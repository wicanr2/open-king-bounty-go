package render

import (
	"bytes"
	"image"
	"image/color"
	"io/fs"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// Sprite 是一張橫向 sprite sheet(N 幀並排,每幀 FrameW×FrameH)。
// 對應 C 版的 IMG_ROW 資源(cursor.png = 12 幀 48×34:船 0-3 / 飛龍 4-7 / 騎馬主角 8-11)。
type Sprite struct {
	sheet          *ebiten.Image
	FrameW, FrameH int
}

// colorKeyTopLeft 把 src 中「與左上角 (0,0) 同色」的像素設為透明,回傳 ebiten.Image。
// 對應 C 版 free-data.c:1579 的 SDL_SetColorKey(surf, SRCCOLORKEY, 0xFF):
// cursor.png 的調色盤 index 255 = (128,0,128) 暗洋紅背景,遊戲中須去背。
func colorKeyTopLeft(src image.Image) *ebiten.Image {
	b := src.Bounds()
	kr, kg, kb, _ := src.At(b.Min.X, b.Min.Y).RGBA()
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := src.At(x, y).RGBA()
			if r == kr && g == kg && bl == kb {
				continue // 與 colorkey 同色 → 留透明(RGBA 預設 0)
			}
			rgba.Set(x-b.Min.X, y-b.Min.Y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), uint8(a >> 8)})
		}
	}
	return ebiten.NewImageFromImage(rgba)
}

// LoadSpriteFS 從檔案系統讀 sprite sheet(去背),需在遊戲迴圈啟動後呼叫(會建 ebiten.Image)。
func LoadSpriteFS(fsys fs.FS, name string, frameW, frameH int) (*Sprite, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return &Sprite{sheet: colorKeyTopLeft(img), FrameW: frameW, FrameH: frameH}, nil
}

// LoadSprite 從 dir 下的 free/name 讀 sprite sheet(去背,桌面用)。
func LoadSprite(dir, name string, frameW, frameH int) (*Sprite, error) {
	data, err := os.ReadFile(dir + "/free/" + name)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return &Sprite{sheet: colorKeyTopLeft(img), FrameW: frameW, FrameH: frameH}, nil
}

// Frame 回傳第 n 幀的子圖(越界回 nil)。
func (s *Sprite) Frame(n int) *ebiten.Image {
	if s == nil || s.sheet == nil {
		return nil
	}
	x := n * s.FrameW
	if x < 0 || x+s.FrameW > s.sheet.Bounds().Dx() {
		return nil
	}
	return s.sheet.SubImage(image.Rect(x, 0, x+s.FrameW, s.FrameH)).(*ebiten.Image)
}

// DrawFrame 把第 n 幀畫到 dst 的 (x,y)。
func (s *Sprite) DrawFrame(dst *ebiten.Image, n, x, y int) {
	f := s.Frame(n)
	if f == nil || dst == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(f, op)
}
