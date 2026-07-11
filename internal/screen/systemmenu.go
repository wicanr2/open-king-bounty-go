// systemmenu.go -- 世界地圖專屬的「選單 dialog」(modal overlay)。
//
// 觸控排版三層化的第三層:主畫面(世界地圖)只留 D-pad + 右側 sidebar 的 view 類熱區,
// 其餘所有動作(存/讀檔、軍隊、解散、搜索、地圖、換洲、主題、作弊、音樂、標題)收進這個
// 由頂部一顆「選單」鈕叫出的 dialog——貼近 C 原版「一顆鍵開完整動作選單」的手感,也讓
// 地圖下緣不再被一排按鈕擋住。桌面鍵盤仍走原快捷鍵(不需開此 dialog)。
//
// 繪製:先重畫底層畫面(世界地圖)當背景 → 壓一層半透明暗色 → 置中畫 dialog 面板與各鈕
// (沿用觸控鈕「深藍底+亮金框+置中白字」視覺)。命中:各鈕矩形收進 Keymap.Buttons 當
// 隱形熱區(視覺由本畫面自繪,故在桌面/觸控兩種模式都正確),點面板外的全螢幕熱區關閉。
package screen

import (
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
	"github.com/wicanr2/open-king-bounty-go/internal/save"
)

// smPanel 是置中 dialog 面板矩形(320×200 邏輯座標)。
var smPanel = input.Rect{X: 48, Y: 28, W: 224, H: 136}

// SystemMenuScreen 是世界地圖的系統/動作選單 dialog(modal overlay)。
type SystemMenuScreen struct {
	parent Screen // 底層畫面(通常世界地圖),用來畫壓暗背景
	gs     *gamestate.GameState
	assets *kbdata.Assets
	msg    string // 存/讀檔結果提示,顯示在面板底
}

// NewSystemMenuScreen 建立系統選單 dialog;parent 為叫出它的底層畫面(拿來當壓暗背景)。
func NewSystemMenuScreen(parent Screen, gs *gamestate.GameState, a *kbdata.Assets) *SystemMenuScreen {
	return &SystemMenuScreen{parent: parent, gs: gs, assets: a}
}

// smItem 是 buttons() 建表用的「字母鍵 + 標籤」。
type smItem struct {
	r rune
	l string
}

// smButton 是 dialog 上一顆鈕:矩形 + 送出的 Action + 標籤。
type smButton struct {
	rect  input.Rect
	act   input.Action
	label string
}

// smRow 把 n 顆等寬鈕以 cx 為中心水平排在 y 列,回傳各矩形。
func smRow(cx, y, bw, bh, gap, n int) []input.Rect {
	total := n*bw + (n-1)*gap
	x0 := cx - total/2
	r := make([]input.Rect, n)
	for i := range r {
		r[i] = input.Rect{X: x0 + i*(bw+gap), Y: y, W: bw, H: bh}
	}
	return r
}

// buttons 依當前狀態(是否乘船)組出 dialog 各鈕。分四列:查看/動作、存讀檔(+換洲)、
// 系統(主題/作弊/音樂)、標題。
func (s *SystemMenuScreen) buttons() []smButton {
	const cx, bw, bh, gap = 160, 46, 20, 8
	var out []smButton

	// 第一列:查看/動作(軍隊/解散/搜索/地圖)。
	ra := smRow(cx, 52, bw, bh, gap, 4)
	for i, it := range []smItem{{'v', "軍隊"}, {'d', "解散"}, {'g', "搜索"}, {'m', "地圖"}} {
		out = append(out, smButton{ra[i], input.Letter(it.r), it.l})
	}

	// 第二列:存/讀檔(乘船時加「換洲」)。
	bItems := []smItem{{'s', "存檔"}, {'l', "讀檔"}}
	if s.gs != nil && s.gs.Mount == gamestate.KBMountSail {
		bItems = append(bItems, smItem{'n', "換洲"})
	}
	rb := smRow(cx, 80, bw, bh, gap, len(bItems))
	for i, it := range bItems {
		out = append(out, smButton{rb[i], input.Letter(it.r), it.l})
	}

	// 第三列:系統(主題/作弊/音樂)。這些由 app.handleSystem 攔 ActThemeCycle/ActMusicToggle
	// 全域處理(切主題/音樂,dialog 保持開啟);ActCheat 走本畫面 Update → 開作弊選單。
	rc := smRow(cx, 108, bw, bh, gap, 3)
	out = append(out,
		smButton{rc[0], input.Action{Kind: input.ActThemeCycle}, "主題"},
		smButton{rc[1], input.Action{Kind: input.ActCheat}, "作弊"},
		smButton{rc[2], input.Action{Kind: input.ActMusicToggle}, "音樂"},
	)

	// 第四列:標題(回標題)。用 Letter('q') 與 dialog 自身的 ActCancel(關閉)區隔——
	// ActCancel 只關 dialog,'q' 才真的回標題。
	out = append(out, smButton{input.Rect{X: cx - 40, Y: 136, W: 80, H: bh}, input.Letter('q'), "標題"})
	return out
}

