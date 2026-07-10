package gamestate

import "testing"

// TestLearnMagicAtAlcove_EnoughGold 驗證金幣足時扣 CostAlcove、習得魔法、回 true。
func TestLearnMagicAtAlcove_EnoughGold(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.Gold = 6000
	gs.KnowsMagic = false

	if got := gs.LearnMagicAtAlcove(); !got {
		t.Fatalf("LearnMagicAtAlcove() = %v, want true(金幣足)", got)
	}
	if gs.Gold != 1000 {
		t.Errorf("Gold = %d, want 1000(6000-%d)", gs.Gold, CostAlcove)
	}
	if !gs.KnowsMagic {
		t.Errorf("KnowsMagic = false, want true")
	}
}

// TestLearnMagicAtAlcove_NotEnoughGold 驗證金幣不足時不扣錢、不習得魔法、回 false。
func TestLearnMagicAtAlcove_NotEnoughGold(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.Gold = 100
	gs.KnowsMagic = false

	if got := gs.LearnMagicAtAlcove(); got {
		t.Fatalf("LearnMagicAtAlcove() = %v, want false(金幣不足)", got)
	}
	if gs.Gold != 100 {
		t.Errorf("Gold 被改動 = %d, want 100(不變)", gs.Gold)
	}
	if gs.KnowsMagic {
		t.Errorf("KnowsMagic = true, want false(不變)")
	}
}

// TestAtAlcove 驗證站在魔法密室座標(land.ini special1)上判 true,離開判 false。
func TestAtAlcove(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)

	c, x, y, ok := AlcoveCoords(a)
	if !ok {
		t.Skip("讀不到魔法密室座標(land.ini special1),略過")
	}

	gs.Continent, gs.X, gs.Y = c, x, y
	if !gs.AtAlcove(a) {
		t.Fatalf("AtAlcove() = false, want true(站在密室座標上)")
	}

	gs.X++
	if gs.AtAlcove(a) {
		t.Errorf("AtAlcove() = true, want false(已離開密室座標)")
	}
}
