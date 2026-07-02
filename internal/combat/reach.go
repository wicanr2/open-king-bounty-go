package combat

// CanReach 回傳 side/unitID 這個單位,能否在剩餘 Moves 步以內、用「每步向八個方向
// 相鄰一格(King-move)」的走法,移動到 (tx,ty)。純幾何查詢,不含攻擊/戰鬥結算。
//
// 幾何規則對應 C 版 move_unit 的單步合法性(src/game.c:3511-3550),推廣成多步搜尋:
//   - 每步偏移 (ox,oy) 各在 {-1,0,1}(King-move,8 方向,含直走與斜走,同一步花費相同),
//     一步耗 1 點 Moves(C: u->moves--)。
//   - 落點需 InBounds(C: 邊界檢查 nx/ny 落在 [0,CLEVEL_W)x[0,CLEVEL_H))。
//   - 落點不可有障礙(C: war->omap[ny][nx] 非 0 直接擋、回傳「未移動」)。
//   - 落點不可被佔用(C: war->umap[ny][nx] 非 0 時,C 版改成觸發攻擊而非移動——
//     本函式只做純幾何可達性,佔用格一律視為擋路,既不能通過也不能落在該格;
//     TODO(flagship):是否能移動攻入敵方格、攻擊互動本身不在此判定範圍)。
//   - 起點視為 0 步可達(與目標同格必真)。
//
// 注意(範圍界定):C 原始碼裡沒有一個現成的「多步可達性」函式可以逐行對照——
// 玩家移動是逐次按鍵、每次呼叫一次 move_unit 走一步,Moves 只是「還能按幾次」的計數器。
// 這裡把該單步合法性規則忠實照搬,套用在「BFS 走滿 Moves 步」這個新介面上:
// 每一條擋路規則都對應 C 的既有判斷,但「组合成多步」是本套件為了滿足
// 「移動可達性查詢」需求做的幾何推廣,而非直接移植某一個 C 函式。
func (c *Combat) CanReach(side, unitID, tx, ty int) bool {
	u := c.Units[side][unitID]

	if u.X == tx && u.Y == ty {
		return true
	}
	if !InBounds(tx, ty) {
		return false
	}
	if u.Moves <= 0 {
		return false
	}

	type pos struct{ x, y int }

	start := pos{u.X, u.Y}
	target := pos{tx, ty}

	visited := map[pos]bool{start: true}
	frontier := []pos{start}

	for step := 0; step < u.Moves; step++ {
		var next []pos
		for _, p := range frontier {
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					np := pos{p.x + dx, p.y + dy}
					if !InBounds(np.x, np.y) {
						continue
					}
					if c.Obstacle(np.x, np.y) {
						continue
					}
					if c.Occupied(np.x, np.y) {
						continue
					}
					if visited[np] {
						continue
					}
					if np == target {
						return true
					}
					visited[np] = true
					next = append(next, np)
				}
			}
		}
		if len(next) == 0 {
			break
		}
		frontier = next
	}

	return false
}
