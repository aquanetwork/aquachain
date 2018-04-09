package argonated

import (
	"fmt"
	"math/big"
	"runtime"
	"time"

	"github.com/aquanetwork/aquachain/common"
	"github.com/aquanetwork/aquachain/common/math"
	"github.com/aquanetwork/aquachain/consensus"
	"github.com/aquanetwork/aquachain/consensus/misc"
	"github.com/aquanetwork/aquachain/core/state"
	"github.com/aquanetwork/aquachain/core/types"
	"github.com/aquanetwork/aquachain/log"
	"github.com/aquanetwork/aquachain/params"
	"github.com/aquanetwork/aquachain/rpc"
)

var (
	big1 = big.NewInt(1)
	big0 = big.NewInt(0)
)

func (c *Consensus) Name() string { return "argonated" }

// Author retrieves the AquaChain address of the account that minted the given
// block, which may be different from the header's coinbase if a consensus
// engine is based on signatures.
func (c *Consensus) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules of a
// given engine. Verifying the seal may be done optionally here, or explicitly
// via the VerifySeal method.
func (c *Consensus) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool) error {
	if !chain.Config().IsHF(5, header.Number) {
		log.Trace("VerifyHeader with aquahash")
		return c.aquahash.VerifyHeader(chain, header, seal)
	}
	log.Trace("VerifyHeader argonated")

	// If we're running a full engine faking, accept any input as valid
	if c.config.PowMode == ModeFullFake {
		return nil
	}
	// Short circuit if the header is known, or it's parent not
	number := header.Number.Uint64()
	if chain.GetHeader(header.Hash(), number) != nil {
		return nil
	}
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	// Sanity checks passed, do a proper verification
	return c.verifyHeader(chain, header, parent, false, seal)

}
func (c *Consensus) verifyHeader(chain consensus.ChainReader, header, parent *types.Header, uncle bool, seal bool) error {
	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
	}
	// Verify the header's timestamp
	if uncle {
		if header.Time.Cmp(math.MaxBig256) > 0 {
			return errLargeBlockTime
		}
	} else {
		if header.Time.Cmp(big.NewInt(time.Now().Add(allowedFutureBlockTime).Unix())) > 0 {
			return consensus.ErrFutureBlock
		}
	}
	if header.Time.Cmp(parent.Time) <= 0 {
		return errZeroBlockTime
	}
	// Verify the block's difficulty based in it's timestamp and parent's difficulty
	expected := c.CalcDifficulty(chain, header.Time.Uint64(), parent)

	if expected.Cmp(header.Difficulty) != 0 {
		return fmt.Errorf("invalid difficulty: have %v, want %v", header.Difficulty, expected)
	}
	// Verify that the gas limit is <= 2^63-1
	cap := uint64(0x7fffffffffffffff)
	if header.GasLimit > cap {
		return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, cap)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	}

	// Verify that the gas limit remains within allowed bounds
	diff := int64(parent.GasLimit) - int64(header.GasLimit)
	if diff < 0 {
		diff *= -1
	}
	limit := parent.GasLimit / params.GasLimitBoundDivisor

	if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
		return fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parent.GasLimit, limit)
	}
	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return consensus.ErrInvalidNumber
	}
	// Verify the engine specific seal securing the block
	if seal {
		if err := c.VerifySeal(chain, header); err != nil {
			return err
		}
	}
	// If all checks passed, validate any special fields for hard forks
	if err := misc.VerifyHFHeaderExtraData(chain.Config(), header); err != nil {
		return err
	}
	if err := misc.VerifyForkHashes(chain.Config(), header, uncle); err != nil {
		return err
	}
	return nil
}

