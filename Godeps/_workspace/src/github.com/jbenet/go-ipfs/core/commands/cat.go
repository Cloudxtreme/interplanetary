package commands

import (
	"io"

	cmds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/commands"
	core "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/core"
	uio "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/unixfs/io"
)

var CatCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Show IPFS object data",
		ShortDescription: `
Retrieves the object named by <ipfs-path> and outputs the data
it contains.
`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("ipfs-path", true, true, "The path to the IPFS object(s) to be outputted"),
	},
	Run: func(req cmds.Request) (interface{}, error) {
		node, err := req.Context().GetNode()
		if err != nil {
			return nil, err
		}

		readers := make([]io.Reader, 0, len(req.Arguments()))

		readers, err = cat(node, req.Arguments())
		if err != nil {
			return nil, err
		}

		reader := io.MultiReader(readers...)
		return reader, nil
	},
}

func cat(node *core.IpfsNode, paths []string) ([]io.Reader, error) {
	readers := make([]io.Reader, 0, len(paths))
	for _, path := range paths {
		dagnode, err := node.Resolver.ResolvePath(path)
		if err != nil {
			return nil, err
		}
		read, err := uio.NewDagReader(dagnode, node.DAG)
		if err != nil {
			return nil, err
		}
		readers = append(readers, read)
	}
	return readers, nil
}
