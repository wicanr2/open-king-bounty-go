package screen

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// ChestScreen 對齊 C take_chest(game.c:3167)的顯示層:呈現寶箱結果訊息;若是金幣/
// 領導力二選一(ChestGoldChoice),提供 A) 取金幣 / B) 換領導力,選定後套用並回地圖。
// 其餘結果(海圖/寶珠/收入/法力/法術/空箱)顯示訊息,按任意鍵回地圖。
//
// TakeChest 已在建構前由呼叫端(worldmap)執行並套用非選擇類效果 + 清掉寶箱 tile;
// 本畫面只負責顯示 + 處理金幣選擇。
type ChestScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	result gamestate.ChestResult
	names  [kbdata.MaxContinents]string
	spells []gamestate.Spell
}

// NewChestScreen 建立寶箱結果畫面。
func NewChestScreen(gs *gamestate.GameState, a *kbdata.Assets, r gamestate.ChestResult) *ChestScreen {
	return &ChestScreen{
		gs:     gs,
		assets: a,
		result: r,
		names:  gamestate.ContinentNames(a),
		spells: gamestate.LoadSpells(a),
	}
}

func (s *ChestScreen) Update(a input.Action) Transition {
	if s.result.Kind == gamestate.ChestGoldChoice {
		switch a.Kind {
		case input.ActLetter:
			if a.Rune == 'a' {
				s.gs.ApplyGoldChoice(s.result, true)
				return Pop()
			}
			if a.Rune == 'b' {
				s.gs.ApplyGoldChoice(s.result, false)
				return Pop()
			}
		case input.ActCancel:
			// 未選擇離開:對齊「不強迫」——預設取金幣(C 版必選,這裡取較安全的金幣)。
			s.gs.ApplyGoldChoice(s.result, true)
			return Pop()
		}
		return Stay()
	}
	// 非選擇類:任意鍵回地圖。
	if !a.IsNone() {
		return Pop()
	}
	return Stay()
}

// lines 依結果類別組出顯示文字(逐一對齊 C take_chest 各分支的 KB_BottomBox 文字)。
func (s *ChestScreen) lines() []string {
	r := s.result
	switch r.Kind {
	case gamestate.ChestNavmap:
		name := ""
		if r.Continent >= 0 && r.Continent < len(s.names) {
			name = s.names[r.Continent]
		}
		return []string{"在一只古老的箱子裡,", "你找到了描繪通往", name, "航路的地圖與海圖。"}
	case gamestate.ChestOrb:
		return []string{"透過一顆魔法寶珠,", "你得以俯瞰整片大陸。", "你這片區域的地圖", "已經完整。"}
	case gamestate.ChestGoldChoice:
		return []string{
			"你發現了一處隱藏的寶藏。",
			"你可以:",
			fmt.Sprintf("A) 取走這 %d 金幣。", r.Gold),
			"B) 把金幣分給農民,",
			fmt.Sprintf("使你的領導力提升 %d。", r.Leadership),
		}
	case gamestate.ChestCommission:
		return []string{"你發現此地蘊藏豐富的礦產。", "", fmt.Sprintf("每週收入提升了 %d。", r.Points)}
	case gamestate.ChestSpellPower:
		return []string{"你放出一名強大的精靈,", fmt.Sprintf("法術威力提升 %d。", r.Points)}
	case gamestate.ChestMaxSpell:
		return []string{"部族薩滿傳授你魔法奧秘,", fmt.Sprintf("法術上限提升了 %d。", r.Points)}
	case gamestate.ChestNewSpell:
		name := ""
		if r.SpellType >= 0 && r.SpellType < len(s.spells) {
			name = s.spells[r.SpellType].Name
		}
		return []string{"你捕獲了一隻作亂的小惡魔,", "為換取自由,你獲得:", fmt.Sprintf("   %d 個 %s。", r.Points, name)}
	default:
		return []string{"", "這只箱子是空的!"}
	}
}

func (s *ChestScreen) Draw(dst *ebiten.Image) {
	dst.Fill(colorBorder)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	drawSidebar(dst, s.gs, 0)
	drawBottomFrame(dst, s.assets, s.lines())
}

func (s *ChestScreen) Keymap() input.Keymap {
	if s.result.Kind == gamestate.ChestGoldChoice {
		return input.Keymap{
			Directions: false, Cancel: "離開",
			Letters: []input.LetterItem{{Rune: 'a', Label: "取金幣"}, {Rune: 'b', Label: "換領導"}},
		}
	}
	return input.Keymap{Directions: false, Confirm: "繼續", Cancel: "離開"}
}
