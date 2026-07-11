// openkb 桌面進入點:載入資料 → app.Game 跑畫面流程(title→選角→地圖→戰鬥/城鎮)。
package main

import (
	"flag"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/open-king-bounty-go/internal/app"
	"github.com/wicanr2/open-king-bounty-go/internal/bgm"
	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/screen"
)

const defaultDataDir = "/home/anr2/openkb/openkb-code/data"

var (
	datadir           = flag.String("datadir", defaultDataDir, "遊戲資料目錄(cjk24.bin / free/*.ini / free/*.png)")
	startClass        = flag.Int("startclass", -1, "debug:>=0 直接以該職業建角進世界地圖(截圖驗證用)")
	startCombat       = flag.Bool("startcombat", false, "debug:直接切入一場戰鬥(截圖驗證用)")
	combatShoot       = flag.Bool("combatshoot", false, "debug:配合 -startcombat,玩家首格換弓箭手並直接進射擊瞄準(驗證射擊 UI)")
	startSelect       = flag.Bool("startselect", false, "debug:直接進選角畫面(截圖驗證用)")
	startTown         = flag.Bool("starttown", false, "debug:直接進城鎮畫面(截圖驗證用)")
	townID            = flag.Int("townid", 0, "debug:配合 -starttown,指定要進哪個鎮(0-25,對應 gamestate.Town.ID)")
	startRecruit      = flag.Bool("startrecruit", false, "debug:直接進招兵(棲地)畫面(截圖驗證用)")
	startViewArmy     = flag.Bool("startviewarmy", false, "debug:直接進軍隊檢視畫面(截圖驗證用)")
	startViewChar     = flag.Bool("startviewchar", false, "debug:直接進角色檢視畫面(截圖驗證用)")
	startViewContract = flag.Bool("startviewcontract", false, "debug:直接進懸賞契約檢視畫面(截圖驗證用)")
	contractVillain   = flag.Int("contractvillain", -1, "debug:配合 -startviewcontract,>=0 直接把 gs.Contract 設為該惡棍 id(0-16;起手預設 0xFF=無契約)")
	startCastleHome   = flag.Bool("startcastlehome", false, "debug:直接進家鄉城堡(王座廳選單)畫面")
	startCastleOwn    = flag.Bool("startcastleown", false, "debug:直接進自家城堡(駐防/撤離)畫面;castle id 由 -castleid 指定")
	startCastleSiege  = flag.Bool("startcastlesiege", false, "debug:直接進敵方城堡(圍攻詢問)畫面;castle id 由 -castleid 指定")
	startRecruitSol   = flag.Bool("startrecruitsoldiers", false, "debug:直接進家鄉招兵畫面")
	startAudience     = flag.Bool("startaudience", false, "debug:直接進謁見國王畫面")
	castleID          = flag.Int("castleid", 0, "debug:配合 -startcastleown/-startcastlesiege,指定城堡 id(0-25)")
	startMinimap      = flag.Bool("startminimap", false, "debug:直接進本洲小地圖畫面")
	startPuzzle       = flag.Bool("startpuzzle", false, "debug:直接進權杖拼圖畫面(掀開部分格)")
	startCheat        = flag.Bool("startcheat", false, "debug:直接進作弊選單(截圖驗證用)")
	startMenu         = flag.Bool("startmenu", false, "debug:直接在世界地圖上疊出系統選單 dialog(截圖驗證用)")
	startEndWeek      = flag.Bool("startendofweek", false, "debug:直接進週結算畫面(第 N 週,占星+收支,截圖驗證用)")
	startLose         = flag.Bool("startlose", false, "debug:直接進時間耗盡的遊戲結束畫面(截圖驗證用)")
	startOptions      = flag.Bool("startoptions", false, "debug:直接進遊戲調整設定畫面(opt_* 選項,截圖驗證用)")
	theme             = flag.String("theme", "", "debug:強制起始美術主題(dos/genesis/amiga/free),空=預設偏好序")
	recruitRtype      = flag.Int("rtype", 0, "debug:配合 -startrecruit,棲地類型(0=平原 1=森林 2=山丘 3=地下城)")
	recruitTroop      = flag.Int("recruittroop", 0, "debug:配合 -startrecruit,招募兵種 TroopID")
	worldSeed         = flag.Uint("seed", uint(gamestate.DefaultWorldSeed), "debug:配合 -starttown/-startclass,指定 salt_spells 等世界生成用的 RNG seed")
	shotPath          = flag.String("shot", "", "debug:在第 -shotframe 幀把畫面存成此 PNG 後結束(截圖驗證用)")
	shotFrame         = flag.Int("shotframe", 3, "debug:配合 -shot,第幾幀截圖(預設 3,讓首幀狀態穩定)")
	forceTouch        = flag.Bool("touch", false, "debug:桌面也繪製觸控控制疊層(供按鈕 UX 設計/截圖驗證用)")
	toastText         = flag.String("toast", "", "debug:第一幀即顯示此 toast 提示(截圖驗證 toast 外觀用)")
)

