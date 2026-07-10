package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// BuildBridge 對齊 C build_bridge(game.c:3873):往 (dirX,dirY) 方向(其中一個為 ±1、
// 另一為 0)在接下來 1-2 格「水域」上架橋(水平方向用 TileBridgeH,垂直用 TileBridgeV),
// 遇到非水域即停。回傳實際架設的橋格數(0 = 這裡不適合造橋、白白浪費一個法術)。
//
// 座標:x = X + i*dirX,y = Y - i*dirY(對齊 C `game->y - i*dj`,配合本專案 Y 向上為
// 高 y 的資料座標系;dirY=-1 表示「上」→ y 增加)。
func (gs *GameState) BuildBridge(dirX, dirY int) int {
	if gs.WorldMap == nil {
		return 0
	}
	tile := kbdata.TileBridgeH
	if dirY != 0 {
		tile = kbdata.TileBridgeV
	}
	built := 0
	for i := 1; i < 3; i++ {
		x := gs.X + i*dirX
		y := gs.Y - i*dirY
		m := gs.WorldMap.Tile(gs.Continent, x, y)
		if !kbdata.IsWater(m) {
			break
		}
		gs.WorldMap.Set(gs.Continent, x, y, tile)
		built++
	}
	return built
}

// FindVillain 對齊 C find_villain(play.c:1979):找出目前 last_contract 惡棍所佔的城堡,
// 標記為「已知統治者」(KBCastleKnown)並回傳城堡 id;找不到回 -1。之後呼叫端會開
// view_contract 顯示(此時城堡欄已能顯示名字而非「未知」)。
func (gs *GameState) FindVillain() int {
	for i := 0; i < len(gs.CastleOwner); i++ {
		if int(gs.CastleOwner[i]&0x3f) == int(gs.LastContract) {
			gs.CastleOwner[i] |= KBCastleKnown
			return i
		}
	}
	return -1
}
