package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// TestPlaceScepter 驗證埋權杖:選出合法洲、埋在草地格,AtScepter 在該格為真、他處為假。
func TestPlaceScepter(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("land.org 未載入,略過")
	}
	// NewGame 已呼叫 PlaceScepter;驗證結果落在草地。
	if gs.ScepterContinent < 0 || gs.ScepterContinent >= 4 {
		t.Fatalf("權杖洲別越界:%d", gs.ScepterContinent)
	}
	tile := gs.WorldMap.Tile(gs.ScepterContinent, gs.ScepterX, gs.ScepterY)
	if !isGrassTile(tile) {
		t.Errorf("權杖未埋在草地:tile=%#x @ (%d,%d,%d)", tile, gs.ScepterContinent, gs.ScepterX, gs.ScepterY)
	}

	// 站在權杖處 → AtScepter true。
	gs.Continent, gs.X, gs.Y = gs.ScepterContinent, gs.ScepterX, gs.ScepterY
	if !gs.AtScepter() {
		t.Error("站在權杖埋藏處 AtScepter 應為 true")
	}
	// 移開一格 → false。
	gs.X++
	if gs.AtScepter() {
		t.Error("離開權杖格後 AtScepter 應為 false")
	}
}

// TestPlaceScepter_Deterministic 驗證同 seed 埋點可重現(同一 rng 序列)。
func TestPlaceScepter_Deterministic(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("land.org 未載入,略過")
	}
	c, x, y := gs.ScepterContinent, gs.ScepterX, gs.ScepterY
	gs.PlaceScepter(kbrng.NewGlibc(42))
	c2, x2, y2 := gs.ScepterContinent, gs.ScepterX, gs.ScepterY
	gs.PlaceScepter(kbrng.NewGlibc(42))
	if gs.ScepterContinent != c2 || gs.ScepterX != x2 || gs.ScepterY != y2 {
		t.Error("同 seed 埋點應可重現")
	}
	_ = c
	_ = x
	_ = y
}
