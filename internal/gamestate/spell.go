package gamestate

import (
	"errors"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// ErrSpellLimitReached:法術數已達上限(對應 C visit_town() key==4 分支
// known_spells(game) >= game->max_spells 那個判斷,game.c:2809)。
var ErrSpellLimitReached = errors.New("已達法術上限")

// maxSpells 對齊 C 版 MAX_SPELLS(bounty.h:43,`7 * 2`)。
const maxSpells = 14

// spellCombatTypes 是「戰鬥法術」的 ini `type` 字串集合,對應 C 版
// src/bounty.h 的 SPELL_DAMAGE(0)..SPELL_TURNUNDEAD(5) 這六個動作碼。
// 依據 src/lib/free-data.c:GNU_spell_downto_byte() 的 names[] 表,ini 裡
// 的 type 字串與 SPELL_* 數值是逐一對應(索引即動作碼):
//
//	magic_damage=0(SPELL_DAMAGE) clone=1 teleport=2 freeze=3
//	resurrect=4 holy_damage=5(SPELL_TURNUNDEAD) bridge=6(SPELL_BRIDGE,起分界)
//	time_stop=7 find_villain=8 castle_gate=9 town_gate=10
//	instant_army=11 raise_control=12
//
// 再對照 src/game.c:4312 choose_spell() 的 switch(spell_actions[spell_id]):
// SPELL_CLONE/TELEPORT/DAMAGE/FREEZE/RESURRECT/TURNUNDEAD 呼叫的是
// clone_army/teleport_army/damage_army/freeze_army/resurrect_army,
// 這些函式都吃 KBcombat*combat 並直接操作戰鬥中的敵我隊伍 —— 是戰鬥法術。
// SPELL_BRIDGE 起(bridge/time_stop/find_villain/castle_gate/town_gate/
// instant_army/raise_control)呼叫的是 build_bridge/time_stop/find_villain/
// select_gate/instant_army/raise_control,只吃 KBgame*game,是在地圖(冒險)
// 情境下操作世界狀態 —— 是冒險法術。
//
// 交叉驗證:同一段 game.c:4256-4260 的法術選單本身也用「前 7 格戰鬥 / 後 7 格
// 冒險」的版面(spell0..spell6 標「戰鬥」,spell7..spell13 標「冒險」),而
// data/free/spells.ini 剛好照這個順序把前 7 個 type 填成戰鬥動作字串、
// 後 7 個填成冒險動作字串 —— 兩條線索(動作碼語意 / ini 版面順序)完全對齊,
// 不是巧合湊出來的規則。
var spellCombatTypes = map[string]bool{
	"magic_damage": true, // SPELL_DAMAGE(0) —— 火球術/閃電術共用此 type
	"clone":        true, // SPELL_CLONE(1)
	"teleport":     true, // SPELL_TELEPORT(2)
	"freeze":       true, // SPELL_FREEZE(3)
	"resurrect":    true, // SPELL_RESURRECT(4)
	"holy_damage":  true, // SPELL_TURNUNDEAD(5) —— 超渡術(對 undead 傷害)
}

// Spell 是單一法術的靜態資料,來源為 data/free/spells.ini 的 `[spellN]` 區塊
// (經 kbdata.FreeIni 解析,欄位語意對照 src/lib/free-data.c 的
// DAT_SACTION/WDAT_SDAMAGE/WDAT_SCOST/DAT_SFILTER 四個 KB_Resolve case)。
//
// 本型別只涵蓋「資料 + 分類 + 查詢」,不含施放效果(見套件內
// TODO(P2+) —— 施放會改到戰鬥狀態(combat package)或地圖/世界狀態,
// 且城鎮法術庫存需世界生成 RNG,兩者都不在本切片)。
type Spell struct {
	ID     int    // 索引(0..13),對齊 ini `[spellN]` 與 C 版 spells[MAX_SPELLS] 的位置
	Name   string // 繁中法術名(ini `name`)
	Type   string // ini 原始 `type` 字串(如 "clone"、"holy_damage"),對照 SPELL_* 動作碼語意見上方 spellCombatTypes 註解
	Gold   int    // 城鎮購買價(ini `gold`,對應 C WDAT_SCOST/spell_costs[])
	Damage int    // 法術威力/傷害(ini `damage`,對應 C WDAT_SDAMAGE/spell_powers[];部份法術此值語意是「回合數」等而非傷害,如 time_stop,以 C 版原始用法為準,未逐一區分)
	Filter string // 目標篩選(ini `filter`,目前只有 spell6 超渡術填 "undead",對應 C SFILTER_UNDEAD;其餘法術缺此 key)
	Combat bool   // true = 戰鬥法術(只能戰鬥中施放),false = 冒險法術(只能地圖上施放);依 Type 用 spellCombatTypes 判定
}

// IsCombat 回傳這是不是戰鬥法術,依 Type 對照 spellCombatTypes。
// 未知/無法辨識的 type 字串(理論上不該發生,ini 資料應與 C 版
// GNU_spell_downto_byte() 的 names[] 表完全對應)回 false,歸類為冒險法術
// 這一側 —— 因為 C 版原始行為是「Failed to parse」時該格法術動作維持
// 記憶體預設值 0(SPELL_DAMAGE,戰鬥法術),所以退回 false 較不會誤判成
// 戰鬥法術;呼叫端若在意應優先檢查 Type 是否為空字串。
func (s Spell) IsCombat() bool {
	return spellCombatTypes[s.Type]
}

// LoadSpells 從 a.Strings["spells"](spells.ini 解析結果)建出 14 個 Spell。
// 缺檔或缺區塊時該筆 Spell 只剩 ID(其餘欄位零值),不 panic —— 對齊
// kbdata.Assets 的 best-effort 慣例(見 kbdata/assets.go Load 註解)。
func LoadSpells(a *kbdata.Assets) []Spell {
	spells := make([]Spell, maxSpells)
	for i := range spells {
		spells[i].ID = i
	}
	if a == nil {
		return spells
	}
	ini := a.Strings["spells"]
	if ini == nil {
		return spells
	}
	for i := 0; i < maxSpells; i++ {
		sec, ok := ini.NumberedSection("spell", i)
		if !ok {
			continue
		}
		sp := &spells[i]
		sp.Name = sec.GetDefault("name", "")
		sp.Type = sec.GetDefault("type", "")
		sp.Gold = sec.IntDefault("gold", 0)
		sp.Damage = sec.IntDefault("damage", 0)
		sp.Filter = sec.GetDefault("filter", "")
		sp.Combat = sp.IsCombat()
	}
	return spells
}

// KnownSpells 回傳玩家目前已學會的法術總數,對應 C known_spells(play.c:661):
// 把 Spells[] 所有格加總(允許同一法術學會多次,逐格計數而非布林已學/未學)。
func (gs *GameState) KnownSpells() int {
	total := 0
	for _, c := range gs.Spells {
		total += c
	}
	return total
}

// LearnSpell 在城鎮購買/學習一個法術,對應 C visit_town() key==4 分支
// (game.c:2807-2827)。呼叫端(TownScreen)已用 gs.TownSpell[townID] 查出
// spellID、用 kbdata spells.ini 查出 cost,故這裡直接吃兩個已解析的值,不重複
// 解析資料 —— 對照組是 BuyTroop(army.go)吃 *kbdata.Assets 現查 troop cost,
// 兩者取用資料的方式不同是因為呼叫端(town.go)本來就已經持有這兩個值
// (NewTownScreen 建構時就查好 s.spell),沒有必要在 gamestate 側重複 LoadSpells。
//
// 判斷順序逐一對齊 C(不可調換):
//  1. known_spells(game) >= game->max_spells → ErrSpellLimitReached
//     (「你已學會上限數量的法術」)。
//  2. game->gold <= cost → ErrNotEnoughGold(「你的金幣不足」)。注意是 <=,
//     金幣剛好等於價格也算不足,需嚴格大於才能購買。
//  3. 否則:該法術 ++、扣錢,回 nil。
func (gs *GameState) LearnSpell(spellID, cost int) error {
	if gs.KnownSpells() >= gs.MaxSpells {
		return ErrSpellLimitReached
	}
	if gs.Gold <= cost {
		return ErrNotEnoughGold
	}
	if spellID >= 0 && spellID < len(gs.Spells) {
		gs.Spells[spellID]++
	}
	gs.Gold -= cost
	return nil
}
