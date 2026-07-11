package render

import (
	"image/color"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// 觸控按鈕視覺 = 對齊遊戲既有 UI 語言:C 版狀態列/底部選單是「深藍底 + 金色框」,
// 觸控鈕沿用同一套 —— 半透明深藍底 + 統一亮金框(對齊金框 frame.png 的 #e6cf00)+
// 置中白字。刻意「平」不做假 3D 浮雕(先前白色斜角浮雕像 Win95 鈕,與像素金框不搭)。
// 只有確認/取消用微調底色(綠/紅)輔助辨識,金框保持一致。
var (
	fillDefault = color.RGBA{0x00, 0x00, 0x8c, 210}  // 預設:遊戲藍(略亮於 #0000aa 好在地圖上分離)
	fillConfirm = color.RGBA{0x18, 0x62, 0x30, 210}  // 確認/是:深綠(金框仍一致)
	fillCancel  = color.RGBA{0x86, 0x22, 0x22, 210}  // 取消/否:深紅
	borderGold  = color.RGBA{0xe6, 0xcf, 0x00, 0xff} // 金框:對齊 frame.png 主金 #e6cf00
	borderDark  = color.RGBA{0x00, 0x00, 0x00, 0xcc} // 金框外一圈暗線,在亮地圖上定義邊界
)

// fillFor 依 Action 種類挑底色(確認綠/取消紅/其餘藍);金框一律相同,維持 C 版金框一致性。
func fillFor(k input.ActionKind) color.RGBA {
	switch k {
	case input.ActConfirm, input.ActYes:
		return fillConfirm
	case input.ActCancel, input.ActNo:
		return fillCancel
	default:
		return fillDefault
	}
}

// DrawTouchControls 把觸控控制畫成半透明疊層:深藍底 + 亮金框 + 置中白字。
// font 可為 nil(桌面無 atlas 時):只畫按鈕底,不畫字,不 panic。
func DrawTouchControls(dst *ebiten.Image, controls []input.Control, font *kbdata.CJKAtlas) {
	for _, c := range controls {
		if c.Hidden {
			continue // 隱形熱區(如城鎮選單行):只作點擊命中,不畫按鈕,避免蓋住選單文字
		}
		drawButton(dst, c.Rect, fillFor(c.Action.Kind))

		if font == nil || c.Label == "" {
			continue // 無 atlas 或無標籤:只留底,別畫字
		}
		// 置中:每字元(ASCII/CJK 皆同)前進 CJKCell=8 邏輯 px,字高視為 8。
		n := utf8.RuneCountInString(c.Label)
		tx := c.Rect.X + (c.Rect.W-n*CJKCell)/2
		ty := c.Rect.Y + (c.Rect.H-8)/2
		DrawText(dst, font, c.Label, tx, ty, color.White) // atlas 已自帶 4 向黑外框,任何底上皆清楚
	}
}

// drawButton 畫一顆「深藍底 + 亮金框」按鈕(pixel art,1px 邏輯 = 輸出 3px)。
// 平面、乾淨、四邊一致——外 1px 暗線定義邊界、內 1px 亮金框呼應遊戲金框,不做白色假浮雕。
func drawButton(dst *ebiten.Image, r input.Rect, fill color.RGBA) {
	x, y := float32(r.X), float32(r.Y)
	w, h := float32(r.W), float32(r.H)
	if w < 2 || h < 2 {
		vector.DrawFilledRect(dst, x, y, w, h, fill, false)
		return
	}
	// 1) 底(半透明深藍,讓地圖透一點又夠對比)
	vector.DrawFilledRect(dst, x, y, w, h, fill, false)
	// 2) 最外 1px 暗線:在亮地圖上把按鈕邊界壓出來
	strokeRect1px(dst, x, y, w, h, borderDark)
	// 3) 內 1px 亮金框:對齊遊戲金框,四邊一致(無左上亮/右下暗的假 3D)
	if w >= 4 && h >= 4 {
		strokeRect1px(dst, x+1, y+1, w-2, h-2, borderGold)
	}
}

// strokeRect1px 用 4 條 1px filled rect 畫出脆利的 1px 邊框(vector.StrokeRect 會落在
// 半像素、放大後糊掉;pixel art 用 filled rect 對齊整數格才銳利)。
func strokeRect1px(dst *ebiten.Image, x, y, w, h float32, c color.RGBA) {
	vector.DrawFilledRect(dst, x, y, w, 1, c, false)     // 上
	vector.DrawFilledRect(dst, x, y+h-1, w, 1, c, false) // 下
	vector.DrawFilledRect(dst, x, y, 1, h, c, false)     // 左
	vector.DrawFilledRect(dst, x+w-1, y, 1, h, c, false) // 右
}
