package screen

// loadart.go -- 美術主題(module)載入與切換。
//
// 一個「主題」= 一個美術 PNG 目錄名(如 "free" / "dos" / "fmtowns"):tileset + 所有
// sprite + UI 圖都取自該目錄。切換主題 = 用新 dir 重跑 LoadArt,把全域 sprite/tileset
// setter 全部覆寫,一次全畫面生效(對齊 C openkb 的 module 系統,但只換美術,不換
// 遊戲資料——DOS/FMTowns/free 的兵種數值/城鎮/land 相同,資料層維持 free)。
//
// 缺檔沿用 best-effort:個別 asset 缺就 log 跳過,故版權主題只要抽好一部分也能漸進上線。
// 詳見 docs/theme-switching-plan.md。

import (
	"io/fs"
	"log"

	"github.com/wicanr2/open-king-bounty-go/internal/render"
	"github.com/wicanr2/open-king-bounty-go/internal/save"
)

// portraitFiles 是四職業立繪檔名(view_character 用),index = 職業,對齊桌面/行動載入。
var portraitFiles = [4]string{"knig", "pala", "sorc", "barb"}

// locationBgFiles 是棲地背景(招募畫面用)的 sub_id → 檔名對照,對齊 main.go / mobile.go。
var locationBgFiles = []struct {
	subID int
	file  string
}{
	{0, "cstl.png"},
	{2, "plai.png"},
	{3, "frst.png"},
	{4, "cave.png"},
	{5, "dngn.png"},
}

// baseTheme 是缺檔回退來源:任何主題缺的美術檔一律用 free 補(free 恆為完整開放集)。
const baseTheme = "free"

// LoadArt 載入 dir 主題的整套美術到全域 setter。先鋪 free 當底(保證每個 asset 都有),
// 再用 dir 主題覆蓋——版權主題(DOS/Genesis/Amiga)若某些 UI 小件沒抽到,就自動沿用
// free 的那件,避免切換後殘留前一主題或空白。dir=="free" 時只鋪一次。
// 桌面傳 os.DirFS(datadir),行動傳 embed.FS。必須在遊戲迴圈啟動後(繪圖就緒)呼叫。
func LoadArt(fsys fs.FS, dir string) {
	SetFrameArt(nil) // 重置:frame.png 只有版權主題(DOS)有,切到無此檔的主題要清掉殘留
	for i := 0; i < 3; i++ {
		SetEndTile(i, nil) // 重置破關動畫 tile(切主題清殘留)
	}
	loadArtDir(fsys, baseTheme)
	if dir != baseTheme {
		loadArtDir(fsys, dir)
	}
}

// loadArtDir 從 fsys 的 dir 目錄載入整套美術並灌進全域 setter(tileset + 所有 sprite + UI)。
// best-effort:個別檔缺就 log 跳過(由 LoadArt 的 free 底層補),不中斷。
func loadArtDir(fsys fs.FS, dir string) {
	p := dir + "/"

	if ts, err := render.LoadTilesetFS(fsys, dir); err == nil {
		SetTileset(ts)
	} else {
		log.Printf("load tileset [%s]: %v(地圖退回色塊)", dir, err)
	}
	if hero, err := render.LoadSpriteFS(fsys, p+"cursor.png", 48, 34); err == nil {
		SetHero(hero)
	} else {
		log.Printf("load hero [%s]: %v", dir, err)
	}
	if sidebar, err := render.LoadSpriteOpaqueFS(fsys, p+"sidebar.png", 48, 34); err == nil {
		SetSidebar(sidebar)
	} else {
		log.Printf("load sidebar [%s]: %v", dir, err)
	}
	if piece, err := render.LoadSpriteOpaqueFS(fsys, p+"piece.png", 9, 6); err == nil {
		SetPiece(piece)
	} else {
		log.Printf("load piece [%s]: %v", dir, err)
	}
	if coins, err := render.LoadSpriteOpaqueFS(fsys, p+"coins.png", 16, 5); err == nil {
		SetCoins(coins)
	} else {
		log.Printf("load coins [%s]: %v", dir, err)
	}
	if art, err := render.LoadPNGTileFS(fsys, p+"select-0.png"); err == nil {
		SetSelectArt(art)
	} else {
		log.Printf("load select art [%s]: %v", dir, err)
	}
	if art, err := render.LoadPNGTileFS(fsys, p+"title.png"); err == nil {
		SetTitleArt(art)
	} else {
		log.Printf("load title art [%s]: %v", dir, err)
	}
	if art, err := render.LoadPNGTileFS(fsys, p+"nwcp.png"); err == nil {
		SetLogoArt(art)
	} else {
		log.Printf("load logo art [%s]: %v", dir, err)
	}
	// 金框邊框圖(只有版權主題有;缺就維持 nil，drawChromeFrame 退回平面框，不 log 噪音)
	if art, err := render.LoadPNGTileFS(fsys, p+"frame.png"); err == nil {
		SetFrameArt(art)
	}
	// 破關動畫 tile(endpic-2/-3/-4 = 草/橋/英雄,free 開放美術;缺就跳過動畫,不 log 噪音)。
	for i, n := range []string{"endpic-2", "endpic-3", "endpic-4"} {
		if art, err := render.LoadPNGTileFS(fsys, p+n+".png"); err == nil {
			SetEndTile(i, art)
		}
	}
	if comtiles, err := render.LoadSpriteOpaqueFS(fsys, p+"comtiles.png", 48, 34); err == nil {
		SetComtiles(comtiles)
	} else {
		log.Printf("load comtiles [%s]: %v", dir, err)
	}
	if art, err := render.LoadPNGTileFS(fsys, p+"town.png"); err == nil {
		SetLocation(art)
	} else {
		log.Printf("load town background [%s]: %v", dir, err)
	}
	for _, d := range locationBgFiles {
		if art, err := render.LoadPNGTileFS(fsys, p+d.file); err == nil {
			SetLocationBg(d.subID, art)
		} else {
			log.Printf("load location bg %s [%s]: %v", d.file, dir, err)
		}
	}
	for id, name := range TroopFileNames {
		if sp, err := render.LoadSpriteFS(fsys, p+name+".png", 48, 34); err == nil {
			SetTroopSprite(id, sp)
		} else {
			log.Printf("load troop sprite %s [%s]: %v", name, dir, err)
		}
	}
	for class, name := range portraitFiles {
		if art, err := render.LoadPNGTileFS(fsys, p+name+".png"); err == nil {
			SetPortrait(class, art)
		} else {
			log.Printf("load portrait %s [%s]: %v", name, dir, err)
		}
	}
	if items, err := render.LoadSpriteOpaqueFS(fsys, p+"view.png", 40, 34); err == nil {
		SetViewItems(items)
	} else {
		log.Printf("load view items [%s]: %v", dir, err)
	}
	for id, name := range VillainFileNames {
		if sp, err := render.LoadSpriteOpaqueFS(fsys, p+name+".png", 48, 34); err == nil {
			SetVillainFace(id, sp)
		} else {
			log.Printf("load villain face %s [%s]: %v", name, dir, err)
		}
	}
}

