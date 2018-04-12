package argonated

import (
	"encoding/binary"
	"hash"
	"runtime"

	"golang.org/x/crypto/argon2"

	"github.com/aquanetwork/aquachain/crypto"
	"github.com/aquanetwork/aquachain/crypto/sha3"
)

// hasher is a repetitive hasher allowing the same hash data structures to be
// reused between hash runs instead of requiring new ones to be created.
type hasher func(dest []byte, data []byte)

// makeHasher creates a repetitive hasher, allowing the same hash data structures
// to be reused between hash runs instead of requiring new ones to be created.
// The returned function is not thread safe!
func makeHasher(h hash.Hash) hasher {
	return func(dest []byte, data []byte) {
		h.Write(data)
		h.Sum(dest[:0])
		h.Reset()
	}
}

// seedHash is the seed to use for generating a verification cache and the mining
// dataset.
func seedHash(block uint64) []byte {
	seed := make([]byte, 32)
	if block < epochLength {
		return seed
	}
	keccak256 := makeHasher(sha3.NewKeccak256())
	for i := 0; i < int(block/epochLength); i++ {
		keccak256(seed, seed)
	}
	return seed
}

const (
	argonTime   uint32 = 1 // argon rounds
	argonMem    uint32 = 1024 * 32
	argonKeyLen uint32 = 32
)

var argonThreads uint8 = uint8(runtime.NumCPU()) * 2

// HashFull runs argon2id(keccak256(hash,nonce))
func HashFull(hash []byte, nonce uint64) []byte {
	seed := make([]byte, 40)
	copy(seed, hash)
	binary.LittleEndian.PutUint64(seed[32:], nonce)
	seed = argon2.IDKey(seed, seed[:32], argonTime, argonMem, argonThreads, argonKeyLen)
	return crypto.Keccak256(seed)
}
