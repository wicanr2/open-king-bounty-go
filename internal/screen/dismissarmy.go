package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// dismissMode 是 DismissArmyScreen 的子狀態。
type dismissMode int

const (
	dismissModeList        dismissMode = iota // 列出部隊供選
	dismissModeConfirmLast                    // 選到唯一一支部隊 → 先問 y/n
)

// DismissArmyScreen 對齊 C dismiss_army(game.c:2027):解散指定部隊。若只剩一支部隊
// 且要解散它,先確認(「若解散最後一支部隊,你將蒙羞被遣送回國王身邊。」),確認後
// 隊伍清空 → temp_death(對齊 C 呼叫端 game.c:6788 的 test_defeat→temp_death)。
type DismissArmyScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	mode   dismissMode
	pend   int // dismissModeConfirmLast 時待確認解散的 slot
}

// NewDismissArmyScreen 建立解散部隊畫面。
func NewDismissArmyScreen(gs *gamestate.GameState, a *kbdata.Assets) *DismissArmyScreen {
	return &DismissArmyScreen{gs: gs, assets: a}
}

// armyCount 回傳目前非空的部隊格數(對齊 C dismiss_army 的 max)。
func (s *DismissArmyScreen) armyCount() int {
	n := 0
	if s.gs == nil {
		return 0
	}
	for i := 0; i < 5; i++ {
		if s.gs.Army[i].TroopID == 0xFF || s.gs.Army[i].Count == 0 {
			break
		}
		n++
	}
	return n
}

func (s *DismissArmyScreen) Update(a input.Action) Transition {
	if s.mode == dismissModeConfirmLast {
		switch a.Kind {
		case input.ActLetter:
			if a.Rune == 'y' {
				return s.doDismiss(s.pend)
			}
			if a.Rune == 'n' {
				return Pop()
			}
		case input.ActConfirm:
			return s.doDismiss(s.pend)
		case input.ActCancel:
			return Pop()
		}
		return Stay()
	}

	switch a.Kind {
	case input.ActLetter:
		if a.Rune >= 'a' && a.Rune <= 'e' {
			slot := int(a.Rune - 'a')
			if slot >= s.armyCount() {
				return Stay() // 空格,忽略
			}
			// 只剩一支且要解散它 → 先確認(對齊 C max==1 分支)。
			if s.armyCount() == 1 {
				s.mode = dismissModeConfirmLast
				s.pend = slot
				return Stay()
			}
			return s.doDismiss(slot)
		}
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

// doDismiss 解散第 slot 格;若解散後隊伍清空,觸發 temp_death(對齊 C 呼叫端),然後回地圖。
func (s *DismissArmyScreen) doDismiss(slot int) Transition {
	if s.gs == nil {
		return Pop()
	}
	// 沿用 gamestate 的隊伍移除(與 garrison 撤離同一套 dismiss_troop 語意)。
	for i := slot; i < 4; i++ {
		s.gs.Army[i] = s.gs.Army[i+1]
	}
	s.gs.Army[4] = gamestate.Squad{TroopID: 0xFF, Count: 0}

	// 隊伍清空 → temp_death(對齊 C game.c:6788 test_defeat 失敗分支)。
	if s.armyCount() == 0 {
		s.gs.TempDeath(s.assets)
	}
	return Pop()
}

func (s *DismissArmyScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 0)

	var lines []string
	if s.mode == dismissModeConfirmLast {
		lines = []string{
			"若解散最後一支部隊,",
			"你將蒙羞被遣送回",
			"國王身邊。",
			"確定解散最後部隊 (y/n)?",
		}
	} else {
		lines = []string{"解散哪一支部隊", ""}
		for i := 0; i < s.armyCount(); i++ {
			name := ""
			tid := s.gs.Army[i].TroopID
			if s.assets != nil && tid >= 0 && tid < len(s.assets.Troops) {
				name = s.assets.Troops[tid].Name
			}
			lines = append(lines, fmt.Sprintf("  %c) %3d %s", 'A'+i, s.gs.Army[i].Count, name))
		}
	}
	drawBottomFrame(dst, s.assets, lines)
}

func (s *DismissArmyScreen) Keymap() input.Keymap {
	if s.mode == dismissModeConfirmLast {
		return input.Keymap{
			Directions: false, Confirm: "確定", Cancel: "取消",
			Letters: []input.LetterItem{{Rune: 'y', Label: "確定"}, {Rune: 'n', Label: "取消"}},
		}
	}
	letters := make([]input.LetterItem, s.armyCount())
	for i := 0; i < s.armyCount(); i++ {
		letters[i] = input.LetterItem{Rune: rune('a' + i), Label: fmt.Sprintf("%c", 'A'+i)}
	}
	return input.Keymap{Directions: false, Cancel: "離開", Letters: letters}
}
