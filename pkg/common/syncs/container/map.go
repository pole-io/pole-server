package container

import (
	"maps"
	"sync"
)

func NewSegmentMap[K comparable, V any](soltNum int, hashFunc func(k K) int) *SegmentMap[K, V] {
	locks := make([]*sync.RWMutex, 0, soltNum)
	solts := make([]map[K]V, 0, soltNum)
	for i := 0; i < int(soltNum); i++ {
		locks = append(locks, &sync.RWMutex{})
		solts = append(solts, map[K]V{})
	}
	return &SegmentMap[K, V]{
		soltNum:  soltNum,
		locks:    locks,
		solts:    solts,
		hashFunc: hashFunc,
	}
}

type SegmentMap[K comparable, V any] struct {
	soltNum  int
	locks    []*sync.RWMutex
	solts    []map[K]V
	hashFunc func(k K) int
}

func (s *SegmentMap[K, V]) Put(k K, v V) {
	lock, solt := s.caulIndex(k)
	lock.Lock()
	defer lock.Unlock()
	solt[k] = v
}

func (s *SegmentMap[K, V]) ComputeIfAbsent(k K, supplier func(k K) V) (V, bool) {
	lock, solt := s.caulIndex(k)
	lock.Lock()
	defer lock.Unlock()
	oldVal, ok := solt[k]
	if !ok {
		v := supplier(k)
		solt[k] = v
		return v, true
	}
	return oldVal, false
}

func (s *SegmentMap[K, V]) PutIfAbsent(k K, v V) (V, bool) {
	lock, solt := s.caulIndex(k)
	lock.Lock()
	defer lock.Unlock()
	oldVal, ok := solt[k]
	if !ok {
		solt[k] = v
		return oldVal, true
	}
	return oldVal, false
}

func (s *SegmentMap[K, V]) Get(k K) (V, bool) {
	lock, solt := s.caulIndex(k)
	lock.RLock()
	defer lock.RUnlock()

	v, ok := solt[k]
	return v, ok
}

func (s *SegmentMap[K, V]) Del(k K) bool {
	lock, solt := s.caulIndex(k)
	lock.Lock()
	defer lock.Unlock()

	_, ok := solt[k]
	delete(solt, k)
	return ok
}

func (s *SegmentMap[K, V]) Range(f func(k K, v V)) {
	for i := 0; i < s.soltNum; i++ {
		lock := s.locks[i]
		solt := s.solts[i]
		func() {
			lock.RLock()
			defer lock.RUnlock()
			for k, v := range solt {
				f(k, v)
			}
		}()
	}
}

func (s *SegmentMap[K, V]) Count() uint64 {
	count := uint64(0)
	for i := 0; i < s.soltNum; i++ {
		lock := s.locks[i]
		solt := s.solts[i]
		func() {
			lock.RLock()
			defer lock.RUnlock()
			count += uint64(len(solt))
		}()
	}
	return count
}

func (s *SegmentMap[K, V]) caulIndex(k K) (*sync.RWMutex, map[K]V) {
	index := s.hashFunc(k) % s.soltNum
	lock := s.locks[index]
	solt := s.solts[index]
	return lock, solt
}

// NewSyncMap
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		m: make(map[K]V, 16),
	}
}

// SyncMap
type SyncMap[K comparable, V any] struct {
	lock sync.RWMutex
	m    map[K]V
}

// ComputeIfAbsent
func (s *SyncMap[K, V]) ComputeIfAbsent(k K, supplier func(k K) V) (V, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	actual, exist := s.m[k]
	if exist {
		return actual, false
	}
	val := supplier(k)
	s.m[k] = val
	return val, true
}

// Load
func (s *SyncMap[K, V]) Load(key K) (V, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	v, ok := s.m[key]
	if ok {
		return v, ok
	}
	var empty V
	return empty, false
}

// Store
func (s *SyncMap[K, V]) Store(key K, val V) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.m[key] = val
}

// Values
func (s *SyncMap[K, V]) Values() []V {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ret := make([]V, 0, len(s.m))
	for _, v := range s.m {
		ret = append(ret, v)
	}
	return ret
}

// Range
func (s *SyncMap[K, V]) Range(f func(key K, val V)) {
	s.lock.RLock()
	snapshot := map[K]V{}
	for k, v := range s.m {
		snapshot[k] = v
	}
	s.lock.RUnlock()

	for k, v := range snapshot {
		f(k, v)
	}
}

// ReadRange .
func (s *SyncMap[K, V]) ReadRange(f func(key K, val V)) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for k, v := range s.m {
		f(k, v)
	}
}

// Delete
func (s *SyncMap[K, V]) Delete(key K) (V, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	v, exist := s.m[key]
	delete(s.m, key)
	return v, exist
}

// Len
func (s *SyncMap[K, V]) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.m)
}

func (s *SyncMap[K, V]) ToMap() map[K]V {
	s.lock.RLock()
	defer s.lock.RUnlock()

	m := map[K]V{}
	maps.Copy(m, s.m)
	return m
}
