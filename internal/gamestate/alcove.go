package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// CostAlcove 對齊 C COST_ALCOVE(bounty.h:25):大法師傳授魔法的價碼。
const CostAlcove = 5000

// AtAlcove 回報玩家目前是否站在魔法密室(land.ini special1 = SP_ALCOVE)座標上。
func (gs *GameState) AtAlcove(a *kbdata.Assets) bool {
	c, x, y, ok := AlcoveCoords(a)
	return ok && gs.Continent == c && gs.X == x && gs.Y == y
}

// LearnMagicAtAlcove 對齊 C visit_alcove 的接受分支(game.c):金幣足則付 CostAlcove、
// 習得魔法(KnowsMagic=true),回 true;金幣不足回 false(不扣錢)。
func (gs *GameState) LearnMagicAtAlcove() bool {
	if gs.Gold < CostAlcove {
		return false
	}
	gs.Gold -= CostAlcove
	gs.KnowsMagic = true
	return true
}
