package gamestate

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// worldgen.go 移植 C 版 src/play.c 的世界生成邏輯:roll_creature、populate_dwelling
// (+enforce_dwelling)、repopulate_foe、salt_continent。這些函式在 spawn_game 建局時
// 把「salt 前」的原始地圖(chest tile 0x8B / foe tile 0x91,見 kbdata land.go 註解)
// 轉成真正的棲地/敵人/寶物佈置,填進 GameState 的世界狀態欄位(見 gamestate.go)。
//
// 逐句對照 play.c,不精簡、不改順序、不「修正」C 原始的資料/控制流怪癖——
// 這些怪癖(見下方個別函式註解)照抄,並在發現處誠實標註,不擅自「修好」。

// rollCreature 對齊 C play.c:28 roll_creature(difficulty, &id, &number)。
//
// rng 呼叫順序刻意對齊 C(先 dwelling 再 chance),即使目前 NewGame 尚未逐一重現
// spawn_game 完整 rand 序列(見 newgame.go parity 註記),仍維持這個內部順序忠實,
// 為未來要對 seed 時少一個要修的地方。
func rollCreature(rng kbrng.Rand, difficulty int) (troopID, number int) {
	dwelling := rng.Between(0, 3)
	chance := rng.Between(1, 100)

	chanceTable := kbdata.TroopChanceTable()
	dwellToTroop := kbdata.DwellingToTroopTable()
	numbers := kbdata.TroopNumbersTable()

	index := 0
	for chance > int(chanceTable[difficulty][index]) {
		index++
		// 對齊 C `if (index >= MAX_TROOP_CHANCE_CURVE-1) break;`:安全閥,避免
		// index 再往上會越界存取 chanceTable(每列只有 MaxTroopChanceCurve-1 欄)。
		if index >= kbdata.MaxTroopChanceCurve-1 {
			break
		}
	}

	troopID = int(dwellToTroop[dwelling][index])
	number = int(numbers[troopID][difficulty])
	if number <= 1 {
		number = 2 // 對齊 C `if (troop_count <= 1) troop_count = 2;`
	}
	return troopID, number
}

// enforceDwelling 對齊 C play.c:50 enforce_dwelling:把棲地兵種與庫存寫進世界狀態,
// 庫存 = troops[troop_id].max_population(對齊修正後的 kbdata.Troop.MaxPopulation 欄,
// 見 kbdata/assets.go 的欄位勘誤註記)。
func (gs *GameState) enforceDwelling(a *kbdata.Assets, continent, dwellingID, troopID int) {
	gs.DwellingTroop[continent][dwellingID] = troopID
	pop := 0
	if a != nil && troopID >= 0 && troopID < len(a.Troops) {
		pop = a.Troops[troopID].MaxPopulation
	}
	gs.DwellingPopulation[continent][dwellingID] = pop
}

// populateDwelling 對齊 C play.c:56 populate_dwelling:決定棲地兵種(有 continent_dwellings
// 固定表就用固定值,否則在 dwelling_ranges 區間內隨機 roll),寫入世界狀態,
// 回傳 troops[troop_id].dwells(棲地型別 0..3,呼叫端用來算地圖 tile = 0x8C+回傳值)。
func (gs *GameState) populateDwelling(a *kbdata.Assets, rng kbrng.Rand, continent, dwellingID int) int {
	continentDwellings := kbdata.ContinentDwellingsTable()
	dwellingRanges := kbdata.DwellingRangesTable()

	var troopID int
	if dwellingID >= kbdata.MaxDwellings || continentDwellings[continent][dwellingID] == 0xFF {
		troopID = rng.Between(int(dwellingRanges[continent][0]), int(dwellingRanges[continent][1]))
	} else {
		troopID = int(continentDwellings[continent][dwellingID])
	}

	gs.enforceDwelling(a, continent, dwellingID, troopID)

	if a != nil && troopID >= 0 && troopID < len(a.Troops) {
		return int(a.Troops[troopID].Home)
	}
	return 0
}

