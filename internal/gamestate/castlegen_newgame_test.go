// 本檔用 external test package(gamestate_test),理由同 worldgen_test.go:需要
// 同時 import gamestate 與 save,而 save 套件本身 import gamestate,用 internal
// test package 會構成 import cycle。
package gamestate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/save"
)

// loadCastleWorldTestAssets 複製 land.org + castles.ini 進 t.TempDir()/free/,
// 讓 a.World != nil 且 a.Strings["castles"] != nil,足以驅動 NewGame 的城堡/惡棍
// 世界生成(castlegen.go)。對齊 loadWorldTestAssets(worldgen_test.go)的手法。
func loadCastleWorldTestAssets(t *testing.T) *kbdata.Assets {
	t.Helper()
	dir := t.TempDir()
	freeDir := filepath.Join(dir, "free")
	if err := os.MkdirAll(freeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"land.org", "castles.ini"} {
		src := filepath.Join("..", "kbdata", "testdata", name)
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("讀 %s: %v", src, err)
		}
		if err := os.WriteFile(filepath.Join(freeDir, name), b, 0o644); err != nil {
			t.Fatalf("寫 %s: %v", name, err)
		}
	}
	a, err := kbdata.Load(dir)
	if err != nil {
		t.Fatalf("kbdata.Load: %v", err)
	}
	if a.World == nil {
		t.Fatal("a.World 未載入,測試前提不成立")
	}
	if a.Strings["castles"] == nil {
		t.Fatal("castles.ini 未載入,測試前提不成立")
	}
	return a
}

// TestNewGame_CastleWorldGen_PerContinentVillainCounts 驗證 NewGame 跑完
// salt_villains 後,每洲被惡棍佔領的城堡數恰好等於 villains_per_continent[洲]
// (對齊 castles.ini 真實分佈 9/7/6/4,皆 >= 各洲需求 6/4/4/3,不會觸發「城堡不足」
// 安全閥)。
func TestNewGame_CastleWorldGen_PerContinentVillainCounts(t *testing.T) {
	a := loadCastleWorldTestAssets(t)
	castles := gamestate.LoadCastles(a)
	perContinent := kbdata.VillainsPerContinentTable()

	gs := gamestate.NewGame(a, "測試角色", 0, gamestate.DefaultWorldSeed)

	got := [kbdata.MaxContinents]int{}
	for i, owner := range gs.CastleOwner {
		if owner != 0x7F {
			got[castles[i].Continent]++
		}
	}
	for cont := 0; cont < kbdata.MaxContinents; cont++ {
		if got[cont] != int(perContinent[cont]) {
			t.Errorf("洲%d 被惡棍佔領的城堡數 = %d, want %d", cont, got[cont], perContinent[cont])
		}
	}
}

