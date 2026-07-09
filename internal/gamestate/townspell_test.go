package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/kbrng"
)

// TestNewGame_TownSpell 驗證 salt_spells 移植(play.c:336)的核心行為對齊 C:
//  1. 0x15 號鎮(獵人村)固定教造橋術(spell 7)——C 硬編特例。
//  2. 所有 26 鎮的 town_spell 值落在 [0, maxSpells-1]。
//  3. 13 種法術(除造橋術外)逐一分配到不同鎮的迴圈,保證這 13 種全部出現過
//     (加上固定的造橋術,共 14 種法術應全部至少出現一次)——這是修掉舊落差
//     (26 鎮全教造橋術)的關鍵斷言。
//  4. 並非全部鎮都教造橋術(舊落差的直接反證)。
//
// 用多個不同 seed 跑同一組斷言,確認這些性質不是巧合單一 seed 撞出來的,
// 而是演算法結構保證(見 townspell.go 的迴圈:抽鎮直到命中空位)。
func TestNewGame_TownSpell(t *testing.T) {
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatalf("kbdata.Load: %v", err)
	}

	for _, seed := range []uint32{1, 2, 42, 12345} {
		gs := NewGame(a, "測試角色", 0, seed)

		if gs.TownSpell[bridgeTownID] != bridgeSpellID {
			t.Errorf("seed=%d: TownSpell[0x%02X] = %d, want %d(造橋術固定鎮)",
				seed, bridgeTownID, gs.TownSpell[bridgeTownID], bridgeSpellID)
		}

		seen := map[int]bool{}
		allBridge := true
		for i, sp := range gs.TownSpell {
			if sp < 0 || sp >= maxSpells {
				t.Errorf("seed=%d: TownSpell[%d] = %d 超出範圍 [0,%d]", seed, i, sp, maxSpells-1)
			}
			seen[sp] = true
			if sp != bridgeSpellID {
				allBridge = false
			}
		}
		if allBridge {
			t.Errorf("seed=%d: 全部 26 鎮都教造橋術,未修掉舊落差", seed)
		}
		if len(seen) < maxSpells {
			t.Errorf("seed=%d: 只出現 %d 種法術,want 全部 %d 種都至少出現一次(分佈=%v)",
				seed, len(seen), maxSpells, gs.TownSpell)
		}
	}
}

// TestSaltSpells_DirectAlgorithm 繞過 NewGame,直接測 saltSpells 本體,搭配
// kbrng.GlibcRand(注入,非 math/rand)驗證演算法本身(不依賴 NewGame 其餘欄位)。
func TestSaltSpells_DirectAlgorithm(t *testing.T) {
	rng := kbrng.NewGlibc(7)
	ts := saltSpells(rng)

	if ts[bridgeTownID] != bridgeSpellID {
		t.Fatalf("TownSpell[0x15] = %d, want %d", ts[bridgeTownID], bridgeSpellID)
	}

	countBySpell := map[int]int{}
	for _, sp := range ts {
		if sp < 0 || sp >= maxSpells {
			t.Fatalf("值超出範圍: %d", sp)
		}
		countBySpell[sp]++
	}
	// 13 個非造橋法術經由「唯一分配迴圈」各自恰好落在 1 個鎮 + 之後隨機補位可能疊加,
	// 因此每個法術出現次數至少 1(不強求恰好 1,补位鎮可能重複選中同一法術)。
	for sp := 0; sp < maxSpells; sp++ {
		if countBySpell[sp] < 1 {
			t.Errorf("法術 %d 從未分配到任何鎮(分佈=%v)", sp, ts)
		}
	}
}
