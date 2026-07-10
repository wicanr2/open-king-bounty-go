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

	// VillainDescs[id] 是 view_contract 用的懸賞契約多行描述文字,對應 C 版
	// STRL_VDESCS(free-data.c:1132 `KB_strcpy(tmp, DOS_villain_names[sub_id]);
	// KB_strcat(tmp, ".txt"); GNU_read_textfile(mod, tmp, 0)`):讀
	// data/free/<villains.ini villainN.file>.txt 整檔原文(含真實 \n、含兩個
	// %s 佔位:第一個是最後現身大陸洲名、第二個是城堡名,呼叫端 fmt.Sprintf 代入)。
	// 缺檔該筆維持空字串,呼叫端須 nil/空字串安全。
	VillainDescs [MaxVillains]string

	// Signs 是路標文字清單,對應 C STRL_SIGNS(free-data.c:1099):讀 signs.txt 後用
	// 「每隔一個換行當分隔」的規則切成多則(見 parseSigns)。read_signpost 依路標在
	// 地圖上的掃描序當索引取用。缺檔為 nil,呼叫端須越界安全。
	Signs []string

	// Credits 是製作名單原文(對應 C STRL_CREDITS = credits.txt 整檔),show_credits 顯示。
	Credits string
}

// freeIniNames 是 Load 會嘗試載入的 free 資料清單(缺檔不致命)。
//
// "land" 收錄 land.ini(洲名/國王居所/魔法密室座標),與 land.org(世界地圖 tile
// grid 本體,見 land.go)是不同檔案、不同用途 —— 這裡只解「中繼資料」的 ini 格式,
// 不影響地圖 tile。加入這份清單是 view_contract 移植時發現的缺口:C 版
// refill_names()(game.c:221)在每次開新遊戲後,把 bounty.c 硬編的
// continent_names[]("大陸洲"/"森林洲"/"群島洲"/"撒哈洲")整個覆寫成
// STRL_CONTINENTS 讀到的 land.ini continent0-3 name 欄位("弗蘭德利亞"/"第二洲"/
// "第三洲"/"第四洲")——bounty.c 那組字面值在 free 模組實際執行時從未真正顯示過,
// 是死預設值(核對見 gamestate/continent.go 檔頭注記),故本移植不採用 bounty.c
// 硬編陣列,改讀 land.ini,對齊 refill_names() 實際覆寫後的結果。
var freeIniNames = []string{"troops", "spells", "towns", "villains", "artifacts", "ui", "colors", "castles", "land"}

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
	loadVillainDescs(fsys, a)
	if b, err := fs.ReadFile(fsys, "free/signs.txt"); err == nil {
		a.Signs = parseSigns(string(b))
	}
	if b, err := fs.ReadFile(fsys, "free/credits.txt"); err == nil {
		a.Credits = string(b)
	}
	return a, nil
}

// parseSigns 把 signs.txt 切成多則路標,逐一對齊 C STRL_SIGNS(free-data.c:1099)的
// 「toggle」規則:掃過每個 '\n',以 0/1 交替——遇到偶數個(j==0)的換行保留成則內
// 換行,遇到奇數個(j==1)的換行當作「則」的分隔。故相鄰兩則由一個「空行」隔開,
// 但則內可含單一換行(多行路標)。此規則脆弱但忠實照抄 C。
func parseSigns(data string) []string {
	var signs []string
	var cur []byte
	j := 0
	for i := 0; i < len(data); i++ {
		ch := data[i]
		if ch == '\n' {
			if j == 1 {
				signs = append(signs, string(cur))
				cur = cur[:0]
			} else {
				cur = append(cur, '\n')
			}
			j = 1 - j
			continue
		}
		cur = append(cur, ch)
	}
	if len(cur) > 0 {
		signs = append(signs, string(cur))
	}
	return signs
}

// loadVillainDescs 讀 17 個惡棍的 STRL_VDESCS 描述文字檔(見 Assets.VillainDescs
// 欄位註解)。檔名取自 villains.ini 各 villainN 區塊的 file 欄位(對齊 C 版
// DOS_villain_names[sub_id],兩者索引順序逐一核對一致,見
// gamestate/villain.go 檔頭注記),不是另建一份寫死清單。
func loadVillainDescs(fsys fs.FS, a *Assets) {
	ini := a.Strings["villains"]
	if ini == nil {
		return
	}
	for i := 0; i < MaxVillains; i++ {
		sec, ok := ini.NumberedSection("villain", i)
		if !ok {
			continue
		}
		file := sec.GetDefault("file", "")
		if file == "" {
			continue
		}
		if b, err := fs.ReadFile(fsys, "free/"+file+".txt"); err == nil {
			a.VillainDescs[i] = string(b)
		}
	}
}
