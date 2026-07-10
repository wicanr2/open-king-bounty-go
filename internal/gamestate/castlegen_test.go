package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// TestSaltVillains_DirectAlgorithm 繞過 NewGame,直接測 saltVillains 本體(對齊
// townspell_test.go TestSaltSpells_DirectAlgorithm 的手法):驗證每洲恰好放入
// villains_per_continent[continent] 個惡棍、且都落在該洲的城堡上、守軍資料
// 逐值等於 kbdata.VillainArmyTroopsTable()/VillainArmyNumbersTable()。
func TestSaltVillains_DirectAlgorithm(t *testing.T) {
	a := loadTestAssets(t, "castles")
	castles := LoadCastles(a)
	perContinent := kbdata.VillainsPerContinentTable()
	troopsTable := kbdata.VillainArmyTroopsTable()
	numbersTable := kbdata.VillainArmyNumbersTable()

	for _, seed := range []uint32{1, 2, 42, 12345} {
		gs := &GameState{}
		for i := 0; i < kbdata.MaxCastles; i++ {
			gs.CastleOwner[i] = castleOwnerMonsters
		}
		rng := kbrng.NewGlibc(seed)

		base := 0
		for cont := 0; cont < kbdata.MaxContinents; cont++ {
			base = gs.saltVillains(rng, castles, cont, base)
		}
		if base != kbdata.MaxVillains {
			t.Errorf("seed=%d: 四洲合計放入 %d 個惡棍, want %d(villains_per_continent 總和)", seed, base, kbdata.MaxVillains)
		}

		gotPerContinent := [kbdata.MaxContinents]int{}
		for i, owner := range gs.CastleOwner {
			if owner == castleOwnerMonsters {
				continue
			}
			villainID := int(owner)
			cont := castles[i].Continent
			gotPerContinent[cont]++

			wantTroops := troopsTable[villainID]
			wantNumbers := numbersTable[villainID]
			for k := 0; k < 5; k++ {
				if gs.CastleTroops[i][k] != int(wantTroops[k]) {
					t.Errorf("seed=%d castle%d(惡棍%d)守軍[%d]: got %d, want %d",
						seed, i, villainID, k, gs.CastleTroops[i][k], wantTroops[k])
				}
				if gs.CastleNumbers[i][k] != wantNumbers[k] {
					t.Errorf("seed=%d castle%d(惡棍%d)數量[%d]: got %d, want %d",
						seed, i, villainID, k, gs.CastleNumbers[i][k], wantNumbers[k])
				}
			}
		}
		for cont := 0; cont < kbdata.MaxContinents; cont++ {
			if gotPerContinent[cont] != int(perContinent[cont]) {
				t.Errorf("seed=%d 洲%d 惡棍城堡數 = %d, want %d", seed, cont, gotPerContinent[cont], perContinent[cont])
			}
		}
	}
}

// TestRepopulateCastle_FillsFiveSlots 驗證 repopulateCastle 對指定城堡填滿 5 格
// (troopID/count 皆非零值 pattern,對齊 rollCreature 的 `number<=1 → 2` 保底)。
func TestRepopulateCastle_FillsFiveSlots(t *testing.T) {
	a := loadTestAssets(t, "castles")
	castles := LoadCastles(a)
	rng := kbrng.NewGlibc(7)

	gs := &GameState{}
	const castleID = 0
	gs.repopulateCastle(rng, castles, castleID)

	for i := 0; i < 5; i++ {
		if gs.CastleNumbers[castleID][i] < 2 {
			t.Errorf("CastleNumbers[%d][%d] = %d, want >= 2(rollCreature 保底)", castleID, i, gs.CastleNumbers[castleID][i])
		}
	}
}

// TestSaltVillains_NotEnoughCastles 驗證城堡數不足該洲惡棍需求時,對齊 C 的安全閥
// (KB_errlog + return base_id,不放任何惡棍),不 panic、不無窮迴圈。
//
// saltVillains 內部用 kbdata.MaxCastles-1 當 rng.Between 上限抽城堡 id(對齊 C
// `KB_rand(0, MAX_CASTLES-1)`),故傳入的 castles 切片長度仍須是 kbdata.MaxCastles
// (只是把其餘 25 座的 Continent 設成不可能命中的 -1),否則會索引越界——這不是
// saltVillains 本身的行為,只是本測試建構假資料時要遵守的前提。
func TestSaltVillains_NotEnoughCastles(t *testing.T) {
	castles := make([]Castle, kbdata.MaxCastles)
	for i := range castles {
		castles[i] = Castle{ID: i, Continent: -1}
	}
	castles[0].Continent = 0 // 洲 0 只有 1 座城堡,但 villains_per_continent[0] = 6,遠遠不足。

	gs := &GameState{}
	for i := range gs.CastleOwner {
		gs.CastleOwner[i] = castleOwnerMonsters
	}
	rng := kbrng.NewGlibc(1)

	got := gs.saltVillains(rng, castles, 0, 0)
	if got != 0 {
		t.Errorf("saltVillains 城堡不足時 got base_id=%d, want 0(不放任何惡棍)", got)
	}
	if gs.CastleOwner[0] != castleOwnerMonsters {
		t.Errorf("城堡不足時城堡0不應被佔領,got owner=%d", gs.CastleOwner[0])
	}
}
