package remap

import (
	"math/rand"
	"testing"
)

func TestConflictPicksSmallestPriority(t *testing.T) {
	keymap := NewMap([]Binding{
		{Key: "k", Action: "attack", Priority: 2},
		{Key: "k", Action: "jump", Priority: 1},
		{Key: "k", Action: "run", Priority: 3},
	})
	if got := keymap.Action("k"); got != "jump" {
		t.Fatalf("应取优先级最小的绑定，得到 action=%s", got)
	}
	// 同一份键位表重复查询结果一致。
	if again := keymap.Action("k"); again != "jump" {
		t.Fatalf("重复查询应得到相同结果，得到 action=%s", again)
	}
}

func TestSamePriorityPicksLexicographicallySmallest(t *testing.T) {
	keymap := NewMap([]Binding{
		{Key: "k", Action: "jump", Priority: 1},
		{Key: "k", Action: "attack", Priority: 1},
	})
	if got := keymap.Action("k"); got != "attack" {
		t.Fatalf("同优先级应取动作名字典序最小，得到 action=%s", got)
	}
}

func TestInvalidRebindIsAtomic(t *testing.T) {
	keymap := NewMap([]Binding{
		{Key: "k", Action: "jump", Priority: 1},
		{Key: "k", Action: "attack", Priority: 2},
		{Key: "q", Action: "run", Priority: 1},
	})

	before := keymap.Action("k")
	beforeBound := keymap.Bound()

	if keymap.Rebind("k", "", 1) {
		t.Fatal("action 为空应返回 false")
	}
	if keymap.Rebind("", "jump", 1) {
		t.Fatal("key 为空应返回 false")
	}
	if keymap.Rebind("", "", 1) {
		t.Fatal("key、action 均空应返回 false")
	}

	// 非法重绑必须在触碰任何状态前返回，不产生扫描开销。
	if got := keymap.Scanned(); got != 1 {
		t.Fatalf("非法重绑不应产生扫描开销：%d，期望 1", got)
	}
	if got := keymap.Action("k"); got != before {
		t.Fatalf("非法重绑后动作被改变：action=%s，期望 %s", got, before)
	}
	if got := keymap.Bound(); got != beforeBound {
		t.Fatalf("非法重绑后绑定按键数被改变：%d，期望 %d", got, beforeBound)
	}
}

func TestValidRebindReplacesAllBindingsForKey(t *testing.T) {
	keymap := NewMap([]Binding{
		{Key: "k", Action: "jump", Priority: 1},
		{Key: "k", Action: "attack", Priority: 2},
	})
	if !keymap.Rebind("k", "run", 9) {
		t.Fatal("合法重绑应成功")
	}
	if got := keymap.Action("k"); got != "run" {
		t.Fatalf("重绑后应取新动作，得到 action=%s", got)
	}
	if got := keymap.Bound(); got != 1 {
		t.Fatalf("重绑不应增加按键数，得到 %d", got)
	}

	if !keymap.Rebind("p", "crouch", 1) {
		t.Fatal("给新键加绑定应成功")
	}
	if got := keymap.Bound(); got != 2 {
		t.Fatalf("新增按键后应有 2 个不同按键，得到 %d", got)
	}
}

func TestBoundCountsDistinctKeys(t *testing.T) {
	keymap := NewMap([]Binding{
		{Key: "k", Action: "jump", Priority: 1},
		{Key: "k", Action: "attack", Priority: 1},
		{Key: "q", Action: "run", Priority: 1},
	})
	if got := keymap.Bound(); got != 2 {
		t.Fatalf("Bound 只应数不同按键，得到 %d", got)
	}
}

func TestActionLookupIsIndexed(t *testing.T) {
	const bindingCount = 5000
	const queryCount = 3000

	bindings := make([]Binding, 0, bindingCount)
	for index := 0; index < bindingCount; index++ {
		bindings = append(bindings, Binding{
			Key:      keyName(index),
			Action:   "jump",
			Priority: index,
		})
	}
	keymap := NewMap(bindings)

	for query := 0; query < queryCount; query++ {
		keymap.Action(keyName(bindingCount - 1))
	}

	// 每次查询只做一次索引探测：扫描量等于查询次数，与 5000 条绑定无关
	//（线性扫描会达到 1500 万）。
	if got := keymap.Scanned(); got != queryCount {
		t.Fatalf("查询应走索引、每次只计 1，scanned=%d，期望 %d", got, queryCount)
	}
}

