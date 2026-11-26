package ds

// lruEntry 是 LRU 的链表节点（哈希表存 key → 节点指针，链表维持最近使用顺序）。
type lruEntry[K comparable, V any] struct {
	key        K
	val        V
	prev, next *lruEntry[K, V]
}

// LRU 最近最少使用缓存：哈希表 O(1) 定位 + 双向链表 O(1) 调整顺序。
type LRU[K comparable, V any] struct {
	capacity int
	m        map[K]*lruEntry[K, V]
	head     *lruEntry[K, V] // 最近使用
	tail     *lruEntry[K, V] // 最久未用
}

// NewLRU 建一个容量为 capacity 的缓存。
func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	if capacity < 1 {
		capacity = 1
	}
	return &LRU[K, V]{capacity: capacity, m: make(map[K]*lruEntry[K, V], capacity)}
}

// Get 命中时把节点挪到表头。
func (c *LRU[K, V]) Get(k K) (V, bool) {
	var zero V
	e, ok := c.m[k]
	if !ok {
		return zero, false
	}
	c.moveFront(e)
	return e.val, true
}

// Put 写入并挪到表头，超容量时淘汰表尾。
func (c *LRU[K, V]) Put(k K, v V) {
	if e, ok := c.m[k]; ok {
		e.val = v
		c.moveFront(e)
		return
	}
	e := &lruEntry[K, V]{key: k, val: v}
	c.m[k] = e
	c.pushFront(e)
	if len(c.m) > c.capacity {
		old := c.tail
		if old != nil {
			c.unlink(old)
			delete(c.m, old.key)
		}
	}
}

// Invalidate 删除某个 key。
func (c *LRU[K, V]) Invalidate(k K) {
	if e, ok := c.m[k]; ok {
		c.unlink(e)
		delete(c.m, k)
	}
}

// Len 当前条数。
func (c *LRU[K, V]) Len() int { return len(c.m) }

func (c *LRU[K, V]) pushFront(e *lruEntry[K, V]) {
	e.prev = nil
	e.next = c.head
	if c.head != nil {
		c.head.prev = e
	}
	c.head = e
	if c.tail == nil {
		c.tail = e
	}
}

func (c *LRU[K, V]) unlink(e *lruEntry[K, V]) {
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		c.head = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		c.tail = e.prev
	}
	e.prev, e.next = nil, nil
}

func (c *LRU[K, V]) moveFront(e *lruEntry[K, V]) {
	if c.head == e {
		return
	}
	c.unlink(e)
	c.pushFront(e)
}
