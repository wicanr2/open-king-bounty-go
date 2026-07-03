package screen

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/combat"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
	"github.com/wicanr2/open-king-bounty-go/internal/save"
)

// demoRecruitTroopID 是踩到棲地(TileDwelling1..4)時暫代的招募兵種(index 0=農夫)。
// TODO: 真正的「棲地住哪種兵」需要世界狀態 dwelling_troop(land.ini 尚未解析),
// 屆時應依 tile 種類 / 座標查表決定 troopID,而不是固定寫死。
const demoRecruitTroopID = 0

// placeholderFoe 是暫代的敵方部隊(野狼×5),之後改取自世界狀態 foe_troops。
func placeholderFoe() [combat.MaxUnits]gamestate.Squad {
	var f [combat.MaxUnits]gamestate.Squad
	f[0] = gamestate.Squad{TroopID: 3, Count: 5} // 野狼
	for i := 1; i < combat.MaxUnits; i++ {
		f[i] = gamestate.Squad{TroopID: 255}
	}
	return f
}

// 版面對齊 C 版 openkb(data/free/ui.ini [map]/[tile] + game.c draw_map):
//   地圖 viewport 16,21 寬240高170;每 tile 48×34 → 5×5 格;玩家永遠在正中格。
//   Y 軸翻轉:遊戲 y 向上為北,螢幕上方 = 高 y(C: pos.y=(perim-1-j)*h+mapY)。
const (
	mapTileW = 48
	mapTileH = 34
	mapX     = 16
	mapY     = 21
	perim    = 5 // 240/48 = 170/34 = 5(5×5 viewport)
	radii    = 2 // (perim-1)/2:玩家置中,每邊 2 格
)

// WorldMapScreen 是世界地圖:顯示 land.org 的 tile(暫以色塊分類)、玩家可走動、踩城鎮進城。
type WorldMapScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	cont   int
	px, py int    // 玩家在該洲的 tile 座標
	msg    string // 存讀檔提示(「已存檔/已讀檔/無存檔」等),不計時,下次動作前持續顯示
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
	// 存讀檔:'s' 存進 slot 0,'l' 讀 slot 0。無關乎世界地圖資產是否載入,故放在
	// nil 檢查之前;結果(成功/失敗)寫進 s.msg,由 Draw 顯示。
	if a.Kind == input.ActLetter {
		switch a.Rune {
		case 's':
			s.handleSave()
			return Stay()
		case 'l':
			return s.handleLoad()
		}
	}

	if s.assets == nil || s.assets.World == nil {
		if a.Kind == input.ActCancel {
			return Replace(NewTitleScreen(s.assets))
		}
		return Stay()
	}
	nx, ny := s.px, s.py
	switch a.Kind {
	case input.ActUp:
		ny++ // Y 翻轉:向上 = 北 = 高 y(對齊 C 版螢幕座標)
	case input.ActDown:
		ny--
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
	// 踩到棲地 → 招兵(疊上 RecruitScreen,離開後回地圖)
	if tile >= kbdata.TileDwelling1 && tile <= kbdata.TileDwelling4 {
		// TODO: 棲地實際教哪個兵種需世界狀態 dwelling_troop(land.ini 尚未解析);
		// 暫寫死 demoRecruitTroopID 示範流程。
		return Push(NewRecruitScreen(s.gs, s.assets, demoRecruitTroopID))
	}
	return Stay()
}

// handleSave 把目前 GameState 存進 slot 0,結果寫進 s.msg。
func (s *WorldMapScreen) handleSave() {
	if s.gs == nil {
		s.msg = "無角色資料,無法存檔"
		return
	}
	if err := save.SaveGame(s.gs, 0); err != nil {
		s.msg = "存檔失敗: " + err.Error()
		return
	}
	s.msg = "已存檔"
}

