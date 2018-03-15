// Copyright 2017 The aquachain Authors
// This file is part of aquachain.
//
// aquachain is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// aquachain is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with aquachain. If not, see <http://www.gnu.org/licenses/>.

// This is a simple Whisper node. It could be used as a stand-alone bootstrap node.
// Also, could be used for different test and diagnostics purposes.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/aquanetwork/aquachain/cmd/internal/browser"
	"github.com/spkg/zipfs"
)

func main() {

	addr := flag.String("addr", "127.0.0.1:8042", "address:port to listen on")
	flag.Parse()
	log.SetFlags(0)

	link := "http://" + *addr
	if _, err := strconv.Atoi(*addr); err == nil {
		link = "http://localhost:" + *addr
		*addr = ":" + *addr
	}
	fmt.Println("Serving", link)
	fmt.Println("Keep this window open while using web wallet")
	if !browser.Open(link) {
		fmt.Println("Use this link to open wallet:", link)
	}

	webwallet := New(*addr)
	log.Fatal(webwallet.Serve())
}

type walletserver struct {
	l   net.Listener
	zfs *zipfs.FileSystem
}

func loadAssets() {
	RestoreAsset(".", "MAW.zip")
}
func New(addr string) *walletserver {
	fs, err := zipfs.New("MAW.zip")
	if err != nil {
		loadAssets()
		fs, err = zipfs.New("MAW.zip")
		if err != nil {
			log.Fatal(err)
		}
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	return &walletserver{
		l:   l,
		zfs: fs,
	}
}

func (ws *walletserver) Serve() error {
	srv := zipfs.FileServer(ws.zfs)
	return http.Serve(ws.l, srv)
}

func (ws *walletserver) Exit() {
	if err := ws.l.Close(); err != nil {
		panic(err)
	}
	os.Exit(0)
}

func (ws *walletserver) String() string {
	return "Aquachain Webwallet"
}

func marshal(i interface{}) []byte {
	b, _ := json.Marshal(i)
	return b
}