func (c *Consensus) verifyHeaderWorker(chain consensus.ChainReader, headers []*types.Header, seals []bool, index int) error {
	var parent *types.Header
	if index == 0 {
		parent = chain.GetHeader(headers[0].ParentHash, headers[0].Number.Uint64()-1)
	} else if headers[index-1].Hash() == headers[index].ParentHash {
		parent = headers[index-1]
	}
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	if chain.GetHeader(headers[index].Hash(), headers[index].Number.Uint64()) != nil {
		return nil // known block
	}
	return c.verifyHeader(chain, headers[index], parent, false, seals[index])
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications (the order is that of
// the input slice).
func (c *Consensus) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	// If we're running a full engine faking, accept any input as valid
	if c.config.PowMode == ModeFullFake || len(headers) == 0 {
		abort, results := make(chan struct{}), make(chan error, len(headers))
		for i := 0; i < len(headers); i++ {
			results <- nil
		}
		return abort, results
	}

	// Spawn as many workers as allowed threads
	workers := runtime.GOMAXPROCS(0)
	if len(headers) < workers {
		workers = len(headers)
	}

	// Create a task channel and spawn the verifiers
	var (
		inputs = make(chan int)
		done   = make(chan int, workers)
		errors = make([]error, len(headers))
		abort  = make(chan struct{})
	)
	for i := 0; i < workers; i++ {
		go func() {
			for index := range inputs {
				errors[index] = c.verifyHeaderWorker(chain, headers, seals, index)
				done <- index
			}
		}()
	}

	errorsOut := make(chan error, len(headers))
	go func() {
		defer close(inputs)
		var (
			in, out = 0, 0
			checked = make([]bool, len(headers))
			inputs  = inputs
		)
		for {
			select {
			case inputs <- in:
				if in++; in == len(headers) {
					// Reached end of headers. Stop sending to workers.
					inputs = nil
				}
			case index := <-done:
				for checked[index] = true; checked[out]; out++ {
					errorsOut <- errors[out]
					if out == len(headers)-1 {
						return
					}
				}
			case <-abort:
				return
			}
		}
	}()
	return abort, errorsOut
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of a given engine.
func (c *Consensus) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	if !chain.Config().IsHF(5, block.Number()) {
		log.Trace("VerifyUncles with aquahash")
		return c.aquahash.VerifyUncles(chain, block)
	}

	if len(block.Uncles()) > 0 {

		log.Trace("No argonated uncles")
		return errTooManyUncles

	}
	return nil
}

// VerifySeal checks whether the crypto seal on a header is valid according to
// the consensus rules of the given engine.
func (c *Consensus) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	if !chain.Config().IsHF(5, header.Number) {
		log.Trace("VerifySeal with aquahash")
		return c.aquahash.VerifySeal(chain, header)
	}
	log.Trace("VerifySeal argonated")
	// If we're running a fake PoW, accept any seal as valid
	if c.config.PowMode == ModeFake || c.config.PowMode == ModeFullFake {
		<-time.After(c.fakeDelay)
		if c.fakeFail == header.Number.Uint64() {
			return errInvalidPoW
		}
		return nil
	}

	// If we're running a shared PoW, delegate verification to it
	if c.shared != nil {
		return c.shared.VerifySeal(chain, header)
	}

	// Ensure that we have a valid difficulty for the block
	if header.Difficulty.Sign() <= 0 {
		return errInvalidDifficulty
	}

	hash := HashFull(header.HashNoNonce().Bytes(), header.Nonce.Uint64())

	target := new(big.Int).Div(maxUint256, header.Difficulty)
	if new(big.Int).SetBytes(hash).Cmp(target) > 0 {
		return errInvalidPoW
	}
	return nil
}

// Prepare initializes the consensus fields of a block header according to the
// rules of a particular engine. The changes are executed inline.
func (c *Consensus) Prepare(chain consensus.ChainReader, header *types.Header) error {
	if !chain.Config().IsHF(5, header.Number) {
		log.Trace("Preparing aquahash block")
		return c.aquahash.Prepare(chain, header)
	}

	log.Trace("Preparing argonated block")
	parent := chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	if header.Number.Cmp(chain.Config().GetHF(5)) == 0 {
		header.Difficulty = big.NewInt(2000)
	}
	header.Difficulty = c.CalcDifficulty(chain, header.Time.Uint64(), parent)
	return nil
}

// Finalize runs any post-transaction state modifications (e.g. block rewards)
// and assembles the final block.
// Note: The block header and state database might be updated to reflect any
// consensus rules that happen at finalization (e.g. block rewards).
func (c *Consensus) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction,
	uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	if !chain.Config().IsHF(5, header.Number) {
		log.Trace("Finalizing with aquahash")
		return c.aquahash.Finalize(chain, header, state, txs, uncles, receipts)
	}
	log.Trace("Finalizing argonated")
	// Accumulate any block and uncle rewards and commit the final state root
	accumulateRewards(chain.Config(), state, header, uncles)
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))

	// Header seems complete, assemble into a block and return
	return types.NewBlock(header, txs, uncles, receipts), nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns the difficulty
// that a new block should have.
func (c *Consensus) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header) *big.Int {
	blockNum := new(big.Int).Add(parent.Number, big1)
	cfg := chain.Config()
	if !cfg.IsHF(5, blockNum) {
		log.Trace("Calculating aquahash difficulty")
		return c.aquahash.CalcDifficulty(chain, time, parent)
	}
	if blockNum.Cmp(cfg.GetHF(5)) == 0 {
		log.Info("Killing A.S.I.C. miners!")
		return big.NewInt(8192)
	}
	log.Trace("Calculating argonated difficulty")
	diff := new(big.Int)
	adjust := new(big.Int).Div(parent.Difficulty, params.DifficultyBoundDivisorHF5)
	bigTime := new(big.Int)
	bigParentTime := new(big.Int)

	bigTime.SetUint64(time)
	bigParentTime.Set(parent.Time)

	if bigTime.Sub(bigTime, bigParentTime).Cmp(params.DurationLimit) < 0 {
		diff.Add(parent.Difficulty, adjust)
	} else {
		diff.Sub(parent.Difficulty, adjust)
	}
	return diff
}

// APIs returns the RPC APIs this consensus engine provides.
func (c *Consensus) APIs(chain consensus.ChainReader) []rpc.API {
	return nil
}
