// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package leveldb

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/journal"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/memdb"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/table"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/util"
)

// DB is a LevelDB database.
type DB struct {
	// Need 64-bit alignment.
	seq uint64

	// Session.
	s *session

	// MemDB.
	memMu             sync.RWMutex
	memPool           chan *memdb.DB
	mem, frozenMem    *memDB
	journal           *journal.Writer
	journalWriter     storage.Writer
	journalFile       storage.File
	frozenJournalFile storage.File
	frozenSeq         uint64

	// Snapshot.
	snapsMu   sync.Mutex
	snapsRoot snapshotElement

	// Stats.
	aliveSnaps, aliveIters int32

	// Write.
	writeC       chan *Batch
	writeMergedC chan bool
	writeLockC   chan struct{}
	writeAckC    chan error
	journalC     chan *Batch
	journalAckC  chan error

	// Compaction.
	tcompCmdC     chan cCmd
	tcompPauseC   chan chan<- struct{}
	tcompTriggerC chan struct{}
	mcompCmdC     chan cCmd
	mcompTriggerC chan struct{}
	compErrC      chan error
	compErrSetC   chan error
	compStats     [kNumLevels]cStats

	// Close.
	closeW sync.WaitGroup
	closeC chan struct{}
	closed uint32
	closer io.Closer
}

func openDB(s *session) (*DB, error) {
	s.log("db@open opening")
	start := time.Now()
	db := &DB{
		s: s,
		// Initial sequence
		seq: s.stSeq,
		// MemDB
		memPool: make(chan *memdb.DB, 1),
		// Write
		writeC:       make(chan *Batch),
		writeMergedC: make(chan bool),
		writeLockC:   make(chan struct{}, 1),
		writeAckC:    make(chan error),
		journalC:     make(chan *Batch),
		journalAckC:  make(chan error),
		// Compaction
		tcompCmdC:     make(chan cCmd),
		tcompPauseC:   make(chan chan<- struct{}),
		tcompTriggerC: make(chan struct{}, 1),
		mcompCmdC:     make(chan cCmd),
		mcompTriggerC: make(chan struct{}, 1),
		compErrC:      make(chan error),
		compErrSetC:   make(chan error),
		// Close
		closeC: make(chan struct{}),
	}
	db.initSnapshot()

	if err := db.recoverJournal(); err != nil {
		return nil, err
	}

	// Remove any obsolete files.
	if err := db.checkAndCleanFiles(); err != nil {
		// Close journal.
		if db.journal != nil {
			db.journal.Close()
			db.journalWriter.Close()
		}
		return nil, err
	}

	// Don't include compaction error goroutine into wait group.
	go db.compactionError()

	db.closeW.Add(3)
	go db.tCompaction()
	go db.mCompaction()
	go db.jWriter()
	go db.mpoolDrain()

	s.logf("db@open done T·%v", time.Since(start))

	runtime.SetFinalizer(db, (*DB).Close)
	return db, nil
}

// Open opens or creates a DB for the given storage.
// The DB will be created if not exist, unless ErrorIfMissing is true.
// Also, if ErrorIfExist is true and the DB exist Open will returns
// os.ErrExist error.
//
// Open will return an error with type of ErrCorrupted if corruption
// detected in the DB. Corrupted DB can be recovered with Recover
// function.
//
// The returned DB instance is goroutine-safe.
// The DB must be closed after use, by calling Close method.
func Open(stor storage.Storage, o *opt.Options) (db *DB, err error) {
	s, err := newSession(stor, o)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			s.close()
			s.release()
		}
	}()

	err = s.recover()
	if err != nil {
		if !os.IsNotExist(err) || s.o.GetErrorIfMissing() {
			return
		}
		err = s.create()
		if err != nil {
			return
		}
	} else if s.o.GetErrorIfExist() {
		err = os.ErrExist
		return
	}

	return openDB(s)
}

// OpenFile opens or creates a DB for the given path.
// The DB will be created if not exist, unless ErrorIfMissing is true.
// Also, if ErrorIfExist is true and the DB exist OpenFile will returns
// os.ErrExist error.
//
// OpenFile uses standard file-system backed storage implementation as
// desribed in the leveldb/storage package.
//
// OpenFile will return an error with type of ErrCorrupted if corruption
// detected in the DB. Corrupted DB can be recovered with Recover
// function.
//
// The returned DB instance is goroutine-safe.
// The DB must be closed after use, by calling Close method.
func OpenFile(path string, o *opt.Options) (db *DB, err error) {
	stor, err := storage.OpenFile(path)
	if err != nil {
		return
	}
	db, err = Open(stor, o)
	if err != nil {
		stor.Close()
	} else {
		db.closer = stor
	}
	return
}

