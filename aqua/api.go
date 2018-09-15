// Copyright 2015 The aquachain Authors
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

package aqua

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math/big"
	"os"
	"sort"
	"strings"

	"gitlab.com/aquachain/aquachain/common"
	"gitlab.com/aquachain/aquachain/common/hexutil"
	"gitlab.com/aquachain/aquachain/common/log"
	"gitlab.com/aquachain/aquachain/core"
	"gitlab.com/aquachain/aquachain/core/state"
	"gitlab.com/aquachain/aquachain/core/types"
	"gitlab.com/aquachain/aquachain/opt/miner"
	"gitlab.com/aquachain/aquachain/params"
	"gitlab.com/aquachain/aquachain/rlp"
	"gitlab.com/aquachain/aquachain/rpc"
)

// PublicTestingAPI provides an API to access new features
type PublicTestingAPI struct {
	cfg   *params.ChainConfig
	agent *miner.RemoteAgent
	e     *AquaChain
}

// NewPublicAquaChainAPI creates a new AquaChain protocol API for full nodes.
func NewPublicTestingAPI(cfg *params.ChainConfig, e *AquaChain) *PublicTestingAPI {
	agent := miner.NewRemoteAgent(e.BlockChain(), e.Engine())
	e.Miner().Register(agent)
	return &PublicTestingAPI{cfg, agent, e}
}

// PublicAquaChainAPI provides an API to access AquaChain full node-related
// information.
type PublicAquaChainAPI struct {
	e *AquaChain
}

// NewPublicAquaChainAPI creates a new AquaChain protocol API for full nodes.
func NewPublicAquaChainAPI(e *AquaChain) *PublicAquaChainAPI {
	return &PublicAquaChainAPI{e}
}

// Aquabase is the address that mining rewards will be send to
func (api *PublicAquaChainAPI) Aquabase() (common.Address, error) {
	return api.e.Aquabase()
}

// Coinbase is the address that mining rewards will be send to (alias for Aquabase)
func (api *PublicAquaChainAPI) Coinbase() (common.Address, error) {
	return api.Aquabase()
}

// Hashrate returns the POW hashrate
func (api *PublicAquaChainAPI) Hashrate() hexutil.Uint64 {
	return hexutil.Uint64(api.e.Miner().HashRate())
}

// PublicMinerAPI provides an API to control the miner.
// It offers only methods that operate on data that pose no security risk when it is publicly accessible.
type PublicMinerAPI struct {
	e     *AquaChain
	agent *miner.RemoteAgent
}

// NewPublicMinerAPI create a new PublicMinerAPI instance.
func NewPublicMinerAPI(e *AquaChain) *PublicMinerAPI {
	agent := miner.NewRemoteAgent(e.BlockChain(), e.Engine())
	e.Miner().Register(agent)

	return &PublicMinerAPI{e, agent}
}

// Mining returns an indication if this node is currently mining.
func (api *PublicMinerAPI) Mining() bool {
	return api.e.IsMining()
}

// SubmitWork can be used by external miner to submit their POW solution. It returns an indication if the work was
// accepted. Note, this is not an indication if the provided work was valid!
func (api *PublicMinerAPI) SubmitWork(nonce types.BlockNonce, solution, digest common.Hash) bool {
	return api.agent.SubmitWork(nonce, digest, solution)
}

// SubmitBlock can be used by external miner to submit their POW solution. It returns an indication if the work was
// accepted. Note, this is not an indication if the provided work was valid!
func (api *PublicTestingAPI) SubmitBlock(encodedBlock []byte) bool {
	var block types.Block
	if encodedBlock == nil {
		log.Warn("submitblock rlp got nil")
		return false
	}
	if err := rlp.DecodeBytes(encodedBlock, &block); err != nil {
		log.Warn("submitblock rlp decode error", "err", err)
		return false
	}
	if block.Nonce() == 0 {
		log.Warn("submitblock got 0 nonce")
		return false
	}
	block.SetVersion(api.e.chainConfig.GetBlockVersion(block.Number()))
	log.Debug("RPC client submitted block:", "block", block.Header())
	return api.agent.SubmitBlock(&block)
}

func (api *PublicTestingAPI) GetBlockTemplate(addr common.Address) ([]byte, error) {
	log.Debug("Got block template request:", "coinbase", addr)
	if !api.e.IsMining() {
		if err := api.e.StartMining(false); err != nil {
			return nil, err
		}
	}
	return api.agent.GetBlockTemplate(addr)
}

