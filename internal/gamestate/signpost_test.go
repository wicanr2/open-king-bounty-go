package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// TestSignpostIndex 驗證路標索引依 (continent,y,x) 掃描序計數。
func TestSignpostIndex(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("land.org 未載入,略過")
	}
	// 清掉洲 0 既有路標干擾,擺三個已知路標(掃描序 = y 由小到大、同 y 再 x)。
	for y := 0; y < kbdata.LevelH; y++ {
		for x := 0; x < kbdata.LevelW; x++ {
			if gs.WorldMap.Tile(0, x, y) == kbdata.TileSignpost {
				gs.WorldMap.Set(0, x, y, 0)
			}
		}
	}
	gs.WorldMap.Set(0, 10, 5, kbdata.TileSignpost) // 索引 0(y 最小)
	gs.WorldMap.Set(0, 3, 8, kbdata.TileSignpost)  // 索引 1
	gs.WorldMap.Set(0, 40, 8, kbdata.TileSignpost) // 索引 2(同 y=8,x 較大)

	gs.Continent = 0
	gs.X, gs.Y = 3, 8
	if idx := gs.SignpostIndex(); idx != 1 {
		t.Errorf("(3,8) 路標索引 = %d, want 1", idx)
	}
	gs.X, gs.Y = 40, 8
	if idx := gs.SignpostIndex(); idx != 2 {
		t.Errorf("(40,8) 路標索引 = %d, want 2", idx)
	}
	// 不在路標上 → -1。
	gs.X, gs.Y = 1, 1
	if idx := gs.SignpostIndex(); idx != -1 {
		t.Errorf("非路標格索引 = %d, want -1", idx)
	}
}
