package ds

// WordCount 是一条词频统计。
type WordCount struct {
	Word  string
	Count int
}

// MinHeap 以计数为关键字的小顶堆，用来在海量词频里留 Top-K。
type MinHeap struct {
	items []WordCount
}

func (h *MinHeap) Len() int { return len(h.items) }

func (h *MinHeap) less(i, j int) bool {
	a, b := h.items[i], h.items[j]
	if a.Count != b.Count {
		return a.Count < b.Count
	}
	return a.Word > b.Word // 计数相同时字典序大的先出堆，保证结果稳定
}

func (h *MinHeap) swap(i, j int) { h.items[i], h.items[j] = h.items[j], h.items[i] }

// Push 上浮调整，O(log n)。
func (h *MinHeap) Push(w WordCount) {
	h.items = append(h.items, w)
	i := len(h.items) - 1
	for i > 0 {
		parent := (i - 1) / 2
		if !h.less(i, parent) {
			break
		}
		h.swap(i, parent)
		i = parent
	}
}

// Pop 弹出最小值，下沉调整，O(log n)。
func (h *MinHeap) Pop() WordCount {
	top := h.items[0]
	last := len(h.items) - 1
	h.items[0] = h.items[last]
	h.items = h.items[:last]
	i := 0
	for {
		l, r := 2*i+1, 2*i+2
		small := i
		if l < len(h.items) && h.less(l, small) {
			small = l
		}
		if r < len(h.items) && h.less(r, small) {
			small = r
		}
		if small == i {
			break
		}
		h.swap(i, small)
		i = small
	}
	return top
}

// Peek 看堆顶（当前第 K 名的门槛）。
func (h *MinHeap) Peek() WordCount { return h.items[0] }

// TopK 取前 k 个高频词，返回按计数降序的切片。
func TopK(items []WordCount, k int) []WordCount {
	if k <= 0 || len(items) == 0 {
		return nil
	}
	h := &MinHeap{}
	for _, it := range items {
		switch {
		case h.Len() < k:
			h.Push(it)
		case it.Count > h.Peek().Count || (it.Count == h.Peek().Count && it.Word < h.Peek().Word):
			h.Pop()
			h.Push(it)
		}
	}
	out := make([]WordCount, h.Len())
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = h.Pop()
	}
	return out
}
