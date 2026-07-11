package screen

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/combat"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
	"github.com/wicanr2/open-king-bounty-go/internal/save"
)

// findDwellingID 在 gs.DwellingCoords[cont] 找座標 (x,y) 對應的棲地 id,對齊 C
// game.c:3006-3010 visit_dwelling() 的線性掃描(不 break,取最後一個相符的 i——
// 座標理論上唯一,結果與有無 break 無差異,這裡忠實照抄不 break)。找不到回傳 -1
// (對齊 C `if (id == -1) return 0;`——呼叫端應維持原地不動)。
func findDwellingID(gs *gamestate.GameState, cont, x, y int) int {
	id := -1
	for i := 0; i < kbdata.MaxDwellings; i++ {
		if gs.DwellingCoords[cont][i][0] == x && gs.DwellingCoords[cont][i][1] == y {
			id = i
		}
	}
	return id
}

// findFoeID 在 gs.FoeCoords[cont] 找座標 (x,y) 對應的 foe id,對齊 C
// game.c:3711-3720 attack_foe() 的線性掃描(找到就 break)。找不到回傳 0
// (對齊 C 預設值 `int id = 0;`)。
func findFoeID(gs *gamestate.GameState, cont, x, y int) int {
	id := 0
	for i := 0; i < kbdata.MaxFoes; i++ {
		if gs.FoeCoords[cont][i][0] == x && gs.FoeCoords[cont][i][1] == y {
			id = i
			break
		}
	}
	return id
}

// foeSquadsFrom 把世界狀態的 foe_troops/foe_numbers(5 格,對齊 C game->foe_troops/
// foe_numbers[cont][id][5])轉成 combat 套件要的 [MaxUnits]Squad,數量為 0 的格
// 填 TroopID=255(空格哨兵值,對齊 gamestate.Squad 慣例;C 版本身沒有這個哨兵值,
// 是 combat.PrepareUnitsFoe/gamestate.Squad 這層轉換介面的既有慣例,見 setup.go)。
//
// 誠實標註(對齊 worldgen.go saltContinent 註解):id < FRIENDLY_FOES(=5,見
// kbdata.MaxFoes 相關常數與 C bounty.h FRIENDLY_FOES)的「友善」foe 在世界生成時
// 刻意未呼叫 repopulateFoe(C 原始碼此處的 populate_foe 呼叫本身被註解掉),故這 5 個
// 槽位在真正的 C 版遊戲中應觸發 game.c attack_foe() 的 accept_foe()「免費入隊」分支,
// 而非戰鬥;本次世界生成移植只還原資料佈置,尚未實作 accept_foe 分支,故目前踩到
// 這些槽位仍會走一般戰鬥流程,只是遇到的會是「全空」的敵軍(與先前 placeholderFoe
// 佔位相比,至少不再是寫死的野狼,但這個特例本身是延續 C 既有的未完成功能,
// 不是本次移植引入的新行為)。
func foeSquadsFrom(troops, numbers [5]int) [combat.MaxUnits]gamestate.Squad {
	var f [combat.MaxUnits]gamestate.Squad
	for i := 0; i < combat.MaxUnits && i < 5; i++ {
		if numbers[i] <= 0 {
			f[i] = gamestate.Squad{TroopID: 255}
			continue
		}
		f[i] = gamestate.Squad{TroopID: troops[i], Count: numbers[i]}
	}
	return f
}

// 版面對齊 C 版 openkb(data/free/ui.ini [map]/[tile] + game.c draw_map):
//
//	地圖 viewport 16,21 寬240高170;每 tile 48×34 → 5×5 格;玩家永遠在正中格。
//	Y 軸翻轉:遊戲 y 向上為北,螢幕上方 = 高 y(C: pos.y=(perim-1-j)*h+mapY)。
const (
	mapTileW = 48
	mapTileH = 34
	mapX     = 16
	mapY     = 21
	perim    = 5 // 240/48 = 170/34 = 5(5×5 viewport)
	radii    = 2 // (perim-1)/2:玩家置中,每邊 2 格

	// 玩家起點,對齊 C spawn_game(play.c:403-405):
	//   continent = special_coords[SP_HOME][0]
	//   x         = special_coords[SP_HOME][1]
	//   y         = special_coords[SP_HOME][2] - 2
	// ★ 注意:bounty.c 的 special_coords 是 DOS 地圖硬編值(11,7),但 free 模組會用
	//   data/free/land.ini [special0]「國王的居所」覆寫(game.c:299-301 KB_Resolve),
	//   free 真值 = (continent 0, x 15, y 10)。本移植是 free 版,故用 free 的家座標。
	//   px/py 與 C game->x/y 同一座標系(Y-flip 只在渲染,不影響資料座標),直接用。
	// TODO: 改成解析 land.ini [special0] 載入(現先硬編 free 值,免另建 parser)。
	homeContinent = 0  // land.ini [special0] continent
	homeStartX    = 15 // land.ini [special0] x
	homeStartY    = 8  // land.ini [special0] y(10) - 2
)

