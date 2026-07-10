package screen

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// ArtifactScreen 對齊 C take_artifact 的顯示層(game.c:3360):拾得神器時顯示其描述
// 文字 + 「……以及一片通往失竊權杖的地圖。」。任意鍵回世界地圖。神器效果與 tile 清除
// 已在 gamestate.TakeArtifact 完成,本畫面只負責顯示。
type ArtifactScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	lines  []string
}

// NewArtifactScreen 建立神器拾得畫面。pickup 由呼叫端先呼叫 gs.TakeArtifact 取得。
func NewArtifactScreen(gs *gamestate.GameState, a *kbdata.Assets, pickup gamestate.ArtifactPickup) *ArtifactScreen {
	s := &ArtifactScreen{gs: gs, assets: a}
	// artifacts.ini 的 description 用字面 "\n" 記換行,展開成實際換行再逐行顯示。
	desc := strings.ReplaceAll(pickup.Desc, `\n`, "\n")
	for _, ln := range strings.Split(desc, "\n") {
		s.lines = append(s.lines, ln)
	}
	s.lines = append(s.lines, "", "……以及一片通往", "失竊權杖的地圖。")
	return s
}

func (s *ArtifactScreen) Update(a input.Action) Transition {
	if !a.IsNone() {
		return Pop()
	}
	return Stay()
}

func (s *ArtifactScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 1) // C: draw_sidebar(game, 1)——拾得神器時用 tick 1
	drawBottomFrame(dst, s.assets, s.lines)
}

func (s *ArtifactScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: false, Confirm: "繼續", Cancel: "離開"}
}
