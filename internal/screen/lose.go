package screen

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// LoseScreen 對齊 C lose_game(game.c:4601):剩餘天數歸零(時間耗盡)時的遊戲結束畫面
// (簡化版:C 版放 GR_ENDING 結局圖 + 長文,那些美術屬版權素材不入 free 包,這裡以文字
// 呈現失敗訊息)。任意鍵回標題畫面。
type LoseScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
}

// NewLoseScreen 建立時間耗盡的遊戲結束畫面。
func NewLoseScreen(gs *gamestate.GameState, a *kbdata.Assets) *LoseScreen {
	return &LoseScreen{gs: gs, assets: a}
}

func (s *LoseScreen) Update(a input.Action) Transition {
	if !a.IsNone() {
		return Replace(NewTitleScreen(s.assets)) // 遊戲結束後回標題
	}
	return Stay()
}

func (s *LoseScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	drawTopBox(dst, s.assets, "時間耗盡")
	drawBottomFrame(dst, s.assets, []string{
		"你未能在期限內",
		"尋回秩序權杖。",
		"",
		"馬克馬斯國王的國度",
		"就此沒入黑暗,",
		"你的使命以失敗告終。",
		"",
		"(按任意鍵回到標題)",
	})
}

func (s *LoseScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: false, Confirm: "回標題"}
}
