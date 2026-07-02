package gamestate

import (
	"errors"

	"github.com/wicanr2/open-king-bounty-go/internal/kbdata"
)

var (
	// ErrNotEnoughGold:金錢不足以招募(對應 C buy_troop 回傳 1)。
	ErrNotEnoughGold = errors.New("gold 不足")
	// ErrNoSlot:隊伍五格已滿且無同兵種可合併(對應 C buy_troop 回傳 2 / add_troop 回傳 1)。
	ErrNoSlot = errors.New("隊伍無空格")
)

// ArmyLeadership 回傳「還能再放多少該兵種」的剩餘領導力,對應 C army_leadership(play.c:590):
// 目前 Leadership 扣掉「該兵種現有數量 × 每隻 HP」,下限 0。
func (gs *GameState) ArmyLeadership(a *kbdata.Assets, troopID int) int {
	free := gs.Leadership
	for i := 0; i < 5; i++ {
		if gs.Army[i].TroopID == troopID {
			free -= a.Troops[troopID].HP * gs.Army[i].Count
			break
		}
	}
	if free < 0 {
		free = 0
	}
	return free
}

// ArmyMaxTroopCount 回傳領導力允許再招的最大隻數,對應 C army_max_troop_count(play.c:607)。
func (gs *GameState) ArmyMaxTroopCount(a *kbdata.Assets, troopID int) int {
	hp := a.Troops[troopID].HP
	if hp <= 0 {
		return 0
	}
	return gs.ArmyLeadership(a, troopID) / hp
}

// addTroop 把 number 隻某兵種加進隊伍,對應 C add_troop(play.c:863):
// 找「同兵種格」或「第一個空格(Count==0)」;都沒有回 ErrNoSlot。不檢查金錢/領導力。
func (gs *GameState) addTroop(troopID, number int) error {
	slot := -1
	for i := 0; i < 5; i++ {
		if gs.Army[i].TroopID == troopID {
			slot = i
			break
		}
		if gs.Army[i].Count == 0 {
			slot = i
			break
		}
	}
	if slot == -1 {
		return ErrNoSlot
	}
	gs.Army[slot].TroopID = troopID
	gs.Army[slot].Count += number
	return nil
}

// BuyTroop 招募 number 隻某兵種,對應 C buy_troop(play.c:887):
// 先檢查金錢(cost = recruit_cost × number),加兵成功才扣金。
// 注意:與 C 一致,BuyTroop 本身**不**檢查領導力上限;呼叫端(UI)應先用 ArmyMaxTroopCount
// 夾住 number(C 版也是由 game.c 的招募畫面限制,buy_troop 只管金錢與格子)。
func (gs *GameState) BuyTroop(a *kbdata.Assets, troopID, number int) error {
	cost := a.Troops[troopID].GoldCost * number
	if cost > gs.Gold {
		return ErrNotEnoughGold
	}
	if err := gs.addTroop(troopID, number); err != nil {
		return err
	}
	gs.Gold -= cost
	return nil
}
