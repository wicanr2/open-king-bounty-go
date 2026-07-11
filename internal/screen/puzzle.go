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
	drawChromeFrame(dst)
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
				s.drawOpenedCell(dst, px, py, s.gs.PuzzleTile(x, y))
			} else {
				s.drawCoveredCell(dst, px, py, gamestate.PuzzleCellID(x, y))
			}
			vector.StrokeRect(dst, float32(px), float32(py), mapTileW, mapTileH, 1, color.RGBA{30, 30, 30, 255}, false)
		}
	}
}

// drawOpenedCell 畫掀開後露出的實際地圖 tile(有 tileset 用真美術,否則退回色塊)。
func (s *PuzzleScreen) drawOpenedCell(dst *ebiten.Image, px, py int, tile byte) {
	if worldTileset != nil {
		worldTileset.DrawTileAt(dst, tile, px, py)
		return
	}
	vector.DrawFilledRect(dst, float32(px), float32(py), mapTileW, mapTileH, tileColor(tile), false)
}

// drawCoveredCell 畫遮蓋物(對齊 C view_puzzle):惡棍格(id>=0)畫惡棍臉、神器格
// (id<0)畫神器圖示;素材缺時退回灰底。
func (s *PuzzleScreen) drawCoveredCell(dst *ebiten.Image, px, py, id int) {
	if id >= 0 {
		if face := VillainFace(id); face != nil {
			face.DrawFrame(dst, 0, px, py)
			return
		}
	} else {
		artID := -id - 1
		if viewItemSprite != nil && artID >= 0 {
			viewItemSprite.DrawFrame(dst, artID, px, py)
			return
		}
	}
	vector.DrawFilledRect(dst, float32(px), float32(py), mapTileW, mapTileH, color.RGBA{70, 70, 70, 255}, false)
}

func (s *PuzzleScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: false, Cancel: "離開"}
}
