package kbdata

import (
	"path/filepath"
	"strings"
	"testing"
)

func loadTestIni(t *testing.T, name string) *FreeIni {
	t.Helper()
	ini, err := LoadFreeIni(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("LoadFreeIni(%s) failed: %v", name, err)
	}
	return ini
}

func TestTroopsIni_SectionCountAndFields(t *testing.T) {
	ini := loadTestIni(t, "troops.ini")

	if len(ini.Sections) != 25 {
		t.Fatalf("expected 25 [troopN] sections, got %d", len(ini.Sections))
	}

	troop0, ok := ini.NumberedSection("troop", 0)
	if !ok {
		t.Fatalf("expected [troop0] section")
	}
	if name, _ := troop0.Get("name"); name != "農夫" {
		t.Errorf("troop0 name = %q, want 農夫", name)
	}
	if v, ok := troop0.Int("skill"); !ok || v != 1 {
		t.Errorf("troop0 skill = %v (ok=%v), want 1", v, ok)
	}
	if v, ok := troop0.Int("hp"); !ok || v != 1 {
		t.Errorf("troop0 hp = %v (ok=%v), want 1", v, ok)
	}
	if v, _ := troop0.Get("melee"); v != "1-1" {
		t.Errorf("troop0 melee = %q, want 1-1", v)
	}
	if v, _ := troop0.Get("abilities"); v != "none" {
		t.Errorf("troop0 abilities = %q, want none", v)
	}

	troop24, ok := ini.NumberedSection("troop", 24)
	if !ok {
		t.Fatalf("expected [troop24] section")
	}
	if name, _ := troop24.Get("name"); name != "火龍" {
		t.Errorf("troop24 name = %q, want 火龍", name)
	}

	names := ini.StringSlice("troop", "name", 25)
	if len(names) != 25 {
		t.Fatalf("StringSlice length = %d, want 25", len(names))
	}
	if names[0] != "農夫" || names[24] != "火龍" {
		t.Errorf("StringSlice troop names[0]=%q names[24]=%q", names[0], names[24])
	}
}

func TestSpellsIni_ShippedValues(t *testing.T) {
	// testdata/spells.ini 已同步到修正版(移除 [spell6] 那行 stray "gold = 2000"),
	// 與 openkb-cht 出貨資料一致:超渡術(spell6)= 250、增控術(spell13)= 500。
	// (上游 GNU_extract_ini 的全域計數器 + 重複 key 會讓 spell13 讀成 0,見 free_ini.go
	//  檔頭;本 Go parser 依 section 索引,不重現該 bug。last-wins 性質另由合成測試涵蓋。)
	ini := loadTestIni(t, "spells.ini")

	spell6, ok := ini.NumberedSection("spell", 6)
	if !ok {
		t.Fatalf("expected [spell6] section")
	}
	if v, ok := spell6.Int("gold"); !ok || v != 250 {
		t.Errorf("spell6 gold = %v (ok=%v), want 250", v, ok)
	}
	if v, ok := spell6.Int("damage"); !ok || v != 50 {
		t.Errorf("spell6 damage = %v (ok=%v), want 50", v, ok)
	}
	if v, _ := spell6.Get("filter"); v != "undead" {
		t.Errorf("spell6 filter = %q, want undead", v)
	}

	spell13, ok := ini.NumberedSection("spell", 13)
	if !ok {
		t.Fatalf("expected [spell13] section")
	}
	if v, ok := spell13.Int("gold"); !ok || v != 500 {
		t.Errorf("spell13 gold = %v (ok=%v), want 500 (this Go parser does NOT reproduce the C-side global-counter bug)", v, ok)
	}
	if name, _ := spell13.Get("name"); name != "增控術" {
		t.Errorf("spell13 name = %q, want 增控術", name)
	}

	goldSlice := ini.IntSlice("spell", "gold", 14)
	if len(goldSlice) != 14 {
		t.Fatalf("IntSlice length = %d, want 14", len(goldSlice))
	}
	if goldSlice[13] != 500 {
		t.Errorf("IntSlice gold[13] = %d, want 500", goldSlice[13])
	}
}

