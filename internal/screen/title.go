package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// TitleScreen 是開場畫面:顯示標題,按確認進選角。
type TitleScreen struct {
	assets *kbdata.Assets
}

func NewTitleScreen(a *kbdata.Assets) *TitleScreen { return &TitleScreen{assets: a} }

func (s *TitleScreen) Update(a input.Action) Transition {
	if a.Kind == input.ActConfirm {
		return Replace(NewCharSelectScreen(s.assets))
	}
	return Stay()
}

func (s *TitleScreen) Draw(dst *ebiten.Image) {
	if s.assets != nil && s.assets.Font != nil {
		render.DrawText(dst, s.assets.Font, "御封戰將", 112, 70, color.White)
	}
	ebitenutil.DebugPrintAt(dst, "King's Bounty (Go/Ebiten)", 100, 110)
	ebitenutil.DebugPrintAt(dst, "Press ENTER / SPACE", 110, 150)
}

func (s *TitleScreen) Keymap() input.Keymap {
	return input.Keymap{Confirm: "開始"}
}
