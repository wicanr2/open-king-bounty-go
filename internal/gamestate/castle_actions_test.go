package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/embedded"
	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// loadAssetsT 用 embedded FS 載入真實 free 資料(land.ini/castles.ini 等)。
// 不能用 kbdata.Load("")——那條路徑 LoadFS(nil) 只回硬編 troop 表,Strings 全空。
func loadAssetsT(t *testing.T) *kbdata.Assets {
	t.Helper()
	a, err := kbdata.LoadFS(embedded.FS())
	if err != nil {
		t.Fatalf("kbdata.LoadFS: %v", err)
	}
	return a
}

// TestSaveCastleOwnerKnowledge 驗證三分支:玩家/怪物佔領不動,惡棍佔領 OR 進 KNOWN 位。
func TestSaveCastleOwnerKnowledge(t *testing.T) {
	gs := &GameState{}
	gs.CastleOwner[0] = KBCastlePlayer
	gs.CastleOwner[1] = castleOwnerMonsters
	gs.CastleOwner[2] = 5 // villain 5

	gs.SaveCastleOwnerKnowledge(0)
	gs.SaveCastleOwnerKnowledge(1)
	gs.SaveCastleOwnerKnowledge(2)

	if gs.CastleOwner[0] != KBCastlePlayer {
		t.Errorf("玩家城堡 owner 被改動 = %#x", gs.CastleOwner[0])
	}
	if gs.CastleOwner[1] != castleOwnerMonsters {
		t.Errorf("怪物城堡 owner 被改動 = %#x", gs.CastleOwner[1])
	}
	if gs.CastleOwner[2] != (5|KBCastleKnown) {
		t.Errorf("惡棍城堡 owner = %#x, want %#x", gs.CastleOwner[2], 5|KBCastleKnown)
	}
	if gs.CastleOwner[2]&KBCastleVillain != 5 {
		t.Errorf("加 KNOWN 後 villain id 應仍可取回 5, got %d", gs.CastleOwner[2]&KBCastleVillain)
	}
}

// TestGarrisonTroop_LastTroopBlocked 驗證只剩一支部隊時駐防被擋(回 2)。
func TestGarrisonTroop_LastTroopBlocked(t *testing.T) {
	gs := &GameState{}
	gs.Army[0] = Squad{TroopID: 3, Count: 10}
	gs.Army[1] = Squad{TroopID: 0xFF, Count: 0} // 只剩一支
	for i := range gs.CastleTroops[0] {
		gs.CastleTroops[0][i] = 0xFF
	}

	if got := gs.GarrisonTroop(0, 0); got != 2 {
		t.Fatalf("GarrisonTroop = %d, want 2(最後一支部隊)", got)
	}
	if gs.Army[0].Count != 10 {
		t.Errorf("被擋下卻仍動了隊伍:Army[0].Count = %d", gs.Army[0].Count)
	}
}

// TestGarrisonTroop_Success 驗證成功駐防:城堡加兵、玩家該格移除並遞補。
func TestGarrisonTroop_Success(t *testing.T) {
	gs := &GameState{}
	gs.Army[0] = Squad{TroopID: 3, Count: 10}
	gs.Army[1] = Squad{TroopID: 7, Count: 20}
	for i := range gs.CastleTroops[0] {
		gs.CastleTroops[0][i] = 0xFF
	}

	if got := gs.GarrisonTroop(0, 0); got != 0 {
		t.Fatalf("GarrisonTroop = %d, want 0", got)
	}
	if gs.CastleTroops[0][0] != 3 || gs.CastleNumbers[0][0] != 10 {
		t.Errorf("城堡未正確加兵:troop=%d num=%d", gs.CastleTroops[0][0], gs.CastleNumbers[0][0])
	}
	// 玩家隊伍:slot0 被移除,原 slot1(7/20)遞補到 slot0
	if gs.Army[0].TroopID != 7 || gs.Army[0].Count != 20 {
		t.Errorf("dismiss 後遞補錯誤:Army[0] = %+v", gs.Army[0])
	}
	if gs.Army[4].TroopID != 0xFF {
		t.Errorf("dismiss 後末格未清空:Army[4] = %+v", gs.Army[4])
	}
}

