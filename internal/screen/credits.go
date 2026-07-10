package screen

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// CreditsScreen 對齊 C show_credits(game.c:706):顯示製作名單(credits.txt 原文)。
// 任意鍵回標題。由標題畫面按 'c' 開啟。
type CreditsScreen struct {
	assets *kbdata.Assets
	lines  []string
}

// NewCreditsScreen 建立製作名單畫面。
func NewCreditsScreen(a *kbdata.Assets) *CreditsScreen {
	s := &CreditsScreen{assets: a}
	text := ""
	if a != nil {
		text = a.Credits
	}
	if text == "" {
		text = "openkb-go"
	}
	s.lines = strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	return s
}

func (s *CreditsScreen) Update(a input.Action) Transition {
	if !a.IsNone() {
		return Pop()
	}
	return Stay()
}

func (s *CreditsScreen) Draw(dst *ebiten.Image) {
	dst.Fill(color.Black)
	if s.assets == nil || s.assets.Font == nil {
		return
	}
	for i, ln := range s.lines {
		render.DrawText(dst, s.assets.Font, ln, 24, 12+i*render.CJKCell, color.White)
	}
}

func (s *CreditsScreen) Keymap() input.Keymap {
	return input.Keymap{Confirm: "返回", Cancel: "返回"}
}
