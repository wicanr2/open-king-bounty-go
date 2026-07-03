package screen

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// classNames 是四職業的起手頭銜(rank 0),索引 = class。
var classNames = []string{"騎士", "遊俠", "女巫師", "蠻俠"}

// selectArt 尺寸與置中位移(對齊 C select_game:RECT_Center 把 288×184 的
// select-0.png 置中在 320×200 螢幕 → 邊距 x=16,y=8,與世界地圖同一套 ui.ini
// 邊框規格)。A/B/C/D 立繪標籤已畫進美術本身,不必另外疊字。
const (
	selectArtW = 288
	selectArtH = 184
	selectArtX = (320 - selectArtW) / 2 // 16
	selectArtY = (200 - selectArtH) / 2 // 8
)

// CharSelectScreen 是選角畫面:A–D 選職業 → 建角 → 進遊戲。
type CharSelectScreen struct {
	assets *kbdata.Assets
	sel    int // 目前高亮的職業(方向鍵移動)
}

func NewCharSelectScreen(a *kbdata.Assets) *CharSelectScreen {
	return &CharSelectScreen{assets: a}
}

func (s *CharSelectScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActUp:
		if s.sel > 0 {
			s.sel--
		}
	case input.ActDown:
		if s.sel < len(classNames)-1 {
			s.sel++
		}
	case input.ActLetter:
		// a–d 直接選對應職業
		if a.Rune >= 'a' && a.Rune <= 'd' {
			return s.start(int(a.Rune - 'a'))
		}
	case input.ActConfirm:
		return s.start(s.sel)
	case input.ActCancel:
		return Replace(NewTitleScreen(s.assets))
	}
	return Stay()
}

// start 用選定職業建角,進世界地圖。
func (s *CharSelectScreen) start(class int) Transition {
	gs := gamestate.NewGame(s.assets, "Sir Loin", class)
	return Replace(NewWorldMapScreen(gs, s.assets))
}

// Draw 對齊 C select_game(game.c:736):整畫面先清黃框底(見 worldmap.go 的
// colorBorder 註解,四周沒蓋到的部分就是視覺邊框)→ select-0.png 置中(288×184,
// A/B/C/D 立繪標籤已內建在美術裡)→ 頂部一行狀態字「選角 A-D 或 L-讀取存檔」。
// C 版這行字沒有底色 FillRect,是直接印在美術圖上,靠 render.DrawText 的雙層
// 合成(4 方位黑外框)在任何底色上都清楚,所以這裡也不必畫底色。
func (s *CharSelectScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	if selectArt != nil {
		render.DrawTile(dst, selectArt, selectArtX, selectArtY)
	}
	if s.assets != nil && s.assets.Font != nil {
		render.DrawText(dst, s.assets.Font, "選角 A-D 或 L-讀取存檔", selectArtX, selectArtY, color.White)
	}
}

func (s *CharSelectScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: true,
		Confirm:    "選擇",
		Cancel:     "返回",
		Letters: []input.LetterItem{
			{Rune: 'a', Label: "騎士"}, {Rune: 'b', Label: "遊俠"},
			{Rune: 'c', Label: "女巫師"}, {Rune: 'd', Label: "蠻俠"},
		},
	}
}
