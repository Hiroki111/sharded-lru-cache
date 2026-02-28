package shard

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

const ttl = 5 * time.Second

func BenchmarkShardedCache_Parallel(b *testing.B) {
	numOfShards := 32
	cache := NewCacheManager[string, int](numOfShards, 1000, 3, "")

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
	cache := NewCacheManager[string, int](32, 100, 3, "")

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

func TestCacheManager_Set(t *testing.T) {
	type User struct {
		id   uint
		name string
	}
	users := []User{{id: 1, name: "Alice"}, {id: 2, name: "Bob"}, {id: 3, name: "Carol"}}
	mapNameToId := map[string]int{"Alice": 1, "Bob": 2, "Carol": 3}
	intChan := make(chan int, 5)

	tests := []struct {
		testName string
		value    any
	}{
		{testName: "map", value: mapNameToId},
		{testName: "slice", value: users},
		{testName: "pointer", value: &users},
		{testName: "float64", value: float64(0.123)},
		{testName: "channel", value: intChan},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mgr := NewCacheManager[string, any](1, 1, 1, "")
			key := "k"
			ttl := 60 * time.Second
			mgr.Set(key, test.value, ttl)

			value, found := mgr.Get(key)
			if !found {
				t.Fatalf("expected key to be found")
			}

			if !reflect.DeepEqual(value, test.value) {
				t.Fatalf("expected %v, got %v", test.value, value)
			}
		})
	}
}
