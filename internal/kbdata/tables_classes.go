package kbdata

// classTable 回傳四職業 × 四難度表,移植自 C 版 src/bounty.c 的 classes[4][4]。
//
// ⚠ P0 佔位:僅放一筆讓模組編得過。
// P1 由移植 subagent 依 src/bounty.h 的 KBclass 定義補齊全部欄位與 16 筆。
func classTable() [4][4]Class {
	var t [4][4]Class
	t[0][0] = Class{Name: "騎士", Leadership: 100}
	// TODO(P1): 補齊 4×4(移植 subagent)。
	return t
}
