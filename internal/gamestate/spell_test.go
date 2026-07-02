package gamestate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

// loadTestAssets 把 internal/kbdata/testdata 下的指定 ini 檔複製進
// t.TempDir()/free/ 子夾,再用 kbdata.Load 讀回,對齊 kbdata/assets_test.go
// TestLoadRealLayout 的作法(kbdata.Load("") 時 Strings 是空 map,LoadSpells/
// LoadArtifacts 才需要真的資料)。
func loadTestAssets(t *testing.T, names ...string) *kbdata.Assets {
	t.Helper()
	dir := t.TempDir()
	freeDir := filepath.Join(dir, "free")
	if err := os.Mkdir(freeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		src := filepath.Join("..", "kbdata", "testdata", n+".ini")
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("讀 %s: %v", src, err)
		}
		if err := os.WriteFile(filepath.Join(freeDir, n+".ini"), b, 0o644); err != nil {
			t.Fatalf("寫 %s: %v", n, err)
		}
	}
	a, err := kbdata.Load(dir)
	if err != nil {
		t.Fatalf("kbdata.Load: %v", err)
	}
	return a
}

func TestLoadSpells(t *testing.T) {
	a := loadTestAssets(t, "spells")
	spells := LoadSpells(a)
	if len(spells) != 14 {
		t.Fatalf("len(spells) = %d, want 14", len(spells))
	}

	// spell13 增控術:kbdata 的 FreeIni 用「per-section 後者覆寫前者」語意
	// (非 C 版全域計數器 bug),gold 應正確讀到 500 —— 見
	// internal/kbdata/free_ini.go 檔頭大註解與 free_ini_test.go 同一筆斷言。
	spell13 := spells[13]
	if spell13.Name != "增控術" {
		t.Errorf("spell13 Name = %q, want 增控術", spell13.Name)
	}
	if spell13.Type != "raise_control" {
		t.Errorf("spell13 Type = %q, want raise_control", spell13.Type)
	}
	if spell13.Gold != 500 {
		t.Errorf("spell13 Gold = %d, want 500", spell13.Gold)
	}
	if spell13.Combat {
		t.Errorf("spell13(raise_control) 應為冒險法術,IsCombat() 卻回 true")
	}

	// spell6 超渡術:testdata/spells.ini 已同步到 openkb-cht 修正版(移除上游那行
	// stray "gold = 2000"),故正確值為 250(與 free_ini_test.go 一致)。
	spell6 := spells[6]
	if spell6.Name != "超渡術" {
		t.Errorf("spell6 Name = %q, want 超渡術", spell6.Name)
	}
	if spell6.Type != "holy_damage" {
		t.Errorf("spell6 Type = %q, want holy_damage", spell6.Type)
	}
	if spell6.Gold != 250 {
		t.Errorf("spell6 Gold = %d, want 250", spell6.Gold)
	}
	if spell6.Damage != 50 {
		t.Errorf("spell6 Damage = %d, want 50", spell6.Damage)
	}
	if spell6.Filter != "undead" {
		t.Errorf("spell6 Filter = %q, want undead", spell6.Filter)
	}
	if !spell6.Combat {
		t.Errorf("spell6(holy_damage) 應為戰鬥法術,IsCombat() 卻回 false")
	}

	// 分類抽驗:7 個戰鬥法術(spell0..spell6)+ 7 個冒險法術(spell7..spell13),
	// 對齊 src/game.c:4256-4260 選單版面與 choose_spell() switch-case 語意
	// (見 spell.go spellCombatTypes 註解)。
	wantCombat := []bool{true, true, true, true, true, true, true, false, false, false, false, false, false, false}
	for i, want := range wantCombat {
		if got := spells[i].Combat; got != want {
			t.Errorf("spells[%d](%s/%s) Combat = %v, want %v", i, spells[i].Name, spells[i].Type, got, want)
		}
	}

	// 費用抽驗(直接對照 data/free/spells.ini 原始值,無重複 key 干擾的欄位)。
	if spells[0].Gold != 2000 || spells[0].Type != "clone" {
		t.Errorf("spell0 = %+v, want 複製術/clone/gold=2000", spells[0])
	}
	if spells[2].Gold != 1500 || spells[2].Damage != 25 || spells[2].Type != "magic_damage" {
		t.Errorf("spell2 = %+v, want 火球術/magic_damage/gold=1500/damage=25", spells[2])
	}
}

func TestLoadSpellsNilAssets(t *testing.T) {
	spells := LoadSpells(nil)
	if len(spells) != 14 {
		t.Fatalf("len(spells) = %d, want 14", len(spells))
	}
	for i, sp := range spells {
		if sp.ID != i {
			t.Errorf("spells[%d].ID = %d, want %d", i, sp.ID, i)
		}
		if sp.Name != "" || sp.Type != "" {
			t.Errorf("spells[%d] 應為零值(除 ID 外),got %+v", i, sp)
		}
	}
}

func TestLoadSpellsEmptyAssets(t *testing.T) {
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatalf("kbdata.Load(\"\"): %v", err)
	}
	spells := LoadSpells(a)
	if len(spells) != 14 {
		t.Fatalf("len(spells) = %d, want 14", len(spells))
	}
	if spells[0].Name != "" {
		t.Errorf("空 dir 應讀不到 spells.ini,spells[0].Name = %q", spells[0].Name)
	}
}
