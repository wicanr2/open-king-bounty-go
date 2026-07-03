package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// TitleScreen 是開場標題(對齊 C display_title:黑底貼滿 title.png「Royal Reward」,
// 按任意鍵進選角)。title.png 為 320×200 全螢幕圖。
type TitleScreen struct {
	assets *kbdata.Assets
}

func NewTitleScreen(a *kbdata.Assets) *TitleScreen { return &TitleScreen{assets: a} }

func (s *TitleScreen) Update(a input.Action) Transition {
	// 任意確認鍵/點擊(對齊 C press_any_key)→ 進選角。
	if a.Kind == input.ActConfirm {
		return Replace(NewCharSelectScreen(s.assets))
	}
	return Stay()
}

func (s *TitleScreen) Draw(dst *ebiten.Image) {
	dst.Fill(color.Black) // 對齊 C:黑底
	if titleArt != nil {
		render.DrawTile(dst, titleArt, 0, 0) // 320×200 全螢幕
		return
	}
	// 無素材時的退回字樣(不影響正常流程)。
	if s.assets != nil && s.assets.Font != nil {
		render.DrawText(dst, s.assets.Font, "御封戰將", 132, 90, color.White)
	}
}

func (s *TitleScreen) Keymap() input.Keymap {
	return input.Keymap{Confirm: "開始"}
}