// 頂部狀態列(對齊 C update_ui_frames:local.status.x/y = left_frame.w/top_frame.h,
// local.status.w = screen.w - left_frame.w - right_frame.w,h = fs.h + zoom;
// data/free/ui.ini [top] h=8、[left]/[right] w=16,故 statusY=8、statusW=320-16-16=288)。
const (
	statusX = mapX // 16,與地圖左緣同(left_frame.w)
	statusY = 8
	statusW = 288
	statusH = 8
)

// 邊框與狀態列色(對齊 C data/free/colors.ini:frame1/frame2 = #ffff55、background = #0000aa)。
//
// ⚠ 修正(2026-07-11):舊版誤以為「C 把整個畫面清成 COLOR_FRAME(金)再蓋內容」,
// 於是 dst.Fill(colorBorder) 把整片填成亮黃 → 進遊戲滿版黃(實測黃佔比 ~19%,C 版只
// ~9%)。實際上 C 的世界地圖**從不清整片為金**:draw_map/draw_sidebar 只畫各自區域,
// 外圍保持 SDL 初始的**黑底**;金色只是「烤進 sidebar/狀態列圖 + 各處 SDL_FillRect
// (COLOR_FRAME) 細框線」的窄框帶(實測 C 最外圈=黑、金框帶約 6-7px)。故 Go 版改為
// 「背景填黑 + 只在 play area 外圈畫 worldFrameBand 寬的金框帶」,內容隨後蓋掉內部,
// 元件間隙自然露出金色分隔線 —— 這才對齊 C 的結構(見 game.c draw_map/draw_sidebar)。
var (
	colorBorder = color.RGBA{0xff, 0xff, 0x55, 0xff}
	colorStatus = color.RGBA{0x00, 0x00, 0xaa, 0xff}
	colorOuter  = color.RGBA{0x00, 0x00, 0x00, 0xff} // 外圍黑(對齊 C 未清區)
)

// 世界地圖「內容區」邊界(320×200 邏輯座標):狀態列/地圖/側欄的外接矩形。
// 左=mapX(16);上=statusY(8);右=側欄右緣(mapX+perim*mapTileW=256 起、側欄圖 48 寬 → 304);
// 下=地圖底(mapY+perim*mapTileH=191)。金框帶畫在此矩形外圈 worldFrameBand 寬,再往外是黑。
const (
	worldContentL  = mapX                       // 16
	worldContentT  = statusY                    // 8
	worldContentR  = mapX + perim*mapTileW + 48 // 304(側欄圖寬 48)
	worldContentB  = mapY + perim*mapTileH      // 191
	worldFrameBand = 4                          // 金框帶寬(對齊 C 實測金框 ~4-6px;取 4 收窄亮黃面積)
)

// sidebar 五個資訊面板的矩形(邏輯座標),對齊 chrome.go drawSidebar 的堆疊:
// sx = mapX+perim*mapTileW = 256、y=mapY(21)起每 mapTileH(34)一格、每格 48×34。
// 世界地圖把「點面板 = 用對應功能」的隱形觸控熱區疊在其中幾個面板上(語意「點什麼看什麼」),
// 由 Keymap() 提供給 TouchLayout 當 LetterRects,並由 drawSidebarTouchHints 畫可點提示。
var (
	sidebarPanelContract = input.Rect{X: 256, Y: 21, W: 48, H: 34}  // 幀8 契約框 → 檢視契約(k)
	sidebarPanelSiege    = input.Rect{X: 256, Y: 55, W: 48, H: 34}  // 幀9 攻城武器(暫無對應畫面,純資訊)
	sidebarPanelMagic    = input.Rect{X: 256, Y: 89, W: 48, H: 34}  // 幀10 魔法星 → 施法(z,僅會魔法時)
	sidebarPanelPuzzle   = input.Rect{X: 256, Y: 123, W: 48, H: 34} // 幀11 拼圖地圖 → 看拼圖(p)
	sidebarPanelPurse    = input.Rect{X: 256, Y: 157, W: 48, H: 34} // 幀12 錢袋金幣 → 看角色/金幣(c)
)

// WorldMapScreen 是世界地圖:顯示 land.org 的 tile(暫以色塊分類)、玩家可走動、踩城鎮進城。
type WorldMapScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
	msg    string // 存讀檔提示(「已存檔/已讀檔/無存檔」等),不計時,下次動作前持續顯示

	drawTick int // Draw() 每次呼叫遞增(對齊 chrome.go drawSidebar 的 frame 節奏注解)
}

