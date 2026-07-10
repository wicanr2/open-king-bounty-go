package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// TestAlcoveScreen_AcceptWithEnoughGold 驗證金幣足時按 y 接受 → Stay 顯示結果、習得魔法。
func TestAlcoveScreen_AcceptWithEnoughGold(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.Gold = 6000

	s := NewAlcoveScreen(gs, a)
	tr := s.Update(input.Letter('y'))
	if tr.Kind != KindStay {
		t.Fatalf("接受且金幣足應 Stay(顯示結果),got %v", tr.Kind)
	}
	if s.mode != alcoveResult {
		t.Errorf("mode = %v, want alcoveResult", s.mode)
	}
	if !gs.KnowsMagic {
		t.Errorf("KnowsMagic = false, want true(已習得魔法)")
	}
}

// TestAlcoveScreen_Decline 驗證按 n 拒絕 → Pop。
func TestAlcoveScreen_Decline(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)

	s := NewAlcoveScreen(gs, a)
	if tr := s.Update(input.Letter('n')); tr.Kind != KindPop {
		t.Errorf("按 n 應 Pop,got %v", tr.Kind)
	}
}

// TestAlcoveScreen_AnyActionAfterResultPops 驗證顯示結果後任意動作都會 Pop。
func TestAlcoveScreen_AnyActionAfterResultPops(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.Gold = 6000

	s := NewAlcoveScreen(gs, a)
	s.Update(input.Letter('y')) // 接受 → 進 alcoveResult
	if tr := s.Update(input.Action{Kind: input.ActConfirm}); tr.Kind != KindPop {
		t.Errorf("結果畫面後任意動作應 Pop,got %v", tr.Kind)
	}
}
