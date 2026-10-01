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
