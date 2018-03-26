// paper command is meant to be easily auditable and can safely generate offline wallets
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aquanetwork/aquachain/crypto"
)

const usage = `This 'paper' command generates aquachain wallets.

Here are 3 ways of using:

Generate 10 in json array:          paper -json 10
Generate 1                          paper
Generate 200                        paper 200

`

type paperWallet struct {
	Private string `json:"private"`
	Public  string `json:"public"`
}

func main() {
	flag.Usage = func() {
		fmt.Println(`                                    _           _
        __ _  __ _ _   _  __ _  ___| |__   __ _(_)_ __
       / _ '|/ _' | | | |/ _' |/ __| '_ \ / _' | | '_ \
      | (_| | (_| | |_| | (_| | (__| | | | (_| | | | | |
       \__,_|\__, |\__,_|\__,_|\___|_| |_|\__,_|_|_| |_|
                |_|` + " https://github.com/aquanetwork/aquachain\n\n")

		fmt.Println(usage)
	}
	log.SetPrefix("")
	log.SetFlags(0)
	jsonFlag := flag.Bool("json", false, "output json")
	flag.Parse()
	n := flag.Args()
	if len(n) != 1 {
		fmt.Println("expecting zero or one argument\n", usage)
		os.Exit(111)
	}
	count, err := strconv.Atoi(n[0])
	if err != nil {
		fmt.Println("expecting digits", usage)
		os.Exit(111)
	}
	wallets := []paperWallet{}
	for i := 0; i < count; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			log.Println("fatal:", err)
			os.Exit(111)
		}

		addr := crypto.PubkeyToAddress(key.PublicKey)
		wallet := paperWallet{
			Private: hex.EncodeToString(crypto.FromECDSA(key)),
			Public:  "0x" + hex.EncodeToString(addr[:]),
		}

		if *jsonFlag {
			wallets = append(wallets, wallet)
		} else {
			fmt.Println(wallet.Private, wallet.Public)
		}
	}
	if *jsonFlag {
		b, _ := json.Marshal(wallets)
		fmt.Println(string(b))
	}

}
