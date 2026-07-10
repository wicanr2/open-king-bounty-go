package gamestate

import "testing"

// TestPuzzleOpened 驗證拼圖掀開判定:神器格看 ArtifactFound、惡棍格看 VillainCaught。
func TestPuzzleOpened(t *testing.T) {
	gs := &GameState{}

	// puzzleMap[0][0] = -1 → 神器 0。未找到 → 未掀開。
	if gs.PuzzleOpened(0, 0) {
		t.Error("未找到神器 0,格 (0,0) 不應掀開")
	}
	gs.ArtifactFound[0] = 1
	if !gs.PuzzleOpened(0, 0) {
		t.Error("找到神器 0 後格 (0,0) 應掀開")
	}

	// puzzleMap[2][2] = 16 → 惡棍 16(中心格,權杖所在)。
	if gs.PuzzleOpened(2, 2) {
		t.Error("未捕獲惡棍 16,中心格不應掀開")
	}
	gs.VillainCaught[16] = 1
	if !gs.PuzzleOpened(2, 2) {
		t.Error("捕獲惡棍 16 後中心格應掀開")
	}

	// 界外 → false。
	if gs.PuzzleOpened(-1, 0) || gs.PuzzleOpened(0, 5) {
		t.Error("界外格應回 false")
	}
}

// TestPuzzleTile 驗證拼圖格對應的地圖座標 = 權杖周邊(中心 = 權杖)。
func TestPuzzleTile(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("land.org 未載入,略過")
	}
	// 中心格 (2,2) 應對到權杖所在 tile。
	center := gs.PuzzleTile(2, 2)
	want := gs.WorldMap.Tile(gs.ScepterContinent, gs.ScepterX, gs.ScepterY)
	if center != want {
		t.Errorf("拼圖中心格 tile = %#x, 應為權杖格 %#x", center, want)
	}
}
