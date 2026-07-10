package screen

import (
	"strings"
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/embedded"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

func castleTestAssets(t *testing.T) *kbdata.Assets {
	t.Helper()
	a, err := kbdata.LoadFS(embedded.FS())
	if err != nil {
		t.Fatalf("kbdata.LoadFS: %v", err)
	}
	return a
}

// TestCastleSiegeScreen_YesPushesCombat 驗證 y/Confirm 進圍攻戰(Push),n/Cancel 離開(Pop)。
func TestCastleSiegeScreen_YesPushesCombat(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)

	s := NewCastleSiegeScreen(gs, a, 3)
	if tr := s.Update(input.Letter('y')); tr.Kind != KindPush {
		t.Errorf("按 y 應 Push 戰鬥,got %v", tr.Kind)
	}

	s2 := NewCastleSiegeScreen(gs, a, 3)
	if tr := s2.Update(input.Letter('n')); tr.Kind != KindPop {
		t.Errorf("按 n 應 Pop,got %v", tr.Kind)
	}
	s3 := NewCastleSiegeScreen(gs, a, 3)
	if tr := s3.Update(input.Action{Kind: input.ActCancel}); tr.Kind != KindPop {
		t.Errorf("ESC 應 Pop,got %v", tr.Kind)
	}
}

// TestCastleSiegeScreen_MonsterVsVillainText 驗證佔據者敘述依 owner 分兩種文字。
func TestCastleSiegeScreen_MonsterVsVillainText(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)

	gs.CastleOwner[3] = gamestate.KBCastleMonsters
	m := NewCastleSiegeScreen(gs, a, 3)
	if !containsLine(m.lines, "各路怪物群") {
		t.Errorf("怪物城堡應顯示「各路怪物群」,got %v", m.lines)
	}

	gs.CastleOwner[3] = 2 // villain 2
	v := NewCastleSiegeScreen(gs, a, 3)
	if !anyLineContains(v.lines, "與其") {
		t.Errorf("惡棍城堡應顯示「X 與其」,got %v", v.lines)
	}
	// 惡棍佔領建構時應標記為已知(SaveCastleOwnerKnowledge)。
	if gs.CastleOwner[3]&gamestate.KBCastleKnown == 0 {
		t.Errorf("圍攻惡棍城堡後 owner 應標記 KBCastleKnown,got %#x", gs.CastleOwner[3])
	}
}

// TestCastleHomeScreen_Actions 驗證 A/B 各 Push 一個子畫面,ESC Pop。
func TestCastleHomeScreen_Actions(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)

	for _, r := range []rune{'a', 'b'} {
		s := NewCastleHomeScreen(gs, a)
		if tr := s.Update(input.Letter(r)); tr.Kind != KindPush {
			t.Errorf("按 %q 應 Push,got %v", r, tr.Kind)
		}
	}
	s := NewCastleHomeScreen(gs, a)
	if tr := s.Update(input.Action{Kind: input.ActCancel}); tr.Kind != KindPop {
		t.Errorf("ESC 應 Pop,got %v", tr.Kind)
	}
}

// TestCastleOwnScreen_EmptyRowsIdentical 驗證空守軍格產生完全相同的文字(排除
// 「空空空」那類疑似渲染 artifact 其實出在字串邏輯的可能——邏輯層必須逐字一致)。
func TestCastleOwnScreen_EmptyRowsIdentical(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.CastleOwner[3] = gamestate.KBCastlePlayer
	gs.CastleTroops[3] = [5]int{0, 3, 0xFF, 0xFF, 0xFF}
	gs.CastleNumbers[3] = [5]int{20, 8, 0, 0, 0}

	s := NewCastleOwnScreen(gs, a, 3) // 預設 remove 模式,讀城堡守軍
	lines := s.menuLines()
	// menuLines[0]=表頭,[1..5]=A-E 五格。C/D/E(索引 3,4,5)都是空格,去掉字母後應一致。
	stripLetter := func(l string) string { return strings.TrimPrefix(l, string(l[0])) }
	c, d, e := stripLetter(lines[3]), stripLetter(lines[4]), stripLetter(lines[5])
	if c != d || d != e {
		t.Errorf("空守軍格文字不一致(邏輯 bug):C=%q D=%q E=%q", c, d, e)
	}
	if !strings.Contains(c, "空") || !strings.Contains(c, "無") {
		t.Errorf("空格應顯示「空…無」,got %q", c)
	}
}

