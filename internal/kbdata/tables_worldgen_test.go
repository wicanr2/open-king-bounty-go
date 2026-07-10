package kbdata

import "testing"

// TestWorldgenConstants 驗證 MAX_ 維度常數與 C 版 bounty.h 對齊。
func TestWorldgenConstants(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"MaxCastles", MaxCastles, 26},
		{"MaxDwellings", MaxDwellings, 11},
		{"MaxTroops", MaxTroops, 25},
		{"MaxTroopDifficulty", MaxTroopDifficulty, 4},
		{"MaxTroopChanceCurve", MaxTroopChanceCurve, 5},
		{"MaxContinents (沿用 land.go)", MaxContinents, 4},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}

// TestCastleDifficultyTable 抽樣比對 castle_difficulty[MAX_CASTLES](bounty.c:612)。
func TestCastleDifficultyTable(t *testing.T) {
	table := castleDifficultyTable()
	if len(table) != MaxCastles {
		t.Fatalf("castleDifficultyTable length: got %d, want %d", len(table), MaxCastles)
	}
	tests := []struct {
		idx  int
		want int
	}{
		{0, 0},
		{4, 2},  // 第 5 筆 = 2
		{18, 3}, // 第 19 筆 = 3(表中第一個出現的難度3)
		{25, 3}, // 最後一筆
	}
	for _, tt := range tests {
		if table[tt.idx] != tt.want {
			t.Errorf("castleDifficultyTable[%d]: got %d, want %d", tt.idx, table[tt.idx], tt.want)
		}
	}
}

// TestTroopChanceTableTable 抽樣比對 troop_chance_table(bounty.c:764)。
func TestTroopChanceTableTable(t *testing.T) {
	table := troopChanceTableTable()
	if len(table) != MaxTroopDifficulty {
		t.Fatalf("troopChanceTableTable rows: got %d, want %d", len(table), MaxTroopDifficulty)
	}
	for i, row := range table {
		if len(row) != MaxTroopChanceCurve-1 {
			t.Errorf("troopChanceTableTable row %d cols: got %d, want %d", i, len(row), MaxTroopChanceCurve-1)
		}
	}
	want0 := [4]byte{0x3C, 0x5A, 0x62, 0x65}
	if table[0] != want0 {
		t.Errorf("troopChanceTableTable[0]: got %v, want %v", table[0], want0)
	}
	// 任務範例明確指定的斷言值:troop_chance_table[3] = {0x03,0x06,0x0A,0x28}
	want3 := [4]byte{0x03, 0x06, 0x0A, 0x28}
	if table[3] != want3 {
		t.Errorf("troopChanceTableTable[3]: got %v, want %v", table[3], want3)
	}
}

// TestDwellingToTroopTable 抽樣比對 dwelling_to_troop(bounty.c:770)。
func TestDwellingToTroopTable(t *testing.T) {
	table := dwellingToTroopTable()
	if len(table) != MaxTroopDifficulty {
		t.Fatalf("dwellingToTroopTable rows: got %d, want %d", len(table), MaxTroopDifficulty)
	}
	for i, row := range table {
		if len(row) != MaxTroopChanceCurve {
			t.Errorf("dwellingToTroopTable row %d cols: got %d, want %d", i, len(row), MaxTroopChanceCurve)
		}
	}
	// 任務範例明確指定的斷言值:dwelling_to_troop[0] = {0x00,0x03,0x0b,0x10,0x14}(Plains)
	want0 := [5]byte{0x00, 0x03, 0x0b, 0x10, 0x14}
	if table[0] != want0 {
		t.Errorf("dwellingToTroopTable[0] (Plains): got %v, want %v", table[0], want0)
	}
	want3 := [5]byte{0x04, 0x05, 0x0d, 0x15, 0x17} // Dungeon
	if table[3] != want3 {
		t.Errorf("dwellingToTroopTable[3] (Dungeon): got %v, want %v", table[3], want3)
	}
}

// TestTroopNumbersTable 抽樣比對 troop_numbers[MAX_TROOPS][MAX_TROOP_DIFFICULTY](bounty.c:780)。
func TestTroopNumbersTable(t *testing.T) {
	table := troopNumbersTable()
	if len(table) != MaxTroops {
		t.Fatalf("troopNumbersTable rows: got %d, want %d", len(table), MaxTroops)
	}
	for i, row := range table {
		if len(row) != MaxTroopDifficulty {
			t.Errorf("troopNumbersTable row %d cols: got %d, want %d", i, len(row), MaxTroopDifficulty)
		}
	}
	want0 := [4]byte{0x0A, 0x14, 0x32, 0x64} // id0 農夫
	if table[0] != want0 {
		t.Errorf("troopNumbersTable[0]: got %v, want %v", table[0], want0)
	}
	want24 := [4]byte{0x01, 0x01, 0x01, 0x02} // id24 火龍
	if table[24] != want24 {
		t.Errorf("troopNumbersTable[24]: got %v, want %v", table[24], want24)
	}
	// 全 0 列:僅招募於城堡、不出現在野外棲地的兵種(id 8/10/14/18)
	for _, idx := range []int{8, 10, 14, 18} {
		want := [4]byte{0, 0, 0, 0}
		if table[idx] != want {
			t.Errorf("troopNumbersTable[%d]: got %v, want all-zero %v", idx, table[idx], want)
		}
	}
	// 對照組:id2(義勇軍)同為城堡兵種,但 troop_numbers 非零——C 原始資料如此,非零項不可誤删。
	want2 := [4]byte{0x0A, 0x14, 0x32, 0x64}
	if table[2] != want2 {
		t.Errorf("troopNumbersTable[2]: got %v, want %v", table[2], want2)
	}
}

// TestContinentDwellingsTable 抽樣比對 continent_dwellings(bounty.c:808),
// 含 C 靜態陣列零填部分(不是只有 0xFF 之前的顯式值)。
func TestContinentDwellingsTable(t *testing.T) {
	table := continentDwellingsTable()
	if len(table) != MaxContinents {
		t.Fatalf("continentDwellingsTable rows: got %d, want %d", len(table), MaxContinents)
	}
	for i, row := range table {
		if len(row) != MaxDwellings {
			t.Errorf("continentDwellingsTable row %d cols: got %d, want %d", i, len(row), MaxDwellings)
		}
	}
	want0 := [11]byte{0x00, 0x01, 0x07, 0x04, 0x03, 0x06, 0xFF, 0x00, 0x00, 0x00, 0x00}
	if table[0] != want0 {
		t.Errorf("continentDwellingsTable[0]: got %v, want %v", table[0], want0)
	}
	want2 := [11]byte{0x0D, 0x10, 0x11, 0x13, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if table[2] != want2 {
		t.Errorf("continentDwellingsTable[2]: got %v, want %v", table[2], want2)
	}
}

// TestDwellingRangesTable 抽樣比對 dwelling_ranges[MAX_CONTINENTS][2](bounty.c:814)。
func TestDwellingRangesTable(t *testing.T) {
	table := dwellingRangesTable()
	if len(table) != MaxContinents {
		t.Fatalf("dwellingRangesTable rows: got %d, want %d", len(table), MaxContinents)
	}
	want := [MaxContinents][2]byte{
		{0x00, 0x0e},
		{0x01, 0x0e},
		{0x02, 0x0e},
		{0x14, 0x18},
	}
	if table != want {
		t.Errorf("dwellingRangesTable: got %v, want %v", table, want)
	}
}
