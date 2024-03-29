// package set contains various different types of 'BlockSet's
package set

import (
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/blocks/bloom"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"
)

var log = util.Logger("blockset")

// BlockSet represents a mutable set of keyed blocks
type BlockSet interface {
	AddBlock(util.Key)
	RemoveBlock(util.Key)
	HasKey(util.Key) bool
	GetBloomFilter() bloom.Filter

	GetKeys() []util.Key
}

func SimpleSetFromKeys(keys []util.Key) BlockSet {
	sbs := &simpleBlockSet{blocks: make(map[util.Key]struct{})}
	for _, k := range keys {
		sbs.blocks[k] = struct{}{}
	}
	return sbs
}

func NewSimpleBlockSet() BlockSet {
	return &simpleBlockSet{blocks: make(map[util.Key]struct{})}
}

type simpleBlockSet struct {
	blocks map[util.Key]struct{}
}

func (b *simpleBlockSet) AddBlock(k util.Key) {
	b.blocks[k] = struct{}{}
}

func (b *simpleBlockSet) RemoveBlock(k util.Key) {
	delete(b.blocks, k)
}

func (b *simpleBlockSet) HasKey(k util.Key) bool {
	_, has := b.blocks[k]
	return has
}

func (b *simpleBlockSet) GetBloomFilter() bloom.Filter {
	f := bloom.BasicFilter()
	for k := range b.blocks {
		f.Add([]byte(k))
	}
	return f
}

func (b *simpleBlockSet) GetKeys() []util.Key {
	var out []util.Key
	for k := range b.blocks {
		out = append(out, k)
	}
	return out
}
