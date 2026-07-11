package combat

import "github.com/wicanr2/open-king-bounty-go/internal/gamestate"

// 棋盤常量,對應 C 版 src/bounty.h:249-253。
const (
	BoardW   = 6 // CLEVEL_W:棋盤欄數
	BoardH   = 5 // CLEVEL_H:棋盤列數
	MaxSides = 2 // MAX_SIDES:玩家/敵方兩側
	MaxUnits = 5 // MAX_UNITS:每側最多 5 格單位
)

// Combat 對應 C 版 KBcombat(src/bounty.h:281-300)——一場戰鬥的完整狀態容器。
//
// Omap/Umap 的陣列尺寸沿用 C 版宣告的 [H+1][W+1](多一格 padding);
// 實際可站立/可通行的棋盤範圍仍是 [0,BoardW) x [0,BoardH),存取一律用
// InBounds 先判斷再用座標索引。
type Combat struct {
	Units [MaxSides][MaxUnits]Unit // C: KBunit units[MAX_SIDES][MAX_UNITS]

	Omap [BoardH + 1][BoardW + 1]byte // 障礙圖(非 0 = 該格有障礙);C: byte omap[CLEVEL_H+1][CLEVEL_W+1]
	Umap [BoardH + 1][BoardW + 1]byte // 單位佔格圖(值 = PackUID,0 = 空);C: byte umap[CLEVEL_H+1][CLEVEL_W+1]

	Spoils [MaxSides]int  // 該側目前可獲得的戰利品金;C: word spoils[MAX_SIDES]
	Powers [MaxSides]byte // 該側生效的神器加成位元(POWER_*);C: byte powers[MAX_SIDES]
	// TODO(flagship): Powers 目前恆為 0——gamestate 尚未移植 artifact_found 世界狀態
	// (見 internal/gamestate/gamestate.go 開頭註解),等旗艦補上後再對回 POWER_* 效果。

	Heroes [MaxSides]*gamestate.GameState // 該側對應的玩家角色;敵方/城堡側為 nil。C: KBgame *heroes[MAX_SIDES]

	YourTurn bool // C: int your_turn(C 原註解:TODO 該換成 macro)

	Turn  int // 回合計數(C 原註解:目前沒真的用到);C: byte turn
	Phase int // 每次玩家/AI 回合最多兩個 phase;C: byte phase

	Spells int // 本回合已施放法術數;C: byte spells

	Side   int // 目前行動方,索引 Units;C: byte side
	UnitID int // 目前行動單位,Units[Side][UnitID];C: byte unit_id

	// AIEvolved:true 時 AI 走進化版決策(對齊 C opt_ai_mode → ai_unit_think_evolved,
	// game.c:6507)。由建 combat 時從 gs.OptAIMode 設定;預設 false = 原版 AI。
	AIEvolved bool
}

// InBounds 回傳 (x,y) 是否落在可通行棋盤範圍內([0,BoardW) x [0,BoardH))。
func InBounds(x, y int) bool {
	return x >= 0 && x < BoardW && y >= 0 && y < BoardH
}

// Obstacle 回傳 (x,y) 是否有障礙物(C: war->omap[y][x] 非 0)。
// 呼叫前需自行確保 InBounds(x,y),否則可能索引越界。
func (c *Combat) Obstacle(x, y int) bool {
	return c.Omap[y][x] != 0
}

// Occupied 回傳 (x,y) 是否已被某單位佔據(C: war->umap[y][x] 非 0)。
// 呼叫前需自行確保 InBounds(x,y),否則可能索引越界。
func (c *Combat) Occupied(x, y int) bool {
	return c.Umap[y][x] != 0
}

// PackUID 把 (side,id) 編碼成 umap 的儲存值,對應 C 版巨集
// PACK_UID(SIDE,ID)(src/play.c:26、src/game.c:3475):SIDE*MaxUnits + ID + 1。
func PackUID(side, id int) byte {
	return byte(side*MaxUnits + id + 1)
}

// UnpackUID 把 umap 的儲存值解回 (side,id),對應 C 版巨集
// UID_AS_SIDE/UID_AS_ID(src/play.c:24-25):以 uid > MaxUnits 判斷是否屬於 side 1。
func UnpackUID(uid byte) (side, id int) {
	if int(uid) > MaxUnits {
		return 1, int(uid) - MaxUnits - 1
	}
	return 0, int(uid) - 1
}

// Adjacent 回傳兩格是否「相鄰」(含對角、含同格本身),對應 C 版 unit_touching
// (src/play.c:1332-1346):Chebyshev 距離 <= 1(|dx|<=1 且 |dy|<=1)。
func Adjacent(x1, y1, x2, y2 int) bool {
	return abs(x1-x2) <= 1 && abs(y1-y2) <= 1
}

// Distance 回傳兩格的曼哈頓距離。
//
// 對應 C 版 calc_distance 巨集「實際生效」的那一版(src/play.c:1641-1645):
// 檔案裡先定義了一版 isqrt32 的畢氏距離(1642 行),但緊接著被 #undef 並重定義成
// abs(dx)+abs(dy)(1645 行)——後者才是真正編譯進去、被 unit_closest_offset 等函式
// 使用的版本(見 rulebook 62:描述一段碼前先確認它真的被執行/是不是死碼)。
// 移植以「實際生效」版為準,故此為曼哈頓距離,不是歐幾里得距離。
func Distance(x1, y1, x2, y2 int) int {
	return abs(x1-x2) + abs(y1-y2)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
