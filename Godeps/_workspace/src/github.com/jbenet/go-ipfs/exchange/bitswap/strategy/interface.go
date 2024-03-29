package strategy

import (
	bsmsg "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/exchange/bitswap/message"
	peer "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/peer"
	u "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"
)

type Strategy interface {
	// Returns a slice of Peers with whom the local node has active sessions
	Peers() []peer.Peer

	// BlockIsWantedByPeer returns true if peer wants the block given by this
	// key
	BlockIsWantedByPeer(u.Key, peer.Peer) bool

	// ShouldSendTo(Peer) decides whether to send data to this Peer
	ShouldSendBlockToPeer(u.Key, peer.Peer) bool

	// Seed initializes the decider to a deterministic state
	Seed(int64)

	// MessageReceived records receipt of message for accounting purposes
	MessageReceived(peer.Peer, bsmsg.BitSwapMessage) error

	// MessageSent records sending of message for accounting purposes
	MessageSent(peer.Peer, bsmsg.BitSwapMessage) error

	NumBytesSentTo(peer.Peer) uint64

	NumBytesReceivedFrom(peer.Peer) uint64
}
