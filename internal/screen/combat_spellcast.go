package screen

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/combat"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// spellState 是戰鬥施法 UI 的子狀態。
type spellState int

const (
	spellNone     spellState = iota // 正常戰鬥,未施法
	spellMenu                       // 選戰鬥法術(A-G)
	spellTarget                     // 選單一目標
	spellTeleDest                   // 瞬移術:已選來源,選目的地
	spellDone                       // 顯示施法結果訊息,任意鍵返回
)

// combatSpellCount 是戰鬥法術數(index 0-6,對齊 choose_spell 左欄)。
const combatSpellCount = 7

// playerGS 回傳玩家側(side 0)的 GameState(施法威力/法術庫存來源);無則 nil。
func (s *CombatScreen) playerGS() *gamestate.GameState {
	return s.combat.Heroes[0]
}

// openSpellMenu 對齊 C choose_spell 進入前的檢查(game.c:4283-4295):不會魔法 → 提示;
// 本回合已施法(c.Spells>0)→ 提示;否則開戰鬥法術選單。
func (s *CombatScreen) openSpellMenu() {
	gs := s.playerGS()
	if gs == nil || !gs.KnowsMagic {
		s.spellState = spellDone
		s.spellMsg = "你不會魔法。"
		return
	}
	if s.combat.Spells > 0 {
		s.spellState = spellDone
		s.spellMsg = "每回合只能施放 1 個法術!"
		return
	}
	s.spellState = spellMenu
	s.spellSel = 0
}

func (s *CombatScreen) updateSpell(a input.Action) Transition {
	switch s.spellState {
	case spellDone:
		if !a.IsNone() {
			s.spellState = spellNone
			s.spellMsg = ""
		}
		return Stay()
	case spellMenu:
		switch a.Kind {
		case input.ActUp:
			if s.spellSel > 0 {
				s.spellSel--
			}
		case input.ActDown:
			if s.spellSel < combatSpellCount-1 {
				s.spellSel++
			}
		case input.ActConfirm:
			s.selectSpell(s.spellSel)
		case input.ActLetter:
			if a.Rune >= 'a' && a.Rune < rune('a'+combatSpellCount) {
				s.selectSpell(int(a.Rune - 'a'))
			}
		case input.ActCancel:
			s.spellState = spellNone
		}
		return Stay()
	case spellTarget, spellTeleDest:
		return s.updateTarget(a)
	}
	return Stay()
}

// selectSpell 選定戰鬥法術 idx(0-6):未學會則提示;否則進目標選取(游標放當前單位)。
func (s *CombatScreen) selectSpell(idx int) {
	gs := s.playerGS()
	if gs == nil || idx < 0 || idx >= len(gs.Spells) || gs.Spells[idx] == 0 {
		s.spellState = spellDone
		s.spellMsg = "你還不會這個法術!"
		return
	}
	s.spellID = idx
	u := &s.combat.Units[s.combat.Side][s.combat.UnitID]
	s.curX, s.curY = u.X, u.Y
	s.spellState = spellTarget
}

// updateTarget 處理目標游標:D-pad 移動,Confirm 選定,Cancel 取消施法。
func (s *CombatScreen) updateTarget(a input.Action) Transition {
	switch a.Kind {
	case input.ActUp:
		if s.curY > 0 {
			s.curY--
		}
	case input.ActDown:
		if s.curY < combat.BoardH-1 {
			s.curY++
		}
	case input.ActLeft:
		if s.curX > 0 {
			s.curX--
		}
	case input.ActRight:
		if s.curX < combat.BoardW-1 {
			s.curX++
		}
	case input.ActConfirm:
		if s.spellState == spellTeleDest {
			s.resolveTeleportDest()
		} else {
			s.resolveTarget()
		}
	case input.ActCancel:
		s.spellState = spellNone
	}
	return Stay()
}

// unitAt 回報 (x,y) 上是否有單位及其 (side,id)。
func (s *CombatScreen) unitAt(x, y int) (side, id int, ok bool) {
	c := s.combat
	for sd := 0; sd < combat.MaxSides; sd++ {
		for i := 0; i < combat.MaxUnits; i++ {
			u := &c.Units[sd][i]
			if u.Count > 0 && u.X == x && u.Y == y {
				return sd, i, true
			}
		}
	}
	return 0, 0, false
}

