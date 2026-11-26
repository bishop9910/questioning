package ds

// Node 是双向链表节点。
type Node struct {
	Value int64
	prev  *Node
	next  *Node
}

// DoublyList 双向链表：题目拖拽排序时只重连指针，不移动数据。
type DoublyList struct {
	head *Node
	tail *Node
	size int
}

// NewList 建立空链表。
func NewList() *DoublyList { return &DoublyList{} }

// Push 在尾部追加，返回新节点（后续 MoveAfter / Remove 直接用这个句柄）。
func (l *DoublyList) Push(v int64) *Node {
	n := &Node{Value: v}
	if l.tail == nil {
		l.head, l.tail = n, n
	} else {
		n.prev = l.tail
		l.tail.next = n
		l.tail = n
	}
	l.size++
	return n
}

func (l *DoublyList) detach(n *Node) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		l.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		l.tail = n.prev
	}
	n.prev, n.next = nil, nil
	l.size--
}

// Remove 摘除节点。
func (l *DoublyList) Remove(n *Node) {
	if n == nil {
		return
	}
	l.detach(n)
}

// MoveAfter 把 n 挪到 at 之后（at 为 nil 表示放到表头），整体 O(1)。
func (l *DoublyList) MoveAfter(n, at *Node) {
	if n == nil || n == at {
		return
	}
	l.detach(n)
	if at == nil {
		n.next = l.head
		if l.head != nil {
			l.head.prev = n
		} else {
			l.tail = n
		}
		l.head = n
	} else {
		n.prev = at
		n.next = at.next
		if at.next != nil {
			at.next.prev = n
		} else {
			l.tail = n
		}
		at.next = n
	}
	l.size++
}

// Values 顺序导出，O(n)。
func (l *DoublyList) Values() []int64 {
	out := make([]int64, 0, l.size)
	for n := l.head; n != nil; n = n.next {
		out = append(out, n.Value)
	}
	return out
}

// Len 节点数。
func (l *DoublyList) Len() int { return l.size }