// GetWork returns a work package for external miner. The work package consists of 3 strings
// result[0], 32 bytes hex encoded current block header pow-hash
// result[1], 32 bytes hex encoded auxiliary chunk (pre hf5: dag seed, hf5: zeros, hf8: header version)
// result[2], 32 bytes hex encoded boundary condition ("target"), 2^256/difficulty
func (api *PublicMinerAPI) GetWork() ([3]string, error) {
	if !api.e.IsMining() {
		if err := api.e.StartMining(false); err != nil {
			return [3]string{}, err
		}
	}
	work, err := api.agent.GetWork()
	if err != nil {
		return work, fmt.Errorf("mining not ready: %v", err)
	}
	return work, nil
}

// SubmitHashrate can be used for remote miners to submit their hash rate. This enables the node to report the combined
// hash rate of all miners which submit work through this node. It accepts the miner hash rate and an identifier which
// must be unique between nodes.
func (api *PublicMinerAPI) SubmitHashrate(hashrate hexutil.Uint64, id common.Hash) bool {
	// api.agent.SubmitHashrate(id, uint64(hashrate))
	return true
}

// PrivateMinerAPI provides private RPC methods to control the miner.
// These methods can be abused by external users and must be considered insecure for use by untrusted users.
type PrivateMinerAPI struct {
	e *AquaChain
}

// NewPrivateMinerAPI create a new RPC service which controls the miner of this node.
func NewPrivateMinerAPI(e *AquaChain) *PrivateMinerAPI {
	return &PrivateMinerAPI{e: e}
}

// Start the miner with the given number of threads. If threads is nil the number
// of workers started is equal to the number of logical CPUs that are usable by
// this process. If mining is already running, this method adjust the number of
// threads allowed to use.
func (api *PrivateMinerAPI) Start(threads *int) error {
	// Set the number of threads if the seal engine supports it
	if threads == nil {
		threads = new(int)
	} else if *threads == 0 {
		*threads = -1 // Disable the miner from within
	}
	type threaded interface {
		SetThreads(threads int)
	}
	if th, ok := api.e.engine.(threaded); ok {
		log.Info("Updated mining threads", "threads", *threads)
		th.SetThreads(*threads)
	}
	// Start the miner and return
	if !api.e.IsMining() {
		// Propagate the initial price point to the transaction pool
		api.e.lock.RLock()
		price := api.e.gasPrice
		api.e.lock.RUnlock()

		api.e.txPool.SetGasPrice(price)
		return api.e.StartMining(true)
	}
	return nil
}

// Stop the miner
func (api *PrivateMinerAPI) Stop() bool {
	type threaded interface {
		SetThreads(threads int)
	}
	if th, ok := api.e.engine.(threaded); ok {
		th.SetThreads(-1)
	}
	api.e.StopMining()
	return true
}

// SetExtra sets the extra data string that is included when this miner mines a block.
func (api *PrivateMinerAPI) SetExtra(extra string) (bool, error) {
	if err := api.e.Miner().SetExtra([]byte(extra)); err != nil {
		return false, err
	}
	return true, nil
}

// SetGasPrice sets the minimum accepted gas price for the miner.
func (api *PrivateMinerAPI) SetGasPrice(gasPrice hexutil.Big) bool {
	api.e.lock.Lock()
	api.e.gasPrice = (*big.Int)(&gasPrice)
	api.e.lock.Unlock()

	api.e.txPool.SetGasPrice((*big.Int)(&gasPrice))
	return true
}

// SetAquabase sets the aquabase of the miner
func (api *PrivateMinerAPI) SetAquabase(aquabase common.Address) bool {
	api.e.SetAquabase(aquabase)
	return true
}

// GetHashrate returns the current hashrate of the miner.
func (api *PrivateMinerAPI) GetHashrate() uint64 {
	return uint64(api.e.miner.HashRate())
}

// PrivateAdminAPI is the collection of AquaChain full node-related APIs
// exposed over the private admin endpoint.
type PrivateAdminAPI struct {
	aqua *AquaChain
}

// NewPrivateAdminAPI creates a new API definition for the full node private
// admin methods of the AquaChain service.
func NewPrivateAdminAPI(aqua *AquaChain) *PrivateAdminAPI {
	return &PrivateAdminAPI{aqua: aqua}
}

