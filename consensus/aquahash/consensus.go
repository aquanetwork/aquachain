// Copyright 2017 The aquachain Authors
// This file is part of the aquachain library.
//
// The aquachain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The aquachain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the aquachain library. If not, see <http://www.gnu.org/licenses/>.

package aquahash

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"time"

	"gitlab.com/aquachain/aquachain/common"
	"gitlab.com/aquachain/aquachain/common/log"
	"gitlab.com/aquachain/aquachain/common/math"
	"gitlab.com/aquachain/aquachain/consensus"
	"gitlab.com/aquachain/aquachain/core/state"
	"gitlab.com/aquachain/aquachain/core/types"
	"gitlab.com/aquachain/aquachain/crypto"
	"gitlab.com/aquachain/aquachain/params"

	set "gopkg.in/fatih/set.v0"
)

// Aquahash proof-of-work protocol constants.
var (
	BlockReward            = params.BlockReward
	maxUncles              = 2                // Maximum number of uncles allowed in a single block
	maxUnclesHF5           = 1                // Maximum number of uncles allowed in a single block after HF5 is activated
	allowedFutureBlockTime = 15 * time.Second // Max time from current time allowed for blocks, before they're considered future blocks
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	errLargeBlockTime    = errors.New("timestamp too big")
	errZeroBlockTime     = errors.New("timestamp equals parent's")
	errTooManyUncles     = errors.New("too many uncles")
	errDuplicateUncle    = errors.New("duplicate uncle")
	errUncleIsAncestor   = errors.New("uncle is ancestor")
	errDanglingUncle     = errors.New("uncle's parent is not ancestor")
	errNonceOutOfRange   = errors.New("nonce out of range")
	errInvalidDifficulty = errors.New("non-positive difficulty")
	errInvalidMixDigest  = errors.New("invalid mix digest")
	errInvalidPoW        = errors.New("invalid proof-of-work")

	errUnknownGrandparent = errors.New("nil grandparent")
)

// Author implements consensus.Engine, returning the header's coinbase as the
// proof-of-work verified author of the block.
func (aquahash *Aquahash) Author(header *types.Header) (common.Address, error) {
	return header.Coinbase, nil
}

