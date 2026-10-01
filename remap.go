// Package remap 是按键重映射：冲突按优先级取最优先、重绑原子、按动作索引查询。
package remap

// Binding 是一条按键绑定。
type Binding struct {
	Key      string
	Action   string
	Priority int // 越小越优先
}

// Map 是键位表。按按键建索引，每个按键只保留最优先的一条绑定。
type Map struct {
	best    map[string]Binding
	scanned int
}

// better 判断 a 是否比 b 更优先：优先级小者优先，同优先级动作名字典序小者优先。
func better(a, b Binding) bool {
	if a.Priority != b.Priority {
		return a.Priority < b.Priority
	}
	return a.Action < b.Action
}

// NewMap 建键位表。
func NewMap(bindings []Binding) *Map {
	m := &Map{best: make(map[string]Binding, len(bindings))}
	for _, binding := range bindings {
		if current, ok := m.best[binding.Key]; !ok || better(binding, current) {
			m.best[binding.Key] = binding
		}
	}
	return m
}

// Action 返回某个键触发的动作。按索引查询，单次 O(1)。
func (m *Map) Action(key string) string {
	m.scanned++
	return m.best[key].Action
}

// Rebind 把一个键绑到新动作。参数不合法时返回 false 且键位表完全不变。
func (m *Map) Rebind(key, action string, priority int) bool {
	if key == "" || action == "" {
		return false
	}
	m.best[key] = Binding{Key: key, Action: action, Priority: priority}
	return true
}

// Bound 是不同按键的个数。
func (m *Map) Bound() int {
	return len(m.best)
}

// Scanned 是累计扫描的绑定数（规模观测）。
func (m *Map) Scanned() int { return m.scanned }
