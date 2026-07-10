// viewcharacter.go -- 角色檢視畫面,對齊 C view_character()(game.c:1736-1854)。
package screen

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/input"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// 版面幾何(320×200 邏輯座標,zoom=0),逐字對齊 C view_character() 開頭
// (game.c:1738-1762):
//
//	portrait = SDL_TakeSurface(GR_PORTRAIT, class, 0) — GR_PORTRAIT 這個 case
//	(free-data.c:982-989)沒設 image_cutout,回傳整張 knig/pala/sorc/barb.png,
//	不切幀。四職業立繪實測皆 96×102(data/free/{knig,pala,sorc,barb}.png),故
//	直接寫死常數,不查詢執行期圖檔尺寸(與 mapTileW/H 等既有常數同慣例)。
//	pos = { left_frame.w, top_frame.h+bar_frame.h+fs.h+zoom, screen.w-left-right, 0 }
//	即 (mapX, mapY, vaPosW)——與 view_army 完全同一組幾何(viewarmy.go 已驗證
//	vaPosW = perim*mapTileW = 240),故本檔直接沿用 vaPosW,不重算。
//	C 版 view_character 全程不呼叫 draw_sidebar,右側維持底色,故本移植也不畫
//	sidebar(同 viewarmy.go 的處理)。
const (
	vcPortraitW = 96  // knig/pala/sorc/barb.png 實測寬(四職業立繪同尺寸)
	vcPortraitH = 102 // 同上,實測高

	vcBoxX = mapX + vcPortraitW   // 112:數值框左緣(立繪右側)
	vcBoxW = vaPosW - vcPortraitW // 144

	// stats.y = pos.y + fs.h/4 + fs.h/8(game.c:1760),fs.h=CJKCell=8 → +2+1=+3。
	vcStatsX = vcBoxX
	vcStatsY = mapY + render.CJKCell/4 + render.CJKCell/8 // 24

	// 分隔線(line rect,game.c:1762):x 比框左緣再往左 fs.w/8,寬度對應加回來;
	// 厚度 fs.h/8。fs.w == fs.h == CJKCell(ASCII/CJK 同格,見 render.CJKCell 註解)。
	vcSepX = vcBoxX - render.CJKCell/8 // 111
	vcSepW = vcBoxW + render.CJKCell/8 // 145
	vcSepH = render.CJKCell / 8        // 1

	// 底部道具帶(game.c:1798-1846):view.png 不設 image_cutout,game.c 自行用
	// 「pos.w/6」重切子圖寬——不是 view.png 原生 tile 寬,是畫面幾何常數,故稱
	// vcItemW 不稱 tile 寬,提醒這是版面常數不是素材常數。view.png 實測高 34px。
	vcItemW    = vaPosW / 6 // 40
	vcItemH    = 34         // view.png 實測高
	vcItemBelt = 4          // BELT:寶物列每列格數
	vcMapBelt  = 2          // MAP_BELT:地圖列每列格數

	vcMaxArtifacts  = 8                                    // MAX_ARTIFACTS(bounty.h:44)
	vcMaxContinents = 4                                    // MAX_CONTINENTS(bounty.h:39)
	vcEmptySlot     = vcMaxArtifacts + vcMaxContinents     // 12:EMPTY_SLOT
	vcEmptyMap      = vcMaxArtifacts + vcMaxContinents + 1 // 13:EMPTY_MAP

	vcInventoryX = mapX
	vcInventoryY = mapY + vcPortraitH // 123:portrait 正下方
)

// viewCharInventoryBg 對齊 C `SDL_RemapColor(sys->screen, 0xFFFF00)`(game.c:1811):
// 底部道具帶固定黃底,與 CS_VIEWCHAR 色scheme 無關,是字面常數色,不查 colors.ini。
var viewCharInventoryBg = color.RGBA{0xff, 0xff, 0x00, 0xff}

// vcGroupY 回傳第 n 組(0-4)文字區塊的起始 y,逐字對齊 C 的 KB_iloc 呼叫序列
// (game.c:1766-1791):
//
//	第 0 組(姓名/稱號 + 領導力):KB_iloc(stats.x, stats.y),無額外位移。
//	第 1-4 組(佣金金幣 / 法術威力上限 / 捕獲寶物 / 城堡陣亡分數):
//	  KB_iloc(stats.x, stats.y + (fs.h+zoom)*2n + fs.h/8),n=1..4。
//
// 巧合但逐算驗證成立的性質:第 k 組(k=0..3)結束後的分隔線 y,剛好等於
// vcGroupY(k+1)(見 game.c `line.y = cursor_y*fs.h + base_y [+fs.h/8 僅 k=0]`
// 展開後與下一組 KB_iloc 位移算式相同),故分隔線直接複用本函式,不必另外重算。
func vcGroupY(n int) int {
	if n <= 0 {
		return vcStatsY
	}
	return vcStatsY + 2*n*render.CJKCell + render.CJKCell/8
}

