package ds

// KV 是哈希表的一项。
type KV struct {
	Key string
	Val int
}

// HashMap 开放寻址（线性探测）哈希表，用来统计选项票数与填空词频。
// 负载因子超过 0.75 时容量翻倍并整体再散列。
type HashMap struct {
	keys []string
	vals []int
	used []bool
	size int
	mask uint64
}

// NewHashMap 建一张容量 16 的表。
func NewHashMap() *HashMap { return newSizedHashMap(16) }

func newSizedHashMap(n int) *HashMap {
	return &HashMap{
		keys: make([]string, n),
		vals: make([]int, n),
		used: make([]bool, n),
		mask: uint64(n - 1),
	}
}

// hashKey 是 FNV-1a 64 位散列。
func hashKey(s string) uint64 {
	const (
		offset64 = 14695981039346656037
		prime64  = 1099511628211
	)
	var h uint64 = offset64
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}
	return h
}

// Add 给 key 的计数加上 delta，返回累加后的值。
func (m *HashMap) Add(key string, delta int) int {
	if (m.size+1)*4 >= len(m.keys)*3 {
		m.grow()
	}
	h := hashKey(key) & m.mask
	for {
		if !m.used[h] {
			m.used[h] = true
			m.keys[h] = key
			m.vals[h] = delta
			m.size++
			return delta
		}
		if m.keys[h] == key {
			m.vals[h] += delta
			return m.vals[h]
		}
		h = (h + 1) & m.mask
	}
}

// Get 取计数，不存在时返回 0。
func (m *HashMap) Get(key string) int {
	h := hashKey(key) & m.mask
	for {
		if !m.used[h] {
			return 0
		}
		if m.keys[h] == key {
			return m.vals[h]
		}
		h = (h + 1) & m.mask
	}
}

// Len 不同 key 的个数。
func (m *HashMap) Len() int { return m.size }

// Items 导出全部键值对（顺序不保证）。
func (m *HashMap) Items() []KV {
	out := make([]KV, 0, m.size)
	for i, ok := range m.used {
		if ok {
			out = append(out, KV{Key: m.keys[i], Val: m.vals[i]})
		}
	}
	return out
}

func (m *HashMap) grow() {
	old := *m
	m.keys = make([]string, len(old.keys)*2)
	m.vals = make([]int, len(old.keys)*2)
	m.used = make([]bool, len(old.keys)*2)
	m.mask = uint64(len(m.keys) - 1)
	m.size = 0
	for i, ok := range old.used {
		if !ok {
			continue
		}
		m.Add(old.keys[i], old.vals[i])
	}
}
