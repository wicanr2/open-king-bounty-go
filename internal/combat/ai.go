package combat

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// AIPickTarget 對應 C 版 ai_pick_target(src/game.c:5868-5915)——幫目前行動單位
// (c.Side/c.UnitID)在棋盤上挑一個目標。
//
//	nearby=true  對應 C 的 nearby=1:只考慮跟自己 Adjacent(C: unit_touching)的目標
//	             (用來找「近戰/被貼身」的對象)。
//	nearby=false 對應 C 的 nearby=0(「far」模式):不限距離,但排序邏輯偏好
//	             還有 Shots 的目標(射手優先),語意見下方逐行對照。
//
// 遍歷規則(逐條對齊 C,順序不可變——決定「同分時最後找到的蓋過前面」的行為):
//   - 依序看兩側(side 0 再 side 1)、每側依序看 5 格。
//   - 跳過自己;若自己是「受控」狀態(非 OutOfControl),跳過整個己方側
//     (out-of-control 的單位才會把己方也當目標,對應失控暴走)。
//   - nearby=true 時,只留下與自己 Adjacent 的目標。
//   - 排序:far 模式下,若已選到的目標還有 Shots、而候選沒有 Shots,跳過候選
//     (射手優先);其餘情況(near 模式,或 far 模式下雙方都沒有 Shots)比較
//     兵種 HP,留 HP 較小或相等者(相等時「後找到蓋過先找到」,因為判斷式是
//     嚴格 `>` 才跳過,C 原文如此,照搬不改)。
//
// 回傳 (side, id, ok);ok=false 對應 C 的 pick==-1(全場沒有合法目標)。
func (c *Combat) AIPickTarget(a *kbdata.Assets, nearby bool) (side, id int, ok bool) {
	actor := &c.Units[c.Side][c.UnitID]
	underControl := !actor.OutOfControl
	far := !nearby

	pick := -1

	for j := 0; j < MaxSides; j++ {
		for i := 0; i < MaxUnits; i++ {
			u := &c.Units[j][i]
			if u.Count == 0 {
				continue
			}
			if j == c.Side && i == c.UnitID {
				continue
			}
			if j == c.Side && underControl {
				continue
			}
			if nearby && !Adjacent(u.X, u.Y, actor.X, actor.Y) {
				continue
			}

			if pick != -1 {
				pSide, pID := UnpackUID(byte(pick))
				picked := &c.Units[pSide][pID]

				if far {
					if picked.Shots > 0 && u.Shots == 0 {
						continue
					}
				}
				if !far || u.Shots == 0 {
					if troopHP(a, u.TroopID) > troopHP(a, picked.TroopID) {
						continue
					}
				}
			}

			pick = int(PackUID(j, i))
		}
	}

	if pick == -1 {
		return 0, 0, false
	}
	side, id = UnpackUID(byte(pick))
	return side, id, true
}

// troopHP 回傳 troopID 的 HP,troopID 非法索引時回 0(避免越界 panic;
// 見 setup.go validTroop 的同一套防呆理由)。
func troopHP(a *kbdata.Assets, troopID int) int {
	if !validTroop(a, troopID) {
		return 0
	}
	return a.Troops[troopID].HP
}

