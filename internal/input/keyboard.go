package input

// Sym 是中性的按鍵符號(不綁任何引擎)。cmd/render 層把 Ebiten 的 ebiten.Key
// 轉成 Sym/Key 再交給本套件,如此 input 層可完全 headless 測、不依賴 GL。
type Sym uint8

const (
	SymNone Sym = iota
	SymUp
	SymDown
	SymLeft
	SymRight
	SymEnter
	SymSpace
	SymEsc
	SymChar // 可列印字元,實際字元放 Key.Ch('a'..'z' / '0'..'9' / 'y' / 'n'…)
	SymPageUp
	SymPageDown
	SymHome
	SymEnd
	SymF8
	SymF9
	SymF10
)

// Key 是一次原始按鍵:符號 + (若為 SymChar)字元。
type Key struct {
	Sym Sym
	Ch  rune
}

// KeyToAction 把一個原始按鍵映成語義 Action。純函式、可測。
// 字元鍵依內容分流:y/n → 是非;0-9 → 數字;其餘字母 → 字母快捷。
// screen 收到 Action 後自行決定意義(同一個 ActLetter 在城鎮是選單、在命名是輸入字元)。
func KeyToAction(k Key) Action {
	switch k.Sym {
	case SymUp:
		return Action{Kind: ActUp}
	case SymDown:
		return Action{Kind: ActDown}
	case SymLeft:
		return Action{Kind: ActLeft}
	case SymRight:
		return Action{Kind: ActRight}
	case SymEnter, SymSpace:
		return Action{Kind: ActConfirm}
	case SymEsc:
		return Action{Kind: ActCancel}
	case SymPageUp:
		return Action{Kind: ActPageUp}
	case SymPageDown:
		return Action{Kind: ActPageDown}
	case SymHome:
		return Action{Kind: ActHome}
	case SymEnd:
		return Action{Kind: ActEnd}
	case SymF8:
		return Action{Kind: ActThemeCycle}
	case SymF9:
		return Action{Kind: ActMusicToggle}
	case SymF10:
		return Action{Kind: ActQuitSave}
	case SymChar:
		switch {
		case k.Ch == 'y' || k.Ch == 'Y':
			return Action{Kind: ActYes}
		case k.Ch == 'n' || k.Ch == 'N':
			return Action{Kind: ActNo}
		case k.Ch >= '0' && k.Ch <= '9':
			return Digit(k.Ch)
		case (k.Ch >= 'a' && k.Ch <= 'z') || (k.Ch >= 'A' && k.Ch <= 'Z'):
			return Letter(k.Ch)
		}
	}
	return None
}
