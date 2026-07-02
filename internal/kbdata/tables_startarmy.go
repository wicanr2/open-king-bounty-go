package kbdata

// 起手資源表,移植自 C 版 src/bounty.c:
//   starting_gold          (bounty.c:193)
//   starting_army_troop    (bounty.c:200) — 每職業起手兩格兵種 id,0xFF = 空
//   starting_army_numbers  (bounty.c:206) — 對應數量
// 索引 = 職業 (0 騎士 / 1 遊俠 / 2 女巫師 / 3 蠻俠)。

var startGold = [4]int{7500, 10000, 10000, 7500}

var startArmyTroop = [4][2]int{
	{0x02, 0x08}, // 騎士:義勇軍, 弓箭手
	{0x00, 0x02}, // 遊俠:農夫, 義勇軍
	{0x00, 0x01}, // 女巫師:農夫, 小妖精
	{0x03, 0xFF}, // 蠻俠:野狼, (空)
}

var startArmyCount = [4][2]int{
	{20, 2},
	{20, 20},
	{30, 10},
	{20, 0},
}
