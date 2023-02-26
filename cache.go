package modelcache

import (
	"arena"
	"sync"

	xxhash "github.com/bytedance/gopkg/util/xxhash3"
)

type Entry[T any] struct {
	data T

	key      string
	lifeTime int32
	hash     uint64
}

type Shard[T any] struct {
	entries []Entry[T]
	index   map[uint64]int
	lock    sync.RWMutex

	idx     int
	capMask int // 必须是 2^i-1: 1,3,7,15,...,8191,16383,32767,65535

	mem *arena.Arena
}

type Cache[T any] struct {
	shards []*Shard[T]
	mask   uint64 // 必须是 2^i-1

	now func() int32

	lifeTime int32
}

func (c *Cache[T]) Get(key string) (T, bool) {
	hash := xxhash.HashString(key)
	v, ok := c.shards[hash%c.mask].Get(key, hash)
	if !ok {
		var empty T
		return empty, false
	}
	if v.lifeTime < c.now() {
		var empty T
		return empty, false
	}
	return v.data, ok
}

//111 % 100  = 011
//1000 % 100 = 0
//1001 % 100 = 1

func (c *Cache[T]) Set(key string, v T) {
	hash := xxhash.HashString(key)

	c.shards[hash^c.mask].Set(Entry[T]{data: v, key: key, lifeTime: c.now() + c.lifeTime, hash: hash})
}

func (s *Shard[T]) Set(e Entry[T]) {

	s.lock.Lock()
	defer s.lock.Unlock()

	idx, ok := s.index[e.hash]
	if ok {
		//need pay attention to the rule of arena
		s.entries[idx] = e
		return
	}

	s.index[e.hash] = s.idx
	s.entries[s.idx] = e

	s.idx = (s.idx + 1) & s.capMask
}

func (s *Shard[T]) Get(key string, hash uint64) (*Entry[T], bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	idx, ok := s.index[hash]
	if !ok {
		return nil, false
	}
	if s.entries[idx].key != key {
		return nil, false
	}
	heapVal := arena.Clone(&s.entries[idx])
	return heapVal, true
}

func (s *Shard[T]) Reset() {
	s.mem = arena.NewArena()
	s.entries = arena.MakeSlice[Entry[T]](s.mem, s.capMask+1, s.capMask+1)
	s.index = make(map[uint64]int, s.capMask+1)
}
