package gamestate

// GameState 對應 C 版 src/bounty.h 的 KBgame struct。本檔只涵蓋「建角起手狀態」
// 這條垂直切片,不含世界生成(惡棍/巢穴/寶物佈置等需要 RNG 的部分,那些在 C 版
// spawn_game 內但不屬於本切片範圍)。

// Squad 是玩家隊伍裡的一格兵(對應 C 版 KBgame.player_troops[i] / player_numbers[i]
// 這一組 parallel array 的其中一格)。TroopID == 255(0xFF)代表空格。
type Squad struct {
	TroopID int // 兵種 id(索引 kbdata.Assets.Troops);255/0xFF = 空格
	Count   int // 數量
}

// GameState 是玩家角色的執行期狀態,欄位命名與註解對照 C 版 KBgame(src/bounty.h)。
// 本切片只填「建角當下」會用到的欄位;地圖、惡棍、城堡等世界狀態不在此結構。
type GameState struct {
	Name string // 玩家輸入的角色名(C: game->name)

	Class int // 職業(0 騎士/1 遊俠/2 女巫師/3 蠻俠)(C: byte class)
	Rank  int // 階級,起手固定 0(C: byte rank)

	Gold int // 起手金(C: dword gold)

	BaseLeadership int // 累積領導力基數(C: word base_leadership;逐階 += classes[][].leadership)
	Leadership     int // 目前可用領導力(C: word leadership;起手 = base_leadership)
	Commission     int // 每週俸祿(C: word commission;逐階累積)

	SpellPower int  // 法術威力(C: byte spell_power;逐階累積)
	MaxSpells  int  // 法術上限(C: byte max_spells;逐階累積)
	KnowsMagic bool // 是否天生會魔法(C: byte knows_magic;僅女巫師起手為 1/true)

	Army [5]Squad // 隊伍五格(C: player_troops[5] + player_numbers[5] 兩個 parallel array 合併)
}
