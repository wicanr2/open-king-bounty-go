package screen

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// homeCastleLoc 是家鄉城堡背景的 GR_LOCATION sub_id(0=cstl,對齊 C
// visit_home_castle 的 draw_location(0, ...))。
const homeCastleLoc = 0

// homeCastleLetters 是家鄉城堡選單(對齊 C visit_home_castle 的
// throne_room_or_barracks:A) 招募士兵 B) 晉見國王)。
var homeCastleLetters = []input.LetterItem{
	{Rune: 'a', Label: "招募士兵"},
	{Rune: 'b', Label: "晉見國王"},
}

// CastleHomeScreen 對齊 C visit_home_castle(game.c:2307):踩到家鄉城堡(land.ini
// special0 座標)時的王座廳選單。A) 招募士兵 → RecruitSoldiersScreen;B) 晉見國王 →
// AudienceScreen。背景 draw_location(0, random_home_troop, frame)。
type CastleHomeScreen struct {
	gs      *gamestate.GameState
	assets  *kbdata.Assets
	bgTroop int // 背景迎賓兵種(C: random_troop = home_troops[rand()%5])
	sel     int // D-pad 高亮選項
}

// NewCastleHomeScreen 建立家鄉城堡畫面,隨機抽一個城堡棲地兵種當背景立繪
// (對齊 C random_troop = home_troops[rand() % 5])。
func NewCastleHomeScreen(gs *gamestate.GameState, a *kbdata.Assets) *CastleHomeScreen {
	s := &CastleHomeScreen{gs: gs, assets: a}
	if home := gamestate.HomeTroops(a); len(home) > 0 {
		s.bgTroop = home[rand.Intn(len(home))]
	}
	return s
}

func (s *CastleHomeScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActUp:
		if s.sel > 0 {
			s.sel--
		}
	case input.ActDown:
		if s.sel < len(homeCastleLetters)-1 {
			s.sel++
		}
	case input.ActConfirm:
		return s.activate(s.sel)
	case input.ActLetter:
		if a.Rune >= 'a' && a.Rune < rune('a'+len(homeCastleLetters)) {
			return s.activate(int(a.Rune - 'a'))
		}
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

// activate 觸發選單項(對齊 C key==1 招募 / key==2 謁見)。
func (s *CastleHomeScreen) activate(idx int) Transition {
	s.sel = idx
	switch idx {
	case 0:
		return Push(NewRecruitSoldiersScreen(s.gs, s.assets, homeCastleLoc, s.bgTroop))
	case 1:
		return Push(NewAudienceScreen(s.gs, s.assets))
	}
	return Stay()
}

func (s *CastleHomeScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawLocationBg(dst, LocationBg(homeCastleLoc), s.bgTroop, 0)
	drawSidebar(dst, s.gs, 0)
	drawBottomFrame(dst, s.assets, []string{
		"城堡 " + gamestate.HomeCastleName,
		"",
		"A) 招募士兵",
		"B) 晉見國王",
	})
}

func (s *CastleHomeScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: false,
		Confirm:    "選擇",
		Cancel:     "離開",
		Letters:    homeCastleLetters,
	}
}
