// free_ini.go -- free 模組(GNU family)*.ini 解析器。
//
// C 版對照(逐讀法移植,見 openkb-code/src/lib/free-data.c):
//   - GNU_string_ini()  第 266-341 行:掃描 "[section]" 標頭,section 內找
//     "key = value" 行,值以 sscanf 格式 "%s = %%[^\t\n;]" 擷取
//     (=後略過空白/tab,遇 tab 或 ';' 就截斷,不裁掉結尾空白)。
//   - GNU_extract_ini() 第 343-451 行:同樣的 section/key 掃描,但數值改用
//     "%s = %%d"(十進位)失敗再退 "%s = #%%6c" + hex2dec()(六碼十六進位,
//     見 kbstd.c:508 hex2dec ── 首字元非合法十六進位就回 0)。
//   - GNU_parse_troop() 第 173-197 行:villains.ini 的 armyN 欄位是
//     "N x 兵種名" 這種迷你格式,五種 sscanf 變體的聯集其實等於
//     「數字、空白、可有可無的 'x'、空白、非空白 token(兵種名)」。
//
// 本檔不是逐行搬 C 版的「累積計數 + 跨 section 全域比對」演算法 ── 那套演算法
// 有個真實存在的副作用 bug,見下方 §已知 C 版怪異行為。這裡改用標準 ini 語意:
// 每個 section 各自維護自己的 key=value map,同一 section 內重複 key 後者覆寫前者、
// 用 section 名稱直接索引(不會因為別的 section 缺欄位而錯位)。這是刻意的設計選擇,
// 不是誤讀 C ── 細節見檔尾大註解。
//
// § 已知 C 版怪異行為(反讀時發現,回報用,不在此複製):
//   data/free/spells.ini 第 36-42 行 [spell6] 區塊裡 "gold" 這個 key 出現兩次
//   (gold = 250 之後又被 gold = 2000 蓋過)。C 版 GNU_extract_ini 用一個「跨整個
//   檔案累加的 filled 計數器」決定何時提早跳出整個 while 迴圈(filled >= num-first
//   就 break,不分是在哪個 section 比對到的)。多出來的這一次重複比對,會讓後面
//   查 "spell13"(增控術)的 gold 欄位時,迴圈在讀到 [spell13] 之前就已經
//   提前 break ── 導致 spell13 的 gold cost 實際上讀到記憶體預設值 0,
//   而不是檔案裡寫的 500。這看起來是資料檔案本身的複製貼上筆誤(不像是刻意的
//   遊戲設計),但因為 C 版當前程式碼的實際行為確實會把它讀成 0,所以在此明確
//   標註、交給後續決策:要不要在 Go 版刻意複製這個「省 token」bug 以求兩邊
//   行為位元對位元一致,還是視為資料檔案錯字予以修正(本檔目前的 Section/Get
//   設計預設是「修正」── 每個 section 用自己的名字查,不會被別的 section 拖累)。
package kbdata

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// FreeIniSection 是解析後的單一 "[xxx]" 區塊。
type FreeIniSection struct {
	Name string            // 區塊名稱,原樣保留(如 "troop0"、"top"、"minimap")
	Keys map[string]string // key 一律轉小寫存放(比對時大小寫不敏感,對齊 C 版 strncasecmp)
}

// FreeIni 是整份 *.ini 檔解析後的結果。
type FreeIni struct {
	Sections []FreeIniSection
	index    map[string]int // strings.ToLower(section name) -> Sections 的 index
}

// newFreeIni 建立一個空的 FreeIni。
func newFreeIni() *FreeIni {
	return &FreeIni{index: make(map[string]int)}
}

