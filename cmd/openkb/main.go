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
	datadir           = flag.String("datadir", defaultDataDir, "遊戲資料目錄(cjk24.bin / free/*.ini / free/*.png)")
	startClass        = flag.Int("startclass", -1, "debug:>=0 直接以該職業建角進世界地圖(截圖驗證用)")
	startCombat       = flag.Bool("startcombat", false, "debug:直接切入一場戰鬥(截圖驗證用)")
	startSelect       = flag.Bool("startselect", false, "debug:直接進選角畫面(截圖驗證用)")
	startTown         = flag.Bool("starttown", false, "debug:直接進城鎮畫面(截圖驗證用)")
	townID            = flag.Int("townid", 0, "debug:配合 -starttown,指定要進哪個鎮(0-25,對應 gamestate.Town.ID)")
	startRecruit      = flag.Bool("startrecruit", false, "debug:直接進招兵(棲地)畫面(截圖驗證用)")
	startViewArmy     = flag.Bool("startviewarmy", false, "debug:直接進軍隊檢視畫面(截圖驗證用)")
	startViewChar     = flag.Bool("startviewchar", false, "debug:直接進角色檢視畫面(截圖驗證用)")
	startViewContract = flag.Bool("startviewcontract", false, "debug:直接進懸賞契約檢視畫面(截圖驗證用)")
	contractVillain   = flag.Int("contractvillain", -1, "debug:配合 -startviewcontract,>=0 直接把 gs.Contract 設為該惡棍 id(0-16;起手預設 0xFF=無契約)")
	recruitRtype      = flag.Int("rtype", 0, "debug:配合 -startrecruit,棲地類型(0=平原 1=森林 2=山丘 3=地下城)")
	recruitTroop      = flag.Int("recruittroop", 0, "debug:配合 -startrecruit,招募兵種 TroopID")
	worldSeed         = flag.Uint("seed", uint(gamestate.DefaultWorldSeed), "debug:配合 -starttown/-startclass,指定 salt_spells 等世界生成用的 RNG seed")
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
	if *startRecruit {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		// debug 直接進招兵:世界生成(saltContinent)已把洲0棲地0撒成隨機兵種,
		// 這裡覆寫成 -recruittroop 指定值,讓截圖驗證能固定看到想測的兵種
		// (對齊 recruit_test.go 的同一手法:直接寫世界狀態,不依賴實際撒鹽結果)。
		gs.DwellingTroop[0][0] = *recruitTroop
		gs.DwellingPopulation[0][0] = 999
		return screen.NewRecruitScreen(gs, a, 0, 0, *recruitRtype)
	}
	if *startViewArmy {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		return screen.NewViewArmyScreen(gs, a)
	}
	if *startViewChar {
		class := 0
		if *startClass >= 0 && *startClass < 4 {
			class = *startClass
		}
		gs := gamestate.NewGame(a, "Sir Loin", class, uint32(*worldSeed))
		return screen.NewViewCharacterScreen(gs, a)
	}
	if *startViewContract {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		// debug 直接進懸賞契約檢視:起手 gs.Contract 恆為 0xFF(尚未接契約,對齊
		// spawn_game),-contractvillain 讓截圖驗證能固定看「有契約」畫面,不必真的
		// 走完整套「城鎮領契約」流程(那是另一案,見 docs/PORT-STATUS.md)。
		if *contractVillain >= 0 && *contractVillain < kbdata.MaxVillains {
			gs.Contract = byte(*contractVillain)
		}
		return screen.NewViewContractScreen(gs, a)
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
	// 棲地背景(招募畫面用):sub_id 0=home(cstl) 2=平原(plai) 3=森林(frst)
	// 4=山丘/洞穴(cave) 5=地下城(dngn),對齊 C DOS_location_names。
	for _, d := range []struct {
		subID int
		file  string
	}{
		{0, "cstl.png"},
		{2, "plai.png"},
		{3, "frst.png"},
		{4, "cave.png"},
		{5, "dngn.png"},
	} {
		if art, err := render.LoadPNGTileNamed(*datadir, d.file); err == nil {
			screen.SetLocationBg(d.subID, art)
		} else {
			log.Printf("load location bg %s: %v", d.file, err)
		}
	}
	for id, name := range screen.TroopFileNames {
		sp, err := render.LoadSprite(*datadir, name+".png", 48, 34)
		if err != nil {
			log.Printf("load troop sprite %s: %v", name, err)
			continue
		}
		screen.SetTroopSprite(id, sp)
	}
	// 班底立繪(view_character 用):sub_id 對照見 screen.SetPortrait 註解
	// (DOS_class_names 順序,與 GameState.Class 直接同一欄位)。
	for class, name := range [4]string{"knig", "pala", "sorc", "barb"} {
		if art, err := render.LoadPNGTileNamed(*datadir, name+".png"); err == nil {
			screen.SetPortrait(class, art)
		} else {
			log.Printf("load portrait %s: %v", name, err)
		}
	}
	if items, err := render.LoadSpriteOpaque(*datadir, "view.png", 40, 34); err == nil {
		screen.SetViewItems(items)
	} else {
		log.Printf("load view items: %v", err)
	}
	// 惡棍臉部(view_contract 用):不去背(free-data.c case GR_VILLAIN 設
	// is_transparent=0),同 sidebar.png 的載入方式。
	for id, name := range screen.VillainFileNames {
		sp, err := render.LoadSpriteOpaque(*datadir, name+".png", 48, 34)
		if err != nil {
			log.Printf("load villain face %s: %v", name, err)
			continue
		}
		screen.SetVillainFace(id, sp)
	}

	ebiten.SetWindowSize(app.LogicalW*3, app.LogicalH*3)
	ebiten.SetWindowTitle("御封戰將 (openkb-go)")
	if err := ebiten.RunGame(app.New(rootScreen(assets), assets, nil, false)); err != nil {
		log.Fatal(err)
	}
}
