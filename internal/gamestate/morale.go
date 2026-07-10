package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// Morale* 對齊 C bounty.h(187-189 行):MORALE_NORMAL=0 MORALE_LOW=1 MORALE_HIGH=2。
// 索引即 MoraleNames 的索引,務必保持這個數值與名稱的對齊(見下)。
const (
	MoraleNormal = 0
	MoraleLow    = 1
	MoraleHigh   = 2
)

// MoraleNames 逐字對齊 C bounty.c morale_names[3](591 行):
//
//	morale_names[MORALE_NORMAL] = "普通"
//	morale_names[MORALE_LOW]    = "低落"
//	morale_names[MORALE_HIGH]   = "高昂"
var MoraleNames = [3]string{"普通", "低落", "高昂"}

// moraleChart 逐字對齊 C bounty.c morale_chart[5][5](150-157 行)。士氣群 A-E
// 對映到索引 0-4(與 kbdata.Troop.MoraleGroup 的既有索引一致,見 tables_troops.go)。
// 列 = 友軍士氣群(groupJ),欄 = 被查詢部隊士氣群(groupI):
//
//	     A  B  C  D  E
//	A  { N, N, N, N, N }
//	B  { N, N, N, N, N }
//	C  { N, N, H, N, N }
//	D  { L, N, L, H, N }
//	E  { L, L, L, N, N }
var moraleChart = [5][5]int{
	{MoraleNormal, MoraleNormal, MoraleNormal, MoraleNormal, MoraleNormal}, // A
	{MoraleNormal, MoraleNormal, MoraleNormal, MoraleNormal, MoraleNormal}, // B
	{MoraleNormal, MoraleNormal, MoraleHigh, MoraleNormal, MoraleNormal},   // C
	{MoraleLow, MoraleNormal, MoraleLow, MoraleHigh, MoraleNormal},         // D
	{MoraleLow, MoraleLow, MoraleLow, MoraleNormal, MoraleNormal},          // E
}

// moraleRank 對齊 C troop_morale() 內的 morale_cnv[3] = { MORALE_LOW, MORALE_NORMAL,
// MORALE_HIGH }(play.c:644):把「士氣值」轉成排序用的嚴重度,因為 MORALE_* 常數
// 本身的數值順序(NORMAL=0 LOW=1 HIGH=2)不是嚴重度順序(LOW 最差、HIGH 最好)。
// moraleRank[nm] 越小代表越差;取全隊最差士氣時要比較這個排序值,不是 nm 本身。
var moraleRank = [3]int{MoraleLow, MoraleNormal, MoraleHigh}

// TroopMorale 計算隊伍第 slot 格部隊的士氣,逐字對應 C troop_morale(play.c:642-658):
// 掃描全隊(從第 0 格開始,含 slot 自己;遇到 Count==0 立刻停止,對齊 C 假設隊伍
// 前段緊湊、無中間空格的慣例),對每支友軍用 moraleChart[友軍群][本部隊群] 查出
// 一個士氣值,取「嚴重度最差」的那個回傳(不是掃到第一個就停,是取最小 rank)。
// slot/troop 越界或資產未載入時回傳 MoraleHigh(對齊 C 初始值,不影響空格顯示——
// 呼叫端只在 Count>0 時才使用回傳值)。
func TroopMorale(gs *GameState, a *kbdata.Assets, slot int) int {
	if gs == nil || a == nil || slot < 0 || slot >= len(gs.Army) {
		return MoraleHigh
	}
	troopID := gs.Army[slot].TroopID
	if troopID < 0 || troopID >= len(a.Troops) {
		return MoraleHigh
	}
	groupI := a.Troops[troopID].MoraleGroup

	morale := MoraleHigh
	for j := 0; j < len(gs.Army); j++ {
		if gs.Army[j].Count == 0 {
			break
		}
		ctroopID := gs.Army[j].TroopID
		if ctroopID < 0 || ctroopID >= len(a.Troops) {
			continue
		}
		groupJ := a.Troops[ctroopID].MoraleGroup
		if groupI < 0 || groupI >= 5 || groupJ < 0 || groupJ >= 5 {
			continue
		}
		nm := moraleChart[groupJ][groupI]
		if moraleRank[nm] < moraleRank[morale] {
			morale = nm
		}
	}
	return morale
}
