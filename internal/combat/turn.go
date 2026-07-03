package combat

import (
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// castleUmap 對應 C 版 castle_umap[5][6](src/play.c:1210-1216)——城堡佈局裡
// 每格站的是哪個 unit(值 = PACK_UID,0 = 空);索引 [y][x],尺寸對齊
// [BoardH][BoardW](= [CLEVEL_H][CLEVEL_W])。
var castleUmap = [BoardH][BoardW]int{
	{0, 8, 6, 7, 9, 0},
	{0, 0, 10, 0, 0, 0},
	{0, 0, 0, 0, 0, 0},
	{0, 5, 3, 4, 0, 0},
	{0, 0, 1, 2, 0, 0},
}

// castleOmap 對應 C 版 castle_omap[5][6](src/play.c:1218-1224)——城堡佈局的
// 固定障礙圖(取代一般戰鬥隨機產生障礙那段)。
var castleOmap = [BoardH][BoardW]int{
	{8, 0, 0, 0, 0, 9},
	{8, 0, 0, 0, 0, 9},
	{8, 0, 0, 0, 0, 9},
	{8, 0, 0, 0, 0, 9},
	{5, 7, 0, 0, 10, 6},
}

// ResetTurn 對應 C 版 reset_turn(src/play.c:1178-1200)——每次換手(玩家/AI 互換,
// 或城堡佈陣完成)都會呼叫,重設雙方所有單位的 turn_count/moves/flights/acted/retaliated,
// 並在「即將從側 1 換回側 0」的那一刻(C: if (war->side == 1))套用解凍與再生效果。
//
// 欄位來源(對齊 C):
//   - Moves ← troops[troop_id].move_rate,本套件對應 kbdata.Troop.Move。
//   - Flights ← 有 ABIL_FLY 才給 2,否則 0(C: flights=0; if (abilities&ABIL_FLY) flights=2;)。
//   - Shots 不在這裡設(C 版 reset_turn 本身也不碰 shots,彈藥只在 ResetMatch/reset_match 給一次,
//     整場戰鬥消耗到底,不會每回合恢復)。
//
// 偏離 C 版之處(刻意,避免 Go 版 panic):C 版對 troops[war->units[j][i].troop_id] 的存取
// 不檢查 troop_id 是否合法(空格 troop_id 常是哨兵值,C 陣列越界讀取在此情境是「無害」的,
// 見 setup.go 的 validTroop 說明);Go 版用 validTroop 守門,非法索引一律當作 Moves=0、無 Flights。
func (c *Combat) ResetTurn(a *kbdata.Assets) {
	for j := 0; j < MaxSides; j++ {
		for i := 0; i < MaxUnits; i++ {
			u := &c.Units[j][i]

			u.TurnCount = u.Count
			u.Retaliated = false
			u.Acted = false

			u.Moves = 0
			u.Flights = 0
			if validTroop(a, u.TroopID) {
				t := a.Troops[u.TroopID]
				u.Moves = t.Move
				if t.Abilities&kbdata.AbilFly != 0 {
					u.Flights = 2
				}
			}

			// End-of-turn effects:對應 C 版 `if (war->side == 1)`——注意這裡讀的是
			// *尚未翻轉* 的 c.Side(ResetTurn 在 NextTurn 裡先於側別互換被呼叫,
			// 與 C 版呼叫順序一致,見 NextTurn 注解)。
			if c.Side == 1 {
				u.Frozen = false
				if validTroop(a, u.TroopID) && a.Troops[u.TroopID].Abilities&kbdata.AbilRegen != 0 {
					u.Injury = 0
				}
			}
		}
	}
	c.Phase = 0
	c.Spells = 0
}

// ResetMatch 對應 C 版 reset_match(src/play.c:1226-1289)——整場戰鬥開打前呼叫一次:
// 視 castle 決定是否套用城堡固定佈局(castle_umap/castle_omap),幫所有有兵的格子
// 補上 Shots/Injury/Frozen 並登記 Umap,清空 Omap 後產生(或套用)障礙圖,最後呼叫 ResetTurn。
//
// rng 注入(對齊 kbrng.Rand,語意同 C 版 KB_rand/rand()):C 版障礙產生每格先擲一次
// `rand()%10`(1/10 機率有障礙),中選才再擲一次 `rand()%3+1` 決定障礙種類——兩次呼叫
// 對應 rng.Between(0,9)==0 與 rng.Between(1,3),呼叫次數與順序需與 C 一致才能 parity。
//
// 略去/簡化之處(比照 setup.go PrepareUnitsCastle 的既有保留範圍):
//   - 障礙生成演算法本身 C 原始碼自己註明是有問題的("A) 不是原版 King's Bounty 的做法
//     B) 有淹沒整層棋盤的風險 //TODO"),此處忠實照搬該演算法,不修正它。
func (c *Combat) ResetMatch(a *kbdata.Assets, rng kbrng.Rand, castle bool) {
	if castle {
		for y := 0; y < BoardH; y++ {
			for x := 0; x < BoardW; x++ {
				id := castleUmap[y][x] - 1
				if id < 0 {
					continue
				}
				side := 0
				if id >= MaxUnits {
					id -= MaxUnits
					side = 1
				}
				u := &c.Units[side][id]
				u.X = x
				u.Y = y
			}
		}
	}

	// Place units
	for j := 0; j < MaxSides; j++ {
		for i := 0; i < MaxUnits; i++ {
			u := &c.Units[j][i]
			if u.Count == 0 {
				continue
			}
			if validTroop(a, u.TroopID) {
				u.Shots = a.Troops[u.TroopID].RangedShots
			}
			u.Injury = 0
			u.Frozen = false
			c.Umap[u.Y][u.X] = PackUID(j, i)
		}
	}

	// wipe_battlefield
	for y := 0; y < BoardH; y++ {
		for x := 0; x < BoardW; x++ {
			c.Omap[y][x] = 0
		}
	}

	if castle {
		for y := 0; y < BoardH; y++ {
			for x := 0; x < BoardW; x++ {
				c.Omap[y][x] = byte(castleOmap[y][x])
			}
		}
	} else {
		for y := 0; y < BoardH; y++ {
			for x := 1; x < BoardW-2; x++ {
				c.Omap[y][x] = 0
				if rng.Between(0, 9) == 0 {
					c.Omap[y][x] = byte(rng.Between(1, 3))
				}
			}
		}
	}

	c.ResetTurn(a)
}

// NextTurn 對應 C 版 next_turn(src/play.c:1292-1306)——換手:回合數 +1、呼叫 ResetTurn
// (此時 c.Side 仍是「即將結束」的那一側),接著在*舊*側裡找第一個還有兵力(Count>0)的
// 格子,記下它的索引為 UnitID,然後才把 Side 換成另一側,回傳是否找到。
//
// ⚠ 忠實保留 C 版一個容易誤會的細節:找到的 UnitID 是「舊側」某個有兵格子的*索引*,
// 但翻面之後,這個索引被直接套用到*新側*(C 版沒有針對新側重新驗證這個索引)。
// 這在 C 裡是安全的,因為呼叫端(combat_loop)之後每次都會檢查
// `!combat->units[combat->side][combat->unit_id].count` 為真就 pass=1、
// 觸發 NextUnit 重新在新側找一個合法單位——這個自我修正發生在驅動迴圈,
// 不在 NextTurn 本身。呼叫端(測試/未來旗艦的迴圈)必須做一樣的事:換側後
// 先檢查 UnitID 指到的格子是否有兵,沒有就呼叫 NextUnit 往下找。
func (c *Combat) NextTurn(a *kbdata.Assets) bool {
	c.Turn++
	c.ResetTurn(a)

	for i := 0; i < MaxUnits; i++ {
		if c.Units[c.Side][i].Count > 0 {
			c.UnitID = i
			c.Side = 1 - c.Side
			return true
		}
	}
	return false
}

// NextUnit 對應 C 版 next_unit(src/play.c:1309-1329)——在目前側(c.Side)裡,
// 從 UnitID+1 開始找下一個「有兵且尚未行動」的格子;找不到就進下一個 phase
// (c.Phase++)後從頭(0)再找一次;兩輪都找不到回傳 -1。
//
// 找到時會順手更新該單位的 OutOfControl(C: u->out_of_control = !unit_under_control(...)),
// 對齊 C 版同一行為(呼叫端拿到索引前,OutOfControl 已經是最新值)。
func (c *Combat) NextUnit(a *kbdata.Assets) int {
	for i := c.UnitID + 1; i < MaxUnits; i++ {
		u := &c.Units[c.Side][i]
		if u.Count == 0 || u.Acted {
			continue
		}
		u.OutOfControl = !c.unitUnderControl(a, c.Side, i)
		return i
	}

	c.Phase++
	for i := 0; i < MaxUnits; i++ {
		u := &c.Units[c.Side][i]
		if u.Count == 0 || u.Acted {
			continue
		}
		u.OutOfControl = !c.unitUnderControl(a, c.Side, i)
		return i
	}

	return -1
}

// unitUnderControl 對應 C 版 unit_under_control(src/play.c:1363-1368),
// 包成 (side,id) 索引形式的小 wrapper,內部直接沿用 setup.go 既有的
// unitUnderControl(gs, a, troopID, count) —— 判斷規則完全相同,不重複實作。
func (c *Combat) unitUnderControl(a *kbdata.Assets, side, id int) bool {
	u := &c.Units[side][id]
	return unitUnderControl(c.Heroes[side], a, u.TroopID, u.Count)
}
