// chrome.go -- 世界地圖與城鎮共用的畫面「外殼」:右側 sidebar、置中頂部狀態列、
// 底部選單框。這三者在 C 版(game.c draw_sidebar / ui.c KB_TopBox / KB_BottomFrame)
// 本來就是全域函式,不屬於任何單一畫面,故在 Go 版也抽成 package 級函式,讓
// WorldMapScreen 與 TownScreen 呼叫同一份實作,保證兩畫面的側欄/框視覺一致。
package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// townFrameDivisor 決定「畫面動畫幀(0-3 循環)」相對於 Draw() 呼叫次數的節奏。
// 對齊 C troop_delay 節奏(game.c:2845-2847):計時器 hotspot 每 90(單位)觸發一次,
// troop_delay > 1(每 2 次觸發,約 180ms)才前進一幀。本引擎裡 Screen.Update(a) 只在
// 收到實際輸入動作時才呼叫(見 internal/app/game.go Game.Update),並非每畫面幀都觸發;
// 真正保證每一幀(桌面 ~60fps)都會執行的是 Draw(),所以動畫計數器改在 Draw() 遞增
// ——這是依本引擎架構做的必要修正,不是照抄「Update 每次 +1」。60fps 下每 10 次
// Draw()(~167ms)前進一幀,節奏對齊 C 版約 180ms 換一幀。
const townFrameDivisor = 10

// drawSidebar 畫右側資訊欄(對齊 C draw_sidebar,game.c:1119-1210):sidebar.png
// 13 幀(48×34)由上而下堆疊在 map 右側(x=256,y=21 起):
//
//	幀8=合約框;contract != 0xFF 時疊上目標 villain 臉(VillainFace,frame 動畫)
//	幀9=攻城武器圖示(暫無欄位,恆顯示「未擁有」幀)
//	幀10/(4+frame)=魔法星:依 GameState.KnowsMagic,有魔法時走 4-7 四幀循環
//	幀11=拼圖地圖框(靜態)
//	幀12=錢袋 + coins.png 疊加金幣堆
//
// frame 對應 C draw_sidebar(game, tick) 的 tick 參數 —— 在 C 版 visit_town() 裡這個
// tick 與迎賓兵種的走路動畫幀是同一個變數(0-3 循環,見 game.c:2845-2847
// troop_delay 節奏),故呼叫端應把「當下畫面用來驅動動畫的 0-3 幀」原樣傳進來,
// 不是隨時間無界遞增的計時器。
// 素材未載入時退回舊版深色色塊,維持無資產環境仍可運作。
func drawSidebar(dst *ebiten.Image, gs *gamestate.GameState, frame int) {
	sx := mapX + perim*mapTileW // 256
	if sidebarSprite == nil {
		vector.DrawFilledRect(dst, float32(sx), 0, float32(320-sx), 200, color.RGBA{40, 30, 22, 255}, false)
		return
	}
	knowsMagic := gs != nil && gs.KnowsMagic

	y := mapY // 21,對齊地圖頂
	sidebarSprite.DrawFrame(dst, 8, sx, y) // 幀8=契約框
	// 契約臉:對齊 C draw_sidebar(game.c:1147)——contract != 0xFF 時,把目標
	// villain 的臉(GR_VILLAIN[contract],frame 動畫)疊在契約框上(不去背,整格覆蓋)。
	if gs != nil && gs.Contract != 0xFF {
		if face := VillainFace(int(gs.Contract)); face != nil {
			face.DrawFrame(dst, frame, sx, y)
		}
	}
	y += mapTileH
	sidebarSprite.DrawFrame(dst, 9, sx, y)
	y += mapTileH
	if knowsMagic {
		sidebarSprite.DrawFrame(dst, 4+frame, sx, y)
	} else {
		sidebarSprite.DrawFrame(dst, 10, sx, y)
	}
	y += mapTileH
	sidebarSprite.DrawFrame(dst, 11, sx, y)
	y += mapTileH
	sidebarSprite.DrawFrame(dst, 12, sx, y)
	if gs != nil {
		drawCoins(dst, sx, y, gs.Gold)
	}
}

