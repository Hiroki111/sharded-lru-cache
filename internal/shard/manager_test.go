package shard

import (
	"fmt"
	"testing"
)

func BenchmarkShardedCache_Parallel(b *testing.B) {
	numOfShards := 32
	cache := NewCacheManager[string, int](numOfShards, 1000)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i)
			cache.Set(key, i)
			cache.Get(key)
			i++
		}
	})
}
