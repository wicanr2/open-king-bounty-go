package screen

import (
	"testing"
	"testing/fstest"
)

// fakeThemeFS 造一個假美術檔案系統:每個給定模組目錄放一個 tileseta.png(themeExists 的判據)。
func fakeThemeFS(mods ...string) fstest.MapFS {
	m := fstest.MapFS{}
	for _, mod := range mods {
		m[mod+"/tileseta.png"] = &fstest.MapFile{Data: []byte("x")}
	}
	return m
}

func TestResolveThemes_FiltersMissingKeepsOrder(t *testing.T) {
	// 只有 amiga / free 內建,dos / genesis 缺 → 過濾掉,順序不變。
	fsys := fakeThemeFS("amiga", "free")
	got := resolveThemes(fsys, []string{"dos", "genesis", "amiga", "free"})
	want := []string{"amiga", "free"}
	if len(got) != len(want) {
		t.Fatalf("resolveThemes = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("resolveThemes[%d] = %q, want %q (full %v)", i, got[i], want[i], got)
		}
	}
}

func TestResolveThemes_DefaultIsFirstExisting(t *testing.T) {
	// dos 缺、genesis 在 → 預設(第一個)應為 genesis。
	fsys := fakeThemeFS("genesis", "free")
	got := resolveThemes(fsys, []string{"dos", "genesis", "amiga", "free"})
	if len(got) == 0 || got[0] != "genesis" {
		t.Fatalf("default theme = %v, want first=genesis", got)
	}
}

func TestResolveThemes_NoneExist(t *testing.T) {
	fsys := fakeThemeFS() // 空
	if got := resolveThemes(fsys, []string{"dos", "free"}); len(got) != 0 {
		t.Fatalf("resolveThemes on empty FS = %v, want empty", got)
	}
}

// TestThemeStartIndex 驗證起始主題 index 選擇(持久化偏好套用邏輯,不觸發 LoadArt/建圖)。
func TestThemeStartIndex(t *testing.T) {
	mods := []string{"dos", "amiga", "free"}
	cases := []struct {
		pref string
		want int
	}{
		{"", 0},        // 無偏好 → 預設(第一個 dos)
		{"dos", 0},     // 偏好=預設
		{"amiga", 1},   // 偏好在中間
		{"free", 2},    // 偏好在末尾
		{"genesis", 0}, // 偏好不在可用清單(未內建)→ 退回預設
	}
	for _, c := range cases {
		if got := themeStartIndex(mods, c.pref); got != c.want {
			t.Errorf("themeStartIndex(%v, %q) = %d, want %d", mods, c.pref, got, c.want)
		}
	}
}

// TestThemeCycleIndex 驗證循環索引邏輯(不觸發 LoadArt/建圖):直接操作套件級狀態。
func TestThemeCycleIndex(t *testing.T) {
	// 保存並還原全域狀態,避免污染其他測試。
	savedMods, savedActive, savedFS := themeMods, themeActive, themeFS
	defer func() { themeMods, themeActive, themeFS = savedMods, savedActive, savedFS }()

	themeMods = []string{"amiga", "free"}
	themeActive = 0
	themeFS = nil // nil FS → CycleTheme 早退不建圖,只驗少於前提;改用純索引檢查

	// 純索引前進(模擬 CycleTheme 的 index 部分,不呼叫會建圖的 LoadArt)。
	advance := func() { themeActive = (themeActive + 1) % len(themeMods) }
	if ActiveTheme() != "amiga" {
		t.Fatalf("start ActiveTheme = %q, want amiga", ActiveTheme())
	}
	advance()
	if ActiveTheme() != "free" {
		t.Fatalf("after 1 cycle = %q, want free", ActiveTheme())
	}
	advance()
	if ActiveTheme() != "amiga" {
		t.Fatalf("after wrap = %q, want amiga", ActiveTheme())
	}
	if got := AvailableThemes(); len(got) != 2 {
		t.Fatalf("AvailableThemes = %v, want 2", got)
	}
}
