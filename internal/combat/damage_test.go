package combat

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

type combatCase struct {
	Atk       int `json:"atk"`
	AtkN      int `json:"atk_n"`
	Def       int `json:"def"`
	DefN      int `json:"def_n"`
	Kills     int `json:"kills"`
	DefLeft   int `json:"def_left"`
	DefInjury int `json:"def_injury"`
}

type combatGolden struct {
	Seed        int          `json:"seed"`
	CombatMelee []combatCase `json:"combat_melee"`
}

// TestDealDamage_ParityWithCOracle 對 C 版 deal_damage 的黃金樣本逐案例比對:
// 相同 seed、相同攻守配置,Go 的 UnitHitUnit 應算出相同 kills/剩餘/injury。
// golden 由 openkb-cht 的 KB_ORACLE(melee 1v1, retaliation=1)產生。
func TestDealDamage_ParityWithCOracle(t *testing.T) {
	raw, err := os.ReadFile("testdata/golden_seed1.json")
	if err != nil {
		t.Fatalf("讀 golden: %v", err)
	}
	var g combatGolden
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatal(err)
	}
	if len(g.CombatMelee) == 0 {
		t.Fatal("golden 無 combat_melee 案例")
	}

	for i, cse := range g.CombatMelee {
		c := &Combat{}
		c.Units[0][0] = Unit{TroopID: cse.Atk, Count: cse.AtkN, TurnCount: cse.AtkN, MaxCount: cse.AtkN}
		c.Units[1][0] = Unit{TroopID: cse.Def, Count: cse.DefN, TurnCount: cse.DefN, MaxCount: cse.DefN}
		rng := kbrng.NewGlibc(uint32(g.Seed))

		kills := c.UnitHitUnit(a, rng, 0, 0, 1, 0)

		if kills != cse.Kills {
			t.Errorf("案例 %d (atk %d×%d → def %d×%d): kills Go=%d, C=%d",
				i, cse.Atk, cse.AtkN, cse.Def, cse.DefN, kills, cse.Kills)
		}
		if got := c.Units[1][0].Count; got != cse.DefLeft {
			t.Errorf("案例 %d: def_left Go=%d, C=%d", i, got, cse.DefLeft)
		}
		if got := c.Units[1][0].Injury; got != cse.DefInjury {
			t.Errorf("案例 %d: def_injury Go=%d, C=%d", i, got, cse.DefInjury)
		}
	}
}
