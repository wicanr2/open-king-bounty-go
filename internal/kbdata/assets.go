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

import "path/filepath"

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

// Class 是角色職業某一階(對應 C 版 KBclass,classes[4][4]:四職業 × 四階)。
// 欄位與 src/bounty.h 的 KBclass 對齊。C 版部分欄位以「逐階增量」寫,這裡一律存累積後的絕對值。
type Class struct {
	Name           string // 該階頭銜(武士/將軍/元帥/領主…)
	VillainsNeeded int    // 晉此階所需擊敗的惡棍數(絕對門檻)
	Leadership     int    // 領導力(累積)
	MaxSpell       int    // 法術上限(累積)
	SpellPower     int    // 法術威力(累積)
	Commission     int    // 每週俸祿(累積)
	KnowsMagic     int    // 是否天生會魔法(C: knows_magic;僅女巫師起手為 1)
	InstantArmy    int    // 起手即戰兵種旗標(C: instant_army,絕對值)
}

// Assets 是載入後的一整包唯讀遊戲資料。render/gamestate 只讀不寫。
type Assets struct {
	Troops  []Troop             // 由 tables_troops.go 提供(移植自 bounty.c)
	Classes [4][4]Class         // 由 tables_classes.go 提供
	Font    *CJKAtlas           // cjk24.bin 點陣字(nil = 未載入)
	Strings map[string]*FreeIni // free *.ini,key 為檔名去副檔名(troops/spells/…)
	// TODO(P2+): Tiles、Map(land.org)由各 loader 填入。
}

// freeIniNames 是 Load 會嘗試載入的 free 資料清單(缺檔不致命)。
var freeIniNames = []string{"troops", "spells", "towns", "villains", "artifacts", "ui", "colors", "castles"}

// Load 讀取 dir 下的遊戲資料並回傳 *Assets。
// 靜態表(Troops/Classes)恆嵌入;檔案型資料(cjk 字、free ini)採 best-effort:
// 缺檔不報錯(dir=="" 或部分檔缺時仍回傳可用 Assets),讓純邏輯測試免帶資料檔。
func Load(dir string) (*Assets, error) {
	a := &Assets{
		Troops:  troopTable(),
		Classes: classTable(),
		Strings: make(map[string]*FreeIni),
	}
	if dir == "" {
		return a, nil
	}
	if atlas, err := LoadCJKAtlas(filepath.Join(dir, "cjk24.bin")); err == nil {
		a.Font = atlas
	}
	for _, name := range freeIniNames {
		if ini, err := LoadFreeIni(filepath.Join(dir, "free", name+".ini")); err == nil {
			a.Strings[name] = ini
		}
	}
	return a, nil
}
