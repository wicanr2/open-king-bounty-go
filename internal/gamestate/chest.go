package gamestate

// AddChestLeadership 套用寶箱的領導力加成,對應 C 版 game.c 的寶箱處理
// (game->base_leadership += leadership; game->leadership += leadership;)。
//
// 關鍵(issue #5):必須同時加進 BaseLeadership 與 Leadership。只加 Leadership 的話,
// 下一次 EndWeek 會把 Leadership 重設回 BaseLeadership,加成當週就流失。加進 BaseLeadership
// 才是永久加成;同步加 Leadership 讓當週立即生效。
func (gs *GameState) AddChestLeadership(bonus int) {
	gs.BaseLeadership += bonus
	gs.Leadership += bonus
}
