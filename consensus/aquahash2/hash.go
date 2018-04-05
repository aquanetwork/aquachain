package aquahash2

import (
	"encoding/binary"

	"github.com/aquanetwork/aquachain/common"
	argon2 "golang.org/x/crypto/argon2"
)

func hashingFuncFull(hash []byte, nonce uint64) (digest, result []byte) {
	return hashingFunc(hash, nonce)
}
func hashingFunc(hash []byte, nonce uint64) (digest, result []byte) {
	// Combine header+nonce into a 64 byte seed
	seed := make([]byte, 40)
	salt := []byte("aquahashv0")
	// Calculate the number of theoretical rows (we use one buffer nonetheless)

	copy(seed, hash)
	binary.LittleEndian.PutUint64(seed[32:], nonce)
	seed = argon2.Key(seed, salt, 3, 1024*32, 1, hashBytes)

	////seedHead := binary.LittleEndian.Uint32(seed)

	mix := make([]uint32, mixBytes/4)
	for i := 0; i < len(mix); i++ {
		mix[i] = binary.LittleEndian.Uint32(seed[i%16*4:])
	}

	// Compress mix
	for i := 0; i < len(mix); i += 4 {
		mix[i/4] = fnv(fnv(fnv(mix[i], mix[i+1]), mix[i+2]), mix[i+3])
	}
	mix = mix[:len(mix)/4]

	digest = make([]byte, common.HashLength)
	for i, val := range mix {
		binary.LittleEndian.PutUint32(digest[i*4:], val)
	}
	return digest, argon2.IDKey(append(seed, digest...), salt, 3, 1024*32, 1, hashBytes)
}
