// openkb 桌面進入點:載入資料 → app.Game 跑畫面流程(title→選角→地圖→戰鬥/城鎮)。
package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/app"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/render"
	"github.com/wicanr2/open-king-bounty-go/internal/screen"
)

const defaultDataDir = "/home/anr2/openkb/openkb-code/data"

var (
	datadir     = flag.String("datadir", defaultDataDir, "遊戲資料目錄(cjk24.bin / free/*.ini / free/*.png)")
	startClass  = flag.Int("startclass", -1, "debug:>=0 直接以該職業建角進世界地圖(截圖驗證用)")
	startCombat = flag.Bool("startcombat", false, "debug:直接切入一場戰鬥(截圖驗證用)")
	startSelect = flag.Bool("startselect", false, "debug:直接進選角畫面(截圖驗證用)")
	startTown   = flag.Bool("starttown", false, "debug:直接進城鎮畫面(截圖驗證用)")
	townID      = flag.Int("townid", 0, "debug:配合 -starttown,指定要進哪個鎮(0-25,對應 gamestate.Town.ID)")
	worldSeed   = flag.Uint("seed", uint(gamestate.DefaultWorldSeed), "debug:配合 -starttown/-startclass,指定 salt_spells 等世界生成用的 RNG seed")
)

func rootScreen(a *kbdata.Assets) screen.Screen {
	if *startCombat {
		gs := gamestate.NewGame(a, "Sir Loin", 0, gamestate.DefaultWorldSeed)
		return screen.NewDebugCombatScreen(gs, a)
	}
	if *startTown {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		return screen.NewTownScreen(gs, a, *townID)
	}
	if *startClass >= 0 && *startClass < 4 {
		gs := gamestate.NewGame(a, "Sir Loin", *startClass, gamestate.DefaultWorldSeed)
		return screen.NewWorldMapScreen(gs, a)
	}
	if *startSelect {
		return screen.NewCharSelectScreen(a)
	}
	return screen.NewTitleScreen(a)
}

func main() {
	flag.Parse()

	assets, err := kbdata.Load(*datadir)
	if err != nil {
		log.Printf("load assets: %v(以空資料續行)", err)
	}
	if ts, err := render.LoadTileset(*datadir); err == nil {
		screen.SetTileset(ts)
	} else {
		log.Printf("load tileset: %v(地圖退回色塊)", err)
	}
	if hero, err := render.LoadSprite(*datadir, "cursor.png", 48, 34); err == nil {
		screen.SetHero(hero)
	} else {
		log.Printf("load hero: %v", err)
	}
	if sidebar, err := render.LoadSpriteOpaque(*datadir, "sidebar.png", 48, 34); err == nil {
		screen.SetSidebar(sidebar)
	} else {
		log.Printf("load sidebar: %v", err)
	}
	if coins, err := render.LoadSpriteOpaque(*datadir, "coins.png", 16, 5); err == nil {
		screen.SetCoins(coins)
	} else {
		log.Printf("load coins: %v", err)
	}
	if art, err := render.LoadPNGTileNamed(*datadir, "select-0.png"); err == nil {
		screen.SetSelectArt(art)
	} else {
		log.Printf("load select art: %v", err)
	}
	if art, err := render.LoadPNGTileNamed(*datadir, "title.png"); err == nil {
		screen.SetTitleArt(art)
	} else {
		log.Printf("load title art: %v", err)
	}
	if comtiles, err := render.LoadSpriteOpaque(*datadir, "comtiles.png", 48, 34); err == nil {
		screen.SetComtiles(comtiles)
	} else {
		log.Printf("load comtiles: %v", err)
	}
	if art, err := render.LoadPNGTileNamed(*datadir, "town.png"); err == nil {
		screen.SetLocation(art)
	} else {
		log.Printf("load town background: %v", err)
	}
	for id, name := range screen.TroopFileNames {
		sp, err := render.LoadSprite(*datadir, name+".png", 48, 34)
		if err != nil {
			log.Printf("load troop sprite %s: %v", name, err)
			continue
		}
		screen.SetTroopSprite(id, sp)
	}

	ebiten.SetWindowSize(app.LogicalW*3, app.LogicalH*3)
	ebiten.SetWindowTitle("御封戰將 (openkb-go)")
	if err := ebiten.RunGame(app.New(rootScreen(assets), assets, nil, false)); err != nil {
		log.Fatal(err)
	}
}
