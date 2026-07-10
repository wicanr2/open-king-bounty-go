package gamestate

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// maxTroops 對齊 C 版 MAX_TROOPS(bounty.h = 25)。
const maxTroops = 25

// EndWeek 執行每週結算的「玩家面」,對應 C 版 end_week(play.c:999)的前半:
//   - Leadership 重設回 BaseLeadership(升階/寶箱累積的 base 因此保留 → issue #5 的正確行為)
//   - 挑本週增長生物:每 4 週補農夫(creature 0),否則 KB_rand(1, MAX_TROOPS-1)
//   - gold += Commission
//   - 扣隊伍維護費:Σ Count × (recruit_cost/10),不足歸 0(對齊 spend_gold)
//   - 若本週生物為農夫:帶 ABIL_ABSORB 的兵(鬼)化為農夫
// 回傳本週選中的生物 id。
//
// ⚠ 暫緩(需世界狀態,後續切片):巢穴/城堡/foe 的增兵、城堡再佈署、船隻維護費、
// 以及精確的 week_id(C 用 passed_days/WEEK_DAYS;這裡用 Week 計數近似)。
// rng 由呼叫端注入(kbrng),parity 測試才能對 seed。
func (gs *GameState) EndWeek(a *kbdata.Assets, rng kbrng.Rand) int {
	// 停時歸零(對齊 C end_week 開頭 game->time_stop = 0)
	gs.TimeStop = 0
	// 重設當週可用領導力(base 保留 → 寶箱/升階加成不流失)
	gs.Leadership = gs.BaseLeadership

	gs.Week++

	// 本週增長生物:每 4 週農夫,否則隨機一種
	creature := 0
	if gs.Week%4 != 0 {
		creature = rng.Between(1, maxTroops-1)
	}

	// 俸祿
	gs.Gold += gs.Commission

	// 隊伍維護費
	credit := 0
	for i := 0; i < 5; i++ {
		if gs.Army[i].Count == 0 {
			continue
		}
		credit += gs.Army[i].Count * (a.Troops[gs.Army[i].TroopID].GoldCost / 10)
	}
	if credit <= gs.Gold {
		gs.Gold -= credit
	} else {
		gs.Gold = 0
	}

	// 農夫週:吸收型單位(鬼)化為農夫
	if creature == 0 {
		for i := 0; i < 5; i++ {
			if gs.Army[i].Count == 0 {
				continue
			}
			if a.Troops[gs.Army[i].TroopID].Abilities&kbdata.AbilAbsorb != 0 {
				gs.Army[i].TroopID = 0
			}
		}
	}

	// === 每週世界事件(對齊 C end_week 的 World events 段,play.c) ===
	gs.weeklyWorldGrowth(a, rng, creature)

	return creature
}

// weeklyWorldGrowth 對齊 C end_week 的世界事件:補滿空的玩家城堡守軍、讓「本週生物 =
// creature」的巢穴補到 max_population、foe/非玩家城堡的 creature 兵種依 growth 增長。
func (gs *GameState) weeklyWorldGrowth(a *kbdata.Assets, rng kbrng.Rand, creature int) {
	if creature < 0 || creature >= len(a.Troops) {
		return
	}
	growth := a.Troops[creature].Growth
	maxPop := a.Troops[creature].MaxPopulation

	// 補滿空的玩家城堡(對齊 C:owner==0xFF 且 castle_numbers[j][0]==0 → repopulate)。
	castles := LoadCastles(a)
	for j := 0; j < kbdata.MaxCastles; j++ {
		if gs.CastleOwner[j] == KBCastlePlayer && gs.CastleNumbers[j][0] == 0 {
			gs.repopulateCastle(rng, castles, j)
		}
	}

	for cont := 0; cont < kbdata.MaxContinents; cont++ {
		// 巢穴:本週生物的巢穴補到 max_population。
		for i := 0; i < kbdata.MaxDwellings; i++ {
			if gs.DwellingTroop[cont][i] == creature {
				gs.DwellingPopulation[cont][i] = maxPop
			}
		}
		// foe:本週生物 +growth。
		for j := 0; j < kbdata.MaxFoes; j++ {
			for i := 0; i < 3; i++ {
				if gs.FoeTroops[cont][j][i] == creature {
					gs.FoeNumbers[cont][j][i] += growth
				}
			}
		}
	}

	// 非玩家城堡:本週生物 +growth。
	for j := 0; j < kbdata.MaxCastles; j++ {
		if gs.CastleOwner[j] != KBCastlePlayer {
			for i := 0; i < 5; i++ {
				if gs.CastleTroops[j][i] == creature {
					gs.CastleNumbers[j][i] += growth
				}
			}
		}
	}
}
