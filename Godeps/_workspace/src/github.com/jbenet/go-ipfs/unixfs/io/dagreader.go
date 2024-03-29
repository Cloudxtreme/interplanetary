package io

import (
	"bytes"
	"errors"
	"io"

	proto "github.com/maybebtc/interplanetary/Godeps/_workspace/src/code.google.com/p/goprotobuf/proto"
	mdag "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/merkledag"
	ft "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/unixfs"
	ftpb "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/unixfs/pb"
)

var ErrIsDir = errors.New("this dag node is a directory")

// DagReader provides a way to easily read the data contained in a dag.
type DagReader struct {
	serv     mdag.DAGService
	node     *mdag.Node
	position int
	buf      io.Reader
}

// NewDagReader creates a new reader object that reads the data represented by the given
// node, using the passed in DAGService for data retreival
func NewDagReader(n *mdag.Node, serv mdag.DAGService) (io.Reader, error) {
	pb := new(ftpb.Data)
	err := proto.Unmarshal(n.Data, pb)
	if err != nil {
		return nil, err
	}

	switch pb.GetType() {
	case ftpb.Data_Directory:
		// Dont allow reading directories
		return nil, ErrIsDir
	case ftpb.Data_File:
		return &DagReader{
			node: n,
			serv: serv,
			buf:  bytes.NewBuffer(pb.GetData()),
		}, nil
	case ftpb.Data_Raw:
		// Raw block will just be a single level, return a byte buffer
		return bytes.NewBuffer(pb.GetData()), nil
	default:
		return nil, ft.ErrUnrecognizedType
	}
}

// precalcNextBuf follows the next link in line and loads it from the DAGService,
// setting the next buffer to read from
func (dr *DagReader) precalcNextBuf() error {
	if dr.position >= len(dr.node.Links) {
		return io.EOF
	}
	nxt, err := dr.node.Links[dr.position].GetNode(dr.serv)
	if err != nil {
		return err
	}
	pb := new(ftpb.Data)
	err = proto.Unmarshal(nxt.Data, pb)
	if err != nil {
		return err
	}
	dr.position++

	switch pb.GetType() {
	case ftpb.Data_Directory:
		// A directory should not exist within a file
		return ft.ErrInvalidDirLocation
	case ftpb.Data_File:
		//TODO: this *should* work, needs testing first
		log.Warning("Running untested code for multilayered indirect FS reads.")
		subr, err := NewDagReader(nxt, dr.serv)
		if err != nil {
			return err
		}
		dr.buf = subr
		return nil
	case ftpb.Data_Raw:
		dr.buf = bytes.NewBuffer(pb.GetData())
		return nil
	default:
		return ft.ErrUnrecognizedType
	}
}

// Read reads data from the DAG structured file
func (dr *DagReader) Read(b []byte) (int, error) {
	// If no cached buffer, load one
	if dr.buf == nil {
		err := dr.precalcNextBuf()
		if err != nil {
			return 0, err
		}
	}
	total := 0
	for {
		// Attempt to fill bytes from cached buffer
		n, err := dr.buf.Read(b[total:])
		total += n
		if err != nil {
			// EOF is expected
			if err != io.EOF {
				return total, err
			}
		}

		// If weve read enough bytes, return
		if total == len(b) {
			return total, nil
		}

		// Otherwise, load up the next block
		err = dr.precalcNextBuf()
		if err != nil {
			return total, err
		}
	}
}

/*
func (dr *DagReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case os.SEEK_SET:
		for i := 0; i < len(dr.node.Links); i++ {
			nsize := dr.node.Links[i].Size - 8
			if offset > nsize {
				offset -= nsize
			} else {
				break
			}
		}
		dr.position = i
		err := dr.precalcNextBuf()
		if err != nil {
			return 0, err
		}
	case os.SEEK_CUR:
	case os.SEEK_END:
	default:
		return 0, errors.New("invalid whence")
	}
	return 0, nil
}
*/
