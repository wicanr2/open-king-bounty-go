// villain_assets.go -- view_contract 用的惡棍臉部素材(GR_VILLAIN)。
//
// 獨立成檔,同 combat_assets.go 的理由:載入邏輯掛在 cmd/openkb/main.go、
// mobile/mobile.go,只是多加幾行呼叫,不動既有程式碼。
package screen

import "github.com/wicanr2/open-king-bounty-go/internal/render"

// VillainFileNames 是 VillainID(0-16)→ data/free/*.png 檔名(不含副檔名)。
// 索引順序對齊 C 版 src/lib/dos-data.c:696 `DOS_villain_names[]`(free-data.c
// case GR_VILLAIN 直接用它當檔名 stem),同時也對齊 data/free/villains.ini 各
// villainN 區塊的 file= 欄位——兩邊順序逐一核對一致(2026-07-10)。
var VillainFileNames = [...]string{
	"mury", "hack", "ammi", "baro", "drea", "cane", "mora", "barr", "barg", // 0-8
	"rina", "ragf", "mahk", "auri", "czar", "magu", "urth", "arec", // 9-16
}

// villainFaceSprites 是各惡棍的臉部 sprite,索引 = villain id,每張 192×34
// (4 幀 48×34:對齊 game.c:1721 `frame*tile->w` 取幀,frame 0-3 循環)。
// free-data.c case GR_VILLAIN 設 is_transparent=0(不去背,同 sidebar.png)。
var villainFaceSprites [len(VillainFileNames)]*render.Sprite

// SetVillainFace 由進入點在載入資產後設定第 villainID 個惡棍臉部 sprite。
func SetVillainFace(villainID int, s *render.Sprite) {
	if villainID >= 0 && villainID < len(villainFaceSprites) {
		villainFaceSprites[villainID] = s
	}
}

// VillainFace 回傳 villainID 對應的臉部 sprite(越界或未載入回 nil)。
func VillainFace(villainID int) *render.Sprite {
	if villainID < 0 || villainID >= len(villainFaceSprites) {
		return nil
	}
	return villainFaceSprites[villainID]
}
