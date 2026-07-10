package screen

import (
	"fmt"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// townMode 是 TownScreen 目前顯示的子畫面。對應 C 版 game.c visit_town() 的
// msg_hold 概念:選單本身之外,某一項的結果會蓋一層訊息,直到玩家再按鍵才回選單。
type townMode int

const (
	townModeMenu        townMode = iota // 只顯示 A-E 選單
	townModeInfo                        // C) 蒐集情報:疊顯既有 GameState 數值(GOLD/LEADERSHIP)
	townModeStub                        // A/B/E:需世界狀態,先顯示「未實作」提示
	townModeLearnResult                 // D) 學習法術:顯示結果訊息(對齊 C msg_hold 三種分支)
)

// 世界狀態尚未建模的固定值,對應 C bounty.h 的常數;等 GameState 補上對應欄位
// (Boat/SiegeWeapons/POWER_CHEAPER_BOAT_RENTAL)後,這裡應改讀真值而非常數。
const (
	costSiegeWeapons  = 3000 // C COST_SIEGE_WEAPONS;GameState 尚無 SiegeWeapons 欄位,恆顯示「未購買」價格
	costBoatExpensive = 500  // C COST_BOAT_EXPENSIVE;GameState 尚無 Boat 欄位,恆走「無船」分支(POWER_CHEAPER_BOAT_RENTAL 亦未建模)
)

// townLetters 是城鎮選單的情境字母列(觸控 Keymap 用的短標籤),對應 C 版
// game.c visit_town()(2660 行起)的 five_choices:A) 領取新契約 B) 租船/取消租船
// C) 蒐集情報 D) 學習法術 E) 購買攻城武器。KB_BottomFrame 實際顯示的完整文字見
// menuLines()/infoLines()——兩者是不同關注點(前者是觸控提示短標籤,後者是逐字
// 對齊 C 的畫面文字)。
var townLetters = []input.LetterItem{
	{Rune: 'a', Label: "接任務"},
	{Rune: 'b', Label: "租船"},
	{Rune: 'c', Label: "情報"},
	{Rune: 'd', Label: "學法術"},
	{Rune: 'e', Label: "買攻城"},
}

// TownScreen 是城鎮畫面,對齊 C visit_town()/draw_location()/draw_sidebar():
// 頂部置中「按 'ESC' 離開」+ town.png 背景 + 迎賓兵種 sprite(4 幀走路動畫)+
// 右側 sidebar + 底部 A-E 選單框。完整的契約/租船/世界狀態邏輯不在本切片(需
// 世界狀態);C(情報)顯示既有 GameState 數值,其餘項先顯示「未實作」提示。
type TownScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets

	town  gamestate.Town  // 對齊 C town_names[id]/town_coords[id]
	spell gamestate.Spell // gs.TownSpell[townID] 對應的法術(名稱+售價),D 選項用;
	// 來源是 GameState.TownSpell(salt_spells 的 per-save 隨機結果),不是
	// town.SpellID(towns.ini 靜態欄,26 鎮全為 7=造橋術,已不供實際邏輯使用)。

	welcomeTroop int // 迎賓兵種(對齊 C `rand() % MAX_TROOPS`,建構時抽一次)
	drawTick     int // Draw() 被呼叫的次數,驅動 0-3 動畫幀(見 chrome.go townFrameDivisor 注解)

	sel  int      // 方向鍵目前高亮的選項(0-4,對應 townLetters 索引)
	mode townMode // 目前顯示的子畫面
	stub string   // townModeStub 時顯示的項目標籤

	learnMsg []string // townModeLearnResult 時顯示的結果訊息(見 learnSpell)
}

// NewTownScreen 建立城鎮畫面。townID 是 gamestate.Town 的索引(對應 C 版
// visit_town() 開頭「掃描 town_coords[] 找 continent/x/y 相符」得到的 id);
// 呼叫端(WorldMapScreen)已用 gamestate.FindTown 算好,這裡只需查表取用。
func NewTownScreen(gs *gamestate.GameState, a *kbdata.Assets, townID int) *TownScreen {
	ts := &TownScreen{gs: gs, assets: a}

	towns := gamestate.LoadTowns(a)
	if townID >= 0 && townID < len(towns) {
		ts.town = towns[townID]
	}
	spells := gamestate.LoadSpells(a)
	spellID := -1
	if gs != nil && townID >= 0 && townID < len(gs.TownSpell) {
		spellID = gs.TownSpell[townID]
	}
	if spellID >= 0 && spellID < len(spells) {
		ts.spell = spells[spellID]
	}
	ts.welcomeTroop = rand.Intn(len(TroopFileNames))
	return ts
}

