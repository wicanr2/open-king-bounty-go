package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// 城堡守軍操作模式,對齊 C visit_own_castle 的 MODE_GARRISON/MODE_REMOVE(game.c:2385)。
// C 初始為 MODE_REMOVE(撤離)。
const (
	castleModeGarrison = 0 // 駐防:把玩家部隊放進城堡
	castleModeRemove   = 1 // 撤離:把城堡守軍收回玩家隊伍
)

var castleModeNames = [2]string{"駐防", "撤離"}

// CastleOwnScreen 對齊 C visit_own_castle(game.c:2381):踩到自家城堡時的守軍管理。
// 兩種模式(駐防/撤離)以「切換」互換;A-E 對左欄五格執行對應操作。
//
// C 用空白鍵切換模式、A-E 選格。Go 版沒有空白鍵動作,改用 Confirm 當「切換模式」
// (本畫面 Confirm 沒有其他用途),A-E 直選格子,對齊 five_choices_and_space 的語意。
type CastleOwnScreen struct {
	gs       *gamestate.GameState
	assets   *kbdata.Assets
	castleID int
	name     string
	mode     int
	msg      string
}

// NewCastleOwnScreen 建立自家城堡守軍管理畫面,初始為撤離模式(對齊 C mode=MODE_REMOVE)。
func NewCastleOwnScreen(gs *gamestate.GameState, a *kbdata.Assets, castleID int) *CastleOwnScreen {
	s := &CastleOwnScreen{gs: gs, assets: a, castleID: castleID, mode: castleModeRemove}
	castles := gamestate.LoadCastles(a)
	if castleID >= 0 && castleID < len(castles) {
		s.name = castles[castleID].Name
	}
	return s
}

func (s *CastleOwnScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActConfirm:
		// 對齊 C key==6(空白鍵)切換模式。
		s.mode = 1 - s.mode
		s.msg = ""
	case input.ActLetter:
		if a.Rune >= 'a' && a.Rune <= 'e' {
			s.act(int(a.Rune - 'a'))
		}
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

// act 對左欄第 slot 格執行目前模式的操作,依 C 回傳碼設定錯誤訊息。
func (s *CastleOwnScreen) act(slot int) {
	if s.gs == nil || s.castleID < 0 || s.castleID >= len(s.gs.CastleTroops) {
		return
	}
	if s.mode == castleModeGarrison {
		switch s.gs.GarrisonTroop(s.castleID, slot) {
		case 2:
			s.msg = "你不能駐防你的最後一支部隊！"
		case 1:
			s.msg = "沒有空的部隊欄位了！"
		default:
			s.msg = ""
		}
		return
	}
	if s.gs.UngarrisonTroop(s.castleID, slot) == 1 {
		s.msg = "沒有空的部隊欄位了！"
	} else {
		s.msg = ""
	}
}

// slotAt 回傳左欄第 i 格目前顯示的 (troopID, number),依模式取玩家隊伍或城堡守軍。
func (s *CastleOwnScreen) slotAt(i int) (troopID, number int) {
	if s.gs == nil {
		return 0xFF, 0
	}
	if s.mode == castleModeGarrison {
		return s.gs.Army[i].TroopID, s.gs.Army[i].Count
	}
	if s.castleID < 0 || s.castleID >= len(s.gs.CastleTroops) {
		return 0xFF, 0
	}
	return s.gs.CastleTroops[s.castleID][i], s.gs.CastleNumbers[s.castleID][i]
}

// menuLines 逐字對齊 C visit_own_castle 底部框(game.c:2458-2500):左欄城堡名 + 五格
// (兵種名/數量,空格顯示「空 無」),右欄 GP/模式名/範圍/切換提示。右欄以空白 padding
// 近似 C 的 KB_iloc 絕對定位(同 town.go 做法,右對齊參差是已知 cosmetic)。
func (s *CastleOwnScreen) menuLines() []string {
	gold := 0
	if s.gs != nil {
		gold = s.gs.Gold
	}
	lines := []string{fmt.Sprintf("城堡 %s        GP=%dK", s.name, gold/1000)}
	max := 0
	for i := 0; i < 5; i++ {
		tid, num := s.slotAt(i)
		if tid == 0xFF || num == 0 {
			lines = append(lines, fmt.Sprintf("%c) %-10s %s", 'A'+i, "空", "無"))
			continue
		}
		name := ""
		if s.assets != nil && tid >= 0 && tid < len(s.assets.Troops) {
			name = s.assets.Troops[tid].Name
		}
		lines = append(lines, fmt.Sprintf("%c) %-10s %d", 'A'+i, name, num))
		max = i + 1
	}
	rangeHint := ""
	if max > 0 {
		rangeHint = fmt.Sprintf("(A-%c)", 'A'+max-1)
	}
	lines = append(lines, fmt.Sprintf("%s %s  空白鍵→%s", castleModeNames[s.mode], rangeHint, castleModeNames[1-s.mode]))
	if s.msg != "" {
		lines = append(lines, s.msg)
	}
	return lines
}

func (s *CastleOwnScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 0)
	drawBottomFrame(dst, s.assets, s.menuLines())
}

func (s *CastleOwnScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: false,
		Confirm:    "切換",
		Cancel:     "離開",
		Letters: []input.LetterItem{
			{Rune: 'a', Label: "A"},
			{Rune: 'b', Label: "B"},
			{Rune: 'c', Label: "C"},
			{Rune: 'd', Label: "D"},
			{Rune: 'e', Label: "E"},
		},
	}
}
