package shard

import (
	"hash/fnv"
	"sync"

	"github.com/Hiroki111/sharded-lru-cache/internal/lru"
)

type Shard[K comparable, V any] struct {
	mu    sync.RWMutex
	cache *lru.LRU[K, V]
}

type CacheManager[K comparable, V any] struct {
	shardCount uint32
	shads      []*Shard[K, V]
}

func NewCacheManager[K comparable, V any](shardCount int, shardCapacity int) *CacheManager[K, V] {
	m := &CacheManager[K, V]{
		shardCount: uint32(shardCount),
		shads:      make([]*Shard[K, V], shardCount),
	}

	for i := 0; i < shardCount; i++ {
		m.shads[i] = &Shard[K, V]{
			cache: lru.NewLRUCache[K, V](shardCapacity),
		}
	}
	return m
}

func (m *CacheManager[K, V]) Get(key K) (V, bool) {
	shard := m.getShard(any(key).(string))

	shard.mu.Lock()
	defer shard.mu.Unlock()
	return shard.cache.Get(key)
}

func (m *CacheManager[K, V]) Set(key K, value V) {
	shard := m.getShard(any(key).(string))

	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.cache.Set(key, value)
}

func (m *CacheManager[K, V]) getShard(key string) *Shard[K, V] {
	h := fnv.New32a()
	h.Write([]byte(key))
	hash := h.Sum32()
	return m.shads[hash%m.shardCount]
}
