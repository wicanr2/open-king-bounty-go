package kbdata

// troopTable 回傳全兵種靜態表,移植自 C 版 src/bounty.c 的 troops[MAX_TROOPS]。
//
// 25 筆兵種,對照 C 版 troops[] 初始化資料與 KBtroop 結構(src/bounty.h)。
// 最後四欄逐字對應 C 版順序:(dwells, max_population, growth, morale_group);
// 詳見 assets.go Troop 型別定義下方的欄位勘誤註記。
func troopTable() []Troop {
	return []Troop{
		// 0: 農夫
		{Name: "農夫", SkillLevel: 1, HP: 1, Move: 1, MeleeMin: 1, MeleeMax: 1,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 10, Spoils: 1, Abilities: 0,
			Home: DwellPlains, MaxPopulation: 250, Growth: 6, MoraleGroup: 0},
		// 1: 小妖精
		{Name: "小妖精", SkillLevel: 1, HP: 1, Move: 1, MeleeMin: 1, MeleeMax: 2,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 15, Spoils: 1, Abilities: AbilFly,
			Home: DwellForest, MaxPopulation: 200, Growth: 6, MoraleGroup: 2},
		// 2: 義勇軍
		{Name: "義勇軍", SkillLevel: 2, HP: 2, Move: 2, MeleeMin: 1, MeleeMax: 2,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 50, Spoils: 5, Abilities: 0,
			Home: DwellCastle, MaxPopulation: 0, Growth: 5, MoraleGroup: 0},
		// 3: 野狼
		{Name: "野狼", SkillLevel: 2, HP: 3, Move: 3, MeleeMin: 1, MeleeMax: 3,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 40, Spoils: 4, Abilities: 0,
			Home: DwellPlains, MaxPopulation: 150, Growth: 5, MoraleGroup: 3},
		// 4: 骷髏兵
		{Name: "骷髏兵", SkillLevel: 2, HP: 3, Move: 2, MeleeMin: 1, MeleeMax: 2,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 40, Spoils: 4, Abilities: AbilUndead,
			Home: DwellDungeon, MaxPopulation: 150, Growth: 5, MoraleGroup: 4},
		// 5: 僵屍
		{Name: "僵屍", SkillLevel: 2, HP: 5, Move: 1, MeleeMin: 2, MeleeMax: 2,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 50, Spoils: 5, Abilities: AbilUndead,
			Home: DwellDungeon, MaxPopulation: 100, Growth: 5, MoraleGroup: 4},
		// 6: 地精
		{Name: "地精", SkillLevel: 2, HP: 5, Move: 1, MeleeMin: 1, MeleeMax: 3,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 60, Spoils: 6, Abilities: 0,
			Home: DwellForest, MaxPopulation: 250, Growth: 5, MoraleGroup: 2},
		// 7: 牛獸人
		{Name: "牛獸人", SkillLevel: 2, HP: 5, Move: 2, MeleeMin: 2, MeleeMax: 3,
			RangedMin: 1, RangedMax: 2, RangedShots: 10, GoldCost: 75, Spoils: 7, Abilities: 0,
			Home: DwellHillCave, MaxPopulation: 200, Growth: 5, MoraleGroup: 3},
		// 8: 弓箭手
		{Name: "弓箭手", SkillLevel: 2, HP: 10, Move: 2, MeleeMin: 1, MeleeMax: 2,
			RangedMin: 1, RangedMax: 3, RangedShots: 12, GoldCost: 250, Spoils: 25, Abilities: 0,
			Home: DwellCastle, MaxPopulation: 0, Growth: 5, MoraleGroup: 1},
		// 9: 精靈族
		{Name: "精靈族", SkillLevel: 3, HP: 10, Move: 3, MeleeMin: 1, MeleeMax: 2,
			RangedMin: 2, RangedMax: 4, RangedShots: 24, GoldCost: 200, Spoils: 20, Abilities: 0,
			Home: DwellForest, MaxPopulation: 100, Growth: 4, MoraleGroup: 2},
		// 10: 槍兵
		{Name: "槍兵", SkillLevel: 3, HP: 10, Move: 2, MeleeMin: 2, MeleeMax: 4,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 800, Spoils: 30, Abilities: 0,
			Home: DwellCastle, MaxPopulation: 0, Growth: 4, MoraleGroup: 1},
		// 11: 流浪者
		{Name: "流浪者", SkillLevel: 3, HP: 15, Move: 2, MeleeMin: 2, MeleeMax: 4,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 300, Spoils: 30, Abilities: 0,
			Home: DwellPlains, MaxPopulation: 150, Growth: 4, MoraleGroup: 2},
		// 12: 矮人
		{Name: "矮人", SkillLevel: 3, HP: 20, Move: 1, MeleeMin: 2, MeleeMax: 4,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 350, Spoils: 30, Abilities: 0,
			Home: DwellHillCave, MaxPopulation: 100, Growth: 4, MoraleGroup: 2},
		// 13: 鬼
		{Name: "鬼", SkillLevel: 4, HP: 10, Move: 3, MeleeMin: 3, MeleeMax: 4,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 400, Spoils: 40, Abilities: AbilAbsorb | AbilUndead,
			Home: DwellDungeon, MaxPopulation: 25, Growth: 3, MoraleGroup: 4},
		// 14: 武士
		{Name: "武士", SkillLevel: 5, HP: 35, Move: 1, MeleeMin: 6, MeleeMax: 10,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 1000, Spoils: 100, Abilities: 0,
			Home: DwellCastle, MaxPopulation: 250, Growth: 3, MoraleGroup: 1},
		// 15: 食人怪
		{Name: "食人怪", SkillLevel: 4, HP: 40, Move: 1, MeleeMin: 3, MeleeMax: 5,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 750, Spoils: 75, Abilities: 0,
			Home: DwellHillCave, MaxPopulation: 200, Growth: 3, MoraleGroup: 3},
		// 16: 野人
		{Name: "野人", SkillLevel: 4, HP: 40, Move: 3, MeleeMin: 1, MeleeMax: 6,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 750, Spoils: 75, Abilities: 0,
			Home: DwellPlains, MaxPopulation: 100, Growth: 3, MoraleGroup: 2},
		// 17: 洞穴巨人
		{Name: "洞穴巨人", SkillLevel: 4, HP: 50, Move: 1, MeleeMin: 2, MeleeMax: 5,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 1000, Spoils: 100, Abilities: AbilRegen,
			Home: DwellForest, MaxPopulation: 25, Growth: 3, MoraleGroup: 3},
		// 18: 騎兵
		{Name: "騎兵", SkillLevel: 4, HP: 20, Move: 4, MeleeMin: 3, MeleeMax: 5,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 800, Spoils: 80, Abilities: 0,
			Home: DwellCastle, MaxPopulation: 0, Growth: 2, MoraleGroup: 1},
		// 19: 德魯依
		{Name: "德魯依", SkillLevel: 5, HP: 25, Move: 2, MeleeMin: 2, MeleeMax: 3,
			RangedMin: 10, RangedMax: 0, RangedShots: 3, GoldCost: 700, Spoils: 70, Abilities: AbilMagic,
			Home: DwellForest, MaxPopulation: 25, Growth: 2, MoraleGroup: 2},
		// 20: 法師
		{Name: "法師", SkillLevel: 5, HP: 25, Move: 1, MeleeMin: 2, MeleeMax: 3,
			RangedMin: 25, RangedMax: 0, RangedShots: 2, GoldCost: 1200, Spoils: 120, Abilities: AbilFly | AbilMagic,
			Home: DwellPlains, MaxPopulation: 25, Growth: 2, MoraleGroup: 2},
		// 21: 吸血鬼
		{Name: "吸血鬼", SkillLevel: 5, HP: 30, Move: 1, MeleeMin: 3, MeleeMax: 6,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 1500, Spoils: 150, Abilities: AbilLeech | AbilFly | AbilUndead,
			Home: DwellDungeon, MaxPopulation: 50, Growth: 2, MoraleGroup: 4},
		// 22: 巨人
		{Name: "巨人", SkillLevel: 5, HP: 60, Move: 3, MeleeMin: 10, MeleeMax: 20,
			RangedMin: 5, RangedMax: 10, RangedShots: 6, GoldCost: 2000, Spoils: 200, Abilities: 0,
			Home: DwellHillCave, MaxPopulation: 50, Growth: 2, MoraleGroup: 2},
		// 23: 惡魔
		{Name: "惡魔", SkillLevel: 6, HP: 50, Move: 1, MeleeMin: 5, MeleeMax: 7,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 3000, Spoils: 300, Abilities: AbilFly | AbilScythe,
			Home: DwellDungeon, MaxPopulation: 25, Growth: 1, MoraleGroup: 4},
		// 24: 火龍
		{Name: "火龍", SkillLevel: 6, HP: 200, Move: 1, MeleeMin: 25, MeleeMax: 50,
			RangedMin: 0, RangedMax: 0, RangedShots: 0, GoldCost: 5000, Spoils: 500, Abilities: AbilFly | AbilImmune,
			Home: DwellHillCave, MaxPopulation: 25, Growth: 1, MoraleGroup: 3},
	}
}
