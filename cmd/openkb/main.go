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
)

func rootScreen(a *kbdata.Assets) screen.Screen {
	if *startCombat {
		gs := gamestate.NewGame(a, "Sir Loin", 0)
		return screen.NewDebugCombatScreen(gs, a)
	}
	if *startClass >= 0 && *startClass < 4 {
		gs := gamestate.NewGame(a, "Sir Loin", *startClass)
		return screen.NewWorldMapScreen(gs, a)
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

	ebiten.SetWindowSize(app.LogicalW*3, app.LogicalH*3)
	ebiten.SetWindowTitle("御封戰將 (openkb-go)")
	if err := ebiten.RunGame(app.New(rootScreen(assets), assets, nil)); err != nil {
		log.Fatal(err)
	}
}
