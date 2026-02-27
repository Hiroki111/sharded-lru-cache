package shard

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	hashRing   *HashRing
	aof        *os.File
	writer     *bufio.Writer
	mu         sync.RWMutex
}

func NewCacheManager[K comparable, V any](shardCount int, shardCapacity int, shardReplica int, aofPath string) *CacheManager[K, V] {
	var f *os.File
	var w *bufio.Writer
	if aofPath != "" {
		f, _ = os.OpenFile(aofPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		w = bufio.NewWriter(f)
	}

	m := &CacheManager[K, V]{
		shardCount: uint32(shardCount),
		shards:     make([]*Shard[K, V], shardCount),
		stopChan:   make(chan struct{}),
		hashRing:   NewHashRing(shardCount, shardReplica),
		aof:        f,
		writer:     w,
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
	shard.cache.Set(key, value, ttl)
	shard.mu.Unlock()

	if m.writer != nil {
		kBuf, _ := json.Marshal(key)
		vBuf, _ := json.Marshal(value)
		expiry := time.Now().Add(ttl).Unix()

		m.mu.Lock()
		// Format: SET|base64_key|base64_val|expiry
		fmt.Fprintf(m.writer, "SET|%s|%s|%d\n", kBuf, vBuf, expiry)
		m.mu.Unlock()
	}
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

func (m *CacheManager[K, V]) StartAofSyncer() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if m.writer != nil {
					m.mu.Lock()
					m.writer.Flush()
					m.aof.Sync()
					m.mu.Unlock()
				}
			case <-m.stopChan:
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
	close(m.stopChan)

	if m.aof != nil {
		m.aof.Sync()
		m.aof.Close()
	}
}

func (m *CacheManager[K, V]) LoadAOF() error {
	if m.aof == nil {
		return nil
	}
	// Seek to the beginning of the file
	m.aof.Seek(0, 0)
	scanner := bufio.NewScanner(m.aof)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "|")
		if len(parts) != 4 {
			continue
		}

		var k K
		var v V
		json.Unmarshal([]byte(parts[1]), &k)
		json.Unmarshal([]byte(parts[2]), &v)

		expiry, _ := strconv.ParseInt(parts[3], 10, 64)
		remaining := time.Unix(expiry, 0).Sub(time.Now())

		if remaining > 0 {
			m.setInternal(k, v, remaining)
		}
	}
	return nil
}

func (m *CacheManager[K, V]) setInternal(key K, value V, ttl time.Duration) {
	shard := m.getShard(key)

	shard.mu.Lock()
	shard.cache.Set(key, value, ttl)
	shard.mu.Unlock()
}

func (m *CacheManager[K, V]) getShard(key K) *Shard[K, V] {
	var shardIndex int

	switch v := any(key).(type) {
	case string:
		shardIndex = m.hashRing.GetShardIndex(v)
	case int:
		shardIndex = m.hashRing.GetShardIndex(strconv.Itoa(v))
	default:
		shardIndex = m.hashRing.GetShardIndex(fmt.Sprintf("%v", v))
	}

	return m.shards[shardIndex]
}

func (m *CacheManager[K, V]) cleanup() {
	for _, shard := range m.shards {
		shard.mu.Lock()
		shard.cache.DeleteExpired()
		shard.mu.Unlock()
	}
}
