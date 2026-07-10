# 移植狀態:C 版 openkb → Go/Ebiten

目標:把 C 版 openkb(free 模組)**逐畫面像素級忠實移植**到 Go/Ebiten,唯一差異是
Android 觸控 + CJK 雙層更銳利。C 版為行為真值 oracle,以 C 源碼為規格直接移植,
用 Linux 桌面 debug flag 秒截圖驗證(桌面=手機同份 screen 碼),模擬器做 Android 驗收。

本檔是移植的**單一真相清單**(gap analysis)。C 功能函式見 openkb-code/src/game.c。

## ✅ 已完成

| 畫面 / 系統 | C 對應 | Go |
|---|---|---|
| 標題 | display_title | title.go |
| 選角(職業+命名) | create_game | charselect.go |
| 世界地圖(移動/tile/起點/踩點進場) | 主迴圈 | worldmap.go |
| 戰鬥(棋盤/佈陣/回合/AI/傷害 parity/勝負) | combat_loop | combatscreen.go + combat 套件 |
| 城鎮 | visit_town | town.go |
| 招募(棲地) | visit_dwelling | recruit.go |
| 存讀檔 | KB_saveDAT/loadDAT | save 套件 |
| 領導力/晉升/寶箱/招兵/週結算(部分) | play.c | gamestate |
| 法術分配各鎮 | salt_spells | gamestate/townspell.go |
| 觸控層(每畫面 Keymap 自動生成) | — | input/touch.go |
| CJK 雙層合成 | env-sdl cjk overlay | render/cjktext.go |
| **軍隊檢視** | view_army | viewarmy.go(+ gamestate/morale.go:troop_morale/morale_chart/morale_names) |
| **角色檢視** | view_character | viewcharacter.go(班底立繪 + 數值框 + 底部道具帶;player_captured/num_artifacts/castles/score/followers_killed 為世界狀態佔位 0,artifact_found 全空、continent_found 僅 home 洲) |
| **世界生成(salt_continent)** | salt_continent/populate_dwelling/repopulate_foe/roll_creature(play.c:28-330) | gamestate/worldgen.go(NewGame 逐洲跑,填 gs.WorldMap + Dwelling*/Foe*/Orb/Navmap/Telecave 座標;worldmap.go/recruit.go 接線讀真實世界狀態,取代舊 placeholderFoe/demoRecruitTroopID 佔位) |

## ⬜ 剩餘畫面(以 C 源為規格逐一移植)

### A. 檢視畫面(主要是顯示既有 gamestate,自足、優先)
- [ ] **view_contract** 懸賞契約(game.c:1641)— 目標 villain 頭像 + 資訊(需 contract 世界狀態)
- [ ] **view_puzzle** 拼圖(game.c:1392)— 5×5 拼圖(需 artifact/villain 世界狀態)
- [ ] **view_minimap** 小地圖/導航(game.c:1533)— 該洲縮圖(需 orb/navmap 狀態)

### B. 地點畫面(多需世界狀態)
- [ ] **visit_castle / visit_own_castle / visit_home_castle** 城堡(2307/2381/2567)— 需 castle_owner/troops/numbers
- [ ] **lay_siege** 攻城(2520)
- [ ] **audience_with_king** 謁見國王(2130)
- [ ] **visit_alcove** 魔法密室(2896)
- [ ] **visit_telecave / select_gate** 傳送洞(2973/4146)
- [ ] **read_signpost** 路標(3095)

