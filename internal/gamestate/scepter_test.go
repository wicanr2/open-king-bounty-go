package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// TestPlaceScepter 驗證埋權杖:選出合法洲、埋在草地格,AtScepter 在該格為真、他處為假。
func TestPlaceScepter(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("land.org 未載入,略過")
	}
	// NewGame 已呼叫 PlaceScepter;驗證結果落在草地。
	if gs.ScepterContinent < 0 || gs.ScepterContinent >= 4 {
		t.Fatalf("權杖洲別越界:%d", gs.ScepterContinent)
	}
	tile := gs.WorldMap.Tile(gs.ScepterContinent, gs.ScepterX, gs.ScepterY)
	if !isGrassTile(tile) {
		t.Errorf("權杖未埋在草地:tile=%#x @ (%d,%d,%d)", tile, gs.ScepterContinent, gs.ScepterX, gs.ScepterY)
	}

	// 站在權杖處 → AtScepter true。
	gs.Continent, gs.X, gs.Y = gs.ScepterContinent, gs.ScepterX, gs.ScepterY
	if !gs.AtScepter() {
		t.Error("站在權杖埋藏處 AtScepter 應為 true")
	}
	// 移開一格 → false。
	gs.X++
	if gs.AtScepter() {
		t.Error("離開權杖格後 AtScepter 應為 false")
	}
}

// TestPlaceScepter_VariesBySeed 驗證 replayability 契約:不同 seed 開局的權杖埋點
// 不會全部相同(還原原版「每局世界隨機」)。正式入口 charselect 用 RandomWorldSeed(),
// 這裡以一批固定 seed 代表不同開局,確認至少出現兩種相異埋點。
func TestPlaceScepter_VariesBySeed(t *testing.T) {
	a := loadAssetsT(t)
	if NewGame(a, "T", 0, DefaultWorldSeed).WorldMap == nil {
		t.Skip("land.org 未載入,略過")
	}
	type pos struct{ c, x, y int }
	seen := map[pos]bool{}
	for seed := uint32(1); seed <= 30; seed++ {
		gs := NewGame(a, "T", 0, seed)
		seen[pos{gs.ScepterContinent, gs.ScepterX, gs.ScepterY}] = true
	}
	if len(seen) < 2 {
		t.Errorf("30 個不同 seed 只產生 %d 種權杖埋點,replayability 失效", len(seen))
	}
}

// TestNewGame_WorldVariesBySeed 驗證整體世界(不只權杖)隨 seed 變化:town_spell
// 分配在不同 seed 下不應全同。
func TestNewGame_WorldVariesBySeed(t *testing.T) {
	a := loadAssetsT(t)
	g1 := NewGame(a, "T", 0, 1)
	g2 := NewGame(a, "T", 0, 2)
	if g1.TownSpell == g2.TownSpell && g1.ScepterX == g2.ScepterX &&
		g1.ScepterY == g2.ScepterY && g1.ScepterContinent == g2.ScepterContinent {
		t.Error("兩個不同 seed 的世界(town_spell + 權杖)完全相同,seed 未生效")
	}
}

// TestPlaceScepter_Deterministic 驗證同 seed 埋點可重現(同一 rng 序列)。
func TestPlaceScepter_Deterministic(t *testing.T) {
	a := loadAssetsT(t)
	gs := NewGame(a, "Tester", 0, DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("land.org 未載入,略過")
	}
	c, x, y := gs.ScepterContinent, gs.ScepterX, gs.ScepterY
	gs.PlaceScepter(kbrng.NewGlibc(42))
	c2, x2, y2 := gs.ScepterContinent, gs.ScepterX, gs.ScepterY
	gs.PlaceScepter(kbrng.NewGlibc(42))
	if gs.ScepterContinent != c2 || gs.ScepterX != x2 || gs.ScepterY != y2 {
		t.Error("同 seed 埋點應可重現")
	}
	_ = c
	_ = x
	_ = y
}
