// Package render 是 Ebiten 繪製層:把 kbdata 的唯讀資料(CJK 點陣字、free
// tile PNG…)畫成畫面。只讀 kbdata/gamestate 提供的狀態,不持有遊戲邏輯。
//
// 目前提供:
//   - cjktext.go:CJKAtlas 點陣字 → 貼圖 + 逐字繪製字串。
//   - tile.go:free tile PNG 載入與繪製。
package render
