package argonated

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/aquanetwork/aquachain/consensus"
	"github.com/aquanetwork/aquachain/consensus/aquahash"
	"github.com/aquanetwork/aquachain/log"
	"github.com/aquanetwork/aquachain/metrics"
)

const (
	maxUnclesHF5           = 1                // Maximum number of uncles allowed in a 10 block span after HF5 is activated
	allowedFutureBlockTime = 15 * time.Second // Max time from current time allowed for blocks, before they're considered future blocks
	ModeNormal             = consensus.ModeNormal
	ModeTest               = consensus.ModeTest
	ModeFake               = consensus.ModeFake
	ModeFullFake           = consensus.ModeFullFake
	ModeShared             = consensus.ModeShared
)

var (
	// sharedConsensus is a full instance that can be shared between multiple users.
	sharedConsensus = New(Config{PowMode: ModeNormal})
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

// Config are the configuration parameters of the aquahash.
type Config struct {
	PowMode          consensus.Mode
	AquahashConfig   aquahash.Config
	AquahashCacheDir string
}

// Consensus is a consensus engine based on proot-of-work implementing the consensus engine and pow interfaces
type Consensus struct {
	config Config

	// Mining related fields
	rand     *rand.Rand    // Properly seeded random source for nonces
	threads  int           // Number of threads to mine on if mining
	update   chan struct{} // Notification channel to update mining parameters
	hashrate metrics.Meter // Meter tracking the average hashrate

	// The fields below are hooks for testing
	shared    *Consensus    // Shared PoW verifier to avoid cache regeneration
	fakeFail  uint64        // Block number which fails PoW check even in fake mode
	fakeDelay time.Duration // Time delay to sleep for before returning from verify

	lock     sync.Mutex // Ensures thread safety for the in-memory caches and mining fields
	aquahash *aquahash.Aquahash
}

// New creates a full sized aquahash PoW scheme.
func New(config Config) *Consensus {
	aquahashConfig := config.AquahashConfig
	var aquahashEngine *aquahash.Aquahash
	switch config.PowMode {
	case ModeFake:
		log.Warn("Aquahash used in fake mode")
		aquahashEngine = aquahash.NewFaker()
	case ModeTest:
		log.Warn("Aquahash used in test mode")
		aquahashEngine = aquahash.NewTester()
	case ModeShared:
		log.Warn("Aquahash used in shared mode")
		aquahashEngine = aquahash.NewShared()
	default:
		aquahashEngine = aquahash.New(aquahash.Config{
			CacheDir:       config.AquahashCacheDir,
			CachesInMem:    aquahashConfig.CachesInMem,
			CachesOnDisk:   aquahashConfig.CachesOnDisk,
			DatasetDir:     aquahashConfig.DatasetDir,
			DatasetsInMem:  aquahashConfig.DatasetsInMem,
			DatasetsOnDisk: aquahashConfig.DatasetsOnDisk,
		})
		aquahashEngine.SetThreads(-1) // Disable CPU mining non argonated
	}
	return &Consensus{
		config:   config,
		update:   make(chan struct{}),
		hashrate: metrics.NewMeter(),
		aquahash: aquahashEngine,
	}
}

// NewTester creates a small sized aquahash PoW scheme useful only for testing
// purposes.
func NewTester() *Consensus {
	return New(Config{PowMode: ModeTest, AquahashConfig: aquahash.Config{
		CachesInMem: 1, PowMode: aquahash.ModeTest,
	}})
}

// NewFaker creates a aquahash consensus engine with a fake PoW scheme that accepts
// all blocks' seal as valid, though they still have to conform to the AquaChain
// consensus rules.
func NewFaker() *Consensus {
	return New(Config{PowMode: ModeFake, AquahashConfig: aquahash.Config{
		PowMode: aquahash.ModeFake,
	}})
}

// NewFakeFailer creates a aquahash consensus engine with a fake PoW scheme that
// accepts all blocks as valid apart from the single one specified, though they
// still have to conform to the AquaChain consensus rules.
func NewFakeFailer(fail uint64) *Consensus {
	c := New(Config{PowMode: ModeTest, AquahashConfig: aquahash.Config{
		CachesInMem: 1, PowMode: aquahash.ModeTest,
	}})
	c.aquahash = aquahash.NewFullFaker()
	return c
}

// NewFakeDelayer creates a aquahash consensus engine with a fake PoW scheme that
// accepts all blocks as valid, but delays verifications by some time, though
// they still have to conform to the AquaChain consensus rules.
func NewFakeDelayer(delay time.Duration) *Consensus {
	c := New(Config{PowMode: ModeTest})
	c.fakeDelay = delay
	return c
}

// NewFullFaker creates an aquahash consensus engine with a full fake scheme that
// accepts all blocks as valid, without checking any consensus rules whatsoever.
func NewFullFaker() *Consensus {
	return New(Config{PowMode: ModeFullFake})
}

// NewShared creates a full sized aquahash PoW shared between all requesters running
// in the same process.
func NewShared() *Consensus {
	return &Consensus{shared: sharedConsensus}
}

// // SeedHash is the seed to use for generating a verification cache and the mining
// // dataset.
// func SeedHash(block uint64) []byte {
// 	return seedHash(block)
// }

// Threads returns the number of mining threads currently enabled. This doesn't
// necessarily mean that mining is running!
func (c *Consensus) Threads() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.threads
}

// SetThreads updates the number of mining threads currently enabled. Calling
// this method does not start mining, only sets the thread count. If zero is
// specified, the miner will use all cores of the machine. Setting a thread
// count below zero is allowed and will cause the miner to idle, without any
// work being done.
func (c *Consensus) SetThreads(threads int) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// If we're running a shared PoW, set the thread count on that instead
	if c.shared != nil {
		c.shared.SetThreads(threads)
		return
	}
	// Update the threads and ping any running seal to pull in any changes
	c.threads = threads
	select {
	case c.update <- struct{}{}:
	default:
	}
}

// Hashrate implements PoW, returning the measured rate of the search invocations
// per second over the last minute.
func (c *Consensus) Hashrate() float64 {
	return c.hashrate.Rate1()
}
