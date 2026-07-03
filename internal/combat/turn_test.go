package combat

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// findTroopWithAbility 在 a.Troops 裡找第一個帶有指定能力位元的兵種索引,
// 供測試用(不寫死具體 troop id,兵種表若微調也不會壞測試)。
func findTroopWithAbility(t *testing.T, a *kbdata.Assets, ability kbdata.Ability) int {
	t.Helper()
	for i, tr := range a.Troops {
		if tr.Abilities&ability != 0 {
			return i
		}
	}
	t.Fatalf("找不到帶有能力位元 %v 的兵種", ability)
	return -1
}

func TestResetTurn_MovesFlightsFromTroop(t *testing.T) {
	a := loadTestAssets(t)
	flyerID := findTroopWithAbility(t, a, kbdata.AbilFly)

	var c Combat
	c.Units[0][0] = Unit{TroopID: flyerID, Count: 5}
	c.Units[0][1] = Unit{TroopID: 0, Count: 3} // 農夫,無飛行

	c.ResetTurn(a)

	if got, want := c.Units[0][0].Moves, a.Troops[flyerID].Move; got != want {
		t.Errorf("flyer Moves = %d, want %d (troop.Move)", got, want)
	}
	if got := c.Units[0][0].Flights; got != 2 {
		t.Errorf("flyer Flights = %d, want 2(ABIL_FLY)", got)
	}
	if got, want := c.Units[0][1].Moves, a.Troops[0].Move; got != want {
		t.Errorf("農夫 Moves = %d, want %d", got, want)
	}
	if got := c.Units[0][1].Flights; got != 0 {
		t.Errorf("農夫 Flights = %d, want 0(無 ABIL_FLY)", got)
	}
}

func TestResetTurn_ResetsActedRetaliatedTurnCount(t *testing.T) {
	a := loadTestAssets(t)

	var c Combat
	c.Units[0][0] = Unit{TroopID: 0, Count: 4, Acted: true, Retaliated: true}

	c.ResetTurn(a)

	u := c.Units[0][0]
	if u.Acted {
		t.Error("Acted 應被重設為 false")
	}
	if u.Retaliated {
		t.Error("Retaliated 應被重設為 false")
	}
	if u.TurnCount != u.Count {
		t.Errorf("TurnCount = %d, want %d(= Count)", u.TurnCount, u.Count)
	}
	if c.Phase != 0 {
		t.Errorf("Phase = %d, want 0", c.Phase)
	}
	if c.Spells != 0 {
		t.Errorf("Spells = %d, want 0", c.Spells)
	}
}

// TestResetTurn_UnfreezeRegenGatedBySideOne 對應 C 版 reset_turn 的
// `if (war->side == 1)` 分支:只有「即將從側 1 換手」那一刻才解凍 + 套用再生。
func TestResetTurn_UnfreezeRegenGatedBySideOne(t *testing.T) {
	a := loadTestAssets(t)
	regenID := findTroopWithAbility(t, a, kbdata.AbilRegen)

	// c.Side == 0:不該解凍、不該套用再生。
	var c0 Combat
	c0.Side = 0
	c0.Units[0][0] = Unit{TroopID: regenID, Count: 3, Frozen: true, Injury: 5}
	c0.ResetTurn(a)
	if !c0.Units[0][0].Frozen {
		t.Error("side==0 時不該解凍")
	}
	if c0.Units[0][0].Injury != 5 {
		t.Error("side==0 時不該套用再生(Injury 不該被清零)")
	}

	// c.Side == 1:應解凍 + 套用再生。
	var c1 Combat
	c1.Side = 1
	c1.Units[0][0] = Unit{TroopID: regenID, Count: 3, Frozen: true, Injury: 5}
	c1.ResetTurn(a)
	if c1.Units[0][0].Frozen {
		t.Error("side==1 時應解凍")
	}
	if c1.Units[0][0].Injury != 0 {
		t.Errorf("side==1 時應套用再生,Injury = %d, want 0", c1.Units[0][0].Injury)
	}
}

