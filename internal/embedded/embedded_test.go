package embedded

import (
	"os"
	"path/filepath"
	"testing"
)

// Extract 應把嵌入的 free 資料落到目標目錄,關鍵檔案齊全。
func TestExtract(t *testing.T) {
	dir := t.TempDir()
	root, err := Extract(dir)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	for _, rel := range []string{
		"cjk24.bin",
		"free/troops.ini",
		"free/land.org",
		"free/tileseta.png",
		"free/tilesetb.png",
	} {
		p := filepath.Join(root, rel)
		fi, err := os.Stat(p)
		if err != nil {
			t.Errorf("缺檔 %s: %v", rel, err)
			continue
		}
		if fi.Size() == 0 {
			t.Errorf("%s 為空", rel)
		}
	}
	// 冪等:再跑一次不報錯。
	if _, err := Extract(dir); err != nil {
		t.Fatalf("重跑 Extract: %v", err)
	}
}
