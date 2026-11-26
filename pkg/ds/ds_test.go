package ds

import (
	"fmt"
	"reflect"
	"testing"
)

func TestDoublyList(t *testing.T) {
	l := NewList()
	a := l.Push(1)
	b := l.Push(2)
	c := l.Push(3)
	if got := l.Values(); !reflect.DeepEqual(got, []int64{1, 2, 3}) {
		t.Fatalf("push: %v", got)
	}

	// 把 1 拖到 3 后面
	l.MoveAfter(a, c)
	if got := l.Values(); !reflect.DeepEqual(got, []int64{2, 3, 1}) {
		t.Fatalf("moveAfter tail: %v", got)
	}

	// 把 1 拖到表头
	l.MoveAfter(a, nil)
	if got := l.Values(); !reflect.DeepEqual(got, []int64{1, 2, 3}) {
		t.Fatalf("moveAfter head: %v", got)
	}

	l.Remove(b)
	if got := l.Values(); !reflect.DeepEqual(got, []int64{1, 3}) || l.Len() != 2 {
		t.Fatalf("remove: %v len=%d", got, l.Len())
	}

	// 删除全部节点后链表应为空且可继续追加
	l.Remove(l.head)
	l.Remove(l.head)
	if l.Len() != 0 || len(l.Values()) != 0 {
		t.Fatalf("drain: len=%d", l.Len())
	}
	l.Push(9)
	if got := l.Values(); !reflect.DeepEqual(got, []int64{9}) {
		t.Fatalf("reuse: %v", got)
	}
}

func TestStack(t *testing.T) {
	s := NewStack[int]()
	if _, ok := s.Pop(); ok {
		t.Fatal("empty pop should fail")
	}
	for i := 1; i <= 3; i++ {
		s.Push(i)
	}
	if v, _ := s.Peek(); v != 3 {
		t.Fatalf("peek: %d", v)
	}
	var got []int
	for {
		v, ok := s.Pop()
		if !ok {
			break
		}
		got = append(got, v)
	}
	if !reflect.DeepEqual(got, []int{3, 2, 1}) {
		t.Fatalf("lifo: %v", got)
	}
}

func TestHashMap(t *testing.T) {
	m := NewHashMap()
	words := []string{"甲", "乙", "丙"}
	for i := 0; i < 500; i++ {
		m.Add(words[i%3], 1)
	}
	if m.Len() != 3 {
		t.Fatalf("len: %d", m.Len())
	}
	for _, w := range words[:2] {
		if got := m.Get(w); got != 167 {
			t.Fatalf("%s = %d, want 167", w, got)
		}
	}
	if got := m.Get("丙"); got != 166 {
		t.Fatalf("丙 = %d, want 166", got)
	}
	if got := m.Get("不存在"); got != 0 {
		t.Fatalf("missing: %d", got)
	}
	if len(m.Items()) != 3 {
		t.Fatalf("items: %v", m.Items())
	}
	// 触发扩容后仍要能查到旧 key
	for i := 0; i < 100; i++ {
		m.Add(fmt.Sprintf("词%d", i), i+1)
	}
	if got := m.Get("甲"); got != 167 {
		t.Fatalf("after grow: %d", got)
	}
	if got := m.Get("词99"); got != 100 {
		t.Fatalf("after grow new key: %d", got)
	}
}

func TestSkipList(t *testing.T) {
	s := NewSkipList()
	// 插入顺序打乱，score 为提交时间（越大越新）
	scores := []int64{50, 10, 90, 70, 30, 20, 80, 60, 40, 100}
	for i, sc := range scores {
		s.Insert(sc, int64(i+1))
	}
	if s.Len() != len(scores) {
		t.Fatalf("len: %d", s.Len())
	}

	got := s.RangeAsc(30, 60)
	want := []int64{5, 9, 1, 8} // score 30,40,50,60
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("range: %v want %v", got, want)
	}

	page := s.Page(0, 3)
	if !reflect.DeepEqual(page, []int64{10, 3, 7}) {
		t.Fatalf("page0: %v", page)
	}
	page = s.Page(3, 3)
	if !reflect.DeepEqual(page, []int64{4, 8, 1}) {
		t.Fatalf("page1: %v", page)
	}

	// 相同时间戳也要能共存
	s.Insert(100, 11)
	if s.Len() != 11 {
		t.Fatalf("dup score len: %d", s.Len())
	}
	if got := s.RangeAsc(100, 100); len(got) != 2 {
		t.Fatalf("dup score range: %v", got)
	}
}

func TestGraph(t *testing.T) {
	g := NewGraph()
	// 1 → 2 → 3 → 1 成环
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)
	g.AddEdge(3, 1)
	if c := g.Cycle(); len(c) < 2 || c[0] != c[len(c)-1] {
		t.Fatalf("cycle: %v", c)
	}

	// 无环的跳题图：1→3, 2→3, 3→0(交卷)
	g2 := NewGraph()
	g2.AddEdge(1, 3)
	g2.AddEdge(2, 3)
	g2.AddEdge(3, 0)
	if c := g2.Cycle(); c != nil {
		t.Fatalf("unexpected cycle: %v", c)
	}
	if !g2.Reachable(1)[3] || g2.Reachable(1)[2] {
		t.Fatalf("reachable: %v", g2.Reachable(1))
	}
	if got := g2.Nodes(); !reflect.DeepEqual(got, []int64{1, 2, 3}) {
		t.Fatalf("nodes: %v", got)
	}
}

func TestTopK(t *testing.T) {
	items := []WordCount{
		{"的", 50}, {"调研", 30}, {"问卷", 20}, {"答辩", 5}, {"课程", 1},
	}
	got := TopK(items, 3)
	want := []WordCount{{"的", 50}, {"调研", 30}, {"问卷", 20}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("topk: %v", got)
	}
	if got := TopK(items, 10); len(got) != len(items) || got[0].Count != 50 {
		t.Fatalf("topk k>n: %v", got)
	}
	if got := TopK(nil, 3); got != nil {
		t.Fatalf("topk empty: %v", got)
	}
}

func TestLRU(t *testing.T) {
	c := NewLRU[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("get a: %v %v", v, ok)
	}
	c.Put("c", 3) // 淘汰最久未用的 b
	if _, ok := c.Get("b"); ok {
		t.Fatal("b should be evicted")
	}
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Fatalf("a should stay: %v %v", v, ok)
	}
	c.Invalidate("a")
	if _, ok := c.Get("a"); ok || c.Len() != 1 {
		t.Fatalf("invalidate: len=%d", c.Len())
	}
}
