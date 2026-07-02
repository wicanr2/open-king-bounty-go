// openkb 桌面進入點(P0:先開得起視窗,遊戲迴圈之後接 screen 狀態機)。
package main

import (
	"flag"
	"image/color"
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// 原版邏輯解析度 320x200(DOS/free 共用),放大後給桌面/手機。
const (
	logicalW = 320
	logicalH = 200
)

// defaultDataDir 目前寫死指到 openkb-code(oracle repo)的原始資料目錄,方便
// P0 demo 直接跑;之後應改成 App 安裝目錄旁的 data/。用 -datadir 可覆寫。
const defaultDataDir = "/home/anr2/openkb/openkb-code/data"

// demoTileFile 是 P0 demo 用來證明「free tile 畫得出來」的樣本檔
// (data/free/cstl.png ── 城堡 tile)。
const demoTileFile = "cstl.png"

var datadir = flag.String("datadir", defaultDataDir, "遊戲資料目錄(cjk24.bin / free/*.ini / free/*.png 所在)")

// Game 是 Ebiten 遊戲迴圈的最外層;P0 載入 kbdata.Assets,畫一行 CJK 字
// 證明點陣字管線可跑,以及一張 free tile 證明 PNG 管線可跑。
// 之後由 internal/screen 的狀態機接管 Update/Draw。
type Game struct {
	assets  *kbdata.Assets
	tile    *ebiten.Image
	loadErr error
}

// NewGame 載入 dir 下的遊戲資料。找不到資料(dir 不存在、缺檔)不 crash ──
// kbdata.Load 本身是 best-effort,Draw 端再依 assets.Font 是否為 nil 決定
// 要不要退回文字提示畫面。
func NewGame(dir string) *Game {
	g := &Game{}
	assets, err := kbdata.Load(dir)
	if err != nil {
		g.loadErr = err
		return g
	}
	g.assets = assets

	if tile, err := render.LoadPNGTile(filepath.Join(dir, "free", demoTileFile)); err == nil {
		g.tile = tile
	}
	return g
}

func (g *Game) Update() error { return nil }

func (g *Game) Draw(screen *ebiten.Image) {
	if g.assets == nil || g.assets.Font == nil {
		msg := "openkb (Go/Ebiten) — 找不到遊戲資料(cjk24.bin)"
		if g.loadErr != nil {
			msg += ": " + g.loadErr.Error()
		}
		ebitenutil.DebugPrint(screen, msg)
		return
	}

	render.DrawText(screen, g.assets.Font, "御封戰將", 8, 8, color.White)

	if g.tile != nil {
		render.DrawTile(screen, g.tile, 8, 40)
	} else {
		ebitenutil.DebugPrintAt(screen, "(tile 未載入)", 8, 180)
	}
}

// Layout 回傳邏輯解析度;Ebiten 負責縮放到實際視窗/螢幕(桌面與手機共用)。
func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	return logicalW, logicalH
}

func main() {
	flag.Parse()

	ebiten.SetWindowSize(logicalW*3, logicalH*3)
	ebiten.SetWindowTitle("御封戰將 (openkb-go)")
	if err := ebiten.RunGame(NewGame(*datadir)); err != nil {
		log.Fatal(err)
	}
}
