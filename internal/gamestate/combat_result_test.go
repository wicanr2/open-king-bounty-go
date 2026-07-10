package gamestate

import "testing"

// TestFulfillContract 驗證履約:領賞、標記已捕獲、清當前契約、循環補新惡棍、max_contract++。
func TestFulfillContract(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)

	// 佈置一個可控的契約循環:villain 4 在 slot 0,當前契約 = 4,max_contract=5。
	gs.Contract = 4
	gs.ContractCycle = [5]byte{4, 1, 2, 3, 0xFF}
	gs.MaxContract = 5
	goldBefore := gs.Gold
	villains := LoadVillains(a)
	reward := villains[4].Reward

	gs.FulfillContract(a, 4)

	if gs.VillainCaught[4] != 1 {
		t.Errorf("villain 4 未標記已捕獲")
	}
	if gs.Contract != 0xFF {
		t.Errorf("履約後當前契約應清成 0xFF,got %#x", gs.Contract)
	}
	if gs.Gold != goldBefore+reward {
		t.Errorf("賞金未入帳:gold %d→%d,want +%d", goldBefore, gs.Gold, reward)
	}
	if gs.ContractCycle[0] == 4 {
		t.Errorf("契約循環 slot 0 仍是已履約的 villain 4:%v", gs.ContractCycle)
	}
	if gs.MaxContract != 6 {
		t.Errorf("max_contract 應 ++ 到 6,got %d", gs.MaxContract)
	}
	// slot 0 應補入從 max_contract(=5,遞增前)起第一個未捕獲的惡棍 = villain 5。
	if gs.ContractCycle[0] != 5 {
		t.Errorf("契約循環 slot 0 應補入 villain 5,got %d", gs.ContractCycle[0])
	}
}

// TestTempDeath 驗證戰敗:隊伍清空後只剩第 0 格 20 名農夫。
func TestTempDeath(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	gs.Army = [5]Squad{{TroopID: 5, Count: 30}, {TroopID: 7, Count: 12}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}}

	gs.TempDeath(a)

	if gs.Army[0].TroopID != peasantTroopID || gs.Army[0].Count != 20 {
		t.Errorf("戰敗後第 0 格應為 20 農夫,got %+v", gs.Army[0])
	}
	for i := 1; i < 5; i++ {
		if gs.Army[i].TroopID != 0xFF || gs.Army[i].Count != 0 {
			t.Errorf("戰敗後第 %d 格應清空,got %+v", i, gs.Army[i])
		}
	}
}
