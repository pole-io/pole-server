package container

import (
	"sync"
)

func NewSegmentMap[K comparable, V any](soltNum int, hashFunc func(k K) int) *SegmentMap[K, V] {
	solts := make([]*SyncMap[K, V], 0, soltNum)
	for i := 0; i < int(soltNum); i++ {
		solts = append(solts, NewSyncMap[K, V]())
	}
	return &SegmentMap[K, V]{
		soltNum:  soltNum,
		solts:    solts,
		hashFunc: hashFunc,
	}
}

type SegmentMap[K comparable, V any] struct {
	soltNum  int
	solts    []*SyncMap[K, V]
	hashFunc func(k K) int
}

func (s *SegmentMap[K, V]) Put(k K, v V) {
	solt := s.caulIndex(k)
	solt.Store(k, v)
}

func (s *SegmentMap[K, V]) PutIfAbsent(k K, v V) (V, bool) {
	solt := s.caulIndex(k)
	oldVal, ok := solt.Load(k)
	if !ok {
		solt.Store(k, v)
		return oldVal, true
	}
	return oldVal, false
}

func (s *SegmentMap[K, V]) Get(k K) (V, bool) {
	solt := s.caulIndex(k)
	v, ok := solt.Load(k)
	return v, ok
}

func (s *SegmentMap[K, V]) Del(k K) bool {
	solt := s.caulIndex(k)
	_, ok := solt.m.LoadAndDelete(k)
	return ok
}

func (s *SegmentMap[K, V]) Range(f func(k K, v V)) {
	for i := 0; i < s.soltNum; i++ {
		solt := s.solts[i]
		solt.Range(f)
	}
}

func (s *SegmentMap[K, V]) Count() uint64 {
	count := uint64(0)
	for i := 0; i < s.soltNum; i++ {
		solt := s.solts[i]
		count += uint64(solt.Len())
	}
	return count
}

func (s *SegmentMap[K, V]) caulIndex(k K) *SyncMap[K, V] {
	index := s.hashFunc(k) % s.soltNum
	solt := s.solts[index]
	return solt
}

// NewSyncMap
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		m: new(sync.Map),
	}
}

// SyncMap
type SyncMap[K comparable, V any] struct {
	m *sync.Map
}

// ComputeIfAbsent
func (s *SyncMap[K, V]) ComputeIfAbsent(k K, supplier func(k K) V) (V, bool) {
	actual, exist := s.m.Load(k)
	if exist {
		return actual.(V), false
	}
	val := supplier(k)
	s.m.Store(k, val)
	return val, true
}

// Load
func (s *SyncMap[K, V]) Load(key K) (V, bool) {
	v, ok := s.m.Load(key)
	if ok {
		return v.(V), ok
	}
	var empty V
	return empty, false
}

// Load
func (s *SyncMap[K, V]) MustLoad(key K) V {
	v, _ := s.m.Load(key)
	return v.(V)
}

// Store
func (s *SyncMap[K, V]) Store(key K, val V) {
	s.m.Store(key, val)
}

// Values
func (s *SyncMap[K, V]) Values() []V {
	ret := make([]V, 0, s.Len())
	s.m.Range(func(key any, val any) bool {
		ret = append(ret, val.(V))
		return true
	})
	return ret
}

// Range
func (s *SyncMap[K, V]) Range(f func(key K, val V)) {
	s.m.Range(func(key any, val any) bool {
		f(key.(K), val.(V))
		return true
	})
}

// ReadRange .
func (s *SyncMap[K, V]) ReadRange(f func(key K, val V)) {
	s.m.Range(func(key any, val any) bool {
		f(key.(K), val.(V))
		return true
	})
}

// Delete
func (s *SyncMap[K, V]) Delete(key K) (V, bool) {
	oldVal, exist := s.m.LoadAndDelete(key)
	if !exist {
		var empty V
		return empty, false
	}
	return oldVal.(V), exist
}

// Len
func (s *SyncMap[K, V]) Len() int {
	length := 0
	s.m.Range(func(key any, val any) bool {
		length++
		return true
	})
	return length
}

func (s *SyncMap[K, V]) ToMap() map[K]V {
	m := map[K]V{}
	s.m.Range(func(key any, val any) bool {
		m[key.(K)] = val.(V)
		return true
	})
	return m
}
