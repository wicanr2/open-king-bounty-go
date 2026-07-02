// Package mobile 是 ebitenmobile bind 的進入點(Android AAR)。
// 用 `ebitenmobile bind -target android -o openkb.aar ./mobile` 產生,
// 由薄殼 Android Activity 載入。桌面走 cmd/openkb,兩者共用 internal/。
package mobile

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/mobile"
)

const (
	logicalW = 320
	logicalH = 200
)

type game struct{}

func (g *game) Update() error { return nil }

func (g *game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "openkb mobile — P0 scaffold")
}

func (g *game) Layout(outsideW, outsideH int) (int, int) {
	return logicalW, logicalH
}

func init() {
	// SetGame 讓 ebitenmobile 生成的 Java/Kotlin 綁定驅動這顆遊戲迴圈。
	mobile.SetGame(&game{})
}

// Dummy 是 gomobile bind 要求 package 至少匯出一個符號的佔位;實際入口在 init()。
func Dummy() {}
