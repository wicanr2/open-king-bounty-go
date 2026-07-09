package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// 黃金樣本由 C 版 KB_ORACLE=1 KB_ORACLE_SEED=1 實跑產出(spawn_game 起手數值,
// seed 不影響這些欄位,只影響世界生成,本切片不涉及)。
func TestNewGame_GoldenSamples(t *testing.T) {
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatalf("kbdata.Load: %v", err)
	}

	tests := []struct {
		class          int
		name           string
		gold           int
		baseLeadership int
		leadership     int
		commission     int
		troops         [5]int
		numbers        [5]int
	}{
		{
			class: 0, name: "騎士",
			gold: 7500, baseLeadership: 100, leadership: 100, commission: 1000,
			troops:  [5]int{2, 8, 255, 255, 255},
			numbers: [5]int{20, 2, 0, 0, 0},
		},
		{
			class: 1, name: "遊俠",
			gold: 10000, baseLeadership: 80, leadership: 80, commission: 1000,
			troops:  [5]int{0, 2, 255, 255, 255},
			numbers: [5]int{20, 20, 0, 0, 0},
		},
		{
			class: 2, name: "女巫師",
			gold: 10000, baseLeadership: 60, leadership: 60, commission: 3000,
			troops:  [5]int{0, 1, 255, 255, 255},
			numbers: [5]int{30, 10, 0, 0, 0},
		},
		{
			class: 3, name: "蠻俠",
			gold: 7500, baseLeadership: 100, leadership: 100, commission: 2000,
			troops:  [5]int{3, 255, 255, 255, 255},
			numbers: [5]int{20, 0, 0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := NewGame(a, "測試角色", tt.class, DefaultWorldSeed)

			if gs.Class != tt.class {
				t.Errorf("Class: got %d, want %d", gs.Class, tt.class)
			}
			if gs.Rank != 0 {
				t.Errorf("Rank: got %d, want 0", gs.Rank)
			}
			if gs.Gold != tt.gold {
				t.Errorf("Gold: got %d, want %d", gs.Gold, tt.gold)
			}
			if gs.BaseLeadership != tt.baseLeadership {
				t.Errorf("BaseLeadership: got %d, want %d", gs.BaseLeadership, tt.baseLeadership)
			}
			if gs.Leadership != tt.leadership {
				t.Errorf("Leadership: got %d, want %d", gs.Leadership, tt.leadership)
			}
			if gs.Commission != tt.commission {
				t.Errorf("Commission: got %d, want %d", gs.Commission, tt.commission)
			}

			for i := 0; i < 5; i++ {
				if gs.Army[i].TroopID != tt.troops[i] {
					t.Errorf("Army[%d].TroopID: got %d, want %d", i, gs.Army[i].TroopID, tt.troops[i])
				}
				if gs.Army[i].Count != tt.numbers[i] {
					t.Errorf("Army[%d].Count: got %d, want %d", i, gs.Army[i].Count, tt.numbers[i])
				}
			}
		})
	}
}

// TestNewGame_SorceressKnowsMagic 驗證只有女巫師起手天生會魔法。
func TestNewGame_SorceressKnowsMagic(t *testing.T) {
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatalf("kbdata.Load: %v", err)
	}

	for class := 0; class < 4; class++ {
		gs := NewGame(a, "測試角色", class, DefaultWorldSeed)
		want := class == 2
		if gs.KnowsMagic != want {
			t.Errorf("class %d KnowsMagic: got %v, want %v", class, gs.KnowsMagic, want)
		}
	}
}
