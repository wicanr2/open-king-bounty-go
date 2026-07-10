package gamestate

import "testing"

// TestVisitTelecave 驗證站在傳送洞一端 → 瞬移到另一端;不在任何洞口 → 不動回 false。
func TestVisitTelecave(t *testing.T) {
	gs := &GameState{}
	gs.Continent = 0
	gs.TelecaveCoords[0][0] = [2]int{5, 6}
	gs.TelecaveCoords[0][1] = [2]int{20, 30}

	// 站在端 0 → 瞬移到端 1。
	gs.X, gs.Y = 5, 6
	if !gs.VisitTelecave() {
		t.Fatal("站在傳送洞端 0 應瞬移(回 true)")
	}
	if gs.X != 20 || gs.Y != 30 {
		t.Errorf("未瞬移到端 1:got (%d,%d), want (20,30)", gs.X, gs.Y)
	}

	// 站在端 1 → 瞬移回端 0。
	gs.X, gs.Y = 20, 30
	if !gs.VisitTelecave() || gs.X != 5 || gs.Y != 6 {
		t.Errorf("端 1 未瞬移回端 0:(%d,%d)", gs.X, gs.Y)
	}

	// 不在任何洞口 → 不動。
	gs.X, gs.Y = 1, 1
	if gs.VisitTelecave() {
		t.Error("非洞口不應瞬移")
	}
	if gs.X != 1 || gs.Y != 1 {
		t.Errorf("非洞口座標被改動:(%d,%d)", gs.X, gs.Y)
	}
}
