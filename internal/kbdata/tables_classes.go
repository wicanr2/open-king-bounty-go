package kbdata

// classTable 回傳四職業 × 四階表,移植自 C 版 src/bounty.c 的 classes[4][4]。
//
// 16 筆(4 職業 × 4 階:騎士/遊俠/女巫師/蠻俠)。欄位對齊 KBclass(src/bounty.h)。
// C 版把 Leadership/MaxSpell/SpellPower/Commission 寫成逐階「+增量」,此處存累積後絕對值;
// VillainsNeeded/KnowsMagic/InstantArmy 在 C 本就是絕對值,照抄。
func classTable() [4][4]Class {
	var t [4][4]Class
	// Row 0: 騎士系列
	t[0][0] = Class{Name: "武士", VillainsNeeded: 0, Leadership: 100, MaxSpell: 2, SpellPower: 1, Commission: 1000, KnowsMagic: 0, InstantArmy: 0x00}
	t[0][1] = Class{Name: "將軍", VillainsNeeded: 2, Leadership: 200, MaxSpell: 5, SpellPower: 2, Commission: 2000, KnowsMagic: 0, InstantArmy: 0x02}
	t[0][2] = Class{Name: "元帥", VillainsNeeded: 8, Leadership: 500, MaxSpell: 9, SpellPower: 3, Commission: 4000, KnowsMagic: 0, InstantArmy: 0x08}
	t[0][3] = Class{Name: "領主", VillainsNeeded: 14, Leadership: 1000, MaxSpell: 14, SpellPower: 5, Commission: 8000, KnowsMagic: 0, InstantArmy: 0x0E}

	// Row 1: 遊俠系列
	t[1][0] = Class{Name: "遊俠", VillainsNeeded: 0, Leadership: 80, MaxSpell: 3, SpellPower: 1, Commission: 1000, KnowsMagic: 0, InstantArmy: 0x00}
	t[1][1] = Class{Name: "聖戰士", VillainsNeeded: 2, Leadership: 160, MaxSpell: 6, SpellPower: 2, Commission: 2000, KnowsMagic: 0, InstantArmy: 0x02}
	t[1][2] = Class{Name: "復仇者", VillainsNeeded: 7, Leadership: 400, MaxSpell: 12, SpellPower: 4, Commission: 4000, KnowsMagic: 0, InstantArmy: 0x08}
	t[1][3] = Class{Name: "勇士", VillainsNeeded: 13, Leadership: 800, MaxSpell: 17, SpellPower: 6, Commission: 8000, KnowsMagic: 0, InstantArmy: 0x12}

	// Row 2: 女巫師系列
	t[2][0] = Class{Name: "女巫師", VillainsNeeded: 0, Leadership: 60, MaxSpell: 5, SpellPower: 2, Commission: 3000, KnowsMagic: 1, InstantArmy: 0x01}
	t[2][1] = Class{Name: "魔術師", VillainsNeeded: 3, Leadership: 120, MaxSpell: 13, SpellPower: 5, Commission: 4000, KnowsMagic: 0, InstantArmy: 0x06}
	t[2][2] = Class{Name: "法師", VillainsNeeded: 6, Leadership: 300, MaxSpell: 23, SpellPower: 10, Commission: 5000, KnowsMagic: 0, InstantArmy: 0x09}
	t[2][3] = Class{Name: "大法師", VillainsNeeded: 12, Leadership: 600, MaxSpell: 35, SpellPower: 15, Commission: 6000, KnowsMagic: 0, InstantArmy: 0x13}

	// Row 3: 蠻俠系列
	t[3][0] = Class{Name: "蠻俠", VillainsNeeded: 0, Leadership: 100, MaxSpell: 2, SpellPower: 0, Commission: 2000, KnowsMagic: 0, InstantArmy: 0x00}
	t[3][1] = Class{Name: "酋長", VillainsNeeded: 1, Leadership: 200, MaxSpell: 4, SpellPower: 1, Commission: 4000, KnowsMagic: 0, InstantArmy: 0x03}
	t[3][2] = Class{Name: "戰王", VillainsNeeded: 5, Leadership: 500, MaxSpell: 7, SpellPower: 2, Commission: 6000, KnowsMagic: 0, InstantArmy: 0x07}
	t[3][3] = Class{Name: "霸主", VillainsNeeded: 10, Leadership: 1000, MaxSpell: 10, SpellPower: 3, Commission: 8000, KnowsMagic: 0, InstantArmy: 0x0F}

	return t
}
