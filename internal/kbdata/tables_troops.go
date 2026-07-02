package kbdata

// troopTable 回傳全兵種靜態表,移植自 C 版 src/bounty.c 的 troops[MAX_TROOPS]。
//
// ⚠ P0 佔位:此處僅放前幾筆讓模組編得過。
// P1 由移植 subagent 依 src/bounty.c 補齊全部 MAX_TROOPS 筆,並加對照測試。
func troopTable() []Troop {
	return []Troop{
		{Name: "農夫", SkillLevel: 1, HP: 1, Move: 1, MeleeMin: 1, MeleeMax: 1,
			GoldCost: 10, Spoils: 1, Home: DwellPlains, Growth: 250, MoraleTop: 6, MoraleGroup: 0},
		{Name: "小妖精", SkillLevel: 1, HP: 1, Move: 1, MeleeMin: 1, MeleeMax: 2,
			GoldCost: 15, Spoils: 1, Abilities: AbilFly, Home: DwellForest, Growth: 200, MoraleTop: 6, MoraleGroup: 2},
		// TODO(P1): 補齊其餘兵種(移植 subagent)。
	}
}
