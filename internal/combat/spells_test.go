package combat

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// 測試用兵種 id(對照 internal/kbdata/tables_troops.go 的固定表順序):
//   - troopNormal = 6  地精,HP=5,Abilities=0(無不死、無免疫)
//   - troopUndead = 4  骷髏兵,HP=3,Abilities=AbilUndead
//   - troopImmune = 24 火龍,HP=200,Abilities=AbilFly|AbilImmune
const (
	troopNormal = 6
	troopUndead = 4
	troopImmune = 24
)

// TestCloneTroop_CleanDivision 驗證整除情況:injury 從 0 開始,
// spellPower*10 剛好是 hp 的倍數時,clones/Count/MaxCount 依公式增加,Injury 歸 0。
func TestCloneTroop_CleanDivision(t *testing.T) {
	a := loadTestAssets(t)
	hp := a.Troops[troopNormal].HP // 5

	var c Combat
	c.Units[0][0] = Unit{TroopID: troopNormal, Count: 10, MaxCount: 10, Injury: 0}

	spellPower := 5 // damage = 5*10+0 = 50, 50/5 = 10 整除
	wantClones := (spellPower*10 + 0) / hp

	got := c.CloneTroop(a, spellPower, 0)

	if got != wantClones {
		t.Errorf("CloneTroop 回傳 = %d, want %d", got, wantClones)
	}
	if c.Units[0][0].Count != 10+wantClones {
		t.Errorf("Count = %d, want %d", c.Units[0][0].Count, 10+wantClones)
	}
	if c.Units[0][0].MaxCount != 10+wantClones {
		t.Errorf("MaxCount = %d, want %d", c.Units[0][0].MaxCount, 10+wantClones)
	}
	if c.Units[0][0].Injury != 0 {
		t.Errorf("Injury = %d, want 0", c.Units[0][0].Injury)
	}
}

// TestCloneTroop_WithRemainder 驗證有餘數情況:起始 Injury 非 0,
// (spellPower*10+Injury) 除以 hp 有餘數時,餘數應寫回 Injury。
func TestCloneTroop_WithRemainder(t *testing.T) {
	a := loadTestAssets(t)
	hp := a.Troops[troopNormal].HP // 5

	var c Combat
	c.Units[0][0] = Unit{TroopID: troopNormal, Count: 10, MaxCount: 10, Injury: 3}

	spellPower := 1 // damage = 1*10+3 = 13, 13/5 = 2 餘 3
	damage := spellPower*10 + 3
	wantClones := damage / hp
	wantInjury := damage % hp

	got := c.CloneTroop(a, spellPower, 0)

	if got != wantClones {
		t.Errorf("CloneTroop 回傳 = %d, want %d", got, wantClones)
	}
	if c.Units[0][0].Count != 10+wantClones {
		t.Errorf("Count = %d, want %d", c.Units[0][0].Count, 10+wantClones)
	}
	if c.Units[0][0].MaxCount != 10+wantClones {
		t.Errorf("MaxCount = %d, want %d", c.Units[0][0].MaxCount, 10+wantClones)
	}
	if c.Units[0][0].Injury != wantInjury {
		t.Errorf("Injury = %d, want %d", c.Units[0][0].Injury, wantInjury)
	}
}

// TestFreezeTroop_Normal 驗證一般兵種可被凍結:回傳 1,Frozen 設為 true。
func TestFreezeTroop_Normal(t *testing.T) {
	a := loadTestAssets(t)

	var c Combat
	c.Units[0][0] = Unit{TroopID: troopNormal, Count: 5, MaxCount: 5}

	got := c.FreezeTroop(a, 0, 0)

	if got != 1 {
		t.Errorf("FreezeTroop 回傳 = %d, want 1", got)
	}
	if !c.Units[0][0].Frozen {
		t.Errorf("Frozen 應為 true")
	}
}

// TestFreezeTroop_Immune 驗證魔法免疫兵種不受凍結:回傳 -1,Frozen 維持 false。
func TestFreezeTroop_Immune(t *testing.T) {
	a := loadTestAssets(t)

	var c Combat
	c.Units[0][0] = Unit{TroopID: troopImmune, Count: 1, MaxCount: 1}

	got := c.FreezeTroop(a, 0, 0)

	if got != -1 {
		t.Errorf("FreezeTroop 回傳 = %d, want -1", got)
	}
	if c.Units[0][0].Frozen {
		t.Errorf("免疫兵種不該被 Frozen")
	}
}

