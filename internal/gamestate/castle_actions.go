package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// 城堡佔領者(castle_owner)語意常數,對齊 C bounty.h:34-37。
// castleOwnerMonsters(0x7F)已在 castlegen.go 定義,這裡補其餘三個。
const (
	KBCastlePlayer  = 0xFF // 玩家佔領(bounty.h:35 KBCASTLE_PLAYER)
	KBCastleKnown   = 0x40 // 「已知統治者」旗標,OR 進 owner 高位(bounty.h:34 KBCASTLE_KNOWN)
	KBCastleVillain = 0x1F // 取 villain id 的低 5 位遮罩(bounty.h:37 KBCASTLE_VILLAIN)
)

// HomeCastleID 是家鄉城堡的特殊 id,對齊 C visit_castle 用 MAX_CASTLES 當 home id
// (castle_names[MAX_CASTLES] = "馬克馬斯國王",不受 free 覆寫,見 bounty.c:357)。
const HomeCastleID = kbdata.MaxCastles

// HomeCastleName 是家鄉城堡名。C refill_names 只覆寫 castle_names[0..MAX_TOWNS-1],
// 索引 MAX_CASTLES(家鄉)保留 bounty.c 硬編字面值。
const HomeCastleName = "馬克馬斯國王"

// CastleKind 是 visit_castle 依座標判定出的城堡類別(對齊 C visit_castle 的
// not_found/home/your/enemy 列舉,game.c:2571)。
type CastleKind int

const (
	CastleNone  CastleKind = iota // 該座標沒有城堡(C not_found)
	CastleHome                    // 家鄉城堡(C home)
	CastleYours                   // 玩家已佔領的城堡(C your)
	CastleEnemy                   // 惡棍/怪物佔領的城堡(C enemy)
)

// CastleAt 依 (continent,x,y) 判定該座標上的城堡類別與 id,對齊 C visit_castle
// (game.c:2578-2597)的查找:先比對家鄉城堡座標(land.ini special0),再線性掃
// castles.ini 座標;命中則依 CastleOwner[id] 是否為玩家分成 your/enemy。
// 回傳 (CastleNone, 0) 代表該格不是城堡。
func (gs *GameState) CastleAt(a *kbdata.Assets, continent, x, y int) (CastleKind, int) {
	if hc, hx, hy, ok := HomeCastleCoords(a); ok && hc == continent && hx == x && hy == y {
		return CastleHome, HomeCastleID
	}
	castles := LoadCastles(a)
	for i := range castles {
		if castles[i].Continent == continent && castles[i].X == x && castles[i].Y == y {
			if gs.CastleOwner[i] == KBCastlePlayer {
				return CastleYours, i
			}
			return CastleEnemy, i
		}
	}
	return CastleNone, 0
}

// HomeCastleCoords 讀家鄉城堡座標,來源 land.ini [special0](對齊 C special_coords[SP_HOME])。
// 缺資料時回 ok=false。
func HomeCastleCoords(a *kbdata.Assets) (continent, x, y int, ok bool) {
	if a == nil {
		return 0, 0, 0, false
	}
	ini := a.Strings["land"]
	if ini == nil {
		return 0, 0, 0, false
	}
	sec, found := ini.NumberedSection("special", 0)
	if !found {
		return 0, 0, 0, false
	}
	return sec.IntDefault("continent", 0), sec.IntDefault("x", 0), sec.IntDefault("y", 0), true
}

// HomeTroops 回傳所有「棲地為城堡」的兵種 id(對齊 C visit_home_castle/visit_own_castle
// 起手蒐集 troops[i].dwells == DWELLING_CASTLE 的迴圈)。家鄉城堡招兵與背景立繪的
// 隨機兵種都取自此清單。
func HomeTroops(a *kbdata.Assets) []int {
	var out []int
	if a == nil {
		return out
	}
	for i := range a.Troops {
		if a.Troops[i].Home == kbdata.DwellCastle {
			out = append(out, i)
		}
	}
	return out
}

// SaveCastleOwnerKnowledge 把城堡統治者標記為「已知」(OR 進 KBCastleKnown 位),
// 對齊 C save_castle_owner_knowledge(play.c:1864):玩家/怪物佔領不標記,只有惡棍佔
// 領時記下「我知道是誰」。view_contract 的「所在城堡」欄位靠這個旗標決定顯示名字或「未知」。
func (gs *GameState) SaveCastleOwnerKnowledge(castleID int) {
	if castleID < 0 || castleID >= len(gs.CastleOwner) {
		return
	}
	if gs.CastleOwner[castleID] == KBCastlePlayer || gs.CastleOwner[castleID] == castleOwnerMonsters {
		return
	}
	gs.CastleOwner[castleID] |= KBCastleKnown
}

// dismissTroop 從隊伍移除第 slot 格,其後各格往前遞補、第 4 格清空,
// 對齊 C dismiss_troop(play.c:936)。Squad 空格慣例 = {TroopID:0xFF, Count:0}。
func (gs *GameState) dismissTroop(slot int) {
	for i := slot; i < 4; i++ {
		gs.Army[i] = gs.Army[i+1]
	}
	gs.Army[4] = Squad{TroopID: 0xFF, Count: 0}
}

// GarrisonTroop 把玩家第 slot 格部隊駐防進城堡 castleID,對齊 C garrison_troop
// (play.c:947)。回傳 0 成功、1 城堡無空格、2 這是玩家最後一支部隊(不能全駐防走人)。
func (gs *GameState) GarrisonTroop(castleID, slot int) int {
	// 玩家最後一支部隊:Army[1] 空(只剩第 0 格)則擋下(對齊 C player_troops[1] 檢查)。
	if gs.Army[1].TroopID == 0xFF || gs.Army[1].Count == 0 {
		return 2
	}

	dst := -1
	for i := 0; i < 5; i++ {
		if gs.CastleTroops[castleID][i] == gs.Army[slot].TroopID {
			dst = i
			break
		}
		if gs.CastleNumbers[castleID][i] == 0 {
			dst = i
			break
		}
	}
	if dst == -1 {
		return 1
	}

	gs.CastleTroops[castleID][dst] = gs.Army[slot].TroopID
	gs.CastleNumbers[castleID][dst] += gs.Army[slot].Count
	gs.dismissTroop(slot)
	return 0
}

// UngarrisonTroop 把城堡 castleID 第 slot 格守軍撤回玩家隊伍,對齊 C ungarrison_troop
// (play.c:983)。回傳 0 成功、1 玩家隊伍無空格(撤不回來)。
func (gs *GameState) UngarrisonTroop(castleID, slot int) int {
	if err := gs.addTroop(gs.CastleTroops[castleID][slot], gs.CastleNumbers[castleID][slot]); err != nil {
		return 1
	}
	for i := slot; i < 4; i++ {
		gs.CastleTroops[castleID][i] = gs.CastleTroops[castleID][i+1]
		gs.CastleNumbers[castleID][i] = gs.CastleNumbers[castleID][i+1]
	}
	gs.CastleTroops[castleID][4] = 0xFF
	gs.CastleNumbers[castleID][4] = 0
	return 0
}
