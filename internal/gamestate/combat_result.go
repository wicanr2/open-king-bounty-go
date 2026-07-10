package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// peasantTroopID 是 C temp_death 的「魔術數字」——隊伍第 0 格固定給 20 名農夫
// (play.c:1024 註解 MAGIC NUMBER for "Peasants";troops[0] = 農夫)。
const peasantTroopID = 0

// FulfillContract 履行對 villainID 的懸賞契約,對齊 C fullfill_contract(play.c:543):
// 領賞金、標記已捕獲、清空當前契約,並把該惡棍從可接契約循環移除、補入下一個未捕獲的
// 惡棍(從 max_contract 起找,並讓 max_contract++)。賞金取自 villains.ini 的 reward 欄
// (對齊 C villain_rewards[]/WDAT_VREWARD)。
func (gs *GameState) FulfillContract(a *kbdata.Assets, villainID int) {
	if villainID < 0 || villainID >= len(gs.VillainCaught) {
		return
	}
	villains := LoadVillains(a)
	if villainID < len(villains) {
		gs.Gold += villains[villainID].Reward
	}
	gs.VillainCaught[villainID] = 1
	gs.Contract = 0xFF

	// 從契約循環找到該惡棍的槽位並清空。
	slot := -1
	for i := 0; i < 5; i++ {
		if int(gs.ContractCycle[i]) == villainID {
			slot = i
		}
	}
	if slot == -1 {
		return
	}
	gs.ContractCycle[slot] = 0xFF

	// 從 max_contract 起找下一個未捕獲的惡棍補進該槽位,並讓 max_contract++。
	for i := int(gs.MaxContract); i < kbdata.MaxVillains; i++ {
		if gs.VillainCaught[i] != 0 {
			continue
		}
		gs.ContractCycle[slot] = byte(i)
		break
	}
	gs.MaxContract++
}

// TempDeath 對齊 C temp_death(play.c:1005):戰敗懲罰——清空隊伍、發還 20 名農夫。
//
// ⚠ 偏離(架構限制,誠實標註):C 版還會把玩家傳送回家鄉(special_coords[SP_HOME])
// 並清除攻城武器。本移植的玩家地圖座標仍存在 WorldMapScreen(cont/px/py)而非
// GameState,siege_weapons 亦尚未建模,故 TempDeath 只還原「隊伍」這塊 GameState
// 狀態;傳送回家由呼叫端(戰鬥畫面戰敗 → 重建世界地圖畫面)另行處理,見
// PORT-STATUS「玩家座標移入 GameState」待辦。
func (gs *GameState) TempDeath(a *kbdata.Assets) {
	for i := 0; i < 5; i++ {
		gs.Army[i] = Squad{TroopID: 0xFF, Count: 0}
	}
	gs.Army[0] = Squad{TroopID: peasantTroopID, Count: 20}
}
