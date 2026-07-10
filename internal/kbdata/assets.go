// Package kbdata 是資料層:讀解所有原始格式,吐出乾淨、唯讀的 Go 型別。
//
// 窄介面(deep module):對外只有 Load(dir) → *Assets。呼叫端不需知道
// free ini / DOS .CC / Genesis ROM / cjk24.bin 的任何細節。
//
// 型別對照 C 版:
//
//	Troop  ← KBtroop  (src/bounty.h / src/bounty.c: troops[MAX_TROOPS])
//	Class  ← KBclass  (src/bounty.h / src/bounty.c: classes[4][4])
//
// 這些純資料表由 P1 的移植工作(subagent)填入 tables_troops.go 等檔。
package kbdata

import (
	"io/fs"
	"os"
)

// Dwelling 是兵種棲地(對應 C 版 DWELLING_*)。數值刻意逐一對齊 C 版 bounty.h 的
// #define(DWELLING_PLAINS=0, DWELLING_FOREST=1, DWELLING_HILLCAVE=2, DWELLING_DUNGEON=3,
// DWELLING_CASTLE=4)——這個數值本身有意義:C 版 populate_dwelling() 直接拿它做
// `TILE_DWELLING_1 + troops[troop_id].dwells` 算出地圖 tile,不是純標籤。
//
// ⚠ 勘誤(2026-07,世界生成移植時發現):本檔原先用 iota 由 DwellCastle 起算,
// 順序對不上 C(把 CASTLE 排第一個變成 0,PLAINS/FOREST/HILLCAVE/DUNGEON 各自 +1)。
// 核對全庫(2026-07-10)只有具名常數比較(如 `Home == DwellCastle`)在用這個型別,
// 沒有任何既有程式碼依賴其原始數值,故直接更正 iota 順序對齊 C,不做相容別名。
// worldgen.go 的 populateDwelling 依賴這個數值算 dwelling tile,必須對齊。
type Dwelling uint8

const (
	DwellPlains   Dwelling = iota // 0 (C: DWELLING_PLAINS)
	DwellForest                   // 1 (C: DWELLING_FOREST)
	DwellHillCave                 // 2 (C: DWELLING_HILLCAVE)
	DwellDungeon                  // 3 (C: DWELLING_DUNGEON)
	DwellCastle                   // 4 (C: DWELLING_CASTLE)
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
	Name          string   // 繁中兵種名(沿用 C 版 free 資料)
	SkillLevel    int      // SL
	HP            int      // 每隻血量
	Move          int      // MV 移動力
	MeleeMin      int      // 近戰最小
	MeleeMax      int      // 近戰最大
	RangedMin     int      // 遠攻最小(0 = 無遠攻)
	RangedMax     int      // 遠攻最大
	RangedShots   int      // 彈藥數
	GoldCost      int      // 招募金價
	Spoils        int      // 戰利品(金)
	Abilities     Ability  // 能力位元 OR
	Home          Dwelling // 棲地
	MaxPopulation int      // 棲地單次生成上限(C 版 KBtroop.max_population;見下方勘誤)
	Growth        int      // 每週增長基數(C 版 KBtroop.growth;見下方勘誤)
	MoraleGroup   int      // 士氣群 _A.._E
}

// ⚠ 欄位勘誤(2026-07,世界生成移植時發現):C 版 KBtroop(src/bounty.h)欄位順序是
// `dwells, max_population, growth, morale_group` 四欄(src/bounty.c troops[] 字面值逐一核對,
// 如農夫 `_PLAINS,250,6,_A` = dwells=_PLAINS, max_population=250, growth=6, morale_group=_A;
// 火龍 `_HILL,25,1,_D` = max_population=25, growth=1)。本檔原先把這兩欄命名為 Growth/MoraleTop
// (且順序按 tables_troops.go 舊註解「growth_num, morale_cap」誤植),數值本身逐位置搬得沒錯,
// 只是命名對調、語意寫反。核對全庫(2026-07-10)只有 kbdata 自己的表/測試引用這兩欄,
// gamestate/combat 尚未消費,故直接更正欄位名為 MaxPopulation/Growth(值不變),不做相容別名。

