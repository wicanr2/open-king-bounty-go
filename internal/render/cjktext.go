// cjktext.go -- 把 kbdata.CJKAtlas 的點陣字 alpha 遮罩畫成 Ebiten 貼圖。
//
// 設計:GlyphImage 只做一件事 ── alpha byte 轉「白色前景 + alpha 透明」的
// *ebiten.Image 遮罩圖,並依 *CJKGlyph 指標快取,不重複轉檔。換顏色不必
// 重新產圖:DrawText 用 ebiten.DrawImageOptions.ColorScale.ScaleWithColor(clr)
// 對這張白色遮罩做乘法上色(白 x clr = clr,alpha 不受影響),同一張字圖能被
// 任何顏色重複使用。
package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// glyphImageCache 以 *CJKGlyph 指標為 key(來自同一個 atlas 的查詢結果指標穩定,
// 見 kbdata.CJKAtlas.Glyph 回傳 &a.glyphs[idx]),避免同一個字每次 DrawText 都
// 重新跑一次 24x24 pixel 轉檔迴圈。
var glyphImageCache = make(map[*kbdata.CJKGlyph]*ebiten.Image)

// GlyphImage 把單一 CJKGlyph 的 Width x Height alpha 遮罩轉成 *ebiten.Image。
// RGB 固定白 (255,255,255),A 直接取 alpha byte 當覆蓋率 ── 背景 alpha=0 即透明。
// 找不到資料(nil 或尺寸不合法)回 nil。
func GlyphImage(g *kbdata.CJKGlyph) *ebiten.Image {
	if g == nil || g.Width <= 0 || g.Height <= 0 || len(g.Alpha) < g.Width*g.Height {
		return nil
	}
	if cached, ok := glyphImageCache[g]; ok {
		return cached
	}
	// 用 NRGBA(straight alpha)而非 RGBA(premultiplied)。點陣字是「白前景 + 覆蓋率 alpha」,
	// 若用 premultiplied 的 RGBA{255,255,255,a},a<255 時為非法值(RGB>A),連 alpha=0 的
	// 背景像素 RGB 仍是 255 → ebiten 以 premult 混合會把整格畫成白塊。NRGBA 讓 ebiten 自己
	// 正確 premultiply,alpha=0 即透明、中間值即抗鋸齒白筆畫。
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

// DrawText 把字串 s 逐 rune 查 atlas 畫到 dst,起點為左上角 (x, y)。
// 24x24 點陣字無比例字距,每個字(不論查得到與否)固定前進 atlas.Width 像素,
// 保留排版對齊;查不到的字(缺字)跳過不畫,但仍佔位前進。
// clr 透過 ColorScale 對 GlyphImage 的白色遮罩上色(見檔頭註解)。
func DrawText(dst *ebiten.Image, atlas *kbdata.CJKAtlas, s string, x, y int, clr color.Color) {
	if dst == nil || atlas == nil {
		return
	}
	cx := x
	for _, r := range s {
		if g, ok := atlas.Glyph(r); ok {
			if img := GlyphImage(g); img != nil {
				var op ebiten.DrawImageOptions
				op.GeoM.Translate(float64(cx), float64(y))
				op.ColorScale.ScaleWithColor(clr)
				dst.DrawImage(img, &op)
			}
		}
		cx += atlas.Width
	}
}
