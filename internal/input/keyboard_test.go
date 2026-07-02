package input

import "testing"

func TestKeyToAction_Symbols(t *testing.T) {
	cases := []struct {
		sym  Sym
		want ActionKind
	}{
		{SymUp, ActUp}, {SymDown, ActDown}, {SymLeft, ActLeft}, {SymRight, ActRight},
		{SymEnter, ActConfirm}, {SymSpace, ActConfirm}, {SymEsc, ActCancel},
		{SymPageUp, ActPageUp}, {SymPageDown, ActPageDown}, {SymHome, ActHome}, {SymEnd, ActEnd},
		{SymF8, ActThemeCycle}, {SymF9, ActMusicToggle}, {SymF10, ActQuitSave},
		{SymNone, ActNone},
	}
	for _, c := range cases {
		if got := KeyToAction(Key{Sym: c.sym}); got.Kind != c.want {
			t.Errorf("sym %d: got %d, want %d", c.sym, got.Kind, c.want)
		}
	}
}

func TestKeyToAction_Chars(t *testing.T) {
	// 字母 → ActLetter(小寫化)
	if a := KeyToAction(Key{Sym: SymChar, Ch: 'A'}); a.Kind != ActLetter || a.Rune != 'a' {
		t.Errorf("'A' → %+v, want ActLetter 'a'", a)
	}
	if a := KeyToAction(Key{Sym: SymChar, Ch: 'c'}); a.Kind != ActLetter || a.Rune != 'c' {
		t.Errorf("'c' → %+v", a)
	}
	// y/n → 是非(優先於字母)
	if a := KeyToAction(Key{Sym: SymChar, Ch: 'Y'}); a.Kind != ActYes {
		t.Errorf("'Y' → %+v, want ActYes", a)
	}
	if a := KeyToAction(Key{Sym: SymChar, Ch: 'n'}); a.Kind != ActNo {
		t.Errorf("'n' → %+v, want ActNo", a)
	}
	// 數字 → ActDigit
	if a := KeyToAction(Key{Sym: SymChar, Ch: '7'}); a.Kind != ActDigit || a.Rune != '7' {
		t.Errorf("'7' → %+v, want ActDigit '7'", a)
	}
}

func TestKeymap_TownMenuDeclaration(t *testing.T) {
	// 城鎮選單:方向不需要、確認/取消有、情境字母列 A–E(對齊 ui-design.md 範例)
	km := Keymap{
		Confirm: "選擇",
		Cancel:  "離開",
		Letters: []LetterItem{
			{'a', "接任務"}, {'b', "租船"}, {'c', "情報"}, {'d', "造橋"}, {'e', "買攻城"},
		},
	}
	if len(km.Letters) != 5 {
		t.Fatalf("城鎮字母列應 5 顆, got %d", len(km.Letters))
	}
	if km.Letters[0].Rune != 'a' || km.Letters[0].Label != "接任務" {
		t.Errorf("首項應 [a 接任務], got %+v", km.Letters[0])
	}
	// 點該字母鈕 = 送出對應 ActLetter
	if got := Letter(km.Letters[0].Rune); got.Kind != ActLetter || got.Rune != 'a' {
		t.Errorf("點 [a] → %+v, want ActLetter 'a'", got)
	}
}