func TestNextUnit_SkipsDeadAndActedThenAdvancesPhase(t *testing.T) {
	a := loadTestAssets(t)

	var c Combat
	c.Side = 0
	c.UnitID = 0
	c.Heroes[0] = nil // 無 hero → unit_under_control 恆真(受控)
	c.Units[0][0] = Unit{TroopID: 0, Count: 3}
	c.Units[0][1] = Unit{TroopID: 0, Count: 0} // 死格,應跳過
	c.Units[0][2] = Unit{TroopID: 0, Count: 3, Acted: true}
	c.Units[0][3] = Unit{TroopID: 0, Count: 3}
	c.Units[0][4] = Unit{TroopID: 0, Count: 0}

	next := c.NextUnit(a)
	if next != 3 {
		t.Fatalf("NextUnit = %d, want 3(跳過死格 1 與已行動的 2)", next)
	}
	if c.Units[0][3].OutOfControl {
		t.Error("無 hero 側,OutOfControl 應為 false")
	}

	// 把 3 也標記已行動,全員耗盡 → 應該進下一個 phase,並從頭找到 unit 0
	// (unit_id 目前是呼叫端維護,NextUnit 只回傳索引,不會自己前移 c.UnitID;
	// 這裡手動模擬呼叫端把 UnitID 設成 3,再呼叫一次)。
	c.UnitID = 3
	c.Units[0][3].Acted = true
	c.Units[0][0].Acted = true // 假設 0 也已行動,只剩不存在的候選

	next = c.NextUnit(a)
	if next != -1 {
		t.Fatalf("全部已行動/死亡時 NextUnit = %d, want -1", next)
	}
	if c.Phase != 1 {
		t.Errorf("Phase = %d, want 1(NextUnit 兩輪都找不到時應 ++)", c.Phase)
	}
}

func TestNextTurn_SwitchesSideAndResets(t *testing.T) {
	a := loadTestAssets(t)

	var c Combat
	c.Side = 0
	c.Turn = 5
	c.Units[0][0] = Unit{TroopID: 0, Count: 0}
	c.Units[0][1] = Unit{TroopID: 0, Count: 3, Acted: true}
	c.Units[1][0] = Unit{TroopID: 0, Count: 4}

	ok := c.NextTurn(a)
	if !ok {
		t.Fatal("NextTurn 應找到側 0 裡有兵的格子(index 1),回傳 true")
	}
	if c.Turn != 6 {
		t.Errorf("Turn = %d, want 6", c.Turn)
	}
	if c.Side != 1 {
		t.Errorf("Side = %d, want 1(已從 0 換手)", c.Side)
	}
	if c.UnitID != 1 {
		t.Errorf("UnitID = %d, want 1(舊側第一個有兵格的索引,見 NextTurn 文件的「索引套用到新側」注解)", c.UnitID)
	}
	// ResetTurn 應該已經把 Acted 清掉(NextTurn 內部呼叫 ResetTurn 在換手之前)。
	if c.Units[0][1].Acted {
		t.Error("NextTurn 應該呼叫 ResetTurn 清掉 Acted")
	}
}

func TestNextTurn_NoLivingUnitsReturnsFalse(t *testing.T) {
	a := loadTestAssets(t)
	var c Combat
	c.Side = 0
	// 側 0 全滅
	ok := c.NextTurn(a)
	if ok {
		t.Error("側 0 全滅時 NextTurn 應回傳 false")
	}
	if c.Side != 0 {
		t.Error("找不到單位時不該換側")
	}
}

func TestResetMatch_SetsShotsInjuryUmapAndObstacleColumns(t *testing.T) {
	a := loadTestAssets(t)
	archerID := -1
	for i, tr := range a.Troops {
		if tr.RangedShots > 0 {
			archerID = i
			break
		}
	}
	if archerID == -1 {
		t.Fatal("找不到有 RangedShots 的兵種")
	}

	var c Combat
	c.Units[0][0] = Unit{TroopID: archerID, Count: 5, Injury: 9, Frozen: true, X: 0, Y: 0}
	c.Units[1][0] = Unit{TroopID: 0, Count: 3, X: BoardW - 1, Y: 0}

	rng := kbrng.NewGlibc(1)
	c.ResetMatch(a, rng, false)

	if got, want := c.Units[0][0].Shots, a.Troops[archerID].RangedShots; got != want {
		t.Errorf("Shots = %d, want %d(troop.RangedShots)", got, want)
	}
	if c.Units[0][0].Injury != 0 {
		t.Error("ResetMatch 應清零 Injury")
	}
	if c.Units[0][0].Frozen {
		t.Error("ResetMatch 應清除 Frozen")
	}
	if c.Umap[0][0] != PackUID(0, 0) {
		t.Errorf("Umap[0][0] = %d, want %d", c.Umap[0][0], PackUID(0, 0))
	}
	if c.Umap[0][BoardW-1] != PackUID(1, 0) {
		t.Errorf("Umap[0][%d] = %d, want %d", BoardW-1, c.Umap[0][BoardW-1], PackUID(1, 0))
	}

	// 障礙只會出現在欄 1..BoardW-3(對應 C 的 `for (i=1;i<CLEVEL_W-2;i++)`),
	// 其餘欄(0 與最後兩欄)恆為 0。
	for y := 0; y < BoardH; y++ {
		if c.Omap[y][0] != 0 {
			t.Errorf("Omap[%d][0] = %d, want 0(欄 0 不產生障礙)", y, c.Omap[y][0])
		}
		for x := BoardW - 2; x < BoardW; x++ {
			if c.Omap[y][x] != 0 {
				t.Errorf("Omap[%d][%d] = %d, want 0(尾兩欄不產生障礙)", y, x, c.Omap[y][x])
			}
		}
	}
}
