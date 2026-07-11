package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// EndOfWeekScreen 對齊 C end_of_week(game.c:3758):每過一週顯示兩頁——① 占星師宣告
// 「第 N 週,本週為 X 之週,所有 X 據點已重新補滿」② 收支表(現有/佣金/船隻/軍隊/結餘)。
// 世界增長(gs.EndWeek)已於進場前套用,本畫面只顯示結果;空白/確認翻頁,第二頁後關閉。
// beforeGold = 結算前金幣(供收支「現有」欄,結算已改動 gs.Gold)。
type EndOfWeekScreen struct {
	gs         *gamestate.GameState
	assets     *kbdata.Assets
	creature   int
	beforeGold int
	page       int
}

// NewEndOfWeekScreen 建立週結算畫面(creature = 本週生物 id,beforeGold = 結算前金幣)。
func NewEndOfWeekScreen(gs *gamestate.GameState, a *kbdata.Assets, creature, beforeGold int) *EndOfWeekScreen {
	return &EndOfWeekScreen{gs: gs, assets: a, creature: creature, beforeGold: beforeGold}
}

func (s *EndOfWeekScreen) Update(a input.Action) Transition {
	if a.Kind == input.ActConfirm || a.Kind == input.ActCancel {
		if s.page == 0 {
			s.page = 1 // 占星頁 → 收支頁
			return Stay()
		}
		return Pop() // 收支頁 → 關閉,回世界地圖
	}
	return Stay()
}

// creatureName 回傳本週生物名(界外退回「生物」)。
func (s *EndOfWeekScreen) creatureName() string {
	if s.assets != nil && s.creature >= 0 && s.creature < len(s.assets.Troops) {
		return s.assets.Troops[s.creature].Name
	}
	return "生物"
}

// armyCost 回傳隊伍每週維護費(Σ Count × recruit_cost/10,對齊 EndWeek 的扣費)。
func (s *EndOfWeekScreen) armyCost() int {
	if s.gs == nil || s.assets == nil {
		return 0
	}
	cost := 0
	for i := 0; i < 5; i++ {
		sq := s.gs.Army[i]
		if sq.Count == 0 || sq.TroopID < 0 || sq.TroopID >= len(s.assets.Troops) {
			continue
		}
		cost += sq.Count * (s.assets.Troops[sq.TroopID].GoldCost / 10)
	}
	return cost
}

func (s *EndOfWeekScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	week := 0
	if s.gs != nil {
		week = s.gs.WeekID()
	}
	if s.page == 0 {
		name := s.creatureName()
		drawTopBox(dst, s.assets, fmt.Sprintf("第 %d 週", week))
		drawBottomFrame(dst, s.assets, []string{
			"占星師宣告:",
			"",
			fmt.Sprintf("本週為「%s」之週", name),
			"",
			fmt.Sprintf("所有%s據點已重新補滿。", name),
			"",
			"(按空白鍵繼續)",
		})
		return
	}
	// 收支頁
	boat := 0
	commission, gold := 0, 0
	if s.gs != nil {
		if s.gs.HasBoat() {
			boat = s.gs.BoatCost()
		}
		commission, gold = s.gs.Commission, s.gs.Gold
	}
	drawTopBox(dst, s.assets, fmt.Sprintf("第 %d 週  收支", week))
	drawBottomFrame(dst, s.assets, []string{
		fmt.Sprintf("現有  %6d", s.beforeGold),
		fmt.Sprintf("佣金  %6d", commission),
		fmt.Sprintf("船隻  %6d", boat),
		fmt.Sprintf("軍隊  %6d", s.armyCost()),
		fmt.Sprintf("結餘  %6d", gold),
		"",
		"(按空白鍵繼續)",
	})
}

func (s *EndOfWeekScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: false, Confirm: "繼續"}
}
