package combat

// 飛行:對應 C 版 fly_unit(game.c:3552)+ unit_fly_offset(play.c:1729)。
// 會飛的單位可越過中途障礙,直接落到目標「相鄰的空格」;每次耗 1 flight,歸 0 即 acted。

// relocate 把單位直接搬到 (nx,ny) 並更新 Umap(對齊 unit_relocate),不檢查中途路徑。
func (c *Combat) relocate(side, id, nx, ny int) {
	u := &c.Units[side][id]
	c.Umap[u.Y][u.X] = 0
	c.Umap[ny][nx] = PackUID(side, id)
	u.X = nx
	u.Y = ny
}

// FlyUnit 把飛行單位移到 (nx,ny)(無視中途障礙),耗 1 flight,歸 0 則標記 acted。
// 對齊 fly_unit(game.c:3552)。
func (c *Combat) FlyUnit(side, id, nx, ny int) bool {
	u := &c.Units[side][id]
	c.relocate(side, id, nx, ny)
	u.Flights--
	if u.Flights <= 0 {
		u.Acted = true
	}
	return true
}

// aiFlyOffset 找目標相鄰的空格當飛行落點,對應 unit_fly_offset(play.c:1729):
// 以目標座標為中心用 aiClosestOffset 搜一格;若該格被障礙/單位佔用則回原地(不飛)。
func (c *Combat) aiFlyOffset(side, id, targetSide, targetID int) (int, int) {
	u := &c.Units[side][id]
	other := &c.Units[targetSide][targetID]
	ox, oy := c.aiClosestOffset(other.X, other.Y, u.X, u.Y, other.X, other.Y)
	nx, ny := other.X+ox, other.Y+oy
	if nx < 0 || ny < 0 || nx > BoardW || ny > BoardH {
		return u.X, u.Y
	}
	if c.Omap[ny][nx] != 0 || c.Umap[ny][nx] != 0 {
		return u.X, u.Y // 占用 → 不飛
	}
	return nx, ny
}
