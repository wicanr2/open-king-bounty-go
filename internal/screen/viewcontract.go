// viewcontract.go -- 懸賞契約檢視畫面,對齊 C view_contract()(game.c:1641-1735)。
package screen

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// 版面幾何(320×200 邏輯座標,zoom=0),逐字對齊 C view_contract() 開頭(game.c:1649-1662):
//
//	RECT_Text(&border, 16, 36) — 巨集是 RECT_Text(RECT, ROWS, COLS)(ui.h:148),
//	即 border.h = 16*fs.h = 128,border.w = 36*fs.w = 288。
//	RECT_Center(&border, screen) — border.x = (320-288)/2 = 16,border.y = (200-128)/2 = 36。
//	border.y += fs->h/2 → 36+4 = 40。
//	hdst(臉/空框 blit 目標,同時也是 KB_iloc 文字起點)= { border.x+fs.w, border.y+fs.h,
//	tile.w, tile.h } = { 24, 48, 48, 34 } —— tile 是 GR_PURSE(SDL_CreatePALSurface(
//	TILE_W,TILE_H) 純尺寸佔位,從未真的 blit 到畫面上,只借它的 48×34 尺寸),故本移植
//	直接寫死 48/34 常數,不載入/引用 GR_PURSE 素材本身。
const (
	vctBorderW = 36 * render.CJKCell // 288
	vctBorderH = 16 * render.CJKCell // 128
	vctBorderX = (320 - vctBorderW) / 2
	vctBorderY = (200-vctBorderH)/2 + render.CJKCell/2 // 40

	vctFaceW = 48 // TILE_W(GR_PURSE 尺寸,face/GR_VILLAIN 每幀同尺寸)
	vctFaceH = 34 // TILE_H

	vctFaceX = vctBorderX + render.CJKCell // 24:同時是文字左緣(KB_iloc 同一個 hdst 起點)
	vctFaceY = vctBorderY + render.CJKCell // 48:同時是文字頂緣
)

// ViewContractScreen 是懸賞契約檢視畫面,對齊 C view_contract()(game.c:1641)。
// 兩種狀態:
//   - gs.Contract == 0xFF(尚未接契約):置中訊息框(藍底,對齊 [generic]
//     background=#0000aa)+ sidebar.png 幀 8(空合約框,同 chrome.go drawSidebar
//     的「幀8=合約框」)+「你目前沒有懸賞契約！」提示,靜態單幀、無動畫。
//   - 有契約:框內左上角疊目標惡棍臉部(GR_VILLAIN,192×34 四幀,frame 0-3
//     循環播放,對齊 game.c 的 key==2 timer tick 換幀 —— 本移植改用既有的
//     Draw() 呼叫節奏自動驅動,同 viewarmy.go/townFrameDivisor 慣例)+
//     villains.ini 對應 <file>.txt 的多行懸賞描述(kbdata.Assets.VillainDescs),
//     文字內兩個 %s 佔位分別代入「最後現身洲名」與「城堡名」(find_villain_castle
//     找不到城堡時兩者皆顯示「未知」,對齊 C 版同一行為)。
//
// ⚠ C 版 view_contract 全程不呼叫 KB_TopBox/KB_BottomFrame(逐讀 game.c:1641-1735
// 確認,與 view_army/view_character 不同),故本移植也不畫頂列提示文字——這是
// 核對原始碼後的忠實結果,不是遺漏。
//
// ⚠ 洲名來源(rulebook 62 溯源,詳見 gamestate/continent.go 檔頭注記):不是
// bounty.c 硬編的 continent_names[],是 land.ini(refill_names() 執行期覆寫後的
// 實際顯示值)。城堡名同理讀 castles.ini(見 gamestate/castle.go 既有決策),不是
// bounty.c 的 castle_names[] 硬編英文預設值。
type ViewContractScreen struct {
	gs         *gamestate.GameState
	assets     *kbdata.Assets
	castles    []gamestate.Castle
	continents [kbdata.MaxContinents]string
	drawTick   int // 驅動惡棍臉部 0-3 動畫幀(節奏對齊 chrome.go townFrameDivisor)
}

// NewViewContractScreen 建立懸賞契約檢視畫面。
func NewViewContractScreen(gs *gamestate.GameState, a *kbdata.Assets) *ViewContractScreen {
	s := &ViewContractScreen{gs: gs, assets: a}
	if a != nil {
		s.castles = gamestate.LoadCastles(a)
		s.continents = gamestate.ContinentNames(a)
	}
	return s
}

