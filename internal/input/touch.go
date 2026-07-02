package input

// 觸控層(P7 實作)。此檔先定介面與意圖,把接縫釘住,讓 screen 現在就能寫對;
// 實際手指命中測試 / 按鈕渲染到 P7 才填,屆時只是「照 Keymap 填按鈕」,不改接線。
//
// 核心:觸控不自己猜遊戲邏輯。每幀拿當前畫面的 Keymap:
//   - Directions → 左下虛擬 D-pad(或 swipe)→ ActUp/Down/Left/Right
//   - Confirm/Cancel → 右下 A/B 兩鈕 → ActConfirm/ActCancel
//   - Letters → 情境字母列,每顆帶標籤(如 [A 接任務])→ ActLetter{該字母}
//   - Digits → 數字步進器 [−]123[+][最大][OK]
//   - YesNo → 兩顆是/否 → ActYes/ActNo
//   - Scroll → [▲][▼] → ActPageUp/ActPageDown
//   - System → 右上 ☰ 收合 → 展開後各 System Action
// 命名畫面走系統 IME(SDL/Ebiten TextInput),不經此層。詳見 docs/android/ui-design.md。

// Tap 是一次觸控點(邏輯座標 320×200,與畫布同一座標系)。
type Tap struct {
	X, Y int
}

// TouchLayout 由 Keymap 產生「這個畫面該畫哪些控制、各在哪個矩形」的佈局。
// P7 由 render 層依此畫半透明控制、由本層做命中測試。現為介面佔位。
type TouchLayout struct {
	km Keymap
}

// NewTouchLayout 依當前畫面 Keymap 建佈局(決定浮出哪些控制)。
func NewTouchLayout(km Keymap) TouchLayout { return TouchLayout{km: km} }

// Resolve 把一次觸控點對應成語義 Action(P7 實作命中測試)。
// 現階段回 None:接縫已在,screen 照 Action 寫即可,觸控填實不影響其邏輯。
func (t TouchLayout) Resolve(tap Tap) Action {
	// TODO(P7): 依 t.km 的各控制矩形做命中測試 → 對應 Action。
	return None
}
