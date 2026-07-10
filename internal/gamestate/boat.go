package gamestate

// 座騎狀態,對齊 C bounty.h:30-32 KBMOUNT_*。玩家在陸地為 RIDE、在水上為 SAIL。
const (
	KBMountSail = 0 // 乘船(水上)
	KBMountFly  = 4 // 飛行(時間術等,暫未用)
	KBMountRide = 8 // 騎乘(陸地,起手預設)
)

// 租船費用,對齊 C bounty.h:26-27。
const (
	CostBoatCheap     = 100 // COST_BOAT_CHEAP(有 POWER_CHEAPER_BOAT_RENTAL 神器時)
	CostBoatExpensive = 500 // COST_BOAT_EXPENSIVE(一般)
)

// noBoat 是「未租船」的哨兵值,對齊 C game->boat == 0xFF。
const noBoat = 0xFF

// BoatCost 回傳目前租船費用,對齊 C boat_cost(play.c:681):有「降低租船費」神器時
// 為 CostBoatCheap,否則 CostBoatExpensive。
//
// ⚠ 神器(artifact)世界狀態尚未建模(見 PORT-STATUS 的 artifact/scepter 待辦),
// 故 has_power(POWER_CHEAPER_BOAT_RENTAL) 目前恆 false,一律回 CostBoatExpensive。
// 待 artifact_found[] 建模後改讀真值。
func (gs *GameState) BoatCost() int {
	return CostBoatExpensive
}

// HasBoat 回報玩家目前是否已租船,對齊 C player_has_boat(play.c:688)。
func (gs *GameState) HasBoat() bool {
	return gs.Boat != noBoat
}

// 租船結果,供 UI 顯示對應訊息。
const (
	RentOK           = iota // 成功租船
	RentNotEnoughGold       // 金幣不足
	RentMustDisembark       // 已有船且正在船上,需先離船才能取消
	RentCancelled           // 取消租船
)

// RentOrCancelBoat 對齊 C visit_town key==2(game.c:2775-2800):
//   - 未租船:金幣足則扣費、把船停到該鎮的 boat 停靠點(town.BoatX/Y,對齊 boat_coords),
//     boat = 該鎮所在洲;金幣不足回 RentNotEnoughGold。
//   - 已租船:若正乘船(Mount==SAIL)需先離船(回 RentMustDisembark);否則取消(boat=0xFF)。
func (gs *GameState) RentOrCancelBoat(town Town) int {
	if !gs.HasBoat() {
		cost := gs.BoatCost()
		if gs.Gold <= cost {
			return RentNotEnoughGold
		}
		gs.Gold -= cost
		gs.Boat = byte(town.Continent)
		gs.BoatX = town.BoatX
		gs.BoatY = town.BoatY
		return RentOK
	}
	if gs.Mount == KBMountSail {
		return RentMustDisembark
	}
	gs.Boat = noBoat
	return RentCancelled
}
