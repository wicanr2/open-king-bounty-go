package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// TestBuildBridge_Horizontal 驗證往右(dirX=1)在水域上架 2 格 TileBridgeH。
func TestBuildBridge_Horizontal(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("WorldMap 未載入,略過")
	}
	gs.Continent = 0
	gs.X, gs.Y = 10, 10
	gs.WorldMap.Set(0, 11, 10, 0x14)
	gs.WorldMap.Set(0, 12, 10, 0x14)

	if got := gs.BuildBridge(1, 0); got != 2 {
		t.Fatalf("BuildBridge(1,0) = %d, want 2", got)
	}
	if tile := gs.WorldMap.Tile(0, 11, 10); tile != kbdata.TileBridgeH {
		t.Errorf("(11,10) tile = %#x, want TileBridgeH(%#x)", tile, kbdata.TileBridgeH)
	}
	if tile := gs.WorldMap.Tile(0, 12, 10); tile != kbdata.TileBridgeH {
		t.Errorf("(12,10) tile = %#x, want TileBridgeH(%#x)", tile, kbdata.TileBridgeH)
	}
}

// TestBuildBridge_NoWater 驗證方向上沒有水域時白白浪費法術,回傳 0。
func TestBuildBridge_NoWater(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("WorldMap 未載入,略過")
	}
	gs.Continent = 0
	gs.X, gs.Y = 10, 10
	gs.WorldMap.Set(0, 11, 10, 0) // 草地,非水域

	if got := gs.BuildBridge(1, 0); got != 0 {
		t.Fatalf("BuildBridge(1,0) = %d, want 0(無水可造橋)", got)
	}
}

// TestBuildBridge_Vertical 驗證往上(dirY=-1,對齊 C dj=-1)在水域上架 2 格 TileBridgeV;
// y = Y - i*dirY = 10 + i,故驗證座標為 (10,11)/(10,12)。
func TestBuildBridge_Vertical(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("WorldMap 未載入,略過")
	}
	gs.Continent = 0
	gs.X, gs.Y = 10, 10
	gs.WorldMap.Set(0, 10, 11, 0x14)
	gs.WorldMap.Set(0, 10, 12, 0x14)

	if got := gs.BuildBridge(0, -1); got != 2 {
		t.Fatalf("BuildBridge(0,-1) = %d, want 2", got)
	}
	if tile := gs.WorldMap.Tile(0, 10, 11); tile != kbdata.TileBridgeV {
		t.Errorf("(10,11) tile = %#x, want TileBridgeV(%#x)", tile, kbdata.TileBridgeV)
	}
	if tile := gs.WorldMap.Tile(0, 10, 12); tile != kbdata.TileBridgeV {
		t.Errorf("(10,12) tile = %#x, want TileBridgeV(%#x)", tile, kbdata.TileBridgeV)
	}
}

// TestFindVillain_Found 驗證找到 last_contract 惡棍所佔城堡時回傳其 id 並標記 KBCastleKnown。
func TestFindVillain_Found(t *testing.T) {
	gs := &GameState{}
	gs.LastContract = 3
	gs.CastleOwner[7] = 3

	if got := gs.FindVillain(); got != 7 {
		t.Fatalf("FindVillain() = %d, want 7", got)
	}
	if gs.CastleOwner[7]&KBCastleKnown == 0 {
		t.Errorf("找到後應標記 KBCastleKnown,got owner=%#x", gs.CastleOwner[7])
	}
}

// TestFindVillain_NotFound 驗證沒有城堡符合 last_contract 時回傳 -1。
func TestFindVillain_NotFound(t *testing.T) {
	gs := &GameState{}
	gs.LastContract = 3
	// CastleOwner 全 0,沒有任何城堡的 owner&0x3f == 3

	if got := gs.FindVillain(); got != -1 {
		t.Fatalf("FindVillain() = %d, want -1", got)
	}
}