func (s *SystemMenuScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActCancel:
		return Pop() // ESC / 點面板外 → 關閉 dialog
	case input.ActCheat:
		if s.gs != nil {
			return Replace(NewCheatMenuScreen(s.gs, s.assets))
		}
		return Stay()
	case input.ActLetter:
		switch a.Rune {
		case 's':
			s.handleSave()
			return Stay()
		case 'l':
			return s.handleLoad()
		case 'v':
			return Replace(NewViewArmyScreen(s.gs, s.assets))
		case 'd':
			return Replace(NewDismissArmyScreen(s.gs, s.assets))
		case 'g':
			return Replace(NewSearchScreen(s.gs, s.assets))
		case 'm':
			return Replace(NewMinimapScreen(s.gs, s.assets))
		case 'n':
			if s.gs != nil && s.gs.Mount == gamestate.KBMountSail {
				return Replace(NewNavigateContinentScreen(s.gs, s.assets))
			}
			return Stay()
		case 'q':
			return Replace(NewTitleScreen(s.assets))
		}
	}
	// ActThemeCycle / ActMusicToggle 已由 app.handleSystem 攔截(切主題/音樂),不會到這裡。
	return Stay()
}

// handleSave 把目前 GameState 存進 slot 0(對齊 WorldMapScreen.handleSave)。
func (s *SystemMenuScreen) handleSave() {
	if s.gs == nil {
		s.msg = "無角色資料,無法存檔"
		return
	}
	if err := save.SaveGame(s.gs, 0); err != nil {
		s.msg = "存檔失敗: " + err.Error()
		return
	}
	s.msg = "已存檔"
}

// handleLoad 讀回 slot 0;成功則以載入的 GameState 換成新的世界地圖(Replace,關掉 dialog)。
func (s *SystemMenuScreen) handleLoad() Transition {
	loaded, err := save.LoadGame(0)
	if err != nil {
		if os.IsNotExist(err) {
			s.msg = "無存檔"
		} else {
			s.msg = "讀檔失敗: " + err.Error()
		}
		return Stay()
	}
	return Replace(NewWorldMapScreen(loaded, s.assets))
}

func (s *SystemMenuScreen) Draw(dst *ebiten.Image) {
	// 底層(世界地圖)當背景,再壓一層半透明暗色。
	if s.parent != nil {
		s.parent.Draw(dst)
	} else {
		drawChromeFrame(dst)
	}
	vector.DrawFilledRect(dst, 0, 0, 320, 200, color.RGBA{0, 0, 0, 150}, false)

	// dialog 面板:深藍底 + 暗邊 + 亮金框(沿用觸控鈕/chrome 的視覺語言)。
	p := smPanel
	vector.DrawFilledRect(dst, float32(p.X), float32(p.Y), float32(p.W), float32(p.H), colorStatus, false)
	strokeRect1pxLocal(dst, float32(p.X), float32(p.Y), float32(p.W), float32(p.H), hintDark)
	strokeRect1pxLocal(dst, float32(p.X+1), float32(p.Y+1), float32(p.W-2), float32(p.H-2), colorBorder)

	var font *kbdata.CJKAtlas
	if s.assets != nil {
		font = s.assets.Font
	}
	// 標題文字置中。
	if font != nil {
		const title = "選單"
		tx := p.X + (p.W-len([]rune(title))*render.CJKCell)/2
		render.DrawText(dst, font, title, tx, p.Y+4, color.White)
	}
	// 各鈕(可見):沿用觸控鈕繪製,深藍底+亮金框+置中白字。
	var cs []input.Control
	for _, b := range s.buttons() {
		cs = append(cs, input.Control{Rect: b.rect, Action: b.act, Label: b.label})
	}
	render.DrawTouchControls(dst, cs, font)

	// 存/讀檔結果提示(面板底)。
	if s.msg != "" && font != nil {
		render.DrawText(dst, font, s.msg, p.X+8, p.Y+p.H-11, color.RGBA{240, 220, 40, 255})
	}
}

func (s *SystemMenuScreen) Keymap() input.Keymap {
	// 各鈕矩形收進 Buttons 當隱形熱區(視覺由 Draw 自繪,兩種模式都正確);面板空白吸收
	// 點擊(ActNone,不誤關);面板外全螢幕熱區排最後 → 點外即關閉(ActCancel)。
	var btns []input.Control
	for _, b := range s.buttons() {
		btns = append(btns, input.Control{Rect: b.rect, Action: b.act, Label: b.label, Hidden: true})
	}
	btns = append(btns,
		input.Control{Rect: smPanel, Action: input.None, Hidden: true},
		input.Control{Rect: input.Rect{X: 0, Y: 0, W: 320, H: 200}, Action: input.Action{Kind: input.ActCancel}, Hidden: true},
	)
	return input.Keymap{Buttons: btns}
}
