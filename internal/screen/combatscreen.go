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

// combatKind 標記這場戰鬥的來源,決定戰後把 side 1 存活單位寫回哪塊世界狀態、
// 以及戰勝的額外結算(清 foe tile / 奪城 + 履約)。
type combatKind int

const (
	combatDebug combatKind = iota // debug/截圖用,不做任何戰後世界狀態回寫
	combatFoe                     // 世界地圖遭遇 foe(對齊 C run_combat mode 0)
	combatSiege                   // 圍攻城堡(對齊 C run_combat mode 1)
)

// combatContext 描述戰鬥來源,供 applyOutcome 做 C run_combat 的戰後結算。
type combatContext struct {
	kind      combatKind
	cont      int // foe:所在洲(寫回 FoeTroops/Numbers、清 tile 用)
	x, y      int // foe:戰勝後要清成 0 的地圖 tile 座標
	foeID     int // foe:FoeTroops/Numbers 索引
	castleID  int // siege:CastleTroops/Numbers/Owner 索引
	gsPointer *gamestate.GameState
}

// CombatScreen 是戰鬥畫面:6×5 棋盤,玩家用方向鍵操作 side 0,AI 自動走 side 1。
// 走進相鄰敵格 = 攻擊;Confirm = 待命;Cancel = 撤退(回地圖)。勝負後任意鍵回地圖。
type CombatScreen struct {
	combat  *combat.Combat
	assets  *kbdata.Assets
	rng     kbrng.Rand
	result  int // 0 進行中 / combat.ResultPlayerWon / combat.ResultAIWon
	ctx     combatContext
	applied bool // applyOutcome 只跑一次(避免每幀重複結算)

	// 戰鬥施法 UI 狀態(見 combat_spellcast.go)。
	spellState spellState
	spellSel   int    // 施法選單目前高亮的法術(0-6)
	spellID    int    // 已選定要施放的戰鬥法術(0-6)
	curX, curY int    // 目標游標座標(施法選目標 / 射擊瞄準共用)
	teleSide   int    // 瞬移術:已選來源單位的 side
	teleID     int    // 瞬移術:已選來源單位的 id
	spellMsg   string // 施法結果訊息(頂列顯示)

	shooting bool // 射擊 target mode 進行中(見 combat_shoot.go)
	tick     int  // 每幀 +1,供射擊準星閃爍用
}

// NewCombatScreen 佈陣一場玩家 vs 敵方部隊的戰鬥(debug/一般用,不做戰後世界狀態回寫)。
// seed 給戰鬥 RNG(TODO:改接世界持久 RNG)。有世界狀態回寫需求時用
// NewCombatScreenFoe / NewCombatScreenSiege。
func NewCombatScreen(gs *gamestate.GameState, a *kbdata.Assets, foe [combat.MaxUnits]gamestate.Squad, seed uint32) *CombatScreen {
	c := &combat.Combat{}
	rng := kbrng.NewGlibc(seed)
	if gs != nil {
		c.AIEvolved = gs.OptAIMode // opt_ai_mode:敵方 AI 走進化版決策
	}
	combat.PrepareUnitsPlayer(c, 0, gs, a)
	combat.PrepareUnitsFoe(c, 1, foe, a)
	c.ResetMatch(a, rng, false)
	return &CombatScreen{combat: c, assets: a, rng: rng, ctx: combatContext{kind: combatDebug, gsPointer: gs}}
}

// NewCombatScreenFoe 佈陣一場世界地圖 foe 遭遇戰(對齊 C run_combat mode 0):
// 戰後把 side 1 存活單位寫回 gs.FoeTroops/Numbers[cont][foeID],戰勝則把 (x,y) tile 清 0。
func NewCombatScreenFoe(gs *gamestate.GameState, a *kbdata.Assets, foe [combat.MaxUnits]gamestate.Squad, cont, foeID, x, y int) *CombatScreen {
	s := NewCombatScreen(gs, a, foe, uint32(x*kbdata.LevelH+y+1))
	s.ctx = combatContext{kind: combatFoe, cont: cont, foeID: foeID, x: x, y: y, gsPointer: gs}
	return s
}

