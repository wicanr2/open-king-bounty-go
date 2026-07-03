// Package app 是 Ebiten 遊戲迴圈的共用 glue:把 screen.Manager、輸入(鍵盤+觸控)、
// 觸控疊層接起來。cmd(桌面)與 mobile(Android)都用它,不重複遊戲迴圈。
package app

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
	"github.com/wicanr2/open-king-bounty-go/internal/screen"
)

// 原版邏輯解析度 320×200(DOS/free 共用);遊戲美術/UI/ASCII 都在此座標空間繪製。
// SCALE 是輸出放大倍率:美術層以 nearest 放大保持 pixel art 銳利,CJK 在放大後的輸出
// 解析度以 24×24 疊上(見 render.FlushCJK)。SCALE=3 時 CJK 8 邏輯格 ×3=24 恰好 1:1 像素完美。
const (
	LogicalW = 320
	LogicalH = 200
	SCALE    = 3
	OutputW  = LogicalW * SCALE
	OutputH  = LogicalH * SCALE
)

// Game 實作 ebiten.Game,驅動 screen.Manager + 輸入 + 觸控疊層。
type Game struct {
	mgr    *screen.Manager
	assets *kbdata.Assets
	boot   func() // 第一幀執行一次的初始化(建立 ebiten.Image 等需繪圖/JVM 就緒的工作)
	booted bool
	touch  bool           // 是否繪製觸控疊層(僅行動裝置;桌面用鍵盤,畫了會污染「與 C 一模一樣」的內容區)
	art    *ebiten.Image  // 320×200 美術層 offscreen(每幀重畫,再 nearest 放大到輸出)
}

// New 建立遊戲迴圈,root 為起始畫面。boot 可為 nil;非 nil 時於「第一次 Draw」執行一次——
// 用於行動裝置:ebiten 繪圖(如 tileset 建圖)不能在 init() 做(JVM 尚未 attach 會 panic
// devicescale: no current JVM),必須延後到遊戲迴圈啟動後的第一幀。
// touch 為 true 時才畫觸控控制疊層(Android);桌面傳 false 保持畫面與 C openkb 一致。
func New(root screen.Screen, assets *kbdata.Assets, boot func(), touch bool) *Game {
	return &Game{mgr: screen.NewManager(root), assets: assets, boot: boot, touch: touch}
}

func (g *Game) Update() error {
	// 鍵盤優先;無鍵盤事件才看觸控/滑鼠(讀當前畫面 Keymap 命中觸控控制)。
	if a := pollAction(); !a.IsNone() {
		g.mgr.Update(a)
		return nil
	}
	if tx, ty, ok := pollTap(); ok {
		// 觸控/滑鼠座標在輸出空間(960×600),換算回 320 邏輯供觸控佈局解析。
		layout := input.NewTouchLayout(g.mgr.Keymap())
		if a := layout.Resolve(input.Tap{X: tx / SCALE, Y: ty / SCALE}); !a.IsNone() {
			g.mgr.Update(a)
		}
	}
	return nil
}

func (g *Game) Draw(dst *ebiten.Image) {
	if !g.booted {
		g.booted = true
		if g.boot != nil {
			g.boot() // JVM/繪圖已就緒,此時建 ebiten.Image 安全
		}
	}
	if g.art == nil {
		g.art = ebiten.NewImage(LogicalW, LogicalH)
	}
	// 雙層合成:① 美術/UI/ASCII 畫進 320 美術層,CJK 同時記進 draw-list(不畫進美術層)。
	render.ResetCJK()
	g.art.Clear()
	g.mgr.Draw(g.art)
	if g.touch {
		var font *kbdata.CJKAtlas
		if g.assets != nil {
			font = g.assets.Font
		}
		render.DrawTouchControls(g.art, input.NewTouchLayout(g.mgr.Keymap()).Controls(), font)
	}
	// ② 美術層 nearest 放大到輸出解析度(pixel art 銳利)。
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(SCALE, SCALE)
	dst.DrawImage(g.art, &op)
	// ③ CJK 在輸出解析度以 24×24 + 黑外框疊上(銳利可讀)。
	render.FlushCJK(dst, SCALE)
}

func (g *Game) Layout(outsideW, outsideH int) (int, int) {
	return OutputW, OutputH
}

// LayoutF 是 ebiten 2.8+ 的浮點版 Layout(LayoutFer 介面)。行動裝置走高 DPI 浮點路徑,
// 只實作 int Layout 會被塞進垃圾尺寸(MinInt64)→ viewport 無效 → 全黑。實作此法即修正。
func (g *Game) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	return OutputW, OutputH
}

// pollAction 把「本幀剛按下」的按鍵轉成單一 input.Action(中性 Key → Action)。
func pollAction() input.Action {
	symOf := map[ebiten.Key]input.Sym{
		ebiten.KeyArrowUp:     input.SymUp,
		ebiten.KeyArrowDown:   input.SymDown,
		ebiten.KeyArrowLeft:   input.SymLeft,
		ebiten.KeyArrowRight:  input.SymRight,
		ebiten.KeyEnter:       input.SymEnter,
		ebiten.KeyNumpadEnter: input.SymEnter,
		ebiten.KeySpace:       input.SymSpace,
		ebiten.KeyEscape:      input.SymEsc,
		ebiten.KeyPageUp:      input.SymPageUp,
		ebiten.KeyPageDown:    input.SymPageDown,
		ebiten.KeyHome:        input.SymHome,
		ebiten.KeyEnd:         input.SymEnd,
		ebiten.KeyF8:          input.SymF8,
		ebiten.KeyF9:          input.SymF9,
		ebiten.KeyF10:         input.SymF10,
	}
	for k, sym := range symOf {
		if inpututil.IsKeyJustPressed(k) {
			return input.KeyToAction(input.Key{Sym: sym})
		}
	}
	for k := ebiten.KeyA; k <= ebiten.KeyZ; k++ {
		if inpututil.IsKeyJustPressed(k) {
			return input.KeyToAction(input.Key{Sym: input.SymChar, Ch: rune('a' + (k - ebiten.KeyA))})
		}
	}
	for k := ebiten.KeyDigit0; k <= ebiten.KeyDigit9; k++ {
		if inpututil.IsKeyJustPressed(k) {
			return input.KeyToAction(input.Key{Sym: input.SymChar, Ch: rune('0' + (k - ebiten.KeyDigit0))})
		}
	}
	return input.None
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
