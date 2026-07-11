// viewarmy.go -- 軍隊檢視畫面,對齊 C view_army()(game.c:1856-1952)。
package screen

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// 顏色對齊 C:view_army() 實際用的是 `KB_Resolve(COL_TEXT, CS_VIEWCHAR)`——
// CS_VIEWCHAR(=7)在 free-data.c 對映到 colors.ini 的 [character] 區段,不是看起來
// 對應的 [army] 區段(那個區段存在但 C 原始碼沒用到)。這是原始碼字面如此,移植
// 忠實沿用,不按名稱猜。[character] background=#000000、frame1=#ffffff。
// COL_DWELLING(colors.ini [dwellings] dwelling0-4)free 模組全為 #0000ff,
// 故任何有部隊的列,底色都是這個藍,與棲地種類無關(不必查 Troop.Home)。
var (
	viewArmyEmptyBg = color.RGBA{0x00, 0x00, 0x00, 0xff} // orig_bg:無部隊格底色(黑,對齊 [character] background)
	viewArmySepLine = color.RGBA{0xff, 0xff, 0xff, 0xff} // colors[COLOR_FRAME]:列間分隔線(白,對齊 [character] frame1)
	viewArmyTroopBg = color.RGBA{0x00, 0x00, 0xff, 0xff} // COL_DWELLING 全值(藍)
)

// 版面幾何(320×200 邏輯座標),逐字對齊 C view_army() 開頭與迴圈本體:
//
//	pos.x = left_frame.w = mapX(16);pos.y = top_frame.h+bar_frame.h+fs.h+zoom = mapY(21)
//	pos.w = screen.w - left_frame.w - right_frame.w = perim*mapTileW(240,即世界地圖
//	viewport 寬,不含右側 sidebar——C 版 view_army 完全不呼叫 draw_sidebar,右側
//	維持底色,故本移植也不畫 sidebar)。每列高 = tile.h(mapTileH,34),5 列共 170px。
const (
	vaPosW     = perim * mapTileW                    // 240
	vaBoxX     = mapX + mapTileW                     // 64:文字框左緣(tile 之後)
	vaBoxW     = vaPosW - mapTileW                   // 192
	vaSepH     = render.CJKCell / 8                  // fs.h/8 = 1:分隔線厚度
	vaBoxH     = mapTileH - vaSepH                   // 33
	vaTextPadY = render.CJKCell/2 - render.CJKCell/8 // fs.h/2 - fs.h/8 = 3:文字起點相對 tbox.y 的位移
	vaLineH    = render.CJKCell + 4                  // ilh = fs.h+4 = 12:KB_ilh 設定的行高
	vaRightX   = render.CJKCell * 16                 // fs.w*16 = 128:右欄相對 tbox.x 的位移
)

// ViewArmyScreen 是軍隊檢視畫面,對齊 C view_army()(game.c:1856):5 列部隊(每列
// = gs.Army[i],Count>0 才畫內容),左側 GR_TILE 幀 0 底 + 兵種 sprite(0-3 幀走路
// 動畫)、左欄文字(數量+名 / SL+MV / 士氣或失控)、右欄文字(生命/傷害/造價),
// 列底色 = COL_DWELLING(free 全藍),列間白線分隔(第 5 列不畫,對齊 C `i < 4`)。
// 純檢視畫面,唯一互動是 ESC 離開,不套 D-pad/字母選單(Directions:false)。
type ViewArmyScreen struct {
	gs       *gamestate.GameState
	assets   *kbdata.Assets
	drawTick int // 驅動左側 tile 底兵種 sprite 的 0-3 動畫幀(節奏對齊 chrome.go townFrameDivisor)
}

// NewViewArmyScreen 建立軍隊檢視畫面。
func NewViewArmyScreen(gs *gamestate.GameState, a *kbdata.Assets) *ViewArmyScreen {
	return &ViewArmyScreen{gs: gs, assets: a}
}

func (s *ViewArmyScreen) Update(a input.Action) Transition {
	if a.Kind == input.ActCancel {
		return Pop()
	}
	return Stay()
}

