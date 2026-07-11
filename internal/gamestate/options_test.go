package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// TestOptNoWages 驗證 opt_no_wages 開啟時週結算跳過軍隊維護費扣款。
func TestOptNoWages(t *testing.T) {
	a := loadAssetsT(t)
	mk := func(noWages bool) int {
		gs := NewGame(a, "T", 0, DefaultWorldSeed)
		gs.Army[0] = Squad{TroopID: 0, Count: 50} // 50 農夫 → 有維護費
		gs.Gold = 5000
		gs.Commission = 0
		gs.OptNoWages = noWages
		gs.EndWeek(a, kbrng.NewGlibc(1))
		return gs.Gold
	}
	normal := mk(false)
	skip := mk(true)
	if !(skip > normal) {
		t.Errorf("opt_no_wages 應跳過維護費 → 金幣較多:關=%d、開=%d", normal, skip)
	}
}

// TestOptFoeStrength 驗證 opt_foe_strength 開啟時新產生敵人兵力加倍。
func TestOptFoeStrength(t *testing.T) {
	a := loadAssetsT(t)
	sum := func(strong bool) int {
		gs := NewGame(a, "T", 0, DefaultWorldSeed)
		gs.OptFoeStrength = strong
		gs.repopulateFoe(0, 0, kbrng.NewGlibc(7))
		s := 0
		for i := 0; i < 3; i++ {
			s += gs.FoeNumbers[0][0][i]
		}
		return s
	}
	normal := sum(false)
	strong := sum(true)
	if strong != normal*2 {
		t.Errorf("opt_foe_strength 應使 foe 兵力 ×2:關=%d、開=%d(want %d)", normal, strong, normal*2)
	}
}

// TestOptFoeFreq 驗證 opt_foe_freq 對 foe 每週成長的三種模式(0正常/1加倍/2停止)。
func TestOptFoeFreq(t *testing.T) {
	a := loadAssetsT(t)
	creature := 1 // 用非農夫生物(農夫=0 會走特殊路徑)
	growth := a.Troops[creature].Growth
	grow := func(freq int) int {
		gs := NewGame(a, "T", 0, DefaultWorldSeed)
		gs.OptFoeFreq = freq
		gs.FoeTroops[0][0][0] = creature
		gs.FoeNumbers[0][0][0] = 10
		gs.weeklyWorldGrowth(a, kbrng.NewGlibc(1), creature)
		return gs.FoeNumbers[0][0][0] - 10
	}
	if got := grow(0); got != growth {
		t.Errorf("foe_freq=0(正常):+%d, want +%d", got, growth)
	}
	if got := grow(1); got != growth*2 {
		t.Errorf("foe_freq=1(加倍):+%d, want +%d", got, growth*2)
	}
	if got := grow(2); got != 0 {
		t.Errorf("foe_freq=2(停止):+%d, want +0", got)
	}
}

// TestOptRecruitCaps 驗證 opt_recruit_caps 開啟時高階兵種招募數被夾到「上限-已持有」。
func TestOptRecruitCaps(t *testing.T) {
	a := loadAssetsT(t)
	const dragon = 0x18 // 火龍,cap=15
	gs := NewGame(a, "T", 0, DefaultWorldSeed)
	gs.BaseLeadership = 100000 // 領導力充足,確保上限來自 cap 而非領導力
	gs.Leadership = gs.BaseLeadership

	gs.OptRecruitCaps = false
	uncapped := gs.ArmyMaxTroopCount(a, dragon)
	gs.OptRecruitCaps = true
	capped := gs.ArmyMaxTroopCount(a, dragon)
	if capped != 15 {
		t.Errorf("opt_recruit_caps 火龍上限應為 15(已持有 0),got %d", capped)
	}
	if !(uncapped > capped) {
		t.Errorf("開啟上限後火龍可招募數應變少:關=%d、開=%d", uncapped, capped)
	}
	// 已持有 5 隻 → 上限剩 10。
	gs.Army[0] = Squad{TroopID: dragon, Count: 5}
	if got := gs.ArmyMaxTroopCount(a, dragon); got != 10 {
		t.Errorf("已持有 5 火龍 → 上限剩 10,got %d", got)
	}
}
