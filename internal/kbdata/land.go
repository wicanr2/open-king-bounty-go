// land.go -- 世界地圖(DAT_WORLD / land.org)載入器。
//
// C 版對照:
//
//	尺寸/型別:src/bounty.h:39-41 (MAX_CONTINENTS=4, LEVEL_W=64, LEVEL_H=64)、
//	          src/bounty.h:170 (KBgame.map[MAX_CONTINENTS][LEVEL_H][LEVEL_W])。
//
//	資料佈局(誰決定 land 原始 byte 怎麼切成 [cont][y][x]):
//	  src/play.c:371 spawn_game(..., byte *land) 收到已解析好的整塊 world buffer,
//	  src/play.c:380 `memcpy(game->map, land, LEVEL_W * LEVEL_H * MAX_CONTINENTS);`
//	  是唯一的佈局依據 —— 一次 memcpy 整塊複製進 map[MAX_CONTINENTS][LEVEL_H][LEVEL_W],
//	  代表 land 的原始位元組順序就是 C 陣列的記憶體順序:continent 最外層、
//	  然後 row(y)、最後 column(x)(row-major,continent-major)。
//	  這個順序另外由地圖編輯器 src/tools/kbtmx.py:495-497 的寫檔迴圈
//	  `for cont in range(0,4): for y in range(0,64): for x in range(0,64)` 驗證一致。
//
//	原始檔案來源(DAT_WORLD 解析到哪個檔):
//	  src/lib/free-data.c:1438-1452 與 src/lib/dos-data.c:1652-1666 的
//	  `case DAT_WORLD` 都是「原封不動讀一個 64*64*4 = 16384 bytes 的檔」,
//	  沒有 header、沒有壓縮:
//	    free 版: KB_fopen_with("land.org", "rb", mod)   (小寫檔名)
//	    DOS  版: KB_fopen_with("LAND.ORG", "rb", mod)   (大寫檔名,同格式)
//	  本 repo 附的 free 資料檔案在 openkb-code/data/free/land.org,
//	  實測檔案大小恰為 16384 bytes,與此假設吻合。
//	  land.ini(同目錄)是另外的中繼資料(洲名、國王/密室座標、導航點),
//	  不影響 tile grid 本身,本 loader 不處理它。
//
//	tile byte 語意(src/lib/kbres.h:279-306,`IS_*` 巨集為真值):
//	  0x00-0x01, 0x80        grass/plain            (IS_GRASS)
//	  0x02-0x07              castle 邊界圖磚          (IS_CASTLE)
//	  0x08-0x09              bridge (H/V)            (IS_BRIDGE; TILE_BRIDGE_H/V)
//	  0x0A-0x13              map object(小地物)       (IS_MAPOBJECT)
//	  0x14-0x20              water,0x20 為深水底線     (IS_WATER; TILE_DEEP_WATER=0x20)
//	  0x21-0x2D              tree/森林                (IS_TREE)
//	  0x2E-0x3A              desert/沙漠              (IS_DESERT)
//	  0x3B-0x47              rock/山地                (IS_ROCK)
//	  0x80-0xFF (高位元=1)    interactive(城堡/城鎮/箱子/棲地/路標/敵人/神器…)
//	                          (IS_INTERACTIVE = M & 0x80)
//	  個別具名值:TILE_CASTLE=0x85、TILE_TOWN=0x8A、TILE_CHEST=0x8B、
//	    TILE_DWELLING_1..4=0x8C..0x8F(TILE_TELECAVE=TILE_DWELLING_3=0x8E)、
//	    TILE_SIGNPOST=0x90、TILE_FOE=0x91、TILE_ARTIFACT_1/2=0x92/0x93。
//
// 本檔為純資料層,只做「原始 byte → 型別化網格」與 tile 語意常量/判斷,
// 不處理 land.ini 中繼資料、不做地圖編輯、不做 furnishing(src/rogue.c 的
// 圖磚轉角美化,屬 render 邏輯,不在本 loader 範圍)。
package kbdata

import (
	"fmt"
	"os"
)

// 世界地圖尺寸,對齊 src/bounty.h:39-41。
const (
	LevelW        = 64 // LEVEL_W
	LevelH        = 64 // LEVEL_H
	MaxContinents = 4  // MAX_CONTINENTS
)

// worldMapBytes 是 land.org 應有的原始檔案大小:
// LevelW * LevelH * MaxContinents,對齊 src/play.c:380 的 memcpy 長度、
// 以及 src/lib/free-data.c:1440 / dos-data.c:1655 的 `len = 64*64*4`。
const worldMapBytes = LevelW * LevelH * MaxContinents

// Tile 具名值,對齊 src/lib/kbres.h:280-295 的 TILE_* 定義。
const (
	TileGrass     byte = 0    // TILE_GRASS
	TileBridgeH   byte = 8    // TILE_BRIDGE_H
	TileBridgeV   byte = 9    // TILE_BRIDGE_V
	TileDeepWater byte = 32   // TILE_DEEP_WATER (0x20)
	TileCastle    byte = 0x85 // TILE_CASTLE
	TileTown      byte = 0x8A // TILE_TOWN
	TileChest     byte = 0x8B // TILE_CHEST
	TileDwelling1 byte = 0x8C // TILE_DWELLING_1
	TileDwelling2 byte = 0x8D // TILE_DWELLING_2
	TileDwelling3 byte = 0x8E // TILE_DWELLING_3
	TileDwelling4 byte = 0x8F // TILE_DWELLING_4
	TileTelecave  byte = TileDwelling3
	TileSignpost  byte = 0x90 // TILE_SIGNPOST
	TileFoe       byte = 0x91 // TILE_FOE
	TileArtifact1 byte = 0x92 // TILE_ARTIFACT_1
	TileArtifact2 byte = 0x93 // TILE_ARTIFACT_2
)

