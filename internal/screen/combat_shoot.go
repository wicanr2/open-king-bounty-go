package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/combat"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// 玩家側遠程射擊 + 目標游標(target mode),對齊 C 版 unit_try_shoot(src/game.c:5962)
// 與 pick_target(src/game.c:5823)。缺口背景見 combatscreen.go:目前玩家回合只有
// 移動/近戰/待命/撤退/施法,遠攻兵種只能走過去砍——這裡補上「下令射擊 + 選目標」。
//
// C 對齊摘要:
//   - 進入條件(canShootCurrent):目前是玩家側受控單位、還有彈藥、且未被貼身
//     (combat.CanShoot,對齊 game.c:5970 的 `!u->shots || unit_surrounded(...)`)。
//   - 目標游標(pick_target filter=4):方向鍵自由移動(夾在棋盤內),只有游標落在
//     「敵方(side 1)或失控單位」上按 Confirm 才成立;落在空格/己方/障礙不接受。
//     C 的游標初始在自身格;這裡改為預設落在最近的合法敵人(方向鍵離散操作/觸控更好用),
//     此為刻意偏離、其餘判定照 C。
//   - 射擊結算:走既有 combat.UnitRangedShot(damage.go,對齊 unit_ranged_shot):
//     扣 1 彈藥、算遠程傷害、無反擊;射完該單位 acted=1、行動結束(advance)。
//   - 鍵位:C 版遠攻鍵是 's'(combat_state 表 game.c:4800 → COMBAT_SHOOT),這裡沿用 's'。

// canShootCurrent 回報「目前行動單位」是不是玩家可以下令射擊的對象:
// 玩家側(side 0)、受控(非 OutOfControl)、且 combat.CanShoot 成立。
func (s *CombatScreen) canShootCurrent() bool {
	c := s.combat
	if c.Side != 0 {
		return false
	}
	u := &c.Units[c.Side][c.UnitID]
	if u.OutOfControl {
		return false
	}
	return c.CanShoot(c.Side, c.UnitID)
}

// enterShoot 進入射擊 target mode:把游標預設放在最近的合法敵人上。
// 不可射(無彈藥/被貼身/非玩家受控單位)時直接返回,不進入 target mode(對齊 C:
// unit_try_shoot 開頭 combat_error("Can't Shoot") 後即 return)。
func (s *CombatScreen) enterShoot() {
	if !s.canShootCurrent() {
		return
	}
	c := s.combat
	u := &c.Units[c.Side][c.UnitID]

	bx, by := u.X, u.Y
	best := 1 << 30
	for j := 0; j < combat.MaxSides; j++ {
		for i := 0; i < combat.MaxUnits; i++ {
			o := &c.Units[j][i]
			if o.Count == 0 {
				continue
			}
			if j == c.Side && i == c.UnitID {
				continue
			}
			if !(j == 1 || o.OutOfControl) { // 合法目標 = 敵方或失控(對齊 pick_target filter 4)
				continue
			}
			if d := combat.Distance(u.X, u.Y, o.X, o.Y); d < best {
				best = d
				bx, by = o.X, o.Y
			}
		}
	}
	s.curX, s.curY = bx, by
	s.shooting = true
}

// shootTargetAt 回報 (x,y) 是否為合法射擊目標(敵方或失控單位,count>0),
// 對應 pick_target filter=4 的接受條件(src/game.c:5852-5860)。
func (s *CombatScreen) shootTargetAt(x, y int) (side, id int, ok bool) {
	sd, i, found := s.unitAt(x, y)
	if !found {
		return 0, 0, false
	}
	u := &s.combat.Units[sd][i]
	if sd == 1 || u.OutOfControl {
		return sd, i, true
	}
	return 0, 0, false
}

// updateShoot 處理射擊 target mode 的輸入:方向鍵移游標(夾在棋盤內、自由移動),
// Confirm 對游標處敵人射擊,Cancel 退出回一般指令(對齊 pick_target 的按鍵語意)。
func (s *CombatScreen) updateShoot(a input.Action) Transition {
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
		if side, id, ok := s.shootTargetAt(s.curX, s.curY); ok {
			c := s.combat
			u := &c.Units[c.Side][c.UnitID]
			c.UnitRangedShot(s.assets, s.rng, c.Side, c.UnitID, side, id)
			u.Acted = true // 對齊 unit_try_shoot:射完 u->acted = 1,行動結束
			s.shooting = false
			s.advance(false)
		}
		// 游標不在合法目標上:留在 target mode(對齊 pick_target 只在 accept 時 done)。
	case input.ActCancel:
		s.shooting = false // 取消射擊,回一般指令
	}
	return Stay()
}

// drawShootCursor 畫出射擊目標游標(青色準星,與「輪到誰」的黃框刻意區隔),
// 並把狀態列換成射擊提示。沿用像素風:銳利 1px 線 + 中心十字 + 四角框。
func (s *CombatScreen) drawShootCursor(dst *ebiten.Image) {
	if !s.shooting {
		return
	}
	px, py := mapX+s.curX*mapTileW, mapY+s.curY*mapTileH

	// 閃爍:青/白交替,任一相位都清楚可見(截圖不會落在「全隱」相位)。
	col := color.RGBA{80, 230, 255, 255} // 青
	if s.tick/15%2 == 0 {
		col = color.RGBA{255, 255, 255, 255} // 白
	}

	fx, fy := float32(px), float32(py)
	fw, fh := float32(mapTileW), float32(mapTileH)
	// 外框(2px,與黃色回合框同粗但顏色不同)。
	vector.StrokeRect(dst, fx, fy, fw, fh, 2, col, false)
	// 中心十字準星。
	cx, cy := fx+fw/2, fy+fh/2
	vector.StrokeLine(dst, cx-3, cy, cx+3, cy, 1, col, false)
	vector.StrokeLine(dst, cx, cy-3, cx, cy+3, 1, col, false)

	// 狀態列換成射擊提示(先清底再印,避免和一般指令列文字疊字)。
	if s.assets != nil && s.assets.Font != nil {
		vector.DrawFilledRect(dst, statusX, statusY, statusW, statusH, colorStatus, false)
		render.DrawText(dst, s.assets.Font, "射擊:方向鍵選敵,按「射擊」發射,取消返回", statusX+render.CJKCell, statusY, color.White)
	}
}
