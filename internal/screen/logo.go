package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// LogoScreen 對齊 C display_logo(game.c:924):開場 NWC 商標(nwcp.png)黑底置中,
// 按任意鍵進標題畫面。這是遊戲啟動的第一個畫面(display_logo → display_title)。
type LogoScreen struct {
	assets *kbdata.Assets
}

// NewLogoScreen 建立開場商標畫面。
func NewLogoScreen(a *kbdata.Assets) *LogoScreen { return &LogoScreen{assets: a} }

func (s *LogoScreen) Update(a input.Action) Transition {
	if !a.IsNone() {
		return Replace(NewTitleScreen(s.assets)) // 任意鍵 → 標題
	}
	return Stay()
}

func (s *LogoScreen) Draw(dst *ebiten.Image) {
	dst.Fill(color.Black) // 對齊 C:黑底
	if logoArt == nil {
		return
	}
	// 置中(對齊 C RECT_Center):logoArt 未必是全螢幕圖。
	b := logoArt.Bounds()
	x := (app_LogicalW - b.Dx()) / 2
	y := (app_LogicalH - b.Dy()) / 2
	render.DrawTile(dst, logoArt, x, y)
}

func (s *LogoScreen) Keymap() input.Keymap {
	return input.Keymap{Confirm: "繼續"}
}

// app_LogicalW/H 是邏輯畫面尺寸(320×200),與 app.LogicalW/H 一致。screen 套件不 import
// app(會循環),故在此以常數複述(對齊 app.LogicalW/LogicalH)。
const (
	app_LogicalW = 320
	app_LogicalH = 200
)
