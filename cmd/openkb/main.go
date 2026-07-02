// openkb 桌面進入點(P0:先開得起視窗,遊戲迴圈之後接 screen 狀態機)。
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// 原版邏輯解析度 320x200(DOS/free 共用),放大後給桌面/手機。
const (
	logicalW = 320
	logicalH = 200
)

// Game 是 Ebiten 遊戲迴圈的最外層;P0 僅畫一行字證明視窗與迴圈可跑。
// 之後由 internal/screen 的狀態機接管 Update/Draw。
type Game struct{}

func (g *Game) Update() error { return nil }

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "openkb (Go/Ebiten) — P0 scaffold")
}

// Layout 回傳邏輯解析度;Ebiten 負責縮放到實際視窗/螢幕(桌面與手機共用)。
func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	return logicalW, logicalH
}

func main() {
	ebiten.SetWindowSize(logicalW*3, logicalH*3)
	ebiten.SetWindowTitle("御封戰將 (openkb-go)")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
