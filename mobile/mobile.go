// Package mobile is the Android entry point for ebitenmobile bind.
// NOTE: doc comments on EXPORTED symbols in this package are copied verbatim into
// the gobind-generated Java and compiled with US-ASCII, so they must stay ASCII-only
// (CJK here breaks `ebitenmobile bind` with "unmappable character"). Non-exported
// comments may use CJK. Build: ebitenmobile bind -target android -o openkb.aar ./mobile.
package mobile

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2/mobile"

	"github.com/wicanr2/open-king-bounty-go/internal/app"
	"github.com/wicanr2/open-king-bounty-go/internal/embedded"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
	"github.com/wicanr2/open-king-bounty-go/internal/screen"
)

func init() {
	// 資料直接從嵌入的 embed.FS 讀(不解壓到檔案系統——Android 的 cwd/temp 唯讀或不可寫)。
	// LoadFS 是純資料(不碰 ebiten),可在 init() 安全執行。
	assets, err := kbdata.LoadFS(embedded.FS())
	if err != nil {
		log.Printf("load assets: %v", err)
	}

	// tileset 會建 ebiten.Image → 必須延後到遊戲迴圈第一幀(JVM/繪圖就緒),
	// 否則在 init() 會 panic「devicescale: no current JVM」。交給 app 的 boot 回呼。
	boot := func() {
		if ts, err := render.LoadTilesetFS(embedded.FS()); err == nil {
			screen.SetTileset(ts)
		} else {
			log.Printf("load tileset: %v", err)
		}
		if hero, err := render.LoadSpriteFS(embedded.FS(), "free/cursor.png", 48, 34); err == nil {
			screen.SetHero(hero)
		} else {
			log.Printf("load hero: %v", err)
		}
		if sidebar, err := render.LoadSpriteOpaqueFS(embedded.FS(), "free/sidebar.png", 48, 34); err == nil {
			screen.SetSidebar(sidebar)
		} else {
			log.Printf("load sidebar: %v", err)
		}
		if coins, err := render.LoadSpriteOpaqueFS(embedded.FS(), "free/coins.png", 16, 5); err == nil {
			screen.SetCoins(coins)
		} else {
			log.Printf("load coins: %v", err)
		}
		if art, err := render.LoadPNGTileFS(embedded.FS(), "free/select-0.png"); err == nil {
			screen.SetSelectArt(art)
		} else {
			log.Printf("load select art: %v", err)
		}
		if art, err := render.LoadPNGTileFS(embedded.FS(), "free/title.png"); err == nil {
			screen.SetTitleArt(art)
		} else {
			log.Printf("load title art: %v", err)
		}
		if art, err := render.LoadPNGTileFS(embedded.FS(), "free/nwcp.png"); err == nil {
			screen.SetLogoArt(art)
		} else {
			log.Printf("load logo art: %v", err)
		}
		if comtiles, err := render.LoadSpriteOpaqueFS(embedded.FS(), "free/comtiles.png", 48, 34); err == nil {
			screen.SetComtiles(comtiles)
		} else {
			log.Printf("load comtiles: %v", err)
		}
		if art, err := render.LoadPNGTileFS(embedded.FS(), "free/town.png"); err == nil {
			screen.SetLocation(art)
		} else {
			log.Printf("load town background: %v", err)
		}
		// 棲地背景(招募畫面用),同桌面版 main.go 的 sub_id 對照。
		for _, d := range []struct {
			subID int
			file  string
		}{
			{0, "free/cstl.png"},
			{2, "free/plai.png"},
			{3, "free/frst.png"},
			{4, "free/cave.png"},
			{5, "free/dngn.png"},
		} {
			if art, err := render.LoadPNGTileFS(embedded.FS(), d.file); err == nil {
				screen.SetLocationBg(d.subID, art)
			} else {
				log.Printf("load location bg %s: %v", d.file, err)
			}
		}
		for id, name := range screen.TroopFileNames {
			sp, err := render.LoadSpriteFS(embedded.FS(), "free/"+name+".png", 48, 34)
			if err != nil {
				log.Printf("load troop sprite %s: %v", name, err)
				continue
			}
			screen.SetTroopSprite(id, sp)
		}
		// 班底立繪(view_character 用),同桌面版 main.go 的 sub_id 對照。
		for class, name := range [4]string{"knig", "pala", "sorc", "barb"} {
			if art, err := render.LoadPNGTileFS(embedded.FS(), "free/"+name+".png"); err == nil {
				screen.SetPortrait(class, art)
			} else {
				log.Printf("load portrait %s: %v", name, err)
			}
		}
		if items, err := render.LoadSpriteOpaqueFS(embedded.FS(), "free/view.png", 40, 34); err == nil {
			screen.SetViewItems(items)
		} else {
			log.Printf("load view items: %v", err)
		}
		// 惡棍臉部(view_contract 用),同桌面版 main.go 的載入方式。
		for id, name := range screen.VillainFileNames {
			sp, err := render.LoadSpriteOpaqueFS(embedded.FS(), "free/"+name+".png", 48, 34)
			if err != nil {
				log.Printf("load villain face %s: %v", name, err)
				continue
			}
			screen.SetVillainFace(id, sp)
		}
	}

	// SetGame 讓 ebitenmobile 生成的 Java/Kotlin 綁定驅動這顆遊戲迴圈。
	mobile.SetGame(app.New(screen.NewLogoScreen(assets), assets, boot, true))
}

// Dummy is a placeholder so gomobile bind sees at least one exported symbol.
// The real entry point is init() (mobile.SetGame). ASCII-only on purpose (see package doc).
func Dummy() {}