// resolveTarget 對游標所在目標套用當前法術效果(依 spell action 決定敵/我方限制),
// 成功則消耗法術。瞬移術改走兩段(選來源→選目的地)。
func (s *CombatScreen) resolveTarget() {
	gs := s.playerGS()
	c := s.combat
	side, id, ok := s.unitAt(s.curX, s.curY)
	action := gamestate.SpellAction(s.spellID)

	// 瞬移:第一段選任意單位當來源,轉到選目的地。
	if action == gamestate.SpellTeleport {
		if !ok {
			s.finishSpell("這裡沒有部隊。", false)
			return
		}
		s.teleSide, s.teleID = side, id
		s.spellState = spellTeleDest
		return
	}

	// 其餘法術:依動作決定敵/我方目標。
	ownTarget := action == gamestate.SpellClone || action == gamestate.SpellResurrect
	if !ok || (ownTarget && side != 0) || (!ownTarget && side != 1) {
		s.finishSpell("目標無效。", false)
		return
	}

	msg := ""
	switch action {
	case gamestate.SpellClone:
		n := c.CloneTroop(s.assets, gs.SpellPower, id)
		msg = fmt.Sprintf("複製出 %d 名。", n)
	case gamestate.SpellFreeze:
		if c.FreezeTroop(s.assets, side, id) == -1 {
			s.finishSpell("這個法術似乎沒有效果!", true)
			return
		}
		msg = "敵軍被凍結了。"
	case gamestate.SpellResurrect:
		n := c.ResurrectTroop(gs.SpellPower, side, id)
		if n == 0 {
			s.finishSpell("這個法術似乎沒有效果!", true)
			return
		}
		msg = fmt.Sprintf("復活了 %d 名。", n)
	case gamestate.SpellDamage, gamestate.SpellTurnUndead:
		undead := action == gamestate.SpellTurnUndead
		base := gamestate.SpellBasePower(s.spellID)
		kills := c.MagicDamage(s.assets, s.rng, gs.SpellPower, base, side, id, undead)
		if kills < 0 {
			s.finishSpell("這個法術似乎沒有效果!", true)
			return
		}
		msg = fmt.Sprintf("消滅了 %d 名敵軍。", kills)
	default:
		s.finishSpell("此法術尚未實作。", false)
		return
	}
	s.finishSpell(msg, true)
}

// resolveTeleportDest 把已選來源單位移到游標所在的空格(對齊 unit_relocate)。
func (s *CombatScreen) resolveTeleportDest() {
	c := s.combat
	if _, _, occupied := s.unitAt(s.curX, s.curY); occupied || c.Obstacle(s.curX, s.curY) {
		s.finishSpell("那個位置無法傳送。", false)
		return
	}
	c.RelocateUnit(s.teleSide, s.teleID, s.curX, s.curY)
	s.finishSpell("部隊已傳送。", true)
}

// finishSpell 收尾:consumed 為 true 時消耗一個該法術 + 標記本回合已施法,顯示結果訊息。
func (s *CombatScreen) finishSpell(msg string, consumed bool) {
	if consumed {
		if gs := s.playerGS(); gs != nil && s.spellID >= 0 && s.spellID < len(gs.Spells) && gs.Spells[s.spellID] > 0 {
			gs.Spells[s.spellID]--
		}
		s.combat.Spells++ // 對齊 C combat->spells++(每回合限一個)
	}
	s.spellState = spellDone
	s.spellMsg = msg
}

// drawSpellOverlay 在戰鬥畫面上疊出施法 UI(選單 / 目標游標 / 結果訊息)。
func (s *CombatScreen) drawSpellOverlay(dst *ebiten.Image) {
	if s.spellState == spellNone {
		return
	}
	if s.assets == nil || s.assets.Font == nil {
		return
	}
	switch s.spellState {
	case spellDone:
		render.DrawText(dst, s.assets.Font, s.spellMsg, statusX+render.CJKCell, statusY, color.White)
	case spellMenu:
		s.drawSpellMenu(dst)
	case spellTarget, spellTeleDest:
		// 目標游標(黃框)+ 提示。
		px, py := mapX+s.curX*mapTileW, mapY+s.curY*mapTileH
		vector.StrokeRect(dst, float32(px), float32(py), mapTileW, mapTileH, 2, color.RGBA{255, 80, 80, 255}, false)
		hint := "選擇目標,Confirm 施放"
		if s.spellState == spellTeleDest {
			hint = "選擇傳送目的地"
		}
		render.DrawText(dst, s.assets.Font, hint, statusX+render.CJKCell, statusY, color.White)
	}
}

// drawSpellMenu 畫戰鬥法術選單(7 個,顯示庫存數 + 名稱,高亮目前選項)。
func (s *CombatScreen) drawSpellMenu(dst *ebiten.Image) {
	spells := gamestate.LoadSpells(s.assets)
	gs := s.playerGS()
	lines := []string{"施放哪個戰鬥法術?"}
	for i := 0; i < combatSpellCount; i++ {
		name := ""
		if i < len(spells) {
			name = spells[i].Name
		}
		cnt := 0
		if gs != nil && i < len(gs.Spells) {
			cnt = gs.Spells[i]
		}
		mark := "  "
		if i == s.spellSel {
			mark = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%c) %2d %s", mark, 'A'+i, cnt, name))
	}
	drawBottomFrame(dst, s.assets, lines)
}
