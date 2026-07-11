package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// GameOptionsScreen 對齊 C game_options_menu(game.c:5219,由 F12 debug 選單 → 開啟):
// 六項遊戲調整選項,字母鍵 A-F 切換/循環,ESC 離開。各項值即時寫回 gs.Opt*(隨存檔持久),
// 並各自接進機制(見 gamestate/time.go·week.go·worldgen.go·army.go 與 combat AIEvolved)。
type GameOptionsScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
}

// NewGameOptionsScreen 建立遊戲調整設定畫面。
func NewGameOptionsScreen(gs *gamestate.GameState, a *kbdata.Assets) *GameOptionsScreen {
	return &GameOptionsScreen{gs: gs, assets: a}
}

// optItem 一個選項:字母鍵 + 標籤 + 切換動作 + 目前值文字。
type optItem struct {
	key    rune
	label  string
	toggle func(gs *gamestate.GameState)
	value  func(gs *gamestate.GameState) string
}

func onOff(b bool) string {
	if b {
		return "開"
	}
	return "關"
}

// optItems 對齊 C game_options_menu 的六項(順序即 A-F)。
var optItems = []optItem{
	{'a', "允許不支付工資",
		func(gs *gamestate.GameState) { gs.OptNoWages = !gs.OptNoWages },
		func(gs *gamestate.GameState) string { return onOff(gs.OptNoWages) }},
	{'b', "戰鬥 AI 模式",
		func(gs *gamestate.GameState) { gs.OptAIMode = !gs.OptAIMode },
		func(gs *gamestate.GameState) string {
			if gs.OptAIMode {
				return "進化版"
			}
			return "原版"
		}},
	{'c', "回合數延長兩倍",
		func(gs *gamestate.GameState) {
			// 對齊 C:切換當下即時調整剩餘天數(開=加倍、關=折半還原,play.c/game.c:5300)。
			gs.OptDaysX2 = !gs.OptDaysX2
			if gs.OptDaysX2 {
				gs.DaysLeft *= 2
			} else {
				gs.DaysLeft = (gs.DaysLeft + 1) / 2
			}
		},
		func(gs *gamestate.GameState) string { return onOff(gs.OptDaysX2) }},
	{'d', "敵人每週成長",
		func(gs *gamestate.GameState) { gs.OptFoeFreq = (gs.OptFoeFreq + 1) % 3 },
		func(gs *gamestate.GameState) string {
			switch gs.OptFoeFreq {
			case 1:
				return "加倍"
			case 2:
				return "停止"
			}
			return "正常"
		}},
	{'e', "隨機敵人強度",
		func(gs *gamestate.GameState) { gs.OptFoeStrength = !gs.OptFoeStrength },
		func(gs *gamestate.GameState) string {
			if gs.OptFoeStrength {
				return "強(兵力x2)"
			}
			return "正常"
		}},
	{'f', "限制高階招募上限",
		func(gs *gamestate.GameState) { gs.OptRecruitCaps = !gs.OptRecruitCaps },
		func(gs *gamestate.GameState) string { return onOff(gs.OptRecruitCaps) }},
}

func (s *GameOptionsScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActCancel:
		return Pop()
	case input.ActLetter:
		if s.gs == nil {
			return Stay()
		}
		for _, it := range optItems {
			if it.key == a.Rune {
				it.toggle(s.gs)
				return Stay()
			}
		}
	}
	return Stay()
}

func (s *GameOptionsScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	drawTopBox(dst, s.assets, "遊戲調整設定(按字母切換,ESC 離開)")
	lines := make([]string, 0, len(optItems)+1)
	lines = append(lines, "遊戲調整設定:")
	for _, it := range optItems {
		val := ""
		if s.gs != nil {
			val = it.value(s.gs)
		}
		lines = append(lines, fmt.Sprintf("%c) %s:%s", it.key-32, it.label, val))
	}
	drawBottomFrame(dst, s.assets, lines)
}

func (s *GameOptionsScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions:  false,
		Cancel:      "離開",
		Letters:     optLetters(),
		LetterRects: optLetterRects(),
	}
}

// optLetters 產選項的觸控字母列。
func optLetters() []input.LetterItem {
	items := make([]input.LetterItem, len(optItems))
	for i, it := range optItems {
		items[i] = input.LetterItem{Rune: it.key, Label: it.label}
	}
	return items
}

// optLetterRects 讓每個選項的觸控熱區疊在選單對應行(line 0 = 標題,item i 在 line 1+i)。
func optLetterRects() []input.Rect {
	rects := make([]input.Rect, len(optItems))
	for i := range optItems {
		line := 1 + i
		rects[i] = input.Rect{
			X: bottomTextX,
			Y: bottomHeaderY + line*render.CJKCell,
			W: bottomBorderW - 2*render.CJKCell,
			H: render.CJKCell,
		}
	}
	return rects
}
