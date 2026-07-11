package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Settings 是跨執行的簡易使用者偏好(目前只有美術主題),存於可寫目錄的 openkb/settings.json
// (與存檔同一個 base:桌面 os.UserConfigDir、行動裝置由殼層 SetSaveDir 提供)。存檔本身走
// gamestate JSON,偏好與之分開放,互不影響。
type Settings struct {
	Theme string `json:"theme,omitempty"` // 上次選的美術模組名(dos/genesis/amiga/free);空=未設
}

// settingsPath 回傳 settings.json 路徑並確保 openkb 目錄存在(語意同 SaveDir 的 base 解析,
// 但落在 openkb/ 而非 openkb/saves/)。
func settingsPath() (string, error) {
	base := overrideBaseDir
	if base == "" {
		var err error
		base, err = os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("save: 取得使用者設定目錄失敗(行動裝置需由殼層 SetSaveDir 提供): %w", err)
		}
	}
	dir := filepath.Join(base, "openkb")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("save: 建立設定目錄 %q 失敗: %w", dir, err)
	}
	return filepath.Join(dir, "settings.json"), nil
}

// LoadSettings 讀回使用者偏好;設定檔不存在時回傳零值 Settings 與 nil error(首次執行正常情形,
// 呼叫端可直接用零值走預設)。
func LoadSettings() (Settings, error) {
	var s Settings
	path, err := settingsPath()
	if err != nil {
		return s, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return s, nil // 首次執行:無設定檔,回零值
		}
		return s, fmt.Errorf("save: 讀取設定 %q 失敗: %w", path, err)
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("save: 解析設定 %q 失敗: %w", path, err)
	}
	return s, nil
}

// SaveSettings 把偏好寫回 settings.json。
func SaveSettings(s Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("save: 序列化設定失敗: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("save: 寫入設定 %q 失敗: %w", path, err)
	}
	return nil
}
