package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/combat"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

func newTestCombatScreen(t *testing.T) *CombatScreen {
	t.Helper()
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatal(err)
	}
	gs := gamestate.NewGame(a, "P", 0, gamestate.DefaultWorldSeed) // 騎士:義勇軍/弓箭手
	return NewDebugCombatScreen(gs, a)
}

// 佈陣後玩家側與敵方側都有單位。
func TestCombatScreen_Setup(t *testing.T) {
	s := newTestCombatScreen(t)
	pc := 0
	for i := 0; i < combat.MaxUnits; i++ {
		if s.combat.Units[0][i].Count > 0 {
			pc++
		}
	}
	if pc == 0 {
		t.Fatal("玩家側應有單位")
	}
	if s.combat.Units[1][0].Count == 0 {
		t.Fatal("敵方側應有單位(野狼)")
	}
}

// 玩家回合按撤退(ActCancel)→ Pop 回上一畫面。
func TestCombatScreen_FleePops(t *testing.T) {
	s := newTestCombatScreen(t)
	s.combat.Side = 0 // 確保玩家回合
	tr := s.Update(input.Action{Kind: input.ActCancel})
	if tr.Kind != KindPop {
		t.Fatalf("撤退應回傳 KindPop, got %d", tr.Kind)
	}
}
