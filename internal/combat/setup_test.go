package combat

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

func loadTestAssets(t *testing.T) *kbdata.Assets {
	t.Helper()
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatalf("kbdata.Load(\"\") err: %v", err)
	}
	return a
}

// 兵種 0(農夫,HP=1)拿來湊隊伍,人數刻意壓低避免撞到領導力上限。
func testGameState(leadership int) *gamestate.GameState {
	gs := &gamestate.GameState{Leadership: leadership}
	gs.Army[0] = gamestate.Squad{TroopID: 0, Count: 3}
	gs.Army[1] = gamestate.Squad{TroopID: 2, Count: 2} // 義勇軍
	for i := 2; i < 5; i++ {
		gs.Army[i] = gamestate.Squad{TroopID: emptyTroopID, Count: 0}
	}
	return gs
}

func TestPrepareUnitsPlayer_Placement(t *testing.T) {
	a := loadTestAssets(t)
	gs := testGameState(1000)

	var c Combat
	PrepareUnitsPlayer(&c, 0, gs, a)
	PrepareUnitsPlayer(&c, 1, gs, a)

	for i := 0; i < MaxUnits; i++ {
		u0 := c.Units[0][i]
		u1 := c.Units[1][i]

		if u0.Y != i || u1.Y != i {
			t.Errorf("i=%d: Y = (%d,%d), want (%d,%d)", i, u0.Y, u1.Y, i, i)
		}
		if u0.X != 0 {
			t.Errorf("side0 unit %d: X = %d, want 0", i, u0.X)
		}
		if u1.X != BoardW-1 {
			t.Errorf("side1 unit %d: X = %d, want %d", i, u1.X, BoardW-1)
		}
		if u0.Count != gs.Army[i].Count {
			t.Errorf("unit %d: Count = %d, want %d", i, u0.Count, gs.Army[i].Count)
		}
		if u0.MaxCount != u0.Count {
			t.Errorf("unit %d: MaxCount = %d, want %d (== Count at setup)", i, u0.MaxCount, u0.Count)
		}
	}

	if c.Heroes[0] != gs {
		t.Errorf("Heroes[0] 應指回傳入的 gs")
	}

	// 農夫(troop 0)Spoils = 1,義勇軍(troop 2)Spoils = 5(見 tables_troops.go)。
	wantSpoils := (a.Troops[0].Spoils*5)*3 + (a.Troops[2].Spoils*5)*2
	if c.Spoils[0] != wantSpoils {
		t.Errorf("Spoils[0] = %d, want %d", c.Spoils[0], wantSpoils)
	}
}

func TestPrepareUnitsPlayer_OutOfControl(t *testing.T) {
	a := loadTestAssets(t)

	// 領導力充足:應受控。
	gsOK := testGameState(1000)
	var cOK Combat
	PrepareUnitsPlayer(&cOK, 0, gsOK, a)
	if cOK.Units[0][0].OutOfControl {
		t.Errorf("領導力充足時,unit 0 不該 OutOfControl")
	}

	// 領導力為 0:任何非空格都該脫離掌控。
	gsBad := testGameState(0)
	var cBad Combat
	PrepareUnitsPlayer(&cBad, 0, gsBad, a)
	if !cBad.Units[0][0].OutOfControl {
		t.Errorf("領導力為 0 時,unit 0 應該 OutOfControl")
	}
	// 空格(count==0)比照 C 版負負得正的行為,視為「受控」。
	if cBad.Units[0][2].OutOfControl {
		t.Errorf("空格 unit 不該被標記 OutOfControl")
	}
}

func TestPrepareUnitsFoe_OnlyThreeSlots(t *testing.T) {
	a := loadTestAssets(t)
	squads := [MaxUnits]gamestate.Squad{
		{TroopID: 0, Count: 5},
		{TroopID: 0, Count: 5},
		{TroopID: 0, Count: 5},
		{TroopID: 0, Count: 5}, // 應被忽略(C 版 max_troops 夾到 3)
		{TroopID: 0, Count: 5},
	}

	var c Combat
	PrepareUnitsFoe(&c, 1, squads, a)

	for i := 0; i < 3; i++ {
		u := c.Units[1][i]
		if u.Count != 5 {
			t.Errorf("foe slot %d: Count = %d, want 5", i, u.Count)
		}
		if u.Y != i || u.X != BoardW-1 {
			t.Errorf("foe slot %d: (X,Y) = (%d,%d), want (%d,%d)", i, u.X, u.Y, BoardW-1, i)
		}
		if u.OutOfControl {
			t.Errorf("foe slot %d 不該有 OutOfControl", i)
		}
	}
	for i := 3; i < MaxUnits; i++ {
		if c.Units[1][i].Count != 0 {
			t.Errorf("foe slot %d 超出 max_troops=3,不該被佈置,Count = %d", i, c.Units[1][i].Count)
		}
	}
	if c.Heroes[1] != nil {
		t.Errorf("foe 側 Heroes 應為 nil")
	}
}

func TestPrepareUnitsCastle_AllAtX0(t *testing.T) {
	a := loadTestAssets(t)
	squads := [MaxUnits]gamestate.Squad{
		{TroopID: 0, Count: 1},
		{TroopID: 0, Count: 1},
		{TroopID: 0, Count: 1},
		{TroopID: 0, Count: 1},
		{TroopID: 0, Count: 1},
	}

	for _, side := range []int{0, 1} {
		var c Combat
		PrepareUnitsCastle(&c, side, squads, a)
		for i := 0; i < MaxUnits; i++ {
			u := c.Units[side][i]
			if u.X != 0 {
				t.Errorf("side %d castle slot %d: X = %d, want 0", side, i, u.X)
			}
			if u.Y != i {
				t.Errorf("side %d castle slot %d: Y = %d, want %d", side, i, u.Y, i)
			}
		}
		if c.Heroes[side] != nil {
			t.Errorf("castle 側 Heroes 應為 nil")
		}
	}
}
