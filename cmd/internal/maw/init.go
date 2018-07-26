// maw package contains Asset("MAW.zip")

package maw

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/spkg/zipfs"
)

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
