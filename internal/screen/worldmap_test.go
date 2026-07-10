package screen

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// newTestWorldMapScreen 用真實(embedded)assets 建一個騎士角色與 WorldMapScreen,
// 並把 XDG_CONFIG_HOME 導向暫存目錄,避免存讀檔測試動到真實使用者設定目錄
// (見 internal/save/save_test.go 的 TestSaveDir 同樣手法)。
func newTestWorldMapScreen(t *testing.T) (*WorldMapScreen, *gamestate.GameState, *kbdata.Assets) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), "config"))

	a, err := kbdata.Load("")
	if err != nil {
		t.Fatalf("kbdata.Load: %v", err)
	}
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	return NewWorldMapScreen(gs, a), gs, a
}

// TestWorldMapScreen_MoveWritesGameState 驗證移動會寫回 GameState 的座標(位置已移入
// GameState,不再由 WorldMapScreen 自己維護)——這是存讀檔位置持久 + 戰敗傳送回家的前提。
// 用 embedded 世界資料(Load("") 沒有 land.org,gs.WorldMap 會是 nil)。
func TestWorldMapScreen_MoveWritesGameState(t *testing.T) {
	a := castleTestAssets(t) // embedded FS,含 land.org
	gs := gamestate.NewGame(a, "Tester", 0, gamestate.DefaultWorldSeed)
	if gs.WorldMap == nil {
		t.Skip("embedded land.org 未載入,略過")
	}
	// 找一格「本身可走且右鄰也可走、兩格都非互動 tile」的純草地,把玩家放左格。
	plain := func(tile byte) bool {
		return walkable(tile) && tile != kbdata.TileTown && tile != kbdata.TileCastle &&
			!kbdata.IsCastle(tile) && tile != kbdata.TileFoe && tile != kbdata.TileChest &&
			!(tile >= kbdata.TileDwelling1 && tile <= kbdata.TileDwelling4)
	}
	found := false
	for y := 0; y < kbdata.LevelH && !found; y++ {
		for x := 0; x < kbdata.LevelW-1; x++ {
			if plain(gs.WorldMap.Tile(0, x, y)) && plain(gs.WorldMap.Tile(0, x+1, y)) {
				gs.Continent, gs.X, gs.Y = 0, x, y
				found = true
				break
			}
		}
	}
	if !found {
		t.Skip("找不到相鄰兩格純草地,略過")
	}
	s := NewWorldMapScreen(gs, a)
	startX := gs.X
	if tr := s.Update(input.Action{Kind: input.ActRight}); tr.Kind != KindStay {
		t.Fatalf("走進純草地應留在地圖(Stay),got %v", tr.Kind)
	}
	if gs.X != startX+1 {
		t.Errorf("移動未寫回 GameState:gs.X = %d, want %d", gs.X, startX+1)
	}
}

// TestWorldMapScreen_KeymapHasSaveLoadLetters 驗證 Keymap 的情境字母列含 's'(存檔)
// 與 'l'(讀檔),讓觸控層能浮出對應按鈕。
func TestWorldMapScreen_KeymapHasSaveLoadLetters(t *testing.T) {
	s, _, _ := newTestWorldMapScreen(t)
	km := s.Keymap()

	var hasSave, hasLoad bool
	for _, li := range km.Letters {
		if li.Rune == 's' {
			hasSave = true
		}
		if li.Rune == 'l' {
			hasLoad = true
		}
	}
	if !hasSave {
		t.Error("Keymap.Letters 缺少 's'(存檔)")
	}
	if !hasLoad {
		t.Error("Keymap.Letters 缺少 'l'(讀檔)")
	}
}

// TestWorldMapScreen_SaveThenLoad 驗證按 's' 存檔後按 'l' 能讀回,不 panic,
// 且結果訊息符合預期;讀檔成功會 Replace 成新的 WorldMapScreen。
func TestWorldMapScreen_SaveThenLoad(t *testing.T) {
	s, gs, _ := newTestWorldMapScreen(t)
	gs.Gold = 12345 // 讓 round-trip 有可辨識的值

	tr := s.Update(input.Letter('s'))
	if tr.Kind != KindStay {
		t.Fatalf("按 's' 後 Kind = %v, want KindStay", tr.Kind)
	}
	if s.msg != "已存檔" {
		t.Errorf("存檔後 msg = %q, want \"已存檔\"", s.msg)
	}

	tr = s.Update(input.Letter('l'))
	if tr.Kind != KindReplace {
		t.Fatalf("按 'l' 後 Kind = %v, want KindReplace", tr.Kind)
	}
	loaded, ok := tr.Next.(*WorldMapScreen)
	if !ok {
		t.Fatalf("讀檔後 Next 型別 = %T, want *WorldMapScreen", tr.Next)
	}
	if loaded.gs.Gold != 12345 {
		t.Errorf("讀回的 Gold = %d, want 12345", loaded.gs.Gold)
	}
}

// TestWorldMapScreen_LoadWithoutSave_NoPanic 驗證存檔目錄裡沒有 slot0 時,
// 按 'l' 不 panic、留在原畫面,並顯示「無存檔」。
func TestWorldMapScreen_LoadWithoutSave_NoPanic(t *testing.T) {
	s, _, _ := newTestWorldMapScreen(t)

	tr := s.Update(input.Letter('l'))
	if tr.Kind != KindStay {
		t.Fatalf("無存檔時按 'l' 後 Kind = %v, want KindStay", tr.Kind)
	}
	if s.msg != "無存檔" {
		t.Errorf("無存檔時 msg = %q, want \"無存檔\"", s.msg)
	}
}

// TestWorldMapScreen_DwellingPushesRecruit 驗證踩到棲地 tile 會 Push RecruitScreen。
//
// 找棲地 tile 要掃 gs.WorldMap(NewGame 跑過 salt_continent 之後的地圖),不能掃
// a.World(唯讀、salt 前的原始 land.org——根本沒有 dwelling tile,見
// gamestate/worldgen.go 與 docs/PORT-STATUS.md「世界生成架構計畫」的調查記錄)。
func TestWorldMapScreen_DwellingPushesRecruit(t *testing.T) {
	s, gs, a := newTestWorldMapScreen(t)
	if a.World == nil || gs.WorldMap == nil {
		t.Skip("land.org 未載入,略過需要世界地圖的測試")
	}

	// 找一格棲地 tile,直接把玩家放到它旁邊一步之遙,再走過去驗證 Push。
	found := false
	for y := 0; y < kbdata.LevelH && !found; y++ {
		for x := 1; x < kbdata.LevelW; x++ {
			tile := gs.WorldMap.Tile(0, x, y)
			if tile >= kbdata.TileDwelling1 && tile <= kbdata.TileDwelling4 {
				// 位置改由 GameState 持有(px/py 已移除),直接寫 gs 座標。
				gs.Continent, gs.X, gs.Y = 0, x-1, y
				found = true
				break
			}
		}
	}
	if !found {
		t.Skip("這份 land.org 找不到棲地 tile,略過")
	}

	tr := s.Update(input.Action{Kind: input.ActRight})
	if tr.Kind != KindPush {
		t.Fatalf("踩到棲地後 Kind = %v, want KindPush", tr.Kind)
	}
	if _, ok := tr.Next.(*RecruitScreen); !ok {
		t.Fatalf("踩到棲地後 Next 型別 = %T, want *RecruitScreen", tr.Next)
	}
}
