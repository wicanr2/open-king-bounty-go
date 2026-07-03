// Package mobile 是 ebitenmobile bind 的 Android 進入點。
// 用 `ebitenmobile bind -target android -o openkb.aar ./mobile` 產生 AAR,
// 由薄殼 Android Activity 載入。資料走嵌入(embedded)→ 解到 app 可寫目錄 → 載入。
// 桌面走 cmd/openkb,兩者共用 internal/app 與 internal/screen 等。
package mobile

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2/mobile"

	"github.com/wicanr2/open-king-bounty-go/internal/app"
	"github.com/wicanr2/open-king-bounty-go/internal/embedded"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
	"github.com/wicanr2/open-king-bounty-go/internal/screen"
)

func init() {
	// Android 上沒有外部資料目錄:把嵌入的 free 資料解到 app 可寫目錄後載入。
	dir, err := embedded.Extract(writableDir())
	if err != nil {
		log.Printf("extract embedded assets: %v", err)
	}
	assets, err := kbdata.Load(dir)
	if err != nil {
		log.Printf("load assets: %v", err)
	}
	if ts, err := render.LoadTileset(dir); err == nil {
		screen.SetTileset(ts)
	}
	// SetGame 讓 ebitenmobile 生成的 Java/Kotlin 綁定驅動這顆遊戲迴圈。
	mobile.SetGame(app.New(screen.NewTitleScreen(assets), assets))
}

// writableDir 回傳可寫的資料落地目錄。Android 上 UserConfigDir 由殼層環境提供;
// 取不到時退回暫存目錄(仍可寫)。
func writableDir() string {
	if d, err := os.UserConfigDir(); err == nil {
		return d
	}
	return os.TempDir()
}

// Dummy 是 gomobile bind 要求 package 至少匯出一個符號的佔位;實際入口在 init()。
func Dummy() {}
