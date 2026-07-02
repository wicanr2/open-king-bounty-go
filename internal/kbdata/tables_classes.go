package kbdata

// classTable 回傳四職業 × 四階表,移植自 C 版 src/bounty.c 的 classes[4][4]。
//
// ⚠ 語意(對齊 C `player_accept_rank`,play.c:760):
//   - Leadership / MaxSpell / SpellPower / Commission / KnowsMagic 存**逐階增量**,
//     由 gamestate.acceptRank 用 += 累加(C 版 classes[][] 就是增量,值前有 '+')。
//     這樣 base_leadership 能同時累積「升階增量」與「寶箱加成」等多來源,不會互相清掉。
//   - VillainsNeeded / InstantArmy 存**該階絕對值**(C 版直接查用、不累加)。
// 這些數字即 bounty.c 的字面值,不做累積轉換(C 為真值,保持一致)。
func classTable() [4][4]Class {
	var t [4][4]Class
	// Row 0: 騎士系列
	t[0][0] = Class{Name: "武士", VillainsNeeded: 0, Leadership: 100, MaxSpell: 2, SpellPower: 1, Commission: 1000, KnowsMagic: 0, InstantArmy: 0x00}
	t[0][1] = Class{Name: "將軍", VillainsNeeded: 2, Leadership: 100, MaxSpell: 3, SpellPower: 1, Commission: 1000, KnowsMagic: 0, InstantArmy: 0x02}
	t[0][2] = Class{Name: "元帥", VillainsNeeded: 8, Leadership: 300, MaxSpell: 4, SpellPower: 1, Commission: 2000, KnowsMagic: 0, InstantArmy: 0x08}
	t[0][3] = Class{Name: "領主", VillainsNeeded: 14, Leadership: 500, MaxSpell: 5, SpellPower: 2, Commission: 4000, KnowsMagic: 0, InstantArmy: 0x0E}

	// Row 1: 遊俠系列
	t[1][0] = Class{Name: "遊俠", VillainsNeeded: 0, Leadership: 80, MaxSpell: 3, SpellPower: 1, Commission: 1000, KnowsMagic: 0, InstantArmy: 0x00}
	t[1][1] = Class{Name: "聖戰士", VillainsNeeded: 2, Leadership: 80, MaxSpell: 3, SpellPower: 1, Commission: 1000, KnowsMagic: 0, InstantArmy: 0x02}
	t[1][2] = Class{Name: "復仇者", VillainsNeeded: 7, Leadership: 240, MaxSpell: 6, SpellPower: 2, Commission: 2000, KnowsMagic: 0, InstantArmy: 0x08}
	t[1][3] = Class{Name: "勇士", VillainsNeeded: 13, Leadership: 400, MaxSpell: 5, SpellPower: 2, Commission: 4000, KnowsMagic: 0, InstantArmy: 0x12}

	// Row 2: 女巫師系列
	t[2][0] = Class{Name: "女巫師", VillainsNeeded: 0, Leadership: 60, MaxSpell: 5, SpellPower: 2, Commission: 3000, KnowsMagic: 1, InstantArmy: 0x01}
	t[2][1] = Class{Name: "魔術師", VillainsNeeded: 3, Leadership: 60, MaxSpell: 8, SpellPower: 3, Commission: 1000, KnowsMagic: 0, InstantArmy: 0x06}
	t[2][2] = Class{Name: "法師", VillainsNeeded: 6, Leadership: 180, MaxSpell: 10, SpellPower: 5, Commission: 1000, KnowsMagic: 0, InstantArmy: 0x09}
	t[2][3] = Class{Name: "大法師", VillainsNeeded: 12, Leadership: 300, MaxSpell: 12, SpellPower: 5, Commission: 1000, KnowsMagic: 0, InstantArmy: 0x13}

	// Row 3: 蠻俠系列
	t[3][0] = Class{Name: "蠻俠", VillainsNeeded: 0, Leadership: 100, MaxSpell: 2, SpellPower: 0, Commission: 2000, KnowsMagic: 0, InstantArmy: 0x00}
	t[3][1] = Class{Name: "酋長", VillainsNeeded: 1, Leadership: 100, MaxSpell: 2, SpellPower: 1, Commission: 2000, KnowsMagic: 0, InstantArmy: 0x03}
	t[3][2] = Class{Name: "戰王", VillainsNeeded: 5, Leadership: 300, MaxSpell: 3, SpellPower: 1, Commission: 2000, KnowsMagic: 0, InstantArmy: 0x07}
	t[3][3] = Class{Name: "霸主", VillainsNeeded: 10, Leadership: 500, MaxSpell: 3, SpellPower: 1, Commission: 2000, KnowsMagic: 0, InstantArmy: 0x0F}

	return t
}
