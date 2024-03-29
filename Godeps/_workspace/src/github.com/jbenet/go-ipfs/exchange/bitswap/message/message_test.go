package message

import (
	"bytes"
	"testing"

	blocks "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/blocks"
	pb "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/exchange/bitswap/message/internal/pb"
	peer "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/peer"
	u "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"
)

func TestAppendWanted(t *testing.T) {
	const str = "foo"
	m := New()
	m.AddWanted(u.Key(str))

	if !contains(m.ToProto().GetWantlist(), str) {
		t.Fail()
	}
}

func TestNewMessageFromProto(t *testing.T) {
	const str = "a_key"
	protoMessage := new(pb.Message)
	protoMessage.Wantlist = []string{string(str)}
	if !contains(protoMessage.Wantlist, str) {
		t.Fail()
	}
	m := newMessageFromProto(*protoMessage)
	if !contains(m.ToProto().GetWantlist(), str) {
		t.Fail()
	}
}

func TestAppendBlock(t *testing.T) {

	strs := make([]string, 2)
	strs = append(strs, "Celeritas")
	strs = append(strs, "Incendia")

	m := New()
	for _, str := range strs {
		block := blocks.NewBlock([]byte(str))
		m.AddBlock(*block)
	}

	// assert strings are in proto message
	for _, blockbytes := range m.ToProto().GetBlocks() {
		s := bytes.NewBuffer(blockbytes).String()
		if !contains(strs, s) {
			t.Fail()
		}
	}
}

func TestWantlist(t *testing.T) {
	keystrs := []string{"foo", "bar", "baz", "bat"}
	m := New()
	for _, s := range keystrs {
		m.AddWanted(u.Key(s))
	}
	exported := m.Wantlist()

	for _, k := range exported {
		present := false
		for _, s := range keystrs {

			if s == string(k) {
				present = true
			}
		}
		if !present {
			t.Logf("%v isn't in original list", string(k))
			t.Fail()
		}
	}
}

func TestCopyProtoByValue(t *testing.T) {
	const str = "foo"
	m := New()
	protoBeforeAppend := m.ToProto()
	m.AddWanted(u.Key(str))
	if contains(protoBeforeAppend.GetWantlist(), str) {
		t.Fail()
	}
}

func TestToNetMethodSetsPeer(t *testing.T) {
	m := New()
	p := peer.WithIDString("X")
	netmsg, err := m.ToNet(p)
	if err != nil {
		t.Fatal(err)
	}
	if !(netmsg.Peer().Key() == p.Key()) {
		t.Fatal("Peer key is different")
	}
}

func TestToNetFromNetPreservesWantList(t *testing.T) {
	original := New()
	original.AddWanted(u.Key("M"))
	original.AddWanted(u.Key("B"))
	original.AddWanted(u.Key("D"))
	original.AddWanted(u.Key("T"))
	original.AddWanted(u.Key("F"))

	p := peer.WithIDString("X")
	netmsg, err := original.ToNet(p)
	if err != nil {
		t.Fatal(err)
	}

	copied, err := FromNet(netmsg)
	if err != nil {
		t.Fatal(err)
	}

	keys := make(map[u.Key]bool)
	for _, k := range copied.Wantlist() {
		keys[k] = true
	}

	for _, k := range original.Wantlist() {
		if _, ok := keys[k]; !ok {
			t.Fatalf("Key Missing: \"%v\"", k)
		}
	}
}

func TestToAndFromNetMessage(t *testing.T) {

	original := New()
	original.AddBlock(*blocks.NewBlock([]byte("W")))
	original.AddBlock(*blocks.NewBlock([]byte("E")))
	original.AddBlock(*blocks.NewBlock([]byte("F")))
	original.AddBlock(*blocks.NewBlock([]byte("M")))

	p := peer.WithIDString("X")
	netmsg, err := original.ToNet(p)
	if err != nil {
		t.Fatal(err)
	}

	m2, err := FromNet(netmsg)
	if err != nil {
		t.Fatal(err)
	}

	keys := make(map[u.Key]bool)
	for _, b := range m2.Blocks() {
		keys[b.Key()] = true
	}

	for _, b := range original.Blocks() {
		if _, ok := keys[b.Key()]; !ok {
			t.Fail()
		}
	}
}

func contains(s []string, x string) bool {
	for _, a := range s {
		if a == x {
			return true
		}
	}
	return false
}

func TestDuplicates(t *testing.T) {
	b := blocks.NewBlock([]byte("foo"))
	msg := New()

	msg.AddWanted(b.Key())
	msg.AddWanted(b.Key())
	if len(msg.Wantlist()) != 1 {
		t.Fatal("Duplicate in BitSwapMessage")
	}

	msg.AddBlock(*b)
	msg.AddBlock(*b)
	if len(msg.Blocks()) != 1 {
		t.Fatal("Duplicate in BitSwapMessage")
	}
}
