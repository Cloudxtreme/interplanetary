package dht

import (
	"testing"

	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/peer"
	u "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"

	context "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/go.net/context"
)

func TestProviderManager(t *testing.T) {
	ctx := context.Background()
	mid := peer.ID("testing")
	p := NewProviderManager(ctx, mid)
	a := u.Key("test")
	p.AddProvider(a, peer.WithIDString("testingprovider"))
	resp := p.GetProviders(a)
	if len(resp) != 1 {
		t.Fatal("Could not retrieve provider.")
	}
	p.Close()
}
