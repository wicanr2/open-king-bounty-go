// continent.go -- 洲名查詢,對齊 C 版 continent_names[MAX_CONTINENTS][32]「實際顯示值」。
package gamestate

import "github.com/wicanr2/open-king-bounty-go/internal/kbdata"

// ContinentNames 回傳 4 洲的顯示名稱,來源 data/free/land.ini 的 continent0-3
// name 欄位(經 kbdata.FreeIni 解析)。缺檔或缺區塊時該筆維持空字串,不 panic。
//
// ⚠ 為什麼不用 bounty.c 硬編的 continent_names[4][32] 字面值("大陸洲"/"森林洲"/
// "群島洲"/"撒哈洲"):view_contract 移植時逐句核對 C 版執行路徑(rulebook 62 靜態
// 溯源)發現 —— game.c:1699 讀的 continent_names[] 全域陣列,在**每次開新遊戲後**
// 都會被 refill_names()(game.c:221,呼叫點 game.c:490 create_game 結尾、
// game.c:700/7116/7145)整個覆寫:
//
//	char *continent_list = KB_Resolve(STRL_CONTINENTS, 0);   // free-data.c:1155
//	                                                          // = land.ini continent%d/name
//	for (i = 0; i < MAX_CONTINENTS; i++)
//	    if (continent_list) KB_strcpy(continent_names[i], KB_strlist_peek(continent_list, i));
//
// 也就是說 bounty.c 那組硬編字面值是**從未在 free 模組正常遊戲流程中被實際顯示過
// 的死預設值**——refill_names 在任何畫面看到 continent_names 之前就已經跑過。
// 實際顯示的是 land.ini 的洲名("弗蘭德利亞"/"第二洲"/"第三洲"/"第四洲",這幾個
// 名字本身像是 land.tmx 匯出工具留下的預設佔位文字,但那是資料檔案本身的內容,
// 不影響「執行期實際讀哪個來源」這個判斷)。故本移植對齊實際執行結果,讀 land.ini。
func ContinentNames(a *kbdata.Assets) [kbdata.MaxContinents]string {
	var names [kbdata.MaxContinents]string
	if a == nil {
		return names
	}
	ini := a.Strings["land"]
	if ini == nil {
		return names
	}
	for i := 0; i < kbdata.MaxContinents; i++ {
		sec, ok := ini.NumberedSection("continent", i)
		if !ok {
			continue
		}
		names[i] = sec.GetDefault("name", "")
	}
	return names
}
