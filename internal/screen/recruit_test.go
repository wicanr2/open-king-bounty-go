package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// newTestRecruitScreen 用真實(embedded)assets 建一個騎士角色與 RecruitScreen(棲地
// 洲0/id0,rtype 0),供本檔測試共用。RecruitScreen 現在依 (cont,dwellingID) 即時查
// gs.DwellingTroop/DwellingPopulation(見 recruit.go),world gen 撒鹽出來的兵種是
// 隨機的;測試需要固定兵種 0(農夫)才能斷言造價/庫存等數字,故直接覆寫這兩格
// 世界狀態,不依賴 salt_continent 撒出的實際結果(那部分由 gamestate 自己的
// worldgen_test.go 驗證)。
func newTestRecruitScreen(t *testing.T) (*RecruitScreen, *gamestate.GameState, *kbdata.Assets) {
	t.Helper()
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatalf("kbdata.Load: %v", err)
	}
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.DwellingTroop[0][0] = 0
	gs.DwellingPopulation[0][0] = 999
	return NewRecruitScreen(gs, a, 0, 0, 0), gs, a
}

// TestRecruitScreen_QuantityClampedToCap 驗證數量調整夾在 [0, cap()]:
// 一路往上加會停在上限、不會超出;一路往下減會停在 0、不會變負。
func TestRecruitScreen_QuantityClampedToCap(t *testing.T) {
	s, _, _ := newTestRecruitScreen(t)
	max := s.cap()
	if max <= 0 {
		t.Fatalf("測試前提:cap() = %d,需要 > 0 才能驗證上限夾制(檢查起手金/GoldCost/領導力)", max)
	}

	// 往上加超過上限次數,應停在 max。
	for i := 0; i < max+10; i++ {
		s.Update(input.Action{Kind: input.ActUp})
	}
	if s.sel != max {
		t.Fatalf("加到頂後 sel = %d, want %d(cap)", s.sel, max)
	}

	// 往下減超過次數,應停在 0。
	for i := 0; i < max+10; i++ {
		s.Update(input.Action{Kind: input.ActDown})
	}
	if s.sel != 0 {
		t.Fatalf("減到底後 sel = %d, want 0", s.sel)
	}
}

// TestRecruitScreen_ConfirmBuysTroop 驗證 Confirm 會呼叫 BuyTroop,金錢與隊伍隨之變化。
func TestRecruitScreen_ConfirmBuysTroop(t *testing.T) {
	s, gs, a := newTestRecruitScreen(t)

	goldBefore := gs.Gold
	s.Update(input.Action{Kind: input.ActUp}) // sel = 1
	if s.sel != 1 {
		t.Fatalf("按一次 ActUp 後 sel = %d, want 1", s.sel)
	}

	tr := s.Update(input.Action{Kind: input.ActConfirm})
	if tr.Kind != KindStay {
		t.Fatalf("Confirm 後 Kind = %v, want KindStay(留在招兵畫面看結果)", tr.Kind)
	}

	cost := a.Troops[0].GoldCost
	if gs.Gold != goldBefore-cost {
		t.Errorf("Gold = %d, want %d(扣掉 %d)", gs.Gold, goldBefore-cost, cost)
	}

	found := false
	for _, sq := range gs.Army {
		if sq.TroopID == 0 && sq.Count >= 1 {
			found = true
		}
	}
	if !found {
		t.Errorf("Army 中找不到已招募的兵種 0: %+v", gs.Army)
	}

	if s.msg != "已招募" {
		t.Errorf("msg = %q, want \"已招募\"", s.msg)
	}
	if s.sel != 0 {
		t.Errorf("成功招募後 sel 應重置為 0,實際 = %d", s.sel)
	}
}

// TestRecruitScreen_ConfirmZeroDoesNotBuy 驗證數量為 0 時 Confirm 不會呼叫 BuyTroop
// (不動金錢/隊伍),只顯示提示。
func TestRecruitScreen_ConfirmZeroDoesNotBuy(t *testing.T) {
	s, gs, _ := newTestRecruitScreen(t)
	goldBefore := gs.Gold

	s.Update(input.Action{Kind: input.ActConfirm})

	if gs.Gold != goldBefore {
		t.Errorf("數量 0 時 Confirm 不應扣錢,Gold = %d, want %d", gs.Gold, goldBefore)
	}
	if s.msg == "" {
		t.Errorf("數量 0 時 Confirm 應顯示提示,實際為空字串")
	}
}

// TestRecruitScreen_CancelPops 驗證 ActCancel 回傳 KindPop(回上一畫面/離開招兵)。
func TestRecruitScreen_CancelPops(t *testing.T) {
	s, _, _ := newTestRecruitScreen(t)
	tr := s.Update(input.Action{Kind: input.ActCancel})
	if tr.Kind != KindPop {
		t.Fatalf("Update(ActCancel).Kind = %v, want KindPop", tr.Kind)
	}
}

// TestRecruitScreen_Keymap 驗證 Keymap 宣告方向鍵、確認(招募)、取消(離開)。
func TestRecruitScreen_Keymap(t *testing.T) {
	s, _, _ := newTestRecruitScreen(t)
	km := s.Keymap()
	if !km.Directions {
		t.Error("Keymap.Directions 應為 true")
	}
	if km.Confirm != "招募" {
		t.Errorf("Keymap.Confirm = %q, want \"招募\"", km.Confirm)
	}
	if km.Cancel != "離開" {
		t.Errorf("Keymap.Cancel = %q, want \"離開\"", km.Cancel)
	}
}