// TestCastleOwnScreen_ToggleAndUngarrison 驗證 Confirm 切換模式;撤離把城堡守軍收回玩家隊伍。
func TestCastleOwnScreen_ToggleAndUngarrison(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	gs.CastleOwner[3] = gamestate.KBCastlePlayer
	gs.CastleTroops[3] = [5]int{5, 0xFF, 0xFF, 0xFF, 0xFF}
	gs.CastleNumbers[3] = [5]int{12, 0, 0, 0, 0}
	// 清空玩家隊伍留空格供撤離收兵
	gs.Army = [5]gamestate.Squad{{TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}}

	s := NewCastleOwnScreen(gs, a, 3)
	if s.mode != castleModeRemove {
		t.Fatalf("初始應為撤離模式,got %d", s.mode)
	}
	s.Update(input.Action{Kind: input.ActConfirm}) // 切換到駐防
	if s.mode != castleModeGarrison {
		t.Fatalf("Confirm 後應切到駐防模式,got %d", s.mode)
	}
	s.Update(input.Action{Kind: input.ActConfirm}) // 切回撤離
	// 撤離 A 格(城堡守軍 troop5×12 → 玩家)
	s.Update(input.Letter('a'))
	if gs.Army[0].TroopID != 5 || gs.Army[0].Count != 12 {
		t.Errorf("撤離後玩家未收到兵:Army[0]=%+v", gs.Army[0])
	}
	if gs.CastleNumbers[3][0] != 0 {
		t.Errorf("撤離後城堡該格未清空:CastleNumbers[3][0]=%d", gs.CastleNumbers[3][0])
	}
}

// TestAudienceScreen_PageAdvanceAndPromotion 驗證翻頁 + 達門檻晉升。
func TestAudienceScreen_PageAdvanceAndPromotion(t *testing.T) {
	a := castleTestAssets(t)
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	rankBefore := gs.Rank
	// 起手 rank 0、下一階 villains_needed=2,PlayerCaptured=0 → needed=2 > 0,不晉升。
	s := NewAudienceScreen(gs, a)
	if gs.Rank != rankBefore {
		t.Errorf("未達門檻卻晉升了:rank %d→%d", rankBefore, gs.Rank)
	}
	if tr := s.Update(input.Action{Kind: input.ActConfirm}); tr.Kind != KindStay || s.page != 1 {
		t.Errorf("第一次 Confirm 應翻到第 2 頁(Stay),got kind=%v page=%d", tr.Kind, s.page)
	}
	if tr := s.Update(input.Action{Kind: input.ActConfirm}); tr.Kind != KindPop {
		t.Errorf("第 2 頁 Confirm 應 Pop,got %v", tr.Kind)
	}

	// 捕獲足夠惡棍後謁見 → 晉升。
	gs2 := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	needed := gs2.VillainsNeeded(a)
	for i := 0; i < needed && i < len(gs2.VillainCaught); i++ {
		gs2.VillainCaught[i] = 1
	}
	NewAudienceScreen(gs2, a) // 建構時就會 Promote
	if gs2.Rank != 1 {
		t.Errorf("達門檻謁見後應晉升到 rank 1,got %d", gs2.Rank)
	}
}

func containsLine(lines []string, want string) bool {
	for _, l := range lines {
		if l == want {
			return true
		}
	}
	return false
}

func anyLineContains(lines []string, sub string) bool {
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return true
		}
	}
	return false
}
