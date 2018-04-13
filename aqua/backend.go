// Copyright 2014 The aquachain Authors
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

// Package aqua implements the AquaChain protocol.
package aqua

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/aquanetwork/aquachain/accounts"
	"github.com/aquanetwork/aquachain/aqua/downloader"
	"github.com/aquanetwork/aquachain/aqua/filters"
	"github.com/aquanetwork/aquachain/aqua/gasprice"
	"github.com/aquanetwork/aquachain/aquadb"
	"github.com/aquanetwork/aquachain/common"
	"github.com/aquanetwork/aquachain/common/hexutil"
	"github.com/aquanetwork/aquachain/consensus"
	"github.com/aquanetwork/aquachain/consensus/aquahash"
	"github.com/aquanetwork/aquachain/consensus/clique"
	"github.com/aquanetwork/aquachain/core"
	"github.com/aquanetwork/aquachain/core/bloombits"
	"github.com/aquanetwork/aquachain/core/types"
	"github.com/aquanetwork/aquachain/core/vm"
	"github.com/aquanetwork/aquachain/event"
	"github.com/aquanetwork/aquachain/internal/aquaapi"
	"github.com/aquanetwork/aquachain/log"
	"github.com/aquanetwork/aquachain/miner"
	"github.com/aquanetwork/aquachain/node"
	"github.com/aquanetwork/aquachain/p2p"
	"github.com/aquanetwork/aquachain/params"
	"github.com/aquanetwork/aquachain/rlp"
	"github.com/aquanetwork/aquachain/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// AquaChain implements the AquaChain full node service.
type AquaChain struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the aquachain
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb aquadb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *AquaApiBackend

	miner    *miner.Miner
	gasPrice *big.Int
	aquabase common.Address

	networkId     uint64
	netRPCService *aquaapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and aquabase)
}

func (s *AquaChain) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new AquaChain object (including the
// initialisation of the common AquaChain object)
func New(ctx *node.ServiceContext, config *Config) (*AquaChain, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run aqua.AquaChain in light sync mode, use les.LightAquaChain")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	aqua := &AquaChain{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, &config.Aquahash, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		aquabase:       config.Aquabase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainConfig, chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising AquaChain protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	//if !config.SkipBcVersionCheck {
	bcVersion := core.GetBlockChainVersion(chainDb)
	if bcVersion != core.BlockChainVersion && bcVersion != 0 {
		return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run aquachain upgradedb.\n", bcVersion, core.BlockChainVersion)
	}
	core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	//}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	aqua.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, aqua.chainConfig, aqua.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		aqua.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	aqua.bloomIndexer.Start(aqua.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	aqua.txPool = core.NewTxPool(config.TxPool, aqua.chainConfig, aqua.blockchain)

	if aqua.protocolManager, err = NewProtocolManager(aqua.chainConfig, config.SyncMode, config.NetworkId, aqua.eventMux, aqua.txPool, aqua.engine, aqua.blockchain, chainDb); err != nil {
		return nil, err
	}
	aqua.miner = miner.New(aqua, aqua.chainConfig, aqua.EventMux(), aqua.engine)
	aqua.miner.SetExtra(makeExtraData(config.ExtraData))

	aqua.ApiBackend = &AquaApiBackend{aqua, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	aqua.ApiBackend.gpo = gasprice.NewOracle(aqua.ApiBackend, gpoParams)

	return aqua, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"aquachain",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (aquadb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*aquadb.LDBDatabase); ok {
		db.Meter("aqua/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an AquaChain service
func CreateConsensusEngine(ctx *node.ServiceContext, config *aquahash.Config, chainConfig *params.ChainConfig, db aquadb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.PowMode == aquahash.ModeFake:
		log.Warn("Aquahash used in fake mode")
		return aquahash.NewFaker()
	case config.PowMode == aquahash.ModeTest:
		log.Warn("Aquahash used in test mode")
		return aquahash.NewTester()
	case config.PowMode == aquahash.ModeShared:
		log.Warn("Aquahash used in shared mode")
		return aquahash.NewShared()
	default:
		engine := aquahash.New(aquahash.Config{
			CacheDir:       ctx.ResolvePath(config.CacheDir),
			CachesInMem:    config.CachesInMem,
			CachesOnDisk:   config.CachesOnDisk,
			DatasetDir:     config.DatasetDir,
			DatasetsInMem:  config.DatasetsInMem,
			DatasetsOnDisk: config.DatasetsOnDisk,
		})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs returns the collection of RPC services the aquachain package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *AquaChain) APIs() []rpc.API {
	apis := aquaapi.GetAPIs(s.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "aqua",
			Version:   "1.0",
			Service:   NewPublicAquaChainAPI(s),
			Public:    true,
		}, {
			Namespace: "aqua",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "aqua",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "aqua",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *AquaChain) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *AquaChain) Aquabase() (eb common.Address, err error) {
	s.lock.RLock()
	aquabase := s.aquabase
	s.lock.RUnlock()

	if aquabase != (common.Address{}) {
		return aquabase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			aquabase := accounts[0].Address

			s.lock.Lock()
			s.aquabase = aquabase
			s.lock.Unlock()

			log.Info("Aquabase automatically configured", "address", aquabase)
			return aquabase, nil
		}
	}
	return common.Address{}, fmt.Errorf("aquabase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *AquaChain) SetAquabase(aquabase common.Address) {
	self.lock.Lock()
	self.aquabase = aquabase
	self.lock.Unlock()

	self.miner.SetAquabase(aquabase)
}

func (s *AquaChain) StartMining(local bool) error {
	eb, err := s.Aquabase()
	if err != nil {
		log.Error("Cannot start mining without aquabase", "err", err)
		return fmt.Errorf("aquabase missing: %v", err)
	}
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Aquabase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *AquaChain) StopMining()         { s.miner.Stop() }
func (s *AquaChain) IsMining() bool      { return s.miner.Mining() }
func (s *AquaChain) Miner() *miner.Miner { return s.miner }

func (s *AquaChain) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *AquaChain) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *AquaChain) TxPool() *core.TxPool               { return s.txPool }
func (s *AquaChain) EventMux() *event.TypeMux           { return s.eventMux }
func (s *AquaChain) Engine() consensus.Engine           { return s.engine }
func (s *AquaChain) ChainDb() aquadb.Database           { return s.chainDb }
func (s *AquaChain) IsListening() bool                  { return true } // Always listening
func (s *AquaChain) AquaVersion() int                   { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *AquaChain) NetVersion() uint64                 { return s.networkId }
func (s *AquaChain) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *AquaChain) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// AquaChain protocol implementation.
func (s *AquaChain) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = aquaapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// AquaChain protocol.
func (s *AquaChain) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
