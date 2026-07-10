// tables_worldgen.go -- 世界生成用靜態表,移植自 C 版 src/bounty.c。
//
// 這些表是 salt_continent / populate_dwelling / repopulate_foe(尚未移植,見下一步)
// 用來決定「城堡難度」「棲地該長哪種兵、多少隻」「哪個洲有哪些棲地」的查表資料,
// 本檔**只搬資料,不搬演算法**——邏輯留給下一步的 worldgen 移植。
//
// 逐值照抄 C 版字面(含 0x 十六進位原樣),不做任何轉換或重排。
package kbdata

// 世界生成表的維度常數,對齊 src/bounty.h。
// MaxContinents(=4, MAX_CONTINENTS)已在 land.go 定義,這裡沿用不重複宣告。
const (
	MaxCastles          = 26 // MAX_CASTLES            (bounty.h:50)
	MaxDwellings        = 11 // MAX_DWELLINGS          (bounty.h:49)
	MaxTroops           = 25 // MAX_TROOPS             (bounty.h:185,與 tables_troops.go 的 25 筆一致)
	MaxTroopDifficulty  = 4  // MAX_TROOP_DIFFICULTY   (bounty.h:55)
	MaxTroopChanceCurve = 5  // MAX_TROOP_CHANCE_CURVE (bounty.h:56)
)

// castleDifficultyTable 回傳每座城堡的難度等級(0~3),
// 移植自 C 版 castle_difficulty[MAX_CASTLES](bounty.c:612)。
// 索引 = 城堡 id,語意與 tables_classes.go 的城堡座標表(castle_coords)索引一致。
func castleDifficultyTable() [MaxCastles]int {
	return [MaxCastles]int{
		0, 1, 0, 1, 2, 0, 2, 2, 0, 1,
		0, 2, 1, 0, 0, 0, 1, 0, 3, 2,
		3, 0, 0, 2, 1, 3,
	}
}

// troopChanceTableTable 回傳「撒鹽」用的機率累積曲線,
// 移植自 C 版 troop_chance_table[MAX_TROOP_DIFFICULTY][MAX_TROOP_CHANCE_CURVE-1](bounty.c:764)。
// 每列 4 個門檻值(0~0xFF 累積機率刻度),列 = 洲難度(0=易 ~ 3=難)。
func troopChanceTableTable() [MaxTroopDifficulty][MaxTroopChanceCurve - 1]byte {
	return [MaxTroopDifficulty][MaxTroopChanceCurve - 1]byte{
		{0x3C, 0x5A, 0x62, 0x65},
		{0x14, 0x46, 0x5F, 0x63},
		{0x0A, 0x14, 0x32, 0x5A},
		{0x03, 0x06, 0x0A, 0x28},
	}
}

// dwellingToTroopTable 回傳「棲地難度 → 兵種 id」對照表,
// 移植自 C 版 dwelling_to_troop[MAX_TROOP_DIFFICULTY][MAX_TROOP_CHANCE_CURVE](bounty.c:770)。
// 列 = 棲地類型(依 C 版原始註解排序):0=平原(Plains) 1=森林(Forest) 2=丘陵洞穴(Hill) 3=地城(Dungeon)。
// 欄 = 該棲地類型底下,依 troopChanceTable 曲線挑出的 5 個兵種 id(由弱到強)。
func dwellingToTroopTable() [MaxTroopDifficulty][MaxTroopChanceCurve]byte {
	return [MaxTroopDifficulty][MaxTroopChanceCurve]byte{
		{0x00, 0x03, 0x0b, 0x10, 0x14}, // Plains
		{0x01, 0x06, 0x09, 0x11, 0x13}, // Forest
		{0x07, 0x0c, 0x0f, 0x16, 0x18}, // Hill
		{0x04, 0x05, 0x0d, 0x15, 0x17}, // Dungeon
	}
}

