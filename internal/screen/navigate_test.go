package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// TestNavigate_OnlyListsFoundContinents 驗證換洲選單只列已探索的洲。
func TestNavigate_OnlyListsFoundContinents(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	// 起手僅家鄉洲(0)已探索;再開放洲 2。
	gs.ContinentFound[2] = 1

	s := NewNavigateContinentScreen(gs, a)
	if len(s.options) != 2 {
		t.Fatalf("應只列 2 洲(0 與 2),got %d", len(s.options))
	}
	ids := map[int]bool{s.options[0].id: true, s.options[1].id: true}
	if !ids[0] || !ids[2] {
		t.Errorf("換洲選項洲別錯誤:%+v", s.options)
	}
}

// TestNavigate_SailToMovesAndSpendsWeek 驗證選洲後 SailTo(移到入口)+ 消耗一週(推進天數 +
// 觸發週結算畫面)。換洲會跨一個週界 → Replace 成 EndOfWeekScreen(關閉後回地圖)。
func TestNavigate_SailToMovesAndSpendsWeek(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.ContinentFound[2] = 1
	weekBefore := gs.Week
	daysBefore := gs.DaysLeft
	entry := gamestate.ContinentEntry(a)

	s := NewNavigateContinentScreen(gs, a)
	// options[0]=洲0、options[1]=洲2;選第 2 項(按 '2')前往洲 2。
	tr := s.Update(input.Letter('2'))
	if tr.Kind != KindReplace {
		t.Fatalf("選洲後應 Replace 成週結算畫面,got %v", tr.Kind)
	}
	if _, ok := tr.Next.(*EndOfWeekScreen); !ok {
		t.Fatalf("換洲週結算應 Replace 成 *EndOfWeekScreen,got %T", tr.Next)
	}
	// 換洲從本週剩餘天數推進到下一週界:起手 passed_days=0 → 花 WeekDays(5)天。
	if gs.DaysLeft != daysBefore-gamestate.WeekDays {
		t.Errorf("換洲應花本週剩餘天數:DaysLeft %d→%d(want -%d)", daysBefore, gs.DaysLeft, gamestate.WeekDays)
	}
	if gs.Continent != 2 {
		t.Errorf("未切換到洲 2:Continent=%d", gs.Continent)
	}
	if gs.X != entry[2][0] || gs.Y != entry[2][1] {
		t.Errorf("未移到洲 2 入口 (%d,%d),got (%d,%d)", entry[2][0], entry[2][1], gs.X, gs.Y)
	}
	if int(gs.Boat) != 2 || gs.BoatX != gs.X || gs.BoatY != gs.Y {
		t.Errorf("船未跟到入口:Boat=%d boat=(%d,%d)", gs.Boat, gs.BoatX, gs.BoatY)
	}
	if gs.Week != weekBefore+1 {
		t.Errorf("換洲應消耗一週:Week %d→%d", weekBefore, gs.Week)
	}
}

// TestWorldMap_NavigateOnlyWhileSailing 驗證世界地圖按 'n' 只有乘船時才開換洲選單。
func TestWorldMap_NavigateOnlyWhileSailing(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	s := NewWorldMapScreen(gs, a)

	gs.Mount = gamestate.KBMountRide
	if tr := s.Update(input.Letter('n')); tr.Kind != KindStay {
		t.Errorf("騎乘時按 n 不應開換洲選單,got %v", tr.Kind)
	}
	gs.Mount = gamestate.KBMountSail
	if tr := s.Update(input.Letter('n')); tr.Kind != KindPush {
		t.Errorf("乘船時按 n 應 Push 換洲選單,got %v", tr.Kind)
	}
}