// ExportState exports the current state database into a simplified json file.
func (api *PrivateAdminAPI) ExportState(file string) (bool, error) {
	// Make sure we can create the file to export into
	out, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return false, err
	}
	defer out.Close()
	statedb, err := api.aqua.BlockChain().State()
	if err != nil {
		return false, err
	}
	var writer io.Writer = out
	if strings.HasSuffix(file, ".gz") {
		writer = gzip.NewWriter(writer)
		defer writer.(*gzip.Writer).Close()
	}
	// Export the state
	if err := statedb.TakeSnapshot(writer); err != nil {
		return false, err
	}
	return true, nil
}

// GetDistribution returns a map of address->balance
func (api *PrivateAdminAPI) GetDistribution() (map[string]state.DumpAccount, error) {
	statedb, err := api.aqua.BlockChain().State()
	if err != nil {
		return nil, err
	}
	// Export the state
	dump := statedb.RawDump()
	return dump.Accounts, nil
}

var BigAqua = new(big.Float).SetFloat64(params.Aqua)

func (api *PrivateAdminAPI) GetRichlist(n int) ([]string, error) {
	if n == 0 {
		n = 100
	}
	dist, err := api.GetDistribution()
	if err != nil {
		return nil, err
	}
	type distribResult struct {
		a  string
		ss state.DumpAccount
	}
	var results = []distribResult{}
	for addr, bal := range dist {
		results = append(results, distribResult{addr, bal})
	}
	sort.Slice(results, func(i, j int) bool {
		ii, _ := new(big.Int).SetString(results[i].ss.Balance, 10)
		jj, _ := new(big.Int).SetString(results[j].ss.Balance, 10)
		return ii.Cmp(jj) > 0
	})
	var balances []string
	for i, v := range results {
		if v.ss.Balance != "0" {
			f, _ := new(big.Float).SetString(v.ss.Balance)
			f = f.Quo(f, BigAqua)
			balances = append(balances, fmt.Sprintf("%s: %2.8f", v.a, f))
			if i >= n-1 {
				break
			}
		}
	}
	return balances, nil
}

// Supply returns a map of address->balance
func (api *PrivateAdminAPI) Supply() (*big.Int, error) {
	dump, err := api.GetDistribution()
	if err != nil {
		return nil, err
	}
	total := new(big.Int)

	if len(dump) > 100000 {
		return nil, fmt.Errorf("number of accounts over 100000, bailing")
	}

	bal := make([]string, len(dump))
	n := 0
	for i := range dump {
		bal[n] = dump[i].Balance
		n++
	}

	for i := range bal {
		if bal[i] == "" || bal[i] == "0" {
			continue
		}
		balance, _ := new(big.Int).SetString(bal[i], 10)
		total.Add(total, balance)
	}
	return total, nil
}

// ExportRealloc exports the current state database into a ready to import json file
func (api *PrivateAdminAPI) ExportRealloc(file string) (bool, error) {
	// Make sure we can create the file to export into
	out, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return false, err
	}
	defer out.Close()
	statedb, err := api.aqua.BlockChain().State()
	if err != nil {
		return false, err
	}
	var writer io.Writer = out
	if strings.HasSuffix(file, ".gz") {
		writer = gzip.NewWriter(writer)
		defer writer.(*gzip.Writer).Close()
	}
	writer.Write([]byte(`` +
		`{
"config":{
  "chainId":61717561,
  "homesteadBlock":0,
  "eip150Block":0,
  "eip150Hash":"0x0000000000000000000000000000000000000000000000000000000000000000"
},
  "nonce":"0x2a",
  "timestamp":"0x0",
  "extraData":"0x",
  "gasLimit":"0x401640",
  "difficulty":"0x5f5e0ff",
  "mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000",
  "coinbase":"0x0000000000000000000000000000000000000000",
       "alloc":
`))
	// Export the state
	if err := statedb.TakeSnapshot(writer); err != nil {
		return false, err
	}
	writer.Write([]byte("}"))
	return true, nil
}

// ExportChain exports the current blockchain into a local file.
func (api *PrivateAdminAPI) ExportChain(file string) (bool, error) {
	// Make sure we can create the file to export into
	out, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return false, err
	}
	defer out.Close()

	var writer io.Writer = out
	if strings.HasSuffix(file, ".gz") {
		writer = gzip.NewWriter(writer)
		defer writer.(*gzip.Writer).Close()
	}

	// Export the blockchain
	if err := api.aqua.BlockChain().Export(writer); err != nil {
		return false, err
	}
	return true, nil
}

func hasAllBlocks(chain *core.BlockChain, bs []*types.Block) bool {
	getversion := chain.Config().GetBlockVersion
	for _, b := range bs {
		if !chain.HasBlock(b.SetVersion(getversion(b.Number())), b.NumberU64()) {
			return false
		}
	}

	return true
}