// --- 主題切換 ---
//
// 套件級 ThemeManager 狀態:記住 fsys + 可用模組清單 + 目前選中,讓 F8 / 觸控 ☰ 能循環
// 切換。用套件級變數(而非注入 app.Game)以免動到入口簽名——app 只需呼叫 CycleTheme()。

var (
	themeFS     fs.FS
	themeMods   []string // 實際存在(該 dir 有 tileseta.png)的模組,依偏好序
	themeActive int
)

// InitThemes 依偏好序 order(如 ["dos","fmtowns","free"])過濾出實際內建的美術模組,
// 載入第一個存在者當預設,並記住狀態供之後 CycleTheme。回傳選中的模組名("" = 都不存在)。
func InitThemes(fsys fs.FS, order []string) string {
	themeFS = fsys
	themeMods = resolveThemes(fsys, order)
	themeActive = 0
	if len(themeMods) == 0 {
		return ""
	}
	// 套用上次玩家選的主題偏好(save/settings.json;由 CycleTheme 寫入)。找不到 / 不在
	// 可用清單就維持預設(themeMods[0]),best-effort:讀設定失敗不阻擋啟動。只在此決定起始
	// index,LoadArt 只跑一次(不先載預設再載偏好)。debug 的 -theme 由 main.go 隨後 SetActiveTheme
	// 覆寫(SetActiveTheme 不寫偏好,故 debug run 不污染玩家設定)。
	if s, err := save.LoadSettings(); err == nil {
		themeActive = themeStartIndex(themeMods, s.Theme)
	}
	LoadArt(fsys, themeMods[themeActive])
	return themeMods[themeActive]
}

// themeStartIndex 回傳起始主題 index:偏好 pref 若在 mods 中則用其 index,否則回 0(預設=
// 偏好序第一個)。純邏輯供單元測試(不觸發 LoadArt/建圖)。
func themeStartIndex(mods []string, pref string) int {
	if pref != "" {
		for i, m := range mods {
			if m == pref {
				return i
			}
		}
	}
	return 0
}

// resolveThemes 過濾偏好序 order,只留下實際內建(該 dir 有 tileseta.png)的美術模組,
// 保持 order 的先後(第一個即預設)。純邏輯、不建圖,供單元測試。
func resolveThemes(fsys fs.FS, order []string) []string {
	var mods []string
	for _, m := range order {
		if themeExists(fsys, m) {
			mods = append(mods, m)
		}
	}
	return mods
}

// CycleTheme 切到下一個可用美術模組(繞回)並重載美術。回傳新選中的模組名。
// 少於 2 個模組時不動作,回傳目前模組名(或 "")。
func CycleTheme() string {
	if len(themeMods) < 2 || themeFS == nil {
		return ActiveTheme()
	}
	themeActive = (themeActive + 1) % len(themeMods)
	LoadArt(themeFS, themeMods[themeActive])
	// 記住玩家選擇,下次啟動沿用(best-effort:寫入失敗只記 log,不影響切換本身)。
	if err := save.SaveSettings(save.Settings{Theme: themeMods[themeActive]}); err != nil {
		log.Printf("theme: 儲存主題偏好失敗: %v", err)
	}
	return themeMods[themeActive]
}

// SetActiveTheme 切到指定美術模組(若存在於可用清單)並重載美術;回傳是否成功。
// 供 debug flag / 設定持久化指定起始主題用。
func SetActiveTheme(name string) bool {
	for i, m := range themeMods {
		if m == name {
			themeActive = i
			if themeFS != nil {
				LoadArt(themeFS, m)
			}
			return true
		}
	}
	return false
}

// ActiveTheme 回傳目前美術模組名("" = 尚未 InitThemes 或無可用模組)。
func ActiveTheme() string {
	if themeActive < 0 || themeActive >= len(themeMods) {
		return ""
	}
	return themeMods[themeActive]
}

// AvailableThemes 回傳實際內建、可切換的美術模組清單(依偏好序)。
func AvailableThemes() []string { return append([]string(nil), themeMods...) }

// themeExists 以「該目錄有 tileseta.png」判定美術模組是否內建(版權主題未內建時自動缺席)。
func themeExists(fsys fs.FS, dir string) bool {
	if fsys == nil {
		return false
	}
	f, err := fsys.Open(dir + "/tileseta.png")
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}
