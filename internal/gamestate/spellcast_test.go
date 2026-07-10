package gamestate

import "testing"

// TestCastAdventureSpell_NotLearned 驗證未學會的法術回 AdvNotLearned,且不消耗(本來就是 0)。
func TestCastAdventureSpell_NotLearned(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.Spells[8] = 0 // 確保未學會

	r := gs.CastAdventureSpell(a, 8)
	if r.Kind != AdvNotLearned {
		t.Fatalf("CastAdventureSpell(未學會).Kind = %v, want AdvNotLearned", r.Kind)
	}
	if gs.Spells[8] != 0 {
		t.Errorf("未學會不應消耗法術,Spells[8] = %d, want 0", gs.Spells[8])
	}
}

// TestCastAdventureSpell_Timestop 驗證停時術:TimeStop 增加、消耗一個法術、Amount 正確。
func TestCastAdventureSpell_Timestop(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.Spells[8] = 1
	gs.SpellPower = 2
	timeStopBefore := gs.TimeStop

	r := gs.CastAdventureSpell(a, 8)

	if r.Kind != AdvTimestop {
		t.Fatalf("CastAdventureSpell(timestop).Kind = %v, want AdvTimestop", r.Kind)
	}
	if r.Amount != gs.SpellPower*10 {
		t.Errorf("r.Amount = %d, want %d", r.Amount, gs.SpellPower*10)
	}
	if gs.TimeStop != timeStopBefore+gs.SpellPower*10 {
		t.Errorf("gs.TimeStop = %d, want %d", gs.TimeStop, timeStopBefore+gs.SpellPower*10)
	}
	if gs.Spells[8] != 0 {
		t.Errorf("成功施放應消耗一個法術,Spells[8] = %d, want 0", gs.Spells[8])
	}
}

// TestCastAdventureSpell_RaiseControl 驗證提升控制:Leadership 增加、消耗一個法術、Amount 正確。
func TestCastAdventureSpell_RaiseControl(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.Spells[13] = 1
	gs.SpellPower = 2
	leadershipBefore := gs.Leadership

	r := gs.CastAdventureSpell(a, 13)

	if r.Kind != AdvRaiseControl {
		t.Fatalf("CastAdventureSpell(raisecontrol).Kind = %v, want AdvRaiseControl", r.Kind)
	}
	if r.Amount != gs.SpellPower*100 {
		t.Errorf("r.Amount = %d, want %d", r.Amount, gs.SpellPower*100)
	}
	if gs.Leadership != leadershipBefore+gs.SpellPower*100 {
		t.Errorf("gs.Leadership = %d, want %d", gs.Leadership, leadershipBefore+gs.SpellPower*100)
	}
	if gs.Spells[13] != 0 {
		t.Errorf("成功施放應消耗一個法術,Spells[13] = %d, want 0", gs.Spells[13])
	}
}

// TestCastAdventureSpell_InstantArmy 驗證即時軍隊:清空隊伍後施放應加入該職業的
// InstantArmy 兵種,數量 = (SpellPower+1)*multiplier[rank](rank 0 = 3)。
func TestCastAdventureSpell_InstantArmy(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.Spells[12] = 1
	gs.SpellPower = 2
	gs.Army = [5]Squad{{TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}}

	wantTroopID := a.Classes[gs.Class][gs.Rank].InstantArmy
	wantAmount := (gs.SpellPower + 1) * 3 // rank 0 的 multiplier = 3

	r := gs.CastAdventureSpell(a, 12)

	if r.Kind != AdvInstantArmy {
		t.Fatalf("CastAdventureSpell(instantarmy).Kind = %v, want AdvInstantArmy", r.Kind)
	}
	if r.Amount != wantAmount {
		t.Errorf("r.Amount = %d, want %d", r.Amount, wantAmount)
	}
	if r.TroopID != wantTroopID {
		t.Errorf("r.TroopID = %d, want %d", r.TroopID, wantTroopID)
	}
	if gs.Army[0].TroopID != wantTroopID || gs.Army[0].Count != wantAmount {
		t.Errorf("Army[0] = %+v, want {TroopID:%d Count:%d}", gs.Army[0], wantTroopID, wantAmount)
	}
	if gs.Spells[12] != 0 {
		t.Errorf("成功施放應消耗一個法術,Spells[12] = %d, want 0", gs.Spells[12])
	}
}

// TestCastAdventureSpell_Todo 驗證尚未實作的冒險法術(如 Bridge)回 AdvTodo,不消耗法術。
func TestCastAdventureSpell_Todo(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.Spells[7] = 1 // Bridge

	r := gs.CastAdventureSpell(a, 7)

	if r.Kind != AdvTodo {
		t.Fatalf("CastAdventureSpell(bridge).Kind = %v, want AdvTodo", r.Kind)
	}
	if gs.Spells[7] != 1 {
		t.Errorf("AdvTodo 不應消耗法術,Spells[7] = %d, want 1", gs.Spells[7])
	}
}
