package kbrng

import (
	"encoding/json"
	"os"
	"testing"
)

// goldenSample 對應 C 版 KB_ORACLE hook 印出的黃金樣本(見 openkb-cht src/game.c)。
// 只取 kbrng 需要的欄位;classes 由 gamestate 的 parity 測試使用。
type goldenSample struct {
	Seed      int   `json:"seed"`
	KBRand099 []int `json:"kb_rand_0_99"`
}

// TestParityWithCOracle 是 kbrng 對 C oracle 的正式 parity 釘子:
// 讀 C 版在固定 seed 下實跑產出的 KB_rand(0,99) 序列,驗 Go 逐一重現。
// golden_seed1.json 由 `KB_ORACLE=1 KB_ORACLE_SEED=1 ./openkb ...` 產生。
func TestParityWithCOracle(t *testing.T) {
	raw, err := os.ReadFile("testdata/golden_seed1.json")
	if err != nil {
		t.Fatalf("讀 golden: %v", err)
	}
	var g goldenSample
	if err := json.Unmarshal(raw, &g); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	if len(g.KBRand099) == 0 {
		t.Fatal("golden 無 kb_rand 序列")
	}
	r := NewGlibc(uint32(g.Seed))
	for i, want := range g.KBRand099 {
		got := r.Between(0, 99)
		if got != want {
			t.Fatalf("KB_rand(0,99) 第 %d 個: Go=%d, C oracle=%d", i, got, want)
		}
	}
}