func (s *TownScreen) Update(a input.Action) Transition {
	switch a.Kind {
	case input.ActUp:
		if s.sel > 0 {
			s.sel--
		}
	case input.ActDown:
		if s.sel < len(townLetters)-1 {
			s.sel++
		}
	case input.ActConfirm:
		return s.activate(s.sel)
	case input.ActLetter:
		if a.Rune >= 'a' && a.Rune <= 'e' {
			return s.activate(int(a.Rune - 'a'))
		}
	case input.ActCancel:
		return Pop()
	}
	return Stay()
}

// activate 觸發第 idx 個選單項(0=A..4=E),對應 C 版 key==1..5 分支。
// A(領取新契約,idx==0)接 next_contract;C(情報,idx==2)切到情報顯示;
// D(學法術,idx==3)呼叫 learnSpell;其餘項(B 租船/E 攻城)先切到「未實作」
// 提示(需世界狀態:boat/siege)。
func (s *TownScreen) activate(idx int) Transition {
	if idx < 0 || idx >= len(townLetters) {
		return Stay()
	}
	s.sel = idx
	switch townLetters[idx].Rune {
	case 'a':
		return s.takeContract()
	case 'c':
		s.mode = townModeInfo
	case 'd':
		s.learnSpell()
	default:
		s.mode = townModeStub
		s.stub = townLetters[idx].Label
	}
	return Stay()
}

// takeContract 對齊 C visit_town key==1(game.c:2753-2772):領取下一個懸賞契約。
// next_contract 回 0xFF(已無惡棍)→ 顯示提示;否則設 last_contract/contract
// 後疊出 view_contract 顯示新目標(sidebar 契約臉也會跟著更新)。
func (s *TownScreen) takeContract() Transition {
	if s.gs == nil {
		s.mode = townModeStub
		s.stub = townLetters[0].Label
		return Stay()
	}
	villainID := s.gs.NextContract()
	if villainID == 0xFF {
		s.mode = townModeLearnResult
		s.learnMsg = []string{"這片土地已無惡棍作亂……"}
		return Stay()
	}
	s.gs.LastContract = villainID
	s.gs.Contract = villainID
	return Push(NewViewContractScreen(s.gs, s.assets))
}

// learnSpell 對齊 C visit_town() key==4 分支(game.c:2807-2827):呼叫
// gamestate.GameState.LearnSpell,依結果切到對應提示文字(三種分支逐句對齊 C:
// 已達上限 / 金幣不足 / 成功後告知還可再學幾個)。
func (s *TownScreen) learnSpell() {
	s.mode = townModeLearnResult
	if s.gs == nil {
		s.learnMsg = []string{"尚未實作: " + townLetters[3].Label}
		return
	}
	err := s.gs.LearnSpell(s.spell.ID, s.spell.Gold)
	switch err {
	case gamestate.ErrSpellLimitReached:
		s.learnMsg = []string{"你已學會", "上限數量的法術。"}
	case gamestate.ErrNotEnoughGold:
		s.learnMsg = []string{"你的金幣不足！"}
	case nil:
		left := s.gs.MaxSpells - s.gs.KnownSpells()
		s.learnMsg = []string{fmt.Sprintf("你還可以再學習 %d 個法術。", left)}
	default:
		s.learnMsg = []string{"學習失敗"}
	}
}

func (s *TownScreen) Draw(dst *ebiten.Image) {
	frame := (s.drawTick / townFrameDivisor) % 4
	s.drawTick++

	// 外框:整畫面先清成黃(對齊世界地圖 worldmap.go 的 colorBorder 註解 ——
	// C 版所有畫面共用同一套 top/left/right/bottom frame 黃色裝飾邊,不是世界地圖
	// 專屬);頂列/背景/側欄/底部選單框接著蓋掉內部,四周留白就是視覺邊框。
	dst.Fill(colorBorder)
	drawTopBox(dst, s.assets, "按 'ESC' 離開")
	s.drawLocation(dst, frame)
	drawSidebar(dst, s.gs, frame)

	switch s.mode {
	case townModeInfo:
		drawBottomFrame(dst, s.assets, s.infoLines())
	case townModeStub:
		drawBottomFrame(dst, s.assets, []string{"尚未實作: " + s.stub})
	case townModeLearnResult:
		drawBottomFrame(dst, s.assets, s.learnMsg)
	default:
		drawBottomFrame(dst, s.assets, s.menuLines())
	}
}

