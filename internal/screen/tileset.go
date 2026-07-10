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
