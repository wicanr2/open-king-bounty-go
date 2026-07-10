package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// GameState 對應 C 版 src/bounty.h 的 KBgame struct。涵蓋玩家角色狀態與規則
// (建角、升階、寶箱領導力、招兵、每週結算的玩家面),以及 NewGame 建局時 salt_continent
// 生成的世界狀態(可變地圖 + 棲地/敵人佈置,見下方 World generation 區塊)。
// 城堡/惡棍/契約/船隻等世界狀態仍未納入(見 week.go TODO、docs/PORT-STATUS.md)。

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

	Week int // 已過週數;end_week 每次遞增,用於「每 4 週補農夫」判定(C: week_id = passed_days/WEEK_DAYS)

	// 失竊權杖(破關目標)位置,對應 C game->scepter_continent/scepter_x/scepter_y。
	// 由 NewGame 隨機埋在某洲的某片草地(PlaceScepter);玩家走到該格搜索即獲勝。
	ScepterContinent int
	ScepterX         int
	ScepterY         int

	// TimeStop = 停時術剩餘回合數(對應 C game->time_stop):>0 時世界地圖走路不耗天數
	// (見 worldmap 移動)、戰鬥中敵方跳過。由 SPELL_TIMESTOP 施放累加(spell_power*10)。
	TimeStop int

	// 玩家在世界地圖的位置(對齊 C game->continent/x/y)。過去這三個值只存在
	// WorldMapScreen(cont/px/py),導致存讀檔遺失位置、戰敗無法傳送回家;移入
	// GameState 後隨 encoding/json 自動持久,WorldMapScreen 改讀寫這裡。
	// X/Y 與 C 同座標系(Y-flip 只在渲染);起手值由 NewGame 設為家鄉附近
	// (special0 座標 y-2,對齊 spawn_game)。
	Continent int
	X, Y      int

	// 船隻 / 座騎 / 已探索大陸(對齊 C game->boat/boat_x/boat_y/mount/continent_found)。
	// Boat = 船隻所在洲(0xFF = 未租船);Mount = KBMount*(RIDE/SAIL/FLY);
	// ContinentFound[c] = 是否可航行至該洲(navigate_continent 目的地清單用,起手僅家鄉洲)。
	Boat         byte
	BoatX, BoatY int
	Mount        int
	ContinentFound [kbdata.MaxContinents]byte

	// TownSpell 是「第 N 鎮教哪個法術」的 per-save 隨機分配,對應 C 版
	// byte town_spell[MAX_TOWNS](bounty.h:128),由 salt_spells(play.c:336)在
	// NewGame 建遊戲時填好(見 townspell.go)。索引與 gamestate.Town.ID(town_coords[]
	// 順序)一致;值域 0..maxSpells-1,查 kbdata 對應法術用 LoadSpells(a)[TownSpell[id]]。
	// 注意:這**不是** towns.ini 的靜態 `spell` 欄(那是 Town.SpellID,建置期常數、
	// 26 鎮全填 7=造橋術,現已不供實際遊戲邏輯使用,只保留供 ini 內容溯源)。
	TownSpell [maxTowns]int

	// Spells 是玩家已學會的法術計數,對應 C 版 byte spells[MAX_SPELLS](bounty.h:123);
	// 索引 = 法術 id(0..maxSpells-1),值 = 該法術學會次數(C 允許同一法術重複學習,
	// 見 known_spells() 加總所有格,LearnSpell 每次呼叫只 ++ 一格,不做上限鉗制)。
	Spells [maxSpells]int

	// --- World generation(salt_continent,見 worldgen.go)---
	//
	// 以下欄位對應 C 版 KBgame(src/bounty.h)的世界狀態子集,由 NewGame 呼叫
	// saltContinent 逐洲生成(棲地/敵人的其餘世界狀態,如城堡/惡棍/契約,仍未納入)。
	// 全部 exported,故隨 GameState 一起被 encoding/json 存讀檔(internal/save)
	// 完整持久化,不需另外擴充存檔格式。

	// WorldMap 是「本局」的可變世界地圖:NewGame 時從 kbdata.Assets.World 深拷貝一份,
	// 再讓 saltContinent 就地把 chest tile(0x8B)改寫成 dwelling/artifact/orb/telecave。
	// 對應 C 版 spawn_game 的 `memcpy(game->map, land, ...)` + salt_continent 原地修改
	// game->map。render 層(screen/worldmap.go)應讀這份而非唯讀的 assets.World。
	WorldMap *kbdata.WorldMap

	// DwellingCoords[cont][id] = [x,y],對應 C game->dwelling_coords[MAX_CONTINENTS][MAX_DWELLINGS][2]。
	DwellingCoords [kbdata.MaxContinents][kbdata.MaxDwellings][2]int
	// DwellingTroop[cont][id] = 該棲地的兵種 id,對應 C game->dwelling_troop。
	DwellingTroop [kbdata.MaxContinents][kbdata.MaxDwellings]int
	// DwellingPopulation[cont][id] = 該棲地目前可招募庫存,對應 C game->dwelling_population
	// (招募時會被扣減並持久化,見 screen/recruit.go)。
	DwellingPopulation [kbdata.MaxContinents][kbdata.MaxDwellings]int

	// FoeCoords[cont][id] = [x,y],對應 C game->foe_coords[MAX_CONTINENTS][MAX_FOES][2]。
	FoeCoords [kbdata.MaxContinents][kbdata.MaxFoes][2]int
	// FoeTroops[cont][id][0..4] = 該敵群最多 5 格兵種 id,對應 C game->foe_troops
	// (repopulate_foe 只填前 3 格,見 worldgen.go)。
	FoeTroops [kbdata.MaxContinents][kbdata.MaxFoes][5]int
	// FoeNumbers[cont][id][0..4] = 對應數量,對應 C game->foe_numbers。
	FoeNumbers [kbdata.MaxContinents][kbdata.MaxFoes][5]int

	// OrbFound[cont] = 是否已在該洲取得魔法寶珠(揭示小地圖),對應 C game->orb_found。
	OrbFound [kbdata.MaxContinents]byte
	// Fog[cont][y][x] = 該格是否已探索(對應 C game->fog):玩家走過即揭示周邊 5×5 視窗;
	// view_minimap 在未取得該洲 orb 時只顯示已探索格(取得 orb 則全顯示,見 minimap.go)。
	Fog [kbdata.MaxContinents][kbdata.LevelH][kbdata.LevelW]bool
	// OrbCoords[cont] = [x,y](尋物法球位置),對應 C game->orb_coords[MAX_CONTINENTS][2]。
	// 目前只記座標,尚未接 view_minimap/orb 撿拾流程(見 docs/PORT-STATUS.md B/C 段)。
	OrbCoords [kbdata.MaxContinents][2]int
	// NavmapCoords[cont] = [x,y](導航圖位置),對應 C game->map_coords[MAX_CONTINENTS][2]。
	NavmapCoords [kbdata.MaxContinents][2]int
	// TelecaveCoords[cont][id] = [x,y],對應 C game->teleport_coords[MAX_CONTINENTS][MAX_TELECAVES][2]。
	TelecaveCoords [kbdata.MaxContinents][kbdata.MaxTelecaves][2]int

	// --- Castle / villain / contract world state(salt_villains/repopulate_castle,
	// 見 castlegen.go;由 NewGame 在 saltContinent 之後呼叫)---

	// CastleOwner[id] = 0x7F(怪物,未被惡棍佔領)、0xFF(玩家)、其他值 = 佔領該城堡的
	// 惡棍 id,對應 C game->castle_owner[MAX_CASTLES](bounty.h:136)。
	CastleOwner [kbdata.MaxCastles]byte
	// CastleTroops[id][0..4] = 該城堡守軍兵種 id(惡棍守軍或 repopulate_castle 填的怪物),
	// 對應 C game->castle_troops[MAX_CASTLES][5]。
	CastleTroops [kbdata.MaxCastles][5]int
	// CastleNumbers[id][0..4] = 對應數量,對應 C game->castle_numbers[MAX_CASTLES][5]。
	CastleNumbers [kbdata.MaxCastles][5]int
	// CastleVisited[id] = 玩家是否曾造訪過該城堡(visit_castle 進 your/enemy 分支時設 1),
	// 對應 C game->castle_visited[MAX_CASTLES](bounty.h:126)。select_gate 傳送目的地
	// 清單只列已造訪城堡。
	CastleVisited [kbdata.MaxCastles]byte

	// ArtifactFound[id] = 是否已拾得該神器(對應 C game->artifact_found[MAX_ARTIFACTS]);
	// HasPower 掃這個判斷神器加成(降租金/戰鬥威力等),view_puzzle 拼圖也靠它揭示。
	ArtifactFound [maxArtifacts]byte

	// TownVisited[id] = 玩家是否曾造訪過該鄉鎮(進城時設 1),對應 C game->town_visited。
	// SPELL_TOWNGATE 傳送目的清單只列已造訪的鎮。
	TownVisited [maxTowns]byte

	// VillainCaught[id] = 玩家是否已捕獲該惡棍(打贏其懸賞戰後設 1),對應 C
	// game->villain_caught[MAX_VILLAINS](bounty.h:119)。audience_with_king 用來算
	// 「還需捕獲幾名才能晉升」;晉升條件 classes[class][rank+1].villains_needed。
	VillainCaught [kbdata.MaxVillains]byte

	// Contract 系列,對應 C spawn_game 起手值(play.c:418-425)。
	Contract      byte    // 目前接的懸賞契約 villain id;0xFF = 尚未接(C: byte contract)
	LastContract  byte    // C: byte last_contract,起手固定 0x04
	MaxContract   byte    // C: byte max_contract,起手固定 0x05
	ContractCycle [5]byte // 可接的 5 個惡棍 id 循環,對應 C byte contract_cycle[5]
}
