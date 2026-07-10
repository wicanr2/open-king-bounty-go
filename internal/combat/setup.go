package combat

import (
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// emptyTroopID 是 C 版隊伍格「空格」的哨兵值(0xFF),對應 gamestate.Squad.TroopID 的註解
// 與 gamestate 內部同義的 emptySquad 常量。
const emptyTroopID = 0xFF

// PrepareUnitsPlayer 把玩家隊伍(gs.Army 五格)佈上棋盤,對應 C 版 prepare_units_player
// (src/play.c:1087-1107)。
//
// 佈陣規則(逐條對齊 C):
//   - Army 五格依序對到 Units[side][0..4]:Y = 格索引 i,X = side*(BoardW-1)
//     (C: u->x = side * (CLEVEL_W - 1))——side 0 站棋盤最左欄(X=0),side 1 站最右欄(X=BoardW-1)。
//   - Moves/Shots/Flights 這裡不設定。C 版 prepare_units_player 本身也不設這三個欄位,
//     要等 reset_match/reset_turn(src/play.c:1178-1289)才賦值,那段屬於「回合流程」,
//     不在本套件範圍內,故此處忠實留白(旗艦之後串連回合流程時再填)。
//   - OutOfControl:對應 unit_under_control(src/play.c:1363-1368),用
//     gs.ArmyLeadership(a, troopID) 是否 > 0 判斷(gamestate 已有 ArmyLeadership,直接沿用)。
//   - Spoils[side] 累加 (troop.Spoils * 5) * count,對應 C 的
//     troops[troop_id].spoils_factor * 5 * count(kbdata.Troop.Spoils 對應 spoils_factor)。
//   - Powers[side] 恆為 0:gamestate 尚未移植 artifact_found 世界狀態,TODO(flagship)。
//
// 偏離 C 版之處(刻意,避免 Go 版 panic):C 版空格 troop_id 常是 0xFF,
// troops[0xFF] 的存取結果會被 count(=0)乘掉、屬於「無害」的 C 陣列越界讀取;
// Go slice 越界會直接 panic。因此本函式在 troopID 不是 a.Troops 合法索引時,
// 一律跳過查表(不算 Spoils、OutOfControl 視為受控),只搬 TroopID/Count 等欄位。
func PrepareUnitsPlayer(c *Combat, side int, gs *gamestate.GameState, a *kbdata.Assets) {
	c.Heroes[side] = gs
	for i := 0; i < MaxUnits; i++ {
		squad := gs.Army[i]
		u := &c.Units[side][i]

		u.TroopID = squad.TroopID
		u.Count = squad.Count
		u.MaxCount = u.Count
		u.Y = i
		u.X = side * (BoardW - 1)

		u.OutOfControl = !unitUnderControl(gs, a, squad.TroopID, squad.Count)

		if validTroop(a, squad.TroopID) && squad.Count > 0 {
			c.Spoils[side] += a.Troops[squad.TroopID].Spoils * 5 * squad.Count
		}
	}
	// 神器加成:OR 所有已拾得神器的 power 位元(對齊 C prepare_units_player play.c:1141)。
	c.Powers[side] = byte(gs.ArtifactPowers(a))
}

// PrepareUnitsFoe 把敵方隊伍佈上棋盤,對應 C 版 prepare_units_foe(src/play.c:1109-1128)。
//
// gamestate 目前尚未移植 KBgame 的 foe_troops/foe_numbers 世界狀態(見
// internal/gamestate/gamestate.go 開頭註解:地圖/惡棍/城堡等世界狀態尚未納入),
// 因此本函式直接吃呼叫端準備好的 [MaxUnits]gamestate.Squad,而非整個
// KBgame + continent_id/foe_id 索引——等旗艦補上 foe 世界狀態後,呼叫端自行從那邊
// 組出 squads 傳進來即可,佈陣規則不變。
//
// 佈陣規則:
//   - 只佈前 3 格(C: max_troops 被夾到 3,src/play.c:1111-1112),第 4、5 格保持零值。
//   - Y = 格索引 i,X = side*(BoardW-1),與玩家佈陣同一條規則。
//   - OutOfControl 恆 false(C: out_of_control = 0,敵方沒有 hero/leadership 概念)。
//   - Heroes[side] = nil、Powers[side] = 0(C: heroes[side]=NULL、powers[side]=0)。
func PrepareUnitsFoe(c *Combat, side int, squads [MaxUnits]gamestate.Squad, a *kbdata.Assets) {
	const maxTroops = 3
	c.Spoils[side] = 0
	for i := 0; i < maxTroops; i++ {
		squad := squads[i]
		u := &c.Units[side][i]

		u.TroopID = squad.TroopID
		u.Count = squad.Count
		u.MaxCount = u.Count
		u.OutOfControl = false
		u.Y = i
		u.X = side * (BoardW - 1)

		if validTroop(a, squad.TroopID) && squad.Count > 0 {
			c.Spoils[side] += a.Troops[squad.TroopID].Spoils * 5 * squad.Count
		}
	}
	c.Heroes[side] = nil
	c.Powers[side] = 0
}

// PrepareUnitsCastle 把城堡守軍佈上棋盤,對應 C 版 prepare_units_castle
// (src/play.c:1130-1146)。
//
// 注意(範圍界定):C 版 prepare_units_castle 本身把兩側 X 都設 0、Y=i;
// 實際的城堡棋盤佈局(castle_umap 表,src/play.c:1210-1216,雙方單位依表格重新定位)
// 是後續 reset_match(castle=1)(src/play.c:1226-1289)才覆寫上去的——那段和障礙圖生成、
// reset_turn 綁在一起,屬於對局初始化/回合流程,不在本套件範圍。因此本函式只忠實搬
// prepare_units_castle 那一段,不套用 castle_umap 覆寫;真正城堡佈局留給旗艦串連。
func PrepareUnitsCastle(c *Combat, side int, squads [MaxUnits]gamestate.Squad, a *kbdata.Assets) {
	c.Spoils[side] = 0
	for i := 0; i < MaxUnits; i++ {
		squad := squads[i]
		u := &c.Units[side][i]

		u.TroopID = squad.TroopID
		u.Count = squad.Count
		u.MaxCount = u.Count
		u.OutOfControl = false
		u.Y = i
		u.X = 0

		if validTroop(a, squad.TroopID) && squad.Count > 0 {
			c.Spoils[side] += a.Troops[squad.TroopID].Spoils * 5 * squad.Count
		}
	}
	c.Heroes[side] = nil
	c.Powers[side] = 0
}

// validTroop 回傳 troopID 是否為 a.Troops 的合法索引(避免空格哨兵值 0xFF 之類造成越界)。
func validTroop(a *kbdata.Assets, troopID int) bool {
	return a != nil && troopID >= 0 && troopID < len(a.Troops)
}

// unitUnderControl 對應 C 版 unit_under_control(src/play.c:1363-1368):
//   - 沒有 hero(敵方/城堡側)一律視為受控(C: !war->heroes[side] return 1)。
//   - count==0 的空格,C 版回傳 -1、取反後仍是「受控」(out_of_control=0),此處比照辦理。
//   - troopID 不是合法索引時視為受控(避免查表越界;C 版此情境本就只在 count==0 發生)。
//   - 否則用 gamestate 既有的 ArmyLeadership(該兵種還能再塞多少)是否 > 0 判斷,
//     對應 C 的 army_leadership(war->heroes[side], u->troop_id) > 0。
func unitUnderControl(gs *gamestate.GameState, a *kbdata.Assets, troopID, count int) bool {
	if gs == nil {
		return true
	}
	if count == 0 {
		return true
	}
	if !validTroop(a, troopID) {
		return true
	}
	return gs.ArmyLeadership(a, troopID) > 0
}
