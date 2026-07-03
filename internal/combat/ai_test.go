package combat

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// 以下測試用到的固定 troop id,對照 internal/kbdata/tables_troops.go:
//   0 農夫    HP1  Move1 melee1-1 無遠攻
//   3 野狼    HP3  Move3 melee1-3 無遠攻 SkillLevel2
//   8 弓箭手  HP10 Move2 melee1-2 ranged1-3 shots12

func TestAIPickTarget_NearbyIgnoresFarAndPicksSmallestHP(t *testing.T) {
	a := loadTestAssets(t)

	var c Combat
	c.Side = 0
	c.UnitID = 0
	c.Units[0][0] = Unit{TroopID: 0, Count: 5, X: 2, Y: 2} // 行動單位:農夫

	c.Units[1][0] = Unit{TroopID: 3, Count: 2, X: 3, Y: 2} // 相鄰,HP3(野狼)
	c.Units[1][1] = Unit{TroopID: 0, Count: 2, X: 2, Y: 3} // 相鄰,HP1(農夫)→ 應中選
	c.Units[1][2] = Unit{TroopID: 0, Count: 1, X: 4, Y: 4} // HP 更小但不相鄰 → nearby 模式應排除

	side, id, ok := c.AIPickTarget(a, true)
	if !ok {
		t.Fatal("應找到近戰目標")
	}
	if side != 1 || id != 1 {
		t.Errorf("AIPickTarget(nearby=true) = (%d,%d), want (1,1)(相鄰候選中 HP 最小者)", side, id)
	}
}

func TestAIPickTarget_FarPrefersUnitWithShotsRegardlessOfHP(t *testing.T) {
	a := loadTestAssets(t)

	var c Combat
	c.Side = 0
	c.UnitID = 0
	c.Units[0][0] = Unit{TroopID: 0, Count: 5, X: 0, Y: 0}

	c.Units[1][0] = Unit{TroopID: 8, Count: 3, Shots: 5, X: 5, Y: 4} // 弓箭手,HP10,還有 shots
	c.Units[1][1] = Unit{TroopID: 0, Count: 3, Shots: 0, X: 5, Y: 0} // 農夫,HP1,無 shots

	side, id, ok := c.AIPickTarget(a, false)
	if !ok {
		t.Fatal("應找到遠程目標")
	}
	if side != 1 || id != 0 {
		t.Errorf("AIPickTarget(nearby=false) = (%d,%d), want (1,0)(有 shots 者優先於 HP 更小者)", side, id)
	}
}

func TestAIPickTarget_NoTargetsReturnsFalse(t *testing.T) {
	a := loadTestAssets(t)

	var c Combat
	c.Side = 0
	c.UnitID = 0
	c.Units[0][0] = Unit{TroopID: 0, Count: 5}
	// 側 1 全空,理應找不到目標。

	if _, _, ok := c.AIPickTarget(a, true); ok {
		t.Error("沒有任何目標時 AIPickTarget 應回傳 ok=false")
	}
	if _, _, ok := c.AIPickTarget(a, false); ok {
		t.Error("沒有任何目標時 AIPickTarget(far) 應回傳 ok=false")
	}
}

func TestAIPickTarget_OutOfControlCanTargetOwnSide(t *testing.T) {
	a := loadTestAssets(t)

	var c Combat
	c.Side = 0
	c.UnitID = 0
	c.Units[0][0] = Unit{TroopID: 0, Count: 5, OutOfControl: true, X: 1, Y: 1}
	c.Units[0][1] = Unit{TroopID: 0, Count: 3, X: 2, Y: 1} // 己方,相鄰
	// 側 1 沒有任何單位。

	side, id, ok := c.AIPickTarget(a, true)
	if !ok {
		t.Fatal("失控單位應能把己方當目標")
	}
	if side != 0 || id != 1 {
		t.Errorf("AIPickTarget = (%d,%d), want (0,1)", side, id)
	}
}