// IsGrass 對齊 kbres.h:298 `IS_GRASS(M) ((M) < 2 || (M) == 0x80)`。
func IsGrass(tile byte) bool { return tile < 2 || tile == 0x80 }

// IsCastle 對齊 kbres.h:299 `IS_CASTLE(M) ((M) >= 0x02 && (M) <= 0x07)`。
func IsCastle(tile byte) bool { return tile >= 0x02 && tile <= 0x07 }

// IsBridge 對齊 kbres.h:301 `IS_BRIDGE(M) ((M) >= 0x08 && (M) <= 0x09)`。
func IsBridge(tile byte) bool { return tile >= 0x08 && tile <= 0x09 }

// IsMapObject 對齊 kbres.h:300 `IS_MAPOBJECT(M) ((M) >= 0x0a && (M) <= 0x13)`。
func IsMapObject(tile byte) bool { return tile >= 0x0A && tile <= 0x13 }

// IsWater 對齊 kbres.h:302 `IS_WATER(M) ((M) >= 0x14 && (M) <= 0x20)`(含深水底線值)。
func IsWater(tile byte) bool { return tile >= 0x14 && tile <= 0x20 }

// IsTree 對齊 kbres.h:303 `IS_TREE(M) ((M) >= 0x21 && (M) <= 0x2D)`。
func IsTree(tile byte) bool { return tile >= 0x21 && tile <= 0x2D }

// IsDesert 對齊 kbres.h:304 `IS_DESERT(M) ((M) >= 0x2e && (M) <= 0x3a)`。
func IsDesert(tile byte) bool { return tile >= 0x2E && tile <= 0x3A }

// IsRock 對齊 kbres.h:305 `IS_ROCK(M) ((M) >= 0x3b && (M) <= 0x47)`。
func IsRock(tile byte) bool { return tile >= 0x3B && tile <= 0x47 }

// IsDeepWater 對齊 kbres.h:307 `IS_DEEP_WATER(M) ((M) == TILE_DEEP_WATER)`。
func IsDeepWater(tile byte) bool { return tile == TileDeepWater }

// IsInteractive 對齊 kbres.h:309 `IS_INTERACTIVE(M) ((M) & 0x80)`——
// 高位元為 1(0x80..0xFF):城堡/城鎮/箱子/棲地/路標/敵人/神器等可互動圖磚。
// 注意 0x80 本身同時滿足 IS_GRASS(見上),兩者在 C 版並非互斥判斷,
// 呼叫端依序 if/else if 決定優先權(見 src/rogue.c:70-73)。
func IsInteractive(tile byte) bool { return tile&0x80 != 0 }

// WorldMap 是解析後的世界地圖,對齊 src/bounty.h:170 的
// `byte map[MAX_CONTINENTS][LEVEL_H][LEVEL_W]`。
type WorldMap struct {
	Continents int // 洲數,恆為 MaxContinents(來源檔案固定 4 洲)
	// Tiles[cont][y][x],cont/y/x 皆從 0 起。
	// 索引順序對齊 C 版陣列宣告與 src/play.c:380 memcpy 的位元組順序
	// (continent-major → row(y) → column(x)),並經 src/tools/kbtmx.py:495-497
	// 的寫檔迴圈次序交叉驗證。
	Tiles [][LevelH][LevelW]byte
}

// LoadWorldMap 從路徑讀取 land.org(或 LAND.ORG)並解析成 *WorldMap。
func LoadWorldMap(path string) (*WorldMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("kbdata: LoadWorldMap %s: %w", path, err)
	}
	return ParseWorldMap(data)
}

// ParseWorldMap 解析 land.org 的原始位元組內容。
//
// 依 src/lib/free-data.c:1438-1452 / dos-data.c:1652-1666,合法檔案長度
// 恆為 LevelW*LevelH*MaxContinents = 16384 bytes,無 header、無壓縮。
func ParseWorldMap(data []byte) (*WorldMap, error) {
	if len(data) < worldMapBytes {
		return nil, fmt.Errorf("kbdata: land.org 太短 (%d bytes < %d)", len(data), worldMapBytes)
	}

	m := &WorldMap{
		Continents: MaxContinents,
		Tiles:      make([][LevelH][LevelW]byte, MaxContinents),
	}

	off := 0
	for cont := 0; cont < MaxContinents; cont++ {
		for y := 0; y < LevelH; y++ {
			copy(m.Tiles[cont][y][:], data[off:off+LevelW])
			off += LevelW
		}
	}

	return m, nil
}

// Set 寫入指定洲座標 (x, y) 的 tile byte(界外靜默略過,呼叫端不必先做邊界檢查)。
// 對齊 C 版 `game->map[cont][y][x] = v`(如 run_combat 戰勝後把 foe tile 清成 0)。
func (m *WorldMap) Set(cont, x, y int, v byte) {
	if m == nil || cont < 0 || cont >= len(m.Tiles) {
		return
	}
	if y < 0 || y >= LevelH || x < 0 || x >= LevelW {
		return
	}
	m.Tiles[cont][y][x] = v
}

// Tile 回傳指定洲座標 (x, y) 的原始 tile byte。
// 座標或洲別界外一律回傳 TileGrass(0)當安全值,呼叫端不必自行邊界檢查。
func (m *WorldMap) Tile(cont, x, y int) byte {
	if m == nil {
		return TileGrass
	}
	if cont < 0 || cont >= len(m.Tiles) {
		return TileGrass
	}
	if y < 0 || y >= LevelH || x < 0 || x >= LevelW {
		return TileGrass
	}
	return m.Tiles[cont][y][x]
}