func (s *ViewArmyScreen) Draw(dst *ebiten.Image) {
	// 動畫幀:C 版由方向鍵(key==2)手動切幀(0-3 循環);本移植是純檢視畫面不佔用
	// D-pad,改用 Draw() 呼叫節奏自動驅動(對齊 town.go 的 drawTick 手法)。
	frame := (s.drawTick / townFrameDivisor) % 4
	s.drawTick++

	// 外框:整畫面先清成黃邊框色(對齊其餘畫面 colorBorder 慣例)。C 版 view_army
	// 不清畫面、疊在既有畫面上(右側 sidebar 欄位仍是前一畫面殘留);本引擎每畫面
	// 各自全幅重繪(見 internal/app/game.go),無「前一畫面殘留」可疊,故用邊框色
	// 填滿右側/底部未畫到的區域,是本移植對此差異的忠實近似,非美術走樣。
	drawChromeFrame(dst)

	for i := 0; i < 5; i++ {
		rowY := mapY + i*mapTileH
		s.drawRow(dst, i, rowY, frame)
	}

	drawTopBox(dst, s.assets, "按 'ESC' 離開")
}

// drawRow 畫第 i 列(0-4),逐字對齊 C 迴圈本體(game.c:1902-1941):tile 底與
// tbox/tline 不論有無部隊都先畫;有部隊(Count>0)才疊兵種 sprite、列底色、文字。
func (s *ViewArmyScreen) drawRow(dst *ebiten.Image, i, rowY, frame int) {
	// 底 tile:C `tile = SDL_TakeSurface(GR_TILE, 0, 0)` 固定用 GR_TILE 索引 0,
	// 動畫幀不作用於這個 tile(frm 只用在下面的兵種 sprite 上)。
	if worldTileset != nil {
		worldTileset.DrawTileAt(dst, 0, mapX, rowY)
	}

	// tbox/tline 先清底(對齊 C:兩者在 `continue` 之前就先 FillRect 過一次)。
	sepColor := viewArmyEmptyBg
	if i < 4 {
		sepColor = viewArmySepLine
	}
	vector.DrawFilledRect(dst, float32(vaBoxX), float32(rowY), float32(vaBoxW), float32(vaBoxH), viewArmyEmptyBg, false)
	vector.DrawFilledRect(dst, float32(vaBoxX), float32(rowY+mapTileH-vaSepH), float32(vaBoxW), float32(vaSepH), sepColor, false)

	if s.gs == nil || s.assets == nil || i >= len(s.gs.Army) {
		return
	}
	squad := s.gs.Army[i]
	if squad.Count == 0 {
		return // 對齊 C `if (game->player_numbers[i] == 0) continue;`
	}
	troopID := squad.TroopID
	if troopID < 0 || troopID >= len(s.assets.Troops) {
		return
	}
	troop := s.assets.Troops[troopID]

	// 兵種 sprite:C `SDL_TakeSurface(GR_TROOP, troop_id, 1)`(flip=1)直接貼在跟
	// tile 完全相同的矩形上(48×34,兩者同尺寸,不像 town/recruit 那樣需要偏移)。
	if sp := troopSpriteFor(troopID); sp != nil {
		sp.DrawFrameFlipped(dst, frame, mapX, rowY)
	}

	// 列底色蓋成兵種棲地色(COL_DWELLING,free 全藍)。
	vector.DrawFilledRect(dst, float32(vaBoxX), float32(rowY), float32(vaBoxW), float32(vaBoxH), viewArmyTroopBg, false)

	if s.assets.Font == nil {
		return
	}
	textY := rowY + vaTextPadY
	render.DrawText(dst, s.assets.Font, fmt.Sprintf(" %-3d %s", squad.Count, troop.Name), vaBoxX, textY, color.White)
	render.DrawText(dst, s.assets.Font, fmt.Sprintf(" SL:%2d MV:%2d", troop.SkillLevel, troop.Move), vaBoxX, textY+vaLineH, color.White)

	moraleLine := fmt.Sprintf(" 士氣:%s", gamestate.MoraleNames[gamestate.TroopMorale(s.gs, s.assets, i)])
	if s.gs.ArmyLeadership(s.assets, troopID) <= 0 {
		moraleLine = " 失控"
	}
	render.DrawText(dst, s.assets.Font, moraleLine, vaBoxX, textY+2*vaLineH, color.White)

	rightX := vaBoxX + vaRightX
	render.DrawText(dst, s.assets.Font, fmt.Sprintf("生命:%d", troop.HP*squad.Count), rightX, textY, color.White)
	render.DrawText(dst, s.assets.Font, fmt.Sprintf("傷害:%d-%d", troop.MeleeMin*squad.Count, troop.MeleeMax*squad.Count), rightX, textY+vaLineH, color.White)
	render.DrawText(dst, s.assets.Font, fmt.Sprintf("造價:%d", troop.GoldCost/10*squad.Count), rightX, textY+2*vaLineH, color.White)
}

// Keymap:純檢視,無移動(Directions:false),不套 LetterRects(無字母選單)。
func (s *ViewArmyScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: false,
		Cancel:     "離開",
	}
}
