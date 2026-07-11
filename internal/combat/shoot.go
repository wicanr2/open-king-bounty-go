package combat

// 本檔移植 C 版遠程射擊的「可否射擊」判定(src/game.c:5970 unit_try_shoot 的守門條件
// 與 src/play.c:1386 unit_surrounded)。傷害本身走既有的 UnitRangedShot(damage.go),
// 這裡只補「一個單位這回合能不能射」與「被誰貼身」兩個純判定,供玩家側射擊 UI 與
// 測試共用,不重寫傷害邏輯。
//
// ⚠ 重要:King's Bounty 的遠程「沒有距離射程上限」——C 版 unit_try_shoot 用
// pick_target(filter=4) 選目標,filter 只檢查「該格是敵方/失控單位」,完全不看距離;
// troops[].ranged_min/max 是「傷害」骰的上下界(deal_damage 用 KB_rand(min,max)),
// 不是射程格數,ranged_ammo 才是彈藥數(對應 kbdata.RangedShots / Unit.Shots)。
// 因此射擊唯一的門檻是:① 還有彈藥(Shots>0)② 沒有被敵人貼身(!UnitSurrounded)。
// 被貼身時 C 只允許近戰(見 game.c:6322 grid_heuristic:近敵→MELEE、遠敵才 SHOOT)。

// unitsAreFriendly 對應 C 版 units_are_friendly(src/play.c:1408-1415):
// 任一方失控(out_of_control)就互不為友;否則同側為友。
func (c *Combat) unitsAreFriendly(aSide, aID, oSide, oID int) bool {
	a := &c.Units[aSide][aID]
	o := &c.Units[oSide][oID]
	if a.OutOfControl || o.OutOfControl {
		return false
	}
	return aSide == oSide
}

// UnitSurrounded 對應 C 版 unit_surrounded(src/play.c:1386-1397):場上是否有任一
// 敵對(非友方)且尚有兵力的單位與 (side,id) 相鄰(unit_touching = Chebyshev<=1,
// 見 Adjacent)。被貼身的遠攻單位不能射,只能近戰。
func (c *Combat) UnitSurrounded(side, id int) bool {
	u := &c.Units[side][id]
	for j := 0; j < MaxSides; j++ {
		for i := 0; i < MaxUnits; i++ {
			o := &c.Units[j][i]
			if o.Count == 0 {
				continue
			}
			if j == side && i == id {
				continue
			}
			if c.unitsAreFriendly(side, id, j, i) {
				continue
			}
			if Adjacent(u.X, u.Y, o.X, o.Y) {
				return true
			}
		}
	}
	return false
}

// CanShoot 對應 C 版 unit_try_shoot 的可射守門(src/game.c:5970):
// 還有彈藥(Shots>0)且沒有被敵人貼身(!UnitSurrounded)。與距離無關(見檔頭說明)。
func (c *Combat) CanShoot(side, id int) bool {
	u := &c.Units[side][id]
	if u.Count == 0 {
		return false
	}
	return u.Shots > 0 && !c.UnitSurrounded(side, id)
}
