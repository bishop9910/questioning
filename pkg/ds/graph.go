package ds

import "sort"

// Graph 邻接表有向图，用来表示「选了某选项跳到某题」。
type Graph struct {
	adj map[int64][]int64
}

// NewGraph 建空图。
func NewGraph() *Graph { return &Graph{adj: map[int64][]int64{}} }

// AddEdge 加一条 from → to 的边；to 为 0 表示交卷，不建边。
func (g *Graph) AddEdge(from, to int64) {
	if to == 0 {
		return
	}
	g.adj[from] = append(g.adj[from], to)
}

// Nodes 返回所有出现过的节点（升序，保证遍历顺序稳定）。
func (g *Graph) Nodes() []int64 {
	seen := map[int64]bool{}
	for from, tos := range g.adj {
		seen[from] = true
		for _, t := range tos {
			seen[t] = true
		}
	}
	out := make([]int64, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

type dfsFrame struct {
	node int64
	idx  int
}

// Cycle 用显式栈做 DFS 三色标记，返回环上的节点序列（首尾相同），无环返回 nil。
func (g *Graph) Cycle() []int64 {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[int64]int{}
	var path []int64

	for _, start := range g.Nodes() {
		if color[start] != white {
			continue
		}
		stack := NewStack[dfsFrame]()
		stack.Push(dfsFrame{node: start})
		color[start] = gray
		path = append(path, start)

		for stack.Len() > 0 {
			top, _ := stack.Peek()
			edges := g.adj[top.node]
			if top.idx >= len(edges) {
				stack.Pop()
				path = path[:len(path)-1]
				color[top.node] = black
				continue
			}
			cur, _ := stack.Pop()
			next := edges[cur.idx]
			cur.idx++
			stack.Push(cur)

			switch color[next] {
			case white:
				color[next] = gray
				stack.Push(dfsFrame{node: next})
				path = append(path, next)
			case gray:
				for i := len(path) - 1; i >= 0; i-- {
					if path[i] == next {
						cyc := append([]int64{}, path[i:]...)
						return append(cyc, next)
					}
				}
			}
		}
	}
	return nil
}

// Reachable 从 start 出发能到达的节点集合（BFS）。
func (g *Graph) Reachable(start int64) map[int64]bool {
	seen := map[int64]bool{start: true}
	queue := []int64{start}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		for _, t := range g.adj[n] {
			if !seen[t] {
				seen[t] = true
				queue = append(queue, t)
			}
		}
	}
	return seen
}
