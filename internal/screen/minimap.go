package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// MinimapScreen 對齊 C view_minimap(game.c:1533):把目前所在洲的整張 64×64 地圖縮成
// 小色塊總覽,玩家位置以亮點標示。由世界地圖按 'm' 開啟(對齊 C KEY_ACT(VIEW_MAP))。
//
// 上色用既有 tileColor;迷霧對齊 C view_minimap(game.c:1613):取得該洲魔法寶珠(OrbFound)
// 則全顯示,否則只顯示已探索格(gs.Fog,玩家走過即揭示 5×5),未探索畫黑。
type MinimapScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
}

// NewMinimapScreen 建立小地圖畫面。
func NewMinimapScreen(gs *gamestate.GameState, a *kbdata.Assets) *MinimapScreen {
	return &MinimapScreen{gs: gs, assets: a}
}

func (s *MinimapScreen) Update(a input.Action) Transition {
	if !a.IsNone() {
		return Pop()
	}
	return Stay()
}

func (s *MinimapScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	drawTopBox(dst, s.assets, "本洲地圖(按任意鍵離開)")
	drawSidebar(dst, s.gs, 0)
	if s.gs == nil || s.gs.WorldMap == nil {
		return
	}
	// 把 64×64 洲圖鋪在地圖視窗區(mapX,mapY 起,寬 perim*mapTileW 高 perim*mapTileH)。
	// 每格 blockW×blockH,Y 翻轉(北=上,對齊世界地圖朝向)。
	const n = kbdata.LevelW // = LevelH = 64
	areaW := perim * mapTileW
	areaH := perim * mapTileH
	blockW := float32(areaW) / float32(n)
	blockH := float32(areaH) / float32(n)
	cont := s.gs.Continent
	// 對齊 C view_minimap(game.c:1613):取得該洲魔法寶珠則全顯示,否則只顯示已探索格(fog);
	// 未探索格畫黑(迷霧)。
	orb := s.gs.OrbFound[cont] != 0
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			px := float32(mapX) + float32(x)*blockW
			py := float32(mapY) + float32(n-1-y)*blockH // Y 翻轉
			if !orb && !s.gs.Fog[cont][y][x] {
				vector.DrawFilledRect(dst, px, py, blockW+1, blockH+1, color.RGBA{0, 0, 0, 255}, false)
				continue
			}
			tile := s.gs.WorldMap.Tile(cont, x, y)
			vector.DrawFilledRect(dst, px, py, blockW+1, blockH+1, tileColor(tile), false)
		}
	}
	// 玩家位置亮點(白)。
	if s.gs.Continent == cont {
		hx := float32(mapX) + float32(s.gs.X)*blockW
		hy := float32(mapY) + float32(n-1-s.gs.Y)*blockH
		vector.DrawFilledRect(dst, hx-1, hy-1, blockW+3, blockH+3, color.White, false)
	}
}

func (s *MinimapScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: false, Cancel: "離開"}
}
