package gamestate

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// emptySquad 是 C 版 0xFF 空格慣例(troop id = 255,count = 0)。
const emptySquad = 255

// DefaultWorldSeed 是「可重現世界」的固定 rand seed,供測試 / debug 啟動旗標用
// (cmd/openkb 的 -seed 預設值、單元測試固定期望值)。正式新遊戲入口不用它——
// charselect 改用 RandomWorldSeed() 讓每局世界都不同(見下方溯源說明)。
//
// 逆向溯源:C 版原始碼(src/main.c、src/game.c 正常「開新遊戲」路徑)沒有任何
// srand(time(NULL)) —— 全域 `srand(` 只命中三處:debug 主控台 'r' 鍵手動播種、
// 以及 KB_ORACLE 產 golden sample 前顯式 srand(oseed)(預設 1,game.c:7142)。
// 也就是說 openkb 這個重製版每次「開新遊戲」的世界理論上完全相同(glibc rand()
// 未播種等同 srand(1),POSIX 明文)——這是 openkb 為求確定性/測試的取捨,偏離
// 1990 原版 King's Bounty(原版以系統計時器播種,每局世界不同)。
//
// 玩家要的是原版那種「每局不同、有重玩價值」的體驗,故正式入口改用時間種子
// (RandomWorldSeed);演算法本身逐句對齊 C spawn_game,只是把 C 省略的
// srand 補回真正的新遊戲入口。世界在 NewGame 當下即完全物化並存進 GameState
// (WorldMap/ScepterX,Y/TownSpell…),存讀檔沿用,故 seed 不需另外持久化。
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
//   - PlaceScepter(rng)    先跑(對齊 C spawn_game memcpy 後「立刻」埋權杖),於
//     salt_continent 改動地圖前於原始草地上放置;消耗 scepter_key/continent/grass
//     三次 rand。
//   - TownSpell            = salt_spells(rng),接著 WorldMap/Dwelling*/Foe* 等 =
//     salt_continent(rng) 逐洲跑,rng 為同一個 kbrng 實例延續消耗(對齊 C 同一個
//     rand() 資料流)。
//
// RNG parity:rng 消耗順序逐值對齊 C spawn_game——
//   scepter_key → scepter_continent → scepter_grass(PlaceScepter)
//   → salt_spells → salt_continent(四洲)→ salt_villains(四洲)→ repopulate_castle。
// 各子演算法(bury_scepter/salt_continent/populate_dwelling/repopulate_foe/
// roll_creature/salt_villains/repopulate_castle)的放置規則、roll 範圍與型別、
// population 計算逐句忠實對齊 C。故同一個 seed 在 C/Go 兩邊得到相同的世界佈置。
func NewGame(a *kbdata.Assets, name string, class int, seed uint32) *GameState {
	gs := &GameState{
		Name:       name,
		Class:      class,
		Rank:       0,
		Gold:       a.StartGold[class],
		Difficulty: 1, // Normal(目前無難度選擇畫面,固定 600 天;對齊 C 預設)
	}

	// 時間系統起手(對齊 C spawn_game play.c:398-399):天數依難度、步數滿格。
	gs.DaysLeft = gs.MaxDays()
	gs.StepsLeft = DaySteps
	gs.PendingWeekCreature = -1

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

	// 埋藏失竊權杖(破關目標)。對齊 C spawn_game(play.c:386-389):memcpy 地圖後
	// 「立刻」埋——在 salt_spells/salt_continent 消耗 rand、改動地圖之前,於原始草地
	// 上放置(消耗 scepter_key/continent/grass 三次 rand)。a.World 為 nil 的純邏輯
	// 測試沒有地圖就跳過(不消耗這三次 rand),保持既有「缺資料不 panic」慣例。
	if a != nil && a.World != nil {
		gs.WorldMap = copyWorldMap(a.World)
		gs.PlaceScepter(rng)
	}

	// salt_spells:對齊 C spawn_game rand 順序(scepter 之後、salt_continent 之前)。
	gs.TownSpell = saltSpells(rng)

	// 世界生成(salt_continent 四洲):在原始地圖上放置棲地/敵人/神器等。
	if gs.WorldMap != nil {
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