// NewCombatScreenSiege 佈陣一場圍攻戰(對齊 C run_combat mode 1):以城堡守軍為敵軍,
// 戰後把 side 1 存活單位寫回 gs.CastleTroops/Numbers[castleID];戰勝則奪城(owner=玩家)、
// 若守軍首腦正是當前契約目標則履約(FulfillContract)。
func NewCombatScreenSiege(gs *gamestate.GameState, a *kbdata.Assets, castleID int) *CombatScreen {
	var foe [combat.MaxUnits]gamestate.Squad
	for i := 0; i < combat.MaxUnits; i++ {
		if castleID >= 0 && castleID < len(gs.CastleTroops) && gs.CastleNumbers[castleID][i] > 0 {
			foe[i] = gamestate.Squad{TroopID: gs.CastleTroops[castleID][i], Count: gs.CastleNumbers[castleID][i]}
		} else {
			foe[i] = gamestate.Squad{TroopID: 255}
		}
	}
	c := &combat.Combat{}
	rng := kbrng.NewGlibc(uint32(castleID + 1))
	if gs != nil {
		c.AIEvolved = gs.OptAIMode
	}
	combat.PrepareUnitsPlayer(c, 0, gs, a)
	combat.PrepareUnitsCastle(c, 1, foe, a)
	// 圍攻戰(run_combat mode=1)用城堡佈局:castleUmap 佈防守方 + castleOmap 城牆障礙
	// (對齊 C reset_match 的 castle 分支)。取代先前傳 false 的「無牆平地」佈局。
	c.ResetMatch(a, rng, true)
	return &CombatScreen{combat: c, assets: a, rng: rng, ctx: combatContext{kind: combatSiege, castleID: castleID, gsPointer: gs}}
}

