package shard

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

const ttl = 5 * time.Second

func BenchmarkShardedCache_Parallel(b *testing.B) {
	numOfShards := 32
	cache := NewCacheManager[string, int](numOfShards, 1000)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i)
			cache.Set(key, i, ttl)
			cache.Get(key)
			i++
		}
	})
}

func TestCacheManager_ConcurrentSet(t *testing.T) {
	cache := NewCacheManager[string, int](32, 100)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		cache.Set("a", 1, ttl)
	}()

	go func() {
		defer wg.Done()
		cache.Set("a", 2, ttl)
	}()

	wg.Wait()

	if _, ok := cache.Get("a"); !ok {
		t.Errorf("Expected ok")
	}
}
