package gamestate

// NextContract 回傳「下一個可領取的懸賞契約」villain id,對齊 C next_contract
// (play.c:573):沿 ContractCycle 找出 LastContract 之後的下一個 villain;
// 找不到(契約循環已空)回 0xFF。
//
// 演算法逐句對齊 C:先掃一輪找 LastContract 所在 slot,之後遇到的第一個非
// LastContract slot 即回傳(處理正向循環);第一輪沒回傳(LastContract 在循環
// 尾或不在循環內)則第二輪回傳第一個非該 slot 的 villain。
func (gs *GameState) NextContract() byte {
	slot := -1
	for i := 0; i < 5; i++ {
		if gs.ContractCycle[i] == 0xFF {
			continue
		}
		if gs.LastContract == gs.ContractCycle[i] {
			slot = i
		}
		if slot != -1 && slot != i {
			return gs.ContractCycle[i]
		}
	}
	for i := 0; i < 5; i++ {
		if gs.ContractCycle[i] == 0xFF {
			continue
		}
		if slot != i {
			return gs.ContractCycle[i]
		}
	}
	return 0xFF
}
