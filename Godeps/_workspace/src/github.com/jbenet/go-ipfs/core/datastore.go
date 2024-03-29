package core

import (
	ds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-datastore"
	fsds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-datastore/fs"
	ktds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-datastore/keytransform"
	lds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-datastore/leveldb"
	syncds "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-datastore/sync"
	ldbopts "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/opt"

	config "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/config"
	u "github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util"
	"github.com/maybebtc/interplanetary/Godeps/_workspace/src/github.com/jbenet/go-ipfs/util/debugerror"
)

func makeDatastore(cfg config.Datastore) (u.ThreadSafeDatastoreCloser, error) {
	if len(cfg.Type) == 0 {
		return nil, debugerror.Errorf("config datastore.type required")
	}

	switch cfg.Type {
	case "leveldb":
		return makeLevelDBDatastore(cfg)

	case "memory":
		return u.CloserWrap(syncds.MutexWrap(ds.NewMapDatastore())), nil

	case "fs":
		log.Warning("using fs.Datastore at .datastore for testing.")
		d, err := fsds.NewDatastore(".datastore") // for testing!!
		if err != nil {
			return nil, err
		}
		ktd := ktds.Wrap(d, u.B58KeyConverter)
		return u.CloserWrap(syncds.MutexWrap(ktd)), nil
	}

	return nil, debugerror.Errorf("Unknown datastore type: %s", cfg.Type)
}

func makeLevelDBDatastore(cfg config.Datastore) (u.ThreadSafeDatastoreCloser, error) {
	if len(cfg.Path) == 0 {
		return nil, debugerror.Errorf("config datastore.path required for leveldb")
	}

	ds, err := lds.NewDatastore(cfg.Path, &lds.Options{
		// TODO don't import ldbopts. Get from go-datastore.leveldb
		Compression: ldbopts.NoCompression,
	})
	return ds, debugerror.Wrap(err)
}
