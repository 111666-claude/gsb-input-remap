// Package remap 是按键重映射：冲突按优先级取最优先、重绑原子、按动作索引查询。
// 缺陷：Action 取最后一条匹配（优先级失效）、Rebind 先删后加（失败时留下没绑定的中间态）、
// 每次查询全扫绑定表。
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
	scanned  int
}

// NewMap 建键位表。
func NewMap(bindings []Binding) *Map {
	return &Map{bindings: append([]Binding(nil), bindings...)}
}

// Action 返回某个键触发的动作。
// 缺陷：取最后一条匹配的绑定（优先级形同虚设），而且每次全扫。
func (m *Map) Action(key string) string {
	found := ""
	for _, binding := range m.bindings {
		m.scanned++
		if binding.Key == key {
			found = binding.Action
		}
	}
	return found
}

// Rebind 把一个键绑到新动作。缺陷：先把该键的旧绑定全删掉再加新的，
// 参数不合法时留下"这个键没有任何绑定"的中间态。
func (m *Map) Rebind(key, action string, priority int) bool {
	kept := make([]Binding, 0, len(m.bindings))
	for _, binding := range m.bindings {
		if binding.Key != key {
			kept = append(kept, binding)
		}
	}
	m.bindings = kept
	if key == "" || action == "" {
		return false
	}
	m.bindings = append(m.bindings, Binding{Key: key, Action: action, Priority: priority})
	return true
}

// Bound 是不同按键的个数。
func (m *Map) Bound() int {
	seen := map[string]bool{}
	for _, binding := range m.bindings {
		seen[binding.Key] = true
	}
	return len(seen)
}

// Scanned 是累计扫描的绑定数（规模观测）。
func (m *Map) Scanned() int { return m.scanned }