func rootScreen(a *kbdata.Assets) screen.Screen {
	if *startCombat {
		gs := gamestate.NewGame(a, "Sir Loin", 0, gamestate.DefaultWorldSeed)
		if *combatShoot {
			return screen.NewDebugShootCombatScreen(gs, a)
		}
		return screen.NewDebugCombatScreen(gs, a)
	}
	if *startTown {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		return screen.NewTownScreen(gs, a, *townID)
	}
	if *startCastleHome {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		return screen.NewCastleHomeScreen(gs, a)
	}
	if *startMinimap {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		return screen.NewMinimapScreen(gs, a)
	}
	if *startPuzzle {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		// debug:掀開部分格子(前 4 神器 + 前 8 惡棍),讓拼圖同時看到露出的地圖與遮蓋的臉/圖示。
		for i := 0; i < 4; i++ {
			gs.ArtifactFound[i] = 1
		}
		for i := 0; i < 8; i++ {
			gs.VillainCaught[i] = 1
		}
		return screen.NewPuzzleScreen(gs, a)
	}
	if *startRecruitSol {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		return screen.NewCastleHomeScreen(gs, a) // 家鄉招兵須從王座廳 A) 進入;此旗標保留給流程截圖
	}
	if *startAudience {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		return screen.NewAudienceScreen(gs, a)
	}
	if *startCheat {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		return screen.NewCheatMenuScreen(gs, a)
	}
	if *startEndWeek {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		gs.DaysLeft = gs.MaxDays() - gamestate.WeekDays // passed 一週 → WeekID()=1(顯示「第 1 週」)
		return screen.NewEndOfWeekScreen(gs, a, 3, gs.Gold+250)
	}
	if *startLose {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		gs.DaysLeft = 0
		return screen.NewLoseScreen(gs, a)
	}
	if *startOptions {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		return screen.NewGameOptionsScreen(gs, a)
	}
	if *startCastleOwn {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		// debug:把指定城堡設為玩家所有並塞些守軍,讓「撤離/駐防」畫面有內容可看。
		if *castleID >= 0 && *castleID < kbdata.MaxCastles {
			gs.CastleOwner[*castleID] = gamestate.KBCastlePlayer
			gs.CastleTroops[*castleID] = [5]int{0, 3, 0xFF, 0xFF, 0xFF}
			gs.CastleNumbers[*castleID] = [5]int{20, 8, 0, 0, 0}
		}
		return screen.NewCastleOwnScreen(gs, a, *castleID)
	}
	if *startCastleSiege {
		gs := gamestate.NewGame(a, "Sir Loin", 0, uint32(*worldSeed))
		return screen.NewCastleSiegeScreen(gs, a, *castleID)
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
	if *startMenu {
		class := 0
		if *startClass >= 0 && *startClass < 4 {
			class = *startClass
		}
		gs := gamestate.NewGame(a, "Sir Loin", class, gamestate.DefaultWorldSeed)
		wm := screen.NewWorldMapScreen(gs, a)
		return screen.NewSystemMenuScreen(wm, gs, a) // dialog 以世界地圖為背景,供截圖驗版面
	}
	if *startClass >= 0 && *startClass < 4 {
		gs := gamestate.NewGame(a, "Sir Loin", *startClass, gamestate.DefaultWorldSeed)
		// debug:-contractvillain 可設定當前契約,讓 sidebar 契約格截圖驗證能看到
		// villain 臉(起手 Contract 恆 0xFF、sidebar 契約格空框)。
		if *contractVillain >= 0 && *contractVillain < kbdata.MaxVillains {
			gs.Contract = byte(*contractVillain)
		}
		return screen.NewWorldMapScreen(gs, a)
	}
	if *startSelect {
		return screen.NewCharSelectScreen(a)
	}
	// 正常啟動:開場商標(display_logo)→ 標題 → 選角。
	return screen.NewLogoScreen(a)
}

func main() {
	flag.Parse()

	assets, err := kbdata.Load(*datadir)
	if err != nil {
		log.Printf("load assets: %v(以空資料續行)", err)
	}
	// 美術主題(module)偏好序:DOS → Genesis → Amiga → free(沿用 C openkb 四主題)。
	// 版權主題(dos/genesis/amiga)未內建時自動缺席,退回 free。整套美術(tileset + 所有
	// sprite + UI)由 screen.LoadArt 載入(先鋪 free 當底再覆蓋);執行期 F8 / 觸控 ☰ 切換
	// (screen.CycleTheme)。詳見 docs/theme-switching-plan.md。
	if active := screen.InitThemes(os.DirFS(*datadir), []string{"dos", "genesis", "amiga", "free"}); active != "" {
		if *theme != "" && screen.SetActiveTheme(*theme) {
			active = *theme
		}
		log.Printf("art theme: %s (available: %v)", active, screen.AvailableThemes())
	} else {
		log.Printf("art theme: none found in %s", *datadir)
	}
	app.ShotPath, app.ShotFrame = *shotPath, *shotFrame // debug 截圖(-shot 空則停用)
	app.DebugToast = *toastText                         // debug:首幀顯示 toast(截圖驗證用)
	if *shotPath == "" {
		bgm.Init(os.DirFS(*datadir)) // BGM(data/music/scenes.ini;缺則靜音);headless 截圖時跳過(無 ALSA 裝置)
	}

	ebiten.SetWindowSize(app.LogicalW*3, app.LogicalH*3)
	ebiten.SetWindowTitle("御封戰將 (openkb-go)")
	if err := ebiten.RunGame(app.New(rootScreen(assets), assets, nil, *forceTouch)); err != nil {
		log.Fatal(err)
	}
}
