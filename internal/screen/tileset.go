package screen

import "github.com/wicanr2/open-king-bounty-go/internal/render"

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
