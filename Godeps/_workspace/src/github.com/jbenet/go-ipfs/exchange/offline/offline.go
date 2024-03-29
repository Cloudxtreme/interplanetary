// package offline implements an object that implements the exchange
// interface but returns nil values to every request.
package offline

import (
	"errors"

	context "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/go.net/context"

	blocks "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/blocks"
	exchange "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/exchange"
	u "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"
)

var OfflineMode = errors.New("Block unavailable. Operating in offline mode")

func NewOfflineExchange() exchange.Interface {
	return &offlineExchange{}
}

// offlineExchange implements the Exchange interface but doesn't return blocks.
// For use in offline mode.
type offlineExchange struct {
}

// Block returns nil to signal that a block could not be retrieved for the
// given key.
// NB: This function may return before the timeout expires.
func (_ *offlineExchange) Block(context.Context, u.Key) (*blocks.Block, error) {
	return nil, OfflineMode
}

// HasBlock always returns nil.
func (_ *offlineExchange) HasBlock(context.Context, blocks.Block) error {
	return nil
}
