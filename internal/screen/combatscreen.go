package screen

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/combat"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

const (
	cellW = 40
	cellH = 32
	gridX = 40 // (320 - 6*40)/2 置中
	gridY = 24
)

// CombatScreen 是戰鬥畫面:6×5 棋盤,玩家用方向鍵操作 side 0,AI 自動走 side 1。
// 走進相鄰敵格 = 攻擊;Confirm = 待命;Cancel = 撤退(回地圖)。勝負後任意鍵回地圖。
type CombatScreen struct {
	combat *combat.Combat
	assets *kbdata.Assets
	rng    kbrng.Rand
	result int // 0 進行中 / combat.ResultPlayerWon / combat.ResultAIWon
}

// NewCombatScreen 佈陣一場玩家 vs 敵方部隊的戰鬥。seed 給戰鬥 RNG(TODO:改接世界持久 RNG)。
func NewCombatScreen(gs *gamestate.GameState, a *kbdata.Assets, foe [combat.MaxUnits]gamestate.Squad, seed uint32) *CombatScreen {
	c := &combat.Combat{}
	rng := kbrng.NewGlibc(seed)
	combat.PrepareUnitsPlayer(c, 0, gs, a)
	combat.PrepareUnitsFoe(c, 1, foe, a)
	c.ResetMatch(a, rng, false)
	return &CombatScreen{combat: c, assets: a, rng: rng}
}

// NewDebugCombatScreen 建一場對佔位敵方(野狼×5)的戰鬥,供 debug/截圖用(固定 seed)。
func NewDebugCombatScreen(gs *gamestate.GameState, a *kbdata.Assets) *CombatScreen {
	var foe [combat.MaxUnits]gamestate.Squad
	foe[0] = gamestate.Squad{TroopID: 3, Count: 5}
	for i := 1; i < combat.MaxUnits; i++ {
		foe[i] = gamestate.Squad{TroopID: 255}
	}
	return NewCombatScreen(gs, a, foe, 1)
}

// advance 在當前單位行動結束(pass 或 acted)時換到下一個單位/回合。
func (s *CombatScreen) advance(pass bool) {
	c := s.combat
	u := &c.Units[c.Side][c.UnitID]
	if !pass && !u.Acted {
		return
	}
	u.Frame = 0
	if next := c.NextUnit(s.assets); next == -1 {
		c.NextTurn(s.assets)
	} else {
		c.UnitID = next
	}
}

func (s *CombatScreen) Update(a input.Action) Transition {
	c := s.combat

	// 已分勝負:任意動作回地圖。
	if s.result != 0 {
		if !a.IsNone() {
			return Pop()
		}
		return Stay()
	}
	if w := c.Winner(); w == 0 {
		s.result = combat.ResultPlayerWon
		return Stay()
	} else if w == 1 {
		s.result = combat.ResultAIWon
		return Stay()
	}

	u := &c.Units[c.Side][c.UnitID]
	if u.Count == 0 {
		s.advance(true)
		return Stay()
	}

	// AI 側(敵方或失控單位):每幀走一 tick。
	if c.Side == 1 || u.OutOfControl {
		pass := c.AIUnitThink(s.assets, s.rng)
		s.advance(pass)
		return Stay()
	}

	// 玩家側:等指令。
	dx, dy := 0, 0
	switch a.Kind {
	case input.ActUp:
		dy = -1
	case input.ActDown:
		dy = 1
	case input.ActLeft:
		dx = -1
	case input.ActRight:
		dx = 1
	case input.ActConfirm:
		s.advance(true) // 待命
		return Stay()
	case input.ActCancel:
		return Pop() // 撤退
	default:
		return Stay()
	}

	nx, ny := u.X+dx, u.Y+dy
	if !combat.InBounds(nx, ny) {
		return Stay()
	}
	// 相鄰敵格 → 攻擊;否則嘗試移動。
	if es, eid, ok := s.enemyAt(nx, ny); ok {
		c.UnitHitUnit(s.assets, s.rng, c.Side, c.UnitID, es, eid)
		u.Acted = true
	} else {
		c.MoveUnit(c.Side, c.UnitID, nx, ny)
	}
	s.advance(false)
	return Stay()
}

// enemyAt 回報 (x,y) 是否站著敵方(對玩家而言 side 1)單位。
func (s *CombatScreen) enemyAt(x, y int) (side, id int, ok bool) {
	c := s.combat
	for i := 0; i < combat.MaxUnits; i++ {
		u := &c.Units[1][i]
		if u.Count > 0 && u.X == x && u.Y == y {
			return 1, i, true
		}
	}
	return 0, 0, false
}

func (s *CombatScreen) Draw(dst *ebiten.Image) {
	c := s.combat
	// 棋盤格
	for y := 0; y < combat.BoardH; y++ {
		for x := 0; x < combat.BoardW; x++ {
			px, py := float32(gridX+x*cellW), float32(gridY+y*cellH)
			fill := color.RGBA{40, 40, 55, 255}
			if c.Obstacle(x, y) {
				fill = color.RGBA{90, 80, 70, 255} // 障礙
			}
			vector.DrawFilledRect(dst, px, py, cellW, cellH, fill, false)
			vector.StrokeRect(dst, px, py, cellW, cellH, 1, color.RGBA{70, 70, 90, 255}, false)
		}
	}
	// 單位
	for side := 0; side < combat.MaxSides; side++ {
		for i := 0; i < combat.MaxUnits; i++ {
			u := &c.Units[side][i]
			if u.Count == 0 {
				continue
			}
			px, py := float32(gridX+u.X*cellW), float32(gridY+u.Y*cellH)
			body := color.RGBA{60, 120, 220, 255} // 玩家:藍
			if side == 1 {
				body = color.RGBA{200, 60, 60, 255} // 敵方:紅
			}
			vector.DrawFilledRect(dst, px+2, py+2, cellW-4, cellH-4, body, false)
			if side == c.Side && i == c.UnitID {
				vector.StrokeRect(dst, px, py, cellW, cellH, 2, color.RGBA{240, 220, 40, 255}, false) // 當前單位:黃框
			}
			name := "?"
			if u.TroopID < len(s.assets.Troops) {
				name = string([]rune(s.assets.Troops[u.TroopID].Name)[:1]) // 首字
			}
			if s.assets.Font != nil {
				render.DrawText(dst, s.assets.Font, name, int(px)+2, int(py)+1, color.White)
			}
			ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%d", u.Count), int(px)+2, int(py)+cellH-14)
		}
	}
	// 狀態列(CJK 用 atlas 畫,ASCII 提示用 DebugPrint)
	cur := &c.Units[c.Side][c.UnitID]
	turn := "你的回合"
	if c.Side == 1 {
		turn = "敵方回合"
	}
	cname := "?"
	if cur.TroopID < len(s.assets.Troops) {
		cname = s.assets.Troops[cur.TroopID].Name
	}
	if s.assets.Font != nil {
		render.DrawText(dst, s.assets.Font, turn+" "+cname, 6, 4, color.White)
	}
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("x%d  arrows:move/attack ENTER:wait ESC:flee", cur.Count), 150, 8)

	if s.result != 0 && s.assets.Font != nil {
		msg := "勝利!"
		if s.result == combat.ResultAIWon {
			msg = "戰敗…"
		}
		render.DrawText(dst, s.assets.Font, msg, 130, 178, color.RGBA{255, 230, 60, 255})
		ebitenutil.DebugPrintAt(dst, "(press any key)", 128, 178)
	}
}

func (s *CombatScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: true, Confirm: "待命", Cancel: "撤退"}
}
