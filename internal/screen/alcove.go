package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// alcoveMode 是 AlcoveScreen 的子狀態。
type alcoveMode int

const (
	alcoveAsk alcoveMode = iota
	alcoveResult
)

// AlcoveScreen 對齊 C visit_alcove(game.c):魔法密室的大法師 Aurange 以 5000 金幣傳授
// 施法奧秘。接受且金幣足 → 習得魔法(KnowsMagic);金幣不足則提示。由世界地圖踩到魔法
// 密室座標(land.ini special1)時開啟。
type AlcoveScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	mode   alcoveMode
	msg    []string
}

// NewAlcoveScreen 建立魔法密室畫面。
func NewAlcoveScreen(gs *gamestate.GameState, a *kbdata.Assets) *AlcoveScreen {
	return &AlcoveScreen{gs: gs, assets: a}
}

func (s *AlcoveScreen) Update(a input.Action) Transition {
	if s.mode == alcoveResult {
		if !a.IsNone() {
			return Pop()
		}
		return Stay()
	}
	confirm := a.Kind == input.ActConfirm || (a.Kind == input.ActLetter && a.Rune == 'y')
	if confirm {
		if s.gs.LearnMagicAtAlcove() {
			s.msg = []string{"你已習得施法的奧秘。", "現在你會使用魔法了。"}
		} else {
			s.msg = []string{"你的金幣不足。"}
		}
		s.mode = alcoveResult
		return Stay()
	}
	if a.Kind == input.ActCancel || (a.Kind == input.ActLetter && a.Rune == 'n') {
		return Pop()
	}
	return Stay()
}

func (s *AlcoveScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 0)
	if s.mode == alcoveResult {
		drawBottomFrame(dst, s.assets, s.msg)
		return
	}
	drawBottomFrame(dst, s.assets, []string{
		"德高望重的大法師 Aurange,",
		fmt.Sprintf("將以 %d 金幣", gamestate.CostAlcove),
		"傳授你施法的奧秘。",
		"",
		"接受 (y/n)?",
	})
}

func (s *AlcoveScreen) Keymap() input.Keymap {
	if s.mode == alcoveResult {
		return input.Keymap{Directions: false, Confirm: "繼續", Cancel: "離開"}
	}
	return input.Keymap{
		Directions: false, Confirm: "接受", Cancel: "拒絕",
		Letters: []input.LetterItem{{Rune: 'y', Label: "接受"}, {Rune: 'n', Label: "拒絕"}},
	}
}
