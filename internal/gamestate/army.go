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
	max := gs.ArmyLeadership(a, troopID) / hp

	// opt_recruit_caps 開啟:高階兵種可招募數夾到「上限 - 目前已持有」(對齊 C
	// army_max_troop_count,play.c:614-635)。上限值照抄 C:法師50/火龍15/惡魔20/吸血鬼20/武士20。
	if gs.OptRecruitCaps {
		cap := recruitCapFor(troopID)
		if cap >= 0 {
			held := 0
			for i := 0; i < 5; i++ {
				if gs.Army[i].TroopID == troopID {
					held += gs.Army[i].Count
				}
			}
			room := cap - held
			if room < 0 {
				room = 0
			}
			if max > room {
				max = room
			}
		}
	}
	return max
}

// recruitCapFor 回傳 opt_recruit_caps 下該高階兵種的持有上限(-1 = 不受限),
// 對齊 C army_max_troop_count 的 switch(play.c:614-620)。
func recruitCapFor(troopID int) int {
	switch troopID {
	case 0x14: // Archmage 法師
		return 50
	case 0x18: // Dragon 火龍
		return 15
	case 0x17: // Demon 惡魔
		return 20
	case 0x15: // Vampire 吸血鬼
		return 20
	case 0x0E: // Knight 武士
		return 20
	}
	return -1
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
