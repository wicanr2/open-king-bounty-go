package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// 法術動作類型,對齊 C bounty.h:321-333 SPELL_*。spellActions[i] 把法術索引(0-13)
// 對到動作類型;choose_spell 依此 dispatch。
const (
	SpellDamage      = 0
	SpellClone       = 1
	SpellTeleport    = 2
	SpellFreeze      = 3
	SpellResurrect   = 4
	SpellTurnUndead  = 5
	SpellBridge      = 6
	SpellTimestop    = 7
	SpellFindVillain = 8
	SpellCastleGate  = 9
	SpellTownGate    = 10
	SpellInstantArmy = 11
	SpellRaiseControl = 12
)

// spellActions/spellPowers 逐一對齊 C bounty.c:425/443(spell_actions[]/spell_powers[])。
// 索引 = 法術 id(0-13);前 7 個是戰鬥法術、後 7 個是冒險法術(choose_spell 分兩欄)。
var spellActions = [maxSpells]int{
	SpellClone, SpellTeleport, SpellDamage, SpellDamage, SpellFreeze, SpellResurrect, SpellTurnUndead,
	SpellBridge, SpellTimestop, SpellFindVillain, SpellCastleGate, SpellTownGate, SpellInstantArmy, SpellRaiseControl,
}
var spellPowers = [maxSpells]int{
	10, 0, 25, 10, 1, 1, 50,
	2, 40, 1, 1, 1, 1, 1,
}

// SpellAction 回傳法術 id 的動作類型(spell_actions[id])。
func SpellAction(spellID int) int {
	if spellID < 0 || spellID >= maxSpells {
		return -1
	}
	return spellActions[spellID]
}

// SpellBasePower 回傳法術 id 的基礎威力(spell_powers[id])。
func SpellBasePower(spellID int) int {
	if spellID < 0 || spellID >= maxSpells {
		return 0
	}
	return spellPowers[spellID]
}

// AdventureResultKind 標記冒險法術施放的結果類型,供 ChooseSpellScreen 顯示對應訊息。
type AdventureResultKind int

const (
	AdvNotLearned  AdventureResultKind = iota // 尚未學會該法術
	AdvTimestop                               // 停時術:剩餘停時回合 +N
	AdvRaiseControl                           // 提升控制:領導力 +N
	AdvInstantArmy                            // 即時軍隊:招來 N 名某兵種
	AdvInstantArmyFail                        // 即時軍隊失敗(無空格/無此兵種)
	AdvTodo                                   // 待後續切片實作(bridge/findvillain/gate)
)

// AdventureResult 描述冒險法術施放結果。
type AdventureResult struct {
	Kind    AdventureResultKind
	Amount  int // timestop/raisecontrol/instantarmy 的數量
	TroopID int // instantarmy 招來的兵種
}

// CastAdventureSpell 施放冒險(overworld)法術,對齊 C choose_spell 的 mode==1 dispatch
// (game.c:4385-4396)。目前實作三個純 GameState 效果(停時/提升控制/即時軍隊);
// bridge/findvillain/castlegate/towngate 需目標選取或額外世界狀態,回 AdvTodo 待後續切片。
// 成功施放才消耗一個該法術(對齊 C `game->spells[spell_id]--`)。
func (gs *GameState) CastAdventureSpell(a *kbdata.Assets, spellID int) AdventureResult {
	if spellID < 0 || spellID >= maxSpells || gs.Spells[spellID] == 0 {
		return AdventureResult{Kind: AdvNotLearned}
	}
	var r AdventureResult
	switch SpellAction(spellID) {
	case SpellTimestop:
		gs.TimeStop += gs.SpellPower * 10
		r = AdventureResult{Kind: AdvTimestop, Amount: gs.SpellPower * 10}
	case SpellRaiseControl:
		amt := gs.SpellPower * 100
		gs.Leadership += amt
		r = AdventureResult{Kind: AdvRaiseControl, Amount: amt}
	case SpellInstantArmy:
		r = gs.instantArmy(a)
	default:
		// bridge/findvillain/castlegate/towngate:待後續切片,不消耗法術。
		return AdventureResult{Kind: AdvTodo}
	}
	gs.Spells[spellID]-- // 消耗一個法術
	return r
}

// instantArmy 對齊 C instant_troop/instant_army(game.c):把本階職業的 instant_army
// 兵種 (SpellPower+1)*multiplier[rank] 隻加進隊伍(同兵種併格或找空格);無空格且無同
// 兵種則失敗。
func (gs *GameState) instantArmy(a *kbdata.Assets) AdventureResult {
	if a == nil || gs.Class < 0 || gs.Class >= len(a.Classes) || gs.Rank < 0 || gs.Rank >= len(a.Classes[gs.Class]) {
		return AdventureResult{Kind: AdvInstantArmyFail}
	}
	troopID := a.Classes[gs.Class][gs.Rank].InstantArmy
	slot := -1
	for i := 0; i < 5; i++ {
		if gs.Army[i].TroopID == troopID || gs.Army[i].Count == 0 {
			slot = i
			break
		}
	}
	if slot == -1 {
		return AdventureResult{Kind: AdvInstantArmyFail}
	}
	number := (gs.SpellPower + 1) * instantArmyMultiplier[gs.Rank]
	gs.Army[slot].TroopID = troopID
	gs.Army[slot].Count += number
	return AdventureResult{Kind: AdvInstantArmy, Amount: number, TroopID: troopID}
}

// instantArmyMultiplier 對齊 C bounty.c:213 instant_army_multiplier[MAX_RANKS]。
var instantArmyMultiplier = [maxRanks]int{3, 2, 1, 1}
