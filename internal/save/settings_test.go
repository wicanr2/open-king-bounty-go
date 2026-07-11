package save

import "testing"

// TestSettingsRoundTrip:寫入偏好後讀回應一致;用 SetSaveDir 指向暫存目錄避免碰真實家目錄。
func TestSettingsRoundTrip(t *testing.T) {
	SetSaveDir(t.TempDir())
	defer SetSaveDir("")

	// 首次(無檔)應回零值 + nil error。
	got, err := LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings(首次) error = %v, want nil", err)
	}
	if got.Theme != "" {
		t.Errorf("LoadSettings(首次).Theme = %q, want 空", got.Theme)
	}

	if err := SaveSettings(Settings{Theme: "amiga"}); err != nil {
		t.Fatalf("SaveSettings error = %v", err)
	}

	got, err = LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings(存後) error = %v", err)
	}
	if got.Theme != "amiga" {
		t.Errorf("LoadSettings(存後).Theme = %q, want %q", got.Theme, "amiga")
	}
}

// TestSettingsIndependentOfSaves:settings.json 落在 openkb/(與存檔 openkb/saves/ 分開),
// 存偏好不應建立任何 slot 檔。
func TestSettingsIndependentOfSaves(t *testing.T) {
	SetSaveDir(t.TempDir())
	defer SetSaveDir("")

	if err := SaveSettings(Settings{Theme: "free"}); err != nil {
		t.Fatalf("SaveSettings error = %v", err)
	}
	slots, err := ListSaves()
	if err != nil {
		t.Fatalf("ListSaves error = %v", err)
	}
	if len(slots) != 0 {
		t.Errorf("存偏好後 ListSaves = %v, want 空(偏好不應建 slot 檔)", slots)
	}
}
