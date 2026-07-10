package combat

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// 戰鬥法術效果函式,對齊 C game.c/play.c 的 *_army/*_troop 底層(去掉 pick_target 互動
// 部分——目標由 CombatScreen 的目標選取器選好後傳進來)。spellPower 由呼叫端(gamestate
// GameState.SpellPower)注入。

// MagicDamage 對齊 C damage_army 的效果核心(game.c:3495):對 (tSide,tID) 施放 external
// 傷害 base*spellPower。驅散不死(undead=true)只對不死族有效;魔法免疫一律無效。
// 回傳擊殺數;無效(免疫/非不死)回 -1。
func (c *Combat) MagicDamage(a *kbdata.Assets, rng kbrng.Rand, spellPower, baseDamage, tSide, tID int, undead bool) int {
	if !validUnit(a, c, tSide, tID) {
		return -1
	}
	ab := a.Troops[c.Units[tSide][tID].TroopID].Abilities
	if (undead && ab&kbdata.AbilUndead == 0) || ab&kbdata.AbilImmune != 0 {
		return -1
	}
	return c.DealDamage(a, rng, 0, 0, tSide, tID, false, true, baseDamage*spellPower, true)
}

// CloneTroop 對齊 C clone_troop(play.c):把玩家(side 0)第 unitID 格複製出
// (spellPower*10 + injury)/hp 隻,餘數留 injury。回傳複製數。
func (c *Combat) CloneTroop(a *kbdata.Assets, spellPower, unitID int) int {
	if !validUnit(a, c, 0, unitID) {
		return 0
	}
	u := &c.Units[0][unitID]
	hp := a.Troops[u.TroopID].HP
	if hp <= 0 {
		return 0
	}
	damage := spellPower*10 + u.Injury
	clones := damage / hp
	u.Count += clones
	u.MaxCount += clones
	u.Injury = damage % hp
	return clones
}

// FreezeTroop 對齊 C freeze_troop(play.c:1876 後段):凍結 (side,id) 使其本回合略過;
// 魔法免疫者無效回 -1,成功回 1。
func (c *Combat) FreezeTroop(a *kbdata.Assets, side, id int) int {
	if !validUnit(a, c, side, id) {
		return -1
	}
	u := &c.Units[side][id]
	if a.Troops[u.TroopID].Abilities&kbdata.AbilImmune != 0 {
		return -1
	}
	u.Frozen = true
	return 1
}

// ResurrectTroop 對齊 C resurrect_troop(play.c):復活 (side,id) 本場戰死的部隊,
// 每點 spellPower 復活 1 隻,上限為戰死數(MaxCount-Count)。回傳復活數。
func (c *Combat) ResurrectTroop(spellPower, side, id int) int {
	if side < 0 || side >= MaxSides || id < 0 || id >= MaxUnits {
		return 0
	}
	u := &c.Units[side][id]
	dead := u.MaxCount - u.Count
	revived := spellPower
	if revived > dead {
		revived = dead
	}
	if revived <= 0 {
		return 0
	}
	u.Count += revived
	u.TurnCount = u.Count
	if u.Count > 0 {
		u.Dead = false
	}
	return revived
}

// RelocateUnit 對齊 C unit_relocate(play.c:1635):把 (side,id) 移到 (nx,ny),
// 更新 Umap 佔格。
func (c *Combat) RelocateUnit(side, id, nx, ny int) {
	if side < 0 || side >= MaxSides || id < 0 || id >= MaxUnits {
		return
	}
	u := &c.Units[side][id]
	c.Umap[u.Y][u.X] = 0
	c.Umap[ny][nx] = PackUID(side, id)
	u.X, u.Y = nx, ny
}

// validUnit 檢查 (side,id) 索引合法且該格兵種是 a.Troops 合法索引。
func validUnit(a *kbdata.Assets, c *Combat, side, id int) bool {
	if a == nil || side < 0 || side >= MaxSides || id < 0 || id >= MaxUnits {
		return false
	}
	tid := c.Units[side][id].TroopID
	return tid >= 0 && tid < len(a.Troops)
}
