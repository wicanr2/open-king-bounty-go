package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// TestTimeInit 驗證新遊戲起手時間狀態(對齊 C spawn_game:days_left=難度天數、steps_left=40)。
func TestTimeInit(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.MaxDays() != 600 {
		t.Errorf("MaxDays = %d, want 600(Normal)", gs.MaxDays())
	}
	if gs.DaysLeft != 600 {
		t.Errorf("起手 DaysLeft = %d, want 600", gs.DaysLeft)
	}
	if gs.StepsLeft != DaySteps {
		t.Errorf("起手 StepsLeft = %d, want %d", gs.StepsLeft, DaySteps)
	}
	if gs.PassedDays() != 0 || gs.WeekID() != 0 {
		t.Errorf("起手 PassedDays=%d WeekID=%d, want 0/0", gs.PassedDays(), gs.WeekID())
	}
}

// TestSpendStepDayAndWeek 驗證走 DaySteps 步過一天、走滿 WeekDays 天剛好跨一個週界。
func TestSpendStepDayAndWeek(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)

	// 走一整天(DaySteps 步):最後一步過天,天數 -1、步數重置。
	var dayEnded bool
	for i := 0; i < DaySteps; i++ {
		dayEnded, _ = gs.SpendStep()
	}
	if !dayEnded {
		t.Errorf("走滿 %d 步應過一天(dayEnded)", DaySteps)
	}
	if gs.DaysLeft != 599 {
		t.Errorf("走一天後 DaysLeft = %d, want 599", gs.DaysLeft)
	}
	if gs.StepsLeft != DaySteps {
		t.Errorf("過天後 StepsLeft 應重置為 %d, got %d", DaySteps, gs.StepsLeft)
	}

	// 從頭再走到第一個週界:總共 WeekDays 天。續走剩下的 (WeekDays-1) 天,最後一天跨週界。
	weekCount := 0
	for d := 0; d < WeekDays-1; d++ {
		for i := 0; i < DaySteps; i++ {
			if _, w := gs.SpendStep(); w {
				weekCount++
			}
		}
	}
	if weekCount != 1 {
		t.Errorf("走滿 %d 天應剛好跨 1 個週界,got %d", WeekDays, weekCount)
	}
	if gs.DaysLeft != 600-WeekDays {
		t.Errorf("走 %d 天後 DaysLeft = %d, want %d", WeekDays, gs.DaysLeft, 600-WeekDays)
	}
	if gs.WeekID() != 1 {
		t.Errorf("跨一週界後 WeekID = %d, want 1", gs.WeekID())
	}
}

// TestSpendStepTimeStopFreezes 驗證停時(TimeStop>0)時走路只消耗停時、不扣步(凍結時間)。
func TestSpendStepTimeStopFreezes(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.TimeStop = 3
	stepsBefore := gs.StepsLeft
	for i := 0; i < 3; i++ {
		gs.SpendStep()
	}
	if gs.TimeStop != 0 {
		t.Errorf("停時應被消耗:TimeStop = %d, want 0", gs.TimeStop)
	}
	if gs.StepsLeft != stepsBefore {
		t.Errorf("停時中不應扣步:StepsLeft %d→%d", stepsBefore, gs.StepsLeft)
	}
	// 停時用完後,下一步正常扣步。
	gs.SpendStep()
	if gs.StepsLeft != stepsBefore-1 {
		t.Errorf("停時用完後應正常扣步:StepsLeft = %d, want %d", gs.StepsLeft, stepsBefore-1)
	}
}

// TestSpendWeekAdvancesToBoundary 驗證換洲用的 SpendWeek 花掉本週剩餘天數推進到週界並回傳生物。
func TestSpendWeekAdvancesToBoundary(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	daysBefore := gs.DaysLeft // 600,passed=0 → 應花 WeekDays 天
	creature := gs.SpendWeek(a, kbrng.NewGlibc(1))
	if gs.DaysLeft != daysBefore-WeekDays {
		t.Errorf("SpendWeek 應花 %d 天:DaysLeft %d→%d", WeekDays, daysBefore, gs.DaysLeft)
	}
	if creature < 0 {
		t.Errorf("SpendWeek 應跨一週界並回傳本週生物(>=0),got %d", creature)
	}
	if gs.WeekID() != 1 {
		t.Errorf("SpendWeek 後 WeekID = %d, want 1", gs.WeekID())
	}
}

// TestDaysReachZero 驗證天數走到 0(時間耗盡):EndDay 使 DaysLeft 遞減到 0。
func TestDaysReachZero(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.DaysLeft = 1
	gs.StepsLeft = 1
	gs.SpendStep() // 最後一步 → EndDay → DaysLeft 1→0
	if gs.DaysLeft != 0 {
		t.Errorf("最後一天走完 DaysLeft = %d, want 0(時間耗盡)", gs.DaysLeft)
	}
}
