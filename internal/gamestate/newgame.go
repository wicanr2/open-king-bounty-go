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
// 涵蓋「玩家角色起手數值」(BaseLeadership/Gold/Army 等)、契約起手值、salt_spells
// 的 town_spell 隨機分配、salt_continent 的世界生成(棲地/敵人佈置,見
// worldgen.go),以及 salt_villains/repopulate_castle 的城堡/惡棍世界生成
// (見 castlegen.go)。船隻(boat)/計分等世界狀態仍不在此。難度固定 easy/rank 0。
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
// 世界生成 + 城堡/惡棍):C 版 spawn_game 是一長串固定順序的 rand() 呼叫(角色起手
// 不耗用 rand、salt_spells 耗用一段、salt_continent 每洲各耗用一段、salt_villains/
// repopulate_castle 接續耗用)。本移植目前重現到 salt_spells → salt_continent(四洲)
// → salt_villains(四洲)→ repopulate_castle 這段呼叫序列,但**在此之前** C 版
// spawn_game 已先呼叫過 scepter 相關的 rand(KB_rand(0x00,0xFF) 藏權杖鑰匙、
// KB_rand(100,400) 選洲、grass_on_continent 後再 KB_rand 選格子埋權杖)——Go 版
// NewGame 尚未移植這段 scepter 佈置,兩邊的 rand 呼叫序列在最開頭就已分岔,故
// 「同一個 seed 在 C/Go 兩邊得到完全相同的世界佈置」仍不保證逐 seed 對齊。
// salt_continent/populate_dwelling/repopulate_foe/roll_creature/salt_villains/
// repopulate_castle 演算法本身(放置規則、roll 範圍與型別、population 計算)
// 逐句忠實對齊 C,行為性質一致,只是尚未從第一個 rand() 呼叫起就逐值對齊。
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

	// 玩家起始位置(對齊 C spawn_game:special_coords[SP_HOME],y 減 2)。缺 land 資料
	// (純邏輯測試)時退回 land.ini special0 的字面值 (0,15,10)。
	if hc, hx, hy, ok := HomeCastleCoords(a); ok {
		gs.Continent, gs.X, gs.Y = hc, hx, hy-2
	} else {
		gs.Continent, gs.X, gs.Y = 0, 15, 8
	}
	// 座騎/船隻起手值(對齊 C spawn_game play.c:407-409):騎乘、未租船、僅家鄉洲可航行。
	gs.Mount = KBMountRide
	gs.Boat = noBoat
	gs.ContinentFound[gs.Continent] = 1
	gs.Reveal(gs.Continent, gs.X, gs.Y) // 起點周邊先揭示(fog-of-war)

	// Contract 起手值,對齊 C spawn_game(play.c:418-425)。不消耗 rand,
	// 順序上放在這裡只是照抄 C 原始程式碼位置,對結果無影響。
	gs.Contract = 0xFF
	gs.LastContract = 0x04
	gs.MaxContract = 0x05
	gs.ContractCycle = [5]byte{0, 1, 2, 3, 4}

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

	// 城堡/惡棍世界生成(對齊 C spawn_game,play.c:461-478,緊接在 salt_continent
	// 四洲迴圈之後):先把全部城堡初始化成 castleOwnerMonsters(0x7F),再依序對
	// 四洲呼叫 saltVillains(累加 base_id),最後對仍是 0x7F 的城堡呼叫
	// repopulateCastle 填怪物守軍。
	//
	// 依賴 castles.ini(a.Strings["castles"])而非 a.World:城堡座標/惡棍佈置與
	// 地圖 tile 無關,只要有 castles.ini 就能生成,與 saltContinent 依賴 a.World
	// 是各自獨立的前提(nil-safe 慣例,見 kbdata.Assets 文件)。
	if a != nil && a.Strings["castles"] != nil {
		castles := LoadCastles(a)
		for i := 0; i < kbdata.MaxCastles; i++ {
			gs.CastleOwner[i] = castleOwnerMonsters
		}

		base := 0
		for cont := 0; cont < kbdata.MaxContinents; cont++ {
			base = gs.saltVillains(rng, castles, cont, base)
		}

		for i := 0; i < kbdata.MaxCastles; i++ {
			if gs.CastleOwner[i] == castleOwnerMonsters {
				gs.repopulateCastle(rng, castles, i)
			}
		}
	}

	// 埋藏失竊權杖(破關目標,對齊 C spawn_game play.c:387-389)。放在世界生成之後
	// (掃已就緒的 gs.WorldMap 找草地);⚠ parity 誠實見 PlaceScepter 註記。
	if gs.WorldMap != nil {
		gs.PlaceScepter(rng)
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
