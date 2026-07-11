package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// NavigateContinentScreen 對齊 C navigate_continent(game.c:1972):乘船時選擇前往
// 哪一洲。只列已探索(ContinentFound)的洲;選定後 SailTo(移到目的洲入口)+ 消耗一週。
//
// 觸發:世界地圖乘船(Mount==SAIL)時按 'n'(對齊 C KEY_ACT(NEW_CONTINENT),
// game.c:6768,同樣限定 mount==SAIL 才可用)。
type NavigateContinentScreen struct {
	gs      *gamestate.GameState
	assets  *kbdata.Assets
	options []navOption // 可前往的洲(id + 顯示名)
}

type navOption struct {
	id   int
	name string
}

// NewNavigateContinentScreen 建立換洲選單,只收錄 ContinentFound 為真的洲。
func NewNavigateContinentScreen(gs *gamestate.GameState, a *kbdata.Assets) *NavigateContinentScreen {
	s := &NavigateContinentScreen{gs: gs, assets: a}
	names := gamestate.ContinentNames(a)
	for i := 0; i < kbdata.MaxContinents; i++ {
		if gs != nil && gs.ContinentFound[i] != 0 {
			s.options = append(s.options, navOption{id: i, name: names[i]})
		}
	}
	return s
}

func (s *NavigateContinentScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActLetter:
		// C 用數字鍵 1-4 選洲;這裡把 '1'..'4' 對到 options 的顯示序。
		if a.Rune >= '1' && a.Rune <= '9' {
			idx := int(a.Rune - '1')
			return s.sail(idx)
		}
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

// sail 選第 idx 個(顯示序)可前往的洲:SailTo + 消耗一週,然後回地圖。
func (s *NavigateContinentScreen) sail(idx int) Transition {
	if idx < 0 || idx >= len(s.options) || s.gs == nil {
		return Stay()
	}
	beforeGold := s.gs.Gold
	s.gs.SailTo(s.assets, s.options[idx].id)
	// 對齊 C spend_week(play.c:524):換洲花掉本週剩餘天數推進到下一個週界,觸發一次週結算。
	// seed 取 PassedDays 衍生(增長生物 parity 非關鍵)。
	creature := s.gs.SpendWeek(s.assets, kbrng.NewGlibc(uint32(s.gs.PassedDays()+1)))
	if s.gs.DaysLeft <= 0 {
		return Replace(NewLoseScreen(s.gs, s.assets)) // 時間耗盡
	}
	if creature >= 0 {
		// 用 Replace 換掉本換洲畫面 → 顯示週結算,關閉後回世界地圖。
		return Replace(NewEndOfWeekScreen(s.gs, s.assets, creature, beforeGold))
	}
	return Pop()
}

func (s *NavigateContinentScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 0)
	lines := []string{"前往哪一洲?", ""}
	for i, opt := range s.options {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, opt.name))
	}
	drawBottomFrame(dst, s.assets, lines)
}

func (s *NavigateContinentScreen) Keymap() input.Keymap {
	letters := make([]input.LetterItem, len(s.options))
	for i, opt := range s.options {
		letters[i] = input.LetterItem{Rune: rune('1' + i), Label: opt.name}
	}
	return input.Keymap{Directions: false, Cancel: "離開", Letters: letters}
}