### C. 動作 / 系統
- [ ] **navigate_continent** 航行換洲(1972)— 需 boat
- [ ] **dismiss_army / dismiss_troop** 解散部隊(2027/play.c)
- [ ] **choose_spell** 戰鬥施法(4275)— combat 已有結構,缺施法 caller(見 go issue #1)
- [ ] **end_week** 週結算完整化(目前部分)
- [ ] 寶箱 search_area(踩寶箱)

### D. 開場 / 雜項
- [ ] **display_logo** NWC logo(924)
- [ ] **show_credits** 製作名單(706)
- [ ] **display_cartoon** 開場動畫 / 地球(4513;free 版可能空白/低優先)

## ⬜ 世界狀態(foundational — 解鎖上面多個畫面的真實動作邏輯)

recruit 的兵種/庫存、worldmap 的 foe/dwelling 已改讀 salt_continent 生成的真實世界狀態
(見下方「世界生成」段);town 的契約/租船/情報仍是「佔位 stub」,要讓它們變真實,
還需要把 C spawn_game 剩下的 contract/boat/castle/villain 世界生成移植進 gamestate.NewGame:

- [ ] **contract / villain** 系統(contract 循環、villain 位置、view_contract/sidebar 頭像)
- [ ] **boat**(租船 + 航行 + 上下船)
- [ ] **castle** 世界狀態(castle_owner/troops/numbers、repopulate_castle、salt_villains)
- [x] **dwelling** 真實兵種 + 庫存(populate_dwelling/dwelling_population;recruit.go 已讀真實世界狀態)
- [x] **foe** 真實部隊(foe_troops/foe_numbers;worldmap.go 已讀真實世界狀態,取代 placeholderFoe)
- [ ] **artifact / scepter**(拼圖、尋寶、破關條件;座標已記錄於 map tile,尚未接 UI/撿拾邏輯)
- [ ] **sidebar 動態內容**(contract 頭像、拼圖 piece 疊圖 — 需 B/世界狀態)

### 世界生成(salt_continent)——已完成(2026-07-10)

移植自 `salt_continent`/`populate_dwelling`/`enforce_dwelling`/`repopulate_foe`/`roll_creature`
(play.c:28-330),見 `internal/gamestate/worldgen.go`。逐句對照 C,細節/怪癖照抄:

- **gs.WorldMap**:NewGame 時深拷貝 `assets.World`(`copyWorldMap`),salt 就地 mutate 這份拷貝,
  不動唯讀的 `kbdata.Assets.World`。worldmap.go/recruit.go 已改讀 `gs.WorldMap`/`gs.Dwelling*`/`gs.Foe*`。
- **世界狀態欄位**:DwellingCoords/Troop/Population、FoeCoords/Troops/Numbers、OrbCoords/
  NavmapCoords/TelecaveCoords(gamestate.go),全部 exported,隨 GameState 一起被
  encoding/json 存讀檔完整持久化——**未擴充存檔格式**(當初計畫的「存 seed 重跑 salt」方案
  最終改採「直接持久化整份世界狀態」,更簡單且不必擔心 RNG parity 隨版本改動而讓存檔重播出不同世界)。
- **NewGame 呼叫**:`salt_spells` 之後、`for cont := range 4 { gs.saltContinent(a, rng, cont, 2,1,1,2,10,5) }`
  ——沿用同一個 kbrng 實例延續消耗(RNG parity 誠實註記見 newgame.go/worldgen.go 內文,尚未逐一
  對齊 spawn_game 完整 rand 呼叫序列)。
- **驗收**:單元測試(gamestate/worldgen_test.go)+ 桌面截圖(Xvfb + xdotool 實際走位,
  踩進真實生成的地下城棲地,recruit 畫面顯示「骷髏兵/150隻/造價40」,與程式化 dump 逐字對上)。

**移植過程中發現並修正的既有勘誤(non-obvious,值得留存)**:
1. `kbdata.Troop` 的 `Growth`/`MoraleTop` 兩欄位命名與 C 版 `KBtroop.max_population`/`growth`
   語意對調(數值搬得對,名字寫反),已更正為 `MaxPopulation`/`Growth`(見 assets.go 欄位勘誤註記)。
2. `kbdata.Dwelling` 列舉數值原本以 `DwellCastle=0` 起算,對不上 C 版 `DWELLING_PLAINS=0..
   DWELLING_CASTLE=4` 的實際數值(populate_dwelling 直接拿這個數值做 tile 運算,不是純標籤),
   已更正 iota 順序對齊 C。
3. **確認一個 C 原始資料本身的既有怪癖(非本次移植引入)**:洲2的 `dwelling_ranges=[2,14]`
   涵蓋 troop id=2(義勇軍,Home=Castle),隨機 roll 到它時 tile = `TileDwelling1+4=0x90`
   撞進 `TileSignpost` 的數值——桌面截圖驗收時實測命中此案例(continent=2 id=4,seed=1),
   已在 worldgen.go/worldgen_test.go 逐句記錄成因,不視為 bug、不「修正」遊戲邏輯本身。

## 建議優先序

1. **檢視畫面 A(view_army、view_character 已完成)** — 自足、只讀既有 gamestate、玩家常用,先做累積可見進度。剩 view_contract/view_puzzle/view_minimap 皆需世界狀態,待 2 完成後再回頭。
2. **世界狀態 D(dwelling→foe→contract/boat)** — 逐一把 stub 換真實,解鎖 recruit/combat/town 動作與 view_contract/sidebar。
3. **城堡系列 B** — 待 castle 世界狀態後做。
4. **開場/雜項 D** — 收尾。

> 每項:以 C 源為規格 → 桌面 debug flag 截圖對齊 → gamestate 邏輯旗艦自己做、
> 解碼/render/佈局派便宜 subagent → docker build/test 綠 + 截圖驗收才 commit。

## 已知待打磨(跨畫面,非阻塞)

- [ ] **數字欄右對齊**:C 格式字串用空格 padding 假設每個 CJK 字 = 2 個 ASCII 格,
      但本移植 CJK 渲染是 1 格/字(雙層合成的專案慣例,換來 CJK 更銳利)。故有 `%Nd`
      右對齊數字的畫面(view_character/view_army/recruit)不同長度標籤的數字 x 位置
      略參差、偶有貼近/溢出框邊。**cosmetic、跨畫面一致、非 regression**。正解=寫個
      CJK 格寬感知的欄位對齊 helper(從框右緣算數字 x,不靠空格 padding),之後一次
      套用到所有這類畫面。
