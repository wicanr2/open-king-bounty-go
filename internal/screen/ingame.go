package screen

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// InGameScreen 是進遊戲後的暫代畫面:顯示建好的角色狀態(隊伍/金/領導力)。
// 之後由真正的世界地圖畫面取代(需 land.org 地圖載入)。
type InGameScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
}

func NewInGameScreen(gs *gamestate.GameState, a *kbdata.Assets) *InGameScreen {
	return &InGameScreen{gs: gs, assets: a}
}

func (s *InGameScreen) Update(a input.Action) Transition {
	if a.Kind == input.ActCancel {
		return Replace(NewTitleScreen(s.assets))
	}
	return Stay()
}

func (s *InGameScreen) Draw(dst *ebiten.Image) {
	gs := s.gs
	title := classNames[gs.Class]
	if s.assets != nil && gs.Class < len(s.assets.Classes) {
		title = s.assets.Classes[gs.Class][gs.Rank].Name
	}

	drawCJK := func(str string, x, y int, clr color.Color) {
		if s.assets != nil && s.assets.Font != nil {
			render.DrawText(dst, s.assets.Font, str, x, y, clr)
		}
	}

	drawCJK(title, 8, 8, color.White)
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%s the class#%d", gs.Name, gs.Class), 8, 34)
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("GOLD %d   LEADERSHIP %d   COMMISSION %d",
		gs.Gold, gs.Leadership, gs.Commission), 8, 48)

	drawCJK("隊伍", 8, 70, color.White)
	y := 96
	for _, sq := range gs.Army {
		if sq.Count == 0 || sq.TroopID == 255 {
			continue
		}
		name := "?"
		if s.assets != nil && sq.TroopID < len(s.assets.Troops) {
			name = s.assets.Troops[sq.TroopID].Name
		}
		drawCJK(name, 16, y, color.White)
		ebitenutil.DebugPrintAt(dst, fmt.Sprintf("x %d", sq.Count), 80, y+6)
		y += 28
	}

	ebitenutil.DebugPrintAt(dst, "ESC: title", 8, 182)
}

func (s *InGameScreen) Keymap() input.Keymap {
	return input.Keymap{Cancel: "返回標題"}
}
