// cjktext.go -- CJK 雙層合成文字渲染(移植自 C openkb env-sdl.c 的 1oom CJK 模型)。
//
// 為什麼是「雙層」:老遊戲固定低解析(320×200),若把 24×24 中文字直接畫進 320 邏輯
// 緩衝再整體放大 → 糊。C openkb 的解法(inprint.c + env-sdl.c cjk_draw_overlay):
//   1. 遊戲美術/UI/ASCII 畫進 320×200 層 → nearest 放大到輸出。
//   2. CJK 不進 320 層,而是記進 draw-list,最後在「輸出解析度」以 24×24 atlas glyph
//      + 4 方位黑外框(±1)+ 字色單獨疊上 → 任何彩色背景上都銳利可讀。
// 對齊 C:CJK 佔 1 個 ASCII 格寬(CJKCell=8 邏輯,見 inprint.c cw=ch=s_rect;
// d_rect.x += s_rect.w)。SCALE=3 時 8 邏輯格 ×3 = 24 輸出,glyph 24×24 剛好 1:1 像素完美。
package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// CJKCell 是單一 CJK 字在 320 邏輯空間的格寬/格高(對齊 C 的 ASCII 格 s_rect=8)。
const CJKCell = 8

// glyphImageCache 以 *CJKGlyph 指標為 key 快取白色遮罩圖(換色用 ColorScale,不重轉檔)。
var glyphImageCache = make(map[*kbdata.CJKGlyph]*ebiten.Image)

// cjkEntry 是 draw-list 一筆:glyph + 320 邏輯座標 + 字色。
type cjkEntry struct {
	g    *kbdata.CJKGlyph
	x, y int
	clr  color.Color
}

// cjkList 是本幀累積的 CJK draw-list;由 app 於美術層放大後 FlushCJK 疊上,ResetCJK 清空。
var cjkList []cjkEntry

// GlyphImage 把單一 CJKGlyph 的 alpha 遮罩轉成白色前景 *ebiten.Image(見舊註解:NRGBA straight alpha)。
func GlyphImage(g *kbdata.CJKGlyph) *ebiten.Image {
	if g == nil || g.Width <= 0 || g.Height <= 0 || len(g.Alpha) < g.Width*g.Height {
		return nil
	}
	if cached, ok := glyphImageCache[g]; ok {
		return cached
	}
	img := image.NewNRGBA(image.Rect(0, 0, g.Width, g.Height))
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			a := g.Alpha[y*g.Width+x]
			img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: a})
		}
	}
	eimg := ebiten.NewImageFromImage(img)
	glyphImageCache[g] = eimg
	return eimg
}

// DrawText 不再立即繪製,而是把字串的每個 CJK 字記進 draw-list(320 邏輯座標),
// 由 app 在 FlushCJK 統一於輸出解析度畫出(雙層合成)。dst 參數保留相容,不使用。
// 起點 (x,y) 為左上角;每字前進 CJKCell(對齊 C 的 1 ASCII 格),缺字跳過但仍佔位。
func DrawText(dst *ebiten.Image, atlas *kbdata.CJKAtlas, s string, x, y int, clr color.Color) {
	if atlas == nil {
		return
	}
	cx := x
	for _, r := range s {
		if g, ok := atlas.Glyph(r); ok {
			cjkList = append(cjkList, cjkEntry{g: g, x: cx, y: y, clr: clr})
		}
		cx += CJKCell
	}
}

// ResetCJK 清空 draw-list(每幀開始呼叫)。
func ResetCJK() { cjkList = cjkList[:0] }

// FlushCJK 把本幀 draw-list 以 scale 倍畫到 dst(輸出解析度):每字 24×24 glyph 置於
// (x*scale, y*scale),先畫 4 方位黑外框(±1 輸出像素)再畫字色。對齊 C cjk_draw_overlay。
// 字色過暗(亮度<60)自動提亮成 235,確保深色背景上可讀。
func FlushCJK(dst *ebiten.Image, scale int) {
	if dst == nil {
		return
	}
	// glyph 源恆 24×24,目標格寬 = CJKCell*scale(SCALE=3 → 24,1:1)。
	dstCell := float64(CJKCell * scale)
	var ox = [4]int{-1, 1, 0, 0}
	var oy = [4]int{0, 0, -1, 1}
	for i := range cjkList {
		e := &cjkList[i]
		img := GlyphImage(e.g)
		if img == nil {
			continue
		}
		gw := img.Bounds().Dx()
		sc := dstCell / float64(gw) // 24→dstCell 縮放(SCALE=3 時為 1.0)
		bx := float64(e.x * scale)
		by := float64(e.y * scale)
		// 4 方位黑外框(±1 輸出像素)
		for k := 0; k < 4; k++ {
			var op ebiten.DrawImageOptions
			op.GeoM.Scale(sc, sc)
			op.GeoM.Translate(bx+float64(ox[k]), by+float64(oy[k]))
			op.ColorScale.ScaleWithColor(color.Black)
			if sc != 1 {
				op.Filter = ebiten.FilterLinear
			}
			dst.DrawImage(img, &op)
		}
		// 字色(過暗提亮)
		clr := brighten(e.clr)
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(sc, sc)
		op.GeoM.Translate(bx, by)
		op.ColorScale.ScaleWithColor(clr)
		if sc != 1 {
			op.Filter = ebiten.FilterLinear
		}
		dst.DrawImage(img, &op)
	}
}

// brighten 對齊 C:字色亮度 <60 時提亮成 (235,235,235),否則原色。
func brighten(c color.Color) color.Color {
	r, g, b, a := c.RGBA()
	r8, g8, b8 := r>>8, g>>8, b>>8
	if (r8*30+g8*59+b8*11)/100 < 60 {
		return color.RGBA{235, 235, 235, uint8(a >> 8)}
	}
	return c
}
