// Copyright 2015 The go-aquachain Authors
// This file is part of the go-aquachain library.
//
// The go-aquachain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-aquachain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-aquachain library. If not, see <http://www.gnu.org/licenses/>.

package aqua

import (
	"context"
	"math/big"

	"github.com/aquanetwork/aquachain/accounts"
	"github.com/aquanetwork/aquachain/aqua/downloader"
	"github.com/aquanetwork/aquachain/aqua/gasprice"
	"github.com/aquanetwork/aquachain/aquadb"
	"github.com/aquanetwork/aquachain/common"
	"github.com/aquanetwork/aquachain/common/math"
	"github.com/aquanetwork/aquachain/core"
	"github.com/aquanetwork/aquachain/core/bloombits"
	"github.com/aquanetwork/aquachain/core/state"
	"github.com/aquanetwork/aquachain/core/types"
	"github.com/aquanetwork/aquachain/core/vm"
	"github.com/aquanetwork/aquachain/event"
	"github.com/aquanetwork/aquachain/params"
	"github.com/aquanetwork/aquachain/rpc"
)

// AquaApiBackend implements aquaapi.Backend for full nodes
type AquaApiBackend struct {
	aqua *AquaChain
	gpo  *gasprice.Oracle
}

func (b *AquaApiBackend) ChainConfig() *params.ChainConfig {
	return b.aqua.chainConfig
}

func (b *AquaApiBackend) CurrentBlock() *types.Block {
	return b.aqua.blockchain.CurrentBlock()
}

func (b *AquaApiBackend) SetHead(number uint64) {
	b.aqua.protocolManager.downloader.Cancel()
	b.aqua.blockchain.SetHead(number)
}

func (b *AquaApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.aqua.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.aqua.blockchain.CurrentBlock().Header(), nil
	}
	return b.aqua.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *AquaApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.aqua.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.aqua.blockchain.CurrentBlock(), nil
	}
	return b.aqua.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *AquaApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.aqua.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.aqua.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *AquaApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.aqua.blockchain.GetBlockByHash(blockHash), nil
}

func (b *AquaApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.aqua.chainDb, blockHash, core.GetBlockNumber(b.aqua.chainDb, blockHash)), nil
}

func (b *AquaApiBackend) GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error) {
	receipts := core.GetBlockReceipts(b.aqua.chainDb, blockHash, core.GetBlockNumber(b.aqua.chainDb, blockHash))
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *AquaApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.aqua.blockchain.GetTdByHash(blockHash)
}

func (b *AquaApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.aqua.BlockChain(), nil)
	return vm.NewEVM(context, state, b.aqua.chainConfig, vmCfg), vmError, nil
}

func (b *AquaApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.aqua.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *AquaApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.aqua.BlockChain().SubscribeChainEvent(ch)
}

func (b *AquaApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.aqua.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *AquaApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.aqua.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *AquaApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.aqua.BlockChain().SubscribeLogsEvent(ch)
}

func (b *AquaApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.aqua.txPool.AddLocal(signedTx)
}

func (b *AquaApiBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.aqua.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *AquaApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.aqua.txPool.Get(hash)
}

func (b *AquaApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.aqua.txPool.State().GetNonce(addr), nil
}

func (b *AquaApiBackend) Stats() (pending int, queued int) {
	return b.aqua.txPool.Stats()
}

func (b *AquaApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.aqua.TxPool().Content()
}

func (b *AquaApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.aqua.TxPool().SubscribeTxPreEvent(ch)
}

func (b *AquaApiBackend) Downloader() *downloader.Downloader {
	return b.aqua.Downloader()
}

func (b *AquaApiBackend) ProtocolVersion() int {
	return b.aqua.AquaVersion()
}

func (b *AquaApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *AquaApiBackend) ChainDb() aquadb.Database {
	return b.aqua.ChainDb()
}

func (b *AquaApiBackend) EventMux() *event.TypeMux {
	return b.aqua.EventMux()
}

func (b *AquaApiBackend) AccountManager() *accounts.Manager {
	return b.aqua.AccountManager()
}

func (b *AquaApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.aqua.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *AquaApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.aqua.bloomRequests)
	}
}
