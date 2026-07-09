// Package input 是輸入抽象接縫:桌面鍵盤與手機觸控都收斂成同一套語義 Action,
// screen 狀態機與 gamestate 只吃 Action,永遠不碰 raw key / raw touch。
//
// 這是 Go 重寫相對 C 版的關鍵改良。C 版觸控是「手指 → 合成 SDLK_* → 餵回 KB_event()」
// 的 overlay hack(疊在鍵盤引擎上)。這裡把語義抽出來當一等公民:
//
//	鍵盤 (keyboard.go)  ─┐
//	                      ├─▶ Action ──▶ screen / gamestate
//	觸控 (touch.go, P7) ─┘
//
// 另一半是 Keymap:每個畫面「宣告」它當下接受哪些 Action。觸控層讀 Keymap 決定
// 要浮出哪些按鈕(方向盤/A・B/情境字母列),對齊 openkb 每畫面一張 KB_keymap 表的做法
// (見 docs/android/ui-design.md)。本套件純語義、無渲染、無引擎依賴,可 headless 測。
package input

// ActionKind 是遊戲邏輯真正理解的語義輸入類別。
type ActionKind uint8

const (
	ActNone        ActionKind = iota
	ActUp                     // 方向:上(地圖移動 / 清單上移 / 戰鬥走位)
	ActDown                   // 方向:下
	ActLeft                   // 方向:左
	ActRight                  // 方向:右
	ActConfirm                // 確認 / 繼續(鍵盤 Enter・Space;觸控 A 鈕)
	ActCancel                 // 返回 / 取消(鍵盤 ESC;觸控 B 鈕)
	ActLetter                 // 字母快捷(選單 A–E、選角 A–D…);字元放 Rune
	ActDigit                  // 數字(招募數量、難度…);字元放 Rune('0'..'9')
	ActYes                    // 是(Y)
	ActNo                     // 否(N)
	ActPageUp                 // 清單上一頁
	ActPageDown               // 清單下一頁
	ActHome                   // 清單頂
	ActEnd                    // 清單底
	ActMenu                   // 系統選單(觸控右上 ☰)
	ActThemeCycle             // 切美術主題(F8)
	ActMusicToggle            // 切音樂(F9)
	ActQuitSave               // 存檔離開(F10)
)

// Action 是一次語義輸入。Letter/Digit 用 Rune 帶實際字元,其餘 Rune 為 0。
type Action struct {
	Kind ActionKind
	Rune rune
}

// None 是「無輸入」的零值,方便比較。
var None = Action{Kind: ActNone}

// IsNone 回報是否為無輸入。
func (a Action) IsNone() bool { return a.Kind == ActNone }

// Letter 造一個字母快捷 Action(統一轉小寫,對齊 openkb keymap 的小寫慣例)。
func Letter(r rune) Action {
	if r >= 'A' && r <= 'Z' {
		r += 'a' - 'A'
	}
	return Action{Kind: ActLetter, Rune: r}
}

// Digit 造一個數字 Action。
func Digit(r rune) Action { return Action{Kind: ActDigit, Rune: r} }

// LetterItem 是情境字母列上的一顆按鈕:字母 + 中文標籤(如 {'a', "接任務"})。
type LetterItem struct {
	Rune  rune
	Label string
}

// Keymap 是「當前畫面接受哪些 Action」的宣告。觸控層據此渲染控制;
// 桌面層可用它做提示/過濾。空欄位 = 該畫面不接受該類輸入(不渲染對應控制)。
//
// 用法:每個 screen 提供一個 CurrentKeymap() Keymap;畫面切換時控制列自動跟著換,
// 不會有一堆無用按鈕擋畫面(見 ui-design.md「情境快捷列」)。
type Keymap struct {
	Directions  bool         // D-pad 相關(地圖移動 / 清單 / 戰鬥走位)
	Confirm     string       // A 鈕標籤;"" = 不接受確認
	Cancel      string       // B 鈕標籤;"" = 不接受取消
	Letters     []LetterItem // 情境字母列(選單項);空 = 無
	LetterRects []Rect       // 選填:各字母選項的觸控熱區矩形。提供時,字母熱區改用這些
	//               矩形且設為隱形(疊在畫面既有選單文字行上,不另畫按鈕列),長度
	//               應對齊 Letters;不提供則字母走預設的下方可見按鈕列。
	Digits bool     // 數字步進器相關(數量輸入)
	YesNo  bool     // 是非兩鈕
	Scroll bool     // 清單捲動(PageUp/Down)
	System []Action // 收進 ☰ 的系統項(F8/F9/F10…)
}
