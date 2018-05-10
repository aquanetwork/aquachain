// aquaminer command is an aquachain miner
package main

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"math/big"
	mrand "math/rand"
	"os"
	"runtime"
	"time"

	"github.com/aquanetwork/aquachain/cmd/utils"
	"github.com/aquanetwork/aquachain/common"
	"github.com/aquanetwork/aquachain/core/types"
	"github.com/aquanetwork/aquachain/crypto"
	"github.com/aquanetwork/aquachain/opt/aquaclient"
)

const version = "aquaminer version 0.4 (https://github.com/aquanetwork/aquachain)"

var (
	rawurl      = "http://localhost:8543"
	max         = new(big.Int).SetUint64(math.MaxUint64)
	digest      = common.BytesToHash(make([]byte, common.HashLength))
	maxproc     = flag.Int("t", runtime.NumCPU(), "number of miners to spawn")
	farm        = flag.String("F", "http://localhost:8543", "rpc server to mine to")
	showVersion = flag.Bool("version", false, "show version and exit")
	benching    = flag.Bool("B", false, "offline benchmark mode")
	debug       = flag.Bool("d", false, "debug mode")
	nonceseed   = flag.Int64("seed", 1, "nonce seed multiplier")
	refresh     = flag.Duration("r", time.Second*3, "seconds to wait between asking for more work")
)

// big numbers
var bigOne = big.NewInt(1)
var oneLsh256 = new(big.Int).Lsh(bigOne, 256)

// bench work taken from a testnet work load
var benchdiff = new(big.Int).SetBytes(common.HexToHash("0x08637bd05af6c69b5a63f9a49c2c1b10fd7e45803cd141a6937d1fe64f54").Bytes())
var benchwork = common.HexToHash("0xd3b5f1b47f52fdc72b1dab0b02ab352442487a1d3a43211bc4f0eb5f092403fc")

func init() {
	fmt.Println(version)
}

type workload struct {
	job    common.Hash
	target *big.Int
	err    error
}

type worker struct {
	newwork chan workload
}

func main() {

	flag.Parse()
	if *showVersion {
		os.Exit(0)
	}

	runtime.GOMAXPROCS(*maxproc)
	runtime.GOMAXPROCS(*maxproc)
	if *nonceseed == 1 {
		seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
		if err != nil {
			if err != nil {
				utils.Fatalf("rand err: %v", err)
			}
		}
		*nonceseed = seed.Int64()
	}
	fmt.Println("rand seed:", *nonceseed)
	mrand.Seed(time.Now().UTC().Unix() * *nonceseed)
	var (
		workers    = []*worker{}
		getnewwork = time.Tick(*refresh)
		maxProc    = *maxproc
		err        error
		client     = &aquaclient.Client{}
	)

	if !*benching {
		client, err = aquaclient.Dial(*farm)
		if err != nil {
			utils.Fatalf("dial err: %v", err)
		}
	} else {
		fmt.Println("OFFLINE MODE")
		<-time.After(time.Second)
	}

	// spawn miners
	for i := 0; i < maxProc; i++ {
		w := new(worker)
		w.newwork = make(chan workload, 4) // new work incoming channel
		workers = append(workers, w)
		go miner(fmt.Sprintf("cpu%v", i+1), client, *benching, w.newwork)
	}

	// get work loop
	ctx := context.Background()
	cachework := common.Hash{}
	for range getnewwork { // set -r flag to change this
		work, target, err := refreshWork(ctx, client, *benching)
		if err != nil {
			log.Println("Error fetching new work from pool:", err)
		}
		if work == cachework {
			continue // dont send already known work
		}
		cachework = work
		log.Printf("Begin new work:\n  HashNoNonce: %s\n  Difficulty %v\n", work.Hex(), big2diff(target))
		for i := range workers {
			workers[i].newwork <- workload{work, target, err}
		}
	}
}

// courtesy function to display difficulty for humans
func big2diff(large *big.Int) uint64 {
	if large == nil {
		return 0
	}
	denominator := new(big.Int).Add(large, bigOne)
	return new(big.Int).Div(oneLsh256, denominator).Uint64()

}

// fetch work from a rpc client
func refreshWork(ctx context.Context, client *aquaclient.Client, benchmarking bool) (common.Hash, *big.Int, error) {
	if benchmarking {
		return benchwork, benchdiff, nil
	}
	work, err := client.GetWork(ctx)
	if err != nil {
		return common.Hash{}, benchdiff, fmt.Errorf("getwork err: %v\ncheck address, pool url, and/or local rpc", err)
	}
	if *debug {
		fmt.Println(work)
	}
	target := new(big.Int).SetBytes(common.HexToHash(work[2]).Bytes())
	return common.HexToHash(work[0]), target, nil
}

// single miner loop
func miner(label string, client *aquaclient.Client, offline bool, getworkchan <-chan workload) {

	var (
		second   = time.Tick(*refresh)
		fps      = 0.00
		ctx      = context.Background()
		workHash common.Hash
		target   *big.Int
		err      error
	)

	// remember original nonce
	ononce := mrand.Uint64()
	nonce := ononce
	for {

		// accept new work if available
		select {
		case newwork := <-getworkchan:
			workHash = newwork.job
			target = newwork.target
			err = newwork.err
		default:
		}

		// error fetching work, wait one second and see if theres more work
		if err != nil {
			log.Println("error getting work:", err)
			<-time.After(time.Second)
			continue
		}

		// difficulty isnt set. wait one second for more work.
		if target == nil {
			log.Println(label, "waiting for work...")
			<-time.After(time.Second)
			continue
		}

		// count h/s
		fps++
		select {
		case <-second:
			log.Print("(", label, ")", fps/(*refresh).Seconds(), "H/s\n")
			fps = 0
		default:
		}

		// increment nonce
		nonce++

		// real actual hashing!
		seed := make([]byte, 40)
		copy(seed, workHash.Bytes())
		binary.LittleEndian.PutUint64(seed[32:], nonce)
		result := crypto.Argon2idHash(seed)

		// check difficulty of result
		if new(big.Int).SetBytes(result.Bytes()).Cmp(target) <= 0 {
			log.Print("valid nonce found (", nonce, ")\n")
			blknonce := types.EncodeNonce(nonce)
			if offline {
				continue
			}
			// submit the nonce, with the original job
			if client.SubmitWork(ctx, blknonce, workHash, digest) {
				log.Print("\n\n######\n\nGood Nonce!\n\n#####\n\n")
			} else {
				// there was an error when we send the work. lets get a totally
				// random nonce, instead of incrementing more
				mrand.Seed(int64(nonce))
				nonce = mrand.Uint64()
			}
		}
	}
}
