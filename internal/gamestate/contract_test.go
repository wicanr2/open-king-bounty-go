package gamestate

import "testing"

func TestNextContract(t *testing.T) {
	// 起手 ContractCycle={0,1,2,3,4}, LastContract=4(對齊 spawn_game)。
	gs := &GameState{LastContract: 4, ContractCycle: [5]byte{0, 1, 2, 3, 4}}
	// 第一次領:LastContract=4 在 slot4,第二輪回第一個非該 slot → villain 0。
	if v := gs.NextContract(); v != 0 {
		t.Errorf("first contract: got %d, want 0", v)
	}
	// 領了 villain 0 後 last=0 → 下一個 = 1
	gs.LastContract = 0
	if v := gs.NextContract(); v != 1 {
		t.Errorf("after last=0: got %d, want 1", v)
	}
	// 循環空(全 0xFF)→ 0xFF
	gs.ContractCycle = [5]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	if v := gs.NextContract(); v != 0xFF {
		t.Errorf("empty cycle: got %d, want 0xFF", v)
	}
}
