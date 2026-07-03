package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// newTestTownScreen 用真實(embedded)assets 建一個騎士角色與 TownScreen,供本檔測試共用。
func newTestTownScreen(t *testing.T) *TownScreen {
	t.Helper()
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatalf("kbdata.Load: %v", err)
	}
	gs := gamestate.NewGame(a, "Tester", 0)
	return NewTownScreen(gs, a)
}

// TestTownScreen_KeymapLetters 驗證情境字母列有 5 顆且標籤正確,對應 C 版
// game.c visit_town() 的 five_choices(A 接任務/B 租船/C 情報/D 學法術/E 買攻城)。
func TestTownScreen_KeymapLetters(t *testing.T) {
	s := newTestTownScreen(t)
	km := s.Keymap()

	want := []input.LetterItem{
		{Rune: 'a', Label: "接任務"},
		{Rune: 'b', Label: "租船"},
		{Rune: 'c', Label: "情報"},
		{Rune: 'd', Label: "造橋"},
		{Rune: 'e', Label: "買攻城"},
	}
	if len(km.Letters) != len(want) {
		t.Fatalf("Letters 數量 = %d, want %d", len(km.Letters), len(want))
	}
	for i, got := range km.Letters {
		if got != want[i] {
			t.Errorf("Letters[%d] = %+v, want %+v", i, got, want[i])
		}
	}
}

// TestTownScreen_CancelPops 驗證 ActCancel 回傳 KindPop(回上一畫面),
// 而非留在城鎮或 Replace。
func TestTownScreen_CancelPops(t *testing.T) {
	s := newTestTownScreen(t)
	tr := s.Update(input.Action{Kind: input.ActCancel})
	if tr.Kind != KindPop {
		t.Fatalf("Update(ActCancel).Kind = %v, want KindPop", tr.Kind)
	}
}

// TestTownScreen_LetterC_SwitchesToInfo 驗證按 'c'(情報)後畫面內部狀態切到情報顯示,
// 且其餘四項(A/B/D/E)切到「未實作」stub 狀態,不誤觸情報分支。
func TestTownScreen_LetterC_SwitchesToInfo(t *testing.T) {
	s := newTestTownScreen(t)
	if s.mode != townModeMenu {
		t.Fatalf("初始 mode = %v, want townModeMenu", s.mode)
	}

	tr := s.Update(input.Letter('c'))
	if tr.Kind != KindStay {
		t.Fatalf("按 'c' 應留在城鎮畫面, Kind = %v", tr.Kind)
	}
	if s.mode != townModeInfo {
		t.Fatalf("按 'c' 後 mode = %v, want townModeInfo", s.mode)
	}
	if s.sel != 2 {
		t.Errorf("按 'c' 後 sel = %d, want 2", s.sel)
	}

	for i, r := range []rune{'a', 'b', 'd', 'e'} {
		s2 := newTestTownScreen(t)
		s2.Update(input.Letter(r))
		if s2.mode != townModeStub {
			t.Errorf("按 %q(第 %d 項)mode = %v, want townModeStub", r, i, s2.mode)
		}
		if s2.stub != townLetters[indexOfRune(r)].Label {
			t.Errorf("按 %q 後 stub = %q, want %q", r, s2.stub, townLetters[indexOfRune(r)].Label)
		}
	}
}

func indexOfRune(r rune) int {
	for i, li := range townLetters {
		if li.Rune == r {
			return i
		}
	}
	return -1
}

// TestTownScreen_DirectionsConfirm 驗證方向鍵 + Confirm 也能觸發對應項(不只字母捷徑),
// 對齊 charselect.go 既有的雙輸入模式(字母直選 / 方向+Confirm)。
func TestTownScreen_DirectionsConfirm(t *testing.T) {
	s := newTestTownScreen(t)
	for i := 0; i < 2; i++ {
		s.Update(input.Action{Kind: input.ActDown})
	}
	if s.sel != 2 {
		t.Fatalf("按兩次 ActDown 後 sel = %d, want 2", s.sel)
	}
	s.Update(input.Action{Kind: input.ActConfirm})
	if s.mode != townModeInfo {
		t.Fatalf("方向移到情報項後 Confirm, mode = %v, want townModeInfo", s.mode)
	}
}
