package combat

// MoveUnit 把 side/id 這個單位移動到 (x,y),對應 C 版 move_unit(src/game.c:3511-3550)
// 的「純移動」分支——不含「撞到單位改觸發攻擊」那條:那條由呼叫端(ai.go 的
// AIUnitThink)先用 Adjacent 判斷、直接呼叫 UnitHitUnit 處理,理由與等價性推導
// 見 AIUnitThink/aiClosestOffset 的文件注解。
//
// 前置檢查用現成 CanReach(side,id,x,y)(邊界/障礙/佔用/剩餘 Moves>0 皆由它把關)。
//
// 範圍界定(刻意,非 C 直接可對照的行為):本函式只接受「單步」目標,要求
// (x,y) 與目前座標 Adjacent——即使 CanReach 用 BFS 判斷更遠的格子「可達」,
// MoveUnit 也不接受非相鄰目標。理由:C 版 move_unit 本身只吃 offset(恆在
// {-1,0,1}),從未支援「一次跳好幾格」;AI(ai.go)也只會用相鄰的單步 offset
// 呼叫它。要多步移動,C 版靠的是同一個單位被反覆呼叫 move_unit(逐 tick),
// 不是 move_unit 一次做完——本函式的呼叫慣例與此一致。
//
// 成功時更新 Umap(清舊格、佔新格)與 u.X/u.Y,並扣 1 點 Moves,Moves 歸零則
// 標記 Acted(對應 C: u->moves--; if (!u->moves) u->acted = 1;)。
//
// 偏離 C 版之處(刻意,不重現 byte 下溢位):C 版 move_unit 完全不檢查
// u->moves 是否 >0 就直接做 u->moves--(byte 型別,若在 moves 已經是 0
// 時被呼叫會下溢成 255)——但這個情況在 C 的實際呼叫路徑下不會發生,因為
// moves 归零的同一次呼叫就會把 acted 設為真,驅動迴圈(combat_loop)之後
// 就不會再對這個單位呼叫 move_unit / ai_unit_think。本函式仰賴 CanReach
// 的 `u.Moves <= 0` 檢查代替這個「本來就不會走到」的防線,不刻意重現下溢位。
func (c *Combat) MoveUnit(side, id, x, y int) bool {
	u := &c.Units[side][id]

	if u.X == x && u.Y == y {
		return false
	}
	if !Adjacent(u.X, u.Y, x, y) {
		return false
	}
	if !c.CanReach(side, id, x, y) {
		return false
	}

	c.Umap[u.Y][u.X] = 0
	c.Umap[y][x] = PackUID(side, id)
	u.X = x
	u.Y = y

	u.Moves--
	if u.Moves <= 0 {
		u.Acted = true
	}
	return true
}

// Winner 回傳某側全滅時「還活著」那側的索引,雙方都還有兵或雙方都全滅
// (後者理論上不該發生,防呆保底)回傳 -1。
//
// 對應 C 版 test_victory(src/play.c:1078-1085:側 1 全滅則側 0 勝)並對稱補上
// 「側 0 全滅則側 1 勝」——C 版本身沒有這個對稱函式:玩家陣營的「輸」是靠
// test_defeat(src/play.c:1071-1076)去查 game->player_numbers(戰鬥外層、由
// compact_units 在每次死亡後同步回去的世界狀態),而不是直接查 war->units。
// 本套件的 Combat 不保證一定綁著 gamestate(見 PrepareUnitsFoe/Castle 的
// Heroes[side]==nil 情境),因此用「該側所有 Units[].Count 皆為 0」直接判定;
// 語意等價於 test_defeat,因為 compact_units 每次死亡都會把 count 同步回
// game,兩者恆一致。
func (c *Combat) Winner() int {
	var dead [MaxSides]bool
	for side := 0; side < MaxSides; side++ {
		allDead := true
		for i := 0; i < MaxUnits; i++ {
			if c.Units[side][i].Count > 0 {
				allDead = false
				break
			}
		}
		dead[side] = allDead
	}

	switch {
	case dead[0] && dead[1]:
		return -1
	case dead[0]:
		return 1
	case dead[1]:
		return 0
	default:
		return -1
	}
}
