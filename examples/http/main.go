package main

import (
	"log"
	"net/http"
	"os"

	ipfs "github.com/maybebtc/interplanetary"
)

const (
	daemonHostAddr1 = "/ip4/127.0.0.1/tcp/5001"
)

func main() {
	ipfs, err := ipfs.NewClient(daemonHostAddr1)
	if err != nil {
		os.Exit(1)
	}
	log.Fatal(http.ListenAndServe(":8080", http.FileServer(ipfs)))
}
