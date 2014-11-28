package interplanetary

import (
	mh "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-multihash"
)

type Key interface {
	String() string
}

// mhKey is a key backed by a Multihash
type mhKey struct {
	mh mh.Multihash
}

func parseKey(maybe string) (Key, error) {
	h, err := mh.FromB58String(maybe)
	if err != nil {
		return nil, err
	}
	return &mhKey{mh: h}, nil
}

func (k *mhKey) String() string {
	return k.mh.B58String()
}
