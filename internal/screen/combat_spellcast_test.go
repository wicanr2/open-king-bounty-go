package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/combat"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// newMagicCombat 建一個「玩家會魔法」的戰鬥畫面(女巫師,class 2 → KnowsMagic)。
func newMagicCombat(t *testing.T) *CombatScreen {
	t.Helper()
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "巫師", 2, gamestate.DefaultWorldSeed)
	if !gs.KnowsMagic {
		t.Fatal("女巫師應會魔法")
	}
	var foe [combat.MaxUnits]gamestate.Squad
	foe[0] = gamestate.Squad{TroopID: 6, Count: 20} // 地精(非不死非免疫)
	for i := 1; i < combat.MaxUnits; i++ {
		foe[i] = gamestate.Squad{TroopID: 255}
	}
	return NewCombatScreen(gs, a, foe, 1)
}

// TestCombatSpell_OpenMenuGates 驗證施法選單的前置檢查:不會魔法→提示;會魔法→開選單。
func TestCombatSpell_OpenMenuGates(t *testing.T) {
	s := newMagicCombat(t)
	s.openSpellMenu()
	if s.spellState != spellMenu {
		t.Fatalf("會魔法且未施法應開選單,got %v", s.spellState)
	}

	// 本回合已施法 → 擋下。
	s2 := newMagicCombat(t)
	s2.combat.Spells = 1
	s2.openSpellMenu()
	if s2.spellState != spellDone {
		t.Errorf("本回合已施法應提示(spellDone),got %v", s2.spellState)
	}

	// 不會魔法 → 提示。
	a := castleTestAssets(t)
	gsKnight := gamestate.NewGame(a, "騎士", 0, gamestate.DefaultWorldSeed)
	var foe [combat.MaxUnits]gamestate.Squad
	for i := range foe {
		foe[i] = gamestate.Squad{TroopID: 255}
	}
	s3 := NewCombatScreen(gsKnight, a, foe, 1)
	s3.openSpellMenu()
	if s3.spellState != spellDone {
		t.Errorf("不會魔法應提示(spellDone),got %v", s3.spellState)
	}
}

// TestCombatSpell_SelectUnlearned 驗證選到未學會的法術 → 提示,不進目標選取。
func TestCombatSpell_SelectUnlearned(t *testing.T) {
	s := newMagicCombat(t)
	for i := range s.playerGS().Spells {
		s.playerGS().Spells[i] = 0 // 全部清空
	}
	s.openSpellMenu()
	s.selectSpell(4) // freeze,未學
	if s.spellState != spellDone {
		t.Errorf("未學法術應提示,got %v", s.spellState)
	}
}

// TestCombatSpell_FreezeConsumesAndFreezes 驗證施放冰凍術:選敵方目標 → 敵軍凍結、消耗法術。
func TestCombatSpell_FreezeConsumesAndFreezes(t *testing.T) {
	s := newMagicCombat(t)
	gs := s.playerGS()
	for i := range gs.Spells {
		gs.Spells[i] = 0
	}
	gs.Spells[4] = 1 // 冰凍術(index 4)

	// 找一個存活的敵方(side 1)單位,把游標移到它。
	var ex, ey int
	found := false
	for i := 0; i < combat.MaxUnits; i++ {
		u := &s.combat.Units[1][i]
		if u.Count > 0 {
			ex, ey = u.X, u.Y
			found = true
			break
		}
	}
	if !found {
		t.Fatal("找不到敵方單位")
	}

	s.openSpellMenu()
	s.selectSpell(4)
	if s.spellState != spellTarget {
		t.Fatalf("選冰凍術後應進目標選取,got %v", s.spellState)
	}
	s.curX, s.curY = ex, ey
	s.resolveTarget()

	if gs.Spells[4] != 0 {
		t.Errorf("施法後應消耗冰凍術,got %d", gs.Spells[4])
	}
	if s.combat.Spells != 1 {
		t.Errorf("本回合施法數應為 1,got %d", s.combat.Spells)
	}
	// 目標敵方單位應被凍結。
	frozen := false
	for i := 0; i < combat.MaxUnits; i++ {
		u := &s.combat.Units[1][i]
		if u.X == ex && u.Y == ey && u.Frozen {
			frozen = true
		}
	}
	if !frozen {
		t.Error("目標敵軍未被凍結")
	}
	if s.spellState != spellDone {
		t.Errorf("施法完成應顯示結果(spellDone),got %v", s.spellState)
	}
}

// TestCombatSpell_TargetOwnRejectedForDamage 驗證傷害類法術選我方目標會被拒(無效)。
func TestCombatSpell_TargetOwnRejectedForDamage(t *testing.T) {
	s := newMagicCombat(t)
	gs := s.playerGS()
	for i := range gs.Spells {
		gs.Spells[i] = 0
	}
	gs.Spells[2] = 1 // 火球術(SpellDamage)

	// 游標放到我方(side 0)單位。
	u := &s.combat.Units[0][s.combat.UnitID]
	s.openSpellMenu()
	s.selectSpell(2)
	s.curX, s.curY = u.X, u.Y
	s.resolveTarget()

	if gs.Spells[2] != 1 {
		t.Errorf("對我方施放傷害術應無效、不消耗,got Spells[2]=%d", gs.Spells[2])
	}
}

// TestCombatSpell_CancelFromMenu 驗證施法選單按 Cancel 回到正常戰鬥。
func TestCombatSpell_CancelFromMenu(t *testing.T) {
	s := newMagicCombat(t)
	s.openSpellMenu()
	s.updateSpell(input.Action{Kind: input.ActCancel})
	if s.spellState != spellNone {
		t.Errorf("選單取消應回正常戰鬥,got %v", s.spellState)
	}
}
