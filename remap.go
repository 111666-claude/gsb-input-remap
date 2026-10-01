// Package remap 是按键重映射：同一按键的多条绑定按优先级取最优先（同优先级取动作
// 字典序最小），重绑参数不合法时整体失败且不改状态，查询走按键索引、单次摊还 O(1)。
package remap

// Binding 是一条按键绑定。
type Binding struct {
	Key      string
	Action   string
	Priority int // 越小越优先
}

// Map 是键位表。
type Map struct {
	bindings []Binding
	// index 记录每个键当前胜出的绑定在 bindings 中的下标。
	index   map[string]int
	scanned int
}

// NewMap 建键位表并建立按键索引。构建 O(绑定数)，之后查询摊还 O(1)。
func NewMap(bindings []Binding) *Map {
	m := &Map{
		bindings: append([]Binding(nil), bindings...),
		index:    make(map[string]int),
	}
	for index, binding := range m.bindings {
		m.indexWins(binding.Key, index)
	}
	return m
}

// indexWins 判断 position 上的绑定是否应成为 key 的胜出者。
func (m *Map) indexWins(key string, position int) {
	current, exists := m.index[key]
	if !exists || better(m.bindings[position], m.bindings[current]) {
		m.index[key] = position
	}
}

// better 按"优先级最小，同优先级动作名字典序最小"判断 a 是否胜出。
func better(a, b Binding) bool {
	return a.Priority < b.Priority ||
		(a.Priority == b.Priority && a.Action < b.Action)
}

// Action 返回某个键触发的动作；没有绑定返回空串。
// 走按键索引，每次查询只计一次索引探测，Scanned 与绑定数无关。
func (m *Map) Action(key string) string {
	m.scanned++
	if position, ok := m.index[key]; ok {
		return m.bindings[position].Action
	}
	return ""
}

// Rebind 把一个键绑到新动作（替换该键此前的全部绑定）。
// key 或 action 为空时返回 false，键位表与调用前完全一致：先校验参数，
// 校验通过后才构建新状态并一次性提交。
func (m *Map) Rebind(key, action string, priority int) bool {
	if key == "" || action == "" {
		return false
	}

	kept := make([]Binding, 0, len(m.bindings)+1)
	index := make(map[string]int, len(m.index))
	for _, binding := range m.bindings {
		if binding.Key != key {
			kept = append(kept, binding)
			if current, exists := index[binding.Key]; !exists ||
				better(binding, kept[current]) {
				index[binding.Key] = len(kept) - 1
			}
		}
	}

	kept = append(kept, Binding{Key: key, Action: action, Priority: priority})
	index[key] = len(kept) - 1

	m.bindings = kept
	m.index = index
	return true
}

// Bound 是不同按键的个数。
func (m *Map) Bound() int {
	return len(m.index)
}

// Scanned 是累计扫描的绑定数（规模观测）。
func (m *Map) Scanned() int { return m.scanned }
