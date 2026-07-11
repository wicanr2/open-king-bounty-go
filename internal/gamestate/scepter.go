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

// PlaceScepter 對齊 C spawn_game 的權杖埋藏(play.c:386-389)+ bury_scepter(148),
// 消耗序列與 C 完全一致:先 scepter_key = KB_rand(0x00,0xFF)、再 scepter_continent =
// KB_rand(100,400)/100-1(=0-3)、最後在該洲第 KB_rand(0, grass-1) 片草地埋下權杖。
//
// 呼叫時機須對齊 C:memcpy 地圖後「立刻」埋(在 salt_spells/salt_continent 改動地圖前、
// 於原始草地上),故 NewGame 在 copyWorldMap 之後、saltSpells 之前呼叫本函式——如此
// 同一個 seed 在 C/Go 兩邊的 scepter 佈置與後續 rand 資料流即逐值對齊。
func (gs *GameState) PlaceScepter(rng kbrng.Rand) {
	gs.ScepterKey = byte(rng.Between(0x00, 0xFF))
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
