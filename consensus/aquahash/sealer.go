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
	crand "crypto/rand"
	"encoding/binary"
	"math"
	"math/big"
	"math/rand"
	"runtime"
	"sync"

	"gitlab.com/aquachain/aquachain/common"
	"gitlab.com/aquachain/aquachain/common/log"
	"gitlab.com/aquachain/aquachain/consensus"
	"gitlab.com/aquachain/aquachain/core/types"
	"gitlab.com/aquachain/aquachain/crypto"
	"gitlab.com/aquachain/aquachain/params"
)

// Seal implements consensus.Engine, attempting to find a nonce that satisfies
// the block's difficulty requirements.
func (aquahash *Aquahash) Seal(chain consensus.ChainReader, block *types.Block, stop <-chan struct{}) (*types.Block, error) {
	// If we're running a fake PoW, simply return a 0 nonce immediately
	chaincfg := params.TestChainConfig
	if chain != nil {
		chaincfg = chain.Config()
	}
	if aquahash.config.PowMode == ModeFake || aquahash.config.PowMode == ModeFullFake {
		header := block.Header()
		header.Version = chaincfg.GetBlockVersion(header.Number)
		header.Nonce, header.MixDigest = types.BlockNonce{}, common.Hash{}
		return block.WithSeal(header), nil
	}
	// If we're running a shared PoW, delegate sealing to it
	if aquahash.shared != nil {
		log.Debug("delegating work", "block", block.Number(), "version", block.Version())
		return aquahash.shared.Seal(chain, block, stop)
	}
	// Create a runner and the multiple search threads it directs
	abort := make(chan struct{})
	found := make(chan *types.Block)

	aquahash.lock.Lock()
	threads := aquahash.threads
	if aquahash.rand == nil {
		seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			aquahash.lock.Unlock()
			return nil, err
		}
		aquahash.rand = rand.New(rand.NewSource(seed.Int64()))
	}
	aquahash.lock.Unlock()
	if threads == 0 {
		threads = runtime.NumCPU()
	}
	if threads < 0 {
		threads = 0 // Allows disabling local mining without extra logic around local/remote
	}
	var pend sync.WaitGroup
	version := chaincfg.GetBlockVersion(block.Number())
	for i := 0; i < threads; i++ {
		pend.Add(1)
		go func(id int, nonce uint64) {
			defer pend.Done()
			aquahash.mine(version, block, id, nonce, abort, found)
		}(i, uint64(aquahash.rand.Int63()))
	}
	// Wait until sealing is terminated or a nonce is found
	var result *types.Block
	select {
	case <-stop:
		// Outside abort, stop all miner threads
		close(abort)
	case result = <-found:
		// One of the threads found a block, abort all others
		close(abort)
	case <-aquahash.update:
		// Thread count was changed on user request, restart
		close(abort)
		pend.Wait()
		return aquahash.Seal(chain, block, stop)
	}
	// Wait for all miners to terminate and return the block
	pend.Wait()
	return result, nil
}

// mine is the actual proof-of-work miner that searches for a nonce starting from
// seed that results in correct final block difficulty.
func (aquahash *Aquahash) mine(version params.HeaderVersion, block *types.Block, id int, seed uint64, abort chan struct{}, found chan *types.Block) {
	// Extract some data from the header
	var (
		header  = block.Header()
		hash    = header.HashNoNonce().Bytes()
		target  = new(big.Int).Div(maxUint256, header.Difficulty)
		number  = header.Number.Uint64()
		dataset = aquahash.dataset(number)
	)
	header.Version = version
	// Start generating random nonces until we abort or find a good one
	var (
		attempts = int64(0)
		nonce    = seed
	)
	logger := log.New("miner", id)
	logger.Trace("Started aquahash search for new nonces", "seed", seed)
search:
	for {

		select {
		case <-abort:
			// Mining terminated, update stats and abort
			logger.Trace("Aquahash nonce search aborted", "attempts", nonce-seed)
			aquahash.hashrate.Mark(attempts)
			break search

		default:
			// We don't have to update hash rate on every nonce, so update after after 2^X nonces
			attempts++
			if (attempts % (1 << 15)) == 0 {
				aquahash.hashrate.Mark(attempts)
				attempts = 0
			}

			// Compute the PoW value of this nonce
			var (
				digest []byte
				result []byte
			)
			switch header.Version {
			case 1:
				digest, result = hashimotoFull(dataset.dataset, hash, nonce)
			case 2, 3, 4:
				seed := make([]byte, 40)
				copy(seed, hash)
				binary.LittleEndian.PutUint64(seed[32:], nonce)
				result = crypto.VersionHash(byte(header.Version), seed)
				digest = make([]byte, common.HashLength)
			default:
				logger.Error("Mining incorrect version", "block", number, "version", version)
				break search
			}

			if new(big.Int).SetBytes(result).Cmp(target) <= 0 {
				// Correct nonce found, create a new header with it
				header = types.CopyHeader(header)
				header.Nonce = types.EncodeNonce(nonce)
				header.MixDigest = common.BytesToHash(digest)

				// Seal and return a block (if still needed)
				select {
				case found <- block.WithSeal(header):
					logger.Trace("Aquahash nonce found and reported", "attempts", nonce-seed, "nonce", nonce)
				case <-abort:
					logger.Trace("Aquahash nonce found but discarded", "attempts", nonce-seed, "nonce", nonce)
				}
				break search
			}
			nonce++
		}
	}
	// Datasets are unmapped in a finalizer. Ensure that the dataset stays live
	// during sealing so it's not unmapped while being read.
	runtime.KeepAlive(dataset)
}