// ParseFreeIni 解析 ini 檔內容(UTF-8 位元組),回傳 *FreeIni。
//
// 語法規則(對齊 free-data.c 的實際讀法,細節見檔頭大註解):
//   - 空行、以 ';' 開頭的行(去除前導空白後)視為註解,略過。
//   - "[NAME]" 開新區塊,NAME 是中括號內原文(不含中括號)。
//   - 區塊內 "key = value" 以第一個 '=' 切分;key 去除前後空白後轉小寫存放。
//     value 先吃掉開頭空白/tab,再遇到 tab 或 ';' 就截斷(對齊 "%[^\t\n;]"),
//     結尾空白不特別裁切(和 C 版一致 ── 目前六個目標檔案裡只有 spells.ini
//     "gold = 100 " 這一行結尾多一個空白,已知情況,呼叫端用 Int() 讀取時
//     不受影響)。
//   - 區塊標頭出現之前的雜散行(理論上不該出現,C 版此時 section=-1 也會跳過)一律忽略。
//   - 同一區塊內重複 key:後者蓋過前者(標準 ini 語意 ── 對照 C 版原本用
//     「全域累加計數器」決定何時 break 的演算法不同,差異與理由見檔頭大註解)。
func ParseFreeIni(data []byte) *FreeIni {
	f := newFreeIni()
	var cur *FreeIniSection

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 4096), 1<<20)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if trimmed[0] == ';' {
			continue
		}
		if trimmed[0] == '[' {
			end := strings.IndexByte(trimmed, ']')
			if end > 0 {
				name := trimmed[1:end]
				f.Sections = append(f.Sections, FreeIniSection{
					Name: name,
					Keys: make(map[string]string),
				})
				idx := len(f.Sections) - 1
				f.index[strings.ToLower(name)] = idx
				cur = &f.Sections[idx]
			}
			continue
		}
		if cur == nil {
			continue // 區塊標頭之前的雜散行,對齊 C 版 section<first 時跳過
		}
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		if key == "" {
			continue
		}
		val := line[eq+1:]
		val = strings.TrimLeft(val, " \t") // 對齊 "= %[^...]" 中,'=' 後的空白由 scanf 略過
		if cut := strings.IndexAny(val, "\t;"); cut >= 0 {
			val = val[:cut] // 對齊 "%[^\t\n;]":遇 tab 或 ';' 截斷
		}
		cur.Keys[strings.ToLower(key)] = val
	}
	return f
}

// LoadFreeIni 從路徑讀檔並解析。
func LoadFreeIni(path string) (*FreeIni, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("kbdata: LoadFreeIni %s: %w", path, err)
	}
	return ParseFreeIni(data), nil
}

// Section 依名稱(大小寫不敏感)取回一個區塊。
func (f *FreeIni) Section(name string) (*FreeIniSection, bool) {
	if f == nil {
		return nil, false
	}
	idx, ok := f.index[strings.ToLower(name)]
	if !ok {
		return nil, false
	}
	return &f.Sections[idx], true
}

// NumberedSection 取回如 "troop3"、"villain0"、"spell13" 這類 prefix+索引 命名的區塊,
// 對齊 C 版 GNU_Resolve 呼叫時慣用的 "troop%d"/"villain%d"/... 型式(first 恆為 0)。
func (f *FreeIni) NumberedSection(prefix string, idx int) (*FreeIniSection, bool) {
	return f.Section(fmt.Sprintf("%s%d", prefix, idx))
}

// Get 取區塊內某 key 的原始字串值(已依規則裁切,見 ParseFreeIni 說明)。
func (s *FreeIniSection) Get(key string) (string, bool) {
	if s == nil {
		return "", false
	}
	v, ok := s.Keys[strings.ToLower(key)]
	return v, ok
}

// GetDefault 同 Get,但找不到 key 時回傳 def。
func (s *FreeIniSection) GetDefault(key, def string) string {
	if v, ok := s.Get(key); ok {
		return v
	}
	return def
}

// Int 取數值鍵:先試十進位(對齊 C 版 "name = %d"),失敗再試
// "#RRGGBB" 十六進位色碼(對齊 GNU_extract_ini 的 hex2dec 退路,見
// kbstd.c:508 ── 只看前 6 碼、首字元非合法十六進位就回 0)。
// 兩者都失敗(key 不存在,或值兩種格式都不符)回 (0, false)。
func (s *FreeIniSection) Int(key string) (int, bool) {
	v, ok := s.Get(key)
	if !ok {
		return 0, false
	}
	v = strings.TrimSpace(v)
	if n, ok2 := parseCDecimalPrefix(v); ok2 {
		return n, true
	}
	if strings.HasPrefix(v, "#") {
		hexpart := v[1:]
		if len(hexpart) >= 6 {
			hexpart = hexpart[:6]
			if isHexDigit(hexpart[0]) {
				return parseHexPrefix(hexpart), true
			}
			// 對齊 C 版:sscanf("%6c") 仍算比對成功(消耗 6 字元、回傳 1),
			// 只是 hex2dec() 因首字元不合法而回 0 ── 「有找到,值是 0」。
			return 0, true
		}
	}
	return 0, false
}

