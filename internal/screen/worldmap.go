package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/combat"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// placeholderFoe 是暫代的敵方部隊(野狼×5),之後改取自世界狀態 foe_troops。
func placeholderFoe() [combat.MaxUnits]gamestate.Squad {
	var f [combat.MaxUnits]gamestate.Squad
	f[0] = gamestate.Squad{TroopID: 3, Count: 5} // 野狼
	for i := 1; i < combat.MaxUnits; i++ {
		f[i] = gamestate.Squad{TroopID: 255}
	}
	return f
}

const (
	tileSize = 10 // 每格像素(P2 暫用色塊,tileset 美術之後接)
	viewCols = 320 / tileSize
	viewRows = 200 / tileSize
)

// WorldMapScreen 是世界地圖:顯示 land.org 的 tile(暫以色塊分類)、玩家可走動、踩城鎮進城。
type WorldMapScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	cont   int
	px, py int // 玩家在該洲的 tile 座標
}

// NewWorldMapScreen 建立地圖畫面,玩家起點掃描該洲第一個草地格。
func NewWorldMapScreen(gs *gamestate.GameState, a *kbdata.Assets) *WorldMapScreen {
	s := &WorldMapScreen{gs: gs, assets: a, cont: 0, px: 32, py: 32}
	if a != nil && a.World != nil {
		s.px, s.py = findStart(a.World, 0)
	}
	return s
}

// findStart 掃描該洲找第一個草地格當起點(暫代,真起點在 land.ini 未解析)。
func findStart(m *kbdata.WorldMap, cont int) (int, int) {
	for y := 0; y < kbdata.LevelH; y++ {
		for x := 0; x < kbdata.LevelW; x++ {
			if kbdata.IsGrass(m.Tile(cont, x, y)) {
				return x, y
			}
		}
	}
	return 32, 32
}

// walkable 回報玩家能否踏上該 tile(水/深水/山阻擋;陸地與互動格可走)。
func walkable(tile byte) bool {
	return !kbdata.IsWater(tile) && !kbdata.IsDeepWater(tile) && !kbdata.IsRock(tile)
}

func (s *WorldMapScreen) Update(a input.Action) Transition {
	if s.assets == nil || s.assets.World == nil {
		if a.Kind == input.ActCancel {
			return Replace(NewTitleScreen(s.assets))
		}
		return Stay()
	}
	nx, ny := s.px, s.py
	switch a.Kind {
	case input.ActUp:
		ny--
	case input.ActDown:
		ny++
	case input.ActLeft:
		nx--
	case input.ActRight:
		nx++
	case input.ActCancel:
		return Replace(NewTitleScreen(s.assets))
	default:
		return Stay()
	}
	if nx < 0 || ny < 0 || nx >= kbdata.LevelW || ny >= kbdata.LevelH {
		return Stay()
	}
	tile := s.assets.World.Tile(s.cont, nx, ny)
	if !walkable(tile) {
		return Stay()
	}
	s.px, s.py = nx, ny
	// 踩到城鎮 → 進城(疊上 TownScreen,離開後回地圖)
	if tile == kbdata.TileTown {
		return Push(NewTownScreen(s.gs, s.assets))
	}
	// 踩到敵人 → 進戰鬥(疊上 CombatScreen,結束後回地圖)
	if tile == kbdata.TileFoe {
		// TODO: 敵方部隊應取自世界狀態 foe_troops;暫用佔位(野狼)+ 依座標決定 seed。
		foe := placeholderFoe()
		return Push(NewCombatScreen(s.gs, s.assets, foe, uint32(nx*kbdata.LevelH+ny+1)))
	}
	return Stay()
}

// tileColor 依 tile 分類回色塊顏色(P2 暫代美術)。
func tileColor(tile byte) color.Color {
	switch {
	case tile == kbdata.TileTown:
		return color.RGBA{240, 220, 40, 255} // 城鎮:黃
	case tile == kbdata.TileCastle || kbdata.IsCastle(tile):
		return color.RGBA{230, 230, 230, 255} // 城堡:白
	case tile == kbdata.TileChest:
		return color.RGBA{230, 140, 30, 255} // 寶箱:橙
	case tile == kbdata.TileFoe:
		return color.RGBA{210, 40, 40, 255} // 敵人:紅
	case tile >= kbdata.TileDwelling1 && tile <= kbdata.TileDwelling4:
		return color.RGBA{170, 60, 200, 255} // 棲地:紫
	case kbdata.IsBridge(tile):
		return color.RGBA{150, 100, 50, 255} // 橋:褐
	case kbdata.IsDeepWater(tile):
		return color.RGBA{10, 30, 110, 255}
	case kbdata.IsWater(tile):
		return color.RGBA{30, 70, 200, 255}
	case kbdata.IsTree(tile):
		return color.RGBA{0, 100, 0, 255}
	case kbdata.IsDesert(tile):
		return color.RGBA{205, 180, 130, 255}
	case kbdata.IsRock(tile):
		return color.RGBA{120, 120, 120, 255}
	case kbdata.IsGrass(tile):
		return color.RGBA{40, 150, 50, 255}
	default:
		return color.RGBA{90, 90, 90, 255}
	}
}

// camera 回傳夾制後的視窗左上角 tile 座標(以玩家為中心,但不超出地圖邊界)。
func (s *WorldMapScreen) camera() (int, int) {
	ox := s.px - viewCols/2
	oy := s.py - viewRows/2
	if ox < 0 {
		ox = 0
	}
	if oy < 0 {
		oy = 0
	}
	if ox > kbdata.LevelW-viewCols {
		ox = kbdata.LevelW - viewCols
	}
	if oy > kbdata.LevelH-viewRows {
		oy = kbdata.LevelH - viewRows
	}
	return ox, oy
}

func (s *WorldMapScreen) Draw(dst *ebiten.Image) {
	if s.assets == nil || s.assets.World == nil {
		ebitenutil.DebugPrint(dst, "world map: land.org 未載入")
		return
	}
	ox, oy := s.camera()
	for row := 0; row < viewRows; row++ {
		for col := 0; col < viewCols; col++ {
			tx, ty := ox+col, oy+row
			if tx < 0 || ty < 0 || tx >= kbdata.LevelW || ty >= kbdata.LevelH {
				continue
			}
			tile := s.assets.World.Tile(s.cont, tx, ty)
			vector.DrawFilledRect(dst, float32(col*tileSize), float32(row*tileSize),
				tileSize, tileSize, tileColor(tile), false)
		}
	}
	// 玩家(依相機夾制後的實際螢幕位置)
	vector.DrawFilledRect(dst, float32((s.px-ox)*tileSize), float32((s.py-oy)*tileSize),
		tileSize, tileSize, color.RGBA{0, 220, 220, 255}, false)

	ebitenutil.DebugPrintAt(dst, "arrows: move  ESC: title", 6, 186)
}

func (s *WorldMapScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: true, Cancel: "標題"}
}