// AIUnitThink 對應 C 版 ai_unit_think(src/game.c:5917-6009)——目前行動單位每次
// 「心跳」的決策:凍結就跳過、有近戰目標就優先近戰、沒有近戰目標且還有彈藥就
// 遠攻、否則試著往目標移動一步、都不行就待命。
//
// ⚠ 顆粒度對齊 C 的逐 tick 模型,不是「一次呼叫走完整個回合」:
// C 版 ai_unit_think 每次只做一個動作(射一箭、走一步、或待命),combat_loop
// 每個畫面更新 tick 呼叫一次;同一個單位只要還有剩餘 Moves 且尚未 Acted,
// 就會被同一顆心跳邏輯重複呼叫、逐步走到底或找到攻擊機會為止(每步都重新
// 呼叫 AIPickTarget,因為單位移動後最近目標可能改變——對齊 C 逐 tick 重算)。
// 本函式忠實保留這個「一次一動作」的顆粒度;呼叫端(測試的驅動迴圈,或未來
// 旗艦接的畫面迴圈)要負責在同一個單位尚未 Acted 前反覆呼叫它,直到 Acted
// 被設起來,或本函式回傳的 pass=true,才呼叫 NextUnit/NextTurn 換下一個單位。
//
// 回傳值 pass 對齊 C 版 return 值:C 端函式內部變數雖叫 acted,但最終
// `return !acted`,呼叫端接住的變數名叫 pass——pass=true 代表「這個單位這個
// tick 完全無法行動,fallback 到待命」(C: return unit_try_wait(...) 恆為真值),
// 促使外層即使 Acted 沒被設起來也強制換下一個單位(對應 combat_loop 的
// `if (pass || units[...].acted) { ... }`)。pass=false 代表做了一個真正的
// 動作(射擊/移動/近戰),外層要不要換單位只看 Acted 有沒有被設起來。
//
// 已略過/簡化之處(務必核實,見套件層級的收尾報告):
//  1. 飛行(unit_fly_offset/fly_unit)已於 flight.go 移植並在此接上飛行分支
//     (凍結→遠攻→飛行→近戰/移動,對齊 C ai_unit_think 順序)。
//  2. 「近戰攻擊」不是照抄 C 版「移動演算法算出的落點恰好是敵方佔用格,
//     move_unit 偵測佔用轉呼叫 hit_unit」這個統一機制,而是先用 Adjacent
//     判斷、直接呼叫 UnitHitUnit——兩者結果等價(推導見下方 aiClosestOffset
//     注解:目標自己的格子距離恆為 0,是全域最小值,必定被選中,兩種寫法
//     殊途同歸),但實作路徑不同,回報時特別列出讓旗艦核實推導是否成立。
//  3. 士氣、Powers 神器加成、demonKills 等已經在 damage.go 處理,AIUnitThink
//     不重複套用。
func (c *Combat) AIUnitThink(a *kbdata.Assets, rng kbrng.Rand) (pass bool) {
	u := &c.Units[c.Side][c.UnitID]

	closeSide, closeID, hasClose := c.AIPickTarget(a, true)
	acted := false

	// 凍結:對應 C 的 `if (!acted && u->frozen) { u->acted = 1; acted = 1; }`。
	if !acted && u.Frozen {
		u.Acted = true
		acted = true
	}

	// 遠攻:對應 C 的 `if (!acted && u->shots) { if (close_target == -1) {...} }`——
	// 身邊沒有近戰目標時,才輪到用剩餘彈藥射遠處目標。
	if !acted && u.Shots > 0 {
		if !hasClose {
			farSide, farID, hasFar := c.AIPickTarget(a, false)
			if hasFar {
				c.UnitRangedShot(a, rng, c.Side, c.UnitID, farSide, farID)
				u.Acted = true
				acted = true
			}
		}
	}

	// 飛行:對應 C `if (!acted && u->flights && close_target == -1)`——有飛行點且
	// 身邊無近戰目標時,飛到遠處目標的相鄰空格(越過中途障礙)。
	if !acted && u.Flights > 0 && !hasClose {
		if farSide, farID, hasFar := c.AIPickTarget(a, false); hasFar {
			nx, ny := c.aiFlyOffset(c.Side, c.UnitID, farSide, farID)
			if nx != u.X || ny != u.Y {
				acted = c.FlyUnit(c.Side, c.UnitID, nx, ny)
			}
		}
	}

	// 近戰或移動:對應 C 的 `if (!acted && !u->shots) { ... }`——只有沒有剩餘
	// 彈藥(或本來就不能遠攻)的單位才會嘗試近戰/移動。
	if !acted && u.Shots == 0 {
		targetSide, targetID, hasTarget := closeSide, closeID, hasClose
		if !hasTarget {
			targetSide, targetID, hasTarget = c.AIPickTarget(a, false)
		}

		if hasTarget {
			t := &c.Units[targetSide][targetID]

			if Adjacent(u.X, u.Y, t.X, t.Y) {
				// 對應 move_unit 撞到敵方佔用格時改觸發攻擊的分支
				// (src/game.c:3525-3540);推導見「已略過」第 2 點。
				c.UnitHitUnit(a, rng, c.Side, c.UnitID, targetSide, targetID)
				u.Acted = true
				acted = true
			} else {
				ox, oy := c.aiClosestOffset(u.X, u.Y, u.X, u.Y, t.X, t.Y)
				if ox != 0 || oy != 0 {
					acted = c.MoveUnit(c.Side, c.UnitID, u.X+ox, u.Y+oy)
				}
			}
		}
	}

	if !acted {
		// 對應 unit_try_wait(src/game.c:5749-5758):只有 phase>0 才真的
		// 標記已行動,恆回傳「pass=true」。
		if c.Phase > 0 {
			u.Acted = true
		}
		return true
	}

	return false
}

// aiClosestOffset 對應 C 版 unit_closest_offset(src/play.c:1655-1707)——在
// (posX,posY) 周圍 3x3 鄰域(含原地)裡,找一格使其到 (targetX,targetY) 的
// 曼哈頓距離最小,且該格沒有障礙、也沒有被別的單位佔據(佔據的格子若剛好
// 就是 target 本身則不擋——這是刻意保留的 C 行為:讓「移動」演算法算出來的
// 落點可以恰好落在敵方目標格,由呼叫端 [AIUnitThink] 的 Adjacent 判斷接手,
// 兩者結果等價,見 AIUnitThink 文件的「已略過」第 2 點)。
//
// posX/posY、originX/originY 兩組座標對應 C 的 position_x/y、origin_x/y:
// 走路時兩組相同(= 單位目前座標),飛行時 position 會是「以目標為中心」
// 搜尋(本套件未移植飛行,呼叫端一律傳同一組座標)。當兩者相同時,C 版不會
// 跳過 (0,0)(原地),但原地格必定被「佔據(自己)」擋下,dist 仍會落到
// max_dist,效果上等同跳過——因此本函式不需要另外特判 originX==posX。
//
// 回傳偏移 (ox,oy),對齊 unit_move_offset 的呼叫慣例:全部候選都不可行時
// 回傳 (0,0)(C: `*ox = 0; *oy = 0;`),呼叫端用 ox!=0||oy!=0 判斷是否真的
// 找到路。
func (c *Combat) aiClosestOffset(posX, posY, originX, originY, targetX, targetY int) (ox, oy int) {
	maxDist := Distance(BoardW+1, BoardH+1, 0, 0)
	pickedDist := maxDist
	pickedX, pickedY := 0, 0

	for j := 1; j > -2; j-- {
		for i := -1; i < 2; i++ {
			nx, ny := posX+i, posY+j
			if nx < 0 || ny < 0 || nx >= BoardW || ny >= BoardH {
				continue
			}

			dist := Distance(nx, ny, targetX, targetY)

			if !(posX == originX && posY == originY) {
				if i == 0 && j == 0 {
					continue
				}
			}

			if c.Obstacle(nx, ny) {
				dist = maxDist
			}
			if c.Occupied(nx, ny) && !(nx == targetX && ny == targetY) {
				dist = maxDist
			}

			if dist < pickedDist {
				pickedDist = dist
				pickedX, pickedY = i, j
			}
		}
	}

	if pickedDist < maxDist {
		return pickedX, pickedY
	}
	return 0, 0
}
