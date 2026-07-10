// villain.go -- 惡棍(villain)靜態資料 + 世界狀態查詢,對齊 C 版 view_contract
// 用到的兩個查詢:find_villain_castle(play.c:1853)+ villains.ini 的 name/reward/file 欄位。
//
// 描述文字(STRL_VDESCS,view_contract 實際顯示的多行懸賞公告)不在這裡 ——
// 那份資料是 kbdata.Assets.VillainDescs(見 kbdata/assets.go 載入邏輯),因為它是
// 純資料讀檔(讀 <file>.txt),不是世界狀態查詢。
package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// Villain 是單一惡棍的靜態資料,來源 data/free/villains.ini(經 kbdata.FreeIni 解析)。
//
// ⚠ 名稱與懸賞金額欄位核實(rulebook 62):view_contract 實際顯示的文字**不是**
// 這裡的 Name/Reward —— 是 kbdata.Assets.VillainDescs[id](STRL_VDESCS 讀
// `<file>.txt`,檔案本身已包含「姓名:X / 懸賞:Y 金幣」等文字)。Villain.Name/
// Reward 對應 C 版 STRL_VNAMES/WDAT_VREWARD,是其他畫面(如未來的 visit_castle
// 顯示佔領城堡的惡棍名、town.go 契約 UI)會用到的通用查詢,先在此建好,
// 不是 view_contract 專屬。
type Villain struct {
	ID     int
	Name   string // ini `name`,對應 C STRL_VNAMES / villain_names[id]
	Reward int    // ini `reward`,對應 C WDAT_VREWARD / villain_reward[id]
	File   string // ini `file`,對應 C DOS_villain_names[id](GR_VILLAIN 素材檔名 stem,亦是 STRL_VDESCS 的 <file>.txt 來源)
}

// LoadVillains 從 a.Strings["villains"] 建出 kbdata.MaxVillains(17)個 Villain。
// 缺檔或缺區塊時該筆只剩 ID(其餘欄位零值),不 panic —— 對齊 LoadTowns/LoadCastles
// 的 best-effort 慣例。
func LoadVillains(a *kbdata.Assets) []Villain {
	villains := make([]Villain, kbdata.MaxVillains)
	for i := range villains {
		villains[i].ID = i
	}
	if a == nil {
		return villains
	}
	ini := a.Strings["villains"]
	if ini == nil {
		return villains
	}
	for i := 0; i < kbdata.MaxVillains; i++ {
		sec, ok := ini.NumberedSection("villain", i)
		if !ok {
			continue
		}
		v := &villains[i]
		v.Name = sec.GetDefault("name", "")
		v.Reward = sec.IntDefault("reward", 0)
		v.File = sec.GetDefault("file", "")
	}
	return villains
}

// castleOwnerVillainMask 對齊 C 版 bounty.h:136 註解「LOW 5 bits = villain id」,
// find_villain_castle(play.c:1857)用 `& 0x3f` 濾掉 KBCASTLE_KNOWN(0x40)這個
// 旗標位元,只留惡棍 id 本身(0-16 五位元夠用;0x3f 是 C 原始碼字面遮罩值,
// 沿用不改窄,理由見 play.c 逐字對照)。
const castleOwnerVillainMask = 0x3f

// FindVillainCastle 對齊 C 版 find_villain_castle(game, villain_id)(play.c:1853):
// 掃描 gs.CastleOwner[0..MAX_CASTLES-1],回傳第一個「(owner & 0x3f) == villainID」
// 的城堡 id;找不到回 -1(C 版同樣回 -1,view_contract 據此顯示「未知/未知」)。
//
// 注意:C 版原函式簽章只吃 (game, villain_id),不需要城堡靜態表(座標/名稱查詢
// 是呼叫端另外用 castle id 查 LoadCastles 結果),故本移植維持同樣窄介面。
func FindVillainCastle(gs *GameState, villainID int) int {
	if gs == nil {
		return -1
	}
	for i := 0; i < kbdata.MaxCastles; i++ {
		if int(gs.CastleOwner[i])&castleOwnerVillainMask == villainID {
			return i
		}
	}
	return -1
}