// VerifyHeader checks whether a header conforms to the consensus rules of the
// stock AquaChain aquahash engine.
func (aquahash *Aquahash) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool) error {
	// If we're running a full engine faking, accept any input as valid
	if aquahash.config.PowMode == ModeFullFake {
		return nil
	}
	// Short circuit if the header is known, or it's parent not
	number := header.Number.Uint64()
	if chain.GetHeader(header.Hash(), number) != nil {
		return nil
	}

	var parent, grandparent *types.Header
	parent = chain.GetHeader(header.ParentHash, number-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	if number > 2 {
		grandparent = chain.GetHeader(parent.ParentHash, number-2)
		if grandparent == nil {
			return fmt.Errorf("nil grandparent: %v", number-2)
		}
	}
	// Sanity checks passed, do a proper verification
	return aquahash.verifyHeader(chain, header, parent, grandparent, false, seal)
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers
// concurrently. The method returns a quit channel to abort the operations and
// a results channel to retrieve the async verifications.
func (aquahash *Aquahash) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	// If we're running a full engine faking, accept any input as valid
	if aquahash.config.PowMode == ModeFullFake || len(headers) == 0 {
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

	// set version
	for i := range headers {
		if headers[i].Version == 0 {
			panic("hf5: verifyheaders did not receive header version")
		}
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
				errors[index] = aquahash.verifyHeaderWorker(chain, headers, seals, index)
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

func (aquahash *Aquahash) verifyHeaderWorker(chain consensus.ChainReader, headers []*types.Header, seals []bool, index int) error {
	var parent, grandparent *types.Header
	if index == 0 {
		parent = chain.GetHeader(headers[0].ParentHash, headers[0].Number.Uint64()-1)
		if headers[0].Number.Uint64() > 2 && parent != nil {
			grandparent = chain.GetHeader(parent.ParentHash, headers[0].Number.Uint64()-2)
		}
	} else if index == 1 {
		parent = headers[0]
		if parent.Number.Uint64() > 1 {
			grandparent = chain.GetHeader(parent.ParentHash, parent.Number.Uint64()-1)
		}
	} else if headers[index-1].Hash() == headers[index].ParentHash {
		parent = headers[index-1]
		grandparent = headers[index-2]
	}
	if parent == nil && headers[0].Number.Uint64() != 0 {
		return consensus.ErrUnknownAncestor
	}
	if grandparent == nil && parent != nil && parent.Number.Uint64() > 1 {
		return errUnknownGrandparent
	}
	if chain.GetHeader(headers[index].Hash(), headers[index].Number.Uint64()) != nil {
		return nil // known block
	}
	return aquahash.verifyHeader(chain, headers[index], parent, grandparent, false, seals[index])
}

// VerifyUncles verifies that the given block's uncles conform to the consensus
// rules of the stock AquaChain aquahash engine.
func (aquahash *Aquahash) VerifyUncles(chain consensus.ChainReader, block *types.Block) error {
	// If we're running a full engine faking, accept any input as valid
	if aquahash.config.PowMode == ModeFullFake {
		return nil
	}
	// Verify that there are at most 2 uncles included in this block
	if len(block.Uncles()) > maxUncles {
		return errTooManyUncles
	}
	// Verify that there are at most 0 uncles included in this block
	if len(block.Uncles()) > maxUnclesHF5 && chain.Config().IsHF(5, block.Number()) {
		return errTooManyUncles
	}
	// Gather the set of past uncles and ancestors
	uncles, ancestors := set.New(), make(map[common.Hash]*types.Header)

	number, parent := block.NumberU64()-1, block.ParentHash()
	for i := 0; i < 7; i++ {
		ancestor := chain.GetBlock(parent, number)
		if ancestor == nil {
			break
		}
		ancestors[ancestor.Hash()] = ancestor.Header()
		for _, uncle := range ancestor.Uncles() {
			uncles.Add(uncle.SetVersion(byte(chain.Config().GetBlockVersion(uncle.Number))))
		}
		parent, number = ancestor.ParentHash(), number-1
	}
	if block.Version() == 0 {
		return fmt.Errorf("verify uncles: block version not set")
	}
	ancestorhash := block.Hash()
	ancestors[ancestorhash] = block.Header()
	uncles.Add(ancestorhash)

	// Verify each of the uncles that it's recent, but not an ancestor
	for _, uncle := range block.Uncles() {

		unum := uncle.Number
		hash := uncle.SetVersion(byte(chain.Config().GetBlockVersion(unum)))

		// Make sure every uncle is rewarded only once
		if uncles.Has(hash) {
			if number > 15000 {
				return errDuplicateUncle
			} else if ancestorhash.Hex() == "0xbac2283407b519ffbb8c47772d1b7cf740646dddf69744ff44219cb868b00548" && unum.Uint64() == 13313 {
			} else if ancestorhash.Hex() == "0xa955c8499ce9c4fb00700a8d97db8600dc50c8a81275627a18e30cfb82c19ac2" && unum.Uint64() == 13315 {
			} else if ancestorhash.Hex() == "0x7da0315b99e059f17b18bfd7f07c57b8e3be3aac261dbf470fb2d6cb0acb9899" && unum.Uint64() == 13998 {
			} else {
				return errDuplicateUncle
			}
		}

		uncles.Add(hash)

		// Make sure the uncle has a valid ancestry
		if ancestors[hash] != nil {
			return errUncleIsAncestor
		}

		if ancestors[uncle.ParentHash] == nil || uncle.ParentHash == block.ParentHash() {
			if number > 15000 {
				return errDanglingUncle
			}
			parentHash := uncle.ParentHash.Hex()
			if parentHash == "0x6b818656fb5059ab4dd070e2c2822a7774065090e74ff31515764212c88e2923" && uncle.Number.Uint64() == 14003 {
				log.Debug("Weird block", "uncle", unum, "number", number)
				return nil
			} else if parentHash == "0x0afd1b00b8e1a49652beeb860e3b58dacc865dd3e3d9d303374ed3ffdfef8eea" && uncle.Number.Uint64() == 14001 {
				log.Debug("Weird block", "uncle", unum, "number", number)
				return nil
			} else if hash.Hex() == "0xed6dae6d2d4f599d78429e127e8a654fe96c30f4b6c9bacb01cfa45d8a57b45e" && uncle.Number.Uint64() == 14004 {
				log.Debug("Weird block", "uncle", unum, "number", number)
				return nil
			} else if hash.Hex() == "0x13cb01d5d3566d076b5e128e5733f17968f95329fb1777ff38db53abdcca3e4c" && uncle.Number.Uint64() == 14008 {
				log.Debug("Weird block", "uncle", unum, "number", number)
				return nil
			} else if hash.Hex() == "0x822735d89d8493434d3ec1f504c9f103d7bb4761cd358370b00dd234621cf1b9" && uncle.Number.Uint64() == 14009 {
				log.Debug("Weird block", "uncle", unum, "number", number)
				return nil
			} else {
				return errDanglingUncle
			}
		}
		parent := ancestors[uncle.ParentHash]
		grandparent := ancestors[parent.ParentHash]
		if err := aquahash.verifyHeader(chain, uncle, parent, grandparent, true, true); err != nil {
			return err
		}
	}
	return nil
}

// verifyHeader checks whether a header conforms to the consensus rules of the
// stock AquaChain aquahash engine.
// See YP section 4.3.4. "Block Header Validity"
func (aquahash *Aquahash) verifyHeader(chain consensus.ChainReader, header, parent, grandparent *types.Header, uncle bool, seal bool) error {
	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("block %d extra-data too long: %d > %d", header.Number, len(header.Extra), params.MaximumExtraDataSize)
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
	expected := aquahash.CalcDifficulty(chain, header.Time.Uint64(), parent, grandparent)

	if expected.Cmp(header.Difficulty) != 0 {
		return fmt.Errorf("block %d invalid difficulty: have %v, want %v", header.Number, header.Difficulty, expected)
	}
	// Verify that the gas limit is <= 2^63-1
	cap := uint64(0x7fffffffffffffff)
	if header.GasLimit > cap {
		return fmt.Errorf("block %d invalid gasLimit: have %v, max %v", header.Number, header.GasLimit, cap)
	}
	// Verify that the gasUsed is <= gasLimit
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf("block %d invalid gasUsed: have %d, gasLimit %d", header.Number, header.GasUsed, header.GasLimit)
	}

	// Verify that the gas limit remains within allowed bounds
	diff := int64(parent.GasLimit) - int64(header.GasLimit)
	if diff < 0 {
		diff *= -1
	}
	limit := parent.GasLimit / params.GasLimitBoundDivisor

	if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
		return fmt.Errorf("block %d invalid gas limit: have %d, want %d += %d", header.Number, header.GasLimit, parent.GasLimit, limit)
	}
	// Verify that the block number is parent's +1
	if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
		return consensus.ErrInvalidNumber
	}
	// Verify the engine specific seal securing the block
	if seal {
		if err := aquahash.VerifySeal(chain, header); err != nil {
			return err
		}
	}
	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func (aquahash *Aquahash) CalcDifficulty(chain consensus.ChainReader, time uint64, parent, grandparent *types.Header) *big.Int {
	if grandparent == nil && parent != nil && parent.Number.Uint64() != 0 {
		grandparent = chain.GetHeader(parent.ParentHash, parent.Number.Uint64()-1)
	}
	return CalcDifficulty(chain.Config(), time, parent, grandparent)
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func CalcDifficulty(config *params.ChainConfig, time uint64, parent, grandparent *types.Header) *big.Int {
	if config == nil {
		panic("calcdiff got nil config")
	}
	return calcDifficultyHFX(config, time, parent, grandparent)
}

// VerifySeal implements consensus.Engine, checking whether the given block satisfies
// the PoW difficulty requirements.
func (aquahash *Aquahash) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	// If we're running a fake PoW, accept any seal as valid
	if aquahash.config.PowMode == ModeFake || aquahash.config.PowMode == ModeFullFake {
		time.Sleep(aquahash.fakeDelay)
		if aquahash.fakeFail == header.Number.Uint64() {
			return errInvalidPoW
		}
		return nil
	}
	// If we're running a shared PoW, delegate verification to it
	if aquahash.shared != nil {
		return aquahash.shared.VerifySeal(chain, header)
	}
	// Sanity check that the block number is below the lookup table size (60M blocks)
	number := header.Number.Uint64()
	if number/epochLength >= maxEpoch {
		// Go < 1.7 cannot calculate new cache/dataset sizes (no fast prime check)
		return errNonceOutOfRange
	}
	// Ensure that we have a valid difficulty for the block
	if header.Difficulty.Sign() <= 0 {
		return errInvalidDifficulty
	}

	// Recompute the digest and PoW value and verify against the header
	cache := aquahash.cache(number)
	size := datasetSize(number)
	if aquahash.config.PowMode == ModeTest {
		size = 32 * 1024
	}
	var (
		digest []byte
		result []byte
	)
	switch header.Version {
	case types.H_UNSET: // 0
		panic("header version not set")
	case types.H_KECCAK256: // 1
		digest, result = hashimotoLight(size, cache.cache, header.HashNoNonce().Bytes(), header.Nonce.Uint64())
	default:
		seed := make([]byte, 40)
		copy(seed, header.HashNoNonce().Bytes())
		binary.LittleEndian.PutUint64(seed[32:], header.Nonce.Uint64())
		result = crypto.VersionHash(byte(header.Version), seed)
		digest = make([]byte, common.HashLength)
	}
	// Caches are unmapped in a finalizer. Ensure that the cache stays live
	// until after the call to hashimotoLight so it's not unmapped while being used.
	runtime.KeepAlive(cache)

	if !bytes.Equal(header.MixDigest[:], digest) {
		//fmt.Printf("Invalid Digest (%v):\n%x (!=) %x\n", header.Number.Uint64(), header.MixDigest[:], digest)
		return errInvalidMixDigest
	}
	target := new(big.Int).Div(maxUint256, header.Difficulty)
	if new(big.Int).SetBytes(result).Cmp(target) > 0 {
		return errInvalidPoW
	}
	return nil
}

// Prepare implements consensus.Engine, initializing the difficulty field of a
// header to conform to the aquahash protocol. The changes are done inline.
func (aquahash *Aquahash) Prepare(chain consensus.ChainReader, header *types.Header) error {
	var parent, grandparent *types.Header
	parent = chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	if header.Number.Uint64() > 2 {
		grandparent = chain.GetHeader(parent.ParentHash, header.Number.Uint64()-2)
		if grandparent == nil {
			return errUnknownGrandparent
		}
	}
	header.Difficulty = aquahash.CalcDifficulty(chain, header.Time.Uint64(), parent, grandparent)
	return nil
}

// Finalize implements consensus.Engine, accumulating the block and uncle rewards,
// setting the final state and assembling the block.
func (aquahash *Aquahash) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	// Accumulate any block and uncle rewards and commit the final state root
	header.SetVersion(byte(chain.Config().GetBlockVersion(header.Number)))
	for i := range uncles {
		uncles[i].Version = header.Version // uncles must have same version
	}
	accumulateRewards(chain.Config(), state, header, uncles)
	header.Root = state.IntermediateRoot(chain.Config().IsEIP158(header.Number))
	// Header seems complete, assemble into a block and return
	return types.NewBlock(header, txs, uncles, receipts), nil
}

// Some weird constants to avoid constant memory allocs for them.
var (
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)
)

