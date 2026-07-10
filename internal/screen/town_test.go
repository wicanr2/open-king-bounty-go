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
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	return NewTownScreen(gs, a, 0)
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
		{Rune: 'd', Label: "學法術"},
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
// 且 A/B/E 三項(需世界狀態,尚未建模)切到「未實作」stub 狀態,不誤觸情報分支。
// D(學法術)已接上真實邏輯,不在本測試涵蓋範圍(見 TestTownScreen_LetterD_*)。
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

	// B(租船)/E(攻城)仍是佔位 stub(需 boat/siege 世界狀態);A 已接契約、
	// C 情報、D 學法術,不在此 stub 迴圈。
	for i, r := range []rune{'b', 'e'} {
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

// TestTownScreen_LetterA_TakesContract 驗證 A(領取新契約)接 next_contract:
// 起手 ContractCycle={0,1,2,3,4}/LastContract=4 → 領到 villain 0,設定
// gs.Contract/LastContract 並 Push view_contract(對齊 C visit_town key==1)。
func TestTownScreen_LetterA_TakesContract(t *testing.T) {
	s := newTestTownScreen(t)
	tr := s.Update(input.Letter('a'))
	if s.gs.Contract != 0 || s.gs.LastContract != 0 {
		t.Errorf("領契約後 Contract=%d LastContract=%d, want 0/0", s.gs.Contract, s.gs.LastContract)
	}
	if tr.Kind != KindPush {
		t.Errorf("領契約應 Push view_contract,got transition kind %v", tr.Kind)
	}
}

// TestTownScreen_LetterD_LearnsSpell 驗證 D(學法術)成功路徑:對齊 C
// visit_town() key==4 成功分支(game.c:2819-2824)—— 該法術 Spells[id] +1、
// Gold 扣掉售價,並切到 townModeLearnResult 顯示「還可以再學習 N 個」。
func TestTownScreen_LetterD_LearnsSpell(t *testing.T) {
	s := newTestTownScreen(t)
	spellID := s.spell.ID
	cost := s.spell.Gold
	goldBefore := s.gs.Gold
	knownBefore := s.gs.KnownSpells()

	tr := s.Update(input.Letter('d'))
	if tr.Kind != KindStay {
		t.Fatalf("按 'd' 應留在城鎮畫面, Kind = %v", tr.Kind)
	}
	if s.mode != townModeLearnResult {
		t.Fatalf("按 'd' 後 mode = %v, want townModeLearnResult", s.mode)
	}
	if s.gs.Spells[spellID] != 1 {
		t.Errorf("Spells[%d] = %d, want 1", spellID, s.gs.Spells[spellID])
	}
	if s.gs.Gold != goldBefore-cost {
		t.Errorf("Gold = %d, want %d(扣除法術售價 %d)", s.gs.Gold, goldBefore-cost, cost)
	}
	if s.gs.KnownSpells() != knownBefore+1 {
		t.Errorf("KnownSpells() = %d, want %d", s.gs.KnownSpells(), knownBefore+1)
	}
	if len(s.learnMsg) == 0 {
		t.Fatal("learnMsg 為空,應顯示學習結果")
	}
}

// TestTownScreen_LetterD_NotEnoughGold 驗證金幣不足時擋下購買且不扣錢/不加法術,
// 對齊 C game->gold <= spell_costs[...] 分支(game.c:2812-2814)。
func TestTownScreen_LetterD_NotEnoughGold(t *testing.T) {
	s := newTestTownScreen(t)
	s.gs.Gold = 0 // 強制金幣不足(法術售價恆 > 0)

	s.Update(input.Letter('d'))

	if s.mode != townModeLearnResult {
		t.Fatalf("mode = %v, want townModeLearnResult", s.mode)
	}
	if s.gs.KnownSpells() != 0 {
		t.Errorf("KnownSpells() = %d, want 0(金幣不足不應學會)", s.gs.KnownSpells())
	}
	if s.gs.Gold != 0 {
		t.Errorf("Gold = %d, want 0(不應被扣錢)", s.gs.Gold)
	}
}

// TestTownScreen_LetterD_SpellLimitReached 驗證已達法術上限時擋下購買,
// 對齊 C known_spells(game) >= game->max_spells 分支(game.c:2808-2811)。
func TestTownScreen_LetterD_SpellLimitReached(t *testing.T) {
	s := newTestTownScreen(t)
	s.gs.Gold = 1_000_000 // 排除金幣不足的干擾
	for i := range s.gs.Spells {
		s.gs.Spells[i] = 1
	}
	knownBefore := s.gs.KnownSpells()
	if knownBefore < s.gs.MaxSpells {
		t.Fatalf("測試前提不成立:KnownSpells()=%d 應 >= MaxSpells=%d", knownBefore, s.gs.MaxSpells)
	}

	s.Update(input.Letter('d'))

	if s.mode != townModeLearnResult {
		t.Fatalf("mode = %v, want townModeLearnResult", s.mode)
	}
	if s.gs.KnownSpells() != knownBefore {
		t.Errorf("KnownSpells() 應維持 %d(已達上限不應再學),got %d", knownBefore, s.gs.KnownSpells())
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
