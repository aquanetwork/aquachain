package main

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.com/aquachain/aquachain/cmd/utils"
	"gitlab.com/aquachain/aquachain/common"
	"gitlab.com/aquachain/aquachain/consensus/lightvalid"
	"gitlab.com/aquachain/aquachain/core/types"
	"gitlab.com/aquachain/aquachain/internal/debug"
	"gitlab.com/aquachain/aquachain/opt/aquaclient"
	"gitlab.com/aquachain/aquachain/params"
	"gitlab.com/aquachain/aquachain/rlp"
	"gitlab.com/aquachain/aquachain/rpc"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	big1   = big.NewInt(1)
	Config = params.TestnetChainConfig
)

const ticker uint64 = 60

var gitCommit = ""

var (
	app = utils.NewApp(gitCommit, "usage")
)

func init() {
	app.Name = "aquacli"
	app.Action = loopit
	_ = filepath.Join
	app.Flags = append(debug.Flags, []cli.Flag{
		cli.StringFlag{
			//Value: filepath.Join(utils.DataDirFlag.Value.String(), "testnet/aquachain.ipc"),
			Value: "https://tx.aquacha.in/testnet/",
			Name:  "rpc",
			Usage: "path or url to rpc",
		},
	}...)
}

//valid block #1 using -testnet2
var header1 = &types.Header{
	Difficulty: big.NewInt(4096),
	Extra:      []byte{0xd4, 0x83, 0x01, 0x07, 0x04, 0x89, 0x61, 0x71, 0x75, 0x61, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x85, 0x6c, 0x69, 0x6e, 0x75, 0x78},
	GasLimit:   4704588,
	GasUsed:    0,
	// Hash: "0x73851a4d607acd8341cf415beeed9c8b8c803e1e835cb45080f6af7a2127e807",
	Coinbase:    common.HexToAddress("0xcf8e5ba37426404bef34c3ca4fa2d2ed9be41e58"),
	MixDigest:   common.Hash{},
	Nonce:       types.BlockNonce{0x70, 0xc2, 0xdd, 0x45, 0xa3, 0x10, 0x17, 0x35},
	Number:      big.NewInt(1),
	ParentHash:  common.HexToHash("0xde434983d3ada19cd43c44d8ad5511bad01ed12b3cc9a99b1717449a245120df"),
	ReceiptHash: common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
	UncleHash:   common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
	Root:        common.HexToHash("0x194b1927f77b77161b58fed1184990d8f7b345fabf8ef8706ee865a844f73bc3"),
	Time:        big.NewInt(1536181711),
	TxHash:      common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
	Version:     2,
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println("fatal:", err)
	}
}

func loopit(ctx *cli.Context) error {
	for {
		if err := runit(ctx); err != nil {
			fmt.Println(err)
		}
	}
}
func runit(ctx *cli.Context) error {
	rpcclient, err := getclient(ctx)
	if err != nil {
		return err
	}
	aqua := aquaclient.NewClient(rpcclient)
	parent, err := aqua.BlockByNumber(context.Background(), nil)
	if err != nil {
		return err
	}
	// prime rpc server for submitting work
	//	_, _ = aqua.GetWork(context.Background())

	var encoded []byte
	// first block is on the house
	if false && parent.Number().Uint64() == 0 {
		parent.SetVersion(Config.GetBlockVersion(parent.Number()))
		block1 := types.NewBlock(header1, nil, nil, nil)
		encoded, err = rlp.EncodeToBytes(&block1)
		if err != nil {
			return err
		}
	} else {
		encoded, err = aqua.GetBlockTemplate(context.Background())
		if err != nil {
			return err
		}
		var bt types.Block
		if err := rlp.DecodeBytes(encoded, &bt); err != nil {
			println("submitblock rlp decode error", err.Error())
			return err
		}
		bt.SetVersion(Config.GetBlockVersion(bt.Number()))
		encoded, err = mine(&bt)
		if err != nil {
			return err
		}
	}
	if encoded == nil {
		return fmt.Errorf("failed to encoded block to rlp")
	}

	if !aqua.SubmitBlock(context.Background(), encoded) {
		fmt.Println("failed")
		return fmt.Errorf("failed")
	} else {
		fmt.Println("success")
	}
	return nil

}

func mine(block *types.Block) ([]byte, error) {
	validator := lightvalid.New()
	rand.Seed(time.Now().UnixNano())
	nonce := uint64(0)
	nonce = rand.Uint64()
	hdr := block.Header()
	fmt.Println("mining algo:", hdr.Version)
	fmt.Printf("#%v, by %x\ndiff: %s\ntx: %s\n", hdr.Number, hdr.Coinbase, hdr.Difficulty, block.Transactions())
	fmt.Printf("starting from nonce: %v\n", nonce)
	second := time.Tick(time.Duration(ticker) * time.Second)
	fps := uint64(0)
	for {
		select {
		case <-second:
			fmt.Printf("%v h/s\n", fps/uint64(ticker))
			fps = 0
		default:
			nonce++
			fps++
			hdr.Nonce = types.EncodeNonce(nonce)
			block = block.WithSeal(hdr)
			if err := validator.VerifyWithError(block); err != nil {
				if err != lightvalid.ErrPOW {
					fmt.Println("error:", err)
				}
				continue
			}
			println("encoding block", block.String())
			b, err := rlp.EncodeToBytes(&block)
			if err != nil {
				return nil, err
			}
			fmt.Println(b)
			return b, nil
		}
	}
}

func getclient(ctx *cli.Context) (*rpc.Client, error) {
	if strings.HasPrefix(ctx.String("rpc"), "http") {
		return rpc.DialHTTP(ctx.String("rpc"))
	} else {
		return rpc.DialIPC(context.Background(), ctx.String("rpc"))
	}
}
