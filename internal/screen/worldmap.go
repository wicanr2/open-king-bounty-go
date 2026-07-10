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

// findDwellingID 在 gs.DwellingCoords[cont] 找座標 (x,y) 對應的棲地 id,對齊 C
// game.c:3006-3010 visit_dwelling() 的線性掃描(不 break,取最後一個相符的 i——
// 座標理論上唯一,結果與有無 break 無差異,這裡忠實照抄不 break)。找不到回傳 -1
// (對齊 C `if (id == -1) return 0;`——呼叫端應維持原地不動)。
func findDwellingID(gs *gamestate.GameState, cont, x, y int) int {
	id := -1
	for i := 0; i < kbdata.MaxDwellings; i++ {
		if gs.DwellingCoords[cont][i][0] == x && gs.DwellingCoords[cont][i][1] == y {
			id = i
		}
	}
	return id
}

// findFoeID 在 gs.FoeCoords[cont] 找座標 (x,y) 對應的 foe id,對齊 C
// game.c:3711-3720 attack_foe() 的線性掃描(找到就 break)。找不到回傳 0
// (對齊 C 預設值 `int id = 0;`)。
func findFoeID(gs *gamestate.GameState, cont, x, y int) int {
	id := 0
	for i := 0; i < kbdata.MaxFoes; i++ {
		if gs.FoeCoords[cont][i][0] == x && gs.FoeCoords[cont][i][1] == y {
			id = i
			break
		}
	}
	return id
}

// foeSquadsFrom 把世界狀態的 foe_troops/foe_numbers(5 格,對齊 C game->foe_troops/
// foe_numbers[cont][id][5])轉成 combat 套件要的 [MaxUnits]Squad,數量為 0 的格
// 填 TroopID=255(空格哨兵值,對齊 gamestate.Squad 慣例;C 版本身沒有這個哨兵值,
// 是 combat.PrepareUnitsFoe/gamestate.Squad 這層轉換介面的既有慣例,見 setup.go)。
//
// 誠實標註(對齊 worldgen.go saltContinent 註解):id < FRIENDLY_FOES(=5,見
// kbdata.MaxFoes 相關常數與 C bounty.h FRIENDLY_FOES)的「友善」foe 在世界生成時
// 刻意未呼叫 repopulateFoe(C 原始碼此處的 populate_foe 呼叫本身被註解掉),故這 5 個
// 槽位在真正的 C 版遊戲中應觸發 game.c attack_foe() 的 accept_foe()「免費入隊」分支,
// 而非戰鬥;本次世界生成移植只還原資料佈置,尚未實作 accept_foe 分支,故目前踩到
// 這些槽位仍會走一般戰鬥流程,只是遇到的會是「全空」的敵軍(與先前 placeholderFoe
// 佔位相比,至少不再是寫死的野狼,但這個特例本身是延續 C 既有的未完成功能,
// 不是本次移植引入的新行為)。
func foeSquadsFrom(troops, numbers [5]int) [combat.MaxUnits]gamestate.Squad {
	var f [combat.MaxUnits]gamestate.Squad
	for i := 0; i < combat.MaxUnits && i < 5; i++ {
		if numbers[i] <= 0 {
			f[i] = gamestate.Squad{TroopID: 255}
			continue
		}
		f[i] = gamestate.Squad{TroopID: troops[i], Count: numbers[i]}
	}
	return f
}

// 版面對齊 C 版 openkb(data/free/ui.ini [map]/[tile] + game.c draw_map):
//
//	地圖 viewport 16,21 寬240高170;每 tile 48×34 → 5×5 格;玩家永遠在正中格。
//	Y 軸翻轉:遊戲 y 向上為北,螢幕上方 = 高 y(C: pos.y=(perim-1-j)*h+mapY)。
const (
	mapTileW = 48
	mapTileH = 34
	mapX     = 16
	mapY     = 21
	perim    = 5 // 240/48 = 170/34 = 5(5×5 viewport)
	radii    = 2 // (perim-1)/2:玩家置中,每邊 2 格

	// 玩家起點,對齊 C spawn_game(play.c:403-405):
	//   continent = special_coords[SP_HOME][0]
	//   x         = special_coords[SP_HOME][1]
	//   y         = special_coords[SP_HOME][2] - 2
	// ★ 注意:bounty.c 的 special_coords 是 DOS 地圖硬編值(11,7),但 free 模組會用
	//   data/free/land.ini [special0]「國王的居所」覆寫(game.c:299-301 KB_Resolve),
	//   free 真值 = (continent 0, x 15, y 10)。本移植是 free 版,故用 free 的家座標。
	//   px/py 與 C game->x/y 同一座標系(Y-flip 只在渲染,不影響資料座標),直接用。
	// TODO: 改成解析 land.ini [special0] 載入(現先硬編 free 值,免另建 parser)。
	homeContinent = 0  // land.ini [special0] continent
	homeStartX    = 15 // land.ini [special0] x
	homeStartY    = 8  // land.ini [special0] y(10) - 2
)

