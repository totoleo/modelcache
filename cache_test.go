package modelcache

import (
	"arena"
	"testing"

	"github.com/bytedance/gopkg/util/xxhash3"
)

func TestShardGet(t *testing.T) {
	type Foo struct {
		Name string
		Id   int64
	}
	s := newShard[Foo](8)
	key := "1"
	s.Set(Entry[Foo]{data: Foo{Name: "test foo"}, key: key, hash: xxhash3.HashString(key), lifeTime: 1})

	e, ok := s.Get(key, xxhash3.HashString(key))
	if !ok {
		t.Error("Expected entry to be found.")
		t.Log(s.index, "\n", s.entries)
	}
	t.Log(e)

	key = "2"
	s.Set(Entry[Foo]{data: Foo{Name: "test foo"}, key: key, hash: xxhash3.HashString(key)})

	t.Log(s.index, "\n", s.entries)
}

func newShard[T any](capPower int) *Shard[T] {
	mem := arena.NewArena()

	s := &Shard[T]{
		capMask: 1<<capPower - 1,
		mem:     mem,
	}
	s.Reset()
	return s
}
