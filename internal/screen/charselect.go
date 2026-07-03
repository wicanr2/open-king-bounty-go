package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// classNames 是四職業的起手頭銜(rank 0),索引 = class。
var classNames = []string{"騎士", "遊俠", "女巫師", "蠻俠"}

// CharSelectScreen 是選角畫面:A–D 選職業 → 建角 → 進遊戲。
type CharSelectScreen struct {
	assets *kbdata.Assets
	sel    int // 目前高亮的職業(方向鍵移動)
}

func NewCharSelectScreen(a *kbdata.Assets) *CharSelectScreen {
	return &CharSelectScreen{assets: a}
}

func (s *CharSelectScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActUp:
		if s.sel > 0 {
			s.sel--
		}
	case input.ActDown:
		if s.sel < len(classNames)-1 {
			s.sel++
		}
	case input.ActLetter:
		// a–d 直接選對應職業
		if a.Rune >= 'a' && a.Rune <= 'd' {
			return s.start(int(a.Rune - 'a'))
		}
	case input.ActConfirm:
		return s.start(s.sel)
	case input.ActCancel:
		return Replace(NewTitleScreen(s.assets))
	}
	return Stay()
}

// start 用選定職業建角,進世界地圖。
func (s *CharSelectScreen) start(class int) Transition {
	gs := gamestate.NewGame(s.assets, "Sir Loin", class)
	return Replace(NewWorldMapScreen(gs, s.assets))
}

func (s *CharSelectScreen) Draw(dst *ebiten.Image) {
	ebitenutil.DebugPrintAt(dst, "SELECT A CLASS (A-D or arrows+ENTER)", 20, 16)
	for i, name := range classNames {
		y := 50 + i*32
		var clr color.Color = color.Gray{Y: 160}
		marker := "  "
		if i == s.sel {
			clr = color.White
			marker = "> "
		}
		ebitenutil.DebugPrintAt(dst, marker, 30, y)
		if s.assets != nil && s.assets.Font != nil {
			render.DrawText(dst, s.assets.Font, name, 50, y-6, clr)
		}
	}
}

func (s *CharSelectScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: true,
		Confirm:    "選擇",
		Cancel:     "返回",
		Letters: []input.LetterItem{
			{Rune: 'a', Label: "騎士"}, {Rune: 'b', Label: "遊俠"},
			{Rune: 'c', Label: "女巫師"}, {Rune: 'd', Label: "蠻俠"},
		},
	}
}
