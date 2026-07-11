package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// CartoonScreen 對齊 C display_cartoon / draw_cartoon_frame(game.c:4420/4513):破關前的
// 過場動畫——6×5 草地上,第 4 欄一座橋自下而上長成、英雄拾級而上跨橋;其餘欄位排滿 25 種
// 兵種(走路幀動畫)當背景。C 版由 win_game() 呼叫,播完(或按鍵)進破關訊息;本移植保留
// 「任意鍵 → WinScreen」語意,動畫由 Draw 逐幀驅動(本引擎無每幀 Update tick,見 LogoScreen)。
// 缺 endpic tile 的主題(非 free)直接跳過動畫、黑底候按鍵。
type CartoonScreen struct {
	gs       *gamestate.GameState
	assets   *kbdata.Assets
	drawTick int
}

// NewCartoonScreen 建立破關過場動畫畫面。
func NewCartoonScreen(gs *gamestate.GameState, a *kbdata.Assets) *CartoonScreen {
	return &CartoonScreen{gs: gs, assets: a}
}

func (s *CartoonScreen) Update(a input.Action) Transition {
	if !a.IsNone() {
		return Replace(NewWinScreen(s.gs, s.assets)) // 任意鍵 → 破關訊息
	}
	return Stay()
}

// cartoonMap* = 動畫的 6×5 tile 網格在畫面上的原點與節奏。
const (
	cartoonCols          = 6
	cartoonRows          = 5
	cartoonOriginX       = mapX // 16
	cartoonOriginY       = statusY + statusH + 5
	cartoonFrameDivisor  = 6 // 每 6 個 Draw 前進一個動畫幀(bridge/hero 進度)
	cartoonWalkDivisor   = 8 // 兵種走路幀(0-3)的切換節奏
	cartoonBridgeCol     = 4 // 橋與英雄所在欄
	cartoonMaxBridgeRows = 5
	cartoonMaxHeroProg   = 4
)

func (s *CartoonScreen) Draw(dst *ebiten.Image) {
	dst.Fill(color.Black)

	grass := EndTile(0)
	bridge := EndTile(1)
	hero := EndTile(2)
	if grass == nil { // 非 free 主題缺 endpic:略過動畫,黑底候按鍵。
		return
	}

	frame := s.drawTick / cartoonFrameDivisor
	if frame > 10 {
		frame = 10 // 停在最後一幀(英雄已過橋),等玩家按鍵
	}
	walk := (s.drawTick / cartoonWalkDivisor) % 4
	s.drawTick++

	bridgeLen := frame
	if bridgeLen > cartoonMaxBridgeRows {
		bridgeLen = cartoonMaxBridgeRows
	}
	heroProg := frame - 5
	if heroProg > cartoonMaxHeroProg {
		heroProg = cartoonMaxHeroProg
	}

	at := func(col, row int) (int, int) {
		return cartoonOriginX + col*mapTileW, cartoonOriginY + row*mapTileH
	}

	// ① 草地鋪滿 6×5(對齊 C 先畫 grass 全格)。
	for row := 0; row < cartoonRows; row++ {
		for col := 0; col < cartoonCols; col++ {
			px, py := at(col, row)
			render.DrawTile(dst, grass, px, py)
		}
	}
	// ② 橋:第 4 欄自 row 4 往上長 bridgeLen 格(對齊 C bridge_len 迴圈)。
	for i := 0; i < bridgeLen; i++ {
		px, py := at(cartoonBridgeCol, 4-i)
		render.DrawTile(dst, bridge, px, py)
	}
	// ③ 英雄:第 4 欄,row = 4 - heroProg(對齊 C,heroProg>=0 才畫)。
	if heroProg >= 0 {
		px, py := at(cartoonBridgeCol, 4-heroProg)
		render.DrawTile(dst, hero, px, py)
	}
	// ④ 25 兵種鋪在其餘欄位(0,1,2,3,5 × 5 列;對齊 C draw troops 迴圈的欄位跳位),
	//    走路幀 = walk;x==5(最右欄)水平翻轉。
	col, row := 0, 0
	for i := 0; i < kbdata.MaxTroops && row < cartoonRows; i++ {
		if sp := troopSpriteFor(i); sp != nil {
			px, py := at(col, row)
			if col == 5 {
				sp.DrawFrameFlipped(dst, walk, px, py)
			} else {
				sp.DrawFrame(dst, walk, px, py)
			}
		}
		col++
		if col == cartoonBridgeCol {
			col = 5 // 跳過第 4 欄(橋/英雄)
		}
		if col >= cartoonCols {
			col = 0
			row++
		}
	}
}

func (s *CartoonScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: false, Confirm: "繼續"}
}