func TestAIUnitThink_MovesOneStepTowardFarTarget(t *testing.T) {
	a := loadTestAssets(t)
	rng := kbrng.NewGlibc(42)

	var c Combat
	c.Side = 1
	c.UnitID = 0
	c.Units[1][0] = Unit{TroopID: 0, Count: 5, Moves: 3, X: 0, Y: 0}
	c.Units[0][0] = Unit{TroopID: 0, Count: 5, X: 5, Y: 4} // 遠端玩家單位
	c.Umap[0][0] = PackUID(1, 0)
	c.Umap[4][5] = PackUID(0, 0)

	beforeDist := Distance(0, 0, 5, 4)

	pass := c.AIUnitThink(a, rng)
	if pass {
		t.Fatal("有合法移動路徑時不該 pass=true")
	}

	u := c.Units[1][0]
	if u.X == 0 && u.Y == 0 {
		t.Fatal("單位應該往目標移動一步,座標不該不變")
	}
	if got := Distance(u.X, u.Y, 5, 4); got >= beforeDist {
		t.Errorf("移動後到目標距離 = %d, 應小於移動前的 %d", got, beforeDist)
	}
	if u.Moves != 2 {
		t.Errorf("Moves = %d, want 2(消耗 1 步)", u.Moves)
	}
	if c.Umap[u.Y][u.X] != PackUID(1, 0) {
		t.Error("Umap 應更新到新座標")
	}
	if c.Umap[0][0] != 0 {
		t.Error("舊格 Umap 應清空")
	}
}

func TestAIUnitThink_MeleeAttacksWhenAdjacent(t *testing.T) {
	a := loadTestAssets(t)
	rng := kbrng.NewGlibc(7)

	var c Combat
	c.Side = 1
	c.UnitID = 0
	c.Units[1][0] = Unit{TroopID: 3, Count: 5, TurnCount: 5, Moves: 3, X: 2, Y: 2} // 野狼
	c.Units[0][0] = Unit{TroopID: 0, Count: 5, TurnCount: 5, X: 3, Y: 2}           // 農夫,相鄰
	c.Umap[2][2] = PackUID(1, 0)
	c.Umap[2][3] = PackUID(0, 0)

	beforeDefCount := c.Units[0][0].Count

	pass := c.AIUnitThink(a, rng)
	if pass {
		t.Fatal("相鄰有目標時不該 pass=true")
	}
	if !c.Units[1][0].Acted {
		t.Error("近戰攻擊後應標記 Acted")
	}
	if c.Units[1][0].X != 2 || c.Units[1][0].Y != 2 {
		t.Error("近戰攻擊不該移動攻擊者位置")
	}
	after := c.Units[0][0]
	if after.Count == beforeDefCount && after.Injury == 0 && !after.Dead {
		t.Error("被近戰攻擊的目標應該受傷、減員或陣亡")
	}
}

// TestCombat_PlayerPassiveAIConverges 對應任務要求①:佈陣後跑「玩家不動、
// AI 自動打」數回合,驗證戰鬥會收斂(不無限迴圈、有勝負),驅動迴圈本身
// 比照 C combat_loop 的 `if (pass || units[...].acted) { next_unit/next_turn }`
// 判斷(見 AIUnitThink 文件的 pass 語意說明)。
func TestCombat_PlayerPassiveAIConverges(t *testing.T) {
	a := loadTestAssets(t)
	rng := kbrng.NewGlibc(99)

	var c Combat
	c.Units[0][0] = Unit{TroopID: 0, Count: 5, X: 0, Y: 2}          // 玩家:農夫 x5
	c.Units[1][0] = Unit{TroopID: 3, Count: 4, X: BoardW - 1, Y: 2} // 敵方:野狼 x4

	c.ResetMatch(a, rng, false)
	// 本測試只驗證回合/AI 收斂,不驗證障礙產生演算法本身;清空隨機障礙,
	// 避免極罕見的隨機佈局把路徑完全擋死而干擾這裡要驗證的目標。
	for y := 0; y < BoardH; y++ {
		for x := 0; x < BoardW; x++ {
			c.Omap[y][x] = 0
		}
	}

	const maxTicks = 5000
	winner := -1
	ticks := 0
	for ; ticks < maxTicks; ticks++ {
		if w := c.Winner(); w != -1 {
			winner = w
			break
		}

		u := &c.Units[c.Side][c.UnitID]
		pass := true
		if u.Count > 0 {
			if c.Side == 0 {
				// 玩家不動:模擬玩家每次心跳都選待命,對應 unit_try_wait 恆
				// 回傳 pass=true(見 AIUnitThink 文件的 pass 語意說明)。
				if c.Phase > 0 {
					u.Acted = true
				}
			} else {
				pass = c.AIUnitThink(a, rng)
			}
		}

		if pass || u.Acted {
			next := c.NextUnit(a)
			if next == -1 {
				if !c.NextTurn(a) {
					t.Fatalf("tick %d: NextTurn 找不到下一個可行動單位,但戰鬥尚未分出勝負", ticks)
				}
			} else {
				c.UnitID = next
			}
		}
	}

	if winner == -1 {
		t.Fatalf("戰鬥在 %d tick 內未收斂出勝負(可能是 AI/回合邏輯卡死)", maxTicks)
	}
	t.Logf("戰鬥在 %d tick 內收斂,勝方 side=%d", ticks, winner)
}
