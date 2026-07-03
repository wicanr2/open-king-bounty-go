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
	gs := gamestate.NewGame(a, "Tester", 0)
	return NewWorldMapScreen(gs, a), gs, a
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
func TestWorldMapScreen_DwellingPushesRecruit(t *testing.T) {
	s, _, a := newTestWorldMapScreen(t)
	if a.World == nil {
		t.Skip("land.org 未載入,略過需要世界地圖的測試")
	}

	// 找一格棲地 tile,直接把玩家放到它旁邊一步之遙,再走過去驗證 Push。
	found := false
	for y := 0; y < kbdata.LevelH && !found; y++ {
		for x := 1; x < kbdata.LevelW; x++ {
			tile := a.World.Tile(0, x, y)
			if tile >= kbdata.TileDwelling1 && tile <= kbdata.TileDwelling4 {
				s.px, s.py = x-1, y
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
