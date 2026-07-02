package gamestate

import (
	"testing"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

func TestLoadArtifacts(t *testing.T) {
	a := loadTestAssets(t, "artifacts")
	artifacts := LoadArtifacts(a)
	if len(artifacts) != 8 {
		t.Fatalf("len(artifacts) = %d, want 8", len(artifacts))
	}

	art0 := artifacts[0]
	if art0.Name != "王者之劍" {
		t.Errorf("artifact0 Name = %q, want 王者之劍", art0.Name)
	}
	if art0.Power != "increased_damage" {
		t.Errorf("artifact0 Power = %q, want increased_damage", art0.Power)
	}
	if art0.InvertID != 7 {
		t.Errorf("artifact0 InvertID = %d, want 7", art0.InvertID)
	}
	if art0.PowerFlag() != 0x80 {
		t.Errorf("artifact0 PowerFlag() = %#x, want 0x80(POWER_INCREASED_DAMAGE)", art0.PowerFlag())
	}

	art6 := artifacts[6]
	if art6.Name != "尼克羅斯之書" {
		t.Errorf("artifact6 Name = %q, want 尼克羅斯之書", art6.Name)
	}
	if art6.Power != "double_max_spells" {
		t.Errorf("artifact6 Power = %q, want double_max_spells", art6.Power)
	}
	if art6.PowerFlag() != 0x40 {
		t.Errorf("artifact6 PowerFlag() = %#x, want 0x40(POWER_DOUBLE_MAX_SPELLS)", art6.PowerFlag())
	}

	// 分類抽驗:8 個 power 字串應各自對應到不重複的 bounty.h POWER_* 位元
	// (見 src/lib/free-data.c:GNU_artifact_downto_byte() names[]/powers[])。
	wantFlags := []int{0x80, 0x02, 0x04, 0x10, 0x08, 0x01, 0x40, 0x20}
	seen := map[int]bool{}
	for i, want := range wantFlags {
		got := artifacts[i].PowerFlag()
		if got != want {
			t.Errorf("artifacts[%d](%s) PowerFlag() = %#x, want %#x", i, artifacts[i].Power, got, want)
		}
		if seen[got] {
			t.Errorf("artifacts[%d] PowerFlag() %#x 與前面重複", i, got)
		}
		seen[got] = true
	}
}

func TestLoadArtifactsNilAssets(t *testing.T) {
	artifacts := LoadArtifacts(nil)
	if len(artifacts) != 8 {
		t.Fatalf("len(artifacts) = %d, want 8", len(artifacts))
	}
	for i, art := range artifacts {
		if art.ID != i {
			t.Errorf("artifacts[%d].ID = %d, want %d", i, art.ID, i)
		}
		if art.Name != "" || art.Power != "" {
			t.Errorf("artifacts[%d] 應為零值(除 ID 外),got %+v", i, art)
		}
		if art.PowerFlag() != 0 {
			t.Errorf("artifacts[%d] 空 Power 的 PowerFlag() 應為 0,got %#x", i, art.PowerFlag())
		}
	}
}

func TestLoadArtifactsEmptyAssets(t *testing.T) {
	a, err := kbdata.Load("")
	if err != nil {
		t.Fatalf("kbdata.Load(\"\"): %v", err)
	}
	artifacts := LoadArtifacts(a)
	if len(artifacts) != 8 {
		t.Fatalf("len(artifacts) = %d, want 8", len(artifacts))
	}
	if artifacts[0].Name != "" {
		t.Errorf("空 dir 應讀不到 artifacts.ini,artifacts[0].Name = %q", artifacts[0].Name)
	}
}
