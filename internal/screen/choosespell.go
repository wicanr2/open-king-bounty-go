package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// spellHalf 是「戰鬥/冒險」各半的法術數(對齊 C half = MAX_SPELLS/2 = 7)。
const spellHalf = 7

// chooseSpellMode 目前只做冒險(overworld)施法;戰鬥施法(mode 0)待 combat 施法切片。
type chooseSpellMode int

const (
	chooseSpellAdventure chooseSpellMode = iota
	chooseSpellResult                    // 施放後顯示結果訊息
)

// ChooseSpellScreen 對齊 C choose_spell(game.c:4275)的冒險(mode==1)分支:雙欄法術
// 表(左戰鬥、右冒險),A-G 選右欄冒險法術施放。目前實作 停時/提升控制/即時軍隊 三個
// 純效果;bridge/findvillain/castlegate/towngate 顯示「待後續」(見 gamestate.CastAdventureSpell)。
type ChooseSpellScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	spells []gamestate.Spell
	mode   chooseSpellMode
	msg    []string
}

// NewChooseSpellScreen 建立冒險施法畫面。
func NewChooseSpellScreen(gs *gamestate.GameState, a *kbdata.Assets) *ChooseSpellScreen {
	return &ChooseSpellScreen{gs: gs, assets: a, spells: gamestate.LoadSpells(a)}
}

func (s *ChooseSpellScreen) Update(a input.Action) Transition {
	if s.mode == chooseSpellResult {
		if !a.IsNone() {
			return Pop()
		}
		return Stay()
	}
	switch a.Kind {
	case input.ActLetter:
		if a.Rune >= 'a' && a.Rune < rune('a'+spellHalf) {
			return s.cast(int(a.Rune - 'a'))
		}
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

// cast 施放右欄第 idx 個冒險法術(spell id = spellHalf + idx),顯示結果。
// 傳送法術(castlegate/towngate)先消耗法術再 Replace 進目的地選單(對齊 C:choose_spell
// 呼叫 select_gate 後仍 `spells[spell_id]--`,即使取消目的地也消耗)。
func (s *ChooseSpellScreen) cast(idx int) Transition {
	spellID := spellHalf + idx
	if s.gs.Spells[spellID] > 0 {
		switch gamestate.SpellAction(spellID) {
		case gamestate.SpellCastleGate:
			s.gs.Spells[spellID]--
			return Replace(NewSelectGateScreen(s.gs, s.assets, gateCastle))
		case gamestate.SpellTownGate:
			s.gs.Spells[spellID]--
			return Replace(NewSelectGateScreen(s.gs, s.assets, gateTown))
		case gamestate.SpellBridge:
			s.gs.Spells[spellID]--
			return Replace(NewBridgeScreen(s.gs, s.assets))
		case gamestate.SpellFindVillain:
			// 對齊 C:find_villain(標記惡棍城堡已知)後開 view_contract。
			s.gs.Spells[spellID]--
			s.gs.FindVillain()
			return Replace(NewViewContractScreen(s.gs, s.assets))
		}
	}
	r := s.gs.CastAdventureSpell(s.assets, spellID)
	s.mode = chooseSpellResult
	switch r.Kind {
	case gamestate.AdvNotLearned:
		s.msg = []string{"你還不會這個法術!"}
	case gamestate.AdvTimestop:
		s.msg = []string{"時間為你停駐,", fmt.Sprintf("停時延續 %d 回合。", r.Amount)}
	case gamestate.AdvRaiseControl:
		s.msg = []string{fmt.Sprintf("你的領導力提升了 %d。", r.Amount)}
	case gamestate.AdvInstantArmy:
		name := ""
		if s.assets != nil && r.TroopID >= 0 && r.TroopID < len(s.assets.Troops) {
			name = s.assets.Troops[r.TroopID].Name
		}
		s.msg = []string{fmt.Sprintf("%s%s", gamestate.NumberName(r.Amount), name), "已加入你的軍隊。"}
	case gamestate.AdvInstantArmyFail:
		s.msg = []string{"沒有空的欄位,", "也沒有此類型的部隊!"}
	default: // AdvTodo
		s.msg = []string{"此冒險法術待後續切片實作。"}
	}
	return Stay()
}

// menuLines 逐列組出雙欄法術表(對齊 C choose_spell 的排版,近似以空白 padding)。
func (s *ChooseSpellScreen) menuLines() []string {
	name := func(id int) string {
		if id >= 0 && id < len(s.spells) {
			return s.spells[id].Name
		}
		return ""
	}
	count := func(id int) int {
		if s.gs != nil && id >= 0 && id < len(s.gs.Spells) {
			return s.gs.Spells[id]
		}
		return 0
	}
	lines := []string{"        法術", "  戰鬥          冒險"}
	for i := 0; i < spellHalf; i++ {
		j := i + spellHalf
		// 左:戰鬥法術 count+name;中:字母;右:冒險法術 name+count。
		lines = append(lines, fmt.Sprintf("%2d %-6s %c %-6s %2d",
			count(i), name(i), 'A'+i, name(j), count(j)))
	}
	lines = append(lines, "施放哪個冒險法術 (A-G)?")
	return lines
}

func (s *ChooseSpellScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 0)
	if s.mode == chooseSpellResult {
		drawBottomFrame(dst, s.assets, s.msg)
		return
	}
	drawBottomFrame(dst, s.assets, s.menuLines())
}

func (s *ChooseSpellScreen) Keymap() input.Keymap {
	if s.mode == chooseSpellResult {
		return input.Keymap{Directions: false, Confirm: "繼續", Cancel: "離開"}
	}
	letters := make([]input.LetterItem, spellHalf)
	for i := 0; i < spellHalf; i++ {
		letters[i] = input.LetterItem{Rune: rune('a' + i), Label: fmt.Sprintf("%c", 'A'+i)}
	}
	return input.Keymap{Directions: false, Cancel: "離開", Letters: letters}
}
