package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
)

// TestBridgeScreen_StartsInPickMode 驗證 NewBridgeScreen 一開始是選方向模式。
func TestBridgeScreen_StartsInPickMode(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("WorldMap 未載入,略過")
	}

	s := NewBridgeScreen(gs, a)
	if s.mode != bridgeModePick {
		t.Fatalf("初始 mode = %v, want bridgeModePick", s.mode)
	}
}

// TestBridgeScreen_DirectionBuildsAndSwitchesToResult 驗證選方向後會架橋並切到結果模式,
// 回傳 KindStay(對齊 C build_bridge 選完方向後留在原畫面顯示結果)。
func TestBridgeScreen_DirectionBuildsAndSwitchesToResult(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("WorldMap 未載入,略過")
	}
	gs.Continent = 0
	gs.X, gs.Y = 10, 10
	gs.WorldMap.Set(0, 11, 10, 0x14) // 右方水域,確保會架橋

	s := NewBridgeScreen(gs, a)
	tr := s.Update(input.Action{Kind: input.ActRight})
	if tr.Kind != KindStay {
		t.Fatalf("選方向後 Update.Kind = %v, want KindStay", tr.Kind)
	}
	if s.mode != bridgeModeResult {
		t.Fatalf("選方向後 mode = %v, want bridgeModeResult", s.mode)
	}
}

// TestBridgeScreen_ResultModeAnyActionPops 驗證進結果模式後任何非 none 動作都會 Pop。
func TestBridgeScreen_ResultModeAnyActionPops(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("WorldMap 未載入,略過")
	}
	gs.Continent = 0
	gs.X, gs.Y = 10, 10
	gs.WorldMap.Set(0, 11, 10, 0x14)

	s := NewBridgeScreen(gs, a)
	s.Update(input.Action{Kind: input.ActRight}) // 切到結果模式

	tr := s.Update(input.Action{Kind: input.ActConfirm})
	if tr.Kind != KindPop {
		t.Fatalf("結果模式下 Confirm 應 Pop,got %v", tr.Kind)
	}
}

// TestBridgeScreen_CancelFromPickPops 驗證選方向模式下按取消直接 Pop(法術已消耗)。
func TestBridgeScreen_CancelFromPickPops(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("WorldMap 未載入,略過")
	}

	s := NewBridgeScreen(gs, a)
	tr := s.Update(input.Action{Kind: input.ActCancel})
	if tr.Kind != KindPop {
		t.Fatalf("選方向模式下 Cancel 應 Pop,got %v", tr.Kind)
	}
}