// ImportChain imports a blockchain from a local file.
func (api *PrivateAdminAPI) ImportChain(file string) (bool, error) {
	// Make sure the can access the file to import
	in, err := os.Open(file)
	if err != nil {
		return false, err
	}
	defer in.Close()

	var reader io.Reader = in
	if strings.HasSuffix(file, ".gz") {
		if reader, err = gzip.NewReader(reader); err != nil {
			return false, err
		}
	}

	// Run actual the import in pre-configured batches
	stream := rlp.NewStream(reader, 0)

	blocks, index := make([]*types.Block, 0, 2500), 0
	for batch := 0; ; batch++ {
		// Load a batch of blocks from the input file
		for len(blocks) < cap(blocks) {
			block := new(types.Block)
			if err := stream.Decode(block); err == io.EOF {
				break
			} else if err != nil {
				return false, fmt.Errorf("block %d: failed to parse: %v", index, err)
			}
			blocks = append(blocks, block)
			index++
		}
		if len(blocks) == 0 {
			break
		}

		if hasAllBlocks(api.aqua.BlockChain(), blocks) {
			blocks = blocks[:0]
			continue
		}
		// Import the batch and reset the buffer
		if _, err := api.aqua.BlockChain().InsertChain(blocks); err != nil {
			return false, fmt.Errorf("batch %d: failed to insert: %v", batch, err)
		}
		blocks = blocks[:0]
	}
	return true, nil
}

// PublicDebugAPI is the collection of AquaChain full node APIs exposed
// over the public debugging endpoint.
type PublicDebugAPI struct {
	aqua *AquaChain
}

// NewPublicDebugAPI creates a new API definition for the full node-
// related public debug methods of the AquaChain service.
func NewPublicDebugAPI(aqua *AquaChain) *PublicDebugAPI {
	return &PublicDebugAPI{aqua: aqua}
}

// DumpBlock retrieves the entire state of the database at a given block.
func (api *PublicDebugAPI) DumpBlock(blockNr rpc.BlockNumber) (state.Dump, error) {
	if blockNr == rpc.PendingBlockNumber {
		// If we're dumping the pending state, we need to request
		// both the pending block as well as the pending state from
		// the miner and operate on those
		_, stateDb := api.aqua.miner.Pending()
		return stateDb.RawDump(), nil
	}
	var block *types.Block
	if blockNr == rpc.LatestBlockNumber {
		block = api.aqua.blockchain.CurrentBlock()
	} else {
		block = api.aqua.blockchain.GetBlockByNumber(uint64(blockNr))
	}
	if block == nil {
		return state.Dump{}, fmt.Errorf("block #%d not found", blockNr)
	}
	stateDb, err := api.aqua.BlockChain().StateAt(block.Root())
	if err != nil {
		return state.Dump{}, err
	}
	return stateDb.RawDump(), nil
}

// PrivateDebugAPI is the collection of AquaChain full node APIs exposed over
// the private debugging endpoint.
type PrivateDebugAPI struct {
	config *params.ChainConfig
	aqua   *AquaChain
}

// NewPrivateDebugAPI creates a new API definition for the full node-related
// private debug methods of the AquaChain service.
func NewPrivateDebugAPI(config *params.ChainConfig, aqua *AquaChain) *PrivateDebugAPI {
	return &PrivateDebugAPI{config: config, aqua: aqua}
}

// Preimage is a debug API function that returns the preimage for a sha3 hash, if known.
func (api *PrivateDebugAPI) Preimage(ctx context.Context, hash common.Hash) (hexutil.Bytes, error) {
	db := core.PreimageTable(api.aqua.ChainDb())
	return db.Get(hash.Bytes())
}

// GetBadBlocks returns a list of the last 'bad blocks' that the client has seen on the network
// and returns them as a JSON list of block-hashes
func (api *PrivateDebugAPI) GetBadBlocks(ctx context.Context) ([]core.BadBlockArgs, error) {
	return api.aqua.BlockChain().BadBlocks()
}

// StorageRangeResult is the result of a debug_storageRangeAt API call.
type StorageRangeResult struct {
	Storage storageMap   `json:"storage"`
	NextKey *common.Hash `json:"nextKey"` // nil if Storage includes the last key in the trie.
}

type storageMap map[common.Hash]storageEntry

type storageEntry struct {
	Key   *common.Hash `json:"key"`
	Value common.Hash  `json:"value"`
}
