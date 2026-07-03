package input

import "testing"

func TestTouchLayout_Controls(t *testing.T) {
	km := Keymap{
		Directions: true,
		Confirm:    "選擇",
		Cancel:     "離開",
		Letters:    []LetterItem{{'a', "接任務"}, {'b', "租船"}, {'c', "情報"}},
	}
	cs := NewTouchLayout(km).Controls()
	// 4 方向 + Confirm + Cancel + 3 字母 = 9
	if len(cs) != 9 {
		t.Fatalf("控制數 = %d, want 9", len(cs))
	}
}

func TestTouchLayout_ResolveDpad(t *testing.T) {
	l := NewTouchLayout(Keymap{Directions: true})
	// 點在「上」鈕矩形內(26..46, 150..166)
	if a := l.Resolve(Tap{X: 34, Y: 156}); a.Kind != ActUp {
		t.Errorf("點上鈕 → %d, want ActUp", a.Kind)
	}
	// 點在「右」鈕(48..68, 166..182)
	if a := l.Resolve(Tap{X: 58, Y: 172}); a.Kind != ActRight {
		t.Errorf("點右鈕 → %d, want ActRight", a.Kind)
	}
	// 點空白處 → None
	if a := l.Resolve(Tap{X: 160, Y: 100}); a.Kind != ActNone {
		t.Errorf("點空白 → %d, want None", a.Kind)
	}
}

func TestTouchLayout_ResolveLetter(t *testing.T) {
	km := Keymap{Letters: []LetterItem{{'a', "接任務"}, {'b', "租船"}}}
	l := NewTouchLayout(km)
	// 第一顆字母鈕在 x=82..112, y=150..170
	a := l.Resolve(Tap{X: 90, Y: 158})
	if a.Kind != ActLetter || a.Rune != 'a' {
		t.Errorf("點第一顆字母 → %+v, want ActLetter 'a'", a)
	}
	// 第二顆 x=114..144
	b := l.Resolve(Tap{X: 120, Y: 158})
	if b.Kind != ActLetter || b.Rune != 'b' {
		t.Errorf("點第二顆字母 → %+v, want ActLetter 'b'", b)
	}
}

func TestTouchLayout_ConfirmCancel(t *testing.T) {
	l := NewTouchLayout(Keymap{Confirm: "OK", Cancel: "返回"})
	if a := l.Resolve(Tap{X: 300, Y: 178}); a.Kind != ActConfirm {
		t.Errorf("點 A 鈕 → %d, want ActConfirm", a.Kind)
	}
	if a := l.Resolve(Tap{X: 268, Y: 178}); a.Kind != ActCancel {
		t.Errorf("點 B 鈕 → %d, want ActCancel", a.Kind)
	}
}
