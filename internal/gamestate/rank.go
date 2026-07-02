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
