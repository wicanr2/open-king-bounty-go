// tileset.go -- 世界地圖地形 tileset 載入器。
//
// 格式(旗艦已 RE 確認,對照 C 版 src/lib/kbres.c:389 KB_LoadTileset_ROWS 與
// src/lib/free-data.c:1035 GR_TILEROW):
//   - 地形 tile 圖分兩檔:free/tileseta.png(tile 0–35)、free/tilesetb.png(tile 36–71)。
//   - 兩檔皆為 1728x34 單列 PNG,橫排 36 個 tile,每 tile 48x34。
//   - 世界地圖 tile byte m(0–71)對應:
//     m < 36  → tileseta 第 m 個  (x = m*48,       y = 0, 48x34)
//     m >= 36 → tilesetb 第 m-36 個 (x = (m-36)*48, y = 0, 48x34)
//   - 互動物件(城鎮/城堡/寶箱/棲地等 byte >= 0x80)不在此範圍,那是另外的 sprite。
package render

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// tileW、tileH 是單一地形 tile 的像素尺寸,tilesPerSheet 是每張圖含的 tile 數。
const (
	tileW         = 48
	tileH         = 34
	tilesPerSheet = 36
)

// Tileset 是世界地圖地形 tile 0-71 的來源圖集。
// tileseta 對應 m=0..35,tilesetb 對應 m=36..71(見檔頭 RE 對照)。
type Tileset struct {
	tileseta *ebiten.Image
	tilesetb *ebiten.Image
}

// LoadTileset 從 dir/free/tileseta.png 與 dir/free/tilesetb.png 載入世界地圖地形 tileset。
// 兩檔皆為 palette PNG,標準庫 image/png 可直接解;缺檔或解碼失敗回 error。
func LoadTileset(dir string) (*Tileset, error) {
	a, err := LoadPNGTile(filepath.Join(dir, "free", "tileseta.png"))
	if err != nil {
		return nil, fmt.Errorf("render: LoadTileset tileseta: %w", err)
	}
	b, err := LoadPNGTile(filepath.Join(dir, "free", "tilesetb.png"))
	if err != nil {
		return nil, fmt.Errorf("render: LoadTileset tilesetb: %w", err)
	}
	return &Tileset{tileseta: a, tilesetb: b}, nil
}

// TileImage 回傳地形 tile m(0-71)的 48x34 子圖。
// m 對應到哪張圖、哪個 x 位移見檔頭 RE 對照;m > 71 時安全回傳 tile 0,不 panic。
func (t *Tileset) TileImage(m byte) *ebiten.Image {
	if t == nil {
		return nil
	}
	sheet := t.tileseta
	idx := int(m)
	if idx >= tilesPerSheet {
		if idx > 71 {
			idx = 0 // 越界安全值:回退成 tile 0
		} else {
			sheet = t.tilesetb
			idx -= tilesPerSheet
		}
	}
	if sheet == nil {
		return nil
	}
	x := idx * tileW
	rect := image.Rect(x, 0, x+tileW, tileH)
	return sheet.SubImage(rect).(*ebiten.Image)
}

// DrawTileAt 把地形 tile m 以原尺寸畫到 dst 的 (x, y) 左上角。
func (t *Tileset) DrawTileAt(dst *ebiten.Image, m byte, x, y int) {
	if t == nil || dst == nil {
		return
	}
	tile := t.TileImage(m)
	if tile == nil {
		return
	}
	DrawTile(dst, tile, x, y)
}
