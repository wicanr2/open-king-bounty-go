package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// gateMode 區分城堡傳送 / 鄉鎮傳送(對齊 C select_gate 的 mode 0/1)。
type gateMode int

const (
	gateCastle gateMode = iota
	gateTown
)

// gateEntry 是一個可傳送目的地(已造訪的城堡/鄉鎮)。
type gateEntry struct {
	id   int
	name string
}

// SelectGateScreen 對齊 C select_gate(game.c:4146):列出玩家造訪過的城堡(SPELL_CASTLEGATE)
// 或鄉鎮(SPELL_TOWNGATE),選一個瞬移過去(乘船中會先下船)。由 ChooseSpellScreen 施放
// 對應法術後 Replace 進來;選定/取消後 Pop 回世界地圖(法術已在施放時消耗)。
type SelectGateScreen struct {
	gs      *gamestate.GameState
	assets  *kbdata.Assets
	mode    gateMode
	entries []gateEntry // 已造訪的目的地(id 為城堡/鎮絕對索引,字母 = 'A'+id)
}

// NewSelectGateScreen 建立傳送目的地選單,只收錄已造訪的城堡/鄉鎮。
func NewSelectGateScreen(gs *gamestate.GameState, a *kbdata.Assets, mode gateMode) *SelectGateScreen {
	s := &SelectGateScreen{gs: gs, assets: a, mode: mode}
	if mode == gateCastle {
		castles := gamestate.LoadCastles(a)
		for i := range castles {
			if i < len(gs.CastleVisited) && gs.CastleVisited[i] != 0 {
				s.entries = append(s.entries, gateEntry{id: i, name: castles[i].Name})
			}
		}
	} else {
		towns := gamestate.LoadTowns(a)
		for i := range towns {
			if i < len(gs.TownVisited) && gs.TownVisited[i] != 0 {
				s.entries = append(s.entries, gateEntry{id: i, name: towns[i].Name})
			}
		}
	}
	return s
}

func (s *SelectGateScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActLetter:
		// 字母 = 目的地絕對索引(A='id' 0);只有已造訪的才傳送(GateTo* 內會再驗)。
		if a.Rune >= 'a' && a.Rune <= 'z' {
			id := int(a.Rune - 'a')
			ok := false
			if s.mode == gateCastle {
				ok = s.gs.GateToCastle(s.assets, id)
			} else {
				ok = s.gs.GateToTown(s.assets, id)
			}
			if ok {
				return Pop() // 傳送完成,回世界地圖(gs 座標已改,worldmap 即時讀到)
			}
		}
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

func (s *SelectGateScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 0)

	header := "你造訪過的城堡:"
	prompt := "重訪哪座城堡?"
	if s.mode == gateTown {
		header = "你造訪過的鄉鎮:"
		prompt = "重訪哪個鄉鎮?"
	}
	lines := []string{header, ""}
	if len(s.entries) == 0 {
		lines = append(lines, "(尚無可傳送的目的地)")
	}
	for _, e := range s.entries {
		lines = append(lines, fmt.Sprintf("%c) %s", 'A'+e.id, e.name))
	}
	lines = append(lines, "", prompt)
	drawBottomFrame(dst, s.assets, lines)
}

func (s *SelectGateScreen) Keymap() input.Keymap {
	letters := make([]input.LetterItem, len(s.entries))
	for i, e := range s.entries {
		letters[i] = input.LetterItem{Rune: rune('a' + e.id), Label: e.name}
	}
	return input.Keymap{Directions: false, Cancel: "離開", Letters: letters}
}
