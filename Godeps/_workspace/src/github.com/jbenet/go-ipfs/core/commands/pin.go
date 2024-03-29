package commands

import (
	"fmt"

	cmds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/commands"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/core"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/merkledag"
	u "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"
)

var pinCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Pin (and unpin) objects to local storage",
	},

	Subcommands: map[string]*cmds.Command{
		"add": addPinCmd,
		"rm":  rmPinCmd,
		"ls":  listPinCmd,
	},
}

var addPinCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Pins objects to local storage",
		ShortDescription: `
Retrieves the object named by <ipfs-path> and stores it locally
on disk.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("ipfs-path", true, true, "Path to object(s) to be pinned"),
	},
	Options: []cmds.Option{
		cmds.BoolOption("recursive", "r", "Recursively pin the object linked to by the specified object(s)"),
	},
	Run: func(req cmds.Request) (interface{}, error) {
		n, err := req.Context().GetNode()
		if err != nil {
			return nil, err
		}

		// set recursive flag
		recursive, found, err := req.Option("recursive").Bool()
		if err != nil {
			return nil, err
		}
		if !found {
			recursive = false
		}

		_, err = pin(n, req.Arguments(), recursive)
		if err != nil {
			return nil, err
		}

		// TODO: create some output to show what got pinned
		return nil, nil
	},
}

var rmPinCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Unpin an object from local storage",
		ShortDescription: `
Removes the pin from the given object allowing it to be garbage
collected if needed.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("ipfs-path", true, true, "Path to object(s) to be unpinned"),
	},
	Options: []cmds.Option{
		cmds.BoolOption("recursive", "r", "Recursively unpin the object linked to by the specified object(s)"),
	},
	Run: func(req cmds.Request) (interface{}, error) {
		n, err := req.Context().GetNode()
		if err != nil {
			return nil, err
		}

		// set recursive flag
		recursive, found, err := req.Option("recursive").Bool()
		if err != nil {
			return nil, err
		}
		if !found {
			recursive = false // default
		}

		_, err = unpin(n, req.Arguments(), recursive)
		if err != nil {
			return nil, err
		}

		// TODO: create some output to show what got unpinned
		return nil, nil
	},
}

var listPinCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "List objects pinned to local storage",
		ShortDescription: `
Returns a list of hashes of objects being pinned. Objects that are indirectly
or recursively pinned are not included in the list.
`,
		LongDescription: `
Returns a list of hashes of objects being pinned. Objects that are indirectly
or recursively pinned are not included in the list.

Use --type=<type> to specify the type of pinned keys to list. Valid values are:
    * "direct"
    * "indirect"
    * "recursive"
    * "all"
(Defaults to "direct")
`,
	},

	Options: []cmds.Option{
		cmds.StringOption("type", "t", "The type of pinned keys to list. Can be \"direct\", \"indirect\", \"recursive\", or \"all\". Defaults to \"direct\""),
	},
	Run: func(req cmds.Request) (interface{}, error) {
		n, err := req.Context().GetNode()
		if err != nil {
			return nil, err
		}

		typeStr, found, err := req.Option("type").String()
		if err != nil {
			return nil, err
		}
		if !found {
			typeStr = "direct"
		}

		switch typeStr {
		case "all", "direct", "indirect", "recursive":
		default:
			return nil, cmds.ClientError("Invalid type '" + typeStr + "', must be one of {direct, indirect, recursive, all}")
		}

		keys := make([]u.Key, 0)
		if typeStr == "direct" || typeStr == "all" {
			keys = append(keys, n.Pinning.DirectKeys()...)
		}
		if typeStr == "indirect" || typeStr == "all" {
			keys = append(keys, n.Pinning.IndirectKeys()...)
		}
		if typeStr == "recursive" || typeStr == "all" {
			keys = append(keys, n.Pinning.RecursiveKeys()...)
		}

		return &KeyList{Keys: keys}, nil
	},
	Type: &KeyList{},
	Marshalers: cmds.MarshalerMap{
		cmds.Text: KeyListTextMarshaler,
	},
}

func pin(n *core.IpfsNode, paths []string, recursive bool) ([]*merkledag.Node, error) {

	dagnodes := make([]*merkledag.Node, 0)
	for _, path := range paths {
		dagnode, err := n.Resolver.ResolvePath(path)
		if err != nil {
			return nil, fmt.Errorf("pin error: %v", err)
		}
		dagnodes = append(dagnodes, dagnode)
	}

	for _, dagnode := range dagnodes {
		err := n.Pinning.Pin(dagnode, recursive)
		if err != nil {
			return nil, fmt.Errorf("pin: %v", err)
		}
	}

	err := n.Pinning.Flush()
	if err != nil {
		return nil, err
	}

	return dagnodes, nil
}

func unpin(n *core.IpfsNode, paths []string, recursive bool) ([]*merkledag.Node, error) {

	dagnodes := make([]*merkledag.Node, 0)
	for _, path := range paths {
		dagnode, err := n.Resolver.ResolvePath(path)
		if err != nil {
			return nil, err
		}
		dagnodes = append(dagnodes, dagnode)
	}

	for _, dagnode := range dagnodes {
		k, _ := dagnode.Key()
		err := n.Pinning.Unpin(k, recursive)
		if err != nil {
			return nil, err
		}
	}

	err := n.Pinning.Flush()
	if err != nil {
		return nil, err
	}
	return dagnodes, nil
}
