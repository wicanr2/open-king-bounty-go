package screen

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/render"
)

// worldTileset 是全域載入一次的世界地圖 tileset(cmd 啟動時 SetTileset 設入)。
// 為 nil 時 WorldMapScreen 退回色塊繪製(讓無資產的測試/環境仍可運作)。
var worldTileset *render.Tileset

// SetTileset 由進入點(cmd)在載入資產後設定世界 tileset。
func SetTileset(t *render.Tileset) { worldTileset = t }

// heroSprite 是世界地圖上的主角 sprite(cursor.png,12 幀 48×34)。
var heroSprite *render.Sprite

// SetHero 由進入點在載入資產後設定主角 sprite。
func SetHero(s *render.Sprite) { heroSprite = s }

// sidebarSprite 是世界地圖右側資訊欄的 UI 素材(sidebar.png,13 幀 48×34,不去背)。
var sidebarSprite *render.Sprite

// SetSidebar 由進入點在載入資產後設定側欄 sprite。
func SetSidebar(s *render.Sprite) { sidebarSprite = s }

// coinsSprite 是側欄錢袋上疊加的金幣圖(coins.png,3 幀 16×5,不去背)。
var coinsSprite *render.Sprite

// SetCoins 由進入點在載入資產後設定金幣 sprite。
func SetCoins(s *render.Sprite) { coinsSprite = s }

// pieceSprite 是側欄拼圖框上的遮片(piece.png,單幀 9×6,不去背):蓋在尚未由「捕獲惡棍
// /尋得神器」掀開的 5×5 拼圖格上(對齊 C GR_PIECE,draw_sidebar game.c:1171)。
var pieceSprite *render.Sprite

// SetPiece 由進入點在載入資產後設定拼圖遮片 sprite。
func SetPiece(s *render.Sprite) { pieceSprite = s }

// endTiles 是破關動畫(display_cartoon)用的三張 48×34 tile:endpic-2=草地、endpic-3=橋、
// endpic-4=英雄(對齊 C GR_ENDTILE,subid 0/1/2)。缺檔時 CartoonScreen 直接跳過動畫。
var endTiles [3]*ebiten.Image

// SetEndTile 由進入點在載入資產後設定第 i 張破關動畫 tile(0=草/1=橋/2=英雄)。
func SetEndTile(i int, img *ebiten.Image) {
	if i >= 0 && i < len(endTiles) {
		endTiles[i] = img
	}
}

// EndTile 回傳第 i 張破關動畫 tile(缺則 nil)。
func EndTile(i int) *ebiten.Image {
	if i >= 0 && i < len(endTiles) {
		return endTiles[i]
	}
	return nil
}

// selectArt 是選角畫面的整張背景美術(select-0.png,288×184,不透明,已內建
// A/B/C/D 立繪標籤,對齊 C 版 GR_SELECT 幀 0)。
var selectArt *ebiten.Image

// SetSelectArt 由進入點在載入資產後設定選角背景圖。
func SetSelectArt(img *ebiten.Image) { selectArt = img }

// townBackground 是城鎮畫面的整張背景美術(town.png,240×102,不透明,對齊 C 版
// GR_LOCATION sub_id=1;free-data.c 該 case 不設 image_cutout,是整檔直接當一張
// 圖用,不是逐幀 sprite sheet)。視覺上緊貼世界地圖 viewport 同一位置(mapX,mapY)。
var townBackground *ebiten.Image

// SetLocation 由進入點在載入資產後設定城鎮(GR_LOCATION)背景圖。
// 同時寫入 locationBg[1],讓 town 也能透過 LocationBg(1) 取得(兩者指同一張圖)。
func SetLocation(img *ebiten.Image) {
	townBackground = img
	locationBg[1] = img
}

// locationBg 是各地點(GR_LOCATION)背景圖,索引對齊 C DOS_location_names 的
// sub_id:0=home(cstl) 1=town(town,亦見 townBackground/SetLocation) 2=平原(plai)
// 3=森林(frst) 4=山丘/洞穴(cave) 5=地下城(dngn)。招募畫面(recruit.go)依
// dwelling rtype(0-3)取 locationBg[2+rtype]。
var locationBg [6]*ebiten.Image

// SetLocationBg 由進入點在載入資產後設定指定 sub_id 的地點背景圖。
func SetLocationBg(subID int, img *ebiten.Image) {
	if subID < 0 || subID >= len(locationBg) {
		return
	}
	locationBg[subID] = img
}

// LocationBg 回傳指定 sub_id 的地點背景圖,可能為 nil(素材未載入),呼叫端須 nil-safe。
func LocationBg(subID int) *ebiten.Image {
	if subID < 0 || subID >= len(locationBg) {
		return nil
	}
	return locationBg[subID]
}

// titleArt 是開場標題整張美術(title.png,320×200 全螢幕,對齊 C 版 GR_TITLE /
// display_title:黑底貼滿此圖,按任意鍵進選角)。
var titleArt *ebiten.Image

// SetTitleArt 由進入點在載入資產後設定標題圖。
func SetTitleArt(img *ebiten.Image) { titleArt = img }

// logoArt 是開場 NWC 商標圖(nwcp.png,對齊 C 版 GR_LOGO / display_logo:黑底置中)。
var logoArt *ebiten.Image

// SetLogoArt 由進入點在載入資產後設定商標圖。
func SetLogoArt(img *ebiten.Image) { logoArt = img }

// frameArt 是「金框」邊框圖(frame.png,320×200,play area 透明 + 四周金色浮雕邊,
// 對齊 C 版 DOS UI 金框浮雕:黑外緣→暗金→亮金→深藍內線)。由 drawChromeFrame 疊在
// play area 外圈,取代舊的平面金框帶。版權主題(DOS)才內建;free/公開版無此檔 →
// frameArt==nil → drawChromeFrame 退回平面金框帶。
var frameArt *ebiten.Image

// SetFrameArt 由載入器設定金框圖(nil = 該主題無 frame.png,退回平面框)。
func SetFrameArt(img *ebiten.Image) { frameArt = img }

// portraitArt 是四職業班底立繪(GR_PORTRAIT,knig/pala/sorc/barb.png,96×102,
// 不去背),索引直接對齊 C DOS_class_names[class] 與 GameState.Class 同一欄位
// (free-data.c:982 case GR_PORTRAIT 用 game->class 原樣當 sub_id,charselect.go
// classNames 的職業順序與此一致,不必重排)。view_character 畫面使用。
var portraitArt [4]*ebiten.Image

// SetPortrait 由進入點在載入資產後設定第 class 個班底立繪(0-3,越界忽略)。
func SetPortrait(class int, img *ebiten.Image) {
	if class < 0 || class >= len(portraitArt) {
		return
	}
	portraitArt[class] = img
}

// Portrait 回傳第 class 個班底立繪,可能為 nil(素材未載入),呼叫端須 nil-safe。
func Portrait(class int) *ebiten.Image {
	if class < 0 || class >= len(portraitArt) {
		return nil
	}
	return portraitArt[class]
}

// viewItemSprite 是角色檢視畫面底部道具帶素材(GR_VIEW,view.png,不去背)。
// 幀寬固定 40px:對齊 view_character 版面常數 pos.w/6(見 viewcharacter.go 檔頭
// 註解),不是 view.png 原生 tile 寬——C 版對這張圖不設 image_cutout,是用畫面
// 幾何常數重切;幀高 34px = view.png 原生高度。
var viewItemSprite *render.Sprite

// SetViewItems 由進入點在載入資產後設定 view.png sprite。
func SetViewItems(s *render.Sprite) { viewItemSprite = s }
