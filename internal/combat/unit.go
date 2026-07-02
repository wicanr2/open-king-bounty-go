package combat

// Unit 對應 C 版 KBunit(src/bounty.h:257-279)——棋盤上單一格戰鬥單位的執行期狀態。
// C 的 byte/word 欄位在 Go 一律用 int 表示,純旗標(0/1)用 bool。
type Unit struct {
	TroopID int // 兵種 id(索引 kbdata.Assets.Troops);C: byte troop_id

	Count     int // 目前隊伍人數(受傷會扣);C: word count
	TurnCount int // 本回合開始前的人數,受傷結算的基準;C: word turn_count
	MaxCount  int // 開戰前的人數上限;C: word max_count

	Dead  bool // 整組陣亡;C: byte dead
	Frame int  // 顯示用動畫格;C: byte frame

	Injury int // 累積傷害餘數(未達一隻份量的傷害);C: word injury

	Acted      bool // 本回合已行動過;C: byte acted
	Retaliated bool // 本回合已反擊過;C: byte retaliated

	Moves   int // 剩餘移動力(單步king-move,每步耗 1);C: byte moves
	Shots   int // 剩餘遠攻彈藥;C: byte shots
	Flights int // 剩餘飛行點數;C: byte flights

	Frozen       bool // 被冰凍,略過本回合行動;C: byte frozen
	OutOfControl bool // 超出領導力上限、脫離掌控;C: byte out_of_control

	X int // 棋盤欄座標,[0, BoardW);C: byte x
	Y int // 棋盤列座標,[0, BoardH);C: byte y
}
