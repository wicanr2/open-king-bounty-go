package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// PuzzleScreen 對齊 C view_puzzle(game.c:1392):5×5 拼圖,以權杖所在為中心。每格由一件
// 神器或一名惡棍遮住,找到神器/捕獲惡棍即掀開,露出權杖周邊的實際地圖 —— 全部掀開後
// 即可看出權杖埋在哪種地形、對照小地圖找出確切位置。由世界地圖按 'p' 開啟。
//
// 簡化:C 版未掀開的格顯示惡棍臉/神器圖示;本移植未掀開格顯示灰底(素材疊圖屬後續),
// 掀開格用 tileColor 顯示該地圖 tile。
type PuzzleScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
}

// NewPuzzleScreen 建立拼圖畫面。
func NewPuzzleScreen(gs *gamestate.GameState, a *kbdata.Assets) *PuzzleScreen {
	return &PuzzleScreen{gs: gs, assets: a}
}

func (s *PuzzleScreen) Update(a input.Action) Transition {
	if !a.IsNone() {
		return Pop()
	}
	return Stay()
}

func (s *PuzzleScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	drawTopBox(dst, s.assets, "權杖拼圖(按任意鍵離開)")
	drawSidebar(dst, s.gs, 0)
	if s.gs == nil {
		return
	}
	// 5×5 拼圖鋪在地圖視窗區,每格用整格大小(perim=5 剛好對上 5×5 拼圖)。
	for y := 0; y < gamestate.PuzzleH; y++ {
		for x := 0; x < gamestate.PuzzleW; x++ {
			px := mapX + x*mapTileW
			// Y 翻轉:拼圖 y=0 是南(低 y),畫在下方,對齊世界地圖朝向。
			py := mapY + (gamestate.PuzzleH-1-y)*mapTileH
			if s.gs.PuzzleOpened(x, y) {
				vector.DrawFilledRect(dst, float32(px), float32(py), mapTileW, mapTileH, tileColor(s.gs.PuzzleTile(x, y)), false)
			} else {
				// 未掀開:灰底(對齊 C 以惡棍臉/神器圖遮蓋,素材疊圖待後續)。
				vector.DrawFilledRect(dst, float32(px), float32(py), mapTileW, mapTileH, color.RGBA{70, 70, 70, 255}, false)
			}
			vector.StrokeRect(dst, float32(px), float32(py), mapTileW, mapTileH, 1, color.RGBA{30, 30, 30, 255}, false)
		}
	}
}

func (s *PuzzleScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: false, Cancel: "離開"}
}
