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
	node, found := c.nodesMap[key]
	if !found {
		var zero V
		return zero, false
	}
	c.extract(node)
	c.pushFront(node)

	return node.Value, true
}

func (c *LRU[K, V]) Set(key K, value V) {
	if node, found := c.nodesMap[key]; found {
		node.Value = value
		c.extract(node)
		c.pushFront(node)
		return
	}

	if c.capacity >= len(c.nodesMap) {
		c.evict()
	}

	newNode := &Node[K, V]{Key: key, Value: value}
	c.nodesMap[key] = newNode
	c.pushFront(newNode)
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
	if oldHead := c.head; oldHead != nil {
		oldHead.Prev = node
		node.Next = oldHead
	}
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
