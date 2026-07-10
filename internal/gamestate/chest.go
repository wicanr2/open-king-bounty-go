package gamestate

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// 寶箱寶藏機率/數量表,逐一對齊 C bounty.c(529-590)的 per-continent 陣列。
// 這些是 bounty.c 硬編字面值,無 DAT 覆寫(chance_for_*/max_gold 等在 game.c 直接以
// game->continent 索引查用,不經 refill_rules),故照抄 C(同 castle_difficulty)。
var (
	chanceForGold       = [kbdata.MaxContinents]int{0x3d, 0x42, 0x4c, 0x47}
	chanceForCommission = [kbdata.MaxContinents]int{0x51, 0x56, 0x56, 0x51}
	chanceForSpellpower = [kbdata.MaxContinents]int{0x56, 0x5c, 0x5d, 0x5b}
	chanceForMaxspell   = [kbdata.MaxContinents]int{0x56, 0x5c, 0x5d, 0x5b}
	chanceForNewspell   = [kbdata.MaxContinents]int{0x65, 0x65, 0x65, 0x65}
	minGold             = [kbdata.MaxContinents]int{0x00, 0x04, 0x09, 0x13}
	maxGold             = [kbdata.MaxContinents]int{0x05, 0x10, 0x15, 0x1f}
	minCommission       = [kbdata.MaxContinents]int{0x09, 0x31, 0x63, 0xc7}
	maxCommission       = [kbdata.MaxContinents]int{0x29, 0x33, 0x65, 0x12d}
	baseMaxspell        = [kbdata.MaxContinents]int{0x01, 0x01, 0x02, 0x02}
)

// ChestKind 是踩到寶箱後判定出的內容類別(對齊 C take_chest 的分支)。
type ChestKind int

const (
	ChestNavmap     ChestKind = iota // 導航海圖:解鎖下一洲(continent_found[c+1])
	ChestOrb                         // 魔法寶珠:揭示本洲小地圖(orb_found[c])
	ChestGoldChoice                  // 金幣/領導力二選一(需玩家選 A/B)
	ChestCommission                  // 每週收入 +points
	ChestSpellPower                  // 法術威力 +1
	ChestMaxSpell                    // 法術上限 +points
	ChestNewSpell                    // 獲得法術 spells[SpellType] += Points
	ChestEmpty                       // 空箱
)

// ChestResult 描述寶箱結果,供 UI 顯示訊息 + 金幣/領導力選擇。非選擇類的獎勵在
// TakeChest 內已直接套用;金幣類則把數額帶出,由呼叫端(ChestScreen)依玩家選擇套用。
type ChestResult struct {
	Kind       ChestKind
	Continent  int // ChestNavmap:新解鎖的洲(c+1)
	Gold       int // ChestGoldChoice:選 A 取得的金幣
	Leadership int // ChestGoldChoice:選 B 取得的領導力
	Points     int // Commission/MaxSpell/NewSpell 的數量
	SpellType  int // ChestNewSpell:法術索引
}

// TakeChest 對齊 C take_chest(game.c:3167):判定玩家目前所在格(gs.Continent/X/Y)的
// 寶箱內容並套用效果,最後把該格 tile 清 0(對齊 C 移除寶箱)。金幣/領導力二選一的
// 情況只回傳數額(Kind=ChestGoldChoice),實際套用由 ApplyGoldChoice 依玩家選擇完成。
//
// ⚠ 神器(POWER_DOUBLE_LEADERSHIP / DOUBLE_MAX_SPELLS)未建模,對應加倍略過。
func (gs *GameState) TakeChest(a *kbdata.Assets, rng kbrng.Rand) ChestResult {
	cont, x, y := gs.Continent, gs.X, gs.Y
	clear := func() {
		if gs.WorldMap != nil {
			gs.WorldMap.Set(cont, x, y, 0)
		}
	}

	// 0a. 導航海圖:解鎖下一洲。
	if cont < kbdata.MaxContinents-1 &&
		gs.NavmapCoords[cont][0] == x && gs.NavmapCoords[cont][1] == y {
		gs.ContinentFound[cont+1] = 1
		clear()
		return ChestResult{Kind: ChestNavmap, Continent: cont + 1}
	}
	// 0b. 魔法寶珠:揭示本洲小地圖。
	if gs.OrbCoords[cont][0] == x && gs.OrbCoords[cont][1] == y {
		gs.OrbFound[cont] = 1
		clear()
		return ChestResult{Kind: ChestOrb}
	}

	// 1. 一般寶藏:擲骰決定類型(對齊 C chance = KB_rand(1,100) 的級聯門檻)。
	chance := rng.Between(1, 100)
	clear()
	switch {
	case chance < chanceForGold[cont]:
		points := rng.Between(1, maxGold[cont]) + minGold[cont]
		gold := points * 100
		leadership := gold / 50
		return ChestResult{Kind: ChestGoldChoice, Gold: gold, Leadership: leadership}
	case chance < chanceForCommission[cont]:
		points := rng.Between(1, maxCommission[cont]) + minCommission[cont]
		gs.Commission += points
		return ChestResult{Kind: ChestCommission, Points: points}
	case chance < chanceForSpellpower[cont]:
		gs.SpellPower++
		return ChestResult{Kind: ChestSpellPower, Points: 1}
	case chance < chanceForMaxspell[cont]:
		points := baseMaxspell[cont]
		gs.MaxSpells += points
		return ChestResult{Kind: ChestMaxSpell, Points: points}
	case chance < chanceForNewspell[cont]:
		spellType := rng.Between(0, len(gs.Spells)-1)
		spellNum := rng.Between(1, cont+1)
		gs.Spells[spellType] += spellNum
		return ChestResult{Kind: ChestNewSpell, Points: spellNum, SpellType: spellType}
	default:
		return ChestResult{Kind: ChestEmpty}
	}
}

// ApplyGoldChoice 套用 ChestGoldChoice 的玩家選擇(對齊 C gold_or_leadership 回傳後的
// 分支):takeGold=true 取金幣;false 取領導力(用 AddChestLeadership 做永久加成)。
func (gs *GameState) ApplyGoldChoice(r ChestResult, takeGold bool) {
	if takeGold {
		gs.Gold += r.Gold
		return
	}
	gs.AddChestLeadership(r.Leadership)
}

// AddChestLeadership 套用寶箱的領導力加成,對應 C 版 game.c 的寶箱處理
// (game->base_leadership += leadership; game->leadership += leadership;)。
//
// 關鍵(issue #5):必須同時加進 BaseLeadership 與 Leadership。只加 Leadership 的話,
// 下一次 EndWeek 會把 Leadership 重設回 BaseLeadership,加成當週就流失。加進 BaseLeadership
// 才是永久加成;同步加 Leadership 讓當週立即生效。
func (gs *GameState) AddChestLeadership(bonus int) {
	gs.BaseLeadership += bonus
	gs.Leadership += bonus
}