func (s *ViewContractScreen) Update(a input.Action) Transition {
	if a.Kind == input.ActCancel {
		return Pop()
	}
	return Stay()
}

func (s *ViewContractScreen) Draw(dst *ebiten.Image) {
	// 外框:同 viewarmy.go/viewcharacter.go 慣例,C 版不清畫面(疊在前一畫面殘留
	// 上),本引擎每畫面全幅重繪,先填邊框色近似之。
	drawChromeFrame(dst)

	// 訊息框(藍底,對齊 SDL_TextRect 的 back=colors[COLOR_BACKGROUND]=[generic]
	// background=#0000aa=colorStatus)。C 版另外用 colors[COLOR_FRAME](#ffff55=
	// colorBorder,與本畫面外框同色)畫 ASCII 花邊——與外框同色會被外框「蓋掉」
	// 看不出輪廓,故本移植不重複畫這條在此處視覺上等於零貢獻的邊框線,只填底色
	// (同 vaBoxX/vcBoxX 等既有色塊慣例,不额外 StrokeRect)。
	vector.DrawFilledRect(dst, vctBorderX, vctBorderY, vctBorderW, vctBorderH, colorStatus, false)

	contract := byte(0xFF)
	if s.gs != nil {
		contract = s.gs.Contract
	}

	if contract == 0xFF {
		s.drawNoContract(dst)
		return
	}
	s.drawContract(dst, int(contract))
}

// drawNoContract 對齊 game.c:1664-1680:sidebar.png 幀 8(空合約框)+ 提示文字。
func (s *ViewContractScreen) drawNoContract(dst *ebiten.Image) {
	if sidebarSprite != nil {
		sidebarSprite.DrawFrame(dst, 8, vctFaceX, vctFaceY)
	}
	if s.assets == nil || s.assets.Font == nil {
		return
	}
	// 對齊 C 字面字串 "\n\n\n\n\n\n      你目前沒有懸賞契約！"(game.c:1669):
	// 6 個 \n 換行(KB_iloc 的 cursor_y 前進 6 列,列高 = fs.h = CJKCell,無額外
	// 行距)+ 6 個全形前導空格(每字仍佔 1 個 CJKCell 寬,同 CJK)。直接把整段
	// 字串拆行後逐行畫,前導空白讓 render.DrawText 自然用「跳過無字圖但仍佔位」
	// 的方式對齊,不必手動算 x 偏移。
	const msg = "\n\n\n\n\n\n      你目前沒有懸賞契約！"
	for i, line := range strings.Split(msg, "\n") {
		render.DrawText(dst, s.assets.Font, line, vctFaceX, vctFaceY+i*render.CJKCell, color.White)
	}
}

// drawContract 對齊 game.c:1682-1729:惡棍臉(GR_VILLAIN,frame 0-3 循環)+
// STRL_VDESCS 描述文字(%s %s 代入洲名/城堡名)。
func (s *ViewContractScreen) drawContract(dst *ebiten.Image, villainID int) {
	frame := (s.drawTick / townFrameDivisor) % 4
	s.drawTick++

	if face := VillainFace(villainID); face != nil {
		face.DrawFrame(dst, frame, vctFaceX, vctFaceY)
	}

	if s.assets == nil || s.assets.Font == nil {
		return
	}
	if villainID < 0 || villainID >= kbdata.MaxVillains {
		return
	}
	text := s.assets.VillainDescs[villainID]
	if text == "" {
		return
	}

	continentName := "未知"
	castleName := "未知"
	if castleID := gamestate.FindVillainCastle(s.gs, villainID); castleID != -1 {
		if castleID >= 0 && castleID < len(s.castles) {
			castleName = s.castles[castleID].Name
		}
		cont := -1
		if castleID >= 0 && castleID < len(s.castles) {
			cont = s.castles[castleID].Continent
		}
		if cont >= 0 && cont < kbdata.MaxContinents && s.continents[cont] != "" {
			continentName = s.continents[cont]
		}
	}

	formatted := fmt.Sprintf(text, continentName, castleName)
	for i, line := range strings.Split(formatted, "\n") {
		render.DrawText(dst, s.assets.Font, line, vctFaceX, vctFaceY+i*render.CJKCell, color.White)
	}
}

// Keymap:純檢視,任意鍵離開(同 viewarmy.go/viewcharacter.go,C 版對應
// KB_Pause()/KB_event 收到任何非 timer-tick 按鍵即 done=1)。
func (s *ViewContractScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: false,
		Cancel:     "離開",
	}
}