// Recover recovers and opens a DB with missing or corrupted manifest files
// for the given storage. It will ignore any manifest files, valid or not.
// The DB must already exist or it will returns an error.
// Also, Recover will ignore ErrorIfMissing and ErrorIfExist options.
//
// The returned DB instance is goroutine-safe.
// The DB must be closed after use, by calling Close method.
func Recover(stor storage.Storage, o *opt.Options) (db *DB, err error) {
	s, err := newSession(stor, o)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			s.close()
			s.release()
		}
	}()

	err = recoverTable(s, o)
	if err != nil {
		return
	}
	return openDB(s)
}

// RecoverFile recovers and opens a DB with missing or corrupted manifest files
// for the given path. It will ignore any manifest files, valid or not.
// The DB must already exist or it will returns an error.
// Also, Recover will ignore ErrorIfMissing and ErrorIfExist options.
//
// RecoverFile uses standard file-system backed storage implementation as desribed
// in the leveldb/storage package.
//
// The returned DB instance is goroutine-safe.
// The DB must be closed after use, by calling Close method.
func RecoverFile(path string, o *opt.Options) (db *DB, err error) {
	stor, err := storage.OpenFile(path)
	if err != nil {
		return
	}
	db, err = Recover(stor, o)
	if err != nil {
		stor.Close()
	} else {
		db.closer = stor
	}
	return
}

func recoverTable(s *session, o *opt.Options) error {
	// Get all tables and sort it by file number.
	tableFiles_, err := s.getFiles(storage.TypeTable)
	if err != nil {
		return err
	}
	tableFiles := files(tableFiles_)
	tableFiles.sort()

	var mSeq uint64
	var good, corrupted int
	rec := new(sessionRecord)
	bpool := util.NewBufferPool(o.GetBlockSize() + 5)
	buildTable := func(iter iterator.Iterator) (tmp storage.File, size int64, err error) {
		tmp = s.newTemp()
		writer, err := tmp.Create()
		if err != nil {
			return
		}
		defer func() {
			writer.Close()
			if err != nil {
				tmp.Remove()
				tmp = nil
			}
		}()

		// Copy entries.
		tw := table.NewWriter(writer, o)
		for iter.Next() {
			key := iter.Key()
			if validIkey(key) {
				err = tw.Append(key, iter.Value())
				if err != nil {
					return
				}
			}
		}
		err = iter.Error()
		if err != nil {
			return
		}
		err = tw.Close()
		if err != nil {
			return
		}
		err = writer.Sync()
		if err != nil {
			return
		}
		size = int64(tw.BytesLen())
		return
	}
	recoverTable := func(file storage.File) error {
		s.logf("table@recovery recovering @%d", file.Num())
		reader, err := file.Open()
		if err != nil {
			return err
		}
		defer reader.Close()

		// Get file size.
		size, err := reader.Seek(0, 2)
		if err != nil {
			return err
		}

		var tSeq uint64
		var tgood, tcorrupted, blockerr int
		var imin, imax []byte
		tr := table.NewReader(reader, size, nil, bpool, o)
		iter := tr.NewIterator(nil, nil)
		iter.(iterator.ErrorCallbackSetter).SetErrorCallback(func(err error) {
			s.logf("table@recovery found error @%d %q", file.Num(), err)
			blockerr++
		})

		// Scan the table.
		for iter.Next() {
			key := iter.Key()
			_, seq, _, ok := parseIkey(key)
			if !ok {
				tcorrupted++
				continue
			}
			tgood++
			if seq > tSeq {
				tSeq = seq
			}
			if imin == nil {
				imin = append([]byte{}, key...)
			}
			imax = append(imax[:0], key...)
		}
		if err := iter.Error(); err != nil {
			iter.Release()
			return err
		}
		iter.Release()

		if tgood > 0 {
			if tcorrupted > 0 || blockerr > 0 {
				// Rebuild the table.
				s.logf("table@recovery rebuilding @%d", file.Num())
				iter := tr.NewIterator(nil, nil)
				tmp, newSize, err := buildTable(iter)
				iter.Release()
				if err != nil {
					return err
				}
				reader.Close()
				if err := file.Replace(tmp); err != nil {
					return err
				}
				size = newSize
			}
			if tSeq > mSeq {
				mSeq = tSeq
			}
			// Add table to level 0.
			rec.addTable(0, file.Num(), uint64(size), imin, imax)
			s.logf("table@recovery recovered @%d N·%d C·%d B·%d S·%d Q·%d", file.Num(), tgood, tcorrupted, blockerr, size, tSeq)
		} else {
			s.logf("table@recovery unrecoverable @%d C·%d B·%d S·%d", file.Num(), tcorrupted, blockerr, size)
		}

		good += tgood
		corrupted += tcorrupted

		return nil
	}

	// Recover all tables.
	if len(tableFiles) > 0 {
		s.logf("table@recovery F·%d", len(tableFiles))

		// Mark file number as used.
		s.markFileNum(tableFiles[len(tableFiles)-1].Num())

		for _, file := range tableFiles {
			if err := recoverTable(file); err != nil {
				return err
			}
		}

		s.logf("table@recovery recovered F·%d N·%d C·%d Q·%d", len(tableFiles), good, corrupted, mSeq)
	}

	// Set sequence number.
	rec.setSeq(mSeq + 1)

	// Create new manifest.
	if err := s.create(); err != nil {
		return err
	}

	// Commit.
	return s.commit(rec)
}

