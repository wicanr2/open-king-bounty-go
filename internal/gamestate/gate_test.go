package gamestate

import "testing"

// TestGateToCastle_Unvisited 驗證未造訪的城堡擋下傳送:回 false、座標不動。
func TestGateToCastle_Unvisited(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	contBefore, xBefore, yBefore := gs.Continent, gs.X, gs.Y

	if ok := gs.GateToCastle(a, 3); ok {
		t.Fatalf("GateToCastle(未造訪) = true, want false")
	}
	if gs.Continent != contBefore || gs.X != xBefore || gs.Y != yBefore {
		t.Errorf("未造訪卻仍動了座標:(%d,%d,%d) -> (%d,%d,%d)",
			contBefore, xBefore, yBefore, gs.Continent, gs.X, gs.Y)
	}
}

// TestGateToCastle_Visited 驗證已造訪城堡傳送到 castle_coords 的 (continent, x, y-1)。
func TestGateToCastle_Visited(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.CastleVisited[3] = 1

	if ok := gs.GateToCastle(a, 3); !ok {
		t.Fatalf("GateToCastle(已造訪) = false, want true")
	}

	c := LoadCastles(a)[3]
	if gs.Continent != c.Continent || gs.X != c.X || gs.Y != c.Y-1 {
		t.Errorf("傳送落點錯誤:got (%d,%d,%d), want (%d,%d,%d)",
			gs.Continent, gs.X, gs.Y, c.Continent, c.X, c.Y-1)
	}
}

// TestGateToCastle_DisembarksWhileSailing 驗證乘船中傳送先下船:mount 改騎乘,
// 船留在傳送前的原座標。
func TestGateToCastle_DisembarksWhileSailing(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.CastleVisited[3] = 1
	gs.Mount = KBMountSail
	gs.X, gs.Y = 5, 6

	if ok := gs.GateToCastle(a, 3); !ok {
		t.Fatalf("GateToCastle(已造訪) = false, want true")
	}
	if gs.Mount != KBMountRide {
		t.Errorf("傳送後 Mount = %d, want KBMountRide(%d)", gs.Mount, KBMountRide)
	}
	if gs.BoatX != 5 || gs.BoatY != 6 {
		t.Errorf("船應留在傳送前原座標 (5,6),got (%d,%d)", gs.BoatX, gs.BoatY)
	}
}

// TestGateToTown_Visited 驗證已造訪鄉鎮傳送到 towngate_coords(continent/gate_x/gate_y)。
func TestGateToTown_Visited(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.TownVisited[0] = 1

	if ok := gs.GateToTown(a, 0); !ok {
		t.Fatalf("GateToTown(已造訪) = false, want true")
	}

	tw := LoadTowns(a)[0]
	if gs.Continent != tw.Continent || gs.X != tw.GateX || gs.Y != tw.GateY {
		t.Errorf("傳送落點錯誤:got (%d,%d,%d), want (%d,%d,%d)",
			gs.Continent, gs.X, gs.Y, tw.Continent, tw.GateX, tw.GateY)
	}
}
