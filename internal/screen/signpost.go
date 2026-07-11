package screen

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// SignpostScreen 對齊 C read_signpost(game.c:3095):顯示「告示寫著:」+ 該路標文字
// (依路標在地圖掃描序的索引查 a.Signs)。任意鍵回地圖。
type SignpostScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	lines  []string
}

// NewSignpostScreen 建立路標畫面,建構時算好要顯示的文字行。
func NewSignpostScreen(gs *gamestate.GameState, a *kbdata.Assets) *SignpostScreen {
	s := &SignpostScreen{gs: gs, assets: a, lines: []string{"告示寫著:"}}
	idx := gs.SignpostIndex()
	text := ""
	if a != nil && idx >= 0 && idx < len(a.Signs) {
		text = a.Signs[idx]
	}
	s.lines = append(s.lines, "")
	for _, ln := range strings.Split(text, "\n") {
		s.lines = append(s.lines, ln)
	}
	return s
}

func (s *SignpostScreen) Update(a input.Action) Transition {
	if !a.IsNone() {
		return Pop()
	}
	return Stay()
}

func (s *SignpostScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 0)
	drawBottomFrame(dst, s.assets, s.lines)
}

func (s *SignpostScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: false, Confirm: "繼續", Cancel: "離開"}
}
