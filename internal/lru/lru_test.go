package lru

import "testing"

func TestLRU_SetAndGet(t *testing.T) {
	// Initialize a small cache
	cache := NewLRUCache[string, int](2)

	cache.Set("a", 1)
	cache.Set("b", 2)

	// Test successful retrieval
	if val, ok := cache.Get("a"); !ok || val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}

	// Test eviction: Adding "c" should kick out "b"
	// because we accessed "a" last (making "b" the LRU).
	cache.Set("c", 3)

	if _, ok := cache.Get("b"); ok {
		t.Error("Expected 'b' to be evicted")
	}

	if val, ok := cache.Get("c"); !ok || val != 3 {
		t.Errorf("Expected 3, got %v", val)
	}
}

func TestLRU_ConcurrentSet(t *testing.T) {
	cache := NewLRUCache[string, int](2)

	go func() {
		cache.Set("a", 1)
	}()

	go func() {
		cache.Set("a", 2)
	}()

	defer func() {
		if _, ok := cache.Get("a"); !ok {
			t.Errorf("Expected ok")
		}
	}()
}

func BenchmarkLRU_Set(b *testing.B) {
	cache := NewLRUCache[int, int](1000)
	b.ResetTimer() // Don't count the setup time
	for i := 0; i < b.N; i++ {
		cache.Set(i, i)
	}
}
