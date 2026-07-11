package gamestate

import "time"

// RandomWorldSeed 回傳一個以當下時間為基礎的世界 rand seed,供正式「開新遊戲」入口
// (charselect)使用,讓每局的權杖位置 / 敵人佈置 / 城鎮法術都不同——還原 1990 原版
// King's Bounty「每局世界隨機、有重玩價值」的體驗(openkb 重製版省略了 srand,固定
// seed 導致每局相同,見 DefaultWorldSeed 溯源說明)。
//
// 測試 / debug 仍用固定的 DefaultWorldSeed 走可重現路徑;唯有真正的玩家新遊戲才呼叫
// 本函式。seed 型別 uint32 對齊 kbrng.NewGlibc 的參數(glibc srand 取 unsigned int)。
func RandomWorldSeed() uint32 {
	// UnixNano 的低 32 bits 已足夠讓相鄰兩局分岔(玩家兩次開局間隔遠大於奈秒解析度)。
	return uint32(time.Now().UnixNano())
}
