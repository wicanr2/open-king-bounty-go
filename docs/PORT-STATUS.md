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
| **世界生成(城堡/惡棍/契約起手值)** | salt_villains/repopulate_castle/num_castles(play.c:77-182)+ spawn_game 城堡/契約段(374-478) | gamestate/castlegen.go + castle.go(NewGame 在 saltContinent 四洲之後跑;CastleOwner/CastleTroops/CastleNumbers/Contract 系列欄位,見下方獨立小節;**visit_castle 仍待後續**) |
| **懸賞契約檢視** | view_contract(game.c:1641-1735) | viewcontract.go(無契約:sidebar.png 幀8 空框 +「你目前沒有懸賞契約！」;有契約:GR_VILLAIN 臉部 4 幀動畫 + STRL_VDESCS 描述文字,%s %s 代入洲名/城堡名;見下方獨立小節的資料溯源紀錄) |

## ⬜ 剩餘畫面(以 C 源為規格逐一移植)

### A. 檢視畫面(主要是顯示既有 gamestate,自足、優先)
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

- [x] **villain 位置 + contract 起手值**(salt_villains 已把惡棍塞進城堡;contract/last_contract/max_contract/contract_cycle 起手值已設,見下方獨立小節)——**view_contract 畫面已完成(見上表);sidebar 頭像疊圖/接受契約動作仍未做**
- [ ] **boat**(租船 + 航行 + 上下船)
- [x] **castle** 世界狀態(castle_owner/troops/numbers、repopulate_castle、salt_villains,見下方獨立小節)——**visit_castle 等城堡畫面仍未做**
- [x] **dwelling** 真實兵種 + 庫存(populate_dwelling/dwelling_population;recruit.go 已讀真實世界狀態)
- [x] **foe** 真實部隊(foe_troops/foe_numbers;worldmap.go 已讀真實世界狀態,取代 placeholderFoe)
- [ ] **artifact / scepter**(拼圖、尋寶、破關條件;座標已記錄於 map tile,尚未接 UI/撿拾邏輯)
- [ ] **sidebar 動態內容**(draw_sidebar 的 contract 頭像疊圖 —— chrome.go drawSidebar 幀 8 目前恆顯示空框註解已過期:Contract 系統與 GR_VILLAIN 素材現已齊備,可疊 gs.Contract 對應的臉部幀,是低成本的後續補完,非世界狀態缺口;拼圖 piece 疊圖仍需 artifact/villain 世界狀態)

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

### 世界生成(城堡/惡棍/契約起手值)——已完成(2026-07-10)

移植自 `num_castles`/`salt_villains`/`repopulate_castle`(play.c:128-182)+ `spawn_game`
的城堡初始化/惡棍呼叫序列/契約起手值段(play.c:374-478),見
`internal/gamestate/castle.go`(靜態城堡表)+ `internal/gamestate/castlegen.go`
(salt_villains/repopulate_castle 邏輯)。逐句對照 C,細節/怪癖照抄:

- **城堡座標來源(關鍵溯源,rulebook 62)**:`castle_coords[MAX_CASTLES][3]`(bounty.c:647)
  **會**被 free 模組透過 `refill_rules()`(game.c:304-307,`DAT_CASTLEX/Y/C`)覆寫,
  且 free 的 `data/free/castles.ini` 座標與 bounty.c 硬編值**確實不同**(castle0:
  bounty.c `{0,30,27}` vs castles.ini `{continent=0,x=44,y=12}`,因為 free 用自製
  land.tmx 地圖、非 DOS 原版佈局)——與先前 town_coords/special_coords 已知的同類坑
  相同模式。故 `gamestate/castle.go` 的 `LoadCastles` 改讀 `castles.ini`(對齊既有
  `town.go` 讀 `towns.ini` 的做法),**不**照抄 bounty.c 的 `castle_coords` 字面。
  `castle_difficulty[MAX_CASTLES]`(bounty.c:612)則相反——free 模組**沒有**
  `DAT_CASTLEDIFF` 覆寫鍵,`castles.ini` 裡的 `difficulty` 欄位是 land.tmx 匯出工具的
  殘留欄位、C 引擎從未讀取,故 `Castle.Difficulty` 仍查 `kbdata.CastleDifficultyTable()`
  (bounty.c 硬編,不受 ini 影響)。
