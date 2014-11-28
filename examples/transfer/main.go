package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	ipfs "github.com/maybebtc/interplanetary"
)

const (
	daemonHostAddr1 = "/ip4/127.0.0.1/tcp/5001"
	// daemonHostAddr1 = "/ip4/104.236.179.241/tcp/5001"
	daemonHostAddr2 = "/ip4/127.0.0.1/tcp/5001"
	// daemonHostAddr2 = "/ip4/104.236.179.241/tcp/49267"
)

func main() {
	fmt.Println(transfer())
}

func transfer() error {
	node1, err := ipfs.NewClient(daemonHostAddr1)
	if err != nil {
		return err
	}
	node2, err := ipfs.NewClient(daemonHostAddr2)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile("main.go")
	if err != nil {
		return err
	}
	k, err := node1.Add(bytes.NewReader(data))
	if err != nil {
		return err
	}
	fmt.Println("added: " + k.String())

	r, err := node2.Cat(k)
	if err != nil {
		return err
	}

	fmt.Println(r)
	return nil
}
