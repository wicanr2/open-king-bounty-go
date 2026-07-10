package gamestate

import "testing"

// TestLoadCastles_RealValues 核對 castles.ini(free 真值,非 bounty.c 硬編座標,見
// castle.go 檔頭核實記錄)castle0/castle25 的欄位,以及 Difficulty 改查
// kbdata.CastleDifficultyTable()(bounty.c 硬編,不受 ini 覆寫)。
func TestLoadCastles_RealValues(t *testing.T) {
	a := loadTestAssets(t, "castles")
	castles := LoadCastles(a)

	if len(castles) != 26 {
		t.Fatalf("len(castles) = %d, want 26", len(castles))
	}

	c0 := castles[0]
	if c0.Name != "阿克羅波斯" || c0.Continent != 0 || c0.X != 44 || c0.Y != 12 {
		t.Errorf("castle0 = %+v, want {阿克羅波斯 0 44 12 ...}", c0)
	}
	if c0.Difficulty != 0 {
		t.Errorf("castle0.Difficulty = %d, want 0(kbdata.CastleDifficultyTable()[0])", c0.Difficulty)
	}

	c25 := castles[25]
	if c25.Name != "昭福" || c25.Continent != 3 || c25.X != 29 || c25.Y != 35 {
		t.Errorf("castle25 = %+v, want {昭福 3 29 35 ...}", c25)
	}
	if c25.Difficulty != 3 {
		t.Errorf("castle25.Difficulty = %d, want 3(kbdata.CastleDifficultyTable()[25])", c25.Difficulty)
	}
}

// TestLoadCastles_MissingIni 驗證缺 castles.ini 時退回 best-effort 零值(不 panic),
// 對齊 town.go LoadTowns 的慣例。
func TestLoadCastles_MissingIni(t *testing.T) {
	castles := LoadCastles(nil)
	if len(castles) != 26 {
		t.Fatalf("len(castles) = %d, want 26", len(castles))
	}
	if castles[0].Name != "" || castles[0].Continent != 0 {
		t.Errorf("castles[0] = %+v, want 零值(除 ID/Difficulty)", castles[0])
	}
	// Difficulty 不依賴 ini,缺 ini 仍應正確填入。
	if castles[4].Difficulty != 2 {
		t.Errorf("castles[4].Difficulty = %d, want 2(缺 ini 不影響 Difficulty 來源)", castles[4].Difficulty)
	}
}

// TestNumCastles 核對 num_castles(continent)(play.c:128)與真實 castles.ini
// 分佈一致:洲別 0/1/2/3 各 9/7/6/4 座(2026-07-10 手動 grep 核對,合計 26)。
func TestNumCastles(t *testing.T) {
	a := loadTestAssets(t, "castles")
	castles := LoadCastles(a)

	tests := []struct {
		continent int
		want      int
	}{
		{0, 9},
		{1, 7},
		{2, 6},
		{3, 4},
	}
	total := 0
	for _, tt := range tests {
		got := numCastles(castles, tt.continent)
		if got != tt.want {
			t.Errorf("numCastles(%d) = %d, want %d", tt.continent, got, tt.want)
		}
		total += got
	}
	if total != 26 {
		t.Errorf("四洲城堡數合計 = %d, want 26", total)
	}
}
