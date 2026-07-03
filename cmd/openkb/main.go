// openkb 桌面進入點:載入資料 → 以 screen.Manager 跑畫面流程(title→選角→遊戲中)。
// 鍵盤事件轉成 input.Action 餵給 Manager;觸控(手機)之後在 input 層接同一套 Action。
package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
	"github.com/wicanr2/open-king-bounty-go/internal/screen"
)

// 原版邏輯解析度 320x200(DOS/free 共用),放大後給桌面/手機。
const (
	logicalW = 320
	logicalH = 200
)

const defaultDataDir = "/home/anr2/openkb/openkb-code/data"

var (
	datadir     = flag.String("datadir", defaultDataDir, "遊戲資料目錄(cjk24.bin / free/*.ini / free/*.png)")
	startClass  = flag.Int("startclass", -1, "debug:>=0 直接以該職業建角進世界地圖(截圖驗證用)")
	startCombat = flag.Bool("startcombat", false, "debug:直接切入一場戰鬥(截圖驗證用)")
)

// Game 把 Ebiten 迴圈接到 screen.Manager。
type Game struct {
	mgr    *screen.Manager
	assets *kbdata.Assets
}

func (g *Game) Update() error {
	// 鍵盤優先;無鍵盤事件才看觸控/滑鼠(讀當前畫面 Keymap 命中觸控控制)。
	if a := pollAction(); !a.IsNone() {
		g.mgr.Update(a)
		return nil
	}
	if tx, ty, ok := pollTap(); ok {
		layout := input.NewTouchLayout(g.mgr.Keymap())
		if a := layout.Resolve(input.Tap{X: tx, Y: ty}); !a.IsNone() {
			g.mgr.Update(a)
		}
	}
	return nil
}

func (g *Game) Draw(dst *ebiten.Image) {
	g.mgr.Draw(dst)
	// 觸控疊層:照當前畫面 Keymap 浮出控制(桌面滑鼠亦可點,先驗)。
	var font *kbdata.CJKAtlas
	if g.assets != nil {
		font = g.assets.Font
	}
	render.DrawTouchControls(dst, input.NewTouchLayout(g.mgr.Keymap()).Controls(), font)
}

// pollTap 回傳本幀剛按下的觸控/滑鼠點(邏輯座標)。手機用觸控,桌面用滑鼠左鍵。
func pollTap() (int, int, bool) {
	if ids := inpututil.AppendJustPressedTouchIDs(nil); len(ids) > 0 {
		x, y := ebiten.TouchPosition(ids[0])
		return x, y, true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return x, y, true
	}
	return 0, 0, false
}

func (g *Game) Layout(outsideW, outsideH int) (int, int) { return logicalW, logicalH }

// pollAction 把「本幀剛按下」的按鍵轉成單一 input.Action(中性 Key → Action)。
func pollAction() input.Action {
	symOf := map[ebiten.Key]input.Sym{
		ebiten.KeyArrowUp:    input.SymUp,
		ebiten.KeyArrowDown:  input.SymDown,
		ebiten.KeyArrowLeft:  input.SymLeft,
		ebiten.KeyArrowRight: input.SymRight,
		ebiten.KeyEnter:      input.SymEnter,
		ebiten.KeyNumpadEnter: input.SymEnter,
		ebiten.KeySpace:      input.SymSpace,
		ebiten.KeyEscape:     input.SymEsc,
		ebiten.KeyPageUp:     input.SymPageUp,
		ebiten.KeyPageDown:   input.SymPageDown,
		ebiten.KeyHome:       input.SymHome,
		ebiten.KeyEnd:        input.SymEnd,
		ebiten.KeyF8:         input.SymF8,
		ebiten.KeyF9:         input.SymF9,
		ebiten.KeyF10:        input.SymF10,
	}
	for k, sym := range symOf {
		if inpututil.IsKeyJustPressed(k) {
			return input.KeyToAction(input.Key{Sym: sym})
		}
	}
	// 字母 a–z
	for k := ebiten.KeyA; k <= ebiten.KeyZ; k++ {
		if inpututil.IsKeyJustPressed(k) {
			return input.KeyToAction(input.Key{Sym: input.SymChar, Ch: rune('a' + (k - ebiten.KeyA))})
		}
	}
	// 數字 0–9
	for k := ebiten.KeyDigit0; k <= ebiten.KeyDigit9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			return input.KeyToAction(input.Key{Sym: input.SymChar, Ch: rune('0' + (k - ebiten.KeyDigit0))})
		}
	}
	return input.None
}

func rootScreen(a *kbdata.Assets) screen.Screen {
	if *startCombat {
		gs := gamestate.NewGame(a, "Sir Loin", 0)
		return screen.NewDebugCombatScreen(gs, a)
	}
	if *startClass >= 0 && *startClass < 4 {
		gs := gamestate.NewGame(a, "Sir Loin", *startClass)
		return screen.NewWorldMapScreen(gs, a)
	}
	return screen.NewTitleScreen(a)
}

func main() {
	flag.Parse()

	assets, err := kbdata.Load(*datadir)
	if err != nil {
		log.Printf("load assets: %v(以空資料續行)", err)
	}
	// 世界地圖 tileset(缺檔則 worldmap 退回色塊繪製)。
	if ts, err := render.LoadTileset(*datadir); err == nil {
		screen.SetTileset(ts)
	} else {
		log.Printf("load tileset: %v(地圖退回色塊)", err)
	}

	g := &Game{mgr: screen.NewManager(rootScreen(assets)), assets: assets}

	ebiten.SetWindowSize(logicalW*3, logicalH*3)
	ebiten.SetWindowTitle("御封戰將 (openkb-go)")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
