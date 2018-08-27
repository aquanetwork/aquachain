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
	BlockReward            *big.Int = big.NewInt(1e+18) // Block reward in wei for successfully mining a block
	ByzantiumBlockReward   *big.Int = big.NewInt(1e+18) // Block reward in wei for successfully mining a block upward from Byzantium
	maxUncles                       = 2                 // Maximum number of uncles allowed in a single block
	maxUnclesHF5                    = 1                 // Maximum number of uncles allowed in a single block after HF5 is activated
	allowedFutureBlockTime          = 15 * time.Second  // Max time from current time allowed for blocks, before they're considered future blocks
	big0                            = big.NewInt(0)
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
	parent := chain.GetHeader(header.ParentHash, number-1)
	if parent == nil && number >= 1 {
		return consensus.ErrUnknownAncestor
	}
	grandparent := chain.GetHeader(parent.ParentHash, number-2)
	if grandparent == nil && number >= 2 {
		log.Error("no grandparent", "number", number)
		return consensus.ErrUnknownAncestor
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
		if parent == nil && headers[0].Number.Uint64() > 0 {
			return consensus.ErrUnknownAncestor
		} else if parent == nil && headers[0].Number.Uint64() == 0 { // genesis has no parent
			// do nothing here
		} else {
			grandparent = chain.GetHeader(parent.ParentHash, parent.Number.Uint64()-1)
			if grandparent == nil && headers[index].Number.Uint64() > 3 {
				return consensus.ErrUnknownAncestor
			}
		}
	} else if index == 1 {
		parent = headers[index-1]
		if parent == nil {
			return consensus.ErrUnknownAncestor
		}
		grandparent = chain.GetHeader(parent.ParentHash, parent.Number.Uint64()-1)
		if grandparent == nil && headers[index].Number.Uint64() > 3 {
			return consensus.ErrUnknownAncestor
		}
	} else if headers[index-1].Hash() == headers[index].ParentHash {
		parent = headers[index-1]
		if parent == nil {
			return consensus.ErrUnknownAncestor
		}
		grandparent = headers[index-2]

		if grandparent == nil && headers[index].Number.Uint64() > 3 {
			return consensus.ErrUnknownAncestor
		}
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
			//switch hash.Hex() {
			//		case "0x13cb01d5d3566d076b5e128e5733f17968f95329fb1777iff38db53abdcca3e4c":
			//default:
			//println("uncle: " + hash.Hex())
			//common.Report(block)
			return errUncleIsAncestor
			//}
		}
		if ancestors[uncle.ParentHash] == nil || uncle.ParentHash == block.ParentHash() {
			parentHash := uncle.ParentHash.Hex()
			if parentHash == "0x6b818656fb5059ab4dd070e2c2822a7774065090e74ff31515764212c88e2923" && uncle.Number.Uint64() == 14003 {
				log.Debug("Weird block", "uncle", uncle.Number, "number", number)
				return nil
			} else if parentHash == "0x0afd1b00b8e1a49652beeb860e3b58dacc865dd3e3d9d303374ed3ffdfef8eea" && uncle.Number.Uint64() == 14001 {
				log.Debug("Weird block", "uncle", uncle.Number, "number", number)
				return nil
			} else if hash.Hex() == "0xed6dae6d2d4f599d78429e127e8a654fe96c30f4b6c9bacb01cfa45d8a57b45e" && uncle.Number.Uint64() == 14004 {
				log.Debug("Weird block", "uncle", uncle.Number, "number", number)
				return nil
			} else if hash.Hex() == "0x13cb01d5d3566d076b5e128e5733f17968f95329fb1777ff38db53abdcca3e4c" && uncle.Number.Uint64() == 14008 {
				log.Debug("Weird block", "uncle", uncle.Number, "number", number)
				return nil
			} else if hash.Hex() == "0x822735d89d8493434d3ec1f504c9f103d7bb4761cd358370b00dd234621cf1b9" && uncle.Number.Uint64() == 14009 {
				log.Debug("Weird block", "uncle", uncle.Number, "number", number)
				return nil
			} else {
				return errDanglingUncle
			}
		}
		if err := aquahash.verifyHeader(chain, uncle, nil, ancestors[uncle.ParentHash], true, true); err != nil {
			return err
		}
	}
	return nil
}

