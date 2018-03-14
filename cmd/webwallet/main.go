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
