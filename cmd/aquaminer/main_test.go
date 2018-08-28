package main

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"

	"gitlab.com/aquachain/aquachain/common"
	"gitlab.com/aquachain/aquachain/crypto"
)

func TestMiner(t *testing.T) {
	// dummy work load
	workHash := common.HexToHash("0xd3b5f1b47f52fdc72b1dab0b02ab352442487a1d3a43211bc4f0eb5f092403fc")
	target := new(big.Int).SetBytes(common.HexToHash("0x08637bd05af6c69b5a63f9a49c2c1b10fd7e45803cd141a6937d1fe64f54").Bytes())

	// good nonce
	nonce := uint64(14649775584697213406)

	seed := make([]byte, 40)
	copy(seed, workHash.Bytes())
	fmt.Printf("hashing work: %x\nless than target:  %s\nnonce: %v\n", workHash, target, nonce)

	// debug
	fmt.Printf("seednononc: %x\n", seed)

	// little endian
	binary.LittleEndian.PutUint64(seed[32:], nonce)

	// pre hash
	fmt.Printf("beforehash: %x\n", seed)

	// hash
	result := crypto.VersionHash(2, seed)

	// difficulty
	out := new(big.Int).SetBytes(result)
	fmt.Printf("result difficulty: %s\n", out)
	fmt.Printf("result difficulty: %x\n", out)

	// test against target difficulty
	testresult := out.Cmp(target) <= 0
	fmt.Printf("%x: %v\n", out, testresult)
	if !testresult {
		t.FailNow()
	}
}
