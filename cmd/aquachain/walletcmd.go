package main

import (
	"fmt"

	"github.com/aquanetwork/aquachain/cmd/internal/browser"
	"github.com/aquanetwork/aquachain/cmd/internal/maw"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	walletCommand = cli.Command{
		Name:     "wallet",
		Usage:    `Launch MyAquaWallet, an offline wallet that connects to this aquachain program via JSON-RPC/HTTP`,
		Category: "ACCOUNT COMMANDS",
		Action:   launchmaw,
		Description: `
  aquachain wallet

will launch browser MAW`,
	}
)

func launchmaw(c *cli.Context) error {
	mawserver := maw.New("127.0.0.1:8042")
	go mawserver.Serve()
	if !c.GlobalBool("rpc") {
		c.GlobalSet("rpc", "true")
	}
	if !c.GlobalBool("rpc") {
		return fmt.Errorf("Please use the -rpc flag when using MAW")
	}
	node := makeFullNode(c)
	if err := node.Start(); err != nil {
		return err
	}
	node.Server().Logger.Info("Serving MAW", "port", "8042", "url", "http://127.0.0.1:8042")
	browser.Open("http://127.0.0.1:8042")
	node.Wait()
	return nil
}
