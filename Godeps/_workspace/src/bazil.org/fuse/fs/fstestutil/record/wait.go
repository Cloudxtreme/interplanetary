package record

import (
	"sync"
	"time"

	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/bazil.org/fuse"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/bazil.org/fuse/fs"
)

type nothing struct{}

// ReleaseWaiter notes whether a FUSE Release call has been seen.
//
// Releases are not guaranteed to happen synchronously with any client
// call, so they must be waited for.
type ReleaseWaiter struct {
	once sync.Once
	seen chan nothing
}

var _ = fs.HandleReleaser(&ReleaseWaiter{})

func (r *ReleaseWaiter) init() {
	r.once.Do(func() {
		r.seen = make(chan nothing, 1)
	})
}

func (r *ReleaseWaiter) Release(req *fuse.ReleaseRequest, intr fs.Intr) fuse.Error {
	r.init()
	close(r.seen)
	return nil
}

// WaitForRelease waits for Release to be called.
//
// With zero duration, wait forever. Otherwise, timeout early
// in a more controller way than `-test.timeout`.
//
// Returns whether a Release was seen. Always true if dur==0.
func (r *ReleaseWaiter) WaitForRelease(dur time.Duration) bool {
	r.init()
	var timeout <-chan time.Time
	if dur > 0 {
		timeout = time.After(dur)
	}
	select {
	case <-r.seen:
		return true
	case <-timeout:
		return false
	}
}
