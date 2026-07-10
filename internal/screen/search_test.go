package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// TestSearch_AtScepterWins 驗證站在權杖處搜索(y)→ Replace 成 WinScreen。
func TestSearch_AtScepterWins(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	// 把玩家放到權杖埋藏處。
	gs.Continent, gs.X, gs.Y = gs.ScepterContinent, gs.ScepterX, gs.ScepterY

	s := NewSearchScreen(gs, a)
	tr := s.Update(input.Letter('y'))
	if tr.Kind != KindReplace {
		t.Fatalf("找到權杖搜索應 Replace 成破關畫面,got %v", tr.Kind)
	}
	if _, ok := tr.Next.(*WinScreen); !ok {
		t.Errorf("破關畫面型別 = %T, want *WinScreen", tr.Next)
	}
}

// TestSearch_NotAtScepterFutile 驗證非權杖處搜索(y)→ 一無所獲(留在畫面),再按鍵回地圖。
func TestSearch_NotAtScepterFutile(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	// 確保玩家不在權杖處。
	gs.Continent, gs.X, gs.Y = gs.ScepterContinent, gs.ScepterX+1, gs.ScepterY
	if gs.AtScepter() {
		gs.X += 2
	}

	s := NewSearchScreen(gs, a)
	if tr := s.Update(input.Letter('y')); tr.Kind != KindStay {
		t.Fatalf("非權杖處搜索應留在畫面顯示結果,got %v", tr.Kind)
	}
	if s.mode != searchFutile {
		t.Errorf("應進一無所獲狀態,got %v", s.mode)
	}
	if tr := s.Update(input.Action{Kind: input.ActConfirm}); tr.Kind != KindPop {
		t.Errorf("一無所獲後任意鍵應 Pop,got %v", tr.Kind)
	}
}

// TestSearch_CancelPops 驗證搜索詢問按 n/Cancel → Pop。
func TestSearch_CancelPops(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	s := NewSearchScreen(gs, a)
	if tr := s.Update(input.Letter('n')); tr.Kind != KindPop {
		t.Errorf("按 n 應 Pop,got %v", tr.Kind)
	}
}
