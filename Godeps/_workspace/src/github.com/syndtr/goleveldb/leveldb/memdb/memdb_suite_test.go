package memdb

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/testutil"
)

func TestMemdb(t *testing.T) {
	testutil.RunDefer()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Memdb Suite")
}
