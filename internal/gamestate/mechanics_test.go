package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// fixedRand 是測試用固定亂數,讓 EndWeek 的 creature pick 可預期。
type fixedRand struct{ v int }

func (f fixedRand) Between(min, max int) int { return f.v }
func (f fixedRand) Seed(uint32)              {}

func loadAssets(t *testing.T) *kbdata.Assets {
	t.Helper()
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// 晉升逐階累加(騎士):base_leadership 100→200→500→1000;commission 1000→2000→4000→8000。
func TestPromote_KnightAccumulates(t *testing.T) {
	a := loadAssets(t)
	gs := NewGame(a, "K", 0)

	wantLead := []int{200, 500, 1000, 1000} // 第 4 次已達最高階,不再變
	wantComm := []int{2000, 4000, 8000, 8000}
	for i := 0; i < 4; i++ {
		gs.Promote(a)
		if gs.BaseLeadership != wantLead[i] {
			t.Errorf("Promote #%d: BaseLeadership=%d, want %d", i+1, gs.BaseLeadership, wantLead[i])
		}
		if gs.Commission != wantComm[i] {
			t.Errorf("Promote #%d: Commission=%d, want %d", i+1, gs.Commission, wantComm[i])
		}
	}
	if gs.Rank != 3 {
		t.Errorf("最高階應為 3, got %d", gs.Rank)
	}
}

// issue #5 回歸:寶箱領導力加成過週不流失(EndWeek 把 Leadership 重設回含加成的 BaseLeadership)。
func TestChestLeadership_SurvivesEndWeek(t *testing.T) {
	a := loadAssets(t)
	gs := NewGame(a, "K", 0) // 騎士 base=lead=100
	if gs.BaseLeadership != 100 || gs.Leadership != 100 {
		t.Fatalf("起手應 base=lead=100, got base=%d lead=%d", gs.BaseLeadership, gs.Leadership)
	}

	gs.AddChestLeadership(50)
	if gs.BaseLeadership != 150 || gs.Leadership != 150 {
		t.Fatalf("寶箱後應 base=lead=150, got base=%d lead=%d", gs.BaseLeadership, gs.Leadership)
	}

	gs.EndWeek(a, fixedRand{v: 5}) // week 1,creature=5(非農夫週,不觸發轉化)
	if gs.Leadership != 150 {
		t.Fatalf("issue #5 回歸失敗:過週後 Leadership=%d,應保留寶箱加成 150(不該回到 100)", gs.Leadership)
	}
	if gs.BaseLeadership != 150 {
		t.Fatalf("BaseLeadership 不該被過週改動, got %d", gs.BaseLeadership)
	}
}

// 寶箱 + 升階疊加:證明採「增量 +=」模型(若誤用賦值絕對值會清掉寶箱 → 200)。
func TestChestThenPromote_Stacks(t *testing.T) {
	a := loadAssets(t)
	gs := NewGame(a, "K", 0) // base 100
	gs.AddChestLeadership(50) // base 150
	gs.Promote(a)             // rank1 增量 +100 → 250(非賦值 200)
	if gs.BaseLeadership != 250 {
		t.Fatalf("寶箱+升階應疊加為 250, got %d(=200 表示誤用賦值清掉了寶箱)", gs.BaseLeadership)
	}
}

// 女巫師的 knows_magic 升階後仍為真(增量累加後恆 1)。
func TestSorceressKnowsMagic_PersistsAcrossPromote(t *testing.T) {
	a := loadAssets(t)
	gs := NewGame(a, "S", 2) // 女巫師
	if !gs.KnowsMagic {
		t.Fatal("女巫師起手應會魔法")
	}
	gs.Promote(a)
	if !gs.KnowsMagic {
		t.Fatal("女巫師升階後仍應會魔法")
	}
}

// 招兵:金錢檢查、扣款、空格合併。
func TestBuyTroop(t *testing.T) {
	a := loadAssets(t)
	gs := NewGame(a, "K", 0) // 騎士 gold 7500,隊伍 [義勇軍x20, 弓箭手x2]
	startGold := gs.Gold

	// 農夫(id 0,GoldCost 10)x10 = 100 金 → 進第一個空格(slot 2)
	if err := gs.BuyTroop(a, 0, 10); err != nil {
		t.Fatalf("招募農夫應成功: %v", err)
	}
	if gs.Gold != startGold-100 {
		t.Errorf("扣款後 gold=%d, want %d", gs.Gold, startGold-100)
	}
	if gs.Army[2].TroopID != 0 || gs.Army[2].Count != 10 {
		t.Errorf("農夫應在 slot2 x10, got %+v", gs.Army[2])
	}
	// 再招 5 隻農夫 → 合併同格
	if err := gs.BuyTroop(a, 0, 5); err != nil {
		t.Fatal(err)
	}
	if gs.Army[2].Count != 15 {
		t.Errorf("農夫應合併為 15, got %d", gs.Army[2].Count)
	}
}

func TestBuyTroop_NotEnoughGold(t *testing.T) {
	a := loadAssets(t)
	gs := NewGame(a, "K", 0)
	gs.Gold = 5
	// 農夫 x1 = 10 金 > 5
	if err := gs.BuyTroop(a, 0, 1); err != ErrNotEnoughGold {
		t.Fatalf("金錢不足應回 ErrNotEnoughGold, got %v", err)
	}
	if gs.Gold != 5 {
		t.Errorf("失敗不該扣款, gold=%d", gs.Gold)
	}
}

// 領導力上限:剩餘領導力 / 每隻 HP。
func TestArmyMaxTroopCount(t *testing.T) {
	a := loadAssets(t)
	gs := NewGame(a, "K", 0) // leadership 100
	// 農夫 HP=1 → 空隊時可放 100 隻(尚無農夫佔用)
	if got := gs.ArmyMaxTroopCount(a, 0); got != 100 {
		t.Errorf("農夫上限 got %d, want 100", got)
	}
}

// 每週結算:領導力重設、俸祿入帳、維護費扣除。
func TestEndWeek_LeadershipResetAndCommission(t *testing.T) {
	a := loadAssets(t)
	gs := NewGame(a, "K", 0) // gold 7500, commission 1000, lead 100
	// 消耗一些當前領導力語意上不改 base;直接改 Leadership 模擬戰損後
	gs.Leadership = 40
	goldBefore := gs.Gold

	gs.EndWeek(a, fixedRand{v: 5})

	if gs.Leadership != gs.BaseLeadership {
		t.Errorf("過週後 Leadership(%d) 應重設回 BaseLeadership(%d)", gs.Leadership, gs.BaseLeadership)
	}
	// 俸祿 +1000;維護費 = 20*(義勇軍50/10=5)+2*(弓箭手250/10=25)=100+50=150
	wantGold := goldBefore + 1000 - 150
	if gs.Gold != wantGold {
		t.Errorf("過週後 gold=%d, want %d (俸祿+1000, 維護-150)", gs.Gold, wantGold)
	}
}
