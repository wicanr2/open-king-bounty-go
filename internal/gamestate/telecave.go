package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// VisitTelecave 對齊 C visit_telecave(game.c:2973,force=1):若玩家目前座標
// (Continent/X/Y)正好是某座傳送洞的一端,就瞬移到成對的另一端並回傳 true;
// 否則回傳 false(該格其實是地下城棲地——TILE_TELECAVE 與 TILE_DWELLING_3 同值 0x8E,
// C 靠座標區分,這裡照做:先問傳送洞座標,不中才當一般棲地處理)。
//
// 傳送洞每洲一對(MAX_TELECAVES=2),i 端 ↔ (1-i) 端;座標由 saltContinent 生成
// (見 worldgen.go TelecaveCoords)。
func (gs *GameState) VisitTelecave() bool {
	for i := 0; i < kbdata.MaxTelecaves; i++ {
		if gs.TelecaveCoords[gs.Continent][i][0] == gs.X &&
			gs.TelecaveCoords[gs.Continent][i][1] == gs.Y {
			other := 1 - i
			gs.X = gs.TelecaveCoords[gs.Continent][other][0]
			gs.Y = gs.TelecaveCoords[gs.Continent][other][1]
			return true
		}
	}
	return false
}
