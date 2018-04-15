// aquaminer command is an aquachain miner
package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"math/big"
	"runtime"
	"time"

	"github.com/aquanetwork/aquachain/aquaclient"
	"github.com/aquanetwork/aquachain/cmd/utils"
	"github.com/aquanetwork/aquachain/common"
	"github.com/aquanetwork/aquachain/core/types"
	"github.com/aquanetwork/aquachain/crypto"
)

var (
	rawurl  = "http://localhost:8543"
	max     = new(big.Int).SetUint64(math.MaxUint64)
	digest  = common.BytesToHash(make([]byte, common.HashLength))
	maxproc = flag.Int("t", runtime.NumCPU(), "cpu to use")
	farm    = flag.String("F", "http://localhost:8543", "rpc server to mine to")
)

func main() {

	flag.Parse()
	runtime.GOMAXPROCS(*maxproc)
	runtime.GOMAXPROCS(*maxproc)

	client, err := aquaclient.Dial(*farm)
	if err != nil {
		utils.Fatalf("dial err: %v", err)
	}

	maxProc := runtime.NumCPU()
	if maxProc > *maxproc {
		maxProc = *maxproc
	}
	for i := 0; i < maxProc; i++ {
		go miner(fmt.Sprintf("cpu%v", i+1), client)
	}

	select {}
}

var bigOne = big.NewInt(1)
var oneLsh256 = new(big.Int).Lsh(bigOne, 256)

func big2diff(large *big.Int) uint64 {
	denominator := new(big.Int).Add(large, bigOne)
	return new(big.Int).Div(oneLsh256, denominator).Uint64()

}
func refreshWork(ctx context.Context, client *aquaclient.Client) (common.Hash, *big.Int) {
	work, err := client.GetWork(ctx)
	if err != nil {
		utils.Fatalf("getwork err: %v", err)
	}
	target := new(big.Int).SetBytes(common.HexToHash(work[2]).Bytes())
	return common.HexToHash(work[0]), target
}
func miner(label string, client *aquaclient.Client) {
	var (
		second = time.Tick(time.Second)
		minute = time.Tick(60 * time.Second)
		fps    = 0
		ctx    = context.Background()
	)
	workHash, target := refreshWork(ctx, client)
	log.Printf("Begin new work:\n  HashNoNonce: %s\n  Difficulty %v\n", workHash.Hex(), big2diff(target))
	for {
		fps++
		select {
		case <-minute:
			log.Printf("(%s) %v H/s\n", label, fps/60)
			fps = 0
		case <-second:
			newWorkHash, newTarget := refreshWork(ctx, client)
			if newWorkHash != workHash {
				workHash = newWorkHash
				target = newTarget
				log.Printf("Got new work: %s\nDifficulty: %v\n", workHash.Hex(), big2diff(target))
			}
		default:
		}
		seed := make([]byte, 40)
		copy(seed, workHash.Bytes())
		nonce, err := rand.Int(rand.Reader, max)
		if err != nil {
			utils.Fatalf("prng err: %v", err)
		}
		nuint := nonce.Uint64()
		binary.LittleEndian.PutUint64(seed[32:], nuint)
		result := crypto.Argon2idHash(seed)
		if new(big.Int).SetBytes(result.Bytes()).Cmp(target) <= 0 {
			log.Printf("valid nonce found, submitting:\n%v: %x\n", nuint, result)
			blknonce := types.EncodeNonce(nuint)
			if client.SubmitWork(ctx, blknonce, workHash, digest) {
				log.Println("\n\ngood nonce!\n\n")
			}
			workHash, target = refreshWork(ctx, client)
		}
	}
}
