package combat

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// 勝負回傳值,對齊 C combat_loop:1 = 玩家(side 0)勝,2 = AI(side 1)勝。
const (
	ResultPlayerWon = 1
	ResultAIWon     = 2
	ResultUndecided = 0
)

// RunCombat 固化 C 版 combat_loop(game.c:6097)的回合契約,headless 跑到分出勝負。
//
// 每一步讓「當前單位」動作一次:AI 側(side==1 / OutOfControl / autoBattle)呼叫
// AIUnitThink(逐 tick,可能只移一步,故對同一單位反覆呼叫直到 Acted 或 pass);
// 非 AI 的玩家側在 headless 下視為待命(pass)。當 pass || Acted 時檢查勝負,
// 未分則 NextUnit(-1 則 NextTurn)換手。maxSteps 為防呆上限(避免卡死不收斂)。
//
// autoBattle=true 時玩家側也交給 AI(對齊 C 的 auto_battle),可跑 AI vs AI。
// 回傳 ResultPlayerWon / ResultAIWon / ResultUndecided(達上限未分)。
func (c *Combat) RunCombat(a *kbdata.Assets, rng kbrng.Rand, autoBattle bool, maxSteps int) int {
	for step := 0; step < maxSteps; step++ {
		u := &c.Units[c.Side][c.UnitID]

		pass := false
		if c.Side == 1 || u.OutOfControl || autoBattle {
			pass = c.AIUnitThink(a, rng)
		} else {
			// headless 玩家側待命,語意對齊 unit_try_wait(game.c:5749):
			// 只有 phase>0 才真的標 Acted,恆回傳 pass。靠 phase 機制讓 NextUnit
			// 最終掃到並標記所有被動單位,不會卡死在 side 0。
			if c.Phase > 0 {
				u.Acted = true
			}
			pass = true
		}
		if u.Count == 0 {
			pass = true // 當前單位已陣亡 → 換手
		}

		if pass || u.Acted {
			// Winner() 回傳「勝方 side」(0=玩家 / 1=AI),-1 未分。
			if w := c.Winner(); w == 0 {
				return ResultPlayerWon
			} else if w == 1 {
				return ResultAIWon
			}
			u.Frame = 0
			if next := c.NextUnit(a); next == -1 {
				c.NextTurn(a)
			} else {
				c.UnitID = next
			}
		}
	}
	return ResultUndecided
}
