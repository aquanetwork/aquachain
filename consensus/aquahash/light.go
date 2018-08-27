package aquahash

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"gitlab.com/aquachain/aquachain/common"
	"gitlab.com/aquachain/aquachain/common/log"
	"gitlab.com/aquachain/aquachain/crypto"
	"gitlab.com/aquachain/aquachain/params"
)

var NoMixDigest = common.Hash{}

func NewLight(config *params.ChainConfig) *Light {
	return &Light{}
}

type Light struct{}
type LightBlock interface {
	Difficulty() *big.Int
	HashNoNonce() common.Hash
	Nonce() uint64
	MixDigest() common.Hash
	NumberU64() uint64
	Version() byte
}

// Verify checks whether the block's nonce is valid.
func (l *Light) Verify(block LightBlock) bool {
	// TODO: do aquahash_quick_verify before getCache in order
	// to prevent DOS attacks.

	if block.Version() < 2 {
		return false
	}

	blockNum := block.NumberU64()
	if blockNum >= epochLength*2048 {
		log.Debug(fmt.Sprintf("block number %d too high, limit is %d", blockNum, epochLength*2048))
		return false
	}

	difficulty := block.Difficulty()
	/* Cannot happen if block header diff is validated prior to PoW, but can
		 happen if PoW is checked first due to parallel PoW checking.
		 We could check the minimum valid difficulty but for SoC we avoid (duplicating)
	   Ethereum protocol consensus rules here which are not in scope of Aquahash
	*/
	if difficulty.Cmp(common.Big0) == 0 {
		log.Debug("invalid block difficulty")
		return false
	}

	// avoid mixdigest malleability as it's not included in a block's "hashNononce"
	if block.MixDigest() != NoMixDigest {
		return false
	}

	seed := make([]byte, 40)
	copy(seed, block.HashNoNonce().Bytes())
	binary.LittleEndian.PutUint64(seed[32:], block.Nonce())
	result := crypto.VersionHash(block.Version(), seed)

	// The actual check.
	target := new(big.Int).Div(maxUint256, difficulty)
	return new(big.Int).SetBytes(result).Cmp(target) <= 0
}
