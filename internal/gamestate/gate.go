package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// gateTeleport 執行 select_gate 的實際傳送(對齊 C select_gate done==2 段,game.c:4256):
// 乘船中先下船(mount=RIDE、船留原地),再把玩家移到目的座標。
func (gs *GameState) gateTeleport(continent, x, y int) {
	if gs.Mount == KBMountSail {
		gs.Mount = KBMountRide
		gs.BoatX, gs.BoatY = gs.X, gs.Y
	}
	gs.Continent, gs.X, gs.Y = continent, x, y
}

// GateToCastle 傳送到已造訪的城堡 castleID,對齊 C select_gate mode 0
// (game.c:4223):落點 = castle_coords[id] 的 (continent, x, y-1)。未造訪則不動、回 false。
func (gs *GameState) GateToCastle(a *kbdata.Assets, castleID int) bool {
	if castleID < 0 || castleID >= len(gs.CastleVisited) || gs.CastleVisited[castleID] == 0 {
		return false
	}
	castles := LoadCastles(a)
	if castleID >= len(castles) {
		return false
	}
	c := castles[castleID]
	gs.gateTeleport(c.Continent, c.X, c.Y-1)
	return true
}

// GateToTown 傳送到已造訪的鄉鎮 townID,對齊 C select_gate mode 1(game.c:4232):
// 落點 = towngate_coords[id] = towns.ini gate_x/gate_y。未造訪則不動、回 false。
func (gs *GameState) GateToTown(a *kbdata.Assets, townID int) bool {
	if townID < 0 || townID >= len(gs.TownVisited) || gs.TownVisited[townID] == 0 {
		return false
	}
	towns := LoadTowns(a)
	if townID >= len(towns) {
		return false
	}
	t := towns[townID]
	gs.gateTeleport(t.Continent, t.GateX, t.GateY)
	return true
}
