package screen

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// bridgeMode 是 BridgeScreen 的子狀態(選方向 / 顯示結果)。
type bridgeMode int

const (
	bridgeModePick bridgeMode = iota
	bridgeModeResult
)

// BridgeScreen 對齊 C build_bridge(game.c:3873):問「往哪個方向造橋」,玩家用方向鍵
// 選一個方向,在該方向接下來 1-2 格水域上架橋;無水可架則「白白浪費一個好法術」。
// 由 ChooseSpellScreen 施放造橋術後 Replace 進來(法術已消耗),完成後 Pop 回世界地圖。
type BridgeScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	mode   bridgeMode
	msg    []string
}

// NewBridgeScreen 建立造橋方向選擇畫面。
func NewBridgeScreen(gs *gamestate.GameState, a *kbdata.Assets) *BridgeScreen {
	return &BridgeScreen{gs: gs, assets: a}
}

func (s *BridgeScreen) Update(a input.Action) Transition {
	if s.mode == bridgeModeResult {
		if !a.IsNone() {
			return Pop()
		}
		return Stay()
	}
	dirX, dirY := 0, 0
	switch a.Kind {
	case input.ActUp:
		dirY = -1 // 對齊 C dj=-1(上);BuildBridge 內 y = Y - i*dirY → 上=y 增加
	case input.ActDown:
		dirY = 1
	case input.ActLeft:
		dirX = -1
	case input.ActRight:
		dirX = 1
	case input.ActCancel:
		return Pop() // 取消造橋(法術已消耗,對齊 C key==0xFF return)
	default:
		return Stay()
	}
	built := s.gs.BuildBridge(dirX, dirY)
	s.mode = bridgeModeResult
	if built == 0 {
		s.msg = []string{"這裡不適合造橋", "白白浪費了一個好法術!"}
	} else {
		s.msg = []string{"橋已架起。"}
	}
	return Stay()
}

func (s *BridgeScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	if s.mode == bridgeModeResult {
		drawTopBox(dst, s.assets, "按 'ESC' 離開")
		drawSidebar(dst, s.gs, 0)
		drawBottomFrame(dst, s.assets, s.msg)
		return
	}
	drawTopBox(dst, s.assets, "往哪個方向造橋")
	drawSidebar(dst, s.gs, 0)
	drawBottomFrame(dst, s.assets, []string{"用方向鍵選擇造橋方向。"})
}

func (s *BridgeScreen) Keymap() input.Keymap {
	if s.mode == bridgeModeResult {
		return input.Keymap{Directions: false, Confirm: "繼續", Cancel: "離開"}
	}
	return input.Keymap{Directions: true, Cancel: "取消"}
}
