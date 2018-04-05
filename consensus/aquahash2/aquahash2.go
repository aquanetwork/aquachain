package aquahash2

import (
	"errors"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/aquanetwork/aquachain/consensus"
	"github.com/aquanetwork/aquachain/metrics"
	"github.com/aquanetwork/aquachain/rpc"
)

var ErrInvalidDumpMagic = errors.New("invalid dump magic")

var (
	// maxUint256 is a big integer representing 2^256-1
	maxUint256      = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), big.NewInt(0))
	maxUint256small = new(big.Int).Exp(big.NewInt(2), big.NewInt(155), big.NewInt(0))

	// sharedAquahash is a full instance that can be shared between multiple users.
	sharedAquahash = New(Config{"", 3, 0, "", 1, 0, ModeNormal})
)

const (
	datasetInitBytes   = 1 << 30 // Bytes in dataset at genesis
	datasetGrowthBytes = 1 << 23 // Dataset growth per epoch
	cacheInitBytes     = 1 << 24 // Bytes in cache at genesis
	cacheGrowthBytes   = 1 << 17 // Cache growth per epoch
	epochLength        = 30000   // Blocks per epoch
	mixBytes           = 128     // Width of mix
	hashBytes          = 64      // Hash length in bytes
	hashWords          = 16      // Number of 32 bit ints in a hash
	datasetParents     = 256     // Number of parents of each dataset element
	cacheRounds        = 3       // Number of rounds in cache production
	loopAccesses       = 64      // Number of accesses in hashimoto loop
)

// Mode defines the type and amount of PoW verification an aquahash engine makes.
type Mode uint

const (
	ModeNormal Mode = iota
	ModeShared
	ModeTest
	ModeFake
	ModeFullFake
)

// Config are the configuration parameters of the aquahash.
type Config struct {
	CacheDir       string
	CachesInMem    int
	CachesOnDisk   int
	DatasetDir     string
	DatasetsInMem  int
	DatasetsOnDisk int
	PowMode        Mode
}

// Aquahash is a consensus engine based on proot-of-work implementing the aquahash
// algorithm.
type Aquahash2 struct {
	config Config

	// Mining related fields
	rand     *rand.Rand    // Properly seeded random source for nonces
	threads  int           // Number of threads to mine on if mining
	update   chan struct{} // Notification channel to update mining parameters
	hashrate metrics.Meter // Meter tracking the average hashrate
	// The fields below are hooks for testing
	shared    *Aquahash2    // Shared PoW verifier to avoid cache regeneration
	fakeFail  uint64        // Block number which fails PoW check even in fake mode
	fakeDelay time.Duration // Time delay to sleep for before returning from verify

	lock sync.Mutex // Ensures thread safety for the in-memory caches and mining fields
}

// fnv is an algorithm inspired by the FNV hash, which in some cases is used as
// a non-associative substitute for XOR. Note that we multiply the prime with
// the full 32-bit input, in contrast with the FNV-1 spec which multiplies the
// prime with one byte (octet) in turn.
func fnv(a, b uint32) uint32 {
	return a*0x01000193 ^ b
}

// fnvHash mixes in data into mix using the aquahash fnv method.
func fnvHash(mix []uint32, data []uint32) {
	for i := 0; i < len(mix); i++ {
		mix[i] = mix[i]*0x01000193 ^ data[i]
	}
}

// Hashrate implements PoW, returning the measured rate of the search invocations
// per second over the last minute.
func (aquahash *Aquahash2) Hashrate() float64 {
	return aquahash.hashrate.Rate1()
}

// APIs implements consensus.Engine, returning the user facing RPC APIs. Currently
// that is empty.
func (aquahash *Aquahash2) APIs(chain consensus.ChainReader) []rpc.API {
	return nil
}

// // SeedHash is the seed to use for generating a verification cache and the mining
// // dataset.
// func SeedHash(block uint64) []byte {
// 	return seedHash(block)
// }

// New creates a full sized aquahash PoW scheme.
func New(config Config) *Aquahash2 {
	// if config.CachesInMem <= 0 {
	// 	log.Warn("One aquahash cache must always be in memory", "requested", config.CachesInMem)
	// 	config.CachesInMem = 1
	// }
	// if config.CacheDir != "" && config.CachesOnDisk > 0 {
	// 	log.Info("Disk storage enabled for aquahash caches", "dir", config.CacheDir, "count", config.CachesOnDisk)
	// }
	// if config.DatasetDir != "" && config.DatasetsOnDisk > 0 {
	// 	log.Info("Disk storage enabled for aquahash DAGs", "dir", config.DatasetDir, "count", config.DatasetsOnDisk)
	// }
	return &Aquahash2{
		config:   config,
		update:   make(chan struct{}),
		hashrate: metrics.NewMeter(),
	}
}

// NewTester creates a small sized aquahash PoW scheme useful only for testing
// purposes.
func NewTester() *Aquahash2 {
	return New(Config{CachesInMem: 1, PowMode: ModeTest})
}

// NewFaker creates a aquahash consensus engine with a fake PoW scheme that accepts
// all blocks' seal as valid, though they still have to conform to the AquaChain
// consensus rules.
func NewFaker() *Aquahash2 {
	return &Aquahash2{
		config: Config{
			PowMode: ModeFake,
		},
	}
}

// NewFakeFailer creates a aquahash consensus engine with a fake PoW scheme that
// accepts all blocks as valid apart from the single one specified, though they
// still have to conform to the AquaChain consensus rules.
func NewFakeFailer(fail uint64) *Aquahash2 {
	return &Aquahash2{
		config: Config{
			PowMode: ModeFake,
		},
		fakeFail: fail,
	}
}

// NewFakeDelayer creates a aquahash consensus engine with a fake PoW scheme that
// accepts all blocks as valid, but delays verifications by some time, though
// they still have to conform to the AquaChain consensus rules.
func NewFakeDelayer(delay time.Duration) *Aquahash2 {
	return &Aquahash2{
		config: Config{
			PowMode: ModeFake,
		},
		fakeDelay: delay,
	}
}

// NewFullFaker creates an aquahash consensus engine with a full fake scheme that
// accepts all blocks as valid, without checking any consensus rules whatsoever.
func NewFullFaker() *Aquahash2 {
	return &Aquahash2{
		config: Config{
			PowMode: ModeFullFake,
		},
	}
}

// NewShared creates a full sized aquahash PoW shared between all requesters running
// in the same process.
func NewShared() *Aquahash2 {
	return &Aquahash2{shared: sharedAquahash}
}