// repopulateFoe 對齊 C play.c:89 repopulate_foe:foe 群固定 3 格(C 迴圈 `i<3`,
// 雖然陣列宣告是 5 格——第 3、4 格刻意留零值,不是漏寫,照抄此怪癖),
// 每格各自 rollCreature 一次。
//
// 省略 C 的 `if (game->opt_foe_strength) troop_count *= 2;`(issue #9 P3 選項):
// GameState 目前沒有 opt_foe_strength 這個玩家選項欄位(見 gamestate.go,options
// 世界狀態尚未移植),沒有來源可讀,故誠實省略,不擅自預設開/關。之後補上該選項時
// 在此加一行即可。
func (gs *GameState) repopulateFoe(continent, foeID int, rng kbrng.Rand) {
	for i := 0; i < 3; i++ {
		troopID, count := rollCreature(rng, continent)
		gs.FoeTroops[continent][foeID][i] = troopID
		gs.FoeNumbers[continent][foeID][i] = count
	}
}

// salt 桶內容列舉,對齊 C play.c:189-197 salt_continent 內的匿名 enum SALT_*。
const (
	saltNone = iota
	saltArtifact
	saltNavmap
	saltOrb
	saltTelecave
	saltDwelling
	saltFriendly
)

// saltContinent 對齊 C play.c:186 salt_continent,逐句移植:
//
//  1. 若地圖 [0][0] 是起點標記 0xFF,換成深水(對齊 C 的兩行防呆)。
//  2. 桶子(barrel)長度 = 該洲 chest tile(0x8B)數;若不夠塞 min 總數就放棄(errlog + return,
//     這裡對應改成直接 return,不 log——GameState 沒有 KB_errlog 對應輸出,呼叫端自行留意)。
//  3. artifact/navmap/orb 三種用「抽中未占用格才算數」的迴圈(guaranteed 剛好放滿 min 個);
//     telecave/dwelling/friendly 三種用「抽 min 次,抽中已占用格就浪費掉那次」的迴圈
//     ——**這是 C 原始碼本身的不對稱行為,不是筆誤**(C: `for(i=0;i<min;i++)` vs
//     `for(i=0;i<min;)` 内層才 i++),照抄,可能導致實際放置數 < min(尤其地圖 chest
//     很少時)。
//  4. 掃描整張地圖:遇到 foe tile(0x91)登記座標+repopulateFoe;遇到 chest tile(0x8B)
//     依桶子內容轉成 artifact/navmap/orb/telecave/dwelling/friendly 對應的 tile 值。
//
// 忠實保留的「死碼」:C 版在 `barrel_index >= barrel_len` 分支裡多一層
// `else if (game->map[continent][j][i] == 0x8C)`——但這段程式碼位於外層
// `else if (game->map[continent][j][i] == 0x8B)` 之內,執行到這裡時 tile 必然還是
// 0x8B(尚未被本次迭代改寫),不可能等於 0x8C,是無法到達的分支(逆向溯源驗證,
// rulebook 62「描述前先驗證它真的成立」)。且 barrel_len 恰等於整個掃描會遇到的
// chest tile 總數,barrel_index 不會在掃描結束前超過 barrel_len,所以連
// `barrel_index >= barrel_len` 這個外層分支本身在實務上都不會被觸發。故 Go 版
// 略去這段對應邏輯,只留這行說明,不引入死碼。
//
// 「friendly」桶(saltFriendly)只把 chest tile 轉成 foe tile(0x91)並登記座標,
// **不呼叫 repopulateFoe**——對齊 C 原始碼把 `populate_foe(game, continent, ??)` 整行
// 註解掉(C 原文保留該行註解在 salt_continent 內,見 play.c 附近),foe_troops/numbers
// 因此維持 KBgame memset(0) 後的全零值。這是 C 版本身未完成的功能(踩到這種
// 「friendly」foe 理論上應觸發 game.c attack_foe 的 `id < FRIENDLY_FOES` 分支走
// accept_foe「免費入隊」流程,而非戰鬥),本次世界生成移植只還原資料佈置,
// accept_foe 的入隊流程不在本任務範圍(見呼叫端 screen/worldmap.go 的對應註記)。
func (gs *GameState) saltContinent(a *kbdata.Assets, rng kbrng.Rand, continent, minArtifacts, minNavmaps, minOrbs, minTelecaves, minDwellings, minFriendly int) {
	if gs.WorldMap == nil || continent < 0 || continent >= len(gs.WorldMap.Tiles) {
		return
	}
	grid := &gs.WorldMap.Tiles[continent]

	minLen := minArtifacts + minNavmaps + minOrbs + minTelecaves + minDwellings + minFriendly

	// 起點標記(對齊 C play.c:213)。
	if grid[0][0] == 0xFF {
		grid[0][0] = kbdata.TileDeepWater
	}

	barrelLen := 0
	for y := 0; y < kbdata.LevelH; y++ {
		for x := 0; x < kbdata.LevelW; x++ {
			if grid[y][x] == kbdata.TileChest {
				barrelLen++
			}
		}
	}
	if barrelLen < minLen {
		return // 對齊 C errlog + return:桶子不夠塞,放棄撒鹽這個洲
	}

	barrel := make([]int, barrelLen)

	for i := 0; i < minArtifacts; {
		idx := rng.Between(0, barrelLen-1)
		if barrel[idx] == saltNone {
			barrel[idx] = saltArtifact
			i++
		}
	}
	for i := 0; i < minNavmaps; {
		idx := rng.Between(0, barrelLen-1)
		if barrel[idx] == saltNone {
			barrel[idx] = saltNavmap
			i++
		}
	}
	for i := 0; i < minOrbs; {
		idx := rng.Between(0, barrelLen-1)
		if barrel[idx] == saltNone {
			barrel[idx] = saltOrb
			i++
		}
	}
	// 以下三種迴圈刻意用 for i:=0;i<min;i++(單次嘗試,不保證命中空格),
	// 對齊 C 的不對稱行為,見函式註解。
	for i := 0; i < minTelecaves; i++ {
		idx := rng.Between(0, barrelLen-1)
		if barrel[idx] == saltNone {
			barrel[idx] = saltTelecave
		}
	}
	for i := 0; i < minDwellings; i++ {
		idx := rng.Between(0, barrelLen-1)
		if barrel[idx] == saltNone {
			barrel[idx] = saltDwelling
		}
	}
	for i := 0; i < minFriendly; i++ {
		idx := rng.Between(0, barrelLen-1)
		if barrel[idx] == saltNone {
			barrel[idx] = saltFriendly
		}
	}

	numFoe := minFriendly // 對齊 C `num_foe = min_friendly;`——一般 foe 的索引從這裡開始
	numArtifacts, numDwellings, numNavmaps, numOrbs, numTelecaves, numFriendly := 0, 0, 0, 0, 0, 0
	barrelIndex := 0

	for y := 0; y < kbdata.LevelH; y++ {
		for x := 0; x < kbdata.LevelW; x++ {
			switch grid[y][x] {
			case kbdata.TileFoe:
				gs.FoeCoords[continent][numFoe][0] = x
				gs.FoeCoords[continent][numFoe][1] = y
				gs.repopulateFoe(continent, numFoe, rng)
				numFoe++

			case kbdata.TileChest:
				if barrelIndex >= barrelLen {
					// 對齊 C 的 else 分支(含前述已驗證為死碼的 0x8C 檢查):
					// 到這裡代表桶子已用罄,tile 維持原樣不動(仍是 chest)。
					continue
				}
				switch barrel[barrelIndex] {
				case saltArtifact:
					grid[y][x] = kbdata.TileArtifact1 + byte(numArtifacts)
					numArtifacts++
				case saltOrb:
					gs.OrbCoords[continent][0] = x
					gs.OrbCoords[continent][1] = y
					numOrbs++
				case saltNavmap:
					gs.NavmapCoords[continent][0] = x
					gs.NavmapCoords[continent][1] = y
					numNavmaps++
				case saltTelecave:
					grid[y][x] = kbdata.TileTelecave
					gs.TelecaveCoords[continent][numTelecaves][0] = x
					gs.TelecaveCoords[continent][numTelecaves][1] = y
					numTelecaves++
				case saltDwelling:
					grid[y][x] = kbdata.TileDwelling1 + byte(gs.populateDwelling(a, rng, continent, numDwellings))
					gs.DwellingCoords[continent][numDwellings][0] = x
					gs.DwellingCoords[continent][numDwellings][1] = y
					numDwellings++
				case saltFriendly:
					grid[y][x] = kbdata.TileFoe
					gs.FoeCoords[continent][numFriendly][0] = x
					gs.FoeCoords[continent][numFriendly][1] = y
					// 刻意不呼叫 repopulateFoe,見函式註解(對齊 C populate_foe 呼叫被註解掉)。
					numFriendly++
				case saltNone:
					// 對齊 C `default: /* Do nothing (leave as chest) */`
				}
				barrelIndex++
			}
		}
	}
}
