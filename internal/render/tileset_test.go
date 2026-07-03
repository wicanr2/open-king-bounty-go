package render

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadTileset 用 testdata/ 的扁平 tileseta.png、tilesetb.png 組成真實佈局
// (dir/free/tileseta.png、dir/free/tilesetb.png),驗 LoadTileset 能成功載入,
// 對照 kbdata/assets_test.go 的 testdata 佈局慣例。
func TestLoadTileset(t *testing.T) {
	dir := t.TempDir()
	freeDir := filepath.Join(dir, "free")
	if err := os.Mkdir(freeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	copyFile := func(src, dst string) {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("讀 %s: %v", src, err)
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			t.Fatalf("寫 %s: %v", dst, err)
		}
	}
	copyFile("testdata/tileseta.png", filepath.Join(freeDir, "tileseta.png"))
	copyFile("testdata/tilesetb.png", filepath.Join(freeDir, "tilesetb.png"))

	ts, err := LoadTileset(dir)
	if err != nil {
		t.Fatalf("LoadTileset err: %v", err)
	}
	if ts == nil {
		t.Fatal("LoadTileset 回傳 nil Tileset")
	}

	checkSize := func(name string, m byte) {
		img := ts.TileImage(m)
		if img == nil {
			t.Fatalf("TileImage(%d) 為 nil", m)
		}
		b := img.Bounds()
		w, h := b.Dx(), b.Dy()
		if w != tileW || h != tileH {
			t.Fatalf("%s: TileImage(%d) 尺寸 = %dx%d, want %dx%d", name, m, w, h, tileW, tileH)
		}
	}

	// tileseta 邊界:tile 0(頭)、tile 35(尾,tileseta 最後一格)。
	checkSize("tileseta 首", 0)
	checkSize("tileseta 尾", 35)
	// tilesetb 邊界:tile 36(tilesetb 第一格)、tile 71(tilesetb 最後一格)。
	checkSize("tilesetb 首", 36)
	checkSize("tilesetb 尾", 71)
}

// TestLoadTileset_MissingDir 驗缺檔時 LoadTileset 回傳 error,不 panic。
func TestLoadTileset_MissingDir(t *testing.T) {
	dir := t.TempDir() // 空目錄,沒有 free/tileseta.png、tilesetb.png
	_, err := LoadTileset(dir)
	if err == nil {
		t.Fatal("LoadTileset 對缺檔目錄應回傳 error,但沒有")
	}
}

// TestTileImage_OutOfRange 驗 m > 71 時 TileImage 不 panic,安全回傳(nil 或 tile 0)。
func TestTileImage_OutOfRange(t *testing.T) {
	dir := t.TempDir()
	freeDir := filepath.Join(dir, "free")
	if err := os.Mkdir(freeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	copyFile := func(src, dst string) {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("讀 %s: %v", src, err)
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			t.Fatalf("寫 %s: %v", dst, err)
		}
	}
	copyFile("testdata/tileseta.png", filepath.Join(freeDir, "tileseta.png"))
	copyFile("testdata/tilesetb.png", filepath.Join(freeDir, "tilesetb.png"))

	ts, err := LoadTileset(dir)
	if err != nil {
		t.Fatalf("LoadTileset err: %v", err)
	}

	// 不應 panic;越界安全回退成有效 tile(或 nil)。
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TileImage(200) panic: %v", r)
		}
	}()
	img := ts.TileImage(200)
	if img != nil {
		b := img.Bounds()
		if b.Dx() != tileW || b.Dy() != tileH {
			t.Fatalf("TileImage(200) 越界安全值尺寸 = %dx%d, want %dx%d", b.Dx(), b.Dy(), tileW, tileH)
		}
	}

	// nil Tileset 呼叫也不該 panic。
	var nilTs *Tileset
	if nilTs.TileImage(0) != nil {
		t.Fatal("nil *Tileset.TileImage 應回傳 nil")
	}
}
