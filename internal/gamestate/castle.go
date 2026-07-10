package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// Castle 是單一城堡的靜態資料,來源 data/free/castles.ini(經 kbdata.FreeIni 解析)
// + kbdata.CastleDifficultyTable()(bounty.c 硬編,不受模組覆寫)。
//
// 對照 C 版全域陣列 castle_coords[MAX_CASTLES][3]({continent,x,y},bounty.c:647)
// + castle_difficulty[MAX_CASTLES](bounty.c:612)。
//
// ⚠ 座標不可照抄 bounty.c 字面(與 gamestate/town.go 讀 towns.ini、不用 bounty.c
// town_coords 同理):C 版 refill_rules()(game.c:304-307)用 DAT_CASTLEX/Y/C 覆寫
// castle_coords,free 模組的 castles.ini 座標是依自製 land.tmx 地圖佈局算出來的,
// 與 bounty.c 的 DOS 硬編值**不同**(核對 2026-07-10:castle0 在 bounty.c 是
// {0,30,27},castles.ini 卻是 {continent=0,x=44,y=12})。故本檔改讀 castles.ini,
// 不沿用 kbdata 的 bounty.c 座標表(該表本來就沒有另外移植,只有 CastleDifficultyTable
// 這個「未被覆寫」的欄位才照抄自 bounty.c)。
//
// Difficulty 則相反:free 模組**沒有** DAT_CASTLEDIFF 這個覆寫鍵(kbres.h/free-data.c
// 搜尋確認),castles.ini 裡的 "difficulty" 欄位是 land.tmx 匯出工具留下的殘留欄位,
// C 引擎從未讀取——實際遊戲邏輯(repopulate_castle)永遠用 bounty.c 硬編的
// castle_difficulty[],故本檔的 Difficulty 欄位查 kbdata.CastleDifficultyTable(),
// 不讀 castles.ini 的同名欄位。
type Castle struct {
	ID         int
	Name       string // ini `name`
	Continent  int    // ini `continent`,對應 C castle_coords[i][0]
	X, Y       int    // ini `x`/`y`,對應 C castle_coords[i][1..2]
	Difficulty int    // kbdata.CastleDifficultyTable()[id],對應 C castle_difficulty[id](不受 ini 覆寫)
}

// LoadCastles 從 a.Strings["castles"](castles.ini 解析結果)+ kbdata.CastleDifficultyTable()
// 建出 kbdata.MaxCastles(26)個 Castle。缺檔或缺區塊時該筆只剩 ID/Difficulty
// (座標零值),不 panic——對齊 kbdata.Assets 的 best-effort 慣例(見 town.go LoadTowns)。
func LoadCastles(a *kbdata.Assets) []Castle {
	diff := kbdata.CastleDifficultyTable()
	castles := make([]Castle, kbdata.MaxCastles)
	for i := range castles {
		castles[i].ID = i
		castles[i].Difficulty = int(diff[i])
	}
	if a == nil {
		return castles
	}
	ini := a.Strings["castles"]
	if ini == nil {
		return castles
	}
	for i := 0; i < kbdata.MaxCastles; i++ {
		sec, ok := ini.NumberedSection("castle", i)
		if !ok {
			continue
		}
		c := &castles[i]
		c.Name = sec.GetDefault("name", "")
		c.Continent = sec.IntDefault("continent", 0)
		c.X = sec.IntDefault("x", 0)
		c.Y = sec.IntDefault("y", 0)
	}
	return castles
}

// numCastles 對齊 C play.c:128 num_castles(continent):數該洲城堡數。
func numCastles(castles []Castle, continent int) int {
	count := 0
	for _, c := range castles {
		if c.Continent == continent {
			count++
		}
	}
	return count
}