- **惡棍守軍資料**(`villain_army_troops`/`villain_army_numbers`,bounty.c:247/266)
  同樣會被 `DAT_VTROOP`/`WDAT_VNUMBER` 覆寫,但**逐筆核對** free 的
  `villains.ini` armyN 欄("N x 兵種名"字串,經 troops.ini 名稱表可還原成 troop id)
  與 bounty.c 字面值**完全一致**(只是把同一批數字換成中文兵種名重新表達),
  故 `kbdata/tables_villains.go` 直接照抄 bounty.c 字面,不需在執行期解析 ini。
  `villains_per_continent`(bounty.c:220)確認**沒有**對應 `DAT_*` 覆寫鍵。
- **spawn_game 呼叫順序**(play.c:461-478,`saltContinent` 四洲之後):
  全部 26 城堡 `CastleOwner` 先設 `castleOwnerMonsters`(0x7F)→ 依序對 4 洲呼叫
  `saltVillains`(累加 base_id,對齊 C `salt_villains(0,0)/(1,i)/(2,i)/(3,i)`)→
  對仍是 0x7F 的城堡呼叫 `repopulateCastle`。契約起手值(`Contract=0xFF`/
  `LastContract=0x04`/`MaxContract=0x05`/`ContractCycle={0,1,2,3,4}`,play.c:418-425)
  在 NewGame 內對齊 C 原始位置設定(不耗用 rand,位置對結果無影響)。
- **世界狀態欄位**:CastleOwner/CastleTroops/CastleNumbers、Contract/LastContract/
  MaxContract/ContractCycle(gamestate.go),全部 exported,隨 GameState 一起被
  encoding/json 存讀檔完整持久化,未擴充存檔格式。
- **RNG parity 誠實註記**:與 salt_continent 同一個 kbrng 實例延續消耗,但 C 版
  spawn_game 在 salt_continent 之前已先呼叫 scepter 相關 rand(藏權杖鑰匙/選洲/選格),
  Go 版尚未移植這段,故兩邊 rand 呼叫序列從最開頭就已分岔,「同 seed 逐值對齊」
  仍不保證;salt_villains/repopulate_castle 演算法本身逐句忠實對齊 C。
- **驗收**:單元測試(gamestate/castle_test.go、castlegen_test.go、
  castlegen_newgame_test.go)涵蓋:castles.ini 真值核對(含缺檔 best-effort)、
  num_castles 洲別分佈(9/7/6/4)、saltVillains 直接演算法(4 個 seed,逐洲惡棍數 +
  守軍資料對表)、城堡不足安全閥(不 panic/不無窮迴圈)、repopulateCastle 保底值、
  NewGame 整合(每洲惡棍數 6/4/4/3、17 座惡棍城堡守軍逐值對表、9 座怪物城堡
  repopulate 非空、契約起手值、存讀檔往返)。docker `golang:1.24-bookworm`
  build/vet/xvfb-run test 全綠(含既有 render/screen/combat 套件回歸)。

### 懸賞契約檢視(view_contract)——已完成(2026-07-10)

移植自 `view_contract()`(game.c:1641-1735),見 `internal/screen/viewcontract.go`
(畫面)+ `internal/gamestate/villain.go`(Villain 靜態表 + `FindVillainCastle`)+
`internal/gamestate/continent.go`(`ContinentNames`)+ `internal/kbdata/assets.go`
(`Assets.VillainDescs`,STRL_VDESCS 描述文字載入)+ `internal/screen/villain_assets.go`
(GR_VILLAIN 臉部 sprite)。逐句對照 C,關鍵資料溯源(rulebook 62):

- **實際顯示的文字不是 villains.ini 的 name/reward,是 `<file>.txt`**:C 版
  `KB_Resolve(STRL_VDESCS, villain_id)`(free-data.c:1132)讀
  `DOS_villain_names[id]+".txt"`(如 `czar.txt`),檔案本身已包含「姓名/別號/懸賞/
  最後現身/城堡/特徵/罪行」等全部文字,只留兩個 `%s` 佔位給洲名與城堡名。
  `villains.ini` 的 name/reward 欄位在這個畫面完全不使用(留給未來 visit_castle
  等其他畫面查詢用,故仍建了 `gamestate.LoadVillains`)。
