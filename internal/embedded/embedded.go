// Package embedded 把 free 遊戲資料(cjk 字、free ini、land.org、tileset)嵌入 binary,
// 讓行動裝置(Android)不必外部檔案即可執行。只嵌自由授權資料,版權素材(DOS/Genesis/
// Amiga 美術、FM-Towns 音樂)一律不嵌,由完整版另外提供。
package embedded

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:data
var dataFS embed.FS

// FS 回傳以 data/ 為根的唯讀檔案系統(路徑如 "cjk24.bin"、"free/troops.ini"),
// 供 kbdata.LoadFS / render.LoadTilesetFS 直接讀,行動裝置免解壓到檔案系統(避開唯讀 cwd)。
func FS() fs.FS {
	sub, err := fs.Sub(dataFS, "data")
	if err != nil {
		return dataFS
	}
	return sub
}

// Extract 把嵌入資料寫到 destDir 下的 openkb-data/,回傳該路徑供 kbdata.Load 使用。
// 冪等:已存在的檔跳過(適合 Android app 每次啟動呼叫)。
func Extract(destDir string) (string, error) {
	root := filepath.Join(destDir, "openkb-data")
	err := fs.WalkDir(dataFS, "data", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("data", p)
		target := filepath.Join(root, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if _, e := os.Stat(target); e == nil {
			return nil // 已存在,跳過
		}
		b, e := dataFS.ReadFile(p)
		if e != nil {
			return e
		}
		if e := os.MkdirAll(filepath.Dir(target), 0o755); e != nil {
			return e
		}
		return os.WriteFile(target, b, 0o644)
	})
	return root, err
}