// 頂部狀態列(對齊 C update_ui_frames:local.status.x/y = left_frame.w/top_frame.h,
// local.status.w = screen.w - left_frame.w - right_frame.w,h = fs.h + zoom;
// data/free/ui.ini [top] h=8、[left]/[right] w=16,故 statusY=8、statusW=320-16-16=288)。
const (
	statusX = mapX // 16,與地圖左緣同(left_frame.w)
	statusY = 8
	statusW = 288
	statusH = 8
)

// 邊框與狀態列色(對齊 C data/free/colors.ini:frame1/frame2 = #ffff55、background = #0000aa)。
// C 版做法是整個畫面先清成 COLOR_FRAME,狀態列/地圖/側欄再蓋掉內部,四周沒蓋到的
// 就是視覺上的黃框——不必額外畫 4 條 FillRect。
var (
	colorBorder = color.RGBA{0xff, 0xff, 0x55, 0xff}
	colorStatus = color.RGBA{0x00, 0x00, 0xaa, 0xff}
)

// worldMapDaysLeft 是狀態列「剩餘天數」暫用的固定值(C: game->days_left)。
// GameState 尚未有 days_left 欄位,先寫死,待世界狀態補上後改真值。
const worldMapDaysLeft = 600

// WorldMapScreen 是世界地圖:顯示 land.org 的 tile(暫以色塊分類)、玩家可走動、踩城鎮進城。
type WorldMapScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	cont   int
	px, py int    // 玩家在該洲的 tile 座標
	msg    string // 存讀檔提示(「已存檔/已讀檔/無存檔」等),不計時,下次動作前持續顯示

	drawTick int // Draw() 每次呼叫遞增(對齊 chrome.go drawSidebar 的 frame 節奏注解)
}

