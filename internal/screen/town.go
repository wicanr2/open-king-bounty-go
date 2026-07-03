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

// townMode 是 TownScreen 目前顯示的子畫面。對應 C 版 game.c visit_town() 的
// msg_hold 概念:選單本身之外,某一項的結果會蓋一層訊息,直到玩家再按鍵才回選單。
type townMode int

const (
	townModeMenu townMode = iota // 只顯示 A-E 選單
	townModeInfo                 // C) 蒐集情報:疊顯既有 GameState 數值(GOLD/LEADERSHIP)
	townModeStub                 // A/B/D/E:需世界狀態,先顯示「未實作」提示
)

// townLetters 是城鎮選單的情境字母列,對應 C 版 game.c visit_town()(2660 行起)的
// five_choices:A) 領取新契約 B) 租船/取消租船 C) 蒐集情報 D) 學習法術(各鎮教的法術
// 不同,C 版原文以 "造橋" 為例,即 town_spell[id]==SPELL_BRIDGE 的情形) E) 購買攻城武器。
// 本切片只做選單框架 + Keymap,尚無世界狀態(哪個城鎮、教哪個法術),故用固定標籤。
var townLetters = []input.LetterItem{
	{Rune: 'a', Label: "接任務"},
	{Rune: 'b', Label: "租船"},
	{Rune: 'c', Label: "情報"},
	{Rune: 'd', Label: "造橋"},
	{Rune: 'e', Label: "買攻城"},
}

// TownScreen 是城鎮畫面:A-E 情境字母列選單,驗證「情境字母列 Keymap」設計在真實畫面
// 走通。完整的契約/租船/世界狀態邏輯不在本切片(需世界狀態);C(情報)顯示既有
// GameState 數值,其餘項先顯示「未實作」提示,重點是選單與 Keymap 走通。
type TownScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets

	sel  int      // 方向鍵目前高亮的選項(0-4,對應 townLetters 索引)
	mode townMode // 目前顯示的子畫面
	stub string   // townModeStub 時顯示的項目標籤
}

// NewTownScreen 建立城鎮畫面。
func NewTownScreen(gs *gamestate.GameState, a *kbdata.Assets) *TownScreen {
	return &TownScreen{gs: gs, assets: a}
}

func (s *TownScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActUp:
		if s.sel > 0 {
			s.sel--
		}
	case input.ActDown:
		if s.sel < len(townLetters)-1 {
			s.sel++
		}
	case input.ActConfirm:
		s.activate(s.sel)
	case input.ActLetter:
		if a.Rune >= 'a' && a.Rune <= 'e' {
			s.activate(int(a.Rune - 'a'))
		}
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

// activate 觸發第 idx 個選單項(0=A..4=E),對應 C 版 key==1..5 分支。
// C(情報,idx==2)切到情報顯示;其餘項先切到「未實作」提示。
func (s *TownScreen) activate(idx int) {
	if idx < 0 || idx >= len(townLetters) {
		return
	}
	s.sel = idx
	if townLetters[idx].Rune == 'c' {
		s.mode = townModeInfo
		return
	}
	s.mode = townModeStub
	s.stub = townLetters[idx].Label
}

func (s *TownScreen) Draw(dst *ebiten.Image) {
	drawCJK := func(str string, x, y int, clr color.Color) {
		if s.assets != nil && s.assets.Font != nil {
			render.DrawText(dst, s.assets.Font, str, x, y, clr)
		}
	}

	drawCJK("城鎮", 8, 4, color.White)

	for i, item := range townLetters {
		y := 32 + i*16
		var clr color.Color = color.Gray{Y: 160}
		marker := "  "
		if i == s.sel {
			clr = color.White
			marker = "> "
		}
		ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%s%c)", marker, 'A'+i), 8, y)
		drawCJK(item.Label, 32, y-6, clr)
	}

	switch s.mode {
	case townModeInfo:
		if s.gs != nil {
			drawCJK("情報", 8, 128, color.White)
			ebitenutil.DebugPrintAt(dst,
				fmt.Sprintf("GOLD %d   LEADERSHIP %d", s.gs.Gold, s.gs.Leadership), 8, 148)
		}
	case townModeStub:
		ebitenutil.DebugPrintAt(dst, "尚未實作: "+s.stub, 8, 148)
	}

	ebitenutil.DebugPrintAt(dst, "ESC: 離開城鎮", 8, 186)
}

func (s *TownScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: true,
		Confirm:    "選擇",
		Cancel:     "離開",
		Letters:    townLetters,
	}
}