// TestParseFreeIni_DuplicateKeyLastWins 用合成輸入涵蓋「同 section 內重複 key 後者覆寫」
// 的 parser 性質(獨立於出貨資料,不再綁 spells.ini 的 stray 行)。
func TestParseFreeIni_DuplicateKeyLastWins(t *testing.T) {
	ini := ParseFreeIni([]byte("[x]\ngold = 250\ngold = 2000\n"))
	sec, ok := ini.Section("x")
	if !ok {
		t.Fatal("expected [x]")
	}
	if v, ok := sec.Int("gold"); !ok || v != 2000 {
		t.Errorf("重複 key 應後者覆寫: got %v, want 2000", v)
	}
}

func TestTownsIni_Coordinates(t *testing.T) {
	ini := loadTestIni(t, "towns.ini")

	if len(ini.Sections) != 26 {
		t.Fatalf("expected 26 [townN] sections, got %d", len(ini.Sections))
	}

	town0, ok := ini.NumberedSection("town", 0)
	if !ok {
		t.Fatalf("expected [town0] section")
	}
	if name, _ := town0.Get("name"); name != "艾米蓋爾" {
		t.Errorf("town0 name = %q, want 艾米蓋爾", name)
	}
	if v, ok := town0.Int("x"); !ok || v != 52 {
		t.Errorf("town0 x = %v (ok=%v), want 52", v, ok)
	}
	if v, ok := town0.Int("y"); !ok || v != 5 {
		t.Errorf("town0 y = %v (ok=%v), want 5", v, ok)
	}
	if v, ok := town0.Int("invert_id"); !ok || v != 0 {
		t.Errorf("town0 invert_id = %v (ok=%v), want 0", v, ok)
	}
}

func TestVillainsIni_ArmyEntryWithTab(t *testing.T) {
	ini := loadTestIni(t, "villains.ini")

	if len(ini.Sections) != 17 {
		t.Fatalf("expected 17 [villainN] sections, got %d", len(ini.Sections))
	}

	villain0, ok := ini.NumberedSection("villain", 0)
	if !ok {
		t.Fatalf("expected [villain0] section")
	}
	if name, _ := villain0.Get("name"); name != "騙徒莫里斯" {
		t.Errorf("villain0 name = %q, want 騙徒莫里斯", name)
	}
	if v, ok := villain0.Int("reward"); !ok || v != 5000 {
		t.Errorf("villain0 reward = %v (ok=%v), want 5000", v, ok)
	}

	army0, ok := villain0.Get("army0")
	if !ok {
		t.Fatalf("expected army0 key")
	}
	count, troop, ok := ParseArmyEntry(army0)
	if !ok || count != 50 || troop != "農夫" {
		t.Errorf("ParseArmyEntry(%q) = (%d, %q, %v), want (50, 農夫, true)", army0, count, troop, ok)
	}

	// army4 在原始檔案裡 "=" 後面是 tab 而非空白 (villains.ini 第 9 行:
	// "army4 =\t25 x 農夫"),驗證 tab 被正確當空白吃掉、值本身不殘留 tab。
	army4, ok := villain0.Get("army4")
	if !ok {
		t.Fatalf("expected army4 key")
	}
	if strings.ContainsAny(army4, "\t") {
		t.Errorf("army4 value %q should not contain a literal tab", army4)
	}
	count4, troop4, ok := ParseArmyEntry(army4)
	if !ok || count4 != 25 || troop4 != "農夫" {
		t.Errorf("ParseArmyEntry(%q) = (%d, %q, %v), want (25, 農夫, true)", army4, count4, troop4, ok)
	}
}