// handleLoad 讀回 slot 0;成功則以載入的 GameState 換掉本畫面(Replace),
// 失敗(含存檔不存在)只更新提示、留在原地圖。
func (s *WorldMapScreen) handleLoad() Transition {
	loaded, err := save.LoadGame(0)
	if err != nil {
		if os.IsNotExist(err) {
			s.msg = "無存檔"
		} else {
			s.msg = "讀檔失敗: " + err.Error()
		}
		return Stay()
	}
	return Replace(NewWorldMapScreen(loaded, s.assets))
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

func (s *WorldMapScreen) Draw(dst *ebiten.Image) {
	if s.assets == nil || s.assets.World == nil {
		ebitenutil.DebugPrint(dst, "world map: land.org 未載入")
		return
	}
	// 5×5 viewport,玩家置中,Y 翻轉(對齊 C 版 draw_map:pos.y=(perim-1-j)*h+mapY)。
	borderX := s.px - radii
	borderY := s.py - radii
	for j := 0; j < perim; j++ {
		for i := 0; i < perim; i++ {
			mx, my := borderX+i, borderY+j
			sx := mapX + i*mapTileW
			sy := mapY + (perim-1-j)*mapTileH
			tile := byte(kbdata.TileDeepWater) // 界外 = 深水(對齊 C 邊界填色)
			if mx >= 0 && my >= 0 && mx < kbdata.LevelW && my < kbdata.LevelH {
				tile = s.assets.World.Tile(s.cont, mx, my)
			}
			if worldTileset != nil {
				worldTileset.DrawTileAt(dst, tile, sx, sy)
			} else {
				vector.DrawFilledRect(dst, float32(sx), float32(sy), mapTileW, mapTileH, tileColor(tile), false)
			}
		}
	}
	// 主角 sprite 置中(cursor.png 幀 8 = KBMOUNT_RIDE 騎馬,對齊 C draw_player:
	// hsrc.x = tile_w*(mount+frame),mount RIDE=8;無 sprite 退回青框)。
	heroX := mapX + radii*mapTileW
	heroY := mapY + (perim-1-radii)*mapTileH
	if heroSprite != nil {
		heroSprite.DrawFrame(dst, 8, heroX, heroY)
	} else {
		vector.StrokeRect(dst, float32(heroX)+2, float32(heroY)+2, mapTileW-4, mapTileH-4, 2, color.RGBA{0, 230, 230, 255}, false)
	}
	s.drawSidebar(dst)
	if s.msg != "" && s.assets.Font != nil {
		render.DrawText(dst, s.assets.Font, s.msg, mapX, mapY+perim*mapTileH-2, color.RGBA{240, 220, 40, 255})
	}
}

// drawSidebar 畫右側資訊欄(map 右緣 256→320)。簡化版:深底 + 金 + 隊伍;
// 完整 villain 臉/coins/piece sprite 對齊 C draw_sidebar 後續補。
func (s *WorldMapScreen) drawSidebar(dst *ebiten.Image) {
	sx := mapX + perim*mapTileW // 256
	vector.DrawFilledRect(dst, float32(sx), 0, float32(320-sx), 200, color.RGBA{40, 30, 22, 255}, false)
	gs := s.gs
	if gs == nil || s.assets == nil {
		return
	}
	if s.assets.Font != nil {
		render.DrawText(dst, s.assets.Font, "金", sx+4, 4, color.RGBA{240, 220, 40, 255})
	}
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%d", gs.Gold), sx+4, 30)
	y := 46
	for _, sq := range gs.Army {
		if sq.Count == 0 || sq.TroopID == 255 {
			continue
		}
		nm := "?"
		if sq.TroopID < len(s.assets.Troops) {
			nm = string([]rune(s.assets.Troops[sq.TroopID].Name)[:1])
		}
		if s.assets.Font != nil {
			render.DrawText(dst, s.assets.Font, nm, sx+4, y, color.White)
		}
		ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%d", sq.Count), sx+30, y+6)
		y += 26
	}
}

func (s *WorldMapScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: true,
		Cancel:     "標題",
		Letters: []input.LetterItem{
			{Rune: 's', Label: "存檔"},
			{Rune: 'l', Label: "讀檔"},
		},
	}
}
