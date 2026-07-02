package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// emptySquad 是 C 版 0xFF 空格慣例(troop id = 255,count = 0)。
const emptySquad = 255

// NewGame 建立角色起手狀態,對應 C 版 src/play.c 的 spawn_game 中「玩家角色起手數值」那部分
// (不含世界生成:惡棍/巢穴/寶物佈置等 RNG 邏輯不在此)。難度固定 easy/rank 0。
//
// 忠實對齊 spawn_game:先 Rank=0,再呼叫 acceptRank(等同 C 的 player_accept_rank),
// 由它把 classes[class][0] 的**增量**累加進 BaseLeadership/MaxSpells/SpellPower/Commission/
// KnowsMagic(與升階共用同一條路徑);Leadership 起手 = BaseLeadership。
//   - Gold                = starting_gold[class]
//   - player_troops/numbers[0..1] = starting_army_troop/numbers[class];[2..4] = 0xFF/0
func NewGame(a *kbdata.Assets, name string, class int) *GameState {
	gs := &GameState{
		Name:  name,
		Class: class,
		Rank:  0,
		Gold:  a.StartGold[class],
	}

	// 套用 rank 0 增量(base 從 0 起),再讓當前可用領導力對齊 base。
	gs.acceptRank(a)
	gs.Leadership = gs.BaseLeadership

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