- **洲名不是 bounty.c 硬編值,是 land.ini(重大溯源發現)**:game.c:1699 讀的
  `continent_names[]` 全域陣列,在**每次開新遊戲後**都被 `refill_names()`
  (game.c:221,呼叫點 490/700/7116/7145)整個覆寫成 `STRL_CONTINENTS`
  (= `land.ini` `continent0-3` 的 `name` 欄位)。也就是說 bounty.c 硬編的
  「大陸洲/森林洲/群島洲/撒哈洲」在 free 模組正常遊戲流程中**從未被實際顯示過**,
  是死預設值——實際顯示的是 land.ini 的「弗蘭德利亞/第二洲/第三洲/第四洲」
  (這幾個名字本身像 land.tmx 匯出工具的佔位文字,但那是資料內容本身,不影響
  「執行期讀哪個來源」的判斷)。`kbdata.freeIniNames` 因此新增 `"land"`,
  `gamestate.ContinentNames` 讀 `a.Strings["land"]`。城堡名同理沿用既有決策
  (`castle.go` 已讀 `castles.ini`,不用 bounty.c `castle_names[]` 硬編英文預設值——
  該陣列同樣被 `refill_names()` 的 `castle_list`/`STRL_CASTLES` 覆寫,只是這條
  溯源在城堡世界生成那次移植就已查清,此次只是再次確認結論一致)。
- **前導空白是刻意版面設計,不是雜訊**:C 版文字與臉部圖(48×34,4 幀)共用同一個
  blit 原點(`hdst` = `border.x+fs.w, border.y+fs.h`),CJK 又是「先畫美術層、
  後疊字层」(雙層合成,見 render/cjktext.go),理論上文字前幾格會蓋在臉部圖上。
  實測(`czar.txt`)每行前面固定 7 個全形空格(`城堡:` 那行更多,刻意再縮排對齊
  `最後現身:` 下方),換算成本畫面的 CJKCell(8px)寬度,恰好把可讀文字推到臉部圖
  (48px=6 cell)右緣之後,兩者不重疊——桌面截圖驗證视覺上確實乾淨對齊,證實這是
  原始資料檔案的刻意留白,移植時原樣保留(不 trim),不是要修正的雜訊。
- **C 版不呼叫 KB_TopBox/KB_BottomFrame**:逐讀 game.c:1641-1735 全函式確認,
  與 view_army/view_character 不同,故本移植也不畫頂列提示文字。
- **驗收**:docker `golang:1.24-bookworm`(+ X11 dev headers)build/vet/test 全綠
  (含既有套件回歸)。桌面截圖(Xvfb + import,`-startviewcontract` /
  `-contractvillain 13`)驗過兩態:①`Contract=0xFF`——藍底框 + sidebar.png 幀 8
  空框 +「你目前沒有懸賞契約！」置中偏移正確;②`Contract=13`(沙皇鮑里斯三世/
  czar)——臉部立繪(皇冠鬍鬚特徵清楚)+ 完整描述文字 13 行皆在框內、無溢出,
  `find_villain_castle` 找到 castle8「艾昂」、洲別「第三洲」代入正確
  (與 castles.ini castle8/land.ini continent2 逐一核對一致)。
- **已知後續(非本次範圍,已在世界狀態清單標註)**:`chrome.go` 的
  `drawSidebar` 幀 8(合約框)目前仍恆顯示空框——舊註解「暫無 contract 系統」
  已過期(Contract 欄位與 GR_VILLAIN 素材現已齊備),補上疊圖是後續低成本工作,
  不在本次任務(view_contract 檢視畫面本體)範圍內,已在上方「世界狀態」清單註記。

## 建議優先序

1. **檢視畫面 A(view_army、view_character、view_contract 已完成)** — 自足、只讀既有 gamestate、玩家常用,先做累積可見進度。剩 view_puzzle/view_minimap 需 artifact/orb 世界狀態,待 2 完成後再回頭。
2. **世界狀態 D(dwelling→foe→boat)** — 逐一把 stub 換真實,解鎖 recruit/combat/town 動作與 sidebar 動態內容。
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
