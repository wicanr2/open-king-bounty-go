package app

import "testing"

func TestThemeLabel(t *testing.T) {
	cases := map[string]string{
		"dos":     "DOS 經典",
		"genesis": "Genesis",
		"amiga":   "Amiga",
		"free":    "開放美術",
		"unknown": "unknown", // 未知模組名原樣回傳
	}
	for mod, want := range cases {
		if got := themeLabel(mod); got != want {
			t.Errorf("themeLabel(%q) = %q, want %q", mod, got, want)
		}
	}
}

func TestShowToastSetsMessageAndTicks(t *testing.T) {
	g := &Game{}
	g.showToast("主題:Amiga")
	if g.toastMsg != "主題:Amiga" {
		t.Errorf("toastMsg = %q, want %q", g.toastMsg, "主題:Amiga")
	}
	if g.toastTicks != toastTicksDefault {
		t.Errorf("toastTicks = %d, want %d", g.toastTicks, toastTicksDefault)
	}
}
