// 本檔用 external test package(gamestate_test),因為需要同時 import gamestate
// 與 save——而 save 套件本身 import gamestate,若用 internal test package(package
// gamestate)會構成 import cycle。所有用到的欄位/函式皆已 exported,external package
// 測試沒有存取限制。
package gamestate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
	"github.com/wicanr2/open-king-bounty-go/internal/save"
)

// loadWorldTestAssets 把 internal/kbdata/testdata/land.org 複製進 t.TempDir()/free/,
// 用 kbdata.Load 讀回,讓 a.World != nil,足以驅動 saltContinent。對齊
// spell_test.go 的 loadTestAssets 手法,只是這裡複製的是二進位地圖檔而非 ini。
func loadWorldTestAssets(t *testing.T) *kbdata.Assets {
	t.Helper()
	dir := t.TempDir()
	freeDir := filepath.Join(dir, "free")
	if err := os.MkdirAll(freeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join("..", "kbdata", "testdata", "land.org")
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("讀 %s: %v", src, err)
	}
	if err := os.WriteFile(filepath.Join(freeDir, "land.org"), b, 0o644); err != nil {
		t.Fatalf("寫 land.org: %v", err)
	}
	a, err := kbdata.Load(dir)
	if err != nil {
		t.Fatalf("kbdata.Load: %v", err)
	}
	if a.World == nil {
		t.Fatal("a.World 未載入,測試前提不成立")
	}
	return a
}

// countDwellingTiles 掃 gs.WorldMap 統計 dwelling tile(0x8C-0x8F)數量。
func countDwellingTiles(gs *gamestate.GameState) int {
	count := 0
	for cont := 0; cont < kbdata.MaxContinents; cont++ {
		for y := 0; y < kbdata.LevelH; y++ {
			for x := 0; x < kbdata.LevelW; x++ {
				tile := gs.WorldMap.Tile(cont, x, y)
				if tile >= kbdata.TileDwelling1 && tile <= kbdata.TileDwelling4 {
					count++
				}
			}
		}
	}
	return count
}

// TestNewGame_SaltContinent_DwellingTilesAppear 驗證 NewGame 跑完 saltContinent 後,
// 世界地圖上真的出現 dwelling tile(取代 salt 前 land.org 完全沒有 dwelling 的狀態,
// 見 docs/PORT-STATUS.md「世界生成架構計畫」的調查記錄)。
//
// 上限抓「每洲 10 個 dwelling、4 洲」= 40,但 saltContinent 的 telecave/dwelling/
// friendly 桶用「單次嘗試,可能因抽中已占用格而放置數 < min」的迴圈(逐句對齊
// C 版本身的不對稱行為,見 worldgen.go 函式註解),所以下限只斷言 > 0,不斷言恰好 40。
func TestNewGame_SaltContinent_DwellingTilesAppear(t *testing.T) {
	a := loadWorldTestAssets(t)
	gs := gamestate.NewGame(a, "測試角色", 0, gamestate.DefaultWorldSeed)

	if gs.WorldMap == nil {
		t.Fatal("gs.WorldMap 為 nil,world generation 未執行")
	}

	count := countDwellingTiles(gs)
	if count <= 0 {
		t.Fatalf("dwelling tile 數量 = %d, want > 0", count)
	}
	if count > kbdata.MaxContinents*10 {
		t.Errorf("dwelling tile 數量 = %d, 超過每洲 10 個的理論上限(4 洲共 40)", count)
	}
	t.Logf("dwelling tile 總數(4 洲合計) = %d", count)
}

