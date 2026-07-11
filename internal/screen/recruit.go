package screen

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// dwellingNames 對齊 C bounty.c dwelling_names[MAX_DWELLINGS](硬編繁中,不被 free
// 資料覆寫)。招募畫面只會走到棲地 rtype 0-3(索引 4「城堡」是玩家自家城堡,不在
// 招募流程內,故不列)。
var dwellingNames = [...]string{"平原", "森林", "山丘", "地下城"}

// RecruitScreen 是招兵(棲地)畫面,對齊 C visit_dwelling()/draw_location()/
// draw_sidebar()(game.c:2991-3093):頂列「按 'ESC' 離開」+ 棲地背景(依 rtype
// 換 plai/frst/cave/dngn)+ 迎賓兵種(即被招募的兵種本身,frame 恆 0,非動畫)+
// 右側 sidebar + 底部選單框(表頭棲地名+底線/可招募/造價+GP/上限/招募數量)。
//
// 兵種與庫存改讀真實世界狀態 gs.DwellingTroop[cont][id] / gs.DwellingPopulation[cont][id]
// (由 gamestate.NewGame 呼叫 saltContinent 生成,見 internal/gamestate/worldgen.go),
// 取代舊版 demoRecruitTroopID/demoDwellingPopulationCap 佔位。cont/dwellingID 由呼叫端
// worldmap.go 依踩到的棲地座標查出(對齊 C visit_dwelling 的線性掃描)。
type RecruitScreen struct {
	gs         *gamestate.GameState
	assets     *kbdata.Assets
	cont       int // 洲別,查 gs.DwellingTroop/DwellingPopulation 用
	dwellingID int // 棲地 id(該洲內),同上
	rtype      int // 棲地類型 0=平原 1=森林 2=山丘 3=地下城,對應 C draw_location(2+rtype,...) 與 dwelling_names[rtype]

	sel int // 目前選定的招募數量(0..cap())
	msg string
}

// NewRecruitScreen 建立招兵畫面,初始數量為 0。兵種/庫存不在建構時快取,
// 每次讀取都直接查 gs.DwellingTroop/DwellingPopulation(對齊 C 版直接查
// game->dwelling_troop/dwelling_population,不快取複本)。
func NewRecruitScreen(gs *gamestate.GameState, a *kbdata.Assets, cont, dwellingID, rtype int) *RecruitScreen {
	return &RecruitScreen{gs: gs, assets: a, cont: cont, dwellingID: dwellingID, rtype: rtype}
}

// validDwelling 回報 cont/dwellingID 是否為合法索引且 gs 已就緒(nil-safe 防呆)。
func (s *RecruitScreen) validDwelling() bool {
	return s.gs != nil &&
		s.cont >= 0 && s.cont < kbdata.MaxContinents &&
		s.dwellingID >= 0 && s.dwellingID < kbdata.MaxDwellings
}

// troopID 回傳目前棲地教的兵種 id,對應 C `game->dwelling_troop[game->continent][id]`。
func (s *RecruitScreen) troopID() int {
	if !s.validDwelling() {
		return 0
	}
	return s.gs.DwellingTroop[s.cont][s.dwellingID]
}

// population 回傳目前棲地可招募庫存,對應 C `game->dwelling_population[game->continent][id]`。
func (s *RecruitScreen) population() int {
	if !s.validDwelling() {
		return 0
	}
	return s.gs.DwellingPopulation[s.cont][s.dwellingID]
}

// troop 回傳目前招募兵種的靜態資料,索引越界或資產未載入時回傳零值(nil-safe)。
func (s *RecruitScreen) troop() kbdata.Troop {
	id := s.troopID()
	if s.assets == nil || id < 0 || id >= len(s.assets.Troops) {
		return kbdata.Troop{}
	}
	return s.assets.Troops[id]
}

// dwellingName 回傳 rtype 對應的棲地中文名,越界回傳空字串。
func (s *RecruitScreen) dwellingName() string {
	if s.rtype < 0 || s.rtype >= len(dwellingNames) {
		return ""
	}
	return dwellingNames[s.rtype]
}

// leadershipMax 回傳「你最多可招募 N 名」的數字,對應 C
// max = army_max_troop_count(game, troop_id)(game.c:3030)——純看領導力,不看金錢
// (金錢不足由 BuyTroop 回傳 ErrNotEnoughGold 另外處理,對齊 C buy_troop 分工)。
func (s *RecruitScreen) leadershipMax() int {
	id := s.troopID()
	if s.gs == nil || s.assets == nil || id < 0 || id >= len(s.assets.Troops) {
		return 0
	}
	return s.gs.ArmyMaxTroopCount(s.assets, id)
}

// cap 回傳步進器(D-pad 調數量,觸控適配,見下方 Keymap 注解)的上限:
// 領導力上限與真實棲地庫存(gs.DwellingPopulation)兩者取小,對應 C 輸入迴圈
// `if (number > max) continue;` 與 `if (number > dwelling_population) continue;`
// 兩道獨立限制。
func (s *RecruitScreen) cap() int {
	max := s.leadershipMax()
	if pop := s.population(); pop < max {
		return pop
	}
	return max
}