// ViewCharacterScreen 是角色檢視畫面,對齊 C view_character()(game.c:1736):
// 左側班底立繪(GR_PORTRAIT,依 class)+ 右側數值框(黑底,分 5 組以白色細線分隔:
// 姓名稱號+領導力 / 佣金+金幣 / 法術威力+上限 / 捕獲惡棍+尋得寶物 / 駐守城堡+
// 陣亡部眾+目前分數)+ 頂列「按 'ESC' 離開」+ 底部黃底道具帶(寶物 8 格 + 大陸
// 地圖 4 格)。純檢視畫面,唯一互動是 ESC 離開,不套 D-pad/字母選單。
//
// 世界狀態未建模的欄位(對齊 town/recruit 既有的「佔位 0 + 誠實 TODO」慣例,
// 詳見 docs/PORT-STATUS.md 世界狀態清單):
//   - player_captured(捕獲惡棍)/ player_num_artifacts(尋得寶物)/
//     player_castles(駐守城堡)/ player_score(目前分數):需
//     villain_caught[]/artifact_found[]/castle_owner[]/難度 等世界狀態,皆佔位 0。
//   - followers_killed(陣亡部眾):C 版由戰鬥損耗累積,Go combat 尚未接回
//     gamestate(見 internal/combat/damage.go:95 既有 TODO),佔位 0。
//   - 寶物帶 artifact_found[]:未建模,全畫 EMPTY_SLOT(空格)。
//   - 大陸地圖帶 continent_found[]:未建模,僅 worldmap.go 的 homeContinent
//     (玩家起始洲)視為已探索,其餘畫 EMPTY_MAP。
type ViewCharacterScreen struct {
	gs     *gamestate.GameState
	assets *kbdata.Assets
}

// NewViewCharacterScreen 建立角色檢視畫面。
func NewViewCharacterScreen(gs *gamestate.GameState, a *kbdata.Assets) *ViewCharacterScreen {
	return &ViewCharacterScreen{gs: gs, assets: a}
}

func (s *ViewCharacterScreen) Update(a input.Action) Transition {
	if a.Kind == input.ActCancel {
		return Pop()
	}
	return Stay()
}

func (s *ViewCharacterScreen) Draw(dst *ebiten.Image) {
	// 外框:同 viewarmy.go 慣例,C 版不清畫面(疊在前一畫面殘留上),本引擎每畫面
	// 全幅重繪,故先填邊框色,未被下方元素蓋到的區域就是視覺上的黃邊框。
	dst.Fill(colorBorder)

	class := 0
	if s.gs != nil {
		class = s.gs.Class
	}
	if portrait := Portrait(class); portrait != nil {
		render.DrawTile(dst, portrait, mapX, mapY)
	}

	// 數值框黑底(colors[COLOR_BACKGROUND],CS_VIEWCHAR → [character] background,
	// 對齊 viewarmy.go 已驗證的 viewArmyEmptyBg 黑色)。
	vector.DrawFilledRect(dst, float32(vcBoxX), float32(mapY), float32(vcBoxW), float32(vcPortraitH), viewArmyEmptyBg, false)

	// 4 條白色分隔線(colors[COLOR_FRAME],同 viewArmySepLine),純幾何,與
	// gs/assets 是否載入無關,故獨立於下面的文字繪製之外。
	for k := 0; k < 4; k++ {
		y := vcGroupY(k + 1)
		vector.DrawFilledRect(dst, float32(vcSepX), float32(y), float32(vcSepW), float32(vcSepH), viewArmySepLine, false)
	}

	s.drawStats(dst)

	drawTopBox(dst, s.assets, "按 'ESC' 離開")

	s.drawInventory(dst)
}

