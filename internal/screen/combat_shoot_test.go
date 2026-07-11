package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// newArcherCombatScreen 佈一場玩家首格是弓箭手(troop 8)的戰鬥,對佔位野狼;
// 不自動進射擊模式,供測試一般指令→射擊入口的完整流程。
func newArcherCombatScreen(t *testing.T) *CombatScreen {
	t.Helper()
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatal(err)
	}
	gs := gamestate.NewGame(a, "P", 0, gamestate.DefaultWorldSeed)
	gs.Army[0] = gamestate.Squad{TroopID: 8, Count: 5} // 弓箭手,人數壓在領導力內
	return NewDebugCombatScreen(gs, a)
}

// 一般指令狀態下,可遠攻的當前單位應浮出「射擊」觸控鈕(ActLetter 's'),
// 且不影響原本的待命/撤退。
func TestCombatShoot_KeymapHasShootButton(t *testing.T) {
	s := newArcherCombatScreen(t)
	if !s.canShootCurrent() {
		t.Fatalf("弓箭手當前單位應可射擊")
	}
	km := s.Keymap()
	found := false
	for _, b := range km.Buttons {
		if b.Action.Kind == input.ActLetter && b.Action.Rune == 's' {
			found = true
		}
	}
	if !found {
		t.Errorf("可遠攻時 Keymap.Buttons 應含射擊鈕(ActLetter 's'),got %+v", km.Buttons)
	}
	if km.Confirm != "待命" || km.Cancel != "撤退" {
		t.Errorf("一般指令列應維持 待命/撤退,got Confirm=%q Cancel=%q", km.Confirm, km.Cancel)
	}
}

// 純近戰兵種(無彈藥)不應浮出射擊鈕。
func TestCombatShoot_NoButtonForMelee(t *testing.T) {
	s := newTestCombatScreen(t) // 首格義勇軍(近戰)
	if s.canShootCurrent() {
		t.Fatalf("近戰兵種當前單位不應可射擊")
	}
	for _, b := range s.Keymap().Buttons {
		if b.Action.Kind == input.ActLetter && b.Action.Rune == 's' {
			t.Errorf("近戰兵種不應出現射擊鈕")
		}
	}
}

// 按 's' 進入射擊 target mode:游標落在敵人上、Keymap 切成 射擊/取消。
func TestCombatShoot_EnterTargetMode(t *testing.T) {
	s := newArcherCombatScreen(t)
	s.Update(input.Letter('s'))
	if !s.shooting {
		t.Fatalf("按 's' 應進入射擊 target mode")
	}
	// 游標應落在合法敵人上(預設最近敵人)。
	if _, _, ok := s.shootTargetAt(s.curX, s.curY); !ok {
		t.Errorf("進入射擊時游標(%d,%d)應落在合法敵人上", s.curX, s.curY)
	}
	km := s.Keymap()
	if km.Confirm != "射擊" || km.Cancel != "取消" {
		t.Errorf("射擊模式 Keymap 應為 射擊/取消,got Confirm=%q Cancel=%q", km.Confirm, km.Cancel)
	}
}

// 對游標處敵人 Confirm:扣 1 彈藥、該單位行動結束(Acted)、離開 target mode。
func TestCombatShoot_ConfirmFires(t *testing.T) {
	s := newArcherCombatScreen(t)
	shooterSide, shooterID := s.combat.Side, s.combat.UnitID
	shotsBefore := s.combat.Units[shooterSide][shooterID].Shots

	s.enterShoot()
	if !s.shooting {
		t.Fatalf("enterShoot 應進入射擊模式")
	}
	s.updateShoot(input.Action{Kind: input.ActConfirm})

	if s.shooting {
		t.Errorf("射擊後應離開 target mode")
	}
	if got := s.combat.Units[shooterSide][shooterID].Shots; got != shotsBefore-1 {
		t.Errorf("射擊後彈藥應 -1:got %d, want %d", got, shotsBefore-1)
	}
	if !s.combat.Units[shooterSide][shooterID].Acted &&
		(s.combat.Side == shooterSide && s.combat.UnitID == shooterID) {
		t.Errorf("射擊後該單位應行動結束(Acted 或已換手)")
	}
}

// 射擊模式按 Cancel:退出回一般指令,不消耗彈藥。
func TestCombatShoot_CancelExits(t *testing.T) {
	s := newArcherCombatScreen(t)
	shotsBefore := s.combat.Units[s.combat.Side][s.combat.UnitID].Shots
	s.enterShoot()
	s.updateShoot(input.Action{Kind: input.ActCancel})
	if s.shooting {
		t.Errorf("Cancel 應退出射擊模式")
	}
	if got := s.combat.Units[s.combat.Side][s.combat.UnitID].Shots; got != shotsBefore {
		t.Errorf("取消不應消耗彈藥:got %d, want %d", got, shotsBefore)
	}
}
