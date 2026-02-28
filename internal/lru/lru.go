package lru

import "time"

type Node[K comparable, V any] struct {
	Key       K
	Value     V
	Prev      *Node[K, V]
	Next      *Node[K, V]
	ExpiresAt time.Time
}

type Stats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
}

type LRU[K comparable, V any] struct {
	capacity int
	nodesMap map[K]*Node[K, V]
	head     *Node[K, V]
	tail     *Node[K, V]
	stats    Stats
}

type Entry[V any] struct {
	Value    V
	ExpiryAt time.Time
}

func NewLRUCache[K comparable, V any](capacity int) *LRU[K, V] {
	return &LRU[K, V]{
		capacity: capacity,
		nodesMap: make(map[K]*Node[K, V]),
	}
}

func (c *LRU[K, V]) Get(key K) (V, bool) {
	var emptyValue V
	node, found := c.nodesMap[key]
	if !found {
		c.stats.Misses++
		return emptyValue, false
	}

	if time.Now().After(node.ExpiresAt) {
		c.stats.Misses++
		return emptyValue, false
	}

	c.extract(node)
	c.pushFront(node)

	c.stats.Hits++
	return node.Value, true
}

func (c *LRU[K, V]) Set(key K, value V, ttl time.Duration) {
	if node, found := c.nodesMap[key]; found {
		node.Value = value
		c.extract(node)
		c.pushFront(node)
		return
	}

	if len(c.nodesMap) >= c.capacity {
		c.evict()
	}

	newNode := &Node[K, V]{Key: key, Value: value, ExpiresAt: time.Now().Add(ttl)}
	c.nodesMap[key] = newNode
	c.pushFront(newNode)
}

func (c *LRU[K, V]) DeleteExpired() {
	now := time.Now()
	for key, node := range c.nodesMap {
		if now.After(node.ExpiresAt) {
			c.extract(node)
			delete(c.nodesMap, key)
		}
	}
}

func (c *LRU[K, V]) Stats() Stats {
	return c.stats
}

func (c *LRU[K, V]) Items() map[K]Entry[V] {
	res := make(map[K]Entry[V])
	for k, node := range c.nodesMap {
		res[k] = Entry[V]{
			Value:    node.Value,
			ExpiryAt: node.ExpiresAt,
		}
	}
	return res
}

func (c *LRU[K, V]) extract(node *Node[K, V]) {
	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		c.head = node.Next // node was the head
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		c.tail = node.Prev // node was the tail
	}

	node.Next = nil
	node.Prev = nil
}

func (c *LRU[K, V]) pushFront(node *Node[K, V]) {
	node.Next = c.head
	node.Prev = nil

	if c.head != nil {
		c.head.Prev = node
	}
	c.head = node

	if c.tail == nil {
		c.tail = node
	}
}

func (c *LRU[K, V]) evict() {
	if c.tail == nil {
		return
	}

	delete(c.nodesMap, c.tail.Key)
	c.stats.Evictions++

	if c.head == c.tail {
		c.head = nil
		c.tail = nil
		return
	}

	c.tail = c.tail.Prev
	c.tail.Next = nil
}
