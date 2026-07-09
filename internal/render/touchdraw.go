package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// DrawTouchControls 把觸控控制(由 input.TouchLayout.Controls 產生)畫成半透明疊層。
// 每顆:半透明底 + 亮框 + 標籤(CJK 走 atlas,ASCII 走內建字;兩者都畫,各顯其字)。
// font 可為 nil(只畫框 + ASCII 標籤)。
func DrawTouchControls(dst *ebiten.Image, controls []input.Control, font *kbdata.CJKAtlas) {
	for _, c := range controls {
		if c.Hidden {
			continue // 隱形熱區(如城鎮選單行):只作點擊命中,不畫按鈕,避免蓋住選單文字
		}
		x, y := float32(c.Rect.X), float32(c.Rect.Y)
		w, h := float32(c.Rect.W), float32(c.Rect.H)
		vector.DrawFilledRect(dst, x, y, w, h, color.RGBA{20, 20, 30, 150}, false) // 半透明底
		vector.StrokeRect(dst, x, y, w, h, 1, color.RGBA{200, 200, 220, 200}, false)

		// 標籤:兩種都畫一次,ASCII 由內建字顯示、CJK 由 atlas 顯示(各只渲染自己支援的字元)。
		ebitenutil.DebugPrintAt(dst, c.Label, c.Rect.X+3, c.Rect.Y+c.Rect.H/2-8)
		if font != nil {
			DrawText(dst, font, c.Label, c.Rect.X+2, c.Rect.Y+1, color.White)
		}
	}
}
