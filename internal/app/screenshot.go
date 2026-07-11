// screenshot.go — debug 用的可重跑截圖 harness:以 -shot 指定路徑時,遊戲在第
// ShotFrame 幀把輸出畫面(含 CJK 疊層)讀回存成 PNG 後 os.Exit(0)。用於「改渲染 →
// 重跑 → 目視/像素比對」的驗證迴圈(對齊 retro-game-playtest skill 的可重現要求),
// 不影響正常執行(ShotPath 為空即完全跳過)。
package app

import (
	"image"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// 由 main/mobile 於 flag.Parse 後設定;ShotPath 為空時截圖邏輯完全停用。
var (
	ShotPath  string
	ShotFrame = 1 // 第幾幀截(1=首幀;需要動畫/狀態穩定時可設大一點)
)

// maybeShot 在達到 ShotFrame 時把 dst 讀回存 PNG 並結束程式。回傳是否已截圖(已截則
// 呼叫端不必再做事,因為程式即將 os.Exit)。
func maybeShot(dst *ebiten.Image, frame int) {
	if ShotPath == "" || frame < ShotFrame {
		return
	}
	b := dst.Bounds()
	img := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	dst.ReadPixels(img.Pix)
	f, err := os.Create(ShotPath)
	if err != nil {
		log.Fatalf("shot: create %s: %v", ShotPath, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatalf("shot: encode %s: %v", ShotPath, err)
	}
	log.Printf("shot: wrote %s (%dx%d, frame %d)", ShotPath, b.Dx(), b.Dy(), frame)
	os.Exit(0)
}
