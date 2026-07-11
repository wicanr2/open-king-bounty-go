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
	// 系統鍵已移出全域 Controls(改進 dialog):4 方向 + Confirm + Cancel + 3 字母 = 9
	if len(cs) != 9 {
		t.Fatalf("控制數 = %d, want 9", len(cs))
	}
	// 不該再出現全域系統鍵(主題/作弊/音樂)
	for _, c := range cs {
		switch c.Action.Kind {
		case ActThemeCycle, ActCheat, ActMusicToggle:
			t.Errorf("系統鍵不應再全域浮出,卻出現 %v", c)
		}
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

// TestTouchLayout_MixedOverlayAndBottomRow 驗證「部分字母疊 sidebar(Hidden 自訂矩形)、
// 部分留下方可見主排」的混用:LetterRects[i].W>0 → 隱形疊層;為零值 → 下方主排,且下方
// 主排用獨立計數排位(不因前面被疊走的字母而留空)。對齊世界地圖新佈局。
func TestTouchLayout_MixedOverlayAndBottomRow(t *testing.T) {
	km := Keymap{
		Directions: true,
		Letters: []LetterItem{
			{'s', "存檔"}, {'l', "讀檔"}, // 下方主排
			{'c', "角色"}, // 疊到 sidebar 錢袋面板
		},
		LetterRects: []Rect{{}, {}, {X: 256, Y: 157, W: 48, H: 34}},
	}
	l := NewTouchLayout(km)

	var letters []Control
	for _, c := range l.Controls() {
		if c.Action.Kind == ActLetter {
			letters = append(letters, c)
		}
	}
	if len(letters) != 3 {
		t.Fatalf("字母控制數 = %d, want 3", len(letters))
	}
	// s,l:可見、下方主排,索引 0/1 → x=82/114(獨立計數,不受 c 被疊走影響)
	if letters[0].Hidden || letters[0].Rect.X != 82 {
		t.Errorf("s 應為下方主排可見鈕 x=82,got %+v", letters[0])
	}
	if letters[1].Hidden || letters[1].Rect.X != 114 {
		t.Errorf("l 應為下方主排可見鈕 x=114,got %+v", letters[1])
	}
	// c:隱形疊層,矩形 = 錢袋面板
	if !letters[2].Hidden || (letters[2].Rect != Rect{X: 256, Y: 157, W: 48, H: 34}) {
		t.Errorf("c 應為隱形疊在錢袋面板,got %+v", letters[2])
	}
	// 點錢袋面板 → 命中 'c'
	if a := l.Resolve(Tap{X: 280, Y: 170}); a.Kind != ActLetter || a.Rune != 'c' {
		t.Errorf("點錢袋面板 → %+v, want ActLetter 'c'", a)
	}
	// 點下方第一顆 → 命中 's'
	if a := l.Resolve(Tap{X: 90, Y: 158}); a.Kind != ActLetter || a.Rune != 's' {
		t.Errorf("點下方第一顆 → %+v, want ActLetter 's'", a)
	}
}

// TestTouchLayout_Buttons 驗證畫面自訂按鈕(km.Buttons)會原樣附加、可被 Resolve 命中,
// 且 Hidden 的按鈕仍能命中(供 dialog 自繪視覺 + 隱形熱區收命中)。
func TestTouchLayout_Buttons(t *testing.T) {
	km := Keymap{
		Buttons: []Control{
			{Rect{200, 1, 52, 14}, Action{Kind: ActMenu}, "選單", false}, // 可見叫出鈕
			{Rect{0, 0, 320, 200}, Action{Kind: ActCancel}, "", true},  // 隱形全螢幕關閉熱區
		},
	}
	l := NewTouchLayout(km)
	// 點叫出鈕 → ActMenu
	if a := l.Resolve(Tap{X: 220, Y: 7}); a.Kind != ActMenu {
		t.Errorf("點叫出鈕 → %d, want ActMenu", a.Kind)
	}
	// 點叫出鈕外(但在全螢幕熱區內)→ ActCancel(隱形熱區仍命中)
	if a := l.Resolve(Tap{X: 10, Y: 100}); a.Kind != ActCancel {
		t.Errorf("點叫出鈕外 → %d, want ActCancel(全螢幕關閉熱區)", a.Kind)
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
