package bitswap

import (
	"bytes"
	"sync"
	"testing"
	"time"

	context "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/go.net/context"

	ds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-datastore"
	ds_sync "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-datastore/sync"
	blocks "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/blocks"
	bstore "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/blockstore"
	exchange "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/exchange"
	notifications "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/exchange/bitswap/notifications"
	strategy "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/exchange/bitswap/strategy"
	tn "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/exchange/bitswap/testnet"
	peer "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/peer"
	mock "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/routing/mock"
	util "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"
)

func TestGetBlockTimeout(t *testing.T) {

	net := tn.VirtualNetwork()
	rs := mock.VirtualRoutingServer()
	g := NewSessionGenerator(net, rs)

	self := g.Next()

	ctx, _ := context.WithTimeout(context.Background(), time.Nanosecond)
	block := blocks.NewBlock([]byte("block"))
	_, err := self.exchange.Block(ctx, block.Key())

	if err != context.DeadlineExceeded {
		t.Fatal("Expected DeadlineExceeded error")
	}
}

func TestProviderForKeyButNetworkCannotFind(t *testing.T) {

	net := tn.VirtualNetwork()
	rs := mock.VirtualRoutingServer()
	g := NewSessionGenerator(net, rs)

	block := blocks.NewBlock([]byte("block"))
	rs.Announce(peer.WithIDString("testing"), block.Key()) // but not on network

	solo := g.Next()

	ctx, _ := context.WithTimeout(context.Background(), time.Nanosecond)
	_, err := solo.exchange.Block(ctx, block.Key())

	if err != context.DeadlineExceeded {
		t.Fatal("Expected DeadlineExceeded error")
	}
}

// TestGetBlockAfterRequesting...

func TestGetBlockFromPeerAfterPeerAnnounces(t *testing.T) {

	net := tn.VirtualNetwork()
	rs := mock.VirtualRoutingServer()
	block := blocks.NewBlock([]byte("block"))
	g := NewSessionGenerator(net, rs)

	hasBlock := g.Next()

	if err := hasBlock.blockstore.Put(block); err != nil {
		t.Fatal(err)
	}
	if err := hasBlock.exchange.HasBlock(context.Background(), *block); err != nil {
		t.Fatal(err)
	}

	wantsBlock := g.Next()

	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	received, err := wantsBlock.exchange.Block(ctx, block.Key())
	if err != nil {
		t.Log(err)
		t.Fatal("Expected to succeed")
	}

	if !bytes.Equal(block.Data, received.Data) {
		t.Fatal("Data doesn't match")
	}
}

