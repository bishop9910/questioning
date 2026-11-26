package ds

import "math/rand"

const maxLevel = 16

type skipNode struct {
	score int64
	value int64
	next  []*skipNode
	prev  *skipNode // 只在第 0 层维护，用于从新到旧翻页
}

// SkipList 跳表：给每份问卷的答卷建时间索引，按时间区间查询和倒序分页都是 O(log n)。
type SkipList struct {
	head  *skipNode
	tail  *skipNode
	level int
	size  int
	rnd   *rand.Rand
}

// NewSkipList 建空表，head 是哨兵（score 取最小）。
func NewSkipList() *SkipList {
	return &SkipList{
		head:  &skipNode{score: -1 << 62, next: make([]*skipNode, maxLevel)},
		level: 1,
		rnd:   rand.New(rand.NewSource(20260101)),
	}
}

func lessThan(aScore, aVal, bScore, bVal int64) bool {
	if aScore != bScore {
		return aScore < bScore
	}
	return aVal < bVal
}

func (s *SkipList) randomLevel() int {
	lvl := 1
	for lvl < maxLevel && s.rnd.Intn(4) == 0 {
		lvl++
	}
	return lvl
}

// Insert 插入一个 (时间戳, 答卷ID)，允许时间戳重复。
func (s *SkipList) Insert(score, value int64) {
	update := make([]*skipNode, maxLevel)
	x := s.head
	for i := s.level - 1; i >= 0; i-- {
		for x.next[i] != nil && lessThan(x.next[i].score, x.next[i].value, score, value) {
			x = x.next[i]
		}
		update[i] = x
	}
	lvl := s.randomLevel()
	if lvl > s.level {
		for i := s.level; i < lvl; i++ {
			update[i] = s.head
		}
		s.level = lvl
	}
	n := &skipNode{score: score, value: value, next: make([]*skipNode, lvl)}
	for i := 0; i < lvl; i++ {
		n.next[i] = update[i].next[i]
		update[i].next[i] = n
	}
	if update[0] != s.head {
		n.prev = update[0]
	}
	if n.next[0] != nil {
		n.next[0].prev = n
	} else {
		s.tail = n
	}
	s.size++
}

// Len 节点数。
func (s *SkipList) Len() int { return s.size }

// RangeAsc 取 [from, to] 内的时间戳对应答卷ID，按时间升序。
func (s *SkipList) RangeAsc(from, to int64) []int64 {
	x := s.head
	for i := s.level - 1; i >= 0; i-- {
		for x.next[i] != nil && x.next[i].score < from {
			x = x.next[i]
		}
	}
	out := make([]int64, 0, 8)
	for n := x.next[0]; n != nil && n.score <= to; n = n.next[0] {
		out = append(out, n.value)
	}
	return out
}

// Page 从新到旧翻页，offset 跳过的条数，limit 每页条数。
func (s *SkipList) Page(offset, limit int) []int64 {
	out := make([]int64, 0, limit)
	n := s.tail
	for i := 0; i < offset && n != nil; i++ {
		n = n.prev
	}
	for ; n != nil && len(out) < limit; n = n.prev {
		out = append(out, n.value)
	}
	return out
}
