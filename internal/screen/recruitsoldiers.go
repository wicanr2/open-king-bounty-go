package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// RecruitSoldiersScreen 對齊 C recruit_soldiers(game.c:2181):家鄉城堡 A) 招募士兵。
// 與棲地招募(RecruitScreen)不同——這裡一次列出全部 5 種城堡棲地兵種(A-E),無庫存
// 上限(max=9999,只受領導力與金錢限制)。兩階段:先選兵種(whom==0),再輸入數量。
//
// C 原始用 text_input 打數字;Go 版沒有數字鍵盤輸入,改用 D-pad 步進器(同 RecruitScreen
// 的觸控友善做法),上限 = army_max_troop_count(對齊 C max = army_max_troop_count)。
type RecruitSoldiersScreen struct {
	gs         *gamestate.GameState
	assets     *kbdata.Assets
	homeTroops []int  // 城堡棲地兵種 id(對齊 C home_troops[5])
	locID      int    // draw_location 背景 id(家鄉=0=cstl)
	bgTroop    int    // 背景迎賓兵種(C 傳入的 random_troop)
	whom       int    // 0=選兵種;1..N=正在為 homeTroops[whom-1] 選數量
	sel        int    // 數量步進器目前值
	msg        string // 購買結果提示(金幣不足/無空格)
}

// NewRecruitSoldiersScreen 建立家鄉招兵畫面。locID/bgTroop 供背景立繪(對齊 C
// draw_location(loc_id, troop_id, frame));家鄉 locID=0(cstl),bgTroop 由呼叫端傳
// 隨機城堡兵種。
func NewRecruitSoldiersScreen(gs *gamestate.GameState, a *kbdata.Assets, locID, bgTroop int) *RecruitSoldiersScreen {
	return &RecruitSoldiersScreen{
		gs:         gs,
		assets:     a,
		homeTroops: gamestate.HomeTroops(a),
		locID:      locID,
		bgTroop:    bgTroop,
	}
}

// max 回傳目前選定兵種的領導力上限(對齊 C army_max_troop_count),whom==0 時回 0。
func (s *RecruitSoldiersScreen) max() int {
	if s.whom <= 0 || s.whom > len(s.homeTroops) || s.gs == nil || s.assets == nil {
		return 0
	}
	return s.gs.ArmyMaxTroopCount(s.assets, s.homeTroops[s.whom-1])
}

func (s *RecruitSoldiersScreen) Update(a input.Action) Transition {
	if s.whom == 0 {
		// 選兵種階段:A-E 直選,或 D-pad 上下 + Confirm。
		switch a.Kind {
		case input.ActLetter:
			if a.Rune >= 'a' && a.Rune < rune('a'+len(s.homeTroops)) {
				s.pick(int(a.Rune-'a') + 1)
			}
		case input.ActCancel:
			return Pop()
		}
		return Stay()
	}

	// 數量階段:D-pad 調數量,Confirm 購買,Cancel 回選兵種。
	max := s.max()
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
		s.buy()
	case input.ActCancel:
		s.whom = 0 // 對齊 C:text_input 回 NULL(取消)→ whom=0 回選兵種
		s.sel = 0
		s.msg = ""
	}
	return Stay()
}

// pick 選定第 whom(1-based)個兵種進入數量階段。
func (s *RecruitSoldiersScreen) pick(whom int) {
	s.whom = whom
	s.sel = 0
	s.msg = ""
}

// buy 對齊 C recruit_soldiers 的購買分支:number 夾到 max 後 buy_troop,依結果提示。
func (s *RecruitSoldiersScreen) buy() {
	if s.gs == nil || s.assets == nil || s.whom <= 0 || s.whom > len(s.homeTroops) {
		return
	}
	if s.sel <= 0 {
		s.msg = "數量為 0,未招募"
		return
	}
	err := s.gs.BuyTroop(s.assets, s.homeTroops[s.whom-1], s.sel)
	switch err {
	case gamestate.ErrNotEnoughGold:
		s.msg = "你的金幣不足！"
	case gamestate.ErrNoSlot:
		s.msg = "沒有空的部隊欄位了！"
	case nil:
		s.msg = "已招募"
		s.sel = 0
	default:
		s.msg = err.Error()
	}
}

// menuLines 逐字對齊 C recruit_soldiers 底部框(game.c:2216-2249):表頭「招募士兵」+
// 5 行兵種(字母/名稱/招募價)+ 右欄 GP/選定字母/上限/數量。右欄以空白 padding 近似
// C 的 KB_iloc 絕對定位(同 town.go menuLines 的做法;右對齊參差是跨畫面已知 cosmetic)。
func (s *RecruitSoldiersScreen) menuLines() []string {
	gold := 0
	if s.gs != nil {
		gold = s.gs.Gold
	}
	lines := []string{fmt.Sprintf("招募士兵            GP=%dK", gold/1000)}
	for i, tid := range s.homeTroops {
		name := ""
		cost := 0
		if s.assets != nil && tid >= 0 && tid < len(s.assets.Troops) {
			name = s.assets.Troops[tid].Name
			cost = s.assets.Troops[tid].GoldCost
		}
		lines = append(lines, fmt.Sprintf("%c) %-8s %d", 'A'+i, name, cost))
	}
	// 右欄狀態(對齊 C:whom==0 顯示 "(A-C)" 提示,否則顯示選定字母 + 上限 + 數量)。
	if s.whom == 0 {
		lines = append(lines, "(A-C) ?")
	} else {
		lines = append(lines, fmt.Sprintf("(A-C) %c  上限=%d  數量 %d", 'A'+s.whom-1, s.max(), s.sel))
	}
	if s.msg != "" {
		lines = append(lines, s.msg)
	}
	return lines
}

func (s *RecruitSoldiersScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	// 背景:家鄉城堡內部(locID=0=cstl)+ 迎賓兵種(bgTroop);對齊
	// C draw_location(loc_id, troop_id, frame),frame 取靜態 0(同 recruit.go 慣例)。
	drawLocationBg(dst, LocationBg(s.locID), s.bgTroop, 0)
	drawSidebar(dst, s.gs, 0)
	drawBottomFrame(dst, s.assets, s.menuLines())
}

func (s *RecruitSoldiersScreen) Keymap() input.Keymap {
	if s.whom == 0 {
		letters := make([]input.LetterItem, len(s.homeTroops))
		for i, tid := range s.homeTroops {
			name := ""
			if s.assets != nil && tid >= 0 && tid < len(s.assets.Troops) {
				name = s.assets.Troops[tid].Name
			}
			letters[i] = input.LetterItem{Rune: rune('a' + i), Label: name}
		}
		return input.Keymap{Directions: false, Cancel: "離開", Letters: letters}
	}
	return input.Keymap{Directions: true, Confirm: "招募", Cancel: "返回"}
}