func TestSwarm(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	net := tn.VirtualNetwork()
	rs := mock.VirtualRoutingServer()
	sg := NewSessionGenerator(net, rs)
	bg := NewBlockGenerator()

	t.Log("Create a ton of instances, and just a few blocks")

	numInstances := 500
	numBlocks := 2

	instances := sg.Instances(numInstances)
	blocks := bg.Blocks(numBlocks)

	t.Log("Give the blocks to the first instance")

	first := instances[0]
	for _, b := range blocks {
		first.blockstore.Put(b)
		first.exchange.HasBlock(context.Background(), *b)
		rs.Announce(first.peer, b.Key())
	}

	t.Log("Distribute!")

	var wg sync.WaitGroup

	for _, inst := range instances {
		for _, b := range blocks {
			wg.Add(1)
			// NB: executing getOrFail concurrently puts tremendous pressure on
			// the goroutine scheduler
			getOrFail(inst, b, t, &wg)
		}
	}
	wg.Wait()

	t.Log("Verify!")

	for _, inst := range instances {
		for _, b := range blocks {
			if _, err := inst.blockstore.Get(b.Key()); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func getOrFail(bitswap instance, b *blocks.Block, t *testing.T, wg *sync.WaitGroup) {
	if _, err := bitswap.blockstore.Get(b.Key()); err != nil {
		_, err := bitswap.exchange.Block(context.Background(), b.Key())
		if err != nil {
			t.Fatal(err)
		}
	}
	wg.Done()
}

// TODO simplify this test. get to the _essence_!
func TestSendToWantingPeer(t *testing.T) {
	net := tn.VirtualNetwork()
	rs := mock.VirtualRoutingServer()
	sg := NewSessionGenerator(net, rs)
	bg := NewBlockGenerator()

	me := sg.Next()
	w := sg.Next()
	o := sg.Next()

	t.Logf("Session %v\n", me.peer)
	t.Logf("Session %v\n", w.peer)
	t.Logf("Session %v\n", o.peer)

	alpha := bg.Next()

	const timeout = 1 * time.Millisecond // FIXME don't depend on time

	t.Logf("Peer %v attempts to get %v. NB: not available\n", w.peer, alpha.Key())
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	_, err := w.exchange.Block(ctx, alpha.Key())
	if err == nil {
		t.Fatalf("Expected %v to NOT be available", alpha.Key())
	}

	beta := bg.Next()
	t.Logf("Peer %v announes availability  of %v\n", w.peer, beta.Key())
	ctx, _ = context.WithTimeout(context.Background(), timeout)
	if err := w.blockstore.Put(&beta); err != nil {
		t.Fatal(err)
	}
	w.exchange.HasBlock(ctx, beta)

	t.Logf("%v gets %v from %v and discovers it wants %v\n", me.peer, beta.Key(), w.peer, alpha.Key())
	ctx, _ = context.WithTimeout(context.Background(), timeout)
	if _, err := me.exchange.Block(ctx, beta.Key()); err != nil {
		t.Fatal(err)
	}

	t.Logf("%v announces availability of %v\n", o.peer, alpha.Key())
	ctx, _ = context.WithTimeout(context.Background(), timeout)
	if err := o.blockstore.Put(&alpha); err != nil {
		t.Fatal(err)
	}
	o.exchange.HasBlock(ctx, alpha)

	t.Logf("%v requests %v\n", me.peer, alpha.Key())
	ctx, _ = context.WithTimeout(context.Background(), timeout)
	if _, err := me.exchange.Block(ctx, alpha.Key()); err != nil {
		t.Fatal(err)
	}

	t.Logf("%v should now have %v\n", w.peer, alpha.Key())
	block, err := w.blockstore.Get(alpha.Key())
	if err != nil {
		t.Fatal("Should not have received an error")
	}
	if block.Key() != alpha.Key() {
		t.Fatal("Expected to receive alpha from me")
	}
}

func NewBlockGenerator() BlockGenerator {
	return BlockGenerator{}
}

type BlockGenerator struct {
	seq int
}

func (bg *BlockGenerator) Next() blocks.Block {
	bg.seq++
	return *blocks.NewBlock([]byte(string(bg.seq)))
}

func (bg *BlockGenerator) Blocks(n int) []*blocks.Block {
	blocks := make([]*blocks.Block, 0)
	for i := 0; i < n; i++ {
		b := bg.Next()
		blocks = append(blocks, &b)
	}
	return blocks
}

func NewSessionGenerator(
	net tn.Network, rs mock.RoutingServer) SessionGenerator {
	return SessionGenerator{
		net: net,
		rs:  rs,
		seq: 0,
	}
}

type SessionGenerator struct {
	seq int
	net tn.Network
	rs  mock.RoutingServer
}

func (g *SessionGenerator) Next() instance {
	g.seq++
	return session(g.net, g.rs, []byte(string(g.seq)))
}

func (g *SessionGenerator) Instances(n int) []instance {
	instances := make([]instance, 0)
	for j := 0; j < n; j++ {
		inst := g.Next()
		instances = append(instances, inst)
	}
	return instances
}

type instance struct {
	peer       peer.Peer
	exchange   exchange.Interface
	blockstore bstore.Blockstore
}

// session creates a test bitswap session.
//
// NB: It's easy make mistakes by providing the same peer ID to two different
// sessions. To safeguard, use the SessionGenerator to generate sessions. It's
// just a much better idea.
func session(net tn.Network, rs mock.RoutingServer, id peer.ID) instance {
	p := peer.WithID(id)

	adapter := net.Adapter(p)
	htc := rs.Client(p)

	blockstore := bstore.NewBlockstore(ds_sync.MutexWrap(ds.NewMapDatastore()))
	const alwaysSendToPeer = true
	bs := &bitswap{
		blockstore:    blockstore,
		notifications: notifications.New(),
		strategy:      strategy.New(alwaysSendToPeer),
		routing:       htc,
		sender:        adapter,
		wantlist:      util.NewKeySet(),
	}
	adapter.SetDelegate(bs)
	return instance{
		peer:       p,
		exchange:   bs,
		blockstore: blockstore,
	}
}
