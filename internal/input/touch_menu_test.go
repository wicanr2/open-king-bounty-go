package input

import "testing"

// 選單型畫面(如城鎮)的觸控佈局:無移動 D-pad、字母是隱形熱區疊在選單各行,
// 點該行即命中該字母。對齊修正「觸控疊層蓋住 A-E 選單文字」的問題。
func TestMenuScreenTouchLayout(t *testing.T) {
	rects := []Rect{{24, 145, 220, 8}, {24, 153, 220, 8}, {24, 161, 220, 8}, {24, 169, 220, 8}, {24, 177, 220, 8}}
	km := Keymap{
		Directions:  false, // 選單畫面不放 D-pad
		Confirm:     "選擇",
		Cancel:      "離開",
		Letters:     []LetterItem{{'a', "A"}, {'b', "B"}, {'c', "C"}, {'d', "D"}, {'e', "E"}},
		LetterRects: rects,
	}
	cs := NewTouchLayout(km).Controls()

	for _, c := range cs {
		switch c.Action.Kind {
		case ActUp, ActDown, ActLeft, ActRight:
			t.Fatalf("選單畫面不該有 D-pad 控制,卻出現 %v", c)
		}
	}
	var letters []Control
	for _, c := range cs {
		if c.Action.Kind == ActLetter {
			letters = append(letters, c)
		}
	}
	if len(letters) != 5 {
		t.Fatalf("want 5 letter controls, got %d", len(letters))
	}
	for i, c := range letters {
		if !c.Hidden {
			t.Errorf("letter %d 應為隱形熱區(Hidden),卻是可見", i)
		}
		if c.Rect != rects[i] {
			t.Errorf("letter %d rect=%v, want %v(疊在選單行)", i, c.Rect, rects[i])
		}
	}
	// 點 A 行(y 落在 rects[0])→ 命中字母 a
	if a := NewTouchLayout(km).Resolve(Tap{100, 148}); a.Kind != ActLetter || a.Rune != 'a' {
		t.Errorf("點 A 選單行應命中字母 a,got %+v", a)
	}
	// 對照:預設(無 LetterRects)字母走可見按鈕列
	km2 := Keymap{Letters: km.Letters}
	for _, c := range NewTouchLayout(km2).Controls() {
		if c.Action.Kind == ActLetter && c.Hidden {
			t.Error("無 LetterRects 時字母不該是隱形")
		}
	}
}