// NewWorldMapScreen 建立地圖畫面。玩家位置(洲/座標)以 GameState 為準——新遊戲由
// NewGame 設為家鄉起點,讀檔則沿用存檔中的位置,戰敗 temp_death 也已寫回 gs;本畫面
// 只讀 gs.Continent/X/Y,不再自己維護一份座標,故存讀檔/傳送回家都自動正確。
func NewWorldMapScreen(gs *gamestate.GameState, a *kbdata.Assets) *WorldMapScreen {
	return &WorldMapScreen{gs: gs, assets: a}
}

// walkable 回報玩家能否踏上該 tile(水/深水/山阻擋;陸地與互動格可走)。
func walkable(tile byte) bool {
	return !kbdata.IsWater(tile) && !kbdata.IsDeepWater(tile) && !kbdata.IsRock(tile)
}

func (s *WorldMapScreen) Update(a input.Action) Transition {
	// 補顯示待處理的週結算:踩城鎮/城堡等 Push 返回後、或上一步在開闊地跨週界時設下。
	// 每步至多跨一個週界,故單一 pending 即可;任意輸入觸發時先消化再處理該輸入。
	if s.gs != nil && s.gs.PendingWeek {
		s.gs.PendingWeek = false
		return Push(NewEndOfWeekScreen(s.gs, s.assets, s.gs.PendingWeekCreature, s.gs.PendingWeekGold))
	}
	// 觸控頂部「選單」鈕 → 疊上系統/動作選單 dialog(存讀檔/軍隊/解散/搜索/地圖/換洲/
	// 主題/作弊/音樂/標題 都在裡面)。桌面鍵盤不需經此 dialog,仍走下面各快捷鍵分支。
	if a.Kind == input.ActMenu {
		if s.gs != nil {
			return Push(NewSystemMenuScreen(s, s.gs, s.assets))
		}
		return Stay()
	}
	// 存讀檔:'s' 存進 slot 0,'l' 讀 slot 0。無關乎世界地圖資產是否載入,故放在
	// nil 檢查之前;結果(成功/失敗)寫進 s.msg,由 Draw 顯示。
	if a.Kind == input.ActLetter {
		switch a.Rune {
		case 's':
			s.handleSave()
			return Stay()
		case 'l':
			return s.handleLoad()
		case 'v':
			// 檢視軍隊(對齊 C combat_options_menu / 世界地圖選單「檢視軍隊」項):
			// 疊上 ViewArmyScreen,ESC 離開後回地圖(Push,非 Replace)。
			return Push(NewViewArmyScreen(s.gs, s.assets))
		case 'c':
			// 檢視角色(對齊 C combat_options_menu / 世界地圖選單「檢視角色」項):
			// 疊上 ViewCharacterScreen,ESC 離開後回地圖(Push,非 Replace)。
			return Push(NewViewCharacterScreen(s.gs, s.assets))
		case 'n':
			// 換洲航行(對齊 C KEY_ACT(NEW_CONTINENT),game.c:6768):只有乘船
			// (Mount==SAIL)時可用,疊上換洲選單。
			if s.gs != nil && s.gs.Mount == gamestate.KBMountSail {
				return Push(NewNavigateContinentScreen(s.gs, s.assets))
			}
			return Stay()
		case 'd':
			// 解散部隊(對齊 C KEY_ACT(DISMISS_ARMY),game.c:6787)。
			return Push(NewDismissArmyScreen(s.gs, s.assets))
		case 'z':
			// 施放冒險法術(對齊 C KEY_ACT(SPELL)→choose_spell(game,NULL))。
			// 不會魔法者 C 顯示 no_spell_banner 後返回;這裡直接不開選單。
			if s.gs != nil && s.gs.KnowsMagic {
				return Push(NewChooseSpellScreen(s.gs, s.assets))
			}
			return Stay()
		case 'g':
			// 搜索當地(對齊 C KEY_ACT(SEARCH)→ask_search):站在權杖埋藏處搜索即破關。
			return Push(NewSearchScreen(s.gs, s.assets))
		case 'm':
			// 本洲小地圖(對齊 C KEY_ACT(VIEW_MAP)→view_minimap)。
			return Push(NewMinimapScreen(s.gs, s.assets))
		case 'p':
			// 權杖拼圖(對齊 C KEY_ACT(VIEW_PUZZLE)→view_puzzle)。
			return Push(NewPuzzleScreen(s.gs, s.assets))
		case 'k':
			// 檢視懸賞契約(對齊 C view_contract)。觸控由點右側 sidebar 的「契約」面板觸發
			// (見 Keymap 的 LetterRects 疊層);無契約時 ViewContractScreen 自行顯示提示。
			return Push(NewViewContractScreen(s.gs, s.assets))
		}
	}

	// 作弊 / debug 選單(對齊 C KEY_ACT(CHEAT) → debug_cheat_menu,game.c:6796):
	// 世界地圖按 F12 或觸控右上「作弊」鈕開啟。war==NULL 語境,'w' = win_game。
	if a.Kind == input.ActCheat {
		if s.gs != nil {
			return Push(NewCheatMenuScreen(s.gs, s.assets))
		}
		return Stay()
	}

	if s.assets == nil || s.gs == nil || s.gs.WorldMap == nil {
		if a.Kind == input.ActCancel {
			return Replace(NewTitleScreen(s.assets))
		}
		return Stay()
	}
	cont := s.gs.Continent
	nx, ny := s.gs.X, s.gs.Y
	switch a.Kind {
	case input.ActUp:
		ny++ // Y 翻轉:向上 = 北 = 高 y(對齊 C 版螢幕座標)
	case input.ActDown:
		ny--
	case input.ActLeft:
		nx--
	case input.ActRight:
		nx++
	case input.ActCancel:
		return Replace(NewTitleScreen(s.assets))
	default:
		return Stay()
	}
	if nx < 0 || ny < 0 || nx >= kbdata.LevelW || ny >= kbdata.LevelH {
		return Stay()
	}
	tile := s.gs.WorldMap.Tile(cont, nx, ny)
	// 登船/航行/上岸(對齊 C map 移動的 boat 段,game.c:6837-6957):
	//   - 目標是玩家船隻停靠格 → 登船(mount=SAIL)。
	//   - 水域:只有乘船(或正登船)才走得過去,否則擋下。
	//   - 陸地:沿用既有 walkable()(擋水/深水/岩;其餘可走)。
	// 移動後:乘船時船隨玩家;若踏上非水非橋(靠岸)則下船(mount=RIDE),船留在上一格水域。
	boarding := s.gs.HasBoat() && int(s.gs.Boat) == cont && s.gs.BoatX == nx && s.gs.BoatY == ny
	if kbdata.IsWater(tile) || kbdata.IsDeepWater(tile) {
		if s.gs.Mount != gamestate.KBMountSail && !boarding {
			return Stay() // 無船不能下水
		}
	} else if !walkable(tile) {
		return Stay()
	}
	lastX, lastY := s.gs.X, s.gs.Y
	s.gs.X, s.gs.Y = nx, ny
	s.gs.Reveal(cont, nx, ny) // 走到新格即揭示周邊(fog-of-war,供 view_minimap)
	if boarding {
		s.gs.Mount = gamestate.KBMountSail
	}
	if s.gs.Mount == gamestate.KBMountSail {
		s.gs.BoatX, s.gs.BoatY = nx, ny // 船隨玩家(對齊 C「tuck the boat with us」)
		if !kbdata.IsWater(tile) && !kbdata.IsDeepWater(tile) && !kbdata.IsBridge(tile) {
			// 靠岸下船(對齊 C「Hitting shore / Leave ship」),船留在上一格水域。
			s.gs.Mount = gamestate.KBMountRide
			s.gs.BoatX, s.gs.BoatY = lastX, lastY
		}
	}
	// 移動耗一步(對齊 C game.c:6945-6969):停時優先、否則扣步;步數耗盡過一天,跨週界
	// (每 WeekDays 天)觸發週結算,天數歸零則時間耗盡失敗。週結算先跑 EndWeek(世界增長)
	// 得本週生物,再以 PendingWeek 排入顯示——若本步同時踩到城鎮/城堡等(下方 Push),
	// 週結算畫面留待該畫面關閉、回世界地圖時由 Update 開頭補顯示;否則在本函式末尾即時顯示。
	beforeGold := s.gs.Gold
	if _, weekEnded := s.gs.SpendStep(); weekEnded {
		creature := s.gs.EndWeek(s.assets, kbrng.NewGlibc(uint32(s.gs.PassedDays()+1)))
		s.gs.PendingWeek = true
		s.gs.PendingWeekCreature = creature
		s.gs.PendingWeekGold = beforeGold
	}
	if s.gs.DaysLeft <= 0 {
		return Push(NewLoseScreen(s.gs, s.assets)) // 時間耗盡:遊戲結束(優先於一切互動)
	}
	// 踩到城鎮 → 進城(疊上 TownScreen,離開後回地圖)
	if tile == kbdata.TileTown {
		townID := 0
		if s.assets != nil {
			towns := gamestate.LoadTowns(s.assets)
			if id := gamestate.FindTown(towns, cont, nx, ny); id >= 0 {
				townID = id
			}
		}
		if townID >= 0 && townID < len(s.gs.TownVisited) {
			s.gs.TownVisited[townID] = 1 // 對齊 C game->town_visited[id]=1(visit_town 開頭)
		}
		return Push(NewTownScreen(s.gs, s.assets, townID))
	}
	// 踩到城堡 → 依 CastleAt 判定 home/your/enemy 分派(對齊 C visit_castle,
	// game.c:2567):家鄉城堡 → 招兵/謁見選單;自家城堡 → 駐防/撤離;敵方城堡 →
	// 圍攻詢問。進 your/enemy 時記下 castle_visited(供 select_gate 傳送清單)。
	if tile == kbdata.TileCastle || kbdata.IsCastle(tile) {
		kind, id := s.gs.CastleAt(s.assets, cont, nx, ny)
		switch kind {
		case gamestate.CastleHome:
			return Push(NewCastleHomeScreen(s.gs, s.assets))
		case gamestate.CastleYours:
			s.gs.CastleVisited[id] = 1
			return Push(NewCastleOwnScreen(s.gs, s.assets, id))
		case gamestate.CastleEnemy:
			s.gs.CastleVisited[id] = 1
			return Push(NewCastleSiegeScreen(s.gs, s.assets, id))
		default:
			// CastleAt 找不到城堡(座標未對上 land.ini/castles.ini)——對齊 C
			// KB_errlog 後 return,不進任何城堡畫面,留在原地。
			return Stay()
		}
	}
	// 踩到敵人 → 進戰鬥(疊上 CombatScreen,結束後回地圖),敵方部隊取自世界狀態
	// gs.FoeTroops/FoeNumbers(對齊 C attack_foe() 依座標查 foe id 再讀 foe_troops/numbers)。
	if tile == kbdata.TileFoe {
		foeID := findFoeID(s.gs, cont, nx, ny)
		foe := foeSquadsFrom(s.gs.FoeTroops[cont][foeID], s.gs.FoeNumbers[cont][foeID])
		// 戰後回寫:存活敵軍寫回 FoeTroops/Numbers,戰勝清 (nx,ny) tile(對齊 C run_combat mode 0)。
		return Push(NewCombatScreenFoe(s.gs, s.assets, foe, cont, foeID, nx, ny))
	}
	// 踩到路標 → 讀告示(對齊 C read_signpost,game.c:3095):疊上 SignpostScreen。
	if tile == kbdata.TileSignpost {
		return Push(NewSignpostScreen(s.gs, s.assets))
	}
	// 踩到神器 → 拾取(對齊 C take_artifact,game.c:3332):num = tile - TileArtifact1
	// (本洲第 0 或第 1 個神器),套用能力後疊上 ArtifactScreen 顯示描述。
	if tile == kbdata.TileArtifact1 || tile == kbdata.TileArtifact2 {
		pickup := s.gs.TakeArtifact(s.assets, int(tile-kbdata.TileArtifact1))
		return Push(NewArtifactScreen(s.gs, s.assets, pickup))
	}
	// 踩到寶箱 → 開箱(對齊 C take_chest,game.c:3167):判定海圖/寶珠/寶藏並套用,
	// 疊上 ChestScreen 顯示結果(金幣類讓玩家選 A/B)。TakeChest 內已清掉寶箱 tile。
	if tile == kbdata.TileChest {
		rng := kbrng.NewGlibc(uint32(cont*1000 + nx*kbdata.LevelH + ny + 1))
		result := s.gs.TakeChest(s.assets, rng)
		return Push(NewChestScreen(s.gs, s.assets, result))
	}
	// 踩到 0x8E 格(TileTelecave 與 TileDwelling3 同值):先當傳送洞試瞬移(對齊 C
	// visit_telecave force=1,game.c:6880),座標中則瞬移到成對另一端、留在地圖;
	// 不中(其實是地下城棲地)則往下走一般棲地招兵流程。
	if tile == kbdata.TileTelecave {
		if s.gs.VisitTelecave() {
			return Stay()
		}
	}
	// 踩到棲地 → 招兵(疊上 RecruitScreen,離開後回地圖)。rtype 對齊 C
	// game.c:6915 visit_dwelling(game, m - TILE_DWELLING_1):TileDwelling1..4 依序
	// 對應 rtype 0=平原 1=森林 2=山丘 3=地下城。棲地 id 依座標查 gs.DwellingCoords
	// (對齊 C visit_dwelling 的線性掃描),RecruitScreen 內部再依 (cont,id) 讀
	// gs.DwellingTroop/DwellingPopulation(真實世界狀態,取代舊佔位)。
	if tile >= kbdata.TileDwelling1 && tile <= kbdata.TileDwelling4 {
		// 對齊 C visit_dwelling 開頭:先比對魔法密室座標(special1),中則進魔法密室
		// 而非一般招兵(game.c:2999)。
		if s.gs.AtAlcove(s.assets) {
			return Push(NewAlcoveScreen(s.gs, s.assets))
		}
		rtype := int(tile - kbdata.TileDwelling1)
		dwellingID := findDwellingID(s.gs, cont, nx, ny)
		if dwellingID == -1 {
			// 對齊 C `if (id == -1) return 0;`:理論上不該發生(tile 本身就是
			// dwelling),防呆保留在原地不進招兵畫面。
			return Stay()
		}
		return Push(NewRecruitScreen(s.gs, s.assets, cont, dwellingID, rtype))
	}
	// 開闊地移動:若本步剛跨過週界(上面設了 PendingWeek),即時顯示週結算(不必等下次輸入)。
	if s.gs.PendingWeek {
		s.gs.PendingWeek = false
		return Push(NewEndOfWeekScreen(s.gs, s.assets, s.gs.PendingWeekCreature, s.gs.PendingWeekGold))
	}
	return Stay()
}