// troopNumbersTable 回傳每個兵種在各難度下的棲地生成隻數,
// 移植自 C 版 troop_numbers[MAX_TROOPS][MAX_TROOP_DIFFICULTY](bounty.c:780)。
// 列 = 兵種 id(0~24,對齊 tables_troops.go 的 troopTable() 順序),欄 = 難度(0~3)。
// 全 0 的列(id 8/10/14/18,對照 tables_troops.go 為弓箭手/槍兵/武士/騎兵,均 Home==DwellCastle)
// 代表該兵種不會出現在野外棲地生成;但同為城堡兵種的 id 2(義勇軍)troop_numbers 非零——
// 這是 C 原始資料本身如此,照抄不推斷/不「修正」。
func troopNumbersTable() [MaxTroops][MaxTroopDifficulty]byte {
	return [MaxTroops][MaxTroopDifficulty]byte{
		{0x0A, 0x14, 0x32, 0x64},
		{0x14, 0x32, 0x64, 0x7F},
		{0x0A, 0x14, 0x32, 0x64},
		{0x05, 0x0F, 0x1E, 0x50},
		{0x05, 0x0A, 0x19, 0x32},
		{0x05, 0x0A, 0x19, 0x4B},
		{0x0A, 0x19, 0x32, 0x64},
		{0x05, 0x0F, 0x1E, 0x50},
		{0x00, 0x00, 0x00, 0x00},
		{0x05, 0x0A, 0x19, 0x32},
		{0x00, 0x00, 0x00, 0x00},
		{0x04, 0x08, 0x0F, 0x1E},
		{0x04, 0x0A, 0x14, 0x32},
		{0x02, 0x04, 0x0A, 0x14},
		{0x00, 0x00, 0x00, 0x00},
		{0x02, 0x04, 0x08, 0x0F},
		{0x02, 0x04, 0x0A, 0x14},
		{0x02, 0x04, 0x08, 0x0F},
		{0x00, 0x00, 0x00, 0x00},
		{0x01, 0x03, 0x06, 0x0A},
		{0x01, 0x02, 0x04, 0x08},
		{0x02, 0x04, 0x0A, 0x19},
		{0x01, 0x02, 0x05, 0x0A},
		{0x01, 0x02, 0x04, 0x08},
		{0x01, 0x01, 0x01, 0x02},
	}
}

// continentDwellingsTable 回傳每個洲擁有哪些棲地兵種 id,
// 移植自 C 版 continent_dwellings[MAX_CONTINENTS][MAX_DWELLINGS](bounty.c:808)。
//
// C 版原始初始化只給每列前幾格顯式值(含 0xFF 結尾標記),
// 其餘欄位在 C 靜態陣列語意下自動零填(static initializer 未列到的元素 = 0)。
// 本表逐格照抄含零填部分,不可只搬「看得到的」前段值,否則會少填 0 而值不對齊。
func continentDwellingsTable() [MaxContinents][MaxDwellings]byte {
	return [MaxContinents][MaxDwellings]byte{
		{0x00, 0x01, 0x07, 0x04, 0x03, 0x06, 0xFF, 0x00, 0x00, 0x00, 0x00},
		{0x0C, 0x05, 0x0B, 0x09, 0x0F, 0x09, 0xFF, 0x00, 0x00, 0x00, 0x00},
		{0x0D, 0x10, 0x11, 0x13, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		{0x16, 0x15, 0x14, 0x18, 0x17, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
}

// dwellingRangesTable 回傳每個洲的兵種難度區間 [下限,上限](兵種 id),
// 移植自 C 版 dwelling_ranges[MAX_CONTINENTS][2](bounty.c:814)。
// 註解直接沿用 C 版原文,標明區間對應的兵種段(peasant-knight 等)。
func dwellingRangesTable() [MaxContinents][2]byte {
	return [MaxContinents][2]byte{
		{0x00, 0x0e}, // (peasant-knight)
		{0x01, 0x0e}, // (sprites-knight)
		{0x02, 0x0e}, // (militia-knight)
		{0x14, 0x18}, // (archmag-dragon)
	}
}
