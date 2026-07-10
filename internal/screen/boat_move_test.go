package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// buildBoatRow 在 gs.WorldMap 洲 0 的 y=8 列鋪一條「草-水-水-草」測試地形,並把玩家
// 放在最左草地。tile 值:草=0x00(IsGrass)、水=0x14(IsWater)。
func buildBoatRow(t *testing.T) (*WorldMapScreen, *gamestate.GameState) {
	t.Helper()
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("embedded land.org 未載入,略過")
	}
	gs.Continent, gs.X, gs.Y = 0, 10, 8
	gs.WorldMap.Set(0, 10, 8, 0x00) // 草(起點)
	gs.WorldMap.Set(0, 11, 8, 0x14) // 水(船停靠格)
	gs.WorldMap.Set(0, 12, 8, 0x14) // 水
	gs.WorldMap.Set(0, 13, 8, 0x00) // 草(上岸)
	return NewWorldMapScreen(gs, a), gs
}

// TestBoat_BlockedWithoutBoat 驗證沒船時走不進水域(留在原地)。
func TestBoat_BlockedWithoutBoat(t *testing.T) {
	s, gs := buildBoatRow(t)
	gs.Boat = 0xFF // 無船
	gs.Mount = gamestate.KBMountRide

	s.Update(input.Action{Kind: input.ActRight})
	if gs.X != 10 {
		t.Errorf("無船卻走進水域:X=%d, want 10(擋下)", gs.X)
	}
}

// TestBoat_BoardSailDisembark 驗證登船→航行→上岸整段流程。
func TestBoat_BoardSailDisembark(t *testing.T) {
	s, gs := buildBoatRow(t)
	gs.Boat = 0 // 船在洲 0
	gs.BoatX, gs.BoatY = 11, 8
	gs.Mount = gamestate.KBMountRide

	// 1) 走進 (11,8) 船停靠格 → 登船。
	s.Update(input.Action{Kind: input.ActRight})
	if gs.X != 11 || gs.Mount != gamestate.KBMountSail {
		t.Fatalf("登船失敗:X=%d Mount=%d, want 11/SAIL", gs.X, gs.Mount)
	}

	// 2) 續航到 (12,8) 水域 → 仍乘船,船隨行。
	s.Update(input.Action{Kind: input.ActRight})
	if gs.X != 12 || gs.Mount != gamestate.KBMountSail {
		t.Fatalf("航行失敗:X=%d Mount=%d, want 12/SAIL", gs.X, gs.Mount)
	}
	if gs.BoatX != 12 || gs.BoatY != 8 {
		t.Errorf("航行時船未隨行:boat=(%d,%d), want (12,8)", gs.BoatX, gs.BoatY)
	}

	// 3) 走上 (13,8) 草地 → 下船,船留在上一格水域 (12,8)。
	s.Update(input.Action{Kind: input.ActRight})
	if gs.X != 13 || gs.Mount != gamestate.KBMountRide {
		t.Fatalf("上岸失敗:X=%d Mount=%d, want 13/RIDE", gs.X, gs.Mount)
	}
	if gs.BoatX != 12 || gs.BoatY != 8 {
		t.Errorf("上岸後船應留在上一格水域 (12,8),got (%d,%d)", gs.BoatX, gs.BoatY)
	}
}