// just for reward log
func weiToAqua(wei *big.Int) string {
	aqu := new(big.Float).SetFloat64(params.Aqua)
	aqu.Set(new(big.Float).Quo(new(big.Float).SetInt(wei), aqu))
	return fmt.Sprintf("%00.2f", aqu)
}

// AccumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward and rewards for
// included uncles. The coinbase of each uncle block is also rewarded.
func accumulateRewards(config *params.ChainConfig, state *state.StateDB, header *types.Header, uncles []*types.Header) {
	// Select the correct block reward based on chain config
	blockReward := BlockReward
	if config == params.EthnetChainConfig {
		blockReward = ethReward(config, header)
	}

	// fees-only after 42,000,000
	// since uncles have a reward too, we will have to adjust this number
	// luckily we have time before we hit anywhere near there
	rewarding := header.Number.Cmp(params.MaxMoney) == -1
	if !rewarding {
		return
	}
	// Accumulate the rewards for the miner and any included uncles
	reward := new(big.Int).Set(blockReward)
	r := new(big.Int)
	for _, uncle := range uncles {
		r.Add(uncle.Number, big8)
		r.Sub(r, header.Number)
		r.Mul(r, blockReward)
		r.Div(r, big8)
		state.AddBalance(uncle.Coinbase, r)
		log.Trace("Uncle reward", "miner", uncle.Coinbase, "reward", weiToAqua(r))

		r.Div(blockReward, big32)
		reward.Add(reward, r)
	}
	state.AddBalance(header.Coinbase, reward)
	log.Trace("Block reward", "miner", header.Coinbase, "reward", weiToAqua(reward))
}
