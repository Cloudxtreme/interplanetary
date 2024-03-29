package offline

import (
	"testing"

	context "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/go.net/context"

	blocks "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/blocks"
	u "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"
)

func TestBlockReturnsErr(t *testing.T) {
	off := NewOfflineExchange()
	_, err := off.Block(context.Background(), u.Key("foo"))
	if err != nil {
		return // as desired
	}
	t.Fail()
}

func TestHasBlockReturnsNil(t *testing.T) {
	off := NewOfflineExchange()
	block := blocks.NewBlock([]byte("data"))
	err := off.HasBlock(context.Background(), *block)
	if err != nil {
		t.Fatal("")
	}
}
