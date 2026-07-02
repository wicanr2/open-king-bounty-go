// Package kbdata 是資料層:讀解所有原始格式,吐出乾淨、唯讀的 Go 型別。
//
// 窄介面(deep module):對外只有 Load(dir) → *Assets。呼叫端不需知道
// free ini / DOS .CC / Genesis ROM / cjk24.bin 的任何細節。
//
// 型別對照 C 版:
//   Troop  ← KBtroop  (src/bounty.h / src/bounty.c: troops[MAX_TROOPS])
//   Class  ← KBclass  (src/bounty.h / src/bounty.c: classes[4][4])
// 這些純資料表由 P1 的移植工作(subagent)填入 tables_troops.go 等檔。
package kbdata

// Dwelling 是兵種棲地(對應 C 版 DWELLING_*)。
type Dwelling uint8

const (
	DwellCastle Dwelling = iota
	DwellPlains
	DwellForest
	DwellHillCave
	DwellDungeon
)

// Ability 是兵種能力位元(對應 C 版 ABIL_*,可 OR 疊加)。
type Ability uint16

const (
	AbilFly Ability = 1 << iota
	AbilRegen
	AbilMagic
	AbilImmune
	AbilAbsorb
	AbilLeech
	AbilScythe
	AbilUndead
)

// Troop 是單一兵種的靜態數值,對應 C 版 KBtroop(src/bounty.h)。
// 欄位語意以 src/bounty.c 的 troops[] 註解為準(SL/HP/MV、近戰、遠攻、金價、戰利品…)。
type Troop struct {
	Name        string   // 繁中兵種名(沿用 C 版 free 資料)
	SkillLevel  int      // SL
	HP          int      // 每隻血量
	Move        int      // MV 移動力
	MeleeMin    int      // 近戰最小
	MeleeMax    int      // 近戰最大
	RangedMin   int      // 遠攻最小(0 = 無遠攻)
	RangedMax   int      // 遠攻最大
	RangedShots int      // 彈藥數
	GoldCost    int      // 招募金價
	Spoils      int      // 戰利品(金)
	Abilities   Ability  // 能力位元 OR
	Home        Dwelling // 棲地
	Growth      int      // 每週增長基數(C 版第一個數字)
	MoraleTop   int      // 士氣群上限(C 版第二個數字)
	MoraleGroup int      // 士氣群 _A.._E
}

// Class 是角色職業(對應 C 版 KBclass,classes[4][4]:四職業 × 四難度)。
// 具體欄位由 P1 移植時對照 src/bounty.h 的 KBclass 定義補齊。
type Class struct {
	Name       string
	Leadership int
	Commission int
	MaxSpell   int
	SpellPower int
	// TODO(P1): 依 src/bounty.h KBclass 補齊其餘欄位(移植 subagent 負責)。
}

// Assets 是載入後的一整包唯讀遊戲資料。render/gamestate 只讀不寫。
type Assets struct {
	Troops  []Troop     // 由 tables_troops.go 提供(移植自 bounty.c)
	Classes [4][4]Class // 由 tables_classes.go 提供
	// TODO(P1): Font(cjk24 atlas)、Tiles、Map、Strings(free ini)等由各 loader 填入。
}

// Load 讀取 dir 下的遊戲資料並回傳 *Assets。
// P0 僅回傳嵌入的靜態表(Troops/Classes);檔案型 loader(ini/atlas/map)P1 接上。
func Load(dir string) (*Assets, error) {
	return &Assets{
		Troops:  troopTable(),
		Classes: classTable(),
	}, nil
}
