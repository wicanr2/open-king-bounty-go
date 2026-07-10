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
// Hidden=true 時是「隱形熱區」:仍可被 Resolve 命中,但 DrawTouchControls 不畫它。
// 用於選單型畫面(如城鎮),讓字母選項的點擊熱區直接疊在畫面既有的選單文字行上,
// 由選單文字本身當視覺,不再另畫一排按鈕蓋住選單(對齊 C 選單的乾淨外觀)。
type Control struct {
	Rect   Rect
	Action Action
	Label  string
	Hidden bool
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

	// 全域系統鍵(觸控裝置無 F8/F12):右上角「主題」切美術主題、其下「作弊」開 debug 選單。
	// 主題由 app.handleSystem 攔 ActThemeCycle → screen.CycleTheme;作弊由 worldmap 接 ActCheat
	// → 開 CheatMenuScreen(非遊戲畫面忽略)。放狀態列右端(該處通常無文字)。
	cs = append(cs,
		Control{Rect{292, 1, 26, 10}, Action{Kind: ActThemeCycle}, "主題", false},
		Control{Rect{292, 12, 26, 10}, Action{Kind: ActCheat}, "作弊", false},
	)

	if t.km.Directions {
		cs = append(cs,
			Control{Rect{26, 150, 20, 16}, Action{Kind: ActUp}, "^", false},
			Control{Rect{26, 182, 20, 16}, Action{Kind: ActDown}, "v", false},
			Control{Rect{4, 166, 20, 16}, Action{Kind: ActLeft}, "<", false},
			Control{Rect{48, 166, 20, 16}, Action{Kind: ActRight}, ">", false},
		)
	}
	if t.km.Confirm != "" {
		cs = append(cs, Control{Rect{288, 168, 28, 22}, Action{Kind: ActConfirm}, t.km.Confirm, false})
	}
	if t.km.Cancel != "" {
		cs = append(cs, Control{Rect{256, 168, 28, 22}, Action{Kind: ActCancel}, t.km.Cancel, false})
	}
	// 情境字母列。兩種佈局:
	//   (a) 若 Keymap 提供 LetterRects(選單型畫面,如城鎮):字母熱區用畫面既有選單
	//       各行的矩形,設 Hidden(不另畫按鈕),點選單那一行即選該項——對齊 C 選單。
	//   (b) 否則(如選角):在下方中央排一排可見按鈕,每顆 30 寬。
	for i, it := range t.km.Letters {
		if i < len(t.km.LetterRects) {
			cs = append(cs, Control{t.km.LetterRects[i], Letter(it.Rune), it.Label, true})
			continue
		}
		x := 82 + i*32
		if x+30 > screenW {
			break // 超出畫面就不再排(避免擠出邊界)
		}
		cs = append(cs, Control{Rect{x, 150, 30, 20}, Letter(it.Rune), it.Label, false})
	}
	if t.km.YesNo {
		cs = append(cs,
			Control{Rect{110, 150, 40, 20}, Action{Kind: ActYes}, "是", false},
			Control{Rect{170, 150, 40, 20}, Action{Kind: ActNo}, "否", false},
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
