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

目前多畫面的動作是「佔位 stub」(town 的契約/租船/情報、recruit 的兵種/庫存、
worldmap 的 foe/dwelling)。要讓它們變真實,需把 C spawn_game/salt_continent
的世界生成移植進 gamestate.NewGame:

- [ ] **contract / villain** 系統(contract 循環、villain 位置、view_contract/sidebar 頭像)
- [ ] **boat**(租船 + 航行 + 上下船)
- [ ] **castle** 世界狀態(castle_owner/troops/numbers、repopulate_castle)
- [ ] **dwelling** 真實兵種 + 庫存(populate_dwelling/dwelling_population;取代 recruit 的佔位)
- [ ] **foe** 真實部隊(foe_troops/foe_numbers;取代 worldmap 的 placeholderFoe)
- [ ] **artifact / scepter**(拼圖、尋寶、破關條件)
- [ ] **sidebar 動態內容**(contract 頭像、拼圖 piece 疊圖 — 需 B/世界狀態)

### ★世界生成架構計畫(調查後,2026-07-10)

**關鍵發現**:Go 的 raw land.org 是 **salt 前**的地圖(land_test.go 證實:178 個 chest 0x8B
+ 92 個 foe 0x91,**無 dwelling tile**)。C 的 dwelling/artifact/orb/telecave 都是
`salt_continent`(play.c:183)在建遊戲時**把 chest tile 轉成**的。**Go 尚未移植 salt_continent**
→ 所以正常遊玩中 dwelling 根本不出現(recruit 只能靠 -startrecruit debug flag 到)、
foe 也用 placeholderFoe。

**移植架構**(下一步邏輯移植照此做):
1. **gamestate 持有 per-game 可變地圖** `gs.Map`(NewGame 時 copy assets.World → gs.Map),
   因為 salt 會 mutate 地圖(chest→dwelling)。worldmap.go 改讀 `gs.Map` 而非 assets.World(共享唯讀)。
2. **NewGame 跑 salt_continent**(移植 play.c:183-330 的 chest 掃描 + 放置):
   salt_continent(game, cont, 2 artifact,1 navmap,1 orb,2 telecave,10 dwelling,5 friendly)。
   放置時用 kbdata 的世界生成表(dwelling_ranges/continent_dwellings/dwelling_to_troop 等,
   go-worldgen-tables 正在移植)+ populate_dwelling(play.c:50)+ repopulate_foe/roll_creature(play.c:28/89)。
3. **gamestate 加地圖級陣列**:DwellingCoords/Troop/Population、FoeCoords/Troops/Numbers
   (per continent),NewGame 填好。
4. **存讀檔**:salt 是 deterministic(給定 seed+map)→ 存檔存 world seed,load 時**用該 seed 重跑 salt**
   還原世界,避免擴充存檔格式(同 roguelike 存 seed 的做法)。
5. **接線**:recruit 讀 gs.DwellingTroop/Population(取代 demo 佔位);worldmap combat 讀
   gs.FoeTroops/Numbers(取代 placeholderFoe)。
6. **RNG parity**:同 salt_spells——NewGame 未逐一重現 spawn_game 完整 rand 序列,故非逐 seed
   對齊 C,但演算法忠實(放置規則/roll 範圍/型別一致)。誠實標註。

> 這塊 correctness 敏感(RNG + 地圖 mutation + 存檔),分兩步:①資料表(go-worldgen-tables,進行中)
> ②生成邏輯 + 每局地圖 + 接線(下一個專項,旗艦嚴審)。

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