func keyName(index int) string {
	return "k-" + itoa(index)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	pos := len(digits)
	for value > 0 {
		pos--
		digits[pos] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[pos:])
}

// referenceMap 是独立的朴素参考实现：每次查询线性扫描取最优先绑定，
// 重绑仅在参数合法时整体替换。
type referenceMap struct {
	bindings []Binding
}

func newReference(bindings []Binding) *referenceMap {
	return &referenceMap{bindings: append([]Binding(nil), bindings...)}
}

func (r *referenceMap) action(key string) string {
	best := -1
	for index := range r.bindings {
		binding := &r.bindings[index]
		if binding.Key == key &&
			(best == -1 || better(*binding, r.bindings[best])) {
			best = index
		}
	}
	if best == -1 {
		return ""
	}
	return r.bindings[best].Action
}

func (r *referenceMap) rebind(key, action string, priority int) bool {
	if key == "" || action == "" {
		return false
	}
	kept := make([]Binding, 0, len(r.bindings)+1)
	for _, binding := range r.bindings {
		if binding.Key != key {
			kept = append(kept, binding)
		}
	}
	r.bindings = append(kept, Binding{Key: key, Action: action, Priority: priority})
	return true
}

func (r *referenceMap) bound() int {
	seen := map[string]struct{}{}
	for _, binding := range r.bindings {
		seen[binding.Key] = struct{}{}
	}
	return len(seen)
}

// TestDifferentialAgainstReference 用 200 组随机「绑定表 + 重绑序列」对拍：
// 每次重绑的成功与否、每个按键的动作、绑定按键数都必须与参考实现一致。
func TestDifferentialAgainstReference(t *testing.T) {
	rng := rand.New(rand.NewSource(20261001))

	for trial := 0; trial < 200; trial++ {
		bindings := randomBindings(rng)
		actual := NewMap(bindings)
		reference := newReference(bindings)

		keys := probeKeys(bindings)

		check := func(stage string) {
			t.Helper()
			for _, key := range keys {
				if got, want := actual.Action(key), reference.action(key); got != want {
					t.Fatalf("第 %d 组(%s) key=%q 动作不一致：%q，参考为 %q",
						trial, stage, key, got, want)
				}
			}
			if got, want := actual.Bound(), reference.bound(); got != want {
				t.Fatalf("第 %d 组(%s) 绑定按键数不一致：%d，参考为 %d",
					trial, stage, got, want)
			}
		}

		check("初始")

		for step := 0; step < 40; step++ {
			key := keys[rng.Intn(len(keys))]
			if rng.Intn(6) == 0 {
				key = "" // 非法 key
			}
			action := []string{"jump", "attack", "run", "crouch"}[rng.Intn(4)]
			if rng.Intn(6) == 0 {
				action = "" // 非法 action
			}
			priority := rng.Intn(4) // 刻意制造大量同优先级冲突

			gotOK := actual.Rebind(key, action, priority)
			wantOK := reference.rebind(key, action, priority)
			if gotOK != wantOK {
				t.Fatalf("第 %d 组第 %d 步重绑 ok 不一致：%v，参考为 %v（key=%q action=%q）",
					trial, step, gotOK, wantOK, key, action)
			}
			check("重绑后")
		}
	}
}

func randomBindings(rng *rand.Rand) []Binding {
	count := rng.Intn(30)
	bindings := make([]Binding, 0, count)
	for index := 0; index < count; index++ {
		keyNumber := rng.Intn(6) // 小键空间制造同键冲突
		bindings = append(bindings, Binding{
			Key:      keyName(keyNumber),
			Action:   []string{"jump", "attack", "run", "crouch"}[rng.Intn(4)],
			Priority: rng.Intn(4),
		})
	}
	return bindings
}

func probeKeys(bindings []Binding) []string {
	seen := map[string]bool{}
	keys := make([]string, 0, 8)
	for _, binding := range bindings {
		if !seen[binding.Key] {
			seen[binding.Key] = true
			keys = append(keys, binding.Key)
		}
	}
	keys = append(keys, keyName(40), keyName(99))
	return keys
}
