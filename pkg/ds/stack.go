package ds

// Stack 顺序栈，泛型实现；图遍历的迭代 DFS 用它保存帧。
type Stack[T any] struct {
	items []T
}

// NewStack 建一个空栈。
func NewStack[T any]() *Stack[T] { return &Stack[T]{} }

// Push 入栈。
func (s *Stack[T]) Push(v T) { s.items = append(s.items, v) }

// Pop 出栈，第二个返回值表示栈是否非空。
func (s *Stack[T]) Pop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

// Peek 看一眼栈顶。
func (s *Stack[T]) Peek() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Len 元素个数。
func (s *Stack[T]) Len() int { return len(s.items) }
