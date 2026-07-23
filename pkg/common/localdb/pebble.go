package localdb

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/cockroachdb/pebble"
)

type PebbleDB struct {
	path string
	db   *pebble.DB
}

type WriteMode int

const (
	WriteNoSync WriteMode = iota
	WriteSync
)

var pebbleRegistry = struct {
	sync.Mutex
	dbs map[string]*sharedPebbleDB
}{
	dbs: map[string]*sharedPebbleDB{},
}

type sharedPebbleDB struct {
	db   *pebble.DB
	refs int
}

func OpenPebble(path string) (*PebbleDB, error) {
	if path == "" {
		return nil, errors.New("pebble path is empty")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absPath, 0o755); err != nil {
		return nil, err
	}

	pebbleRegistry.Lock()
	defer pebbleRegistry.Unlock()

	if shared := pebbleRegistry.dbs[absPath]; shared != nil {
		shared.refs++
		return &PebbleDB{path: absPath, db: shared.db}, nil
	}

	db, err := pebble.Open(absPath, &pebble.Options{})
	if err != nil {
		return nil, err
	}
	pebbleRegistry.dbs[absPath] = &sharedPebbleDB{db: db, refs: 1}
	return &PebbleDB{path: absPath, db: db}, nil
}

func (p *PebbleDB) Close() error {
	if p == nil || p.path == "" {
		return nil
	}

	pebbleRegistry.Lock()
	defer pebbleRegistry.Unlock()

	shared := pebbleRegistry.dbs[p.path]
	if shared == nil {
		return nil
	}
	shared.refs--
	if shared.refs > 0 {
		return nil
	}
	delete(pebbleRegistry.dbs, p.path)
	return shared.db.Close()
}

func (p *PebbleDB) Set(key, value []byte, mode WriteMode) error {
	if p == nil || p.db == nil {
		return errors.New("pebble db is nil")
	}
	return p.db.Set(key, value, writeOptions(mode))
}

func (p *PebbleDB) Get(key []byte) ([]byte, bool, error) {
	if p == nil || p.db == nil {
		return nil, false, errors.New("pebble db is nil")
	}
	value, closer, err := p.db.Get(key)
	if errors.Is(err, pebble.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	defer closer.Close()
	return append([]byte(nil), value...), true, nil
}

func (p *PebbleDB) Delete(key []byte, mode WriteMode) error {
	if p == nil || p.db == nil {
		return errors.New("pebble db is nil")
	}
	err := p.db.Delete(key, writeOptions(mode))
	if errors.Is(err, pebble.ErrNotFound) {
		return nil
	}
	return err
}

func (p *PebbleDB) ScanPrefix(prefix []byte, limit int, fn func(key, value []byte) error) error {
	if p == nil || p.db == nil {
		return errors.New("pebble db is nil")
	}
	if limit <= 0 {
		return nil
	}
	iter, err := p.db.NewIter(&pebble.IterOptions{
		LowerBound: append([]byte(nil), prefix...),
		UpperBound: prefixUpperBound(prefix),
	})
	if err != nil {
		return err
	}
	defer iter.Close()

	count := 0
	for valid := iter.First(); valid && count < limit; valid = iter.Next() {
		key := iter.Key()
		if !bytes.HasPrefix(key, prefix) {
			break
		}
		value, err := iter.ValueAndErr()
		if err != nil {
			return err
		}
		if err := fn(append([]byte(nil), key...), append([]byte(nil), value...)); err != nil {
			return err
		}
		count++
	}
	return iter.Error()
}

func writeOptions(mode WriteMode) *pebble.WriteOptions {
	if mode == WriteSync {
		return pebble.Sync
	}
	return pebble.NoSync
}

func prefixUpperBound(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}
	upper := append([]byte(nil), prefix...)
	for i := len(upper) - 1; i >= 0; i-- {
		if upper[i] != 0xff {
			upper[i]++
			return upper[:i+1]
		}
	}
	return nil
}
