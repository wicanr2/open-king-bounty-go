package gamestate

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// isGrassTile 對齊 C bury_scepter 的草地判定(map 值 0x00 或 0x80)。
func isGrassTile(t byte) bool {
	return t == 0x00 || t == 0x80
}

// grassOnContinent 對齊 C grass_on_continent(play.c):數某洲的草地格總數。
func (gs *GameState) grassOnContinent(continent int) int {
	if gs.WorldMap == nil {
		return 0
	}
	n := 0
	for y := 0; y < kbdata.LevelH; y++ {
		for x := 0; x < kbdata.LevelW; x++ {
			if isGrassTile(gs.WorldMap.Tile(continent, x, y)) {
				n++
			}
		}
	}
	return n
}

// PlaceScepter 對齊 C spawn_game 的權杖埋藏(play.c:387-389)+ bury_scepter(148):
// 隨機選一洲(rand(100,400)/100-1 = 0-3)、在該洲隨機一片草地埋下權杖。
//
// ⚠ parity 誠實(同 NewGame 註記):Go 未逐值重現 C spawn_game 從第一個 rand() 起的
// 完整序列,故權杖埋點不保證與 C 同 seed 一致;但選洲/選草地的演算法忠實對齊。
func (gs *GameState) PlaceScepter(rng kbrng.Rand) {
	gs.ScepterContinent = rng.Between(100, 400)/100 - 1
	if gs.ScepterContinent < 0 || gs.ScepterContinent >= kbdata.MaxContinents {
		gs.ScepterContinent = 0
	}
	count := gs.grassOnContinent(gs.ScepterContinent)
	if count <= 0 {
		return
	}
	target := rng.Between(0, count-1)
	c := 0
	for y := 0; y < kbdata.LevelH; y++ {
		for x := 0; x < kbdata.LevelW; x++ {
			if !isGrassTile(gs.WorldMap.Tile(gs.ScepterContinent, x, y)) {
				continue
			}
			if c == target {
				gs.ScepterX, gs.ScepterY = x, y
				return
			}
			c++
		}
	}
}

// AtScepter 回報玩家目前是否正站在權杖埋藏處(搜索即獲勝)。
func (gs *GameState) AtScepter() bool {
	return gs.Continent == gs.ScepterContinent && gs.X == gs.ScepterX && gs.Y == gs.ScepterY
}