func (db *DB) recoverJournal() error {
	// Get all tables and sort it by file number.
	journalFiles_, err := db.s.getFiles(storage.TypeJournal)
	if err != nil {
		return err
	}
	journalFiles := files(journalFiles_)
	journalFiles.sort()

	// Discard older journal.
	prev := -1
	for i, file := range journalFiles {
		if file.Num() >= db.s.stJournalNum {
			if prev >= 0 {
				i--
				journalFiles[i] = journalFiles[prev]
			}
			journalFiles = journalFiles[i:]
			break
		} else if file.Num() == db.s.stPrevJournalNum {
			prev = i
		}
	}

	var jr *journal.Reader
	var of storage.File
	var mem *memdb.DB
	batch := new(Batch)
	cm := newCMem(db.s)
	buf := new(util.Buffer)
	// Options.
	strict := db.s.o.GetStrict(opt.StrictJournal)
	checksum := db.s.o.GetStrict(opt.StrictJournalChecksum)
	writeBuffer := db.s.o.GetWriteBuffer()
	recoverJournal := func(file storage.File) error {
		db.logf("journal@recovery recovering @%d", file.Num())
		reader, err := file.Open()
		if err != nil {
			return err
		}
		defer reader.Close()

		// Create/reset journal reader instance.
		if jr == nil {
			jr = journal.NewReader(reader, dropper{db.s, file}, strict, checksum)
		} else {
			jr.Reset(reader, dropper{db.s, file}, strict, checksum)
		}

		// Flush memdb and remove obsolete journal file.
		if of != nil {
			if mem.Len() > 0 {
				if err := cm.flush(mem, 0); err != nil {
					return err
				}
			}
			if err := cm.commit(file.Num(), db.seq); err != nil {
				return err
			}
			cm.reset()
			of.Remove()
			of = nil
		}

		// Replay journal to memdb.
		mem.Reset()
		for {
			r, err := jr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			buf.Reset()
			if _, err := buf.ReadFrom(r); err != nil {
				if err == io.ErrUnexpectedEOF {
					continue
				} else {
					return err
				}
			}
			if err := batch.decode(buf.Bytes()); err != nil {
				return err
			}
			if err := batch.memReplay(mem); err != nil {
				return err
			}

			// Save sequence number.
			db.seq = batch.seq + uint64(batch.len())

			// Flush it if large enough.
			if mem.Size() >= writeBuffer {
				if err := cm.flush(mem, 0); err != nil {
					return err
				}
				mem.Reset()
			}
		}

		of = file
		return nil
	}

	// Recover all journals.
	if len(journalFiles) > 0 {
		db.logf("journal@recovery F·%d", len(journalFiles))

		// Mark file number as used.
		db.s.markFileNum(journalFiles[len(journalFiles)-1].Num())

		mem = memdb.New(db.s.icmp, writeBuffer)
		for _, file := range journalFiles {
			if err := recoverJournal(file); err != nil {
				return err
			}
		}

		// Flush the last journal.
		if mem.Len() > 0 {
			if err := cm.flush(mem, 0); err != nil {
				return err
			}
		}
	}

	// Create a new journal.
	if _, err := db.newMem(0); err != nil {
		return err
	}

	// Commit.
	if err := cm.commit(db.journalFile.Num(), db.seq); err != nil {
		// Close journal.
		if db.journal != nil {
			db.journal.Close()
			db.journalWriter.Close()
		}
		return err
	}

	// Remove the last obsolete journal file.
	if of != nil {
		of.Remove()
	}

	return nil
}