// drawStats 印右側五組文字,逐行對齊 C KB_iprintf 呼叫序列(game.c:1768-1794,
// 精確格式字串與空白數見檔頭常數註解引用的原始行)。
func (s *ViewCharacterScreen) drawStats(dst *ebiten.Image) {
	if s.gs == nil || s.assets == nil || s.assets.Font == nil {
		return
	}
	gs := s.gs
	font := s.assets.Font

	title := ""
	if gs.Class >= 0 && gs.Class < len(s.assets.Classes) && gs.Rank >= 0 && gs.Rank < len(s.assets.Classes[gs.Class]) {
		title = s.assets.Classes[gs.Class][gs.Rank].Name
	}

	draw := func(text string, y int) {
		render.DrawText(dst, font, text, vcStatsX, y, color.White)
	}

	draw(fmt.Sprintf("%s，%s", gs.Name, title), vcGroupY(0))
	draw(fmt.Sprintf("領導力           %5d", gs.Leadership), vcGroupY(0)+render.CJKCell)

	// player_commission(game) 就是 game->commission 原樣回傳(play.c:769-771),
	// 無額外加成邏輯;GameState.Commission 本身已在 acceptRank/寶箱等處累加,直接用。
	draw(fmt.Sprintf("每週佣金         %5d", gs.Commission), vcGroupY(1))
	draw(fmt.Sprintf("金幣             %5d", gs.Gold), vcGroupY(1)+render.CJKCell)

	draw(fmt.Sprintf("法術威力         %5d", gs.SpellPower), vcGroupY(2))
	draw(fmt.Sprintf("法術上限         %5d", gs.MaxSpells), vcGroupY(2)+render.CJKCell)

	// TODO(世界狀態):player_captured/player_num_artifacts 需
	// villain_caught[]/artifact_found[] 世界狀態(尚未建模),暫佔位 0。
	draw(fmt.Sprintf("捕獲惡棍         %5d", 0), vcGroupY(3))
	draw(fmt.Sprintf("尋得寶物         %5d", 0), vcGroupY(3)+render.CJKCell)

	// TODO(世界狀態):player_castles 需 castle_owner[];followers_killed 需
	// combat 把戰鬥損耗接回 gamestate(見 internal/combat/damage.go:95 既有
	// TODO);player_score 依賴以上三者加難度加權。三者暫皆佔位 0。
	draw(fmt.Sprintf("駐守城堡         %5d", 0), vcGroupY(4))
	draw(fmt.Sprintf("陣亡部眾         %5d", 0), vcGroupY(4)+render.CJKCell)
	draw(fmt.Sprintf("目前分數         %5d", 0), vcGroupY(4)+2*render.CJKCell)
}

// drawInventory 畫底部道具帶,逐字對齊 C game.c:1798-1846:黃底 + 寶物列(8格,
// 每列 4 個)+ 大陸地圖列(4格,每列 2 個),未拾得/未探索畫空格子圖。
func (s *ViewCharacterScreen) drawInventory(dst *ebiten.Image) {
	vector.DrawFilledRect(dst, float32(vcInventoryX), float32(vcInventoryY), float32(vaPosW), float32(vcItemH*2), viewCharInventoryBg, false)

	if viewItemSprite == nil {
		return
	}

	// 寶物列:artifact_found[] 未建模,全視為未拾得 → 全畫 EMPTY_SLOT。
	for i := 0; i < vcMaxArtifacts; i++ {
		col := i % vcItemBelt
		row := i / vcItemBelt
		x := vcInventoryX + col*vcItemW
		y := vcInventoryY + row*vcItemH
		viewItemSprite.DrawFrame(dst, vcEmptySlot, x, y)
	}

	// 大陸地圖列:continent_found[] 未建模,僅玩家起始洲(homeContinent,對齊
	// worldmap.go 的 homeContinent 常數)視為已探索,其餘畫 EMPTY_MAP。
	baseX := vcInventoryX + vcItemBelt*vcItemW
	for i := 0; i < vcMaxContinents; i++ {
		col := i % vcMapBelt
		row := i / vcMapBelt
		x := baseX + col*vcItemW
		y := vcInventoryY + row*vcItemH
		frame := vcEmptyMap
		if i == homeContinent {
			frame = vcMaxArtifacts + i
		}
		viewItemSprite.DrawFrame(dst, frame, x, y)
	}
}

// Keymap:純檢視,無移動、無字母選單(同 viewarmy.go)。
func (s *ViewCharacterScreen) Keymap() input.Keymap {
	return input.Keymap{
		Directions: false,
		Cancel:     "離開",
	}
}
