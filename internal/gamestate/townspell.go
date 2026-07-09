package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbrng"

// bridgeSpellID / bridgeTownID 對齊 C play.c:336 salt_spells() 的硬編值:
// 造橋術(spell 0x07)固定配給 0x15 號鎮 —— C 原始註解「huntersville sells
// bridge!」(獵人村),並標註「TODO: make this less hardcoded」,故這是 C 版本身
// 承認的硬編特例,不是我們自創。
const (
	bridgeSpellID = 0x07
	bridgeTownID  = 0x15

	unassignedSpell = 0xFF // 分配過程中的「尚未分配」哨兵值,對應 C 的 town_spell[i]=0xFF
)

// saltSpells 把 maxSpells(14)種法術分配到 maxTowns(26)個鄉鎮,對齊 C play.c:336
// salt_spells()(spawn_game 建遊戲時呼叫,play.c:445「隨機分配各鎮教的法術」)。
//
// 演算法逐句對齊 C(不精簡、不改順序):
//  1. 全部 26 鎮先設 unassignedSpell(0xFF)。
//  2. 0x15 號鎮固定分配造橋術(0x07)。
//  3. 依序把 0..13(跳過 0x07,因已在步驟 2 固定分配)分配到隨機挑中、且仍
//     unassignedSpell 的鎮 —— 用「抽鎮→若空則佔用,否則重抽」直到命中空位,
//     故這 13 種法術必定各自落在 13 個不同的鎮(不會重複)。
//  4. 分配完後仍是 unassignedSpell 的鎮(26-14=12 鎮),各自抽一個 0..13 的隨機
//     法術(可能與其他鎮重複 —— C 版允許多鎮教同一法術,只有步驟 3 那 13 鎮保證唯一)。
//
// 省略 C 原碼開頭 `if (MAX_SPELLS > MAX_CASTLES)` 的防呆檢查:該檢查在目前常數下
// (14 <= 26)恆為 false、函式會直接 return,對現行 maxSpells/maxTowns 常數無行為
// 差異,故不搬過來徒增死碼。
//
// parity 誠實註記:C 版此函式是 spawn_game 一長串 rand 呼叫序列中固定位置的一步
// (難度/選項初始化之後、salt_continent 之前)。Go 的 NewGame 目前只重現「玩家角色
// 起手」那段 rand 呼叫(acceptRank 等不消耗 rand),尚未逐一對齊 spawn_game 完整的
// rand 呼叫序列(salt_continent/salt_villains/repopulate_castle 等世界生成邏輯都
// 還沒移植),所以「同一個 seed 值在 C/Go 兩邊得到完全相同的 town_spell 分佈」
// 目前不保證逐 seed 對齊 —— 但演算法本身(造橋固定配 0x15、其餘 13 種法術各自
// 分配到不同鎮、剩餘鎮隨機重複)已忠實移植,行為性質(範圍/唯一性/固定特例)
// 與 C 完全一致。
func saltSpells(rng kbrng.Rand) [maxTowns]int {
	var townSpell [maxTowns]int
	for i := range townSpell {
		townSpell[i] = unassignedSpell
	}

	townSpell[bridgeTownID] = bridgeSpellID

	for i := 0; i < maxSpells; {
		if i == bridgeSpellID {
			i++
			continue
		}
		j := rng.Between(0, maxTowns-1)
		if townSpell[j] == unassignedSpell {
			townSpell[j] = i
			i++
		}
	}

	for i := range townSpell {
		if townSpell[i] == unassignedSpell {
			townSpell[i] = rng.Between(0, maxSpells-1)
		}
	}

	return townSpell
}
