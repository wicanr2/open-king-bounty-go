package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// TestChooseSpellScreen_CastLearnedSpell 驗證按對應字母施放已學會的冒險法術(此處為
// 'b' → 冒險索引 1 → spell id 8 = 停時術):應留在畫面(KindStay)、切到結果 mode、
// 且消耗該法術。
func TestChooseSpellScreen_CastLearnedSpell(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.Spells[8] = 1

	s := NewChooseSpellScreen(gs, a)
	tr := s.Update(input.Letter('b'))

	if tr.Kind != KindStay {
		t.Fatalf("按 'b' 施放後 Kind = %v, want KindStay", tr.Kind)
	}
	if s.mode != chooseSpellResult {
		t.Fatalf("按 'b' 施放後 mode = %v, want chooseSpellResult", s.mode)
	}
	if gs.Spells[8] != 0 {
		t.Errorf("施放後應消耗法術,Spells[8] = %d, want 0", gs.Spells[8])
	}
}

// TestChooseSpellScreen_CancelPops 驗證選單狀態下 ActCancel 回傳 KindPop(離開畫面)。
func TestChooseSpellScreen_CancelPops(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)

	s := NewChooseSpellScreen(gs, a)
	tr := s.Update(input.Action{Kind: input.ActCancel})

	if tr.Kind != KindPop {
		t.Fatalf("Update(ActCancel).Kind = %v, want KindPop", tr.Kind)
	}
}