// handleSave 把目前 GameState 存進 slot 0,結果寫進 s.msg。
func (s *WorldMapScreen) handleSave() {
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

// handleLoad 讀回 slot 0;成功則以載入的 GameState 換掉本畫面(Replace),
// 失敗(含存檔不存在)只更新提示、留在原地圖。
func (s *WorldMapScreen) handleLoad() Transition {
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

// tileColor 依 tile 分類回色塊顏色(P2 暫代美術)。
func tileColor(tile byte) color.Color {
	switch {
	case tile == kbdata.TileTown:
		return color.RGBA{240, 220, 40, 255} // 城鎮:黃
	case tile == kbdata.TileCastle || kbdata.IsCastle(tile):
		return color.RGBA{230, 230, 230, 255} // 城堡:白
	case tile == kbdata.TileChest:
		return color.RGBA{230, 140, 30, 255} // 寶箱:橙
	case tile == kbdata.TileFoe:
		return color.RGBA{210, 40, 40, 255} // 敵人:紅
	case tile >= kbdata.TileDwelling1 && tile <= kbdata.TileDwelling4:
		return color.RGBA{170, 60, 200, 255} // 棲地:紫
	case kbdata.IsBridge(tile):
		return color.RGBA{150, 100, 50, 255} // 橋:褐
	case kbdata.IsDeepWater(tile):
		return color.RGBA{10, 30, 110, 255}
	case kbdata.IsWater(tile):
		return color.RGBA{30, 70, 200, 255}
	case kbdata.IsTree(tile):
		return color.RGBA{0, 100, 0, 255}
	case kbdata.IsDesert(tile):
		return color.RGBA{205, 180, 130, 255}
	case kbdata.IsRock(tile):
		return color.RGBA{120, 120, 120, 255}
	case kbdata.IsGrass(tile):
		return color.RGBA{40, 150, 50, 255}
	default:
		return color.RGBA{90, 90, 90, 255}
	}
}

func (s *WorldMapScreen) Draw(dst *ebiten.Image) {
	if s.assets == nil || s.gs == nil || s.gs.WorldMap == nil {
		ebitenutil.DebugPrint(dst, "world map: land.org 未載入")
		return
	}
	// 背景:填黑 + play area 外圈金框帶(取代舊的整片填黃,見 colorBorder 註解);
	// 地圖/狀態列/側欄隨後蓋掉內部。與所有 chrome 畫面共用同一 drawChromeFrame。
	drawChromeFrame(dst)
	// 5×5 viewport,玩家置中,Y 翻轉(對齊 C 版 draw_map:pos.y=(perim-1-j)*h+mapY)。
	borderX := s.gs.X - radii
	borderY := s.gs.Y - radii
	for j := 0; j < perim; j++ {
		for i := 0; i < perim; i++ {
			mx, my := borderX+i, borderY+j
			sx := mapX + i*mapTileW
			sy := mapY + (perim-1-j)*mapTileH
			tile := byte(kbdata.TileDeepWater) // 界外 = 深水(對齊 C 邊界填色)
			if mx >= 0 && my >= 0 && mx < kbdata.LevelW && my < kbdata.LevelH {
				tile = s.gs.WorldMap.Tile(s.gs.Continent, mx, my)
			}
			if worldTileset != nil {
				worldTileset.DrawTileAt(dst, tile, sx, sy)
			} else {
				vector.DrawFilledRect(dst, float32(sx), float32(sy), mapTileW, mapTileH, tileColor(tile), false)
			}
		}
	}
	// 主角 sprite 置中(對齊 C draw_player:hsrc.x = tile_w*(mount+frame))。
	// 幀 = 座騎值:KBMOUNT_RIDE=8 騎馬、KBMOUNT_SAIL=0 乘船、KBMOUNT_FLY=4 飛行;
	// 乘船時自動顯示船隻 sprite。無 sprite 退回青框。
	heroX := mapX + radii*mapTileW
	heroY := mapY + (perim-1-radii)*mapTileH
	if heroSprite != nil {
		heroSprite.DrawFrame(dst, s.gs.Mount, heroX, heroY)
	} else {
		vector.StrokeRect(dst, float32(heroX)+2, float32(heroY)+2, mapTileW-4, mapTileH-4, 2, color.RGBA{0, 230, 230, 255}, false)
	}
	s.drawStatusBar(dst)
	frame := (s.drawTick / townFrameDivisor) % 4
	s.drawTick++
	drawSidebar(dst, s.gs, frame)
	s.drawSidebarTouchHints(dst) // 觸控模式:在可點面板角落畫金色摺角提示
	if s.msg != "" && s.assets.Font != nil {
		render.DrawText(dst, s.assets.Font, s.msg, mapX, mapY+perim*mapTileH-2, color.RGBA{240, 220, 40, 255})
	}
}

// drawStatusBar 畫頂部狀態列(對齊 C draw_statusbar → KB_TopBox(0, " 選項 / 操作說明 /
// 剩餘天數:%d ", game->days_left)):藍底(colorStatus)+ 白字。cjk24.bin 的 atlas 其實
// 連 ASCII(數字/標點/空白)都收錄成 24×24 glyph(與 CJK 同格),故整串直接走
// render.DrawText 一次印完即可,不必混用 ebitenutil 的內建字(那套字級高與 8px
// 狀態列的 CJKCell 對不上,會整段往下溢出)。
func (s *WorldMapScreen) drawStatusBar(dst *ebiten.Image) {
	vector.DrawFilledRect(dst, statusX, statusY, statusW, statusH, colorStatus, false)
	if s.assets == nil || s.assets.Font == nil {
		return
	}
	// 「選項 / 操作說明」原是照抄 C 的靜態標籤、Go 版不可互動 → 移除(觸控改由左側
	// 「選單」鈕入口);只留有意義的「剩餘天數」,靠右畫,把左段讓給選單鈕。
	days := 0
	if s.gs != nil {
		days = s.gs.DaysLeft
	}
	text := fmt.Sprintf("剩餘天數:%d ", days)
	tw := len([]rune(text)) * render.CJKCell
	render.DrawText(dst, s.assets.Font, text, statusX+statusW-tw, statusY, color.White)
}

// drawCoins 依 C draw_sidebar 的金幣堆疊邏輯(game.c:1119+ cval 計算):三欄分別代表
// 金幣萬/千/百位數,每欄依數字堆疊對應張數的硬幣(coins.png 3 幀 16×5),由底往上疊。
// C 版原始位移是 2*sys->zoom(桌面 320×200/zoom=0 時實際只疊在同一位置);這裡固定
// 抓 2px 位移,讓多枚硬幣在畫面上能看出堆疊,屬簡化近似。
func drawCoins(dst *ebiten.Image, x, y, gold int) {
	if coinsSprite == nil {
		return
	}
	cval := [3]int{gold / 10000, (gold % 10000) / 1000, (gold % 1000) / 100}
	coinW, coinH := coinsSprite.FrameW, coinsSprite.FrameH
	for col, cnt := range cval {
		dy := y + mapTileH - coinH
		for i := 0; i < cnt; i++ {
			coinsSprite.DrawFrame(dst, col, x+col*coinW, dy)
			dy -= 2
		}
	}
}

// touchHints 為 true 時,世界地圖在「可觸控」的 sidebar 面板角落畫輕量金色摺角標記,
// 提示玩家該面板可點(僅觸控/行動模式開;桌面鍵盤模式關,維持與 C 一致的畫面)。
// 由 app.New 依 -touch / 行動裝置設定(見 internal/app/game.go 呼叫 SetTouchHints)。
var touchHints bool

// SetTouchHints 設定是否在可點的 sidebar 面板畫觸控提示標記(app 層依 touch 模式設定)。
func SetTouchHints(on bool) { touchHints = on }

var (
	hintFill = color.RGBA{0x00, 0x00, 0x8c, 0xe6} // 觸控鈕同款深藍(在金色面板框上才顯眼)
	hintGold = color.RGBA{0xe6, 0xcf, 0x00, 0xff} // 對齊 frame 主金,與觸控鈕金框一致
	hintDark = color.RGBA{0x00, 0x00, 0x00, 0xcc}
)

// drawSidebarTouchHints 在可觸控的 sidebar 面板右上角畫小「可點徽章」(深藍底+金框,
// 沿用觸控鈕的視覺語言),提示該面板可點。只畫角落一小塊、不蓋面板中央的資訊圖,維持
// C 版 sidebar 外觀。sidebar 面板本身已是金色框,故徽章用深藍底才能在金框上分離出來。
// 魔法星面板僅會魔法時提示。
func (s *WorldMapScreen) drawSidebarTouchHints(dst *ebiten.Image) {
	if !touchHints {
		return
	}
	panels := []input.Rect{sidebarPanelContract, sidebarPanelPuzzle, sidebarPanelPurse}
	if s.gs != nil && s.gs.KnowsMagic {
		panels = append(panels, sidebarPanelMagic)
	}
	for _, p := range panels {
		drawTapBadge(dst, p)
	}
}

// drawTapBadge 在矩形右上角畫一個小徽章(深藍底 + 1px 暗邊 + 1px 金框 + 中央金點),
// 當「此面板可觸控」的提示,與觸控鈕的深藍底/亮金框語言一致。
func drawTapBadge(dst *ebiten.Image, r input.Rect) {
	const bw, bh = 12, 8
	x := float32(r.X + r.W - bw - 1) // 右上角內縮 1px,避開面板金框
	y := float32(r.Y + 1)
	vector.DrawFilledRect(dst, x, y, bw, bh, hintFill, false)             // 深藍底
	strokeRect1pxLocal(dst, x, y, bw, bh, hintDark)                       // 外 1px 暗邊
	strokeRect1pxLocal(dst, x+1, y+1, bw-2, bh-2, hintGold)               // 內 1px 金框
	vector.DrawFilledRect(dst, x+bw/2-1, y+bh/2-1, 2, 2, hintGold, false) // 中央金點,遠看是「可點」記號
}

// strokeRect1pxLocal 用 4 條 1px filled rect 畫脆利邊框(同 render.strokeRect1px,
// screen 套件內自備一份免跨套件匯出)。
func strokeRect1pxLocal(dst *ebiten.Image, x, y, w, h float32, c color.RGBA) {
	vector.DrawFilledRect(dst, x, y, w, 1, c, false)
	vector.DrawFilledRect(dst, x, y+h-1, w, 1, c, false)
	vector.DrawFilledRect(dst, x, y, 1, h, c, false)
	vector.DrawFilledRect(dst, x+w-1, y, 1, h, c, false)
}

func (s *WorldMapScreen) Keymap() input.Keymap {
	// 觸控三層佈局:
	//   ① 主畫面下方只留 D-pad(Directions),不再有整排功能鈕。
	//   ② 右側 sidebar 的 view 類功能疊成隱形熱區(點面板即看對應內容):錢袋→角色(c)、
	//      拼圖→拼圖(p)、契約→檢視契約(k)、魔法星→施法(z,僅會魔法時)。LetterRects 給
	//      對應面板矩形 → TouchLayout 設 Hidden;drawSidebarTouchHints 在角落畫可點徽章。
	//   ③ 其餘動作(存讀檔/軍隊/解散/搜索/地圖/換洲/主題/作弊/音樂/標題)全收進由頂部
	//      「選單」鈕(ActMenu)叫出的 SystemMenuScreen dialog——貼近 C 原版「一鍵開選單」。
	// 桌面鍵盤仍可直接按 s/l/v/c/d/g/m/p/z/n/k/ESC(見 Update),不受本觸控佈局影響。
	letters := []input.LetterItem{
		{Rune: 'c', Label: "角色"},
		{Rune: 'p', Label: "拼圖"},
		{Rune: 'k', Label: "契約"},
	}
	rects := []input.Rect{sidebarPanelPurse, sidebarPanelPuzzle, sidebarPanelContract}
	if s.gs != nil && s.gs.KnowsMagic {
		letters = append(letters, input.LetterItem{Rune: 'z', Label: "施法"})
		rects = append(rects, sidebarPanelMagic)
	}
	// 頂部「選單」叫出鈕:移到狀態列左段(原「選項/操作說明」的位置,已移除),
	// 剩餘天數改靠右;兩者不重疊。
	menuBtn := input.Control{
		Rect:   input.Rect{X: 16, Y: 1, W: 60, H: 14},
		Action: input.Action{Kind: input.ActMenu},
		Label:  "選單",
	}
	// 常駐「搜索」鈕:搜索(找權杖埋藏處)是高頻操作,依使用者要求放主畫面免開選單。
	// 底部中央偏右,避開左下 D-pad(x≤68)與右側 sidebar(x≥256)。
	searchBtn := input.Control{
		Rect:   input.Rect{X: 200, Y: 176, W: 52, H: 20},
		Action: input.Action{Kind: input.ActLetter, Rune: 'g'},
		Label:  "搜索",
	}
	return input.Keymap{
		Directions:  true,
		Letters:     letters,
		LetterRects: rects,
		Buttons:     []input.Control{menuBtn, searchBtn},
	}
}
