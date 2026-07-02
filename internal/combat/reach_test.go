package combat

import "testing"

func TestCanReach_SameTileAlwaysTrue(t *testing.T) {
	var c Combat
	c.Units[0][0] = Unit{X: 2, Y: 2, Moves: 0}
	if !c.CanReach(0, 0, 2, 2) {
		t.Errorf("原地(0 步)應永遠可達")
	}
}

func TestCanReach_EmptyBoardWithinMoves(t *testing.T) {
	var c Combat
	c.Units[0][0] = Unit{X: 0, Y: 0, Moves: 2}

	if !c.CanReach(0, 0, 2, 0) {
		t.Errorf("空棋盤、Moves=2,水平走 2 格應可達 (2,0)")
	}
	if !c.CanReach(0, 0, 1, 1) {
		t.Errorf("空棋盤、Moves=2,King-move 1 步即可達對角 (1,1)")
	}
}

func TestCanReach_OutOfMoves(t *testing.T) {
	var c Combat
	c.Units[0][0] = Unit{X: 0, Y: 0, Moves: 2}

	if c.CanReach(0, 0, 0, 4) {
		t.Errorf("Moves=2 不該走到 4 步遠的 (0,4)")
	}
}

func TestCanReach_OutOfBoundsTarget(t *testing.T) {
	var c Combat
	c.Units[0][0] = Unit{X: 0, Y: 0, Moves: 5}

	if c.CanReach(0, 0, BoardW, 0) {
		t.Errorf("目標超出棋盤,不該可達")
	}
}

func TestCanReach_BlockedByObstacle(t *testing.T) {
	var c Combat
	c.Units[0][0] = Unit{X: 0, Y: 2, Moves: 3}

	// 在 x=1 整欄放障礙,把左半邊完全封死(棋盤只有 5 列,y=0..4)。
	for y := 0; y < BoardH; y++ {
		c.Omap[y][1] = 1
	}

	if c.CanReach(0, 0, 2, 2) {
		t.Errorf("整欄障礙應完全擋住通往 (2,2) 的路")
	}
	// 但同一格(未跨越障礙)仍可達。
	if !c.CanReach(0, 0, 0, 2) {
		t.Errorf("原地應仍可達")
	}
}

func TestCanReach_BlockedByOccupiedUnit(t *testing.T) {
	var c Combat
	c.Units[0][0] = Unit{X: 0, Y: 0, Moves: 3}
	// 目標格被其他單位佔據(純幾何:不能通過也不能停在被佔用格,攻擊互動不在此判定)。
	c.Umap[0][1] = PackUID(1, 0)

	if c.CanReach(0, 0, 1, 0) {
		t.Errorf("目標格被佔用,不該視為可達(移動幾何不含攻擊)")
	}
	// 繞路仍可達更遠但沒被擋住的格子。
	if !c.CanReach(0, 0, 0, 2) {
		t.Errorf("未被佔用的格子仍應可達")
	}
}
