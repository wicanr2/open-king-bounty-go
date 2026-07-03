// combat_assets.go -- 戰鬥畫面(CombatScreen)專用的素材全域變數。
//
// 獨立成檔(不塞進 tileset.go)是刻意的:避免跟同時在動 tileset.go/worldmap.go
// 的其他工作衝突。載入邏輯掛在 cmd/openkb/main.go、mobile/mobile.go,只是多加
// 幾行呼叫,不動既有程式碼。
package screen

import "github.com/wicanr2/open-king-bounty-go/internal/render"

// comtilesSprite 是戰鬥棋盤地形/障礙圖(data/free/comtiles.png,15 幀 48×34,不透明)。
// 對齊 C draw_combat(game.c:1298):war->omap[y][x] 的值直接當 comtiles 幀索引
// (0 = 平地,1-3 = combat/turn.go 隨機出的障礙外觀)。
var comtilesSprite *render.Sprite

// SetComtiles 由進入點在載入資產後設定戰鬥地形 sprite。
func SetComtiles(s *render.Sprite) { comtilesSprite = s }

// troopSprites 是各兵種的戰鬥 sprite,索引 = TroopID,每張 192×34(4 幀 48×34)。
// 對齊 C draw_combat:SDL_TakeSurface(GR_TROOP, troop_id, 1) 讀取,src 用
// frame*48 選欄。C 版用 flip=1 把同一張圖水平鏡射疊成第二排給 side 1 用
// (兩側共用同一張圖,只差朝向);Go 版不預先做鏡射圖,改在畫的當下對 side 1
// 呼叫 Sprite.DrawFrameFlipped,省一份資料。
var troopSprites [len(TroopFileNames)]*render.Sprite

// SetTroopSprite 由進入點在載入資產後設定第 troopID 兵種的戰鬥 sprite。
func SetTroopSprite(troopID int, s *render.Sprite) {
	if troopID >= 0 && troopID < len(troopSprites) {
		troopSprites[troopID] = s
	}
}

// troopSpriteFor 回傳 troopID 對應的戰鬥 sprite(越界或未載入回 nil)。
func troopSpriteFor(troopID int) *render.Sprite {
	if troopID < 0 || troopID >= len(troopSprites) {
		return nil
	}
	return troopSprites[troopID]
}

// TroopFileNames 是 TroopID → data/free/*.png 檔名(不含副檔名)。索引順序對齊
// internal/kbdata/tables_troops.go 的 troopTable()(移植自 C 版 bounty.c troops[]
// 固定順序),同時也對齊 data/free/troops.ini 的 [troopN] 區塊與其 file= 欄位——
// 兩邊順序一致,故用同一個索引查兩邊都對得上,不需要另外建對照表。
//
// 兩個已知例外(troops.ini 本身的資料缺口/筆誤,這裡直接修正,不照搬):
//   - troop7(牛獸人):troops.ini 沒有 file= 欄位(free 美術包沒收錄專屬圖),
//     借用同尺寸(192×34)但未被其他兵種使用的 orcs.png 頂替,避免戰鬥畫面開天窗。
//   - troop17(洞穴巨人):troops.ini 寫 file=troll,但實際檔名少一個 l,是 trol.png;
//     這裡用真正存在的檔名。
var TroopFileNames = [...]string{
	"peas", "spri", "mili", "wolf", "skel", "zomb", "gnom", "orcs", // 0-7
	"arcr", "elfs", "pike", "noma", "dwar", "ghos", "kght", "ogre", // 8-15
	"brbn", "trol", "cavl", "drui", "arcm", "vamp", "gian", "demo", // 16-23
	"drag", // 24
}