// drawLocation 畫城鎮背景 + 迎賓兵種,對齊 C draw_location(1, random_troop, frame)
// (game.c:1086):town.png(GR_LOCATION sub_id=1,整張貼圖不裁切)貼在
// (left_frame.w, top_frame.h+bar_frame.h+fs.h+zoom) = (mapX, mapY)。迎賓兵種
// sprite 貼在背景左緣往右一個兵種幀寬、底部與背景底對齊(C:
// tpos.x = troop_frame.w*1,RECT_Bottom(&tpos,bg) 再 RECT_AddPos(&tpos,&pos))。
// C 版對 GR_TROOP 用 flip=1(SDL_TakeSurface(GR_TROOP,troop_id,1)),故這裡用
// DrawFrameFlipped 對齊(與戰鬥畫面 side 1 共用同一份鏡射邏輯)。
func (s *TownScreen) drawLocation(dst *ebiten.Image, frame int) {
	if townBackground != nil {
		render.DrawTile(dst, townBackground, mapX, mapY)
	}
	sp := troopSpriteFor(s.welcomeTroop)
	if sp == nil || townBackground == nil {
		return
	}
	bgH := townBackground.Bounds().Dy()
	tx := mapX + sp.FrameW*1
	ty := mapY + bgH - sp.FrameH
	sp.DrawFrameFlipped(dst, frame, tx, ty)
}

// menuLines 是底部選單框逐字對齊 C visit_town() 的顯示文字(game.c:2699-2721):
//
//	"%s 鄉鎮\n" + "                    GP=%dK\n"(20 空格縮排)+ 五個選項行。
//
// 世界狀態不足處(船/攻城武器擁有與否)維持保守預設,見 costBoatExpensive/
// costSiegeWeapons 常數註解。
func (s *TownScreen) menuLines() []string {
	gold := 0
	if s.gs != nil {
		gold = s.gs.Gold
	}
	return []string{
		fmt.Sprintf("%s 鄉鎮", s.town.Name),
		fmt.Sprintf("                    GP=%dK", gold/1000),
		"A) 領取新契約",
		fmt.Sprintf("B) 租船 (%d 週) ", costBoatExpensive),
		"C) 蒐集情報",
		fmt.Sprintf("D) 學習 %s (%d)", s.spell.Name, s.spell.Gold),
		fmt.Sprintf("E) 購買攻城武器 (%d)", costSiegeWeapons),
	}
}

// infoLines 是 C(蒐集情報)按下後顯示的內容。C 版 gather_information() 顯示
// 附近城堡統治者/駐軍(需 castle_owner/castle_numbers 世界狀態,尚未建模);
// 這裡先顯示既有 GameState 數值當佔位,讓選單流程走通。
func (s *TownScreen) infoLines() []string {
	gold, leadership := 0, 0
	if s.gs != nil {
		gold = s.gs.Gold
		leadership = s.gs.Leadership
	}
	return []string{
		fmt.Sprintf("%s 鄉鎮", s.town.Name),
		"",
		"情報",
		fmt.Sprintf("GOLD %d", gold),
		fmt.Sprintf("LEADERSHIP %d", leadership),
	}
}

func (s *TownScreen) Keymap() input.Keymap {
	// 城鎮是選單畫面,不需要移動 D-pad(Directions:false),避免觸控 D-pad 蓋住
	// 底部 A-E 選單文字;字母選項的觸控熱區直接疊在選單各行原位(letterRects),
	// 由選單文字本身當視覺、點該行即選該項,對齊 C 選單的乾淨外觀。
	return input.Keymap{
		Directions:  false,
		Confirm:     "選擇",
		Cancel:      "離開",
		Letters:     townLetters,
		LetterRects: townLetterRects(),
	}
}

// townLetterRects 回傳 A-E 五個選項各自在底部選單框的文字行矩形(觸控熱區)。
// A-E 對應 menuLines() 的索引 2..6,行位置與 drawBottomFrame 的排版一致
// (第 n 行 y = bottomHeaderY + n*render.CJKCell)。
func townLetterRects() []input.Rect {
	rects := make([]input.Rect, len(townLetters))
	for i := range townLetters {
		line := 2 + i // menuLines 索引:0=鎮名 1=GP 2=A … 6=E
		rects[i] = input.Rect{
			X: bottomTextX,
			Y: bottomHeaderY + line*render.CJKCell,
			W: bottomBorderW - 2*render.CJKCell,
			H: render.CJKCell,
		}
	}
	return rects
}
