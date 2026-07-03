package screen

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// RecruitScreen 是招兵畫面:顯示棲地兵種的單價/上限,用方向鍵夾住數量,
// Confirm 呼叫 gamestate.BuyTroop 買下,Cancel 直接離開(回地圖)。
//
// 對應 C 版 game.c 的招募對話框(數量步進 + 買/取消)。棲地實際教哪個兵種需要
// 世界狀態 dwelling_troop(land.ini 尚未解析),本畫面先由呼叫端(worldmap)傳入
// 固定 troopID 示範——見 NewRecruitScreen 呼叫處的 TODO。
type RecruitScreen struct {
	gs      *gamestate.GameState
	assets  *kbdata.Assets
	troopID int

	sel int    // 目前選定的招募數量(0..cap())
	msg string // 上次操作結果提示(「已招募」/ 錯誤訊息);不計時,換動作或再開畫面才變
}

// NewRecruitScreen 建立招兵畫面,初始數量為 0。
func NewRecruitScreen(gs *gamestate.GameState, a *kbdata.Assets, troopID int) *RecruitScreen {
	return &RecruitScreen{gs: gs, assets: a, troopID: troopID}
}

// cap 回傳目前數量上限:金錢買得起的隻數與領導力上限兩者取小。
func (s *RecruitScreen) cap() int {
	if s.gs == nil || s.assets == nil || s.troopID < 0 || s.troopID >= len(s.assets.Troops) {
		return 0
	}
	troop := s.assets.Troops[s.troopID]

	byGold := 0
	if troop.GoldCost > 0 {
		byGold = s.gs.Gold / troop.GoldCost
	}
	byLeadership := s.gs.ArmyMaxTroopCount(s.assets, s.troopID)

	if byGold < byLeadership {
		return byGold
	}
	return byLeadership
}

func (s *RecruitScreen) Update(a input.Action) Transition {
	max := s.cap()
	if s.sel > max {
		s.sel = max
	}
	switch a.Kind {
	case input.ActUp, input.ActRight:
		if s.sel < max {
			s.sel++
		}
	case input.ActDown, input.ActLeft:
		if s.sel > 0 {
			s.sel--
		}
	case input.ActConfirm:
		if s.gs == nil || s.assets == nil {
			return Stay()
		}
		if s.sel <= 0 {
			s.msg = "數量為 0,未招募"
			return Stay()
		}
		if err := s.gs.BuyTroop(s.assets, s.troopID, s.sel); err != nil {
			switch err {
			case gamestate.ErrNotEnoughGold:
				s.msg = "金錢不足"
			case gamestate.ErrNoSlot:
				s.msg = "隊伍無空格"
			default:
				s.msg = err.Error()
			}
			return Stay()
		}
		s.msg = "已招募"
		s.sel = 0
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

func (s *RecruitScreen) Draw(dst *ebiten.Image) {
	drawCJK := func(str string, x, y int, clr color.Color) {
		if s.assets != nil && s.assets.Font != nil {
			render.DrawText(dst, s.assets.Font, str, x, y, clr)
		}
	}

	drawCJK("招募", 8, 4, color.White)

	if s.assets == nil || s.gs == nil || s.troopID < 0 || s.troopID >= len(s.assets.Troops) {
		ebitenutil.DebugPrintAt(dst, "無可招募兵種資料", 8, 32)
		ebitenutil.DebugPrintAt(dst, "ESC: 離開", 8, 186)
		return
	}
	troop := s.assets.Troops[s.troopID]

	drawCJK(troop.Name, 8, 26, color.White)
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("單價 %d", troop.GoldCost), 8, 48)
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("持有金 %d", s.gs.Gold), 8, 64)
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("上限 %d", s.cap()), 8, 80)
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("數量: %d", s.sel), 8, 100)

	if s.msg != "" {
		drawCJK(s.msg, 8, 122, color.RGBA{240, 220, 40, 255})
	}

	ebitenutil.DebugPrintAt(dst, "left/right: 調數量  ENTER: 招募  ESC: 離開", 6, 186)
}

func (s *RecruitScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: true,
		Confirm:    "招募",
		Cancel:     "離開",
	}
}
