package screen

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// townLetterRects 應把 A-E 五個熱區疊在底部選單框的 A-E 文字行上
// (menuLines 索引 2..6,行 y = bottomHeaderY + line*CJKCell),與 drawBottomFrame 排版一致。
func TestTownLetterRectsAlignMenuRows(t *testing.T) {
	rects := townLetterRects()
	if len(rects) != len(townLetters) {
		t.Fatalf("want %d rects, got %d", len(townLetters), len(rects))
	}
	for i := range rects {
		wantY := bottomHeaderY + (2+i)*render.CJKCell // A=line2 … E=line6
		if rects[i].Y != wantY {
			t.Errorf("letter %d Y=%d, want %d(疊在選單第 %d 行)", i, rects[i].Y, wantY, 2+i)
		}
		if rects[i].X != bottomTextX || rects[i].H != render.CJKCell {
			t.Errorf("letter %d rect=%+v,X/H 應對齊選單文字行(X=%d H=%d)", i, rects[i], bottomTextX, render.CJKCell)
		}
	}
}
