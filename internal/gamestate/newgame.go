package gamestate

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// emptySquad 是 C 版 0xFF 空格慣例(troop id = 255,count = 0)。
const emptySquad = 255

// DefaultWorldSeed 是尚未接上「per-playthrough 持久世界種子」前的預設 rand seed。
//
// 誠實記錄一個逆向溯源結果:C 版原始碼(src/main.c、src/game.c 正常「開新遊戲」
// 路徑)**沒有任何 srand(time(NULL)) 呼叫** —— 全域搜尋 `srand(` 只在 game.c 命中
// 三處:debug 主控台 'r' 鍵手動重新播種、以及 KB_ORACLE 產golden sample 前顯式
// srand(oseed)(oseed 未指定時預設 1,見 game.c:7142)。也就是說,C 版玩家若從
// 不使用 debug 主控台重新播種,同一份執行檔每次「開新遊戲」的 world_spell/
// 地圖生成序列理論上完全相同(glibc rand() 未播種時的行為等同 srand(1),POSIX
// 規範明文如此)。
//
// 因此這裡選 1 當預設值,不是隨便挑的數字,也不是我方自創的「加時間戳讓每局
// 不同」的新行為 —— 那樣反而是替 C 版本沒有的行為捏造一個。之後若要接上「每局
// 世界都不同」的持久種子(連動 worldmap 的敵人佈置、地圖生成等尚未建模的世界
// RNG),應該是另一個明確任務,牽動的不只 town_spell 一處。
const DefaultWorldSeed uint32 = 1

// NewGame 建立角色起手狀態,對應 C 版 src/play.c 的 spawn_game。
// 涵蓋「玩家角色起手數值」(BaseLeadership/Gold/Army 等)、salt_spells 的
// town_spell 隨機分配,以及 salt_continent 的世界生成(棲地/敵人佈置,見
// worldgen.go);城堡/惡棍/契約/船隻等世界狀態仍不在此(spawn_game 呼叫
// salt_continent 之後還有 salt_villains/repopulate_castle 等,尚未移植)。
// 難度固定 easy/rank 0。
//
// 忠實對齊 spawn_game:先 Rank=0,再呼叫 acceptRank(等同 C 的 player_accept_rank),
// 由它把 classes[class][0] 的**增量**累加進 BaseLeadership/MaxSpells/SpellPower/Commission/
// KnowsMagic(與升階共用同一條路徑);Leadership 起手 = BaseLeadership。
//   - Gold                = starting_gold[class]
//   - player_troops/numbers[0..1] = starting_army_troop/numbers[class];[2..4] = 0xFF/0
//   - TownSpell            = salt_spells(rng),接著 WorldMap/Dwelling*/Foe* 等 =
//     salt_continent(rng) 逐洲跑(對齊 C spawn_game 呼叫順序:salt_spells 之後、
//     salt_villains 之前),rng 為同一個 kbrng 實例延續消耗(對齊 C 同一個 rand()
//     資料流,見下方 parity 誠實註記)。
//
// world seed / RNG parity 誠實註記(同 townspell.go 的 saltSpells 註記,範圍擴大到
// 世界生成):C 版 spawn_game 是一長串固定順序的 rand() 呼叫(角色起手不耗用 rand、
// salt_spells 耗用一段、salt_continent 每洲各耗用一段、salt_villains/repopulate_castle
// 還會再耗用)。本移植目前只重現到 salt_spells + salt_continent 這段呼叫序列,
// 尚未移植 salt_villains/repopulate_castle(它們不消耗獨立 rand 流,只是還沒實作),
// 所以「同一個 seed 在 C/Go 兩邊得到完全相同的世界佈置」仍不保證逐 seed 對齊
// (C 版本身在 salt_continent 之前已經呼叫過 scepter 相關的 rand,而 Go 版
// NewGame 尚未移植 scepter 佈置,兩邊呼叫序列在那之前就已經分岔)——但
// salt_continent/populate_dwelling/repopulate_foe/roll_creature 演算法本身
// (放置規則、roll 範圍與型別、population 計算)逐句忠實對齊 C,行為性質一致。
func NewGame(a *kbdata.Assets, name string, class int, seed uint32) *GameState {
	gs := &GameState{
		Name:  name,
		Class: class,
		Rank:  0,
		Gold:  a.StartGold[class],
	}

	// 套用 rank 0 增量(base 從 0 起),再讓當前可用領導力對齊 base。
	gs.acceptRank(a)
	gs.Leadership = gs.BaseLeadership

	for i := 0; i < 2; i++ {
		gs.Army[i] = Squad{
			TroopID: a.StartArmyTroop[class][i],
			Count:   a.StartArmyCount[class][i],
		}
	}
	for i := 2; i < 5; i++ {
		gs.Army[i] = Squad{TroopID: emptySquad, Count: 0}
	}

	rng := kbrng.NewGlibc(seed)
	gs.TownSpell = saltSpells(rng)

	// 世界生成:a.World 是 nil-safe 的(kbdata.Load("") 純邏輯測試不帶地圖檔時
	// a.World == nil),沒有地圖就跳過,保持既有「缺資料不 panic」慣例。
	if a != nil && a.World != nil {
		gs.WorldMap = copyWorldMap(a.World)
		for cont := 0; cont < kbdata.MaxContinents; cont++ {
			// 對齊 C spawn_game:`salt_continent(game, i, 2, 1, 1, 2, 10, 5);`
			// (2 artifact / 1 navmap / 1 orb / 2 telecave / 10 dwelling / 5 friendly)。
			gs.saltContinent(a, rng, cont, 2, 1, 1, 2, 10, 5)
		}
	}

	return gs
}

// copyWorldMap 深拷貝 kbdata.WorldMap,讓 salt_continent 能就地修改而不動到
// 唯讀的 kbdata.Assets.World(同一份 assets 可能被多個存檔/角色共用)。
func copyWorldMap(src *kbdata.WorldMap) *kbdata.WorldMap {
	if src == nil {
		return nil
	}
	dst := &kbdata.WorldMap{
		Continents: src.Continents,
		Tiles:      make([][kbdata.LevelH][kbdata.LevelW]byte, len(src.Tiles)),
	}
	copy(dst.Tiles, src.Tiles) // 陣列值複製(非共享底層),[LevelH][LevelW]byte 是值型別
	return dst
}