// TestNewGame_CastleWorldGen_VillainGarrisonMatchesTable 驗證每座被惡棍佔領的
// 城堡,其 CastleTroops/CastleNumbers 逐值等於 kbdata.VillainArmyTroopsTable()/
// VillainArmyNumbersTable() 對應該惡棍 id 的那一列(對齊 C salt_villains 直接
// 複製 villain_army_troops/numbers 進 castle_troops/numbers)。
func TestNewGame_CastleWorldGen_VillainGarrisonMatchesTable(t *testing.T) {
	a := loadCastleWorldTestAssets(t)
	troopsTable := kbdata.VillainArmyTroopsTable()
	numbersTable := kbdata.VillainArmyNumbersTable()

	gs := gamestate.NewGame(a, "測試角色", 0, gamestate.DefaultWorldSeed)

	checked := 0
	for i, owner := range gs.CastleOwner {
		if owner == 0x7F {
			continue
		}
		villainID := int(owner)
		if villainID < 0 || villainID >= kbdata.MaxVillains {
			t.Errorf("castle%d owner=%d 超出惡棍 id 範圍 [0,%d]", i, villainID, kbdata.MaxVillains-1)
			continue
		}
		want := troopsTable[villainID]
		for k := 0; k < 5; k++ {
			if gs.CastleTroops[i][k] != int(want[k]) {
				t.Errorf("castle%d(惡棍%d)守軍[%d]: got %d, want %d", i, villainID, k, gs.CastleTroops[i][k], want[k])
			}
			if gs.CastleNumbers[i][k] != numbersTable[villainID][k] {
				t.Errorf("castle%d(惡棍%d)數量[%d]: got %d, want %d", i, villainID, k, gs.CastleNumbers[i][k], numbersTable[villainID][k])
			}
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("沒有任何惡棍佔領城堡可供比對,測試前提不成立")
	}
	t.Logf("比對了 %d 座惡棍城堡的守軍資料", checked)
}

// TestNewGame_CastleWorldGen_MonsterCastlesRepopulated 驗證未被惡棍佔領
// (castle_owner 仍是 0x7F)的城堡,經 repopulate_castle 後守軍非空(至少一格
// 數量 > 0),對齊 C spawn_game「Populate all castles owned by monsters」。
func TestNewGame_CastleWorldGen_MonsterCastlesRepopulated(t *testing.T) {
	a := loadCastleWorldTestAssets(t)
	gs := gamestate.NewGame(a, "測試角色", 0, gamestate.DefaultWorldSeed)

	checked := 0
	for i, owner := range gs.CastleOwner {
		if owner != 0x7F {
			continue
		}
		if gs.CastleNumbers[i][0] <= 0 {
			t.Errorf("castle%d(怪物守軍)CastleNumbers[0] = %d, want > 0(repopulateCastle 應已填入)", i, gs.CastleNumbers[i][0])
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("沒有任何怪物城堡可供比對,測試前提不成立(理論上 26-17=9 座)")
	}
	t.Logf("比對了 %d 座怪物城堡的守軍非空", checked)
}

// TestNewGame_Contract_InitialValues 驗證契約起手值對齊 C spawn_game(play.c:418-425):
// contract=0xFF(未接)、last_contract=0x04、max_contract=0x05、contract_cycle={0,1,2,3,4}。
func TestNewGame_Contract_InitialValues(t *testing.T) {
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatalf("kbdata.Load: %v", err)
	}
	gs := gamestate.NewGame(a, "測試角色", 0, gamestate.DefaultWorldSeed)

	if gs.Contract != 0xFF {
		t.Errorf("Contract = 0x%02X, want 0xFF", gs.Contract)
	}
	if gs.LastContract != 0x04 {
		t.Errorf("LastContract = 0x%02X, want 0x04", gs.LastContract)
	}
	if gs.MaxContract != 0x05 {
		t.Errorf("MaxContract = 0x%02X, want 0x05", gs.MaxContract)
	}
	want := [5]byte{0, 1, 2, 3, 4}
	if gs.ContractCycle != want {
		t.Errorf("ContractCycle = %v, want %v", gs.ContractCycle, want)
	}
}

// TestNewGame_CastleWorldGen_SaveLoadRoundTrip 驗證城堡/惡棍/契約世界狀態完整
// 隨 JSON 存讀檔往返(對齊 worldgen_test.go 的 SaveLoadRoundTrip 測試,同一套
// 「全部欄位已 exported,不需擴充存檔格式」保證)。
func TestNewGame_CastleWorldGen_SaveLoadRoundTrip(t *testing.T) {
	a := loadCastleWorldTestAssets(t)
	gs := gamestate.NewGame(a, "測試角色", 0, gamestate.DefaultWorldSeed)

	path := filepath.Join(t.TempDir(), "slot0.json")
	if err := save.SaveToPath(gs, path); err != nil {
		t.Fatalf("SaveToPath: %v", err)
	}
	loaded, err := save.LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	if loaded.CastleOwner != gs.CastleOwner {
		t.Error("CastleOwner 存讀檔前後不一致")
	}
	if loaded.CastleTroops != gs.CastleTroops {
		t.Error("CastleTroops 存讀檔前後不一致")
	}
	if loaded.CastleNumbers != gs.CastleNumbers {
		t.Error("CastleNumbers 存讀檔前後不一致")
	}
	if loaded.Contract != gs.Contract || loaded.LastContract != gs.LastContract ||
		loaded.MaxContract != gs.MaxContract || loaded.ContractCycle != gs.ContractCycle {
		t.Error("Contract 系列欄位存讀檔前後不一致")
	}
}
