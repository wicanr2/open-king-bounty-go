// Package mobile is the Android entry point for ebitenmobile bind.
// NOTE: doc comments on EXPORTED symbols in this package are copied verbatim into
// the gobind-generated Java and compiled with US-ASCII, so they must stay ASCII-only
// (CJK here breaks `ebitenmobile bind` with "unmappable character"). Non-exported
// comments may use CJK. Build: ebitenmobile bind -target android -o openkb.aar ./mobile.
package mobile

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2/mobile"

	"github.com/wicanr2/open-king-bounty-go/internal/app"
	"github.com/wicanr2/open-king-bounty-go/internal/embedded"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
	"github.com/wicanr2/open-king-bounty-go/internal/screen"
)

func init() {
	// 資料直接從嵌入的 embed.FS 讀(不解壓到檔案系統——Android 的 cwd/temp 唯讀或不可寫)。
	// LoadFS 是純資料(不碰 ebiten),可在 init() 安全執行。
	assets, err := kbdata.LoadFS(embedded.FS())
	if err != nil {
		log.Printf("load assets: %v", err)
	}

	// tileset 會建 ebiten.Image → 必須延後到遊戲迴圈第一幀(JVM/繪圖就緒),
	// 否則在 init() 會 panic「devicescale: no current JVM」。交給 app 的 boot 回呼。
	boot := func() {
		if ts, err := render.LoadTilesetFS(embedded.FS()); err == nil {
			screen.SetTileset(ts)
		} else {
			log.Printf("load tileset: %v", err)
		}
	}

	// SetGame 讓 ebitenmobile 生成的 Java/Kotlin 綁定驅動這顆遊戲迴圈。
	mobile.SetGame(app.New(screen.NewTitleScreen(assets), assets, boot))
}

// Dummy is a placeholder so gomobile bind sees at least one exported symbol.
// The real entry point is init() (mobile.SetGame). ASCII-only on purpose (see package doc).
func Dummy() {}
