// tables_villains.go -- 惡棍(villain)世界生成用靜態表,移植自 C 版 src/bounty.c。
//
// 本檔只搬資料,不搬演算法(salt_villains/repopulate_castle 邏輯留給
// gamestate/castlegen.go)。逐值照抄 C 版字面(含 0x 十六進位原樣),不做任何轉換。
//
// ⚠ 資料來源核實(rulebook 62 靜態溯源,踩過 town_coords/special_coords 被 free
// 模組 *.ini 覆寫的坑,故本次移植前先逐欄核對 free 真值,見下方兩點):
//   - villain_army_troops/villain_army_numbers:C 版 refill_rules()(game.c:273-274)
//     會用 DAT_VTROOP/WDAT_VNUMBER 覆寫,free 模組實際來源是 data/free/villains.ini
//     的 armyN 欄("N x 兵種名"字串,經 troops.ini 名稱表轉回 troop id)。**逐筆核對
//     villains.ini 與本檔字面值(2026-07-10 手動核對全部 17 筆兵種 id 與數量),
//     兩者完全一致**——villains.ini 的army 資料只是把 bounty.c 同一批數字换成人類
//     可讀的中文兵種名重新表達,數值本身未被 free 模組更動。故本檔直接照抄 bounty.c
//     字面即可,不需要在執行期解析 villains.ini(kbdata.FreeIni + ParseArmyEntry
//     已備妥,如日後要走 ini 解析路徑可重用)。
//   - villains_per_continent:C 版**沒有**對應的 DAT_* 覆寫鍵(kbres.h/free-data.c
//     搜尋確認),永遠讀 bounty.c 硬編值,不受模組影響。
//
// ⚠ 城堡座標(castle_coords)**不在本檔**——那組資料**會**被 free 模組覆寫
// (DAT_CASTLEX/Y/C,且 free 版 castles.ini 的實際座標與 bounty.c 硬編值不同,
// 因為 free 用自製 land.tmx 地圖、非 DOS 原版佈局),故 castle 的 continent/x/y
// 改在 gamestate/castle.go 用 kbdata.FreeIni 解析 free/castles.ini(對齊既有
// gamestate/town.go 讀 towns.ini 的做法),不可比照本檔字面照抄。castle_difficulty
// 則反過來——free 模組**沒有** DAT_CASTLEDIFF 覆寫鍵,castles.ini 裡的 "difficulty"
// 欄位是 land.tmx 匯出時的殘留欄位、C 引擎從未讀取,故 castle_difficulty 仍是
// kbdata.CastleDifficultyTable()(tables_worldgen.go,bounty.c:612)照抄不受影響。
package kbdata

// MaxVillains 對齊 C 版 bounty.h:46 `#define MAX_VILLAINS 17`。
const MaxVillains = 17

// VillainsPerContinentTable 回傳每洲要放幾個惡棍,
// 移植自 C 版 villains_per_continent[MAX_CONTINENTS](bounty.c:220)。
// 索引 = 洲別(0..3),不受 free 模組覆寫(無對應 DAT_* 鍵,見檔頭核實記錄)。
func VillainsPerContinentTable() [MaxContinents]byte {
	return [MaxContinents]byte{6, 4, 4, 3}
}

// VillainArmyTroopsTable 回傳每個惡棍的 5 格守軍兵種 id,
// 移植自 C 版 villain_army_troops[MAX_VILLAINS][5](bounty.c:247)。
// 索引 = 惡棍 id(0..16);與 free 模組 villains.ini 的 armyN 欄逐筆核對一致
// (見檔頭核實記錄),故直接照抄 bounty.c 字面。
func VillainArmyTroopsTable() [MaxVillains][5]byte {
	return [MaxVillains][5]byte{
		{0x00, 0x03, 0x02, 0x00, 0x00}, // 0
		{0x0b, 0x02, 0x02, 0x00, 0x00}, // 1
		{0x01, 0x01, 0x04, 0x05, 0x0f}, // 2
		{0x03, 0x07, 0x08, 0x11, 0x0c}, // 3
		{0x02, 0x02, 0x08, 0x09, 0x10}, // 4
		{0x01, 0x0d, 0x0e, 0x14, 0x14}, // 5
		{0x02, 0x08, 0x0a, 0x12, 0x0e}, // 6
		{0x09, 0x14, 0x13, 0x0a, 0x01}, // 7
		{0x07, 0x0f, 0x11, 0x16, 0x03}, // 8
		{0x04, 0x05, 0x0d, 0x15, 0x17}, // 9
		{0x04, 0x05, 0x0d, 0x15, 0x17}, // 10
		{0x16, 0x18, 0x0f, 0x07, 0x06}, // 11
		{0x06, 0x10, 0x16, 0x0b, 0x00}, // 12
		{0x08, 0x0a, 0x12, 0x0e, 0x18}, // 13
		{0x17, 0x15, 0x14, 0x06, 0x00}, // 14
		{0x17, 0x18, 0x12, 0x0e, 0x14}, // 15
		{0x18, 0x18, 0x18, 0x17, 0x15}, // 16
	}
}

// VillainArmyNumbersTable 回傳每個惡棍的 5 格守軍數量,
// 移植自 C 版 villain_army_numbers[MAX_VILLAINS][5](word,bounty.c:266)。
// 索引 = 惡棍 id(0..16);與 free 模組 villains.ini 的 armyN 欄逐筆核對一致
// (見檔頭核實記錄),故直接照抄 bounty.c 字面。
func VillainArmyNumbersTable() [MaxVillains][5]int {
	return [MaxVillains][5]int{
		{50, 20, 25, 30, 25},     // 0
		{10, 30, 20, 60, 40},     // 1
		{70, 50, 20, 20, 4},      // 2
		{30, 20, 10, 2, 6},       // 3
		{50, 50, 10, 10, 5},      // 4
		{250, 10, 10, 4, 4},      // 5
		{100, 20, 20, 15, 15},    // 6
		{30, 30, 10, 30, 300},    // 7
		{150, 20, 10, 5, 80},     // 8
		{500, 100, 30, 10, 6},    // 9
		{600, 200, 50, 25, 10},   // 10
		{30, 5, 30, 200, 200},    // 11
		{300, 40, 20, 100, 700},  // 12
		{35, 100, 80, 60, 5},     // 13
		{30, 50, 100, 500, 5000}, // 14
		{50, 10, 200, 250, 60},   // 15
		{100, 25, 25, 100, 100},  // 16
	}
}
