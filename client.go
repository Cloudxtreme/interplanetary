package interplanetary

import (
	"bytes"
	"io"
	"net/http"

	cmds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/commands"
	cmds_http "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/commands/http"
	core_cmds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/core/commands"
	errors "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util/debugerror"
	ma "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-multiaddr"
	ma_net "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-multiaddr-net"
)

type Client interface {
	Add(io.Reader) (Key, error)
	Cat(Key) (io.Reader, error)

	SwarmConnect(string) error

	http.FileSystem
}

type client struct {
	httpClient cmds_http.Client
}

func NewClient(addr string) (Client, error) {
	// TODO test returns nil if addr is not a multiaddr
	// TODO allow to connect with either multiaddr or other through configuration option
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return nil, err
	}
	_, host, err := ma_net.DialArgs(maddr)
	if err != nil {
		return nil, err
	}
	return &client{
		httpClient: cmds_http.NewClient(host),
	}, nil
}

func (c *client) Add(r io.Reader) (Key, error) {
	// SliceFile is a workaround for https://github.com/jbenet/go-ipfs/issues/392
	// FIXME pass ReaderFile to NewRequest
	f := &cmds.SliceFile{"TODO",
		[]cmds.File{
			&cmds.ReaderFile{Filename: "TODO", Reader: r},
		},
	}
	req, err := cmds.NewRequest([]string{"add"}, nil, nil, f, core_cmds.AddCmd, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.httpClient.Send(req)
	if err != nil {
		return nil, err
	}
	switch v := res.Output().(type) {
	case *core_cmds.AddOutput:
		if len(v.Objects) < 1 {
			return nil, errors.New("malformed response")
		}
		k, err := parseKey(v.Objects[0].Hash)
		if err != nil {
			return nil, err
		}
		return k, nil
	default:
		return nil, errors.New("unrecognized output format")
	}
}

func (c *client) Cat(k Key) (io.Reader, error) {
	// FIXME workaround for panic in http.Send
	f := &cmds.SliceFile{"TODO",
		[]cmds.File{
			&cmds.ReaderFile{Filename: "TODO", Reader: bytes.NewReader([]byte(""))},
		},
	}
	req, err := cmds.NewRequest([]string{"cat"}, nil, []string{k.String()}, f, core_cmds.CatCmd, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.httpClient.Send(req)
	if err != nil {
		return nil, err
	}
	return res.Reader()
}

func (c *client) SwarmConnect(maddr string) error {
	req, err := cmds.NewRequest([]string{"swarm", "connect"}, nil, []string{maddr}, nil, core_cmds.SwarmCmd, nil)
	if err != nil {
		return err
	}
	if _, err := c.httpClient.Send(req); err != nil {
		return err
	}
	return nil
}

func (c *client) Open(filename string) (http.File, error) {
	return nil, errors.New("TODO")
}

type PeerAddress interface {
	Addr() ma.Multiaddr
	ID() PeerID
}

type PeerID interface {
	Equal(PeerID) bool
}
