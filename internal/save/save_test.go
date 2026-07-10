package save

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// TestSaveLoadRoundTrip 驗證 NewGame 產生的角色經 SaveToPath → LoadFromPath 後
// 逐欄相等。全程使用 t.TempDir(),不觸碰真實 os.UserConfigDir(),避免污染。
func TestSaveLoadRoundTrip(t *testing.T) {
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatalf("kbdata.Load: %v", err)
	}

	gs := gamestate.NewGame(a, "測試騎士", 0, gamestate.DefaultWorldSeed)
	gs.Week = 3 // 讓 round-trip 也覆蓋非起手值(避免欄位剛好等於零值而測不出漏傳)
	// 玩家位置移入 GameState 後應隨存檔持久(過去存在 WorldMapScreen,讀檔會遺失)。
	gs.Continent, gs.X, gs.Y = 2, 7, 19

	path := filepath.Join(t.TempDir(), "slot0.json")

	if err := SaveToPath(gs, path); err != nil {
		t.Fatalf("SaveToPath: %v", err)
	}

	got, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	if got.Name != gs.Name {
		t.Errorf("Name: got %q, want %q", got.Name, gs.Name)
	}
	if got.Class != gs.Class {
		t.Errorf("Class: got %d, want %d", got.Class, gs.Class)
	}
	if got.Rank != gs.Rank {
		t.Errorf("Rank: got %d, want %d", got.Rank, gs.Rank)
	}
	if got.Gold != gs.Gold {
		t.Errorf("Gold: got %d, want %d", got.Gold, gs.Gold)
	}
	if got.BaseLeadership != gs.BaseLeadership {
		t.Errorf("BaseLeadership: got %d, want %d", got.BaseLeadership, gs.BaseLeadership)
	}
	if got.Leadership != gs.Leadership {
		t.Errorf("Leadership: got %d, want %d", got.Leadership, gs.Leadership)
	}
	if got.Commission != gs.Commission {
		t.Errorf("Commission: got %d, want %d", got.Commission, gs.Commission)
	}
	if got.SpellPower != gs.SpellPower {
		t.Errorf("SpellPower: got %d, want %d", got.SpellPower, gs.SpellPower)
	}
	if got.MaxSpells != gs.MaxSpells {
		t.Errorf("MaxSpells: got %d, want %d", got.MaxSpells, gs.MaxSpells)
	}
	if got.KnowsMagic != gs.KnowsMagic {
		t.Errorf("KnowsMagic: got %v, want %v", got.KnowsMagic, gs.KnowsMagic)
	}
	if got.Continent != gs.Continent || got.X != gs.X || got.Y != gs.Y {
		t.Errorf("位置未持久:got (cont=%d,x=%d,y=%d), want (%d,%d,%d)",
			got.Continent, got.X, got.Y, gs.Continent, gs.X, gs.Y)
	}
	if got.Week != gs.Week {
		t.Errorf("Week: got %d, want %d", got.Week, gs.Week)
	}
	for i := range gs.Army {
		if got.Army[i] != gs.Army[i] {
			t.Errorf("Army[%d]: got %+v, want %+v", i, got.Army[i], gs.Army[i])
		}
	}
}

// TestLoadFromPath_NotExist 驗證讀取不存在的檔案時,回傳的 error 滿足 os.IsNotExist。
func TestLoadFromPath_NotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatal("LoadFromPath: 預期回傳 error,實際為 nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("LoadFromPath: 預期 os.IsNotExist(err) 為 true,實際 err = %v", err)
	}
}

// TestSaveDir 驗證 SaveDir 回傳非空、含 "openkb" 的路徑,且目錄確實被建立。
//
// SaveDir() 內部呼叫 os.UserConfigDir() + MkdirAll,直接呼叫會動到真實使用者設定
// 目錄。這裡用 t.Setenv 把 XDG_CONFIG_HOME 導向 t.TempDir(),讓 os.UserConfigDir()
// (Linux 上優先讀 XDG_CONFIG_HOME)解析到暫存目錄,測完自動隨 TempDir 清掉,
// 不寫進真實 UserConfigDir。
func TestSaveDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	dir, err := SaveDir()
	if err != nil {
		t.Fatalf("SaveDir: %v", err)
	}

	if dir == "" {
		t.Fatal("SaveDir: 回傳空字串")
	}
	if !strings.Contains(dir, "openkb") {
		t.Errorf("SaveDir: 路徑 %q 應包含 \"openkb\"", dir)
	}

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Errorf("SaveDir: 目錄 %q 應存在(MkdirAll 後);stat err = %v", dir, err)
	}
}
