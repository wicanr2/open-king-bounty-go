package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// maxTowns 對齊 C 版 bounty.h MAX_TOWNS(26)。
const maxTowns = 26

// Town 是單一鄉鎮的靜態資料,來源 data/free/towns.ini(經 kbdata.FreeIni 解析),
// 欄位對照 C 版 town_coords[]/town_names[]/game->town_spell[](bounty.c/game.c
// visit_town():以 continent/x/y 掃描 town_coords[] 找到 id,town_names[id] 是
// 鄉鎮顯示名,town_spell[id] 是該鎮教的法術索引)。
type Town struct {
	ID           int
	Name         string // ini `name`,對應 C town_names[id]
	Continent    int    // ini `continent`,對應 C town_coords[i][0]
	X, Y         int    // ini `x`/`y`,對應 C town_coords[i][1..2]
	SpellID      int    // ini `spell`,對應 C game->town_spell[id](該鎮教的法術索引,查 spells.ini)
	BoatX, BoatY int    // ini `boat_x`/`boat_y`,對應 C boat_coords[i][1..2](租船時船隻停靠點,DAT_BOATX/Y 覆寫)
	GateX, GateY int    // ini `gate_x`/`gate_y`,對應 C towngate_coords[i][1..2](鄉鎮傳送落點,SPELL_TOWNGATE 用)
}

// LoadTowns 從 a.Strings["towns"](towns.ini 解析結果)建出 26 個 Town。
// 缺檔或缺區塊時該筆只剩 ID(其餘欄位零值),不 panic —— 對齊 kbdata.Assets 的
// best-effort 慣例(見 LoadSpells)。
func LoadTowns(a *kbdata.Assets) []Town {
	towns := make([]Town, maxTowns)
	for i := range towns {
		towns[i].ID = i
	}
	if a == nil {
		return towns
	}
	ini := a.Strings["towns"]
	if ini == nil {
		return towns
	}
	for i := 0; i < maxTowns; i++ {
		sec, ok := ini.NumberedSection("town", i)
		if !ok {
			continue
		}
		t := &towns[i]
		t.Name = sec.GetDefault("name", "")
		t.Continent = sec.IntDefault("continent", 0)
		t.X = sec.IntDefault("x", 0)
		t.Y = sec.IntDefault("y", 0)
		t.SpellID = sec.IntDefault("spell", 0)
		t.BoatX = sec.IntDefault("boat_x", 0)
		t.BoatY = sec.IntDefault("boat_y", 0)
		t.GateX = sec.IntDefault("gate_x", 0)
		t.GateY = sec.IntDefault("gate_y", 0)
	}
	return towns
}

// FindTown 依大陸/座標找鎮 id,對應 C visit_town() 開頭那段「掃描 town_coords[]
// 找 continent/x/y 相符的那一筆」邏輯。找不到回 -1(C 版掃不到時 id 維持初始值 0,
// 呼叫端應自行決定退回 0 還是視為錯誤 —— 見 WorldMapScreen 呼叫處)。
func FindTown(towns []Town, continent, x, y int) int {
	for _, t := range towns {
		if t.Continent == continent && t.X == x && t.Y == y {
			return t.ID
		}
	}
	return -1
}
