package kbdata

import "testing"

// TestParseSigns 驗證 signs.txt 的「每隔一個換行當分隔」toggle 規則(對齊 C STRL_SIGNS)。
// 輸入 "A\n\nB\n\nC\nD\n" → 依 toggle:A / B / "C\nD"(第 5、6 個換行:C 後保留、D 後分隔)。
func TestParseSigns(t *testing.T) {
	// 逐則以「一個空行」相隔;則內可含單一換行。
	// C 的 toggle 規則會把每對雙換行的「前一個」\n 留在則內,故單行路標帶尾隨 \n。
	got := parseSigns("孤獨島\n\n觸發之島\n\n繼續探索\n這片土地\n")
	want := []string{"孤獨島\n", "觸發之島\n", "繼續探索\n這片土地"}
	if len(got) != len(want) {
		t.Fatalf("則數 = %d, want %d(%q)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("第 %d 則 = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestLoadSigns_Embedded 只驗證 signs 從 signs.txt 有載進來(非空),不釘死內容。
func TestLoadSigns_Embedded(t *testing.T) {
	// 直接驗 parseSigns 對多則輸入回非空即可(embedded 載入路徑由 gamestate 測試覆蓋)。
	if s := parseSigns("一\n\n二\n"); len(s) < 2 {
		t.Fatalf("parseSigns 應切出至少 2 則,got %d", len(s))
	}
}