// drawTopBox 畫置中的頂部狀態列(對齊 C KB_TopBox(MSG_CENTERED, str)):清成
// colorStatus 底,文字置中於 local.status 矩形(statusX/Y/W/H,定義於 worldmap.go)。
// C 版置中算法是「(可容納字元格數 - 字串長度)/2」,這裡用 CJKCell 當一格寬度
// (ASCII/CJK 同格,見 render.CJKCell 註解),對應 fs.w。
//
// 城鎮畫面固定用這個顯示「按 'ESC' 離開」。世界地圖的狀態列文字不同(含動態剩餘
// 天數、且不置中),維持 WorldMapScreen 自己的 drawStatusBar,不套用本函式。
func drawTopBox(dst *ebiten.Image, assets *kbdata.Assets, text string) {
	vector.DrawFilledRect(dst, statusX, statusY, statusW, statusH, colorStatus, false)
	if assets == nil || assets.Font == nil {
		return
	}
	runes := []rune(text)
	maxCells := statusW / render.CJKCell
	pad := 0
	if len(runes) < maxCells {
		pad = (maxCells - len(runes)) / 2
	}
	render.DrawText(dst, assets.Font, text, statusX+pad*render.CJKCell, statusY, color.White)
}

// 底部選單框幾何(320×200 邏輯座標,zoom=0),對齊 C KB_BottomFrame()(ui.c:642):
//
//	border = 30x8 字元框 + fs.h/2 額外高度(RECT_Text(8,30) 後 border.h += fs.h/2)
//	border.x = left_frame.w(16),border.y = 200 - bottom_frame.h(8) - border.h
//	text 起點 = border 位置 + (fs.w, fs.h)(RECT_Pos 再位移一格)
//
// 數值:border=(16,124) 240×68;text 起點=(24,132)。
const (
	bottomBorderX = mapX                                // 16 = FRAME_LEFT.w
	bottomBorderW = 30 * render.CJKCell                 // 240 = 30 cols
	bottomBorderH = 8*render.CJKCell + render.CJKCell/2 // 68 = 8 rows + fs.h/2
	bottomBorderY = 200 - 8 - bottomBorderH             // 124(8 = FRAME_BOTTOM.h)
	bottomTextX   = bottomBorderX + render.CJKCell      // 24
	bottomTextY   = bottomBorderY + render.CJKCell      // 132
	// 表頭第一行比 text 起點再往上 fs.h/4+fs.h/8,對齊 C
	// KB_iloc(text->x, text->y - fs->h/4 - fs->h/8) 的「往上幾像素」微調。
	bottomHeaderY = bottomTextY - render.CJKCell/4 - render.CJKCell/8 // 129
)

// drawBottomFrame 畫底部選單框(背景+邊框,顏色沿用世界地圖同一套 colorStatus/
// colorBorder,對齊 C data/free/colors.ini [generic] background/frame1)+ 依序印
// lines,每行行高 CJKCell(對齊 C KB_iloc/KB_iprint 逐行前進 —— visit_town() 的
// header 兩行 + 五個 KB_imenu/KB_iprint 選項行,彼此都在同一 line_height=fs.h 的
// 文字流內,無額外行距)。lines 可以少於 7 行(如情報/尚未實作提示),多出的框內
// 空間留白。
func drawBottomFrame(dst *ebiten.Image, assets *kbdata.Assets, lines []string) {
	vector.DrawFilledRect(dst, bottomBorderX, bottomBorderY, bottomBorderW, bottomBorderH, colorStatus, false)
	vector.StrokeRect(dst, bottomBorderX, bottomBorderY, bottomBorderW, bottomBorderH, 1, colorBorder, false)
	if assets == nil || assets.Font == nil {
		return
	}
	for i, line := range lines {
		render.DrawText(dst, assets.Font, line, bottomTextX, bottomHeaderY+i*render.CJKCell, color.White)
	}
}

// drawLocationBg 畫地點背景 + 迎賓兵種,對齊 C draw_location(loc_id, troop_id, frame)
// (game.c:1086):背景整張貼在 (mapX,mapY),兵種 sprite 貼在背景左緣往右一個幀寬、
// 底部對齊背景底,C 版 GR_TROOP 用 flip=1 故用 DrawFrameFlipped。與 recruit.go/town.go
// 的 drawLocation 同幾何,抽成共用 helper 供城堡系列畫面(家鄉招兵/謁見)重用。
// troopID<0 時只畫背景(無迎賓兵種)。
func drawLocationBg(dst *ebiten.Image, bg *ebiten.Image, troopID, frame int) {
	if bg == nil {
		return
	}
	render.DrawTile(dst, bg, mapX, mapY)
	if troopID < 0 {
		return
	}
	sp := troopSpriteFor(troopID)
	if sp == nil {
		return
	}
	bgH := bg.Bounds().Dy()
	tx := mapX + sp.FrameW*1
	ty := mapY + bgH - sp.FrameH
	sp.DrawFrameFlipped(dst, frame, tx, ty)
}
