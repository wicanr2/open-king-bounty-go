package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// TestMoraleNames_IndexAlignment 鎖住 morale_names 索引與 MORALE_* 常數值的對齊
// (bounty.c:591 morale_names[3] = {"普通","低落","高昂"},索引即 MORALE_NORMAL/
// LOW/HIGH 的字面值 0/1/2)。錯位是最容易犯、也最難從畫面肉眼發現的一種 bug
// (顯示的字沒亂碼,只是「講的是別種士氣」)。
func TestMoraleNames_IndexAlignment(t *testing.T) {
	tests := []struct {
		morale int
		want   string
	}{
		{MoraleNormal, "普通"},
		{MoraleLow, "低落"},
		{MoraleHigh, "高昂"},
	}
	for _, tt := range tests {
		if got := MoraleNames[tt.morale]; got != tt.want {
			t.Errorf("MoraleNames[%d] = %q, want %q", tt.morale, got, tt.want)
		}
	}
}

// troopWithGroup 建立一個只設定 MoraleGroup 的最小 kbdata.Assets(TroopMorale 只讀
// 這個欄位),避免測試耦合到 tables_troops.go 的真實資料內容。
func assetsWithGroups(groups ...int) *kbdata.Assets {
	troops := make([]kbdata.Troop, len(groups))
	for i, g := range groups {
		troops[i] = kbdata.Troop{MoraleGroup: g}
	}
	return &kbdata.Assets{Troops: troops}
}

// TestTroopMorale 逐案例驗證 troop_morale(play.c:642) 的「取全隊最差」邏輯,
// 對照 bounty.c morale_chart[5][5](A=0..E=4)。
func TestTroopMorale(t *testing.T) {
	tests := []struct {
		name   string
		groups []int // troopID i 的 morale group,i 對應 Army 格
		army   [5]Squad
		slot   int
		want   int
	}{
		{
			// 單一部隊、group A,自己對自己查表:chart[A][A] = N。
			name:   "單部隊_同群A對A_普通",
			groups: []int{0},
			army:   [5]Squad{{TroopID: 0, Count: 5}, {TroopID: 255}, {TroopID: 255}, {TroopID: 255}, {TroopID: 255}},
			slot:   0,
			want:   MoraleNormal,
		},
		{
			// 兩支同為 group D 的部隊:chart[D][D] = H,全隊最差仍是 H。
			name:   "兩部隊_同群D對D_高昂",
			groups: []int{3, 3},
			army:   [5]Squad{{TroopID: 0, Count: 5}, {TroopID: 1, Count: 3}, {TroopID: 255}, {TroopID: 255}, {TroopID: 255}},
			slot:   0,
			want:   MoraleHigh,
		},
		{
			// 自己 group A + 友軍 group D:自查 chart[A][A]=N,友軍查 chart[D][A]=L,
			// 取最差 = L。驗證「不是掃到第一個就停,要比較全部取最小」。
			name:   "群A混群D友軍_取最差_低落",
			groups: []int{0, 3},
			army:   [5]Squad{{TroopID: 0, Count: 5}, {TroopID: 1, Count: 3}, {TroopID: 255}, {TroopID: 255}, {TroopID: 255}},
			slot:   0,
			want:   MoraleLow,
		},
		{
			// 群C對群C:chart[C][C] = H。
			name:   "群C對群C_高昂",
			groups: []int{2},
			army:   [5]Squad{{TroopID: 0, Count: 5}, {TroopID: 255}, {TroopID: 255}, {TroopID: 255}, {TroopID: 255}},
			slot:   0,
			want:   MoraleHigh,
		},
		{
			// 中間出現 Count==0 要立刻停止掃描(對齊 C `if (numbers[j]==0) break`),
			// 即使更後面的格子有資料也不看:army[1].Count==0,army[2] 的 group D
			// 不該被納入計算,slot0(group A)應維持只對自己查表 = N,不是 L。
			name:   "中段空格提早停止_不看後面",
			groups: []int{0, 0, 3},
			army:   [5]Squad{{TroopID: 0, Count: 5}, {TroopID: 255, Count: 0}, {TroopID: 2, Count: 5}, {TroopID: 255}, {TroopID: 255}},
			slot:   0,
			want:   MoraleNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := &GameState{Army: tt.army}
			a := assetsWithGroups(tt.groups...)
			if got := TroopMorale(gs, a, tt.slot); got != tt.want {
				t.Errorf("TroopMorale() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestTroopMorale_NilSafe 確認 nil/越界輸入回傳預設 MoraleHigh,不 panic
// (呼叫端 viewarmy.go 只在 Count>0 時使用回傳值,但防禦式檢查仍要正確)。
func TestTroopMorale_NilSafe(t *testing.T) {
	if got := TroopMorale(nil, nil, 0); got != MoraleHigh {
		t.Errorf("nil gs/a: got %d, want %d", got, MoraleHigh)
	}
	gs := &GameState{}
	a := assetsWithGroups(0)
	if got := TroopMorale(gs, a, -1); got != MoraleHigh {
		t.Errorf("負 slot: got %d, want %d", got, MoraleHigh)
	}
	if got := TroopMorale(gs, a, 99); got != MoraleHigh {
		t.Errorf("越界 slot: got %d, want %d", got, MoraleHigh)
	}
}
