package combat

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// archerTroopID / wolfTroopID 對齊 kbdata.tables_troops.go 的固定索引:
//
//	8 = 弓箭手(RangedShots:12,有遠攻);3 = 野狼(無遠攻,純近戰)。
const (
	archerTroopID = 8
	wolfTroopID   = 3
)

// newShootTestCombat 佈一場最小戰鬥:side0 一個弓箭手、side1 一個野狼,
// 位置由呼叫端指定;回傳已 ResetMatch(彈藥/障礙就緒,seed 固定)的 Combat。
func newShootTestCombat(t *testing.T, a *kbdata.Assets, ax, ay, wx, wy int) *Combat {
	t.Helper()
	c := &Combat{}
	c.Units[0][0] = Unit{TroopID: archerTroopID, Count: 5, MaxCount: 5, X: ax, Y: ay}
	c.Units[1][0] = Unit{TroopID: wolfTroopID, Count: 5, MaxCount: 5, X: wx, Y: wy}
	c.Heroes[0] = nil // 無 Hero:略過士氣/領導力,傷害純骰,測試可判定
	c.ResetMatch(a, kbrng.NewGlibc(1), false)
	// ResetMatch 會依隨機障礙圖擺放,測試只關心射擊本身,清掉障礙避免干擾。
	for y := 0; y < BoardH; y++ {
		for x := 0; x < BoardW; x++ {
			c.Omap[y][x] = 0
		}
	}
	return c
}

// TestCanShoot_RangedUnitHitsDistantEnemy:遠攻單位對「不相鄰」的敵人可射,
// 射擊後敵人扣血、且自身彈藥 -1(對齊 unit_ranged_shot / unit_try_shoot)。
func TestCanShoot_RangedUnitHitsDistantEnemy(t *testing.T) {
	a := loadTestAssets(t)
	c := newShootTestCombat(t, a, 0, 0, 5, 4) // 弓箭手 (0,0),野狼 (5,4):距離遠、不相鄰

	if !c.CanShoot(0, 0) {
		t.Fatalf("有彈藥且未被貼身的弓箭手應可射擊,CanShoot=false")
	}
	shotsBefore := c.Units[0][0].Shots
	countBefore := c.Units[1][0].Count
	if shotsBefore <= 0 {
		t.Fatalf("弓箭手起始彈藥應 >0,got %d", shotsBefore)
	}

	kills := c.UnitRangedShot(a, kbrng.NewGlibc(7), 0, 0, 1, 0)

	if got := c.Units[0][0].Shots; got != shotsBefore-1 {
		t.Errorf("射擊後彈藥應 -1:got %d, want %d", got, shotsBefore-1)
	}
	// 傷害「有效」= 整隻擊殺(Count 下降)或累積餘傷(Injury>0)。單發低骰可能
	// 不足一隻 HP(野狼 HP=3),此時 Count 不變但 Injury 記帳,兩者任一成立即算命中。
	dealt := kills > 0 || c.Units[1][0].Count < countBefore || c.Units[1][0].Injury > 0
	if !dealt {
		t.Errorf("射擊應造成傷害:kills=%d, count %d→%d, injury=%d",
			kills, countBefore, c.Units[1][0].Count, c.Units[1][0].Injury)
	}
}

// TestCanShoot_NoAmmo:彈藥耗盡(Shots==0)的單位不能射(對齊 unit_try_shoot 的 !u->shots)。
func TestCanShoot_NoAmmo(t *testing.T) {
	a := loadTestAssets(t)
	c := newShootTestCombat(t, a, 0, 0, 5, 4)
	c.Units[0][0].Shots = 0

	if c.CanShoot(0, 0) {
		t.Errorf("無彈藥單位不應可射,CanShoot=true")
	}
}

// TestCanShoot_Surrounded:被敵人貼身(相鄰)時不能射,只能近戰
// (對齊 unit_try_shoot 的 unit_surrounded 分支 / grid_heuristic 近敵走 MELEE)。
func TestCanShoot_Surrounded(t *testing.T) {
	a := loadTestAssets(t)
	c := newShootTestCombat(t, a, 2, 2, 3, 2) // 野狼 (3,2) 就在弓箭手 (2,2) 右邊 → 相鄰

	if !c.UnitSurrounded(0, 0) {
		t.Fatalf("相鄰敵人應使 UnitSurrounded=true")
	}
	if c.CanShoot(0, 0) {
		t.Errorf("被貼身的弓箭手不應可射(只能近戰),CanShoot=true")
	}
}

// TestCanShoot_NonRangedMeleeUnit:純近戰兵種(野狼,RangedShots=0)永遠不能射。
func TestCanShoot_NonRangedMeleeUnit(t *testing.T) {
	a := loadTestAssets(t)
	c := newShootTestCombat(t, a, 0, 0, 5, 4)

	if c.CanShoot(1, 0) {
		t.Errorf("無遠攻兵種(野狼)不應可射,CanShoot=true")
	}
}
