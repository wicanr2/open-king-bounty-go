package screen

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/combat"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// 戰鬥棋盤格對齊 C draw_combat(game.c:1298):共用世界地圖那套 local.map 座標
// (mapX/mapY/mapTileW/mapTileH,定義在 worldmap.go),棋盤 BoardW×BoardH(6×5)
// 從 (mapX,mapY) 起、每格 mapTileW×mapTileH,不需要另外的置中常數。

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

// Draw 對齊 C draw_combat(game.c:1298-1355):黃框底(同 worldmap/charselect,見
// colorBorder 註解)→ 棋盤地形(comtiles.png,omap 值即幀索引)→ 兩側單位(兵種
// sprite,side 1 水平鏡射同一張圖)+ 部隊數字(bottom-right,對齊 C:印的是
// turn_count 不是 count)→ 頂部狀態列(對齊 draw_combat_statusbar)。
func (s *CombatScreen) Draw(dst *ebiten.Image) {
	c := s.combat
	dst.Fill(colorBorder)

	// 棋盤地形
	for y := 0; y < combat.BoardH; y++ {
		for x := 0; x < combat.BoardW; x++ {
			px, py := mapX+x*mapTileW, mapY+y*mapTileH
			if comtilesSprite != nil {
				comtilesSprite.DrawFrame(dst, int(c.Omap[y][x]), px, py)
			} else {
				fill := color.RGBA{40, 40, 55, 255}
				if c.Obstacle(x, y) {
					fill = color.RGBA{90, 80, 70, 255} // 障礙
				}
				vector.DrawFilledRect(dst, float32(px), float32(py), mapTileW, mapTileH, fill, false)
				vector.StrokeRect(dst, float32(px), float32(py), mapTileW, mapTileH, 1, color.RGBA{70, 70, 90, 255}, false)
			}
		}
	}

	// 單位(C: u->turn_count == 0 才略過不畫,不是 count == 0)
	for side := 0; side < combat.MaxSides; side++ {
		for i := 0; i < combat.MaxUnits; i++ {
			u := &c.Units[side][i]
			if u.TurnCount == 0 {
				continue
			}
			px, py := mapX+u.X*mapTileW, mapY+u.Y*mapTileH
			if sp := troopSpriteFor(u.TroopID); sp != nil {
				if side == 0 {
					sp.DrawFrame(dst, u.Frame, px, py)
				} else {
					sp.DrawFrameFlipped(dst, u.Frame, px, py)
				}
			} else {
				body := color.RGBA{60, 120, 220, 255} // 玩家:藍
				if side == 1 {
					body = color.RGBA{200, 60, 60, 255} // 敵方:紅
				}
				vector.DrawFilledRect(dst, float32(px+2), float32(py+2), mapTileW-4, mapTileH-4, body, false)
			}
			// 當前單位高亮:C 版 draw_combat 本身沒畫這個(靠另一套目標游標機制),
			// 但 Go 版鍵盤操作沒有那套游標,保留這個黃框當唯一的「輪到誰」視覺提示。
			if side == c.Side && i == c.UnitID {
				vector.StrokeRect(dst, float32(px), float32(py), mapTileW, mapTileH, 2, color.RGBA{240, 220, 40, 255}, false)
			}
			if s.assets != nil && s.assets.Font != nil {
				count := fmt.Sprintf("%d", u.TurnCount)
				tx := px + mapTileW - len(count)*render.CJKCell
				ty := py + mapTileH - render.CJKCell
				render.DrawText(dst, s.assets.Font, count, tx, ty, color.White)
			}
		}
	}

	s.drawStatusBar(dst)

	if s.result != 0 && s.assets != nil && s.assets.Font != nil {
		msg := "勝利!"
		if s.result == combat.ResultAIWon {
			msg = "戰敗…"
		}
		render.DrawText(dst, s.assets.Font, msg, 130, 178, color.RGBA{255, 230, 60, 255})
		render.DrawText(dst, s.assets.Font, "(按任意鍵)", 120, 188, color.White)
	}
}

// drawStatusBar 對齊 C draw_combat_statusbar(game.c:5604-5624):藍底(colorStatus,
// 沿用世界地圖那份)一律先清;只有玩家側(side==0)且該單位未失控時才印文字
// 「選項 / <兵種名> [F,]M<剩餘移動>[,S<剩餘彈藥>]」——C 版對 AI 側/失控單位就是
// 留空白藍條,沒有另外的「敵方回合」提示,這裡照實移植。
func (s *CombatScreen) drawStatusBar(dst *ebiten.Image) {
	vector.DrawFilledRect(dst, statusX, statusY, statusW, statusH, colorStatus, false)

	c := s.combat
	u := &c.Units[c.Side][c.UnitID]
	if c.Side != 0 || u.OutOfControl {
		return
	}
	if s.assets == nil || s.assets.Font == nil || u.TroopID < 0 || u.TroopID >= len(s.assets.Troops) {
		return
	}
	t := s.assets.Troops[u.TroopID]

	text := " 選項 / " + t.Name + " "
	if t.Abilities&kbdata.AbilFly != 0 {
		text += "F,"
	}
	text += fmt.Sprintf("M%d", u.Moves)
	if t.RangedShots > 0 {
		text += fmt.Sprintf(",S%d", u.Shots)
	}
	render.DrawText(dst, s.assets.Font, text, statusX, statusY, color.White)
}

func (s *CombatScreen) Keymap() input.Keymap {
	return input.Keymap{Directions: true, Confirm: "待命", Cancel: "撤退"}
}
