package combat

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// TestAIPickTargetEvolved_PrefersDangerousTarget 驗證進化版選敵偏好高威脅(高 SL/傷)目標,
// 而非原版的「最低 HP」——場上農夫(低分)與火龍(高 SL/傷)並存時應選火龍。
func TestAIPickTargetEvolved_PrefersDangerousTarget(t *testing.T) {
	a := loadTestAssets(t)
	var c Combat
	c.Side = 1
	c.UnitID = 0
	c.Units[1][0] = Unit{TroopID: 0, Count: 5, X: 0, Y: 0}     // 行動者
	c.Units[0][0] = Unit{TroopID: 0, Count: 5, X: 5, Y: 0}     // 敵:農夫(低分)
	c.Units[0][1] = Unit{TroopID: 0x18, Count: 1, X: 5, Y: 4}  // 敵:火龍(高 SL/傷 → 高分)

	side, id, ok := c.AIPickTargetEvolved(a, false)
	if !ok {
		t.Fatal("應選到目標")
	}
	if side != 0 || id != 1 {
		t.Errorf("進化版應優先高威脅(火龍)= (0,1),got (%d,%d)", side, id)
	}
}

// TestAIUnitThinkEvolved_SurroundedArcherMelees 驗證進化版 AI 的新增行為:被貼身而無法射擊的
// 射手改近戰攻擊,而非原版的空等。AIEvolved=true 時 AIUnitThink dispatch 到進化版。
func TestAIUnitThinkEvolved_SurroundedArcherMelees(t *testing.T) {
	a := loadTestAssets(t)
	rng := kbrng.NewGlibc(3)
	var c Combat
	c.AIEvolved = true
	c.Side = 1
	c.UnitID = 0
	// 行動者=野狼但仍有彈藥(Shots>0),用來觸發「射手被貼身」分支;貼身敵人相鄰。
	c.Units[1][0] = Unit{TroopID: 3, Count: 5, TurnCount: 5, Shots: 2, Moves: 1, X: 2, Y: 2}
	c.Units[0][0] = Unit{TroopID: 0, Count: 5, TurnCount: 5, X: 3, Y: 2} // 貼身敵人(農夫)
	c.Umap[2][2] = PackUID(1, 0)
	c.Umap[2][3] = PackUID(0, 0)

	beforeCount := c.Units[0][0].Count
	c.AIUnitThink(a, rng) // AIEvolved → 進化版:被貼身射手改近戰
	if !c.Units[1][0].Acted {
		t.Error("進化版:被貼身的射手應行動(近戰),而非空等")
	}
	after := c.Units[0][0]
	if after.Count == beforeCount && after.Injury == 0 && !after.Dead {
		t.Error("進化版:貼身射手應近戰造成傷害(敵方受傷/減員/陣亡)")
	}
}
