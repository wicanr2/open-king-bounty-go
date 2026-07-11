package render

import (
	"image/color"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// 觸控按鈕點綴色(對齊遊戲既有金框中世紀風)。
// 底色統一深色半透明,只有邊框點綴色隨用途變化,整體克制一致。
var (
	accentGold  = color.RGBA{224, 192, 64, 225}  // 預設金框(#e0c040 系),呼應遊戲金框
	accentSteel = color.RGBA{140, 160, 205, 225} // 系統鍵(主題/作弊/音樂)用冷調鋼藍區分
	accentGreen = color.RGBA{72, 200, 176, 225}  // 確認 / 是(翠綠偏青,在草地綠上仍可辨,不與草色同化)
	accentRed   = color.RGBA{210, 108, 92, 225}  // 取消 / 否
)

// accentFor 依 Action 種類挑點綴色,讓系統鍵與確認/取消一眼可辨,其餘走預設金。
func accentFor(k input.ActionKind) color.RGBA {
	switch k {
	case input.ActThemeCycle, input.ActCheat, input.ActMusicToggle:
		return accentSteel
	case input.ActConfirm, input.ActYes:
		return accentGreen
	case input.ActCancel, input.ActNo:
		return accentRed
	default:
		return accentGold
	}
}

// DrawTouchControls 把觸控控制畫成半透明疊層。
// 每顆按鈕:深色半透明底 + 上亮下暗漸層 + 內浮雕(左上亮/右下暗,凸起感) + 點綴金框 + 置中白字。
// font 可為 nil(桌面無 atlas 時):只畫按鈕底,不畫字,不 panic。
func DrawTouchControls(dst *ebiten.Image, controls []input.Control, font *kbdata.CJKAtlas) {
	for _, c := range controls {
		if c.Hidden {
			continue // 隱形熱區(如城鎮選單行):只作點擊命中,不畫按鈕,避免蓋住選單文字
		}
		drawButton(dst, c.Rect, accentFor(c.Action.Kind))

		if font == nil || c.Label == "" {
			continue // 無 atlas 或無標籤:只留美化底,別畫字
		}
		// 置中:每字元(ASCII/CJK 皆同)前進 CJKCell=8 邏輯 px,字高視為 8。
		n := utf8.RuneCountInString(c.Label)
		tx := c.Rect.X + (c.Rect.W-n*CJKCell)/2
		ty := c.Rect.Y + (c.Rect.H-8)/2
		DrawText(dst, font, c.Label, tx, ty, color.White) // atlas 已自帶 4 向黑外框,任何底上皆清楚
	}
}

// scaleRGBA 依倍率調 RGB 亮度(clamp 0..255),alpha 保留。用來由點綴色推導亮/暗金屬邊。
func scaleRGBA(c color.RGBA, f float64) color.RGBA {
	clamp := func(v float64) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	return color.RGBA{clamp(float64(c.R) * f), clamp(float64(c.G) * f), clamp(float64(c.B) * f), c.A}
}

// drawButton 畫一顆「中世紀凸起金屬框」風按鈕(pixel art,1px 邏輯線 = 輸出 3px)。
// 分層(由底到面):深色底 → 上亮下暗漸層 → 兩色點綴金屬框(左上亮金/右下暗金) →
// 內浮雕亮/暗邊 → 斜切角。金屬框自身左上亮、右下暗,配合內浮雕共同做出「凸起可按」的 3D 感。
func drawButton(dst *ebiten.Image, r input.Rect, accent color.RGBA) {
	x, y := float32(r.X), float32(r.Y)
	w, h := float32(r.W), float32(r.H)
	if w < 2 || h < 2 {
		vector.DrawFilledRect(dst, x, y, w, h, color.RGBA{16, 14, 26, 200}, false)
		return
	}
	half := float32(int(h) / 2)

	// 1) 深色半透明底(讓地圖透一點,又夠對比讀字)。
	vector.DrawFilledRect(dst, x, y, w, h, color.RGBA{16, 14, 26, 202}, false)

	// 2) 垂直漸層:上半疊淡亮、下半疊淡暗 → 上緣受光的立體感。
	vector.DrawFilledRect(dst, x+1, y+1, w-2, half-1, color.RGBA{255, 255, 240, 26}, false)
	vector.DrawFilledRect(dst, x+1, y+half, w-2, h-half-1, color.RGBA{0, 0, 0, 66}, false)

	// 3) 兩色金屬框(最外 1px):左上亮金受光、右下暗金入影 → 框本身即有厚度。
	fLight := scaleRGBA(accent, 1.30)
	fDark := scaleRGBA(accent, 0.52)
	vector.DrawFilledRect(dst, x, y, w, 1, fLight, false)    // 上邊亮
	vector.DrawFilledRect(dst, x, y, 1, h, fLight, false)    // 左邊亮
	vector.DrawFilledRect(dst, x, y+h-1, w, 1, fDark, false) // 下邊暗
	vector.DrawFilledRect(dst, x+w-1, y, 1, h, fDark, false) // 右邊暗

	// 4) 內浮雕(inset 1px):左上白亮、右下黑暗,強化凸起。
	vector.DrawFilledRect(dst, x+1, y+1, w-2, 1, color.RGBA{255, 255, 255, 60}, false) // 內頂亮
	vector.DrawFilledRect(dst, x+1, y+1, 1, h-2, color.RGBA{255, 255, 255, 42}, false) // 內左亮
	vector.DrawFilledRect(dst, x+1, y+h-2, w-2, 1, color.RGBA{0, 0, 0, 115}, false)    // 內底暗
	vector.DrawFilledRect(dst, x+w-2, y+1, 1, h-2, color.RGBA{0, 0, 0, 115}, false)    // 內右暗

	// 5) 斜切角(pixel art 收邊):把 4 個角壓暗 1px,弱化尖角、更像素味。
	corner := color.RGBA{0, 0, 0, 170}
	vector.DrawFilledRect(dst, x, y, 1, 1, corner, false)
	vector.DrawFilledRect(dst, x+w-1, y, 1, 1, corner, false)
	vector.DrawFilledRect(dst, x, y+h-1, 1, 1, corner, false)
	vector.DrawFilledRect(dst, x+w-1, y+h-1, 1, 1, corner, false)
}