func (s *RecruitScreen) Update(a input.Action) Transition {
	max := s.cap()
	if s.sel > max {
		s.sel = max
	}
	switch a.Kind {
	case input.ActUp, input.ActRight:
		if s.sel < max {
			s.sel++
		}
	case input.ActDown, input.ActLeft:
		if s.sel > 0 {
			s.sel--
		}
	case input.ActConfirm:
		if s.gs == nil || s.assets == nil {
			return Stay()
		}
		if s.sel <= 0 {
			s.msg = "數量為 0,未招募"
			return Stay()
		}
		if err := s.gs.BuyTroop(s.assets, s.troopID(), s.sel); err != nil {
			switch err {
			case gamestate.ErrNotEnoughGold:
				s.msg = "你的金幣不足！" // 對齊 C KB_BottomBox("\n\n\n你的金幣不足！", ...)
			case gamestate.ErrNoSlot:
				s.msg = "沒有空的部隊欄位了！" // 對齊 C KB_BottomBox("", "沒有空的部隊欄位了！", 1)
			default:
				s.msg = err.Error()
			}
			return Stay()
		}
		// 持久扣減世界狀態的棲地庫存,對齊 C game.c:3080
		// `game->dwelling_population[game->continent][id] -= number;`(買完直接寫回
		// game 本體,不是只存在本畫面的複本——離開棲地再回來庫存仍反映扣減)。
		if s.validDwelling() {
			s.gs.DwellingPopulation[s.cont][s.dwellingID] -= s.sel
		}
		s.msg = "已招募"
		s.sel = 0
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

// drawLocation 畫棲地背景 + 迎賓兵種,對齊 C draw_location(2+rtype, troop_id, 0)
// (game.c:3040)。與 town.go 的 drawLocation 同一套幾何(troop sprite 貼在背景左緣
// 往右一個兵種幀寬、底部與背景底對齊),差異只在:
//   - 背景改依 rtype 取棲地圖(plai/frst/cave/dngn),不是固定 town.png。
//   - 迎賓兵種就是被招募的兵種本身(troopID),不是隨機兵種。
//   - frame 恆為 0(C 呼叫處第三參數固定傳 0,無走路動畫)。
//   - troop sprite 仍用 DrawFrameFlipped:draw_location 內部固定
//     SDL_TakeSurface(GR_TROOP, troop_id, 1)(flip=1),與呼叫端無關。
func (s *RecruitScreen) drawLocation(dst *ebiten.Image) {
	bg := LocationBg(2 + s.rtype)
	if bg != nil {
		render.DrawTile(dst, bg, mapX, mapY)
	}
	sp := troopSpriteFor(s.troopID())
	if sp == nil || bg == nil {
		return
	}
	bgH := bg.Bounds().Dy()
	tx := mapX + sp.FrameW*1
	ty := mapY + bgH - sp.FrameH
	sp.DrawFrameFlipped(dst, 0, tx, ty)
}

// menuLines 是底部選單框逐字對齊 C visit_dwelling()(game.c:3048-3061)的顯示文字:
//
//	"           %s\n"(11 空格 + 棲地名)+ 11 空格 + 底線(長度 = 棲地名 UTF-8 byte
//	長度,對齊 C strlen(dwelling_names[rtype]) 逐一印 "-" 的原始行為——Go 的
//	len(string) 與 C 的 strlen 一樣是 byte 數,故直接沿用不必換算)+ 空行 + 四行訊息。
//
// 空行的由來:C 版底線印完後,游標被 KB_iloc 重設回表頭那一列,再 "\n\n\n" 前進
// 三行才印「可招募」,中間空出一行未寫(視覺上的留白列),這裡逐行對應成
// lines[2]=""。
func (s *RecruitScreen) menuLines() []string {
	name := s.dwellingName()
	troop := s.troop()
	gold := 0
	if s.gs != nil {
		gold = s.gs.Gold
	}
	lines := []string{
		"           " + name,
		"           " + strings.Repeat("-", len(name)),
		"",
		fmt.Sprintf("可招募 %d 名%s", s.population(), troop.Name),
		fmt.Sprintf("每名造價=% 3d      GP=%dK", troop.GoldCost, gold/1000),
		fmt.Sprintf("你最多可招募 %d 名", s.leadershipMax()),
		fmt.Sprintf("招募數量        %d", s.sel),
	}
	// msg 是 Go 版加的提示(C 版用獨立 modal KB_BottomBox 蓋版顯示,本移植未做
	// 全螢幕 modal,先疊在選單框內當額外一行,不影響上面 7 行的逐字對齊)。
	if s.msg != "" {
		lines = append(lines, "", s.msg)
	}
	return lines
}

func (s *RecruitScreen) Draw(dst *ebiten.Image) {
	drawChromeFrame(dst)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	s.drawLocation(dst)
	drawSidebar(dst, s.gs, 0) // C: draw_sidebar(game, 0) —— 棲地畫面 tick 恆 0,無迎賓動畫
	drawBottomFrame(dst, s.assets, s.menuLines())
}

// Keymap 保留現有的數量步進器(D-pad 調 sel,對觸控友善——C 版用數字鍵盤
// text_input 直接打數字,這是給觸控用的等效替代,不逐鍵移植)。Directions:true
// 因為步進器需要 D-pad;招募的資訊都在畫面上方(頂列/背景/側欄/底部選單框),
// 不與觸控 D-pad(畫面左下)重疊,故不套用城鎮那種 LetterRects 隱形熱區。
func (s *RecruitScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: true,
		Confirm:    "招募",
		Cancel:     "離開",
	}
}
