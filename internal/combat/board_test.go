package combat

import "testing"

func TestBoardDimensions(t *testing.T) {
	if BoardW != 6 {
		t.Fatalf("BoardW = %d, want 6 (CLEVEL_W)", BoardW)
	}
	if BoardH != 5 {
		t.Fatalf("BoardH = %d, want 5 (CLEVEL_H)", BoardH)
	}
	if MaxSides != 2 {
		t.Fatalf("MaxSides = %d, want 2", MaxSides)
	}
	if MaxUnits != 5 {
		t.Fatalf("MaxUnits = %d, want 5", MaxUnits)
	}
}

func TestInBounds(t *testing.T) {
	cases := []struct {
		x, y int
		want bool
	}{
		{0, 0, true},
		{BoardW - 1, BoardH - 1, true},
		{-1, 0, false},
		{0, -1, false},
		{BoardW, 0, false},
		{0, BoardH, false},
	}
	for _, c := range cases {
		if got := InBounds(c.x, c.y); got != c.want {
			t.Errorf("InBounds(%d,%d) = %v, want %v", c.x, c.y, got, c.want)
		}
	}
}

func TestAdjacent(t *testing.T) {
	cases := []struct {
		x1, y1, x2, y2 int
		want           bool
	}{
		{0, 0, 0, 0, true},  // 同格
		{0, 0, 1, 0, true},  // 直向相鄰
		{0, 0, 0, 1, true},  // 直向相鄰
		{0, 0, 1, 1, true},  // 對角相鄰
		{0, 0, 2, 0, false}, // 隔一格
		{0, 0, 1, 2, false}, // 超出 Chebyshev 1
	}
	for _, c := range cases {
		if got := Adjacent(c.x1, c.y1, c.x2, c.y2); got != c.want {
			t.Errorf("Adjacent(%d,%d,%d,%d) = %v, want %v", c.x1, c.y1, c.x2, c.y2, got, c.want)
		}
	}
}

func TestDistance(t *testing.T) {
	cases := []struct {
		x1, y1, x2, y2 int
		want           int
	}{
		{0, 0, 0, 0, 0},
		{0, 0, 3, 4, 7}, // 曼哈頓,非畢氏(畢氏會是 5)
		{2, 2, 0, 0, 4},
	}
	for _, c := range cases {
		if got := Distance(c.x1, c.y1, c.x2, c.y2); got != c.want {
			t.Errorf("Distance(%d,%d,%d,%d) = %d, want %d", c.x1, c.y1, c.x2, c.y2, got, c.want)
		}
	}
}

func TestPackUnpackUID(t *testing.T) {
	for side := 0; side < MaxSides; side++ {
		for id := 0; id < MaxUnits; id++ {
			uid := PackUID(side, id)
			gotSide, gotID := UnpackUID(uid)
			if gotSide != side || gotID != id {
				t.Errorf("UnpackUID(PackUID(%d,%d)=%d) = (%d,%d), want (%d,%d)",
					side, id, uid, gotSide, gotID, side, id)
			}
		}
	}
}
