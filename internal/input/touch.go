package input

// 觸控層(P7)。核心:不自己猜遊戲邏輯,而是讀當前畫面的 Keymap 自動生成控制,
// 手指點命中哪個控制就送出對應 Action。桌面滑鼠也走同一條路(可先驗)。
// 座標一律用邏輯解析度 320×200(與畫布同一座標系)。詳見 docs/android/ui-design.md。

// 邏輯畫布尺寸(與 cmd/mobile 的 logicalW/H 一致)。
const (
	screenW = 320
	screenH = 200
)

// Rect 是邏輯座標下的矩形。
type Rect struct{ X, Y, W, H int }

// Contains 回報點 (px,py) 是否落在矩形內。
func (r Rect) Contains(px, py int) bool {
	return px >= r.X && px < r.X+r.W && py >= r.Y && py < r.Y+r.H
}

// Control 是一顆浮出的觸控控制:矩形 + 點擊送出的 Action + 顯示標籤。
type Control struct {
	Rect   Rect
	Action Action
	Label  string
}

// Tap 是一次觸控/點擊點(邏輯座標)。
type Tap struct{ X, Y int }

// TouchLayout 依當前畫面 Keymap 生成控制佈局(決定浮出哪些控制、各在哪)。
type TouchLayout struct {
	km Keymap
}

// NewTouchLayout 依當前畫面 Keymap 建佈局。
func NewTouchLayout(km Keymap) TouchLayout { return TouchLayout{km: km} }

// Controls 依 Keymap 生成要浮出的控制清單(方向盤 / A·B / 情境字母列)。
// 位置固定在拇指範圍:左下 D-pad、右下 A/B、字母列在下方中央一排。
func (t TouchLayout) Controls() []Control {
	var cs []Control

	if t.km.Directions {
		cs = append(cs,
			Control{Rect{26, 150, 20, 16}, Action{Kind: ActUp}, "^"},
			Control{Rect{26, 182, 20, 16}, Action{Kind: ActDown}, "v"},
			Control{Rect{4, 166, 20, 16}, Action{Kind: ActLeft}, "<"},
			Control{Rect{48, 166, 20, 16}, Action{Kind: ActRight}, ">"},
		)
	}
	if t.km.Confirm != "" {
		cs = append(cs, Control{Rect{288, 168, 28, 22}, Action{Kind: ActConfirm}, t.km.Confirm})
	}
	if t.km.Cancel != "" {
		cs = append(cs, Control{Rect{256, 168, 28, 22}, Action{Kind: ActCancel}, t.km.Cancel})
	}
	// 情境字母列:下方中央一排,每顆 30 寬。
	for i, it := range t.km.Letters {
		x := 82 + i*32
		if x+30 > screenW {
			break // 超出畫面就不再排(避免擠出邊界)
		}
		cs = append(cs, Control{Rect{x, 150, 30, 20}, Letter(it.Rune), it.Label})
	}
	if t.km.YesNo {
		cs = append(cs,
			Control{Rect{110, 150, 40, 20}, Action{Kind: ActYes}, "是"},
			Control{Rect{170, 150, 40, 20}, Action{Kind: ActNo}, "否"},
		)
	}
	return cs
}

// Resolve 把一次觸控點對應成 Action:命中哪個控制就回它的 Action,沒命中回 None。
func (t TouchLayout) Resolve(tap Tap) Action {
	for _, c := range t.Controls() {
		if c.Rect.Contains(tap.X, tap.Y) {
			return c.Action
		}
	}
	return None
}
