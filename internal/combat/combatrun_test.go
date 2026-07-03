package combat

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// clearObstacles 清掉隨機障礙,避免罕見佈局擋死路徑干擾收斂驗證。
func clearObstacles(c *Combat) {
	for y := 0; y < BoardH; y++ {
		for x := 0; x < BoardW; x++ {
			c.Omap[y][x] = 0
		}
	}
}

// RunCombat 被動玩家 vs AI:玩家不主動出手(僅自動反擊),戰鬥仍應決定性收斂。
// 註:確切勝方取決於反擊數學與回合順序,需 combat_loop 的 C oracle 才能對 parity;
// 此處只驗「會終止分勝負、不卡死」(RunCombat 契約的核心保證)。
func TestRunCombat_PassivePlayerConverges(t *testing.T) {
	a := loadTestAssets(t)
	rng := kbrng.NewGlibc(99)

	var c Combat
	c.Units[0][0] = Unit{TroopID: 0, Count: 5, X: 0, Y: 2}          // 農夫 x5(被動,僅反擊)
	c.Units[1][0] = Unit{TroopID: 3, Count: 4, X: BoardW - 1, Y: 2} // 野狼 x4(AI)
	c.ResetMatch(a, rng, false)
	clearObstacles(&c)

	result := c.RunCombat(a, rng, false, 5000)
	if result == ResultUndecided {
		t.Fatal("被動玩家戰鬥應收斂分勝負,卻 Undecided(可能卡死不收斂)")
	}
}

// auto-battle:兩側皆 AI,應在上限內分出決定性勝負(非 Undecided)。
func TestRunCombat_AutoBattleDecisive(t *testing.T) {
	a := loadTestAssets(t)
	rng := kbrng.NewGlibc(7)

	var c Combat
	c.Units[0][0] = Unit{TroopID: 2, Count: 8, X: 0, Y: 1}          // 義勇軍 x8
	c.Units[0][1] = Unit{TroopID: 8, Count: 3, X: 0, Y: 3}          // 弓箭手 x3
	c.Units[1][0] = Unit{TroopID: 3, Count: 6, X: BoardW - 1, Y: 1} // 野狼 x6
	c.Units[1][1] = Unit{TroopID: 4, Count: 4, X: BoardW - 1, Y: 3} // 骷髏兵 x4
	c.ResetMatch(a, rng, false)
	clearObstacles(&c)

	result := c.RunCombat(a, rng, true, 5000)
	if result == ResultUndecided {
		t.Fatal("auto-battle 應在上限內分出勝負,卻 Undecided(可能未收斂)")
	}
}

// FlyUnit:relocate + 扣 1 flight,歸 0 標記 acted。
func TestFlyUnit(t *testing.T) {
	var c Combat
	c.Units[0][0] = Unit{TroopID: 1, Count: 3, X: 0, Y: 0, Flights: 2}
	c.Umap[0][0] = PackUID(0, 0)

	c.FlyUnit(0, 0, 3, 2)
	u := &c.Units[0][0]
	if u.X != 3 || u.Y != 2 {
		t.Fatalf("飛行後座標 (%d,%d), want (3,2)", u.X, u.Y)
	}
	if c.Umap[0][0] != 0 || c.Umap[2][3] != PackUID(0, 0) {
		t.Fatal("Umap 未正確更新(舊格未清/新格未登記)")
	}
	if u.Flights != 1 {
		t.Fatalf("flights 應扣為 1, got %d", u.Flights)
	}
	if u.Acted {
		t.Fatal("還有 flights 不該標 acted")
	}
	c.FlyUnit(0, 0, 4, 4) // 第二次 → flights 歸 0 → acted
	if !u.Acted {
		t.Fatal("flights 用盡應標 acted")
	}
}

// 會飛單位 AIUnitThink:無近戰目標時飛到遠處目標相鄰格(越過距離)。
func TestAIFlyReachesAdjacent(t *testing.T) {
	a := loadTestAssets(t)
	rng := kbrng.NewGlibc(1)

	var c Combat
	// 小妖精(id 1,ABIL_FLY,無遠攻)在角落;敵方農夫在對角遠處
	c.Units[0][0] = Unit{TroopID: 1, Count: 3, X: 0, Y: 0}
	c.Units[1][0] = Unit{TroopID: 0, Count: 3, X: BoardW - 1, Y: BoardH - 1}
	c.ResetMatch(a, rng, false)
	clearObstacles(&c)
	c.Side = 0
	c.UnitID = 0

	flyer := &c.Units[0][0]
	tgt := &c.Units[1][0]
	if flyer.Flights <= 0 {
		t.Fatalf("小妖精應有飛行點, got %d", flyer.Flights)
	}
	c.AIUnitThink(a, rng)
	if !Adjacent(flyer.X, flyer.Y, tgt.X, tgt.Y) {
		t.Fatalf("飛行後應貼近目標,got flyer(%d,%d) tgt(%d,%d)", flyer.X, flyer.Y, tgt.X, tgt.Y)
	}
}
