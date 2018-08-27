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

	"github.com/aerth/tgun"
	"gitlab.com/aquachain/aquachain/cmd/utils"
	"gitlab.com/aquachain/aquachain/common"
	"gitlab.com/aquachain/aquachain/core/types"
	"gitlab.com/aquachain/aquachain/crypto"
	"gitlab.com/aquachain/aquachain/opt/aquaclient"
	"gitlab.com/aquachain/aquachain/rpc"
)

const version = "aquaminer version 0.9x (https://gitlab.com/aquachain/aquachain)"

var (
	digest       = common.BytesToHash(make([]byte, common.HashLength))
	maxproc      = flag.Int("t", runtime.NumCPU(), "number of miners to spawn")
	farm         = flag.String("F", "http://localhost:8543", "rpc server to mine to")
	showVersion  = flag.Bool("version", false, "show version and exit")
	benching     = flag.Bool("B", false, "offline benchmark mode")
	debug        = flag.Bool("d", false, "debug mode")
	benchversion = flag.Uint64("v", 4, "hash version (benchmarking only)")
	nonceseed    = flag.Int64("seed", 1, "nonce seed multiplier")
	refresh      = flag.Duration("r", time.Second*3, "seconds to wait between asking for more work")
	proxypath    = flag.String("prx", "", "example: socks5://192.168.1.3:1080 or 'tor' for localhost:9050")
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
	job     common.Hash
	target  *big.Int
	version uint64
	err     error
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
		client     = &aquaclient.Client{}
	)

	if !*benching {
		tgunner := &tgun.Client{
			UserAgent: "Aquadiver v0.9x",
			Proxy:     *proxypath,
		}
		httpClient, err := tgunner.HTTPClient()
		if err != nil {
			utils.Fatalf("dial err: %v", err)
		}
		rpcclient, err := rpc.DialHTTPWithClient(*farm, httpClient)
		if err != nil {
			utils.Fatalf("dial err: %v", err)
		}
		client = aquaclient.NewClient(rpcclient)
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
		work, target, algo, err := refreshWork(ctx, client, *benching)
		if err != nil {
			log.Println("Error fetching new work from pool:", err)
		}
		if work == cachework {
			continue // dont send already known work
		}
		cachework = work
		log.Printf("Begin new work: %s (difficulty: %v) algo %v\n", work.Hex(), big2diff(target), algo)
		for i := range workers {
			workers[i].newwork <- workload{work, target, algo, err}
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
func refreshWork(ctx context.Context, client *aquaclient.Client, benchmarking bool) (common.Hash, *big.Int, uint64, error) {
	if benchmarking {
		return benchwork, benchdiff, *benchversion, nil
	}
	work, err := client.GetWork(ctx)
	if err != nil {
		return common.Hash{}, benchdiff, 0, fmt.Errorf("getwork err: %v\ncheck address, pool url, and/or local rpc", err)
	}
	target := new(big.Int).SetBytes(common.HexToHash(work[2]).Bytes())
	headerVersion := new(big.Int).SetBytes(common.HexToHash(work[1]).Bytes()).Uint64()
	if *debug {
		fmt.Println(work, "diff:", target, "version:", headerVersion)
	}

	// set header version manually for before hf8
	if headerVersion == 0 || headerVersion > 4 {
		headerVersion = 2
	}
	return common.HexToHash(work[0]), target, headerVersion, nil
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
		algo     uint64
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
			algo = newwork.version
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
			log.Printf("( %s %2.0fH/s (algo #%v)", label, fps/(*refresh).Seconds(), algo)
			fps = 0
		default:
		}

		// increment nonce
		nonce++

		// real actual hashing!
		seed := make([]byte, 40)
		copy(seed, workHash.Bytes())
		binary.LittleEndian.PutUint64(seed[32:], nonce)
		result := common.BytesToHash(crypto.VersionHash(byte(algo), seed))
		// check difficulty of result
		if diff := new(big.Int).SetBytes(result.Bytes()); diff.Cmp(target) <= 0 {
			blknonce := types.EncodeNonce(nonce)
			if offline {
				continue
			}
			// submit the nonce, with the original job
			if client.SubmitWork(ctx, blknonce, workHash, digest) {
				log.Println("good nonce:", nonce)
			} else {
				// there was an error when we send the work. lets get a totally
				log.Println("nonce not accepted", nonce)
				// random nonce, instead of incrementing more
				mrand.Seed(int64(nonce))
				nonce = mrand.Uint64()
			}
		}
	}
}
