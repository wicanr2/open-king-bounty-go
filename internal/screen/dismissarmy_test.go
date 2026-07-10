package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// TestDismissArmy_NonLastDismisses 驗證解散非最後一支部隊:直接移除並前移,Pop。
func TestDismissArmy_NonLastDismisses(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.Army = [5]gamestate.Squad{{TroopID: 3, Count: 10}, {TroopID: 7, Count: 20}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}}

	s := NewDismissArmyScreen(gs, a)
	tr := s.Update(input.Letter('a')) // 解散第 0 格
	if tr.Kind != KindPop {
		t.Fatalf("解散後應 Pop,got %v", tr.Kind)
	}
	if gs.Army[0].TroopID != 7 || gs.Army[0].Count != 20 {
		t.Errorf("解散後未前移:Army[0]=%+v", gs.Army[0])
	}
}

// TestDismissArmy_LastNeedsConfirm 驗證解散唯一一支部隊需確認,y 後隊伍清空→temp_death。
func TestDismissArmy_LastNeedsConfirm(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.Army = [5]gamestate.Squad{{TroopID: 3, Count: 10}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}}

	s := NewDismissArmyScreen(gs, a)
	// 選唯一部隊 → 進確認,不立即解散。
	if tr := s.Update(input.Letter('a')); tr.Kind != KindStay {
		t.Fatalf("解散最後部隊應先確認(Stay),got %v", tr.Kind)
	}
	if s.mode != dismissModeConfirmLast {
		t.Fatalf("應進確認模式,got %v", s.mode)
	}
	if gs.Army[0].Count != 10 {
		t.Errorf("確認前不應解散:Army[0]=%+v", gs.Army[0])
	}
	// 確認 y → 解散 → temp_death(隊伍清空後補 20 農夫)。
	if tr := s.Update(input.Letter('y')); tr.Kind != KindPop {
		t.Fatalf("確認後應 Pop,got %v", tr.Kind)
	}
	if gs.Army[0].TroopID != 0 || gs.Army[0].Count != 20 {
		t.Errorf("解散最後部隊應 temp_death 補 20 農夫,got %+v", gs.Army[0])
	}
}

// TestDismissArmy_ConfirmNoKeeps 驗證確認時按 n 保留部隊。
func TestDismissArmy_ConfirmNoKeeps(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.Army = [5]gamestate.Squad{{TroopID: 3, Count: 10}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}}

	s := NewDismissArmyScreen(gs, a)
	s.Update(input.Letter('a'))
	if tr := s.Update(input.Letter('n')); tr.Kind != KindPop {
		t.Fatalf("取消應 Pop,got %v", tr.Kind)
	}
	if gs.Army[0].TroopID != 3 || gs.Army[0].Count != 10 {
		t.Errorf("取消後部隊應保留:Army[0]=%+v", gs.Army[0])
	}
}