// verifyHeader checks whether a header conforms to the consensus rules of the
// stock AquaChain aquahash engine.
// See YP section 4.3.4. "Block Header Validity"
func (aquahash *Aquahash) verifyHeader(chain consensus.ChainReader, header, parent *types.Header, grandparent *types.Header, uncle bool, seal bool) error {
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

	if parent == nil && header.Number.Cmp(big0) != 0 {
		parent = chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
		if parent == nil {
			log.Error("nil parent?", "header", header)
			return fmt.Errorf("invalid block: has no parent")
		}
	}
	if parent != nil && header.Time.Cmp(parent.Time) <= 0 {
		return errZeroBlockTime
	}
	// Verify the block's difficulty based in it's timestamp and parent's difficulty
	expected := aquahash.CalcDifficulty(chain, header.Time.Uint64(), parent, grandparent)

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
		if err := aquahash.VerifySeal(chain, header); err != nil {
			return err
		}
	}
	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func (aquahash *Aquahash) CalcDifficulty(chain consensus.ChainReader, time uint64, parent *types.Header, grandparent *types.Header) *big.Int {
	return CalcDifficultyHF8(chain, time, parent, grandparent)
}

func CalcDifficultyHF8(chain consensus.ChainReader, time uint64, parent *types.Header, grandparent *types.Header) *big.Int {
	var (
		config = chain.Config()
	)

	if config.ChainId == nil {
		panic("calc difficulty: no chainID set")
	}

	var (
		next    = new(big.Int).Add(parent.Number, big1)
		chainID = config.ChainId.Uint64()
		mainnet = config.ChainId.Cmp(params.MainnetChainConfig.ChainId) == 0 // bool
	)

	switch {
	// hardfork 9
	case (config.GetHF(9) != nil && next.Cmp(config.GetHF(9)) == 0):
		diff := new(big.Int).Sub(parent.Difficulty, new(big.Int).Div(parent.Difficulty, params.JumpDifficultyHF9)) // reset diff since pow is much different
		log.Info("Activating Hardfork", "HF", 9, "BlockNumber", config.GetHF(9), "Difficulty", diff)
		log.Debug("HF9 Difficulty Jump", "number", next, "version", config.GetBlockVersion(next), "parentDiff", parent.Difficulty, "gparentDiff", grandparent.Difficulty)
		if mainnet {
			return diff
		} else {
			return params.MinimumDifficultyHF5Testnet
		}

	case config.IsHF(9, next):
		log.Debug("HF9 Difficulty", "number", next, "version", config.GetBlockVersion(next), "parentDiff", parent.Difficulty, "gparentDiff", grandparent.Difficulty)
		return calcDifficultyGrandparent(time, parent, grandparent, 9, chainID)

	// hardfork 8
	case (config.GetHF(8) != nil && next.Cmp(config.GetHF(8)) == 0):
		diff := new(big.Int).Sub(parent.Difficulty, new(big.Int).Div(parent.Difficulty, params.JumpDifficultyHF8)) // reset diff since pow is much different
		log.Debug("HF8 Difficulty", "number", next, "version", config.GetBlockVersion(next))
		log.Info("Activating Hardfork", "HF", 8, "BlockNumber", config.GetHF(8), "Difficulty", diff)
		if mainnet {
			return diff
		} else {
			return params.MinimumDifficultyHF5Testnet
		}

	case config.IsHF(8, next):
		log.Debug("HF8 Difficulty", "number", next, "version", config.GetBlockVersion(next))
		return calcDifficultyHF6(time, parent, 8, chainID)

	default:
		return CalcDifficulty(config, time, parent)
	}

}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func CalcDifficulty(config *params.ChainConfig, time uint64, parent *types.Header) *big.Int {
	if config.IsHF(8, new(big.Int).Add(big.NewInt(1), parent.Number)) {
		panic("should not be using CalcDifficulty after HF8, use CalcDifficultyHF8")
	}

	if config.ChainId == nil {
		panic("calc difficulty: no chainID set")
	}

	var (
		next    = new(big.Int).Add(parent.Number, big1)
		chainID = config.ChainId.Uint64()
		mainnet = config.ChainId.Cmp(params.MainnetChainConfig.ChainId) == 0 // bool
	)

	switch {

	// mainnet difficulty adjustment
	// if a hardfork is not listed here it means there was no difficulty algorithm adjustments in that hf
	//
	// hardfork 6
	case (config.GetHF(6) != nil && next.Cmp(config.GetHF(6)) == 0):
		log.Info("Activating Hardfork", "HF", 6, "BlockNumber", config.GetHF(6))
		return calcDifficultyHF6(time, parent, 6, chainID)
	case config.IsHF(6, next):
		log.Debug("HF6 Difficulty", "number", next)
		return calcDifficultyHF6(time, parent, 6, chainID)

	// hardfork 5
	case (config.GetHF(5) != nil && next.Cmp(config.GetHF(5)) == 0):
		log.Info("Activating Hardfork", "HF", 5, "BlockNumber", config.GetHF(5))
		if mainnet {
			return params.MinimumDifficultyHF5 // reset diff since pow is much different
		} else {
			return params.MinimumDifficultyHF5Testnet
		}
	case config.IsHF(5, next):
		return calcDifficultyHF6(time, parent, 5, chainID)

	// hardfork 3
	case (config.GetHF(3) != nil && next.Cmp(config.GetHF(3)) == 0):
		log.Info("Activating Hardfork", "HF", 3, "BlockNumber", config.GetHF(3))
		return calcDifficultyHF6(time, parent, 3, chainID)
	case config.IsHF(3, next):
		return calcDifficultyHF6(time, parent, 3, chainID)

	// hardfork 2
	case (config.GetHF(2) != nil && next.Cmp(config.GetHF(2)) == 0):
		log.Info("Activating Hardfork", "HF", 2, "BlockNumber", config.GetHF(2))
		return calcDifficultyHF6(time, parent, 2, chainID)
	case config.IsHF(2, next):
		return calcDifficultyHF6(time, parent, 2, chainID)

	// hardfork 1
	case (config.GetHF(1) != nil && next.Cmp(config.GetHF(1)) == 0):
		log.Info("Activating Hardfork", "HF", 1, "BlockNumber", config.GetHF(1))
		return calcDifficultyHF1(time, parent, chainID)

	case config.IsHF(1, next):
		return calcDifficultyHF1(time, parent, chainID)

	// genesis to HF1
	case config.IsHomestead(next):
		return calcDifficultyHomestead(time, parent, chainID)
	default:
		panic("unknown block type: " + next.String())
		return calcDifficultyHomestead(time, parent, chainID)
	}
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
		digest     []byte
		result     []byte
		multiplier = big.NewInt(1)
	)
	switch header.Version {
	default: // types.H_UNSET: // 0
		panic("header version not set")
	case types.H_KECCAK256: // 1
		digest, result = hashimotoLight(size, cache.cache, header.HashNoNonce().Bytes(), header.Nonce.Uint64())
	case types.H_ARGON2ID_A: // 2
		seed := make([]byte, 40)
		copy(seed, header.HashNoNonce().Bytes())
		binary.LittleEndian.PutUint64(seed[32:], header.Nonce.Uint64())
		result = crypto.Argon2idA(seed)
		digest = make([]byte, common.HashLength)
	case types.H_ARGON2ID_B: // 3
		seed := make([]byte, 40)
		copy(seed, header.HashNoNonce().Bytes())
		binary.LittleEndian.PutUint64(seed[32:], header.Nonce.Uint64())
		result = crypto.Argon2idB(seed)
		digest = make([]byte, common.HashLength)
	case types.H_ARGON2ID_C: // 4
		seed := make([]byte, 40)
		copy(seed, header.HashNoNonce().Bytes())
		binary.LittleEndian.PutUint64(seed[32:], header.Nonce.Uint64())
		result = crypto.Argon2idC(seed)
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
	if resultDiff := new(big.Int).Mul(new(big.Int).SetBytes(result), multiplier); resultDiff.Cmp(target) > 0 {
		log.Error("difficulty", "target", target, "result", resultDiff)
		return errInvalidPoW
	}
	return nil
}