func (db *DB) get(key []byte, seq uint64, ro *opt.ReadOptions) (value []byte, err error) {
	ikey := newIKey(key, seq, tSeek)

	em, fm := db.getMems()
	for _, m := range [...]*memDB{em, fm} {
		if m == nil {
			continue
		}
		defer m.decref()

		mk, mv, me := m.mdb.Find(ikey)
		if me == nil {
			ukey, _, t, ok := parseIkey(mk)
			if ok && db.s.icmp.uCompare(ukey, key) == 0 {
				if t == tDel {
					return nil, ErrNotFound
				}
				return append([]byte{}, mv...), nil
			}
		} else if me != ErrNotFound {
			return nil, me
		}
	}

	v := db.s.version()
	value, cSched, err := v.get(ikey, ro)
	v.release()
	if cSched {
		// Trigger table compaction.
		db.compTrigger(db.tcompTriggerC)
	}
	return
}

// Get gets the value for the given key. It returns ErrNotFound if the
// DB does not contain the key.
//
// The returned slice is its own copy, it is safe to modify the contents
// of the returned slice.
// It is safe to modify the contents of the argument after Get returns.
func (db *DB) Get(key []byte, ro *opt.ReadOptions) (value []byte, err error) {
	err = db.ok()
	if err != nil {
		return
	}

	return db.get(key, db.getSeq(), ro)
}

// NewIterator returns an iterator for the latest snapshot of the
// uderlying DB.
// The returned iterator is not goroutine-safe, but it is safe to use
// multiple iterators concurrently, with each in a dedicated goroutine.
// It is also safe to use an iterator concurrently with modifying its
// underlying DB. The resultant key/value pairs are guaranteed to be
// consistent.
//
// Slice allows slicing the iterator to only contains keys in the given
// range. A nil Range.Start is treated as a key before all keys in the
// DB. And a nil Range.Limit is treated as a key after all keys in
// the DB.
//
// The iterator must be released after use, by calling Release method.
//
// Also read Iterator documentation of the leveldb/iterator package.
func (db *DB) NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator {
	if err := db.ok(); err != nil {
		return iterator.NewEmptyIterator(err)
	}

	snap := db.newSnapshot()
	defer snap.Release()
	return snap.NewIterator(slice, ro)
}

// GetSnapshot returns a latest snapshot of the underlying DB. A snapshot
// is a frozen snapshot of a DB state at a particular point in time. The
// content of snapshot are guaranteed to be consistent.
//
// The snapshot must be released after use, by calling Release method.
func (db *DB) GetSnapshot() (*Snapshot, error) {
	if err := db.ok(); err != nil {
		return nil, err
	}

	return db.newSnapshot(), nil
}

