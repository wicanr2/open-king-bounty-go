package kbdata

import "testing"

// TestTroopTableLength 驗證兵種表有完整 25 筆。
func TestTroopTableLength(t *testing.T) {
	troops := troopTable()
	if len(troops) != 25 {
		t.Errorf("troopTable length: got %d, want 25", len(troops))
	}
}

// TestTroopZeroPeasant 驗證第 0 筆(農夫)與 C 版對齊。
func TestTroopZeroPeasant(t *testing.T) {
	troops := troopTable()
	peasant := troops[0]

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Name", peasant.Name, "農夫"},
		{"SkillLevel", peasant.SkillLevel, 1},
		{"HP", peasant.HP, 1},
		{"Move", peasant.Move, 1},
		{"MeleeMin", peasant.MeleeMin, 1},
		{"MeleeMax", peasant.MeleeMax, 1},
		{"GoldCost", peasant.GoldCost, 10},
		{"Spoils", peasant.Spoils, 1},
		{"Home", peasant.Home, DwellPlains},
		{"Abilities", peasant.Abilities, Ability(0)},
		{"MaxPopulation", peasant.MaxPopulation, 250},
		{"Growth", peasant.Growth, 6},
		{"MoraleGroup", peasant.MoraleGroup, 0},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("peasant %s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestTroopLastDragon 驗證最後一筆(火龍)與 C 版對齊。
func TestTroopLastDragon(t *testing.T) {
	troops := troopTable()
	dragon := troops[24]

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Name", dragon.Name, "火龍"},
		{"SkillLevel", dragon.SkillLevel, 6},
		{"HP", dragon.HP, 200},
		{"Move", dragon.Move, 1},
		{"MeleeMin", dragon.MeleeMin, 25},
		{"MeleeMax", dragon.MeleeMax, 50},
		{"GoldCost", dragon.GoldCost, 5000},
		{"Spoils", dragon.Spoils, 500},
		{"Home", dragon.Home, DwellHillCave},
		{"Abilities", dragon.Abilities, AbilFly | AbilImmune},
		{"MaxPopulation", dragon.MaxPopulation, 25},
		{"Growth", dragon.Growth, 1},
		{"MoraleGroup", dragon.MoraleGroup, 3},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("dragon %s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestClassTableDimensions 驗證職業表是 4×4 結構。
func TestClassTableDimensions(t *testing.T) {
	classes := classTable()
	if len(classes) != 4 {
		t.Errorf("classTable rows: got %d, want 4", len(classes))
	}
	for i, row := range classes {
		if len(row) != 4 {
			t.Errorf("classTable row %d: got %d cols, want 4", i, len(row))
		}
	}
}

// TestClassKnight 驗證第 0 行 0 列(騎士系武士)與 C 版對齊。
func TestClassKnight(t *testing.T) {
	classes := classTable()
	knight := classes[0][0]

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Name", knight.Name, "武士"},
		{"Leadership", knight.Leadership, 100},
		{"Commission", knight.Commission, 1000},
		{"MaxSpell", knight.MaxSpell, 2},
		{"SpellPower", knight.SpellPower, 1},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("knight %s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestClassSorceressLast 驗證第 2 行 3 列(女巫師系大法師)與 C 版對齊。
// 值為 bounty.c 字面**逐階增量**(非累積絕對值);累加語意由 gamestate.acceptRank 負責。
func TestClassSorceressLast(t *testing.T) {
	classes := classTable()
	archMage := classes[2][3]

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Name", archMage.Name, "大法師"},
		{"Leadership", archMage.Leadership, 300},  // C 增量 +300(累積後 600)
		{"Commission", archMage.Commission, 1000}, // C 增量 +1000(累積後 6000)
		{"MaxSpell", archMage.MaxSpell, 12},       // C 增量 +12(累積後 35)
		{"SpellPower", archMage.SpellPower, 5},    // C 增量 +5(累積後 15)
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("archMage %s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}
