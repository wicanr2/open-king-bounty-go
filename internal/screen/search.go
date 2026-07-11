package screen

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// searchMode 是 SearchScreen 的子狀態(詢問 / 一無所獲結果)。
type searchMode int

const (
	searchAsk searchMode = iota
	searchFutile
)

// SearchScreen 對齊 C ask_search(game.c:4687):問「搜索這片區域(y/n)?」。若玩家正站
// 在失竊權杖埋藏處 → 獲勝(Replace 成 WinScreen);否則一無所獲(顯示訊息後回地圖)。
// 由世界地圖按 'g' 開啟。
type SearchScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	mode   searchMode
}

// NewSearchScreen 建立搜索詢問畫面。
func NewSearchScreen(gs *gamestate.GameState, a *kbdata.Assets) *SearchScreen {
	return &SearchScreen{gs: gs, assets: a}
}

func (s *SearchScreen) Update(a input.Action) Transition {
	if s.mode == searchFutile {
		if !a.IsNone() {
			return Pop()
		}
		return Stay()
	}
	confirm := a.Kind == input.ActConfirm || (a.Kind == input.ActLetter && a.Rune == 'y')
	if confirm {
		if s.gs.AtScepter() {
			return Replace(NewWinScreen(s.gs, s.assets)) // 找到權杖 → 破關
		}
		s.mode = searchFutile
		return Stay()
	}
	if a.Kind == input.ActCancel || (a.Kind == input.ActLetter && a.Rune == 'n') {
		return Pop()
	}
	return Stay()
}

func (s *SearchScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 0)
	if s.mode == searchFutile {
		drawBottomFrame(dst, s.assets, []string{"你搜索這片區域", "一無所獲。"})
		return
	}
	drawBottomFrame(dst, s.assets, []string{
		"搜索……",
		"",
		"搜索這片區域",
		"將花費 10 天。",
		"",
		"搜索 (y/n)?",
	})
}

func (s *SearchScreen) Keymap() input.Keymap {
	if s.mode == searchFutile {
		return input.Keymap{Directions: false, Confirm: "繼續", Cancel: "離開"}
	}
	return input.Keymap{
		Directions: false, Confirm: "搜索", Cancel: "取消",
		Letters: []input.LetterItem{{Rune: 'y', Label: "搜索"}, {Rune: 'n', Label: "取消"}},
	}
}
