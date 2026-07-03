package kbdata

import (
	"path/filepath"
	"testing"
)

func testLandPath(t *testing.T) string {
	t.Helper()
	return filepath.Join("testdata", "land.org")
}

func TestLoadWorldMap_ShapeAndSize(t *testing.T) {
	m, err := LoadWorldMap(testLandPath(t))
	if err != nil {
		t.Fatalf("LoadWorldMap failed: %v", err)
	}
	if m.Continents != MaxContinents {
		t.Fatalf("expected %d continents, got %d", MaxContinents, m.Continents)
	}
	if len(m.Tiles) != MaxContinents {
		t.Fatalf("expected %d continent grids, got %d", MaxContinents, len(m.Tiles))
	}
	for cont, grid := range m.Tiles {
		if len(grid) != LevelH {
			t.Fatalf("continent %d: expected %d rows, got %d", cont, LevelH, len(grid))
		}
		for y, row := range grid {
			if len(row) != LevelW {
				t.Fatalf("continent %d row %d: expected %d cols, got %d", cont, y, LevelW, len(row))
			}
		}
	}
}

// TestLoadWorldMap_KnownTilesPresent 驗證真實 free 資料檔(openkb-code/data/free/land.org
// 複製而來)含有已知的具名 tile 值,確認 tile byte 解讀語意合理。
// 對照 python 靜態掃描本檔:0x20(TileDeepWater)最多(4929)、0x8A(TileTown)恰 26 格、
// 0x8B(TileChest)178 格、0x91(TileFoe)92 格 —— 三者皆 > 0,證明資料非全零且解析對齊。
func TestLoadWorldMap_KnownTilesPresent(t *testing.T) {
	m, err := LoadWorldMap(testLandPath(t))
	if err != nil {
		t.Fatalf("LoadWorldMap failed: %v", err)
	}

	counts := map[byte]int{}
	for cont := 0; cont < m.Continents; cont++ {
		for y := 0; y < LevelH; y++ {
			for x := 0; x < LevelW; x++ {
				counts[m.Tile(cont, x, y)]++
			}
		}
	}

	wantPresent := []byte{TileGrass, TileDeepWater, TileTown, TileChest, TileFoe}
	for _, tile := range wantPresent {
		if counts[tile] == 0 {
			t.Errorf("expected tile 0x%02X to appear at least once in world map, got 0 occurrences", tile)
		}
	}

	// 與 python 靜態掃描 testdata/land.org 得到的精確計數交叉核實
	// (見任務調查時跑的一次性 Counter(data) 統計)。
	wantExact := map[byte]int{
		TileTown:  26,
		TileChest: 178,
		TileFoe:   92,
	}
	for tile, want := range wantExact {
		if got := counts[tile]; got != want {
			t.Errorf("tile 0x%02X: expected exactly %d occurrences, got %d", tile, want, got)
		}
	}
}

func TestWorldMap_TileOutOfBoundsSafe(t *testing.T) {
	m, err := LoadWorldMap(testLandPath(t))
	if err != nil {
		t.Fatalf("LoadWorldMap failed: %v", err)
	}

	cases := []struct {
		name       string
		cont, x, y int
	}{
		{"negative continent", -1, 0, 0},
		{"continent too large", MaxContinents, 0, 0},
		{"negative x", 0, -1, 0},
		{"negative y", 0, 0, -1},
		{"x too large", 0, LevelW, 0},
		{"y too large", 0, 0, LevelH},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := m.Tile(tc.cont, tc.x, tc.y); got != TileGrass {
				t.Fatalf("out-of-bounds Tile(%d,%d,%d) = 0x%02X, want safe default 0x%02X",
					tc.cont, tc.x, tc.y, got, TileGrass)
			}
		})
	}
}

func TestWorldMap_NilSafe(t *testing.T) {
	var m *WorldMap
	if got := m.Tile(0, 0, 0); got != TileGrass {
		t.Fatalf("nil *WorldMap.Tile = 0x%02X, want 0x%02X", got, TileGrass)
	}
}

func TestParseWorldMap_TooShort(t *testing.T) {
	_, err := ParseWorldMap(make([]byte, 100))
	if err == nil {
		t.Fatalf("expected error for truncated data, got nil")
	}
}

// TestParseWorldMap_SyntheticLayout 用合成 bytes 驗證 continent-major →
// row(y) → column(x) 的解析順序,對齊 src/play.c:380 memcpy 語意與
// src/tools/kbtmx.py:495-497 的寫檔迴圈次序。每個位置填入可從 (cont,y,x)
// 反推的獨特值,解析後逐一比對。
func TestParseWorldMap_SyntheticLayout(t *testing.T) {
	data := make([]byte, worldMapBytes)
	off := 0
	for cont := 0; cont < MaxContinents; cont++ {
		for y := 0; y < LevelH; y++ {
			for x := 0; x < LevelW; x++ {
				// 值域刻意錯開,讓 cont/y/x 任一錯位都會被抓到。
				data[off] = byte((cont*40 + y + x) % 256)
				off++
			}
		}
	}

	m, err := ParseWorldMap(data)
	if err != nil {
		t.Fatalf("ParseWorldMap failed: %v", err)
	}

	for cont := 0; cont < MaxContinents; cont++ {
		for _, pt := range []struct{ x, y int }{{0, 0}, {63, 0}, {0, 63}, {63, 63}, {17, 42}} {
			want := byte((cont*40 + pt.y + pt.x) % 256)
			if got := m.Tile(cont, pt.x, pt.y); got != want {
				t.Fatalf("Tile(cont=%d,x=%d,y=%d) = 0x%02X, want 0x%02X", cont, pt.x, pt.y, got, want)
			}
		}
	}
}

func TestTileSemanticHelpers(t *testing.T) {
	cases := []struct {
		name string
		fn   func(byte) bool
		yes  []byte
		no   []byte
	}{
		{"IsGrass", IsGrass, []byte{0x00, 0x01, 0x80}, []byte{0x02, 0x20, 0x85}},
		{"IsCastle", IsCastle, []byte{0x02, 0x07}, []byte{0x00, 0x08, 0x85}},
		{"IsBridge", IsBridge, []byte{TileBridgeH, TileBridgeV}, []byte{0x00, 0x0A, 0x20}},
		{"IsMapObject", IsMapObject, []byte{0x0A, 0x13}, []byte{0x09, 0x14}},
		{"IsWater", IsWater, []byte{0x14, TileDeepWater}, []byte{0x13, 0x21}},
		{"IsTree", IsTree, []byte{0x21, 0x2D}, []byte{0x20, 0x2E}},
		{"IsDesert", IsDesert, []byte{0x2E, 0x3A}, []byte{0x2D, 0x3B}},
		{"IsRock", IsRock, []byte{0x3B, 0x47}, []byte{0x3A, 0x48}},
		{"IsDeepWater", IsDeepWater, []byte{TileDeepWater}, []byte{0x1F, 0x21}},
		{"IsInteractive", IsInteractive, []byte{0x80, TileTown, TileChest}, []byte{0x00, 0x7F}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, v := range tc.yes {
				if !tc.fn(v) {
					t.Errorf("%s(0x%02X) = false, want true", tc.name, v)
				}
			}
			for _, v := range tc.no {
				if tc.fn(v) {
					t.Errorf("%s(0x%02X) = true, want false", tc.name, v)
				}
			}
		})
	}
}
