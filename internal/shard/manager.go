package shard

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/Hiroki111/sharded-lru-cache/internal/lru"
)

type Shard[K comparable, V any] struct {
	mu    sync.RWMutex
	cache *lru.LRU[K, V]
}

type CacheManager[K comparable, V any] struct {
	shardCount uint32
	shards     []*Shard[K, V]
	stopChan   chan struct{}
}

func NewCacheManager[K comparable, V any](shardCount int, shardCapacity int) *CacheManager[K, V] {
	m := &CacheManager[K, V]{
		shardCount: uint32(shardCount),
		shards:     make([]*Shard[K, V], shardCount),
		stopChan:   make(chan struct{}),
	}

	for i := 0; i < shardCount; i++ {
		m.shards[i] = &Shard[K, V]{
			cache: lru.NewLRUCache[K, V](shardCapacity),
		}
	}
	return m
}

func (m *CacheManager[K, V]) Get(key K) (V, bool) {
	shard := m.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()
	return shard.cache.Get(key)
}

func (m *CacheManager[K, V]) Set(key K, value V, ttl time.Duration) {
	shard := m.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.cache.Set(key, value, ttl)
}

func (m *CacheManager[K, V]) StartJanitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.cleanup()
			case <-m.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (m *CacheManager[K, V]) GetStats() lru.Stats {
	var total lru.Stats
	for _, shard := range m.shards {
		shard.mu.RLock()
		total.Hits += shard.cache.Stats().Hits
		total.Misses += shard.cache.Stats().Misses
		total.Evictions += shard.cache.Stats().Evictions
		shard.mu.RUnlock()
	}
	return total
}

func (m *CacheManager[K, V]) Stop() {
	m.stopChan <- struct{}{}
	close(m.stopChan)
}

func (m *CacheManager[K, V]) getShard(key K) *Shard[K, V] {
	h := fnv.New32a()

	switch v := any(key).(type) {
	case string:
		h.Write([]byte(v))
	case int:
		binary.Write(h, binary.LittleEndian, int64(v))
	default:
		h.Write([]byte(fmt.Sprintf("%v", v)))
	}

	hash := h.Sum32()
	return m.shards[hash%m.shardCount]
}

func (m *CacheManager[K, V]) cleanup() {
	for _, shard := range m.shards {
		shard.mu.Lock()
		shard.cache.DeleteExpired()
		shard.mu.Unlock()
	}
}