func TestArtifactsIni_DescriptionKeepsLiteralEscapes(t *testing.T) {
	ini := loadTestIni(t, "artifacts.ini")

	if len(ini.Sections) != 8 {
		t.Fatalf("expected 8 [artifactN] sections, got %d", len(ini.Sections))
	}

	artifact0, ok := ini.NumberedSection("artifact", 0)
	if !ok {
		t.Fatalf("expected [artifact0] section")
	}
	if name, _ := artifact0.Get("name"); name != "王者之劍" {
		t.Errorf("artifact0 name = %q, want 王者之劍", name)
	}
	if power, _ := artifact0.Get("power"); power != "increased_damage" {
		t.Errorf("artifact0 power = %q, want increased_damage", power)
	}
	desc, ok := artifact0.Get("description")
	if !ok {
		t.Fatalf("expected description key")
	}
	// ini 解析層不展開 "\n" 轉義(那是 GNU_expand_newlines 的職責,屬更高層),
	// 這裡只驗證原始逐字保留、字面上的反斜線-n序列還在。
	if !strings.Contains(desc, `\n`) {
		t.Errorf("artifact0 description %q should still contain literal `\\n` escapes", desc)
	}
	if strings.Contains(desc, "\n") {
		t.Errorf("artifact0 description should not contain an actual newline byte")
	}
}

func TestUiIni_NonNumberedSections(t *testing.T) {
	ini := loadTestIni(t, "ui.ini")

	top, ok := ini.Section("top")
	if !ok {
		t.Fatalf("expected [top] section")
	}
	if w, ok := top.Int("w"); !ok || w != 320 {
		t.Errorf("top.w = %v (ok=%v), want 320", w, ok)
	}
	if h, ok := top.Int("h"); !ok || h != 8 {
		t.Errorf("top.h = %v (ok=%v), want 8", h, ok)
	}

	tile, ok := ini.Section("tile")
	if !ok {
		t.Fatalf("expected [tile] section")
	}
	if w, _ := tile.Int("w"); w != 48 {
		t.Errorf("tile.w = %d, want 48", w)
	}
}

func TestColorsIni_HexParsing(t *testing.T) {
	ini := loadTestIni(t, "colors.ini")

	minimap, ok := ini.Section("minimap")
	if !ok {
		t.Fatalf("expected [minimap] section")
	}
	if v, ok := minimap.Int("shallow_water"); !ok || v != 0x0000ff {
		t.Errorf("minimap.shallow_water = %#v (ok=%v), want 0x0000ff", v, ok)
	}
	if v, ok := minimap.Int("grass"); !ok || v != 0x00ff00 {
		t.Errorf("minimap.grass = %#v (ok=%v), want 0x00ff00", v, ok)
	}
	if v, ok := minimap.Int("castle"); !ok || v != 0xffffff {
		t.Errorf("minimap.castle = %#v (ok=%v), want 0xffffff", v, ok)
	}
}

func TestFreeIni_CaseInsensitiveLookup(t *testing.T) {
	ini := loadTestIni(t, "troops.ini")
	sec, ok := ini.Section("TROOP0")
	if !ok {
		t.Fatalf("expected case-insensitive section lookup to find [troop0]")
	}
	if v, ok := sec.Get("NAME"); !ok || v != "農夫" {
		t.Errorf("case-insensitive key lookup: got %q (ok=%v), want 農夫", v, ok)
	}
}

func TestFreeIni_MissingSectionAndKey(t *testing.T) {
	ini := loadTestIni(t, "troops.ini")
	if _, ok := ini.Section("nope"); ok {
		t.Errorf("expected missing section to report ok=false")
	}
	troop0, _ := ini.NumberedSection("troop", 0)
	if _, ok := troop0.Get("does_not_exist"); ok {
		t.Errorf("expected missing key to report ok=false")
	}
	if v := troop0.GetDefault("does_not_exist", "fallback"); v != "fallback" {
		t.Errorf("GetDefault fallback = %q, want fallback", v)
	}
	if v := troop0.IntDefault("does_not_exist", 42); v != 42 {
		t.Errorf("IntDefault fallback = %d, want 42", v)
	}
}
