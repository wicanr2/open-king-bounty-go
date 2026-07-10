package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// CastleSiegeScreen 對齊 C lay_siege(game.c:2513):踩到惡棍/怪物佔領的城堡時
// 顯示佔據者,問「是否圍攻 (y/n)？」。y → 進圍攻戰(run_combat(game,1,id),Go 以
// 城堡守軍當敵軍 Push CombatScreen);n → 回地圖。
//
// ⚠ 尚未完整處理的部分(誠實標註,對齊 PORT-STATUS.md):
//   - 圍攻勝利後的「奪城」(castle_owner=玩家、守軍換成玩家駐防、villain_caught)
//     屬 run_combat mode=1 的戰後結算,CombatScreen 目前無戰後世界狀態回寫(與一般
//     foe 戰鬥同一個缺口),故本畫面只負責「進戰鬥」,奪城邏輯待 combat 戰後 hook。
//   - 圍攻專屬的城牆障礙佈局(combat mode=1)未建模,先走一般戰鬥佈陣。
type CastleSiegeScreen struct {
	gs       *gamestate.GameState
	assets   *kbdata.Assets
	castleID int
	name     string
	lines    []string // 佔據者敘述(建構時算一次)
}

// NewCastleSiegeScreen 建立圍攻詢問畫面。castleID 為 castles.ini 索引;佔據者敘述
// 依 gs.CastleOwner[castleID] 是怪物或惡棍分兩種文字,惡棍時記下「已知統治者」旗標
// (對齊 C save_castle_owner_knowledge)。
func NewCastleSiegeScreen(gs *gamestate.GameState, a *kbdata.Assets, castleID int) *CastleSiegeScreen {
	s := &CastleSiegeScreen{gs: gs, assets: a, castleID: castleID}
	castles := gamestate.LoadCastles(a)
	if castleID >= 0 && castleID < len(castles) {
		s.name = castles[castleID].Name
	}
	s.lines = []string{fmt.Sprintf("城堡 %s", s.name), "", ""}
	if gs != nil && castleID >= 0 && castleID < len(gs.CastleOwner) {
		owner := gs.CastleOwner[castleID]
		if owner == gamestate.KBCastleMonsters {
			s.lines = append(s.lines, "各路怪物群", "佔據著這座城堡。")
		} else {
			vname := ""
			villains := gamestate.LoadVillains(a)
			vid := int(owner & gamestate.KBCastleVillain)
			if vid >= 0 && vid < len(villains) {
				vname = villains[vid].Name
			}
			s.lines = append(s.lines, vname+" 與其", "軍隊佔據著這座城堡。")
			gs.SaveCastleOwnerKnowledge(castleID)
		}
	}
	s.lines = append(s.lines, "", "是否圍攻 (y/n)？")
	return s
}

func (s *CastleSiegeScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActConfirm:
		return s.siege()
	case input.ActLetter:
		if a.Rune == 'y' {
			return s.siege()
		}
		if a.Rune == 'n' {
			return Pop()
		}
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

// siege 進圍攻戰:以城堡守軍(CastleTroops/CastleNumbers)當敵軍佈陣。
// seed 用 castleID 衍生,對齊「同一座城堡的戰鬥可重現」(戰鬥 RNG 尚未接世界持久 RNG)。
func (s *CastleSiegeScreen) siege() Transition {
	if s.gs == nil || s.castleID < 0 || s.castleID >= len(s.gs.CastleTroops) {
		return Pop()
	}
	foe := foeSquadsFrom(s.gs.CastleTroops[s.castleID], s.gs.CastleNumbers[s.castleID])
	return Push(NewCombatScreen(s.gs, s.assets, foe, uint32(s.castleID+1)))
}

func (s *CastleSiegeScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 0)
	drawBottomFrame(dst, s.assets, s.lines)
}

func (s *CastleSiegeScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: false,
		Confirm:    "圍攻",
		Cancel:     "撤退",
		Letters: []input.LetterItem{
			{Rune: 'y', Label: "圍攻"},
			{Rune: 'n', Label: "離開"},
		},
	}
}
