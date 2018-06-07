package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/aquachain/aquachain/cmd/internal/browser"
	"gitlab.com/aquachain/aquachain/cmd/internal/maw"
	"gitlab.com/aquachain/aquachain/cmd/utils"
	"gitlab.com/aquachain/aquachain/crypto"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	mawCommand = cli.Command{
		Name:     "maw",
		Usage:    `Launch MyAquaWallet, an offline wallet that connects to this aquachain program via JSON-RPC/HTTP`,
		Category: "ACCOUNT COMMANDS",
		Action:   launchmaw,
		Description: `
	aquachain maw

will launch browser MAW`,
	}
	walletCommand = cli.Command{
		Name:     "wallet",
		Usage:    `Launch MyAquaWallet, an offline wallet that connects to this aquachain program via JSON-RPC/HTTP`,
		Category: "ACCOUNT COMMANDS",
		Action:   launchmaw,
		Description: `
  aquachain wallet

will launch browser MAW`,
	}
	paperCommand = cli.Command{
		Name:      "paper",
		Usage:     `Generate paper wallet keypair`,
		Flags:     []cli.Flag{utils.JsonFlag, utils.VanityFlag},
		ArgsUsage: "[number of wallets]",
		Category:  "ACCOUNT COMMANDS",
		Action:    paper,
		Description: `
Generate a number of wallets.`,
	}
)

func launchmaw(c *cli.Context) error {
	mawserver := maw.New("127.0.0.1:8042")
	go mawserver.Serve()
	if !c.GlobalBool("rpc") {
		c.GlobalSet("rpc", "true")
	}
	c.GlobalSet("rpccorsdomain", "http://localhost:8042")
	if !c.GlobalBool("rpc") {
		return fmt.Errorf("Please use the -rpc flag when using MAW")
	}
	node := makeFullNode(c)
	if err := node.Start(); err != nil {
		return err
	}
	node.Server().Logger.Info("Serving MAW", "port", "8042", "url", "http://localhost:8042")
	browser.Open("http://localhost:8042")
	node.Wait()
	return nil
}

type paperWallet struct{ Private, Public string }

func paper(c *cli.Context) error {

	if c.NArg() > 1 {
		return fmt.Errorf("too many arguments")
	}
	var (
		count = 1
		err   error
	)
	if c.NArg() == 1 {
		count, err = strconv.Atoi(c.Args().First())
		if err != nil {
			return err
		}
	}
	wallets := []paperWallet{}
	vanity := c.String("vanity")
	for i := 0; i < count; i++ {
		var wallet paperWallet
		for {
			key, err := crypto.GenerateKey()
			if err != nil {
				return err
			}

			addr := crypto.PubkeyToAddress(key.PublicKey)
			wallet = paperWallet{
				Private: hex.EncodeToString(crypto.FromECDSA(key)),
				Public:  "0x" + hex.EncodeToString(addr[:]),
			}
			if vanity == "" {
				break
			}
			pubkey := hex.EncodeToString(addr[:])
			if strings.HasPrefix(pubkey, vanity) {
				break
			}
			//println(pubkey, "!=", "vanity")
		}
		if c.Bool("json") {
			wallets = append(wallets, wallet)
		} else {
			fmt.Println(wallet.Private, wallet.Public)
		}
	}
	if c.Bool("json") {
		b, _ := json.Marshal(wallets)
		fmt.Println(string(b))
	}
	return nil
}
