package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// SignpostIndex 回傳玩家目前所站路標在「全地圖掃描序」中的索引,對齊 C read_signpost
// (game.c:3095):依 (continent, y, x) 三層迴圈數過所有 TILE_SIGNPOST,數到玩家所在
// 那一格時的計數即索引(供查 a.Signs)。玩家不在路標格上回 -1。
//
// 掃 gs.WorldMap(執行期地圖,salt 後;路標 tile 不受 salt 改動,故與 land.org 同)。
func (gs *GameState) SignpostIndex() int {
	if gs.WorldMap == nil {
		return -1
	}
	id := 0
	for k := 0; k < kbdata.MaxContinents; k++ {
		for j := 0; j < kbdata.LevelH; j++ {
			for i := 0; i < kbdata.LevelW; i++ {
				if gs.WorldMap.Tile(k, i, j) != kbdata.TileSignpost {
					continue
				}
				if k == gs.Continent && i == gs.X && j == gs.Y {
					return id
				}
				id++
			}
		}
	}
	return -1
}
