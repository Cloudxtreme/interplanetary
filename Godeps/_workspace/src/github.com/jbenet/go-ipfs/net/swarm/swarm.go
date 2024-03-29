// package swarm implements a connection muxer with a pair of channels
// to synchronize all network communication.
package swarm

import (
	"errors"
	"fmt"
	"sync"

	conn "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/net/conn"
	msg "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/net/message"
	peer "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/peer"
	u "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"
	ctxc "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util/ctxcloser"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util/eventlog"

	context "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/go.net/context"
	ma "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-multiaddr"
)

var log = eventlog.Logger("swarm")

// ErrAlreadyOpen signals that a connection to a peer is already open.
var ErrAlreadyOpen = errors.New("Error: Connection to this peer already open.")

// ListenErr contains a set of errors mapping to each of the swarms addresses.
// Used to return multiple errors, as in listen.
type ListenErr struct {
	Errors []error
}

func (e *ListenErr) Error() string {
	if e == nil {
		return "<nil error>"
	}
	var out string
	for i, v := range e.Errors {
		if v != nil {
			out += fmt.Sprintf("%d: %s\n", i, v)
		}
	}
	return out
}

// Swarm is a connection muxer, allowing connections to other peers to
// be opened and closed, while still using the same Chan for all
// communication. The Chan sends/receives Messages, which note the
// destination or source Peer.
type Swarm struct {

	// local is the peer this swarm represents
	local peer.Peer

	// peers is a collection of peers for swarm to use
	peers peer.Peerstore

	// Swarm includes a Pipe object.
	*msg.Pipe

	// errChan is the channel of errors.
	errChan chan error

	// conns are the open connections the swarm is handling.
	// these are MultiConns, which multiplex multiple separate underlying Conns.
	conns     conn.MultiConnMap
	connsLock sync.RWMutex

	// listeners for each network address
	listeners []conn.Listener

	// ContextCloser
	ctxc.ContextCloser
}

// NewSwarm constructs a Swarm, with a Chan.
func NewSwarm(ctx context.Context, listenAddrs []ma.Multiaddr, local peer.Peer, ps peer.Peerstore) (*Swarm, error) {
	s := &Swarm{
		Pipe:    msg.NewPipe(10),
		conns:   conn.MultiConnMap{},
		local:   local,
		peers:   ps,
		errChan: make(chan error, 100),
	}

	// ContextCloser for proper child management.
	s.ContextCloser = ctxc.NewContextCloser(ctx, s.close)

	s.Children().Add(1)
	go s.fanOut()
	return s, s.listen(listenAddrs)
}

// close stops a swarm. It's the underlying function called by ContextCloser
func (s *Swarm) close() error {
	// close listeners
	for _, list := range s.listeners {
		list.Close()
	}
	// close connections
	conn.CloseConns(s.Connections()...)
	return nil
}

// Dial connects to a peer.
//
// The idea is that the client of Swarm does not need to know what network
// the connection will happen over. Swarm can use whichever it choses.
// This allows us to use various transport protocols, do NAT traversal/relay,
// etc. to achive connection.
//
// For now, Dial uses only TCP. This will be extended.
func (s *Swarm) Dial(peer peer.Peer) (conn.Conn, error) {
	if peer.ID().Equal(s.local.ID()) {
		return nil, errors.New("Attempted connection to self!")
	}

	// check if we already have an open connection first
	c := s.GetConnection(peer.ID())
	if c != nil {
		return c, nil
	}

	// check if we don't have the peer in Peerstore
	peer, err := s.peers.Add(peer)
	if err != nil {
		return nil, err
	}

	// open connection to peer
	d := &conn.Dialer{
		LocalPeer: s.local,
		Peerstore: s.peers,
	}

	// try to connect to one of the peer's known addresses.
	// for simplicity, we do this sequentially.
	// A future commit will do this asynchronously.
	for _, addr := range peer.Addresses() {
		c, err = d.DialAddr(s.Context(), addr, peer)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	c, err = s.connSetup(c)
	if err != nil {
		c.Close()
		return nil, err
	}

	// TODO replace the TODO ctx with a context passed in from caller
	log.Event(context.TODO(), "dial", peer)
	return c, nil
}

// GetConnection returns the connection in the swarm to given peer.ID
func (s *Swarm) GetConnection(pid peer.ID) conn.Conn {
	s.connsLock.RLock()
	c, found := s.conns[u.Key(pid)]
	s.connsLock.RUnlock()

	if !found {
		return nil
	}
	return c
}

// Connections returns a slice of all connections.
func (s *Swarm) Connections() []conn.Conn {
	s.connsLock.RLock()

	conns := make([]conn.Conn, 0, len(s.conns))
	for _, c := range s.conns {
		conns = append(conns, c)
	}

	s.connsLock.RUnlock()
	return conns
}

// CloseConnection removes a given peer from swarm + closes the connection
func (s *Swarm) CloseConnection(p peer.Peer) error {
	c := s.GetConnection(p.ID())
	if c == nil {
		return u.ErrNotFound
	}

	s.connsLock.Lock()
	delete(s.conns, u.Key(p.ID()))
	s.connsLock.Unlock()

	return c.Close()
}

func (s *Swarm) Error(e error) {
	s.errChan <- e
}

// GetErrChan returns the errors chan.
func (s *Swarm) GetErrChan() chan error {
	return s.errChan
}

// GetPeerList returns a copy of the set of peers swarm is connected to.
func (s *Swarm) GetPeerList() []peer.Peer {
	var out []peer.Peer
	s.connsLock.RLock()
	for _, p := range s.conns {
		out = append(out, p.RemotePeer())
	}
	s.connsLock.RUnlock()
	return out
}
