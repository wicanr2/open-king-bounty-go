// cjk_atlas.go -- cjk24.bin 點陣字 atlas 載入器。
//
// C 版對照(逐位元組移植,見 openkb-code/src/cjkfont.c:1-125 與 cjkfont.h:1-31):
//
//	檔案格式(小端序):
//	  magic[8]   = "OKBCJK\0\1"           (cjkfont.c:18 CJK_MAGIC)
//	  u16 glyphW, u16 glyphH, u32 count   (cjkfont.c:71-74)
//	  count * { u32 codepoint; glyphW*glyphH bytes alpha(0..255) }
//	           (cjkfont.c:27 recsz = 4 + w*h; cjkfont.c:31-35 rec_cp 讀 codepoint)
//	codepoint 依 count 遞增排序,C 版用二分搜尋(cjkfont.c:102-121)。
//
// 每個字不是位元遮罩(1bpp),而是「一整塊 w*h 的 alpha byte 陣列」
// (每像素一個 0..255 的灰階值),cjkfont_get() 直接回傳 recs 裡對應那段
// 連續記憶體當灰階遮罩用(cjkfont.h:26-27 註解:「取得 cp 的 glyph alpha
// (gw*gh bytes,0..255)」)。本檔照實把每個 glyph 存成 []byte(長度 w*h,
// row-major,每 byte 一個像素的 alpha 值),不做位元封裝。
//
// 本檔為純資料層:不 import Ebiten,只回傳 bitmask ([]byte),render 層
// 之後自行轉成貼圖。
package kbdata

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"
)

// cjkMagic 對齊 cjkfont.c:18 的 CJK_MAGIC { 'O','K','B','C','J','K','\0','\1' }。
var cjkMagic = [8]byte{'O', 'K', 'B', 'C', 'J', 'K', 0x00, 0x01}

// CJKGlyph 是單一字的點陣資料。Alpha 是 row-major 的灰階遮罩,
// 長度恆為 Width*Height,每個 byte 是該像素的不透明度(0..255)。
type CJKGlyph struct {
	CodePoint rune
	Width     int
	Height    int
	Alpha     []byte // len == Width*Height,第 row*Width+col 個 byte 對應該像素
}

// CJKAtlas 是載入後的整份 cjk24.bin,依 CodePoint 遞增排序以便二分查找
// (對齊 C 版 cjkfont_get 的二分搜尋,見 cjkfont.c:102-121)。
type CJKAtlas struct {
	Width  int // 單字寬(像素),cjk24.bin 目前恆為 24
	Height int // 單字高(像素),cjk24.bin 目前恆為 24
	glyphs []CJKGlyph
	index  map[rune]int // codepoint -> glyphs 的 index,查詢用(比 C 版二分搜尋直觀,結果相同)
}

// LoadCJKAtlas 從路徑讀取 cjk24.bin 並解析成 *CJKAtlas。
func LoadCJKAtlas(path string) (*CJKAtlas, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("kbdata: LoadCJKAtlas %s: %w", path, err)
	}
	return ParseCJKAtlas(data)
}

// ParseCJKAtlas 解析 cjk24.bin 的原始位元組內容。
func ParseCJKAtlas(data []byte) (*CJKAtlas, error) {
	const headerLen = 16 // magic[8] + u16 w + u16 h + u32 count
	if len(data) < headerLen {
		return nil, fmt.Errorf("kbdata: cjk atlas 太短 (%d bytes < %d)", len(data), headerLen)
	}
	if !bytesEqual(data[0:8], cjkMagic[:]) {
		return nil, fmt.Errorf("kbdata: cjk atlas magic 不符 (要 %q)", cjkMagic)
	}
	w := int(binary.LittleEndian.Uint16(data[8:10]))
	h := int(binary.LittleEndian.Uint16(data[10:12]))
	count := int(binary.LittleEndian.Uint32(data[12:16]))

	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("kbdata: cjk atlas 尺寸不合法 %dx%d", w, h)
	}
	recSize := 4 + w*h // 4-byte codepoint + w*h alpha bytes,對齊 cjkfont.c:27 recsz
	need := headerLen + count*recSize
	if len(data) < need {
		return nil, fmt.Errorf("kbdata: cjk atlas 檔案被截斷 (需要 %d bytes,實際 %d bytes)", need, len(data))
	}

	atlas := &CJKAtlas{
		Width:  w,
		Height: h,
		glyphs: make([]CJKGlyph, count),
		index:  make(map[rune]int, count),
	}

	off := headerLen
	for i := 0; i < count; i++ {
		cp := rune(binary.LittleEndian.Uint32(data[off : off+4]))
		alphaStart := off + 4
		alphaEnd := alphaStart + w*h
		alpha := make([]byte, w*h)
		copy(alpha, data[alphaStart:alphaEnd])

		atlas.glyphs[i] = CJKGlyph{
			CodePoint: cp,
			Width:     w,
			Height:    h,
			Alpha:     alpha,
		}
		atlas.index[cp] = i
		off = alphaEnd
	}

	// C 版假設 codepoint 已遞增排序(靠二分搜尋);這裡不假設檔案一定有序,
	// 用 map 直查即可拿到同樣結果,順便對「萬一沒排序」更寬容。
	if !sort.SliceIsSorted(atlas.glyphs, func(i, j int) bool {
		return atlas.glyphs[i].CodePoint < atlas.glyphs[j].CodePoint
	}) {
		// 保留原始順序不動(map 索引已經正確),只是這代表檔案跟 C 版註解描述的
		// 「codepoint 升序」假設不一致 ── 目前實際樣本 (cjk24.bin) 是有序的。
	}

	return atlas, nil
}

// Glyph 依 rune 查詢字型資料;找不到回 (nil, false)。
func (a *CJKAtlas) Glyph(r rune) (*CJKGlyph, bool) {
	if a == nil {
		return nil, false
	}
	idx, ok := a.index[r]
	if !ok {
		return nil, false
	}
	return &a.glyphs[idx], true
}

// Has 回報 atlas 是否收錄某個 rune。
func (a *CJKAtlas) Has(r rune) bool {
	_, ok := a.Glyph(r)
	return ok
}

// Count 回傳 atlas 收錄的字數。
func (a *CJKAtlas) Count() int {
	if a == nil {
		return 0
	}
	return len(a.glyphs)
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
