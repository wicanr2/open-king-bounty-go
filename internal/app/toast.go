package app

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// toast 是跨畫面的短暫提示(切美術主題 / 切音樂時給觸控使用者視覺回饋——這些是全域
// 快捷,原本靜默切換,玩家(尤其手機上沒鍵盤標籤)看不出切到哪套主題 / 音樂開關)。
// 置中偏上、深藍底 + 金框 + 白字,顯示約 toastTicksDefault 幀後自動消失。
const toastTicksDefault = 110 // ~1.8s @ 60fps

// DebugToast 非空時,遊戲第一幀即顯示此 toast(供 -toast 旗標截圖驗證 toast 外觀)。
var DebugToast string

var (
	toastBg     = color.RGBA{0x00, 0x02, 0x66, 0xe6} // 深藍(對齊 UI dialog 底色),半透明
	toastBorder = color.RGBA{0xe6, 0xcf, 0x00, 0xff} // 鎏金框(對齊觸控按鈕金框色)
)

// showToast 設定要顯示的提示文字並重置計時。
func (g *Game) showToast(msg string) {
	g.toastMsg = msg
	g.toastTicks = toastTicksDefault
}

// drawToast 把提示畫在美術層(置中偏上),文字走 render.DrawText(CJK 於輸出層疊上)。
// 在 Draw 的美術層階段呼叫(觸控疊層之後、放大之前),故盒子疊在最上、字再疊盒子上。
func (g *Game) drawToast(font *kbdata.CJKAtlas) {
	if g.toastTicks <= 0 || g.toastMsg == "" {
		return
	}
	g.toastTicks--

	w := len([]rune(g.toastMsg)) * render.CJKCell
	x := (LogicalW - w) / 2
	y := 18
	bx, by := float32(x-6), float32(y-3)
	bw, bh := float32(w+12), float32(render.CJKCell+6)
	vector.DrawFilledRect(g.art, bx, by, bw, bh, toastBg, false)
	vector.StrokeRect(g.art, bx, by, bw, bh, 1, toastBorder, false)
	render.DrawText(g.art, font, g.toastMsg, x, y, color.White)
}

// themeLabel 把美術模組目錄名轉成給玩家看的友善主題名(toast 用)。
func themeLabel(mod string) string {
	switch mod {
	case "dos":
		return "DOS 經典"
	case "genesis":
		return "Genesis"
	case "amiga":
		return "Amiga"
	case "free":
		return "開放美術"
	default:
		return mod
	}
}
