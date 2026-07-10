package screen

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// WinScreen 對齊 C win_game(game.c:4537)的破關畫面(簡化版:C 版放結局動畫 +
// GR_ENDING 圖 + 長篇文字,那些美術屬版權素材不入 free 包,這裡以文字呈現破關訊息)。
// 任意鍵回標題畫面。
type WinScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
}

// NewWinScreen 建立破關畫面。
func NewWinScreen(gs *gamestate.GameState, a *kbdata.Assets) *WinScreen {
	return &WinScreen{gs: gs, assets: a}
}

func (s *WinScreen) Update(a input.Action) Transition {
	if !a.IsNone() {
		return Replace(NewTitleScreen(s.assets)) // 破關後回標題
	}
	return Stay()
}

func (s *WinScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	drawTopBox(dst, s.assets, "破關!")
	drawBottomFrame(dst, s.assets, []string{
		"你尋回了失竊的",
		"秩序權杖!",
		"",
		"馬克馬斯國王的國度",
		"重歸太平,你的英名",
		"永載史冊。",
		"",
		"(按任意鍵回到標題)",
	})
}

func (s *WinScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: false, Confirm: "回標題"}
}
