// tile.go -- free 模組 tile PNG 載入與繪製。
//
// free tile 是普通 PNG(常見 8-bit colormap/palette,如 data/free/cstl.png),
// 用 Go 標準 image/png 就能解,不需要照搬 C 版任何自製解碼器。
package render

import (
	"fmt"
	"image"
	_ "image/png" // 註冊 PNG decoder 給 image.Decode 用
	"io/fs"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// LoadPNGTile 讀取一張 free tile PNG 並轉成 *ebiten.Image。
func LoadPNGTile(path string) (*ebiten.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("render: LoadPNGTile open %s: %w", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("render: LoadPNGTile decode %s: %w", path, err)
	}
	return ebiten.NewImageFromImage(img), nil
}

// LoadPNGTileNamed 是 LoadPNGTile 的便利版,自動組 dir/free/name(比照 LoadSprite 慣例)。
func LoadPNGTileNamed(dir, name string) (*ebiten.Image, error) {
	return LoadPNGTile(dir + "/free/" + name)
}

// LoadPNGTileFS 是 LoadPNGTile 的 embed.FS 版本(行動裝置用),name 需含 "free/" 前綴。
func LoadPNGTileFS(fsys fs.FS, name string) (*ebiten.Image, error) {
	return loadPNGFromFS(fsys, name)
}

// DrawTile 把 tile 原尺寸畫到 dst 的 (x, y) 左上角,不縮放。
func DrawTile(dst, tile *ebiten.Image, x, y int) {
	if dst == nil || tile == nil {
		return
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(tile, &op)
}