// TestResurrectTroop_PartialRevive 驗證未達戰死上限時,復活數 = spellPower。
func TestResurrectTroop_PartialRevive(t *testing.T) {
	var c Combat
	c.Units[0][0] = Unit{TroopID: troopNormal, MaxCount: 20, Count: 5, Dead: true}

	got := c.ResurrectTroop(3, 0, 0)

	if got != 3 {
		t.Errorf("ResurrectTroop 回傳 = %d, want 3", got)
	}
	if c.Units[0][0].Count != 8 {
		t.Errorf("Count = %d, want 8", c.Units[0][0].Count)
	}
	if c.Units[0][0].TurnCount != c.Units[0][0].Count {
		t.Errorf("TurnCount = %d, want == Count(%d)", c.Units[0][0].TurnCount, c.Units[0][0].Count)
	}
	if c.Units[0][0].Dead {
		t.Errorf("Dead 應為 false")
	}
}

// TestResurrectTroop_CappedByDead 驗證復活數不能超過戰死數(MaxCount-Count)。
func TestResurrectTroop_CappedByDead(t *testing.T) {
	var c Combat
	c.Units[0][0] = Unit{TroopID: troopNormal, MaxCount: 20, Count: 19, Dead: true}

	got := c.ResurrectTroop(5, 0, 0)

	if got != 1 {
		t.Errorf("ResurrectTroop 回傳 = %d, want 1(上限為戰死數)", got)
	}
	if c.Units[0][0].Count != 20 {
		t.Errorf("Count = %d, want 20", c.Units[0][0].Count)
	}
	if c.Units[0][0].Dead {
		t.Errorf("Dead 應為 false")
	}
}

// TestMagicDamage_ImmuneTarget 驗證魔法免疫目標一律無效,回傳 -1(即使 undead=false)。
func TestMagicDamage_ImmuneTarget(t *testing.T) {
	a := loadTestAssets(t)
	rng := kbrng.NewGlibc(1)

	var c Combat
	c.Units[1][0] = Unit{TroopID: troopImmune, Count: 5, MaxCount: 5, TurnCount: 5}

	got := c.MagicDamage(a, rng, 3, 5, 1, 0, false)

	if got != -1 {
		t.Errorf("MagicDamage(免疫目標) = %d, want -1", got)
	}
}

// TestMagicDamage_TurnUndeadOnNormal 驗證驅散不死(undead=true)對非不死目標無效,回傳 -1。
func TestMagicDamage_TurnUndeadOnNormal(t *testing.T) {
	a := loadTestAssets(t)
	rng := kbrng.NewGlibc(1)

	var c Combat
	c.Units[1][0] = Unit{TroopID: troopNormal, Count: 5, MaxCount: 5, TurnCount: 5}

	got := c.MagicDamage(a, rng, 3, 5, 1, 0, true)

	if got != -1 {
		t.Errorf("MagicDamage(驅散不死打一般兵種) = %d, want -1", got)
	}
}

// TestMagicDamage_TurnUndeadOnUndead 驗證驅散不死對不死族有效,應呼叫 DealDamage(不回 -1)。
func TestMagicDamage_TurnUndeadOnUndead(t *testing.T) {
	a := loadTestAssets(t)
	rng := kbrng.NewGlibc(1)

	var c Combat
	c.Units[1][0] = Unit{TroopID: troopUndead, Count: 5, MaxCount: 5, TurnCount: 5}

	got := c.MagicDamage(a, rng, 3, 5, 1, 0, true)

	if got == -1 {
		t.Errorf("MagicDamage(驅散不死打不死族) 不應回 -1")
	}
}

// TestMagicDamage_NormalDamage 驗證一般魔法傷害(undead=false)對非免疫目標有效。
func TestMagicDamage_NormalDamage(t *testing.T) {
	a := loadTestAssets(t)
	rng := kbrng.NewGlibc(1)

	var c Combat
	c.Units[1][0] = Unit{TroopID: troopNormal, Count: 5, MaxCount: 5, TurnCount: 5}

	got := c.MagicDamage(a, rng, 3, 5, 1, 0, false)

	if got == -1 {
		t.Errorf("MagicDamage(一般傷害打一般兵種) 不應回 -1")
	}
}

// TestRelocateUnit 驗證搬移單位會更新座標與 Umap 佔格(舊格清空、新格標記 PackUID)。
func TestRelocateUnit(t *testing.T) {
	var c Combat
	c.Units[0][0] = Unit{TroopID: troopNormal, X: 0, Y: 0}
	c.Umap[0][0] = PackUID(0, 0)

	c.RelocateUnit(0, 0, 3, 4)

	if c.Units[0][0].X != 3 || c.Units[0][0].Y != 4 {
		t.Errorf("(X,Y) = (%d,%d), want (3,4)", c.Units[0][0].X, c.Units[0][0].Y)
	}
	if c.Umap[0][0] != 0 {
		t.Errorf("舊格 Umap[0][0] = %d, want 0", c.Umap[0][0])
	}
	if c.Umap[4][3] != PackUID(0, 0) {
		t.Errorf("新格 Umap[4][3] = %d, want %d", c.Umap[4][3], PackUID(0, 0))
	}
}