// Class 是角色職業某一階(對應 C 版 KBclass,classes[4][4]:四職業 × 四階)。
// 欄位與 src/bounty.h 的 KBclass 對齊,值為 bounty.c 字面值(C 為真值,不做累積轉換)。
//
// ⚠ 混合語意(見 tables_classes.go 與 C player_accept_rank):
//   - Leadership/MaxSpell/SpellPower/Commission/KnowsMagic 是**逐階增量**,
//     由 gamestate.acceptRank 用 += 累加(讓 base_leadership 能疊加升階+寶箱等多來源)。
//   - VillainsNeeded/InstantArmy 是**該階絕對值**,直接查用、不累加。
type Class struct {
	Name           string // 該階頭銜(武士/將軍/元帥/領主…)
	VillainsNeeded int    // 晉此階所需擊敗的惡棍數(絕對門檻)
	Leadership     int    // 領導力增量(逐階,acceptRank 累加)
	MaxSpell       int    // 法術上限增量(逐階)
	SpellPower     int    // 法術威力增量(逐階)
	Commission     int    // 每週俸祿增量(逐階)
	KnowsMagic     int    // 會魔法增量(逐階;僅女巫師 rank0 為 1,累加後恆 1)
	InstantArmy    int    // 起手即戰兵種旗標(絕對值)
}

// Assets 是載入後的一整包唯讀遊戲資料。render/gamestate 只讀不寫。
type Assets struct {
	Troops  []Troop             // 由 tables_troops.go 提供(移植自 bounty.c)
	Classes [4][4]Class         // 由 tables_classes.go 提供
	Font    *CJKAtlas           // cjk24.bin 點陣字(nil = 未載入)
	Strings map[string]*FreeIni // free *.ini,key 為檔名去副檔名(troops/spells/…)
	World   *WorldMap           // land.org 世界地圖(nil = 未載入)

	// 起手資源(索引 = 職業),移植自 bounty.c;供 gamestate 建角使用。
	StartGold      [4]int    // 各職業起手金
	StartArmyTroop [4][2]int // 各職業起手兩格兵種 id(0xFF = 空)
	StartArmyCount [4][2]int // 對應數量
	// TODO(P2+): Tiles(tileset 美術)由 render 層填入。
}

// freeIniNames 是 Load 會嘗試載入的 free 資料清單(缺檔不致命)。
var freeIniNames = []string{"troops", "spells", "towns", "villains", "artifacts", "ui", "colors", "castles"}

// Load 讀取 dir 下的遊戲資料並回傳 *Assets。dir=="" 只回靜態表(供純邏輯測試)。
func Load(dir string) (*Assets, error) {
	if dir == "" {
		return LoadFS(nil)
	}
	return LoadFS(os.DirFS(dir))
}

// LoadFS 從任意檔案系統讀遊戲資料(桌面用 os.DirFS(dir),行動裝置用 embed.FS,免解壓到唯讀 cwd)。
// 路徑相對於 fsys 根:cjk24.bin、free/*.ini、free/land.org。best-effort:缺檔不報錯。
func LoadFS(fsys fs.FS) (*Assets, error) {
	a := &Assets{
		Troops:         troopTable(),
		Classes:        classTable(),
		Strings:        make(map[string]*FreeIni),
		StartGold:      startGold,
		StartArmyTroop: startArmyTroop,
		StartArmyCount: startArmyCount,
	}
	if fsys == nil {
		return a, nil
	}
	if b, err := fs.ReadFile(fsys, "cjk24.bin"); err == nil {
		if atlas, err := ParseCJKAtlas(b); err == nil {
			a.Font = atlas
		}
	}
	for _, name := range freeIniNames {
		if b, err := fs.ReadFile(fsys, "free/"+name+".ini"); err == nil {
			a.Strings[name] = ParseFreeIni(b)
		}
	}
	if b, err := fs.ReadFile(fsys, "free/land.org"); err == nil {
		if wm, err := ParseWorldMap(b); err == nil {
			a.World = wm
		}
	}
	return a, nil
}
