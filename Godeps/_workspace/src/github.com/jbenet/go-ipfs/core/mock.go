package core

import (
	ds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-datastore"
	syncds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-datastore/sync"
	bs "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/blockservice"
	ci "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/crypto"
	mdag "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/merkledag"
	nsys "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/namesys"
	path "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/path"
	peer "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/peer"
	mdht "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/routing/mock"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"
)

// NewMockNode constructs an IpfsNode for use in tests.
func NewMockNode() (*IpfsNode, error) {
	nd := new(IpfsNode)

	// Generate Identity
	sk, pk, err := ci.GenerateKeyPair(ci.RSA, 1024)
	if err != nil {
		return nil, err
	}

	p, err := peer.WithKeyPair(sk, pk)
	if err != nil {
		return nil, err
	}

	nd.Peerstore = peer.NewPeerstore()
	nd.Identity, err = nd.Peerstore.Add(p)
	if err != nil {
		return nil, err
	}

	// Temp Datastore
	dstore := ds.NewMapDatastore()
	nd.Datastore = util.CloserWrap(syncds.MutexWrap(dstore))

	// Routing
	dht := mdht.NewMockRouter(nd.Identity, nd.Datastore)
	nd.Routing = dht

	// Bitswap
	//??

	bserv, err := bs.NewBlockService(nd.Datastore, nil)
	if err != nil {
		return nil, err
	}

	nd.DAG = mdag.NewDAGService(bserv)

	// Namespace resolver
	nd.Namesys = nsys.NewNameSystem(dht)

	// Path resolver
	nd.Resolver = &path.Resolver{DAG: nd.DAG}

	return nd, nil
}
