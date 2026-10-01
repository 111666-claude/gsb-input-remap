package remap

import "testing"

func TestActionOfSingleBinding(t *testing.T) {
	keymap := NewMap([]Binding{{Key: "k", Action: "jump", Priority: 1}})
	if keymap.Action("k") != "jump" {
		t.Fatal("单个绑定应该能查到动作")
	}
}

func TestUnknownKeyHasNoAction(t *testing.T) {
	if NewMap(nil).Action("k") != "" {
		t.Fatal("没有绑定应该返回空动作")
	}
}

func TestRebindAddsBinding(t *testing.T) {
	keymap := NewMap(nil)
	if !keymap.Rebind("k", "jump", 1) {
		t.Fatal("合法重绑应该成功")
	}
	if keymap.Bound() != 1 {
		t.Fatalf("绑定数不对：%d", keymap.Bound())
	}
}

func TestScannedStartsAtZero(t *testing.T) {
	if NewMap(nil).Scanned() != 0 {
		t.Fatal("初始扫描数应该是 0")
	}
}

func TestConflictTakesLowestPriority(t *testing.T) {
	keymap := NewMap([]Binding{
		{Key: "k", Action: "attack", Priority: 2},
		{Key: "k", Action: "jump", Priority: 1},
	})
	if keymap.Action("k") != "jump" {
		t.Fatal("同一个键应该取优先级最小的绑定")
	}
}

func TestConflictSamePriorityTakesLexicographicAction(t *testing.T) {
	keymap := NewMap([]Binding{
		{Key: "k", Action: "jump", Priority: 1},
		{Key: "k", Action: "attack", Priority: 1},
	})
	if keymap.Action("k") != "attack" {
		t.Fatal("同优先级应该取动作名字典序最小的绑定")
	}
}

func TestRebindIsAtomicOnInvalidArgs(t *testing.T) {
	keymap := NewMap([]Binding{{Key: "k", Action: "jump", Priority: 1}})
	before := keymap.Action("k")
	boundBefore := keymap.Bound()

	if keymap.Rebind("k", "", 1) {
		t.Fatal("action 为空应该返回 false")
	}
	if keymap.Rebind("", "jump", 1) {
		t.Fatal("key 为空应该返回 false")
	}

	if keymap.Action("k") != before {
		t.Fatal("非法重绑后原绑定必须保持不变")
	}
	if keymap.Bound() != boundBefore {
		t.Fatal("非法重绑后绑定数必须保持不变")
	}
}

func TestRebindReplacesKeyBindings(t *testing.T) {
	keymap := NewMap([]Binding{
		{Key: "k", Action: "jump", Priority: 1},
		{Key: "k", Action: "attack", Priority: 2},
	})
	if !keymap.Rebind("k", "crouch", 3) {
		t.Fatal("合法重绑应该成功")
	}
	if keymap.Action("k") != "crouch" {
		t.Fatal("重绑后应该是新动作")
	}
	if keymap.Bound() != 1 {
		t.Fatalf("重绑同一个键不应该增加绑定数：%d", keymap.Bound())
	}
}

func TestRepeatedQueriesAreStable(t *testing.T) {
	keymap := NewMap([]Binding{
		{Key: "k", Action: "jump", Priority: 1},
		{Key: "k", Action: "attack", Priority: 2},
	})
	first := keymap.Action("k")
	for i := 0; i < 100; i++ {
		if keymap.Action("k") != first {
			t.Fatal("同一份键位表重复查询结果必须一致")
		}
	}
}

func TestActionScanCostIsConstantPerQuery(t *testing.T) {
	bindings := make([]Binding, 0, 5000)
	for index := 0; index < 5000; index++ {
		bindings = append(bindings, Binding{Key: "k", Action: "dup", Priority: index})
	}
	for index := 0; index < 5000; index++ {
		bindings = append(bindings, Binding{
			Key:      keyName(index),
			Action:   "a",
			Priority: index,
		})
	}
	keymap := NewMap(bindings)

	const queries = 10000
	for i := 0; i < queries; i++ {
		keymap.Action(keyName(4999))
	}
	if got := keymap.Scanned(); got != queries {
		t.Fatalf("Scanned 不应随绑定数线性增长：得到 %d，期望 %d", got, queries)
	}
}

func keyName(index int) string {
	return "unique-" + string(rune('a'+index/26)) + string(rune('a'+index%26))
}

// referenceMap 是参考实现：冲突取优先级首条、重绑原子、按索引查询。
type referenceMap struct {
	best map[string]Binding
}

func newReference(bindings []Binding) *referenceMap {
	r := &referenceMap{best: map[string]Binding{}}
	for _, binding := range bindings {
		current, ok := r.best[binding.Key]
		if !ok || binding.Priority < current.Priority ||
			(binding.Priority == current.Priority && binding.Action < current.Action) {
			r.best[binding.Key] = binding
		}
	}
	return r
}

func (r *referenceMap) action(key string) string { return r.best[key].Action }

func (r *referenceMap) rebind(key, action string, priority int) bool {
	if key == "" || action == "" {
		return false
	}
	r.best[key] = Binding{Key: key, Action: action, Priority: priority}
	return true
}

func (r *referenceMap) bound() int { return len(r.best) }

func TestDifferentialAgainstReference(t *testing.T) {
	rng := newDeterministicRNG(20261001)
	actions := []string{"jump", "attack", "crouch", "dash", "block", ""}

	for group := 0; group < 200; group++ {
		bindingCount := rng.intn(40)
		bindings := make([]Binding, bindingCount)
		for i := range bindings {
			bindings[i] = Binding{
				Key:      keyName(rng.intn(8)),
				Action:   actions[rng.intn(len(actions)-1)],
				Priority: rng.intn(4),
			}
		}

		actual := NewMap(bindings)
		reference := newReference(bindings)

		ops := rng.intn(30)
		for op := 0; op < ops; op++ {
			key := keyName(rng.intn(10))
			if rng.intn(5) == 0 {
				key = ""
			}
			action := actions[rng.intn(len(actions))]
			priority := rng.intn(4)

			gotOK := actual.Rebind(key, action, priority)
			wantOK := reference.rebind(key, action, priority)
			if gotOK != wantOK {
				t.Fatalf("第 %d 组第 %d 步重绑返回值不一致：got %v want %v",
					group, op, gotOK, wantOK)
			}
		}

		if actual.Bound() != reference.bound() {
			t.Fatalf("第 %d 组绑定数不一致：got %d want %d",
				group, actual.Bound(), reference.bound())
		}
		for keyIndex := 0; keyIndex < 12; keyIndex++ {
			key := keyName(keyIndex)
			if actual.Action(key) != reference.action(key) {
				t.Fatalf("第 %d 组按键 %q 动作不一致：got %q want %q",
					group, key, actual.Action(key), reference.action(key))
			}
		}
	}
}

// deterministicRNG 是不依赖 math/rand 状态的简单确定性随机源。
type deterministicRNG struct{ state uint64 }

func newDeterministicRNG(seed uint64) *deterministicRNG {
	return &deterministicRNG{state: seed}
}

func (r *deterministicRNG) intn(n int) int {
	r.state ^= r.state << 13
	r.state ^= r.state >> 7
	r.state ^= r.state << 17
	return int(r.state % uint64(n))
}