// Prepare implements consensus.Engine, initializing the difficulty field of a
// header to conform to the aquahash protocol. The changes are done inline.
func (aquahash *Aquahash) Prepare(chain consensus.ChainReader, header *types.Header) error {
	parent := chain.GetHeader(header.ParentHash, header.Number.Uint64()-1)
	if parent == nil {
		return consensus.ErrUnknownAncestor
	}
	grandparent := chain.GetHeader(parent.ParentHash, header.Number.Uint64()-2)
	if grandparent == nil && header.Number.Uint64() > 2 {
		return consensus.ErrUnknownAncestor
	}
	if header.Version == 0 {
		header.Version = chain.Config().GetBlockVersion(header.Number)
	}
	header.Difficulty = CalcDifficultyHF8(chain, header.Time.Uint64(), parent, grandparent)
	return nil
}

// Finalize implements consensus.Engine, accumulating the block and uncle rewards,
// setting the final state and assembling the block.
func (aquahash *Aquahash) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	// Accumulate any block and uncle rewards and commit the final state root
	if header.Version == 0 {
		header.Version = chain.Config().GetBlockVersion(header.Number)
	}
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

// AccumulateRewards credits the coinbase of the given block with the mining
// reward. The total reward consists of the static block reward and rewards for
// included uncles. The coinbase of each uncle block is also rewarded.
func accumulateRewards(config *params.ChainConfig, state *state.StateDB, header *types.Header, uncles []*types.Header) {
	// Select the correct block reward based on chain progression
	blockReward := BlockReward

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

		r.Div(blockReward, big32)
		reward.Add(reward, r)
	}
	state.AddBalance(header.Coinbase, reward)
}