// GetProperty returns value of the given property name.
//
// Property names:
//	leveldb.num-files-at-level{n}
//		Returns the number of filer at level 'n'.
//	leveldb.stats
//		Returns statistics of the underlying DB.
//	leveldb.sstables
//		Returns sstables list for each level.
//	leveldb.blockpool
//		Returns block pool stats.
//	leveldb.cachedblock
//		Returns size of cached block.
//	leveldb.openedtables
//		Returns number of opened tables.
//	leveldb.alivesnaps
//		Returns number of alive snapshots.
//	leveldb.aliveiters
//		Returns number of alive iterators.
func (db *DB) GetProperty(name string) (value string, err error) {
	err = db.ok()
	if err != nil {
		return
	}

	const prefix = "leveldb."
	if !strings.HasPrefix(name, prefix) {
		return "", errors.New("leveldb: GetProperty: unknown property: " + name)
	}
	p := name[len(prefix):]

	v := db.s.version()
	defer v.release()

	switch {
	case strings.HasPrefix(p, "num-files-at-level"):
		var level uint
		var rest string
		n, _ := fmt.Scanf("%d%s", &level, &rest)
		if n != 1 || level >= kNumLevels {
			err = errors.New("leveldb: GetProperty: invalid property: " + name)
		} else {
			value = fmt.Sprint(v.tLen(int(level)))
		}
	case p == "stats":
		value = "Compactions\n" +
			" Level |   Tables   |    Size(MB)   |    Time(sec)  |    Read(MB)   |   Write(MB)\n" +
			"-------+------------+---------------+---------------+---------------+---------------\n"
		for level, tables := range v.tables {
			duration, read, write := db.compStats[level].get()
			if len(tables) == 0 && duration == 0 {
				continue
			}
			value += fmt.Sprintf(" %3d   | %10d | %13.5f | %13.5f | %13.5f | %13.5f\n",
				level, len(tables), float64(tables.size())/1048576.0, duration.Seconds(),
				float64(read)/1048576.0, float64(write)/1048576.0)
		}
	case p == "sstables":
		for level, tables := range v.tables {
			value += fmt.Sprintf("--- level %d ---\n", level)
			for _, t := range tables {
				value += fmt.Sprintf("%d:%d[%q .. %q]\n", t.file.Num(), t.size, t.imin, t.imax)
			}
		}
	case p == "blockpool":
		value = fmt.Sprintf("%v", db.s.tops.bpool)
	case p == "cachedblock":
		if bc := db.s.o.GetBlockCache(); bc != nil {
			value = fmt.Sprintf("%d", bc.Size())
		} else {
			value = "<nil>"
		}
	case p == "openedtables":
		value = fmt.Sprintf("%d", db.s.tops.cache.Size())
	case p == "alivesnaps":
		value = fmt.Sprintf("%d", atomic.LoadInt32(&db.aliveSnaps))
	case p == "aliveiters":
		value = fmt.Sprintf("%d", atomic.LoadInt32(&db.aliveIters))
	default:
		err = errors.New("leveldb: GetProperty: unknown property: " + name)
	}

	return
}

// SizeOf calculates approximate sizes of the given key ranges.
// The length of the returned sizes are equal with the length of the given
// ranges. The returned sizes measure storage space usage, so if the user
// data compresses by a factor of ten, the returned sizes will be one-tenth
// the size of the corresponding user data size.
// The results may not include the sizes of recently written data.
func (db *DB) SizeOf(ranges []util.Range) (Sizes, error) {
	if err := db.ok(); err != nil {
		return nil, err
	}

	v := db.s.version()
	defer v.release()

	sizes := make(Sizes, 0, len(ranges))
	for _, r := range ranges {
		imin := newIKey(r.Start, kMaxSeq, tSeek)
		imax := newIKey(r.Limit, kMaxSeq, tSeek)
		start, err := v.offsetOf(imin)
		if err != nil {
			return nil, err
		}
		limit, err := v.offsetOf(imax)
		if err != nil {
			return nil, err
		}
		var size uint64
		if limit >= start {
			size = limit - start
		}
		sizes = append(sizes, size)
	}

	return sizes, nil
}

// Close closes the DB. This will also releases any outstanding snapshot and
// abort any in-flight compaction.
//
// It is not safe to close a DB until all outstanding iterators are released.
// It is valid to call Close multiple times. Other methods should not be
// called after the DB has been closed.
func (db *DB) Close() error {
	if !db.setClosed() {
		return ErrClosed
	}

	start := time.Now()
	db.log("db@close closing")

	// Clear the finalizer.
	runtime.SetFinalizer(db, nil)

	// Get compaction error.
	var err error
	select {
	case err = <-db.compErrC:
	default:
	}

	close(db.closeC)

	// Wait for the close WaitGroup.
	db.closeW.Wait()

	// Close journal.
	db.writeLockC <- struct{}{}
	if db.journal != nil {
		db.journal.Close()
		db.journalWriter.Close()
	}

	// Close session.
	db.s.close()
	db.logf("db@close done T·%v", time.Since(start))
	db.s.release()

	if db.closer != nil {
		if err1 := db.closer.Close(); err == nil {
			err = err1
		}
	}

	// NIL'ing pointers.
	db.s = nil
	db.mem = nil
	db.frozenMem = nil
	db.journal = nil
	db.journalWriter = nil
	db.journalFile = nil
	db.frozenJournalFile = nil
	db.snapsRoot = snapshotElement{}
	db.closer = nil

	return err
}
