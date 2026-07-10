package gamestate

import "testing"

// TestTakeArtifactDoubleLeadership 驗證 ordered = continent*2+num 對到 id2
// (double_leadership, invert_id=4):continent=2, num=0 → ordered=4。
func TestTakeArtifactDoubleLeadership(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.Continent = 2

	base := gs.BaseLeadership
	p := gs.TakeArtifact(a, 0)

	if !p.Found {
		t.Fatalf("TakeArtifact(num=0) Found = false, want true")
	}
	if gs.ArtifactFound[2] != 1 {
		t.Errorf("ArtifactFound[2] = %d, want 1", gs.ArtifactFound[2])
	}
	if gs.BaseLeadership != base*2 {
		t.Errorf("BaseLeadership = %d, want %d (base %d *2)", gs.BaseLeadership, base*2, base)
	}
	if gs.Leadership != gs.BaseLeadership {
		t.Errorf("Leadership = %d, want %d (== BaseLeadership)", gs.Leadership, gs.BaseLeadership)
	}
}

// TestTakeArtifactDoubleSpellPower 驗證 id4(double_spell_power, invert_id=6):
// continent=3, num=0 → ordered=6。
func TestTakeArtifactDoubleSpellPower(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.Continent = 3

	sp := gs.SpellPower
	p := gs.TakeArtifact(a, 0)

	if !p.Found {
		t.Fatalf("TakeArtifact(num=0) Found = false, want true")
	}
	if gs.ArtifactFound[4] != 1 {
		t.Errorf("ArtifactFound[4] = %d, want 1", gs.ArtifactFound[4])
	}
	if gs.SpellPower != sp*2 {
		t.Errorf("SpellPower = %d, want %d (base %d *2)", gs.SpellPower, sp*2, sp)
	}
}

// TestTakeArtifactIncreaseCommission 驗證 id3(increase_commission, invert_id=1):
// continent=0, num=1 → ordered=1。
func TestTakeArtifactIncreaseCommission(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.Continent = 0

	commission := gs.Commission
	p := gs.TakeArtifact(a, 1)

	if !p.Found {
		t.Fatalf("TakeArtifact(num=1) Found = false, want true")
	}
	if gs.ArtifactFound[3] != 1 {
		t.Errorf("ArtifactFound[3] = %d, want 1", gs.ArtifactFound[3])
	}
	if gs.Commission != commission+2000 {
		t.Errorf("Commission = %d, want %d (base %d +2000)", gs.Commission, commission+2000, commission)
	}
}

// TestHasPowerAndBoatCostWith 驗證拾得海權之錨(id7, cheaper_boat_rental)前後,
// HasPower/BoatCostWith 的租船費用切換。
func TestHasPowerAndBoatCostWith(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)

	if gs.HasPower(a, PowerCheaperBoatRental) {
		t.Fatalf("HasPower(PowerCheaperBoatRental) = true before pickup, want false")
	}
	if got := gs.BoatCostWith(a); got != CostBoatExpensive {
		t.Errorf("BoatCostWith = %d, want CostBoatExpensive(%d)", got, CostBoatExpensive)
	}

	gs.ArtifactFound[7] = 1

	if !gs.HasPower(a, PowerCheaperBoatRental) {
		t.Fatalf("HasPower(PowerCheaperBoatRental) = false after pickup, want true")
	}
	if got := gs.BoatCostWith(a); got != CostBoatCheap {
		t.Errorf("BoatCostWith = %d, want CostBoatCheap(%d)", got, CostBoatCheap)
	}
}
