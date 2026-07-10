package kbdata

import "testing"

// TestMaxVillains 對齊 bounty.h:46 `#define MAX_VILLAINS 17`。
func TestMaxVillains(t *testing.T) {
	if MaxVillains != 17 {
		t.Errorf("MaxVillains: got %d, want 17", MaxVillains)
	}
}

// TestVillainsPerContinentTable 逐值比對 villains_per_continent(bounty.c:220)。
func TestVillainsPerContinentTable(t *testing.T) {
	table := VillainsPerContinentTable()
	want := [MaxContinents]byte{6, 4, 4, 3}
	if table != want {
		t.Errorf("VillainsPerContinentTable: got %v, want %v", table, want)
	}
	sum := 0
	for _, v := range table {
		sum += int(v)
	}
	if sum != MaxVillains {
		t.Errorf("villains_per_continent 總和 = %d, want MaxVillains=%d", sum, MaxVillains)
	}
}

// TestVillainArmyTroopsTable 抽樣比對 villain_army_troops[MAX_VILLAINS][5](bounty.c:247)。
func TestVillainArmyTroopsTable(t *testing.T) {
	table := VillainArmyTroopsTable()
	if len(table) != MaxVillains {
		t.Fatalf("VillainArmyTroopsTable rows: got %d, want %d", len(table), MaxVillains)
	}
	want0 := [5]byte{0x00, 0x03, 0x02, 0x00, 0x00}
	if table[0] != want0 {
		t.Errorf("VillainArmyTroopsTable[0]: got %v, want %v", table[0], want0)
	}
	want9 := [5]byte{0x04, 0x05, 0x0d, 0x15, 0x17}
	if table[9] != want9 {
		t.Errorf("VillainArmyTroopsTable[9]: got %v, want %v", table[9], want9)
	}
	// villain 9 與 villain 10 字面完全相同——C 原始資料本身如此(bounty.c:257-258),
	// 照抄不推斷/不「修正」。
	if table[9] != table[10] {
		t.Errorf("VillainArmyTroopsTable[9] 應等於 [10](C 原始資料如此):%v vs %v", table[9], table[10])
	}
	want16 := [5]byte{0x18, 0x18, 0x18, 0x17, 0x15}
	if table[16] != want16 {
		t.Errorf("VillainArmyTroopsTable[16]: got %v, want %v", table[16], want16)
	}
}

// TestVillainArmyNumbersTable 抽樣比對 villain_army_numbers[MAX_VILLAINS][5](bounty.c:266)。
func TestVillainArmyNumbersTable(t *testing.T) {
	table := VillainArmyNumbersTable()
	if len(table) != MaxVillains {
		t.Fatalf("VillainArmyNumbersTable rows: got %d, want %d", len(table), MaxVillains)
	}
	want0 := [5]int{50, 20, 25, 30, 25}
	if table[0] != want0 {
		t.Errorf("VillainArmyNumbersTable[0]: got %v, want %v", table[0], want0)
	}
	want14 := [5]int{30, 50, 100, 500, 5000} // 最大值 5000(智者索利度斯 army4)
	if table[14] != want14 {
		t.Errorf("VillainArmyNumbersTable[14]: got %v, want %v", table[14], want14)
	}
	want16 := [5]int{100, 25, 25, 100, 100}
	if table[16] != want16 {
		t.Errorf("VillainArmyNumbersTable[16]: got %v, want %v", table[16], want16)
	}
}