// TestGarrisonTroop_MergesSameType 驗證同兵種併格(對齊 C 找 same troop 分支)。
func TestGarrisonTroop_MergesSameType(t *testing.T) {
	gs := &GameState{}
	gs.Army[0] = Squad{TroopID: 3, Count: 10}
	gs.Army[1] = Squad{TroopID: 7, Count: 20}
	gs.CastleTroops[0] = [5]int{3, 0xFF, 0xFF, 0xFF, 0xFF}
	gs.CastleNumbers[0] = [5]int{5, 0, 0, 0, 0}

	gs.GarrisonTroop(0, 0)
	if gs.CastleNumbers[0][0] != 15 {
		t.Errorf("同兵種未合併:CastleNumbers[0][0] = %d, want 15", gs.CastleNumbers[0][0])
	}
}

// TestUngarrisonTroop 驗證撤離:玩家收兵,城堡該格移除並前移、末格清空。
func TestUngarrisonTroop(t *testing.T) {
	gs := &GameState{}
	gs.Army = [5]Squad{{TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}, {TroopID: 0xFF}}
	gs.CastleTroops[0] = [5]int{3, 7, 0xFF, 0xFF, 0xFF}
	gs.CastleNumbers[0] = [5]int{10, 20, 0, 0, 0}

	if got := gs.UngarrisonTroop(0, 0); got != 0 {
		t.Fatalf("UngarrisonTroop = %d, want 0", got)
	}
	if gs.Army[0].TroopID != 3 || gs.Army[0].Count != 10 {
		t.Errorf("玩家未收到兵:Army[0] = %+v", gs.Army[0])
	}
	if gs.CastleTroops[0][0] != 7 || gs.CastleNumbers[0][0] != 20 {
		t.Errorf("城堡未前移:slot0 troop=%d num=%d", gs.CastleTroops[0][0], gs.CastleNumbers[0][0])
	}
	if gs.CastleTroops[0][4] != 0xFF || gs.CastleNumbers[0][4] != 0 {
		t.Errorf("城堡末格未清空:troop=%d num=%d", gs.CastleTroops[0][4], gs.CastleNumbers[0][4])
	}
}

// TestHomeTroops 驗證家鄉兵清單非空且都是城堡棲地兵種。
func TestHomeTroops(t *testing.T) {
	a := loadAssetsT(t)
	ht := HomeTroops(a)
	if len(ht) == 0 {
		t.Fatal("HomeTroops 為空,應有城堡棲地兵種(對齊 C DWELLING_CASTLE 迴圈)")
	}
	for _, id := range ht {
		if a.Troops[id].Home != kbdata.DwellCastle {
			t.Errorf("HomeTroops 含非城堡棲地兵種 id=%d Home=%v", id, a.Troops[id].Home)
		}
	}
}

// TestCastleAt_Home 驗證家鄉座標(land.ini special0)判為 CastleHome / HomeCastleID。
func TestCastleAt_Home(t *testing.T) {
	a := loadAssetsT(t)
	hc, hx, hy, ok := HomeCastleCoords(a)
	if !ok {
		t.Fatal("讀不到家鄉座標(land.ini special0)")
	}
	gs := &GameState{}
	kind, id := gs.CastleAt(a, hc, hx, hy)
	if kind != CastleHome || id != HomeCastleID {
		t.Fatalf("家鄉座標判定 = (%v,%d), want (CastleHome,%d)", kind, id, HomeCastleID)
	}
}

// TestCastleAt_YoursVsEnemy 驗證同一城堡座標依 CastleOwner 分成 your/enemy。
func TestCastleAt_YoursVsEnemy(t *testing.T) {
	a := loadAssetsT(t)
	castles := LoadCastles(a)
	// 找第一個有座標的非家鄉城堡
	var c *Castle
	for i := range castles {
		if castles[i].X != 0 || castles[i].Y != 0 {
			c = &castles[i]
			break
		}
	}
	if c == nil {
		t.Skip("castles.ini 無座標,略過")
	}
	gs := &GameState{}
	gs.CastleOwner[c.ID] = KBCastlePlayer
	if kind, id := gs.CastleAt(a, c.Continent, c.X, c.Y); kind != CastleYours || id != c.ID {
		t.Errorf("玩家城堡判定 = (%v,%d), want (CastleYours,%d)", kind, id, c.ID)
	}
	gs.CastleOwner[c.ID] = castleOwnerMonsters
	if kind, id := gs.CastleAt(a, c.Continent, c.X, c.Y); kind != CastleEnemy || id != c.ID {
		t.Errorf("怪物城堡判定 = (%v,%d), want (CastleEnemy,%d)", kind, id, c.ID)
	}
}
