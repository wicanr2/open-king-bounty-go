// Package save 負責玩家存檔的序列化與跨平台落地路徑。
//
// C 版用自訂二進位格式直接寫死在遊戲目錄旁;Go 重寫版改用 JSON——不需要對齊
// C 的 binary layout(存檔不是「反編當 oracle」要對照的資產,純粹是本版新設計
// 的 I/O 需求),換來人類可讀、跨平台好除錯的格式。
//
// 落地目錄鐵則:AppImage(squashfs)、macOS .app、Android assets 的 cwd 是唯讀
// 掛載,寫 cwd 會直接失敗。一律用 os.UserConfigDir() 取得「使用者可寫目錄」:
//   - Linux:   ~/.config
//   - macOS:   ~/Library/Application Support
//   - Windows: %AppData%
//   - Android/iOS/mobile shell:由殼層提供對應可寫路徑
//
// 存檔前一律 os.MkdirAll(具 mkdir -p 語意),不假設目錄已存在。
package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/wicanr2/open-king-bounty-go/internal/gamestate"
)

// currentVersion 是目前 saveFile 的格式版本號。未來欄位有不相容變動時遞增,
// LoadFromPath/LoadGame 可依版本號做遷移(目前只有 v1,尚無遷移邏輯)。
const currentVersion = 1

// saveFile 是實際寫進磁碟的信封(envelope):State 之外多帶一個 Version,
// 讓日後改 GameState 欄位時,讀檔端能判斷是否需要轉檔而不是直接解析失敗。
type saveFile struct {
	Version int                  `json:"version"`
	State   *gamestate.GameState `json:"state"`
}

// SaveDir 回傳存檔目錄(使用者可寫目錄下的 openkb/saves),並確保目錄存在。
func SaveDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("save: 取得使用者設定目錄失敗: %w", err)
	}

	dir := filepath.Join(base, "openkb", "saves")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("save: 建立存檔目錄 %q 失敗: %w", dir, err)
	}

	return dir, nil
}

// slotPath 回傳指定 slot 對應的存檔檔名(SaveDir()/slotN.json)。
func slotPath(dir string, slot int) string {
	return filepath.Join(dir, fmt.Sprintf("slot%d.json", slot))
}

// SaveGame 把 gs 序列化寫入指定 slot。
func SaveGame(gs *gamestate.GameState, slot int) error {
	dir, err := SaveDir()
	if err != nil {
		return err
	}
	return SaveToPath(gs, slotPath(dir, slot))
}

// LoadGame 讀回指定 slot 的存檔;slot 不存在時回傳的 error 滿足 os.IsNotExist。
func LoadGame(slot int) (*gamestate.GameState, error) {
	dir, err := SaveDir()
	if err != nil {
		return nil, err
	}
	return LoadFromPath(slotPath(dir, slot))
}

// ListSaves 列出 SaveDir() 下已存在的 slot 編號(由檔名 slotN.json 解析,遞增排序)。
func ListSaves() ([]int, error) {
	dir, err := SaveDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("save: 讀取存檔目錄 %q 失敗: %w", dir, err)
	}

	var slots []int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "slot") || !strings.HasSuffix(name, ".json") {
			continue
		}
		numStr := strings.TrimSuffix(strings.TrimPrefix(name, "slot"), ".json")
		n, err := strconv.Atoi(numStr)
		if err != nil {
			continue // 不符 slotN.json 命名慣例的檔案(使用者手動放的雜檔)直接略過
		}
		slots = append(slots, n)
	}

	sort.Ints(slots)
	return slots, nil
}

// SaveToPath 把 gs 序列化寫到 path(不經過 SaveDir,供測試 / 自訂路徑使用)。
func SaveToPath(gs *gamestate.GameState, path string) error {
	sf := saveFile{Version: currentVersion, State: gs}

	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return fmt.Errorf("save: 序列化存檔失敗: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("save: 建立存檔所在目錄失敗: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("save: 寫入存檔 %q 失敗: %w", path, err)
	}

	return nil
}

// LoadFromPath 從 path 讀回存檔(不經過 SaveDir,供測試 / 自訂路徑使用)。
// path 不存在時,回傳的 error 滿足 os.IsNotExist(err) == true。
func LoadFromPath(path string) (*gamestate.GameState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// 檔案不存在時,直接回傳 os.ReadFile 原始的 *PathError,不用 fmt.Errorf
		// 包一層——os.IsNotExist 只認得 *PathError/*LinkError/*SyscallError 這幾種
		// 具體型別(見 Go stdlib os/error.go underlyingError),包一層 %w 之後型別
		// 變成 *fmt.wrapError,os.IsNotExist 會誤判為 false。呼叫端要判斷「不存在」
		// 用 os.IsNotExist(err) 或 errors.Is(err, fs.ErrNotExist) 皆可。
		if errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		return nil, fmt.Errorf("save: 讀取存檔 %q 失敗: %w", path, err)
	}

	var sf saveFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, fmt.Errorf("save: 解析存檔 %q 失敗: %w", path, err)
	}

	return sf.State, nil
}
