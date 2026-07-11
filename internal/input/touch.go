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

	// 系統鍵(主題/作弊/音樂)不再全域浮出:改收進世界地圖專屬的「選單 dialog」
	// (screen.SystemMenuScreen),由世界地圖頂部一顆叫出鈕(km.Buttons 的 ActMenu)開啟。
	// 故 Controls() 不再無條件附加系統鍵——城鎮/戰鬥等畫面本就不該有這些次要/開發鍵。
	// 各畫面若要任意動作鈕,走 km.Buttons(見本函式尾端附加)。

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
	// 情境字母列。每顆字母是「隱形疊在畫面既有視覺上的熱區」或「下方可見按鈕」二選一,
	// 由 LetterRects[i] 決定(逐顆,可混用):
	//   (a) LetterRects[i] 有效(W>0):字母熱區用該矩形、設 Hidden(不另畫按鈕),點畫面
	//       既有視覺(選單那一行 / sidebar 面板)即命中該字母——對齊 C 選單的乾淨外觀,
	//       也是「功能疊到 sidebar 面板」的做法(如世界地圖:錢袋面板→角色、拼圖面板→拼圖)。
	//   (b) LetterRects[i] 未提供或為零值(W==0):在下方中央排一顆可見按鈕(每顆 30 寬)。
	// 下方可見按鈕用獨立計數 bottomIdx 排位,不受被疊到 sidebar 的隱形熱區影響——這樣同一
	// 畫面可「部分疊 sidebar、部分留下方主排」而下方那排仍緊密無縫。每列最多容納數
	// (x=82 起、每顆 32 寬):(320-82)/32 = 7 顆;超過溢出到「上一排」(y 往上 22),
	// 不像舊版 break 丟掉——否則溢出的字母在觸控上永遠不可達。見 docs/theme-switching-plan.md。
	const lx0, lgap, lw, ly0, lrow = 82, 32, 30, 150, 22
	perRow := (screenW - lx0) / lgap
	bottomIdx := 0
	for i, it := range t.km.Letters {
		if i < len(t.km.LetterRects) && t.km.LetterRects[i].W > 0 {
			cs = append(cs, Control{t.km.LetterRects[i], Letter(it.Rune), it.Label, true})
			continue
		}
		col := bottomIdx % perRow
		row := bottomIdx / perRow
		bottomIdx++
		cs = append(cs, Control{Rect{lx0 + col*lgap, ly0 - row*lrow, lw, 20}, Letter(it.Rune), it.Label, false})
	}
	if t.km.YesNo {
		cs = append(cs,
			Control{Rect{110, 150, 40, 20}, Action{Kind: ActYes}, "是", false},
			Control{Rect{170, 150, 40, 20}, Action{Kind: ActNo}, "否", false},
		)
	}
	// 畫面自訂按鈕(任意 Action)原樣附加在最後。呼叫端自行決定其在清單中的相對順序
	// (Resolve 回傳第一個命中者;如 dialog 把各功能鈕排在前、全螢幕關閉熱區排最後)。
	cs = append(cs, t.km.Buttons...)
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
