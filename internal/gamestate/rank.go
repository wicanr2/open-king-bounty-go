package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// maxRanks 對齊 C 版 MAX_RANKS(bounty.h)。
const maxRanks = 4

// acceptRank 套用「當前 rank」的職業增量,對應 C 版 player_accept_rank(play.c:760)。
// 用 += 累加(不是賦值),因此 BaseLeadership 能同時疊加升階增量與寶箱等其他來源——
// 這正是不能改用「賦值累積絕對值」的原因(否則會清掉寶箱加成,重蹈 issue #5 類 bug)。
// 不改動 Rank(遞增在 Promote 內,對齊 C 的職責切分)。
func (gs *GameState) acceptRank(a *kbdata.Assets) {
	c := a.Classes[gs.Class][gs.Rank]
	gs.BaseLeadership += c.Leadership
	gs.MaxSpells += c.MaxSpell
	gs.SpellPower += c.SpellPower
	gs.Commission += c.Commission
	if c.KnowsMagic != 0 {
		gs.KnowsMagic = true
	}
}

// PlayerCaptured 回傳已捕獲的惡棍數,對應 C player_captured(play.c:747)。
func (gs *GameState) PlayerCaptured() int {
	num := 0
	for _, c := range gs.VillainCaught {
		if c != 0 {
			num++
		}
	}
	return num
}

// VillainsNeeded 回傳「晉升到下一階還需再捕獲幾名惡棍」,對齊 C audience_with_king
// (game.c:2139):classes[class][rank+1].villains_needed - player_captured。已達最高
// 階(無下一階)回 0。<=0 代表可晉升。
func (gs *GameState) VillainsNeeded(a *kbdata.Assets) int {
	if gs.Rank >= maxRanks-1 {
		return 0
	}
	return a.Classes[gs.Class][gs.Rank+1].VillainsNeeded - gs.PlayerCaptured()
}

// Promote 升一階並套用新階增量,對應 C 版 promote_player(play.c:768):
// 已達最高階則不動;否則 Rank+1 後 acceptRank。
// 注意:Promote 不重設 Leadership(當前可用領導力),升階增量進 BaseLeadership,
// 下次 EndWeek 才把 Leadership 補到新的 BaseLeadership(對齊 C)。
func (gs *GameState) Promote(a *kbdata.Assets) {
	if gs.Rank >= maxRanks-1 {
		return
	}
	gs.Rank++
	gs.acceptRank(a)
}
