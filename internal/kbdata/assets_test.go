package kbdata

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadEmptyDir:dir=="" 時仍回傳可用 Assets(純邏輯測試免帶資料檔)。
func TestLoadEmptyDir(t *testing.T) {
	a, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") err: %v", err)
	}
	if len(a.Troops) != 25 {
		t.Fatalf("Troops = %d, want 25", len(a.Troops))
	}
	if a.Font != nil {
		t.Fatalf("空 dir 不該有 Font")
	}
}

// TestLoadRealLayout:把扁平 testdata 組成真實佈局(dir/cjk24.bin + dir/free/*.ini),
// 驗 Load 真的把 Font 與 Strings 載進來。
func TestLoadRealLayout(t *testing.T) {
	dir := t.TempDir()
	freeDir := filepath.Join(dir, "free")
	if err := os.Mkdir(freeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	link := func(src, dst string) {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("讀 %s: %v", src, err)
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			t.Fatalf("寫 %s: %v", dst, err)
		}
	}
	link("testdata/cjk24.bin", filepath.Join(dir, "cjk24.bin"))
	for _, n := range freeIniNames {
		link("testdata/"+n+".ini", filepath.Join(freeDir, n+".ini"))
	}

	a, err := Load(dir)
	if err != nil {
		t.Fatalf("Load err: %v", err)
	}
	if a.Font == nil {
		t.Fatal("Font 未載入")
	}
	if !a.Font.Has('農') {
		t.Fatal("Font 缺「農」字")
	}
	if a.Strings["troops"] == nil {
		t.Fatal("troops.ini 未載入")
	}
	if a.Strings["spells"] == nil {
		t.Fatal("spells.ini 未載入")
	}
}
