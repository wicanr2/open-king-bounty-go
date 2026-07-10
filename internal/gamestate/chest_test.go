package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// TestTakeChest_Navmap 驗證踩到導航海圖格 → 解鎖下一洲 + 清 tile。
func TestTakeChest_Navmap(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("land.org 未載入,略過")
	}
	// 把玩家放到洲 0 的導航海圖座標上。
	gs.Continent = 0
	gs.X, gs.Y = gs.NavmapCoords[0][0], gs.NavmapCoords[0][1]
	gs.WorldMap.Set(0, gs.X, gs.Y, 0x8B) // 任意非 0 值,開箱後應清 0

	r := gs.TakeChest(a, kbrng.NewGlibc(1))
	if r.Kind != ChestNavmap || r.Continent != 1 {
		t.Fatalf("導航海圖判定錯:%+v", r)
	}
	if gs.ContinentFound[1] != 1 {
		t.Errorf("未解鎖洲 1")
	}
	if gs.WorldMap.Tile(0, gs.X, gs.Y) != 0 {
		t.Errorf("開箱後 tile 未清 0")
	}
}

// TestTakeChest_Orb 驗證踩到寶珠格 → orb_found + 清 tile。
func TestTakeChest_Orb(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("land.org 未載入,略過")
	}
	gs.Continent = 0
	gs.X, gs.Y = gs.OrbCoords[0][0], gs.OrbCoords[0][1]
	// 確保不與 navmap 座標重疊(否則會走 navmap 分支)。
	if gs.X == gs.NavmapCoords[0][0] && gs.Y == gs.NavmapCoords[0][1] {
		t.Skip("orb 與 navmap 同格,略過")
	}
	gs.WorldMap.Set(0, gs.X, gs.Y, 0x8B)

	r := gs.TakeChest(a, kbrng.NewGlibc(1))
	if r.Kind != ChestOrb {
		t.Fatalf("寶珠判定錯:%+v", r)
	}
	if gs.OrbFound[0] != 1 {
		t.Errorf("未標記 orb_found[0]")
	}
}

// TestTakeChest_GoldChoiceApply 驗證金幣/領導力二選一的套用(gold vs leadership)。
func TestTakeChest_GoldChoiceApply(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)

	r := ChestResult{Kind: ChestGoldChoice, Gold: 500, Leadership: 10}
	goldBefore := gs.Gold
	gs.ApplyGoldChoice(r, true)
	if gs.Gold != goldBefore+500 {
		t.Errorf("取金幣未入帳:gold %d→%d", goldBefore, gs.Gold)
	}

	gs2 := NewGame(a, "Tester", 0, DefaultWorldSeed)
	baseBefore, leadBefore := gs2.BaseLeadership, gs2.Leadership
	gs2.ApplyGoldChoice(r, false)
	if gs2.BaseLeadership != baseBefore+10 || gs2.Leadership != leadBefore+10 {
		t.Errorf("換領導力應同時進 Base+Leadership:base %d→%d, lead %d→%d",
			baseBefore, gs2.BaseLeadership, leadBefore, gs2.Leadership)
	}
}

// TestTakeChest_TreasureAppliesAndClears 驗證一般寶藏格會套用某種獎勵並清 tile
// (不預設是哪一種,只驗「非 navmap/orb、tile 被清、回傳合法 Kind」)。
func TestTakeChest_TreasureAppliesAndClears(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("land.org 未載入,略過")
	}
	// 放到一個既非 navmap 也非 orb 的格。
	gs.Continent = 0
	gs.X, gs.Y = 40, 40
	if (gs.X == gs.NavmapCoords[0][0] && gs.Y == gs.NavmapCoords[0][1]) ||
		(gs.X == gs.OrbCoords[0][0] && gs.Y == gs.OrbCoords[0][1]) {
		gs.X, gs.Y = 41, 41
	}
	gs.WorldMap.Set(0, gs.X, gs.Y, 0x8B)

	r := gs.TakeChest(a, kbrng.NewGlibc(7))
	if r.Kind == ChestNavmap || r.Kind == ChestOrb {
		t.Fatalf("非海圖/寶珠格卻判成 %v", r.Kind)
	}
	if gs.WorldMap.Tile(0, gs.X, gs.Y) != 0 {
		t.Errorf("開箱後 tile 未清 0")
	}
}
