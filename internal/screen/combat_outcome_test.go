package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/combat"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// forceWin 把 side 1 全部單位歸零(模擬 AI 被殲滅),讓 Winner()==0(玩家勝)。
func zeroSide(c *combat.Combat, side int) {
	for i := 0; i < combat.MaxUnits; i++ {
		c.Units[side][i].Count = 0
		c.Units[side][i].TurnCount = 0
	}
}

// TestApplyOutcome_SiegeCapture 驗證圍攻戰勝:奪城(owner=玩家)、spoils 入帳、
// 守軍首腦正是契約目標時履約(villain_caught + contract 清空)。
func TestApplyOutcome_SiegeCapture(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	// 城堡 3 由 villain 2 佔領,守軍兩格。契約正是 villain 2。
	gs.CastleOwner[3] = 2
	gs.CastleTroops[3] = [5]int{3, 5, 0xFF, 0xFF, 0xFF}
	gs.CastleNumbers[3] = [5]int{10, 8, 0, 0, 0}
	gs.Contract = 2
	gs.ContractCycle = [5]byte{2, 1, 0, 3, 0xFF}
	gs.MaxContract = 5
	goldBefore := gs.Gold

	s := NewCombatScreenSiege(gs, a, 3)
	spoils := s.combat.Spoils[1]
	zeroSide(s.combat, 1) // 敵軍全滅 → 玩家勝
	// 驅動一次 Update 讓 Winner 判定 + applyOutcome。
	s.Update(input.Action{Kind: input.ActConfirm})

	if s.result != combat.ResultPlayerWon {
		t.Fatalf("應判玩家勝,result=%d", s.result)
	}
	if gs.CastleOwner[3] != gamestate.KBCastlePlayer {
		t.Errorf("圍攻勝利未奪城:owner=%#x", gs.CastleOwner[3])
	}
	if gs.VillainCaught[2] != 1 {
		t.Errorf("契約目標 villain 2 未標記已捕獲")
	}
	if gs.Contract != 0xFF {
		t.Errorf("履約後契約未清空:%#x", gs.Contract)
	}
	// spoils 入帳 + 履約賞金(FulfillContract 另加 villain reward)。
	villains := gamestate.LoadVillains(a)
	wantGold := goldBefore + spoils + villains[2].Reward
	if gs.Gold != wantGold {
		t.Errorf("金幣結算錯誤:got %d, want %d(base %d + spoils %d + reward %d)",
			gs.Gold, wantGold, goldBefore, spoils, villains[2].Reward)
	}
}

// TestApplyOutcome_FoeWinClearsTile 驗證 foe 戰勝清地圖 tile + 存活敵軍寫回。
func TestApplyOutcome_FoeWinClearsTile(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	cont, fx, fy, foeID := 0, 5, 6, 1
	gs.FoeTroops[cont][foeID] = [5]int{3, 0xFF, 0xFF, 0xFF, 0xFF}
	gs.FoeNumbers[cont][foeID] = [5]int{9, 0, 0, 0, 0}
	// 確保該 tile 非 0(放一個 foe tile 值),戰勝後應被清 0。
	gs.WorldMap.Set(cont, fx, fy, 0x91)

	var foe [combat.MaxUnits]gamestate.Squad
	foe[0] = gamestate.Squad{TroopID: 3, Count: 9}
	for i := 1; i < combat.MaxUnits; i++ {
		foe[i] = gamestate.Squad{TroopID: 255}
	}
	s := NewCombatScreenFoe(gs, a, foe, cont, foeID, fx, fy)
	zeroSide(s.combat, 1)
	s.Update(input.Action{Kind: input.ActConfirm})

	if s.result != combat.ResultPlayerWon {
		t.Fatalf("應判玩家勝,result=%d", s.result)
	}
	if got := gs.WorldMap.Tile(cont, fx, fy); got != 0 {
		t.Errorf("foe 戰勝後 tile 應清 0,got %#x", got)
	}
	if gs.FoeNumbers[cont][foeID][0] != 0 {
		t.Errorf("存活敵軍未寫回(應全滅為 0):%d", gs.FoeNumbers[cont][foeID][0])
	}
}

// TestApplyOutcome_DefeatTempDeath 驗證戰敗觸發 temp_death(隊伍清空 + 20 農夫)。
func TestApplyOutcome_DefeatTempDeath(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	cont, fx, fy, foeID := 0, 5, 6, 1
	var foe [combat.MaxUnits]gamestate.Squad
	foe[0] = gamestate.Squad{TroopID: 3, Count: 50}
	for i := 1; i < combat.MaxUnits; i++ {
		foe[i] = gamestate.Squad{TroopID: 255}
	}
	s := NewCombatScreenFoe(gs, a, foe, cont, foeID, fx, fy)
	zeroSide(s.combat, 0) // 玩家全滅 → 戰敗
	s.Update(input.Action{Kind: input.ActConfirm})

	if s.result != combat.ResultAIWon {
		t.Fatalf("應判玩家戰敗,result=%d", s.result)
	}
	if gs.Army[0].TroopID != 0 || gs.Army[0].Count != 20 {
		t.Errorf("戰敗後應剩 20 農夫,got %+v", gs.Army[0])
	}
}
