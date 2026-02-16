package lru

type Node[K comparable, V any] struct {
	Key   K
	Value V
	Prev  *Node[K, V]
	Next  *Node[K, V]
}

type LRU[K comparable, V any] struct {
	capacity int
	nodesMap map[K]*Node[K, V]
	head     *Node[K, V]
	tail     *Node[K, V]
}

func (c *LRU[K, V]) Get(key K) (V, bool) {
	var node *Node[K, V]

	node, found := c.nodesMap[key]
	if !found {
		return node.Value, false
	}
	c.moveToFront(node)

	return node.Value, true
}

func (c *LRU[K, V]) Set(key K, value V) {
	node, found := c.nodesMap[key]
	if found {
		c.moveToFront(node)
		c.nodesMap[key].Value = value
		return
	}

	newNode := &Node[K, V]{Key: key, Value: value}
	if c.capacity >= len(c.nodesMap) {
		c.evict()
		c.capacity--
	}
	c.capacity++
	c.nodesMap[key] = newNode
	c.moveToFront(node)
}

func (c *LRU[K, V]) moveToFront(node *Node[K, V]) {
	if c.tail == nil {
		c.tail = node
		c.head = node
		return
	}
	oldHead := c.head
	node.Next = oldHead
	oldHead.Prev = node
	c.head = node
}

func (c *LRU[K, V]) evict() {
	if c.tail == nil {
		return
	}

	oldTail := c.tail
	delete(c.nodesMap, oldTail.Key)
	c.tail = oldTail.Prev
}
