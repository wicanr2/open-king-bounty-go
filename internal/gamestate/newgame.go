package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// emptySquad 是 C 版 0xFF 空格慣例(troop id = 255,count = 0)。
const emptySquad = 255

// NewGame 建立角色起手狀態,對應 C 版 src/play.c 的 spawn_game(name, pclass, difficulty, land)
// 當中「玩家角色起手數值」那部分(不含世界生成:惡棍/巢穴/寶物佈置等 RNG 相關邏輯不在本切片)。
//
// 難度固定為 easy/rank 0,對齊黃金樣本。C 版 spawn_game 的對應:
//   - game->gold          = starting_gold[pclass]                         → Gold
//   - game->rank          = 0; player_accept_rank(game)                   → Rank
//   - game->base_leadership 從 0 累加 classes[pclass][0].leadership       → BaseLeadership
//   - game->leadership    = game->base_leadership(建角當下兩者相等)      → Leadership
//   - game->commission    從 0 累加 classes[pclass][0].commission        → Commission
//   - game->spell_power   從 0 累加 classes[pclass][0].spell_power       → SpellPower
//   - game->max_spells    從 0 累加 classes[pclass][0].max_spell         → MaxSpells
//   - game->knows_magic   從 0 累加 classes[pclass][0].knows_magic       → KnowsMagic
//   - game->player_troops[0..1]/player_numbers[0..1] = starting_army_troop/numbers[pclass]
//   - game->player_troops[2..4] = 0xFF, player_numbers[2..4] = 0         → Army[2..4]
//
// 註:C 版 player_accept_rank 用 `+=` 對 rank 列累加(供之後升階複用),但建角當下
// KBgame 剛 memset 成 0 且只呼叫一次(rank=0),數值上等於直接取 classes[pclass][0] 的值。
// kbdata.Assets.Classes 已在資料層把逐階「+增量」換算成累積後的絕對值(見 assets.go /
// tables_classes.go 註解),所以這裡直接讀 classes[class][0] 即為起手絕對值,不需要再自己累加。
func NewGame(a *kbdata.Assets, name string, class int) *GameState {
	c := a.Classes[class][0]

	gs := &GameState{
		Name: name,

		Class: class,
		Rank:  0,

		Gold: a.StartGold[class],

		BaseLeadership: c.Leadership,
		Leadership:     c.Leadership,
		Commission:     c.Commission,

		SpellPower: c.SpellPower,
		MaxSpells:  c.MaxSpell,
		KnowsMagic: c.KnowsMagic != 0,
	}

	for i := 0; i < 2; i++ {
		gs.Army[i] = Squad{
			TroopID: a.StartArmyTroop[class][i],
			Count:   a.StartArmyCount[class][i],
		}
	}
	for i := 2; i < 5; i++ {
		gs.Army[i] = Squad{TroopID: emptySquad, Count: 0}
	}

	return gs
}