// TestNewGame_SaltContinent_DwellingTroopTypeMatchesTile 驗證每個已放置的棲地,
// 地圖上實際 tile 值恰為 `kbdata.TileDwelling1 + a.Troops[troopID].Home` ——對齊 C
// `game->map[...][...] = TILE_DWELLING_1 + populate_dwelling(...)`,populate_dwelling
// 回傳值就是 troops[troop_id].dwells。
//
// ⚠ 刻意不斷言「tile 必落在 TileDwelling1..TileDwelling4 範圍內」:實測(seed=1)
// 抓到 continent=2 id=4 產生 tile=0x90(TileSignpost),因為該洲 dwelling_ranges[2]=
// [0x02,0x0e] 涵蓋 troop id=2(義勇軍,Home=DwellCastle=4),populate_dwelling 隨機
// roll 到它時 tile = TileDwelling1(0x8C) + 4 = 0x90,超出 TileDwelling4(0x8F)、
// 撞進 TileSignpost 的數值——這是**C 原始資料本身的既有怪癖**(dwelling_ranges 允許
// 隨機 roll 到 Home=Castle 的兵種,不是本移植引入的新行為),逐句對齊 C 邏輯與
// 資料表後如實重現,不是 bug。故本測試只驗證「tile 值與 Home 的算術關係」這個
// 不變量本身,不預設它必然落在 4 個具名 tile 內。
func TestNewGame_SaltContinent_DwellingTroopTypeMatchesTile(t *testing.T) {
	a := loadWorldTestAssets(t)
	gs := gamestate.NewGame(a, "測試角色", 0, gamestate.DefaultWorldSeed)

	checked := 0
	for cont := 0; cont < kbdata.MaxContinents; cont++ {
		for id := 0; id < kbdata.MaxDwellings; id++ {
			x, y := gs.DwellingCoords[cont][id][0], gs.DwellingCoords[cont][id][1]
			if x == 0 && y == 0 {
				continue // 該洲此 id 未被撒鹽放置(座標仍是零值),跳過
			}
			tile := gs.WorldMap.Tile(cont, x, y)

			troopID := gs.DwellingTroop[cont][id]
			if troopID < 0 || troopID >= len(a.Troops) {
				t.Errorf("cont=%d id=%d troopID=%d 越界(len=%d)", cont, id, troopID, len(a.Troops))
				continue
			}
			wantTile := kbdata.TileDwelling1 + byte(a.Troops[troopID].Home)
			if tile != wantTile {
				t.Errorf("cont=%d id=%d 座標(%d,%d): tile=0x%02X,want 0x%02X(TileDwelling1+兵種%d(%s)的Home=%d)",
					cont, id, x, y, tile, wantTile, troopID, a.Troops[troopID].Name, a.Troops[troopID].Home)
			}
			checked++
		}
	}
	if checked == 0 {
		t.Fatal("沒有任何棲地座標可供比對,測試前提不成立")
	}
	t.Logf("比對了 %d 個棲地的兵種型別與 tile 型別", checked)
}

// TestNewGame_SaltContinent_FoeTroopsNonEmpty 驗證世界生成後,一般 foe(非 friendly
// 佔位槽,index >= minFriendly=5)的 FoeTroops/FoeNumbers 確實被 repopulateFoe 填過
// (至少第一格數量 > 0)。
func TestNewGame_SaltContinent_FoeTroopsNonEmpty(t *testing.T) {
	a := loadWorldTestAssets(t)
	gs := gamestate.NewGame(a, "測試角色", 0, gamestate.DefaultWorldSeed)

	const minFriendly = 5 // 對齊 NewGame 呼叫 saltContinent(...,minFriendly=5)
	nonEmpty := 0
	for cont := 0; cont < kbdata.MaxContinents; cont++ {
		for id := minFriendly; id < kbdata.MaxFoes; id++ {
			x, y := gs.FoeCoords[cont][id][0], gs.FoeCoords[cont][id][1]
			if x == 0 && y == 0 {
				continue // 未放置
			}
			if gs.FoeNumbers[cont][id][0] > 0 {
				nonEmpty++
			}
		}
	}
	if nonEmpty <= 0 {
		t.Fatal("找不到任何已填入部隊的一般 foe,repopulateFoe 疑似未執行")
	}
	t.Logf("已生成部隊的一般 foe 數 = %d", nonEmpty)
}

// TestNewGame_SaltContinent_SaveLoadRoundTrip 驗證世界狀態(WorldMap/DwellingTroop/
// FoeTroops)完整隨 JSON 存讀檔往返,不需要另外擴充存檔格式(全部欄位已 exported)。
func TestNewGame_SaltContinent_SaveLoadRoundTrip(t *testing.T) {
	a := loadWorldTestAssets(t)
	gs := gamestate.NewGame(a, "測試角色", 0, gamestate.DefaultWorldSeed)

	path := filepath.Join(t.TempDir(), "slot0.json")
	if err := save.SaveToPath(gs, path); err != nil {
		t.Fatalf("SaveToPath: %v", err)
	}
	loaded, err := save.LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	if loaded.WorldMap == nil {
		t.Fatal("讀回的 WorldMap 為 nil")
	}
	for cont := 0; cont < kbdata.MaxContinents; cont++ {
		for y := 0; y < kbdata.LevelH; y++ {
			for x := 0; x < kbdata.LevelW; x++ {
				want := gs.WorldMap.Tile(cont, x, y)
				got := loaded.WorldMap.Tile(cont, x, y)
				if got != want {
					t.Fatalf("WorldMap[%d][%d][%d]: 存讀檔前後不一致,got 0x%02X want 0x%02X",
						cont, y, x, got, want)
				}
			}
		}
	}
	if loaded.DwellingTroop != gs.DwellingTroop {
		t.Error("DwellingTroop 存讀檔前後不一致")
	}
	if loaded.DwellingPopulation != gs.DwellingPopulation {
		t.Error("DwellingPopulation 存讀檔前後不一致")
	}
	if loaded.FoeTroops != gs.FoeTroops {
		t.Error("FoeTroops 存讀檔前後不一致")
	}
	if loaded.FoeNumbers != gs.FoeNumbers {
		t.Error("FoeNumbers 存讀檔前後不一致")
	}
}