// NewWorldMapScreen 建立地圖畫面,玩家起點掃描該洲第一個草地格。
func NewWorldMapScreen(gs *gamestate.GameState, a *kbdata.Assets) *WorldMapScreen {
	// 起點對齊 C spawn_game 的 home castle 座標(homeContinent/homeStartX/homeStartY),
	// 不再用 findStart 的「第一格草地」佔位。
	s := &WorldMapScreen{gs: gs, assets: a, cont: homeContinent, px: homeStartX, py: homeStartY}
	return s
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
		case 'v':
			// 檢視軍隊(對齊 C combat_options_menu / 世界地圖選單「檢視軍隊」項):
			// 疊上 ViewArmyScreen,ESC 離開後回地圖(Push,非 Replace)。
			return Push(NewViewArmyScreen(s.gs, s.assets))
		case 'c':
			// 檢視角色(對齊 C combat_options_menu / 世界地圖選單「檢視角色」項):
			// 疊上 ViewCharacterScreen,ESC 離開後回地圖(Push,非 Replace)。
			return Push(NewViewCharacterScreen(s.gs, s.assets))
		}
	}

	if s.assets == nil || s.gs == nil || s.gs.WorldMap == nil {
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
	tile := s.gs.WorldMap.Tile(s.cont, nx, ny)
	if !walkable(tile) {
		return Stay()
	}
	s.px, s.py = nx, ny
	// 踩到城鎮 → 進城(疊上 TownScreen,離開後回地圖)
	if tile == kbdata.TileTown {
		townID := 0
		if s.assets != nil {
			towns := gamestate.LoadTowns(s.assets)
			if id := gamestate.FindTown(towns, s.cont, s.px, s.py); id >= 0 {
				townID = id
			}
		}
		return Push(NewTownScreen(s.gs, s.assets, townID))
	}
	// 踩到敵人 → 進戰鬥(疊上 CombatScreen,結束後回地圖),敵方部隊取自世界狀態
	// gs.FoeTroops/FoeNumbers(對齊 C attack_foe() 依座標查 foe id 再讀 foe_troops/numbers)。
	if tile == kbdata.TileFoe {
		foeID := findFoeID(s.gs, s.cont, nx, ny)
		foe := foeSquadsFrom(s.gs.FoeTroops[s.cont][foeID], s.gs.FoeNumbers[s.cont][foeID])
		return Push(NewCombatScreen(s.gs, s.assets, foe, uint32(nx*kbdata.LevelH+ny+1)))
	}
	// 踩到棲地 → 招兵(疊上 RecruitScreen,離開後回地圖)。rtype 對齊 C
	// game.c:6915 visit_dwelling(game, m - TILE_DWELLING_1):TileDwelling1..4 依序
	// 對應 rtype 0=平原 1=森林 2=山丘 3=地下城。棲地 id 依座標查 gs.DwellingCoords
	// (對齊 C visit_dwelling 的線性掃描),RecruitScreen 內部再依 (cont,id) 讀
	// gs.DwellingTroop/DwellingPopulation(真實世界狀態,取代舊佔位)。
	if tile >= kbdata.TileDwelling1 && tile <= kbdata.TileDwelling4 {
		rtype := int(tile - kbdata.TileDwelling1)
		dwellingID := findDwellingID(s.gs, s.cont, nx, ny)
		if dwellingID == -1 {
			// 對齊 C `if (id == -1) return 0;`:理論上不該發生(tile 本身就是
			// dwelling),防呆保留在原地不進招兵畫面。
			return Stay()
		}
		return Push(NewRecruitScreen(s.gs, s.assets, s.cont, dwellingID, rtype))
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
	if s.assets == nil || s.gs == nil || s.gs.WorldMap == nil {
		ebitenutil.DebugPrint(dst, "world map: land.org 未載入")
		return
	}
	// 外框:整畫面先清成黃(對齊 C 版行為,見上方 colorBorder 註解);地圖/狀態列/
	// 側欄接著蓋掉內部,四周留白就是黃色邊框,不必個別畫 4 條 FillRect。
	dst.Fill(colorBorder)
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
				tile = s.gs.WorldMap.Tile(s.cont, mx, my)
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
	s.drawStatusBar(dst)
	frame := (s.drawTick / townFrameDivisor) % 4
	s.drawTick++
	drawSidebar(dst, s.gs, frame)
	if s.msg != "" && s.assets.Font != nil {
		render.DrawText(dst, s.assets.Font, s.msg, mapX, mapY+perim*mapTileH-2, color.RGBA{240, 220, 40, 255})
	}
}

// drawStatusBar 畫頂部狀態列(對齊 C draw_statusbar → KB_TopBox(0, " 選項 / 操作說明 /
// 剩餘天數:%d ", game->days_left)):藍底(colorStatus)+ 白字。cjk24.bin 的 atlas 其實
// 連 ASCII(數字/標點/空白)都收錄成 24×24 glyph(與 CJK 同格),故整串直接走
// render.DrawText 一次印完即可,不必混用 ebitenutil 的內建字(那套字級高與 8px
// 狀態列的 CJKCell 對不上,會整段往下溢出)。
func (s *WorldMapScreen) drawStatusBar(dst *ebiten.Image) {
	vector.DrawFilledRect(dst, statusX, statusY, statusW, statusH, colorStatus, false)
	if s.assets == nil || s.assets.Font == nil {
		return
	}
	text := fmt.Sprintf(" 選項 / 操作說明 / 剩餘天數:%d ", worldMapDaysLeft)
	render.DrawText(dst, s.assets.Font, text, statusX, statusY, color.White)
}

// drawCoins 依 C draw_sidebar 的金幣堆疊邏輯(game.c:1119+ cval 計算):三欄分別代表
// 金幣萬/千/百位數,每欄依數字堆疊對應張數的硬幣(coins.png 3 幀 16×5),由底往上疊。
// C 版原始位移是 2*sys->zoom(桌面 320×200/zoom=0 時實際只疊在同一位置);這裡固定
// 抓 2px 位移,讓多枚硬幣在畫面上能看出堆疊,屬簡化近似。
func drawCoins(dst *ebiten.Image, x, y, gold int) {
	if coinsSprite == nil {
		return
	}
	cval := [3]int{gold / 10000, (gold % 10000) / 1000, (gold % 1000) / 100}
	coinW, coinH := coinsSprite.FrameW, coinsSprite.FrameH
	for col, cnt := range cval {
		dy := y + mapTileH - coinH
		for i := 0; i < cnt; i++ {
			coinsSprite.DrawFrame(dst, col, x+col*coinW, dy)
			dy -= 2
		}
	}
}

func (s *WorldMapScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: true,
		Cancel:     "標題",
		Letters: []input.LetterItem{
			{Rune: 's', Label: "存檔"},
			{Rune: 'l', Label: "讀檔"},
			{Rune: 'v', Label: "軍隊"},
			{Rune: 'c', Label: "角色"},
		},
	}
}