// IntDefault 同 Int,但找不到/解析失敗時回傳 def。
func (s *FreeIniSection) IntDefault(key string, def int) int {
	if n, ok := s.Int(key); ok {
		return n
	}
	return def
}

// StringSlice 依序取回 prefix+0..prefix+(count-1) 各區塊裡某 key 的字串值。
// 缺區塊或缺 key 一律回填空字串,以保持索引對齊(刻意不同於 C 版
// GNU_string_ini 的「附加式 strlist」── 那個版本一旦某個 section 缺這個 key,
// 後面所有結果都會往前錯位;此處改用索引直接定位,更安全,詳見檔頭大註解)。
func (f *FreeIni) StringSlice(prefix, key string, count int) []string {
	out := make([]string, count)
	for i := 0; i < count; i++ {
		sec, ok := f.NumberedSection(prefix, i)
		if !ok {
			continue
		}
		if v, ok := sec.Get(key); ok {
			out[i] = v
		}
	}
	return out
}

// IntSlice 同 StringSlice,但取數值(見 Int 的十進位/十六進位規則)。
func (f *FreeIni) IntSlice(prefix, key string, count int) []int {
	out := make([]int, count)
	for i := 0; i < count; i++ {
		sec, ok := f.NumberedSection(prefix, i)
		if !ok {
			continue
		}
		if n, ok := sec.Int(key); ok {
			out[i] = n
		}
	}
	return out
}

// parseCDecimalPrefix 模擬 sscanf 的 "%d":略過的前導空白由呼叫端先處理,
// 這裡只認「可有可無的正負號 + 至少一個數字」,遇到第一個非數字字元就停止
// (尾端多餘字元一律忽略,不算解析失敗 ── 和 C 版 sscanf %d 一致)。
// 找不到任何數字回 (0, false)。
func parseCDecimalPrefix(s string) (int, bool) {
	i := 0
	n := len(s)
	if i < n && (s[i] == '+' || s[i] == '-') {
		i++
	}
	start := i
	for i < n && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == start {
		return 0, false
	}
	v, err := strconv.Atoi(s[:i])
	if err != nil {
		return 0, false
	}
	return v, true
}

func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// parseHexPrefix 模擬 C 版 hex2dec()(kbstd.c:508):對 6 字元做 strtol(base 16),
// 在第一個不合法的十六進位字元處停止(而不是直接判定失敗),盡量吃到多少算多少。
// 呼叫前呼叫端已確認 hexpart[0] 是合法十六進位字元。
func parseHexPrefix(hexpart string) int {
	i := 0
	for i < len(hexpart) && isHexDigit(hexpart[i]) {
		i++
	}
	v, err := strconv.ParseInt(hexpart[:i], 16, 64)
	if err != nil {
		return 0
	}
	return int(v)
}

// armyLineRe 對齊 GNU_parse_troop()(free-data.c:173-197)五種 sscanf 樣式的聯集:
// "%d x %s" / "%dx%s" / "%d x%s" / "%dx %s" / "%d %s" ── 共同點是
// 「數字、可有可無的空白、可有可無的 'x'、可有可無的空白、非空白 token」。
var armyLineRe = regexp.MustCompile(`^\s*(\d+)\s*x?\s*(\S+)`)

// ParseArmyEntry 解析 villains.ini 裡 "armyN = 50 x 農夫" 這種欄位值,
// 回傳兵力數與兵種名稱(如 troops.ini 的 name 值,尚未查表轉成 troop index ──
// 查表留給呼叫端,因為需要完整兵種名稱清單才能比對,不屬本檔職責)。
func ParseArmyEntry(value string) (count int, troopName string, ok bool) {
	m := armyLineRe.FindStringSubmatch(value)
	if m == nil {
		return 0, "", false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, "", false
	}
	return n, m[2], true
}
