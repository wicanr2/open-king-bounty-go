package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// fogRadius 是探索揭示半徑:玩家置中,每邊 2 格(對齊世界地圖 5×5 viewport,worldmap.go radii=2)。
const fogRadius = 2

// Reveal 把 (cx,cy) 周邊 5×5 範圍標記為已探索(對齊 C 版走動時逐格填 game->fog)。
// 界外座標自動略過。玩家移動與 NewGame 起點各呼叫一次。
func (gs *GameState) Reveal(cont, cx, cy int) {
	if cont < 0 || cont >= kbdata.MaxContinents {
		return
	}
	for dy := -fogRadius; dy <= fogRadius; dy++ {
		for dx := -fogRadius; dx <= fogRadius; dx++ {
			x, y := cx+dx, cy+dy
			if x < 0 || x >= kbdata.LevelW || y < 0 || y >= kbdata.LevelH {
				continue
			}
			gs.Fog[cont][y][x] = true
		}
	}
}
