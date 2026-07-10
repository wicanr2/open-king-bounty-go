package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// TestSelectGateScreen_OnlyListsVisitedCastles 驗證城堡傳送選單只列已造訪的城堡,
// 且 entry id 為城堡的絕對索引(非重新編號)。
func TestSelectGateScreen_OnlyListsVisitedCastles(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.CastleVisited[2] = 1
	gs.CastleVisited[5] = 1

	s := NewSelectGateScreen(gs, a, gateCastle)
	if len(s.entries) != 2 {
		t.Fatalf("entries 數量 = %d, want 2", len(s.entries))
	}
	ids := map[int]bool{s.entries[0].id: true, s.entries[1].id: true}
	if !ids[2] || !ids[5] {
		t.Errorf("entries id 錯誤:%+v, want {2,5}", s.entries)
	}
}

// TestSelectGateScreen_SelectCastleTeleportsAndPops 驗證選定已造訪城堡後傳送並 Pop。
func TestSelectGateScreen_SelectCastleTeleportsAndPops(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.CastleVisited[2] = 1

	s := NewSelectGateScreen(gs, a, gateCastle)
	tr := s.Update(input.Letter(rune('a' + 2)))
	if tr.Kind != KindPop {
		t.Fatalf("選定已造訪城堡後 Update.Kind = %v, want KindPop", tr.Kind)
	}

	c := gamestate.LoadCastles(a)[2]
	if gs.Continent != c.Continent {
		t.Errorf("傳送後 Continent = %d, want %d", gs.Continent, c.Continent)
	}
}

// TestSelectGateScreen_CancelPops 驗證 ESC/ActCancel 直接 Pop,不傳送。
func TestSelectGateScreen_CancelPops(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)

	s := NewSelectGateScreen(gs, a, gateCastle)
	tr := s.Update(input.Action{Kind: input.ActCancel})
	if tr.Kind != KindPop {
		t.Fatalf("ActCancel 應 Pop,got %v", tr.Kind)
	}
}
