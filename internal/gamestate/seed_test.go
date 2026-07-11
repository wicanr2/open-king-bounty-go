package gamestate

import "testing"

// TestRandomWorldSeed_Differs 驗證正式新遊戲入口(charselect 用的 RandomWorldSeed)每次
// 回傳不同 seed —— 這是「每局世界隨機、有重玩價值」的底層保證。連續多次呼叫收集到的
// seed 不應全部相同(時間種子在相鄰呼叫間即分岔)。
func TestRandomWorldSeed_Differs(t *testing.T) {
	seen := map[uint32]bool{}
	for i := 0; i < 8; i++ {
		seen[RandomWorldSeed()] = true
	}
	if len(seen) < 2 {
		t.Errorf("RandomWorldSeed 連續 8 次只得 %d 種 seed,每局隨機失效", len(seen))
	}
}
