package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// CheatMenuScreen 對齊 C debug_cheat_menu(game.c:5052,原 F10 → 改 F12)。移植常用作弊項
// 到 Go/Ebiten,並讓觸控可用(每項是選單一行 + LetterRects 熱區)。由世界地圖 F12 或
// 觸控右上「作弊」鈕開啟。cheat 直接改寫 GameState;'w' 立即勝利(對齊 war==NULL → win_game)。
type CheatMenuScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	msg    string
}

// NewCheatMenuScreen 建立作弊選單。
func NewCheatMenuScreen(gs *gamestate.GameState, a *kbdata.Assets) *CheatMenuScreen {
	return &CheatMenuScreen{gs: gs, assets: a}
}

// cheatItem 一個作弊項:字母鍵 + 選單標籤 + 套用到 gs 的動作(回傳結果訊息)。
// apply 為 nil 表示由 Update 特別處理(如 'w' 勝利要轉場)。
type cheatItem struct {
	key   rune
	label string
	apply func(gs *gamestate.GameState) string
}

// cheatItems 對齊 C debug_cheat_menu 的常用項(g/v/u/m/f/x/n/o/w)。
var cheatItems = []cheatItem{
	{'g', "+5000 金幣", func(gs *gamestate.GameState) string { gs.Gold += 5000; return "已 +5000 金幣" }},
	{'v', "+100 領導力", func(gs *gamestate.GameState) string {
		gs.Leadership += 100
		gs.BaseLeadership += 100
		return "已 +100 領導力"
	}},
	{'u', "學會所有法術", func(gs *gamestate.GameState) string {
		for i := range gs.Spells {
			gs.Spells[i]++
		}
		return "已學會所有法術"
	}},
	{'m', "學會魔法", func(gs *gamestate.GameState) string {
		if gs.KnowsMagic {
			gs.SpellPower++
		}
		gs.KnowsMagic = true
		return "已學會魔法"
	}},
	{'x', "神級隊伍", func(gs *gamestate.GameState) string { applyFlightTeam(gs); return "已獲得神級隊伍" }},
	{'n', "解鎖下一洲海圖", func(gs *gamestate.GameState) string {
		if gs.Continent+1 < kbdata.MaxContinents {
			gs.ContinentFound[gs.Continent+1] = 1
		}
		gs.OrbFound[gs.Continent] = 1   // 順帶揭示本洲小地圖(對齊 C 'o')
		gs.Mount = gamestate.KBMountFly // 順帶飛行(對齊 C 'f',一併給機動力)
		return "已解鎖下一洲 + 寶珠 + 飛行"
	}},
	{'o', "遊戲選項", nil}, // 由 Update 開啟 GameOptionsScreen(對齊 C debug 選單 → game_options_menu)
	{'w', "立即勝利", nil}, // 由 Update 轉場到 WinScreen
}

// applyFlightTeam 對齊 C 'x'(Flight team):精靈/大法師/吸血鬼/惡魔/火龍 各 1。
func applyFlightTeam(gs *gamestate.GameState) {
	team := [5]int{0x01, 0x14, 0x15, 0x17, 0x18}
	for i, id := range team {
		gs.Army[i] = gamestate.Squad{TroopID: id, Count: 1}
	}
}

func (s *CheatMenuScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActCancel:
		return Pop()
	case input.ActLetter:
		if a.Rune == 'w' {
			return Replace(NewCartoonScreen(s.gs, s.assets)) // 過場動畫 → 破關訊息
		}
		if a.Rune == 'o' {
			return Push(NewGameOptionsScreen(s.gs, s.assets))
		}
		for _, c := range cheatItems {
			if c.key == a.Rune && c.apply != nil {
				s.msg = c.apply(s.gs)
				return Stay()
			}
		}
	}
	return Stay()
}

func (s *CheatMenuScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	drawTopBox(dst, s.assets, "作弊選單(按字母,ESC 離開)")
	drawSidebar(dst, s.gs, 0)
	head := "作弊指令(按字母鍵):"
	if s.msg != "" {
		head = s.msg // 套用結果直接蓋在標題行,避免多一行溢出底框
	}
	lines := []string{head}
	for _, c := range cheatItems {
		lines = append(lines, fmt.Sprintf("%c) %s", c.key-32, c.label))
	}
	drawBottomFrame(dst, s.assets, lines)
}

func (s *CheatMenuScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions:  false,
		Cancel:      "離開",
		Letters:     cheatLetters(),
		LetterRects: cheatLetterRects(),
	}
}

// cheatLetters 產作弊項的觸控字母列(短標籤)。
func cheatLetters() []input.LetterItem {
	items := make([]input.LetterItem, len(cheatItems))
	for i, c := range cheatItems {
		items[i] = input.LetterItem{Rune: c.key, Label: c.label}
	}
	return items
}

// cheatLetterRects 讓每個作弊項的觸控熱區疊在選單對應行(對齊 town 的做法)。
// 選單 line 0 = 標題,item i 在 line 1+i。
func cheatLetterRects() []input.Rect {
	rects := make([]input.Rect, len(cheatItems))
	for i := range cheatItems {
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
