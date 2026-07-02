package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// maxArtifacts 對齊 C 版 MAX_ARTIFACTS(bounty.h:44)。
const maxArtifacts = 8

// artifactPowerFlags 把 ini `power` 字串對照到 C 版 bounty.h 的 POWER_* 位元旗標,
// 來源為 src/lib/free-data.c:GNU_artifact_downto_byte() 的 names[]/powers[] 兩張表
// (逐一對應,非猜測)。這些旗標會用 OR 疊加進 C 版 KBgame.artifact_power(
// has_power() 用 & 檢查),但「疊加/生效」屬於世界狀態(玩家已拾得哪些寶物),
// 不在本切片範圍 —— 這裡只提供「這件寶物本身對應哪個旗標」的靜態查詢。
var artifactPowerFlags = map[string]int{
	"unknown":             0x01, // POWER_UNKNOWN_XXX1(bounty.h 未進一步命名效果)
	"quarter_protection":  0x02, // POWER_QUARTER_PROTECTION
	"double_leadership":   0x04, // POWER_DOUBLE_LEADERSHIP
	"double_spell_power":  0x08, // POWER_DOUBLE_SPELL_POWER
	"increase_commission": 0x10, // POWER_INCREASE_COMMISSION
	"cheaper_boat_rental": 0x20, // POWER_CHEAPER_BOAT_RENTAL
	"double_max_spells":   0x40, // POWER_DOUBLE_MAX_SPELLS
	"increased_damage":    0x80, // POWER_INCREASED_DAMAGE
}

// Artifact 是單一寶物的靜態資料,來源為 data/free/artifacts.ini 的
// `[artifactN]` 區塊(經 kbdata.FreeIni 解析,欄位語意對照 artifacts.ini
// 檔頭註解與 src/lib/free-data.c 的 DAT_APOWER/STRL_ARTIFACTS case)。
//
// 本型別只涵蓋「資料 + 屬性查詢」,不含「玩家拾得寶物後生效」(world/combat
// 狀態,見套件 TODO)與「地圖上寶物擺放」(invert_id 只是原始資料透傳,
// 地圖 tileset 對應留給後續地圖 loader)。
type Artifact struct {
	ID          int    // 索引(0..7),對齊 ini `[artifactN]` 與 C 版 artifact_found[MAX_ARTIFACTS] 的位置
	Name        string // 繁中寶物名(ini `name`;C 版註解稱「currently unused」,即遊戲內文字不靠這欄,但仍原樣保留供顯示/工具用)
	Power       string // ini 原始 `power` 字串(如 "increased_damage"),對照 artifactPowerFlags 查旗標
	Description string // 拾得時顯示文字(ini `description`,含字面 "\n" 換行記號,未轉譯——轉譯留給顯示層)
	InvertID    int    // ini `invert_id`:地圖上寶物出現順序(對應 tileseta.png/tilesalt.png 排列),純資料透傳
}

// PowerFlag 回傳這件寶物對應的 C 版 POWER_* 位元旗標;無法辨識的 Power 字串回 0
// (對照 C 版 GNU_artifact_downto_byte()「Failed to parse」時該格維持記憶體
// 預設值 0 的行為 —— 0 不是任何一個 POWER_* 常數,可安全當「無效果」判斷)。
func (a Artifact) PowerFlag() int {
	return artifactPowerFlags[a.Power]
}

// LoadArtifacts 從 a.Strings["artifacts"](artifacts.ini 解析結果)建出 8 個 Artifact。
// 缺檔或缺區塊時該筆 Artifact 只剩 ID(其餘欄位零值),不 panic —— 對齊
// kbdata.Assets 的 best-effort 慣例(見 kbdata/assets.go Load 註解)。
func LoadArtifacts(a *kbdata.Assets) []Artifact {
	artifacts := make([]Artifact, maxArtifacts)
	for i := range artifacts {
		artifacts[i].ID = i
	}
	if a == nil {
		return artifacts
	}
	ini := a.Strings["artifacts"]
	if ini == nil {
		return artifacts
	}
	for i := 0; i < maxArtifacts; i++ {
		sec, ok := ini.NumberedSection("artifact", i)
		if !ok {
			continue
		}
		art := &artifacts[i]
		art.Name = sec.GetDefault("name", "")
		art.Power = sec.GetDefault("power", "")
		art.Description = sec.GetDefault("description", "")
		art.InvertID = sec.IntDefault("invert_id", 0)
	}
	return artifacts
}
