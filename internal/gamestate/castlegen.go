package gamestate

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// castlegen.go 移植 C 版 src/play.c 的城堡/惡棍世界生成邏輯:salt_villains、
// repopulate_castle。這兩支接在 worldgen.go 的 saltContinent 之後執行(對齊
// spawn_game 呼叫順序,見 newgame.go),把「未分配」城堡填上惡棍守軍或怪物守軍。
//
// 逐句對照 play.c,不精簡、不改順序、不「修正」C 原始的資料/控制流怪癖。

// castleOwnerMonsters 對齊 C 版 0x7F(bounty.h:136 註解:
// "0x7F = no one, 0xFF = you, LOW 5 bits = villain")。spawn_game 先把全部城堡
// 初始化成這個值,再讓 salt_villains 佔走一部分,剩下的仍是 0x7F、由
// repopulate_castle 填怪物守軍。
const castleOwnerMonsters = 0x7F

// saltVillains 對齊 C play.c:158 salt_villains(game, continent, base_id):
// 把該洲 villains_per_continent[continent] 個惡棍隨機塞進「該洲且未被佔領」
// (castle_owner == 0x7F)的城堡,設定 castle_owner/troops/numbers。
// 回傳 base_id + 實際放入數(呼叫端串接四洲,累加惡棍 id)。
//
// 忠實保留 C 的安全閥:若該洲城堡數不夠塞(numCastles(continent) <
// villains_per_continent[continent]),C 版 KB_errlog 後直接 return base_id
// (該洲一個惡棍都不放)——GameState 沒有 KB_errlog 對應輸出,故只 return,
// 不 log,呼叫端自行留意。
//
// 迴圈本身對齊 C 的「未命中就重抽,直到放滿」寫法(`for (j=0;...;)`,內層才
// j++):只要「該洲存在足夠未佔領城堡」這個前提成立(由上面的數量檢查保證,
// 因為呼叫時全部 26 座城堡都還是 0x7F),必然會終止,不會無窮迴圈。
func (gs *GameState) saltVillains(rng kbrng.Rand, castles []Castle, continent, baseID int) int {
	perContinent := kbdata.VillainsPerContinentTable()
	need := int(perContinent[continent])

	max := numCastles(castles, continent)
	if max < need {
		return baseID
	}

	troops := kbdata.VillainArmyTroopsTable()
	numbers := kbdata.VillainArmyNumbersTable()

	j := 0
	for j < need {
		villainID := baseID + j
		castleID := rng.Between(0, kbdata.MaxCastles-1)
		// 該城堡在此洲、且尚未被佔領。
		if castles[castleID].Continent == continent && gs.CastleOwner[castleID] == castleOwnerMonsters {
			gs.CastleOwner[castleID] = byte(villainID)
			for i := 0; i < 5; i++ {
				gs.CastleTroops[castleID][i] = int(troops[villainID][i])
				gs.CastleNumbers[castleID][i] = int(numbers[villainID][i])
			}
			j++
		}
	}
	return baseID + j
}

// repopulateCastle 對齊 C play.c:77 repopulate_castle(game, castle_id):
// 用 rollCreature(castle_difficulty[castle_id]) 填滿該城堡 5 格守軍(怪物,
// 非惡棍)。呼叫端只對 salt_villains 之後仍是 castleOwnerMonsters 的城堡呼叫
// (對齊 spawn_game:「Populate all castles owned by monsters」)。
func (gs *GameState) repopulateCastle(rng kbrng.Rand, castles []Castle, castleID int) {
	difficulty := castles[castleID].Difficulty
	for i := 0; i < 5; i++ {
		troopID, count := rollCreature(rng, difficulty)
		gs.CastleTroops[castleID][i] = troopID
		gs.CastleNumbers[castleID][i] = count
	}
}
