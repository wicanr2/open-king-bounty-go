package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// AudienceScreen 對齊 C audience_with_king(game.c:2130):家鄉城堡 B) 謁見國王。
// 兩頁對話(C 用兩個 KB_BottomBox modal + 空白鍵翻頁):
//   - 第 1 頁:國王起身迎接的固定台詞。
//   - 第 2 頁依「還需捕獲幾名惡棍」分三種:可晉升則晉升並道賀(最高階則改催討權杖);
//     未達門檻則告知還差幾名。晉升(promote_player)在算第 2 頁文字前發生,故道賀詞用
//     的是晉升後的新階職稱(對齊 C 先 promote_player 再 sprintf title)。
type AudienceScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	page1  []string
	page2  []string
	page   int // 0 或 1
}

// NewAudienceScreen 建立謁見畫面,並在此判定/執行晉升(對齊 C:audience_with_king
// 本體就會呼叫 promote_player,不是另設按鈕)。
func NewAudienceScreen(gs *gamestate.GameState, a *kbdata.Assets) *AudienceScreen {
	s := &AudienceScreen{gs: gs, assets: a}
	s.page1 = []string{
		"號角以皇家樂聲",
		"宣告你的到來。",
		"",
		"馬克馬斯國王自王座",
		"起身迎接你,並",
		"宣示:      (空白鍵)",
	}
	s.page2 = s.buildPage2()
	return s
}

// buildPage2 對齊 C audience_with_king 的三分支訊息(含晉升副作用)。
func (s *AudienceScreen) buildPage2() []string {
	if s.gs == nil || s.assets == nil {
		return []string{"……"}
	}
	name := s.gs.Name
	needed := s.gs.VillainsNeeded(s.assets)
	if needed <= 0 {
		if s.gs.Rank < 4-1 { // MAX_RANKS-1
			s.gs.Promote(s.assets)
			return []string{
				"",
				fmt.Sprintf("恭喜你,%s,", name),
				"",
				"朕現在晉升你為",
				fmt.Sprintf("%s。", s.rankTitle()),
			}
		}
		title := s.rankTitle()
		return []string{
			"",
			fmt.Sprintf("%s,%s,", name, title),
			"",
			"快去尋回朕的秩序權杖,",
			"否則一切都將",
			"毀於一旦!",
		}
	}
	return []string{
		"",
		fmt.Sprintf("親愛的 %s,", name),
		"",
		fmt.Sprintf("待你再捕獲 %d 名惡棍", needed),
		"之後,朕便能",
		"更好地助你一臂之力。",
	}
}

// rankTitle 回傳玩家目前階級職稱,對齊 C classes[class][rank].title
// (Go 版 Class.Name 就是該階職稱,見 tables_classes.go)。越界回空字串。
func (s *AudienceScreen) rankTitle() string {
	if s.assets == nil || s.gs.Class < 0 || s.gs.Class >= len(s.assets.Classes) {
		return ""
	}
	row := s.assets.Classes[s.gs.Class]
	if s.gs.Rank < 0 || s.gs.Rank >= len(row) {
		return ""
	}
	return row[s.gs.Rank].Name
}

func (s *AudienceScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActConfirm, input.ActLetter:
		if s.page == 0 {
			s.page = 1
			return Stay()
		}
		return Pop()
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

func (s *AudienceScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 0)
	lines := s.page1
	if s.page == 1 {
		lines = s.page2
	}
	drawBottomFrame(dst, s.assets, lines)
}

func (s *AudienceScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: false, Confirm: "繼續", Cancel: "離開"}
}