// applyOutcome 做 C run_combat 收尾的戰後結算(只跑一次)。順序對齊 game.c:3598-3652:
// 先寫回兩側存活單位;foe 戰勝清 tile;戰敗 temp_death;戰勝結算 spoils/奪城/履約。
func (s *CombatScreen) applyOutcome() {
	if s.applied {
		return
	}
	s.applied = true
	gs := s.ctx.gsPointer
	if gs == nil || s.ctx.kind == combatDebug {
		return
	}
	won := s.result == combat.ResultPlayerWon

	// 1) 存活單位寫回世界狀態(對齊 accept_units_player + accept_units_foe/castle)。
	combat.AcceptUnitsPlayer(s.combat, 0, gs)
	switch s.ctx.kind {
	case combatFoe:
		troops, numbers := combat.AcceptSquads(s.combat, 1, 3) // foe 只 3 格
		for i := 0; i < 3; i++ {
			gs.FoeTroops[s.ctx.cont][s.ctx.foeID][i] = troops[i]
			gs.FoeNumbers[s.ctx.cont][s.ctx.foeID][i] = numbers[i]
		}
		if won && gs.WorldMap != nil {
			gs.WorldMap.Set(s.ctx.cont, s.ctx.x, s.ctx.y, 0) // 對齊 C:戰勝把 foe tile 清 0
		}
	case combatSiege:
		troops, numbers := combat.AcceptSquads(s.combat, 1, combat.MaxUnits)
		for i := 0; i < combat.MaxUnits; i++ {
			gs.CastleTroops[s.ctx.castleID][i] = troops[i]
			gs.CastleNumbers[s.ctx.castleID][i] = numbers[i]
		}
	}

	// 2) 戰敗:temp_death(清隊伍發農夫)。
	if s.result == combat.ResultAIWon {
		gs.TempDeath(s.assets)
		return
	}

	// 3) 戰勝:spoils + 圍攻奪城/履約(對齊 game.c:3630-3648)。
	if won {
		gs.Gold += s.combat.Spoils[1]
		if s.ctx.kind == combatSiege {
			owner := gs.CastleOwner[s.ctx.castleID]
			if owner != gamestate.KBCastleMonsters {
				vid := int(owner & gamestate.KBCastleVillain)
				if int(gs.Contract) == vid {
					gs.FulfillContract(s.assets, vid)
				}
			}
			gs.CastleOwner[s.ctx.castleID] = gamestate.KBCastlePlayer
		}
	}
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

// NewDebugShootCombatScreen 同 NewDebugCombatScreen,但把玩家首格換成弓箭手(troop 8,
// 有 12 發彈藥),並直接進入射擊 target mode——供截圖驗證射擊 UI(準星 + 觸控射擊鈕)。
// 人數壓在領導力上限內(5)以免變成失控單位而被 AI 接管。
func NewDebugShootCombatScreen(gs *gamestate.GameState, a *kbdata.Assets) *CombatScreen {
	gs.Army[0] = gamestate.Squad{TroopID: 8, Count: 5} // 弓箭手
	s := NewDebugCombatScreen(gs, a)
	s.enterShoot()
	return s
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
	s.tick++ // 準星閃爍用計時(每幀遞增)

	// 已分勝負:任意動作回地圖。
	if s.result != 0 {
		if !a.IsNone() {
			return Pop()
		}
		return Stay()
	}
	if w := c.Winner(); w == 0 {
		s.result = combat.ResultPlayerWon
		s.applyOutcome()
		return Stay()
	} else if w == 1 {
		s.result = combat.ResultAIWon
		s.applyOutcome()
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

	// 施法 UI 進行中(選單/選目標)時,輸入全交給施法流程處理。
	if s.spellState != spellNone {
		return s.updateSpell(a)
	}

	// 射擊 target mode 進行中時,輸入全交給射擊瞄準流程處理(見 combat_shoot.go)。
	if s.shooting {
		return s.updateShoot(a)
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
	case input.ActLetter:
		switch a.Rune {
		case 'z':
			s.openSpellMenu() // 施放戰鬥法術(對齊 C choose_spell 戰鬥模式)
		case 's':
			s.enterShoot() // 遠程射擊(對齊 C unit_try_shoot,鍵位 's');不可射時 enterShoot 內部忽略
		}
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
	drawChromeFrame(dst)

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
	s.drawSpellOverlay(dst) // 施法 UI(選單/目標游標/結果訊息)疊在戰鬥畫面上
	s.drawShootCursor(dst)  // 射擊瞄準準星 + 提示(見 combat_shoot.go)

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
	km := input.Keymap{Directions: true, Confirm: "待命", Cancel: "撤退"}

	// 射擊 target mode:D-pad 移游標、Confirm 發射、Cancel 取消(對齊 pick_target 操作)。
	if s.shooting {
		km.Confirm = "射擊"
		km.Cancel = "取消"
		return km
	}

	// 施法中 Confirm 的語意改為「施放/選定」;施法選單/目標選取時提供對應提示。
	switch s.spellState {
	case spellMenu:
		km.Confirm = "選擇"
	case spellTarget, spellTeleDest:
		km.Confirm = "施放"
	case spellNone:
		// 玩家回合(會魔法且本回合未施法)時多一顆施法鈕。
		if gs := s.playerGS(); gs != nil && gs.KnowsMagic && s.combat.Spells == 0 {
			km.Letters = []input.LetterItem{{Rune: 'z', Label: "施法"}}
		}
		// 可遠攻(有彈藥且未被貼身)時,右下多浮一顆「射擊」鈕(觸控入口)。
		// 放在確認鈕正上方(y=142),不與撤退/待命(y=168)或 D-pad(左下)重疊。
		if s.canShootCurrent() {
			km.Buttons = append(km.Buttons, input.Control{
				Rect:   input.Rect{X: 288, Y: 142, W: 28, H: 22},
				Action: input.Letter('s'),
				Label:  "射擊",
			})
		}
	}
	return km
}
