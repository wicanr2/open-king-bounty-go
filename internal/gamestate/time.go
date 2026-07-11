package gamestate

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// 天數/時間系統常數,對齊 C(bounty.h:72-73、bounty.c:286)。
const (
	DaySteps = 40 // DAY_STEPS:一天可走幾步,步盡即過一天
	WeekDays = 5  // WEEK_DAYS:幾天為一週(每過一週觸發週結算)
)

// daysPerDifficulty 對齊 C days_per_difficulty[4](bounty.c:286):難度越高、起手天數越少。
var daysPerDifficulty = [4]int{
	900, // Easy
	600, // Normal(目前 NewGame 固定用這檔)
	400, // Hard
	200, // Impossible
}

// MaxDays 回傳本難度的起手總天數(對齊 C days_per_difficulty[difficulty])。
func (gs *GameState) MaxDays() int {
	d := gs.Difficulty
	if d < 0 || d >= len(daysPerDifficulty) {
		d = 1 // 越界安全值:Normal
	}
	return daysPerDifficulty[d]
}

// PassedDays 回傳已過天數(對齊 C passed_days = max_days - days_left)。
func (gs *GameState) PassedDays() int { return gs.MaxDays() - gs.DaysLeft }

// WeekID 回傳目前週序(對齊 C week_id = passed_days / WEEK_DAYS)。週結算畫面顯示「第 N 週」。
func (gs *GameState) WeekID() int { return gs.PassedDays() / WeekDays }

// EndDay 過一天(對齊 C end_day,play.c:504):天數 -1、步數重置為 DaySteps、停時歸零;
// 回傳是否剛好跨過一個週界(days_left % WeekDays == 0 → 該觸發週結算)。
func (gs *GameState) EndDay() bool {
	gs.DaysLeft--
	gs.StepsLeft = DaySteps
	gs.TimeStop = 0
	return gs.DaysLeft%WeekDays == 0
}

// SpendStep 花一步(世界地圖成功移動一格後呼叫),對齊 C game.c:6945-6964:
// 停時中(TimeStop>0)優先消耗停時、不扣步(停時 = 凍結時間);否則扣一步。扣完若步數耗盡
// 就 EndDay 過一天。回傳 (dayEnded, weekEnded):weekEnded 時呼叫端應跑 EndWeek + 顯示週結算,
// 之後應檢查 DaysLeft<=0 判失敗。
func (gs *GameState) SpendStep() (dayEnded, weekEnded bool) {
	if gs.TimeStop > 0 {
		gs.TimeStop--
	} else {
		gs.StepsLeft--
	}
	if gs.StepsLeft <= 0 {
		weekEnded = gs.EndDay()
		dayEnded = true
	}
	return dayEnded, weekEnded
}

// SpendWeek 花掉「本週剩餘天數」推進到下一個週界(對齊 C spend_week,play.c:524):換洲等
// 「消耗一週」的動作用。回傳本週結算選中的生物 id(-1 = 期間未跨週界,理論上不會發生);
// DaysLeft 歸零時提早停止(呼叫端須另判失敗)。
func (gs *GameState) SpendWeek(a *kbdata.Assets, rng kbrng.Rand) int {
	remaining := WeekDays - gs.PassedDays()%WeekDays
	creature := -1
	for i := 0; i < remaining && gs.DaysLeft > 0; i++ {
		if gs.EndDay() {
			creature = gs.EndWeek(a, rng)
		}
	}
	return creature
}
