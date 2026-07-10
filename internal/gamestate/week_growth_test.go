package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// TestWeeklyWorldGrowth 驗證每週世界成長:本週生物的巢穴補滿、foe/非玩家城堡 +growth、
// 停時歸零。
func TestWeeklyWorldGrowth(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)

	creature := 6 // 地精(有 growth/max_population)
	growth := a.Troops[creature].Growth
	maxPop := a.Troops[creature].MaxPopulation

	// 佈置:洲0巢穴0是該生物但庫存被掏空;洲0 foe0 第0格是該生物;非玩家城堡5第0格是該生物。
	gs.DwellingTroop[0][0] = creature
	gs.DwellingPopulation[0][0] = 0
	gs.FoeTroops[0][0][0] = creature
	gs.FoeNumbers[0][0][0] = 10
	gs.CastleOwner[5] = castleOwnerMonsters
	gs.CastleTroops[5][0] = creature
	gs.CastleNumbers[5][0] = 20

	gs.weeklyWorldGrowth(a, kbrng.NewGlibc(1), creature)

	if gs.DwellingPopulation[0][0] != maxPop {
		t.Errorf("巢穴應補到 max_population %d,got %d", maxPop, gs.DwellingPopulation[0][0])
	}
	if gs.FoeNumbers[0][0][0] != 10+growth {
		t.Errorf("foe 應 +growth(%d):got %d, want %d", growth, gs.FoeNumbers[0][0][0], 10+growth)
	}
	if gs.CastleNumbers[5][0] != 20+growth {
		t.Errorf("非玩家城堡應 +growth:got %d, want %d", gs.CastleNumbers[5][0], 20+growth)
	}
}

// TestEndWeekResetsTimeStop 驗證 EndWeek 把停時歸零。
func TestEndWeekResetsTimeStop(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.TimeStop = 50
	gs.EndWeek(a, kbrng.NewGlibc(1))
	if gs.TimeStop != 0 {
		t.Errorf("EndWeek 後 TimeStop 應歸零,got %d", gs.TimeStop)
	}
}
