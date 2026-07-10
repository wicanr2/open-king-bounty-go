package combat

import "github.com/wicanr2/open-king-bounty-go/internal/gamestate"

// AcceptUnitsPlayer 把戰鬥結束後 side 側的存活單位寫回玩家隊伍 gs.Army,對齊 C
// accept_units_player(play.c:1185):逐格覆寫 troop_id/count(含陣亡歸零的格)。
// 對應 C 版在 combat_loop 收尾把每個 hero 的隊伍寫回 game 的行為(play.c:1453)。
func AcceptUnitsPlayer(c *Combat, side int, gs *gamestate.GameState) {
	if gs == nil {
		return
	}
	for i := 0; i < MaxUnits; i++ {
		gs.Army[i] = gamestate.Squad{
			TroopID: c.Units[side][i].TroopID,
			Count:   c.Units[side][i].Count,
		}
	}
}

// AcceptSquads 回傳戰鬥結束後 side 側前 n 格的 (troop_id, count),供呼叫端寫回世界
// 狀態(foe / castle)。對齊 C accept_units_foe(前 3 格)/accept_units_castle(前 5 格)
// ——差別只在寫回目標與格數,故抽成回傳 parallel array 由呼叫端各自寫入。
func AcceptSquads(c *Combat, side, n int) (troops [MaxUnits]int, numbers [MaxUnits]int) {
	for i := 0; i < MaxUnits; i++ {
		troops[i] = 0xFF
	}
	if n > MaxUnits {
		n = MaxUnits
	}
	for i := 0; i < n; i++ {
		troops[i] = c.Units[side][i].TroopID
		numbers[i] = c.Units[side][i].Count
	}
	return troops, numbers
}
