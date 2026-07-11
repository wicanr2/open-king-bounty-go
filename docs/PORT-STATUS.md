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
- [x] **view_puzzle** 拼圖(game.c:1392)——已完成(2026-07-10):5×5 拼圖,神器/惡棍掀開露出權杖周邊地圖。世界地圖 'p'。
- [x] **view_minimap** 小地圖(game.c:1533)——已完成(2026-07-10):本洲 64×64 縮圖。世界地圖 'm'。

### B. 地點畫面(多需世界狀態)
- [x] **visit_castle / visit_own_castle / visit_home_castle** 城堡(2307/2381/2567)——已完成(2026-07-10),見下方獨立小節
- [x] **lay_siege** 攻城(2520)——已完成(進圍攻戰;⚠ 勝利奪城 + 城牆障礙未做,見小節)
- [x] **audience_with_king** 謁見國王(2130)——已完成(兩頁對話 + 達門檻晉升)
- [x] **visit_alcove** 魔法密室(2896)——已完成(2026-07-10):大法師 5000 金幣傳授魔法(KnowsMagic)。
- [x] **visit_telecave** 傳送洞(2973)——已完成(2026-07-10):踩洞口瞬移到成對另一端
      (與 TileDwelling3 同值 0x8E,靠座標區分)。**select_gate**(4146,法術傳送到已訪城堡/鎮)
      仍待 choose_spell 一起做。
- [x] **read_signpost** 路標(3095)——已完成(2026-07-10):踩路標顯示 signs.txt 對應文字。

### C. 動作 / 系統
- [x] **run_combat 戰後結算**(game.c:3598)——已完成(2026-07-10):存活寫回、foe 清 tile、
      圍攻奪城 + 履約、戰敗 temp_death,見下方「combat 戰後 hook」小節。⚠ 圍攻城牆障礙佈局未做。
- [x] **navigate_continent** 航行換洲(1972)——已完成(2026-07-10,見 boat 小節)
- [x] **dismiss_army / dismiss_troop** 解散部隊(2027/play.c)——已完成(2026-07-10):
      世界地圖 'd',列部隊選解散;最後一支需確認 → temp_death。
- [x] **choose_spell** 戰鬥+冒險施法(4275)——已完成(2026-07-10):全 14 法術,見下方剩餘清單。
- [ ] **end_week** 週結算完整化(目前部分)
- [x] 寶箱 take_chest(踩寶箱)——已完成(2026-07-10):海圖解鎖下一洲(補上 boat
      的探索觸發)/寶珠 orb_found /寶藏(金幣·領導力二選一·收入·法力·法術上限·獲法術·空箱)。

### 玩家座標 / 存讀檔——已完成(2026-07-10)
玩家 Continent/X/Y 已從 WorldMapScreen 移入 GameState,修正存讀檔遺失位置 +
戰敗 temp_death 傳送回家。詳見 commit「worldmap: 玩家座標移入 GameState」。
仍未建模:mount / siege_weapons(temp_death 略過)。

### combat 戰後 hook——已完成(2026-07-10)
`internal/screen/combatscreen.go` 的 combatContext + applyOutcome 移植 C run_combat
戰後段(game.c:3598-3652):NewCombatScreenFoe/Siege 標記來源,勝負確定時一次結算——
AcceptUnitsPlayer/AcceptSquads 寫回兩側存活單位、foe 戰勝清 tile、圍攻奪城 +
FulfillContract(讓 audience 晉升門檻能達成)、戰敗 TempDeath。新增
gamestate.FulfillContract/TempDeath、combat.AcceptUnitsPlayer/AcceptSquads、
kbdata.WorldMap.Set。⚠ 圍攻城牆障礙佈局(combat mode=1 的 castle_umap)未做。

### D. 開場 / 雜項
- [x] **display_logo** NWC logo(924)——已完成(2026-07-10):LogoScreen(nwcp.png)→ 標題。
- [x] **show_credits** 製作名單(706)——已完成(2026-07-10):標題按 'c' → CreditsScreen(credits.txt)。
- [ ] **display_cartoon** 開場動畫 / 地球(4513)——free 模組無專屬 cartoon 動畫幀(僅 nwcp.png 商標,已用於 logo);低優先,可視為 N/A。

## ⬜ 世界狀態(foundational — 解鎖上面多個畫面的真實動作邏輯)

recruit 的兵種/庫存、worldmap 的 foe/dwelling 已改讀 salt_continent 生成的真實世界狀態
(見下方「世界生成」段);town 的契約/租船/情報仍是「佔位 stub」,要讓它們變真實,
還需要把 C spawn_game 剩下的 contract/boat/castle/villain 世界生成移植進 gamestate.NewGame:

- [x] **villain 位置 + contract 起手值**(salt_villains 已把惡棍塞進城堡;contract/last_contract/max_contract/contract_cycle 起手值已設,見下方獨立小節)——**view_contract 畫面已完成(見上表);sidebar 頭像疊圖/接受契約動作仍未做**
- [x] **boat**(租船 + 航行 + 上下船 + 換洲)——已完成(2026-07-10):Boat/Mount/
      ContinentFound 世界狀態、城鎮 B) 租船/取消、世界地圖登船/航行/靠岸、navigate_continent
      換洲(sail_to + 消耗一週)。詳見下方「boat 系統」小節。⚠ 神器降租金(BoatCost 恆貴)、
      連通性檢查(哪些洲可航抵)未做。
- [x] **castle** 世界狀態(castle_owner/troops/numbers、repopulate_castle、salt_villains,見下方獨立小節)——**visit_castle 三畫面 + 圍攻/謁見已完成(2026-07-10);garrison/ungarrison/castle_visited/villain_caught 已建模**
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

### 城堡系列(visit_castle)——已完成(2026-07-10)

移植自 `visit_castle`/`visit_home_castle`/`visit_own_castle`/`lay_siege`/`recruit_soldiers`/
`audience_with_king`(game.c:2130-2660)+ `garrison_troop`/`ungarrison_troop`/
`save_castle_owner_knowledge`/`player_captured`(play.c),逐句對照 C:

- **gamestate 動作層** `internal/gamestate/castle_actions.go` + `rank.go`:
  - `CastleAt(cont,x,y)` 分派器(land.ini special0 = 家鄉;castles.ini 座標 → your/enemy)。
  - `GarrisonTroop`/`UngarrisonTroop`/`dismissTroop`(Army[5]Squad ↔ C player_troops/numbers)。
  - `SaveCastleOwnerKnowledge`(OR 進 `KBCastleKnown` 位)、`HomeTroops`(DWELLING_CASTLE)。
  - `VillainCaught[]`/`PlayerCaptured`/`VillainsNeeded`(晉升門檻,audience 用)。
  - 常數 `KBCastlePlayer/Monsters/Known/Villain`、`HomeCastleName`「馬克馬斯國王」
    (index MAX_CASTLES,不受 free refill_names 覆寫)。
- **五個畫面** `internal/screen/castle{home,own,siege}.go` + `recruitsoldiers.go` + `audience.go`:
  - 家鄉:王座廳選單 A)招募士兵→RecruitSoldiersScreen(5 城堡兵種、無庫存上限)
    B)謁見國王→AudienceScreen(兩頁對話,達門檻建構時 Promote)。
  - 自家:駐防/撤離兩模式(Confirm 切換,對齊 C 空白鍵),A-E 對五格操作。
  - 敵方:圍攻詢問 y/n,惡棍佔領時記下已知統治者,y → 以城堡守軍佈陣 Push 戰鬥。
- **worldmap.go** 城堡 tile 分派;進 your/enemy 記 `castle_visited`(供 select_gate)。
- **chrome.go** 抽 `drawLocationBg` 共用(背景 + 迎賓兵種幾何,家鄉/招兵重用)。
- 桌面 ffmpeg 截圖驗證四畫面版面忠實;6 個 screen 測試綠。
- **⚠ 未完成(誠實標註)**:① 圍攻勝利後奪城(castle_owner=玩家、守軍換玩家駐防、
  villain_caught)屬 run_combat mode=1 戰後結算,CombatScreen 目前無戰後世界狀態回寫
  (與一般 foe 戰鬥同一缺口);② 圍攻城牆障礙佈局(combat mode=1)未建模;
  ③ lay_siege 畫面 C 版是疊在世界地圖上(無 draw_location),本移植先清成黃底。

### boat 系統——已完成(2026-07-10)
三切片,對齊 C game->boat/boat_x/boat_y/mount/continent_found 世界狀態 + 移動 boat 段
(game.c:6837-6957)+ navigate_continent(1972)/sail_to(play.c:821):
- **世界狀態**:GameState.Boat/BoatX/BoatY/Mount/ContinentFound;gamestate/boat.go
  (KBMount* 常數、BoatCost、HasBoat、RentOrCancelBoat)。Town 加 BoatX/BoatY。
- **城鎮 B) 租船/取消**:扣費停船 / 乘船中需先離船 / 取消;選單行依 HasBoat 切文字。
- **世界地圖登船/航行/靠岸**:走進船停靠格登船(Mount=SAIL)、水域僅乘船可過、船隨玩家、
  靠岸下船(船留上一格水域);主角 sprite 幀用 Mount(SAIL=0 船 / RIDE=8 馬)。
- **navigate_continent**:乘船時按 'n' → NavigateContinentScreen(列 ContinentFound
  的洲)→ SailTo(切洲 + 移入口 ContinentEntry)+ 消耗一週。
- ⚠ 未做:神器 POWER_CHEAPER_BOAT_RENTAL 降租金(BoatCost 恆 CostBoatExpensive)。
  **探索觸發已補**:寶箱的導航海圖(take_chest ChestNavmap)會 ContinentFound[c+1]=1,
  這是正常遊戲取得可航洲的機制(見「寶箱 take_chest」)。

## 建議優先序

1. **檢視畫面 A(view_army、view_character、view_contract 已完成)** — 自足、只讀既有 gamestate、玩家常用,先做累積可見進度。剩 view_puzzle/view_minimap 需 artifact/orb 世界狀態,待 2 完成後再回頭。
2. **世界狀態 D(dwelling→foe→boat)** — ✅ dwelling/foe/boat 皆完成。
3. **城堡系列 B** — ✅ 已完成(2026-07-10),含戰後奪城。
4. **combat 戰後結算 hook** — ✅ 已完成(2026-07-10)。
5. **開場/雜項 D** — 收尾。

### 2026-07-10 大進度後的剩餘清單(核心遊戲循環已閉合)
title→選角→世界地圖→(城鎮契約·租船·情報·學法術 / 城堡家鄉·自家·圍攻 / 棲地招兵 /
foe 戰鬥含奪城履約 / 寶箱含海圖解鎖洲 / 傳送洞 / 路標 / 解散部隊 / 換洲航行 / 存讀檔含位置)
全通。**剩餘**:
- [x] **choose_spell** — ✅ 已完成(2026-07-10):全 14 法術。冒險 7(造橋·停時·尋敵·城堡傳送·
  鄉鎮傳送·即時軍隊·提升控制,世界地圖 'z')+ 戰鬥 7(分身·瞬移·火球·閃電·冰凍·復活·驅散
  不死,CombatScreen 'z' + 棋盤目標游標)。效果引擎 combat/spells.go + gamestate/spellcast.go
  + adventure_spells.go + gate.go。⚠ 神器加成(Powers)未接。
- [x] **artifact 拾取 + scepter 破關** — ✅ 已完成(2026-07-10):神器拾取(TakeArtifact,
  即時能力 + HasPower + BoatCostWith)、權杖埋藏(PlaceScepter)+ 搜索破關(SearchScreen 'g' +
  WinScreen)。**遊戲可從新遊戲玩到通關。** ⚠ 剩 view_puzzle 拼圖、戰鬥中神器 Powers(增傷/減傷)、
  結局動畫/圖(版權素材)。
- [x] **view_puzzle / view_minimap / visit_alcove** — ✅ 已完成(2026-07-10)。
- [x] **end_week 完整化** — ✅ 已完成(2026-07-10):停時歸零 + 巢穴/foe/城堡每週成長 + 空城堡補守軍。⚠ 剩精確 week_id(用 Week 計數近似)、opt_* 選項旗標。
- [x] **開場 display_logo / show_credits** — ✅ 已完成(2026-07-10)。display_cartoon = free 無素材(N/A)。
- [x] **戰鬥中神器 Powers** — ✅ 已完成(2026-07-10):PrepareUnitsPlayer 從 ArtifactFound 填 combat.Powers。

### 移植大致完成(2026-07-10)——剩餘只餘打磨
核心 + 進階系統(城鎮/城堡/棲地/戰鬥含施法奪城/寶箱/boat/傳送/路標/解散/換洲/
神器/破關/魔法密室/小地圖/拼圖/週結算/開場)全數移植,遊戲可從新遊戲玩到通關。
**剩餘皆屬打磨,非阻塞**:
- ~~view_puzzle 未掀開格的惡棍臉/神器圖示(灰底)~~ ✅ 已完成(2026-07-10)。
- ~~minimap orb 迷霧(顯示整洲)~~ ✅ 已完成(2026-07-10):Fog-of-war + orb 揭示(見下方主題/打磨區)。
- ~~sidebar 契約頭像疊圖~~ ✅ 經查已實作於 chrome.go drawSidebar。
- ~~view_character 數值右對齊~~ ✅ 已完成(2026-07-10,drawLV 右對齊框右緣)。
- ~~view_army/recruit 亦可套 drawLV~~ ❌ **經查證不套(2026-07-11)**:這兩畫面數字是
  C 的 **inline label:value**(view_army「生命:40 / 傷害:20-40 / 造價:100」;recruit
  「可招募 999 名農夫 / 你最多可招募 100 名」),桌面截圖確認已對齊、不參差;drawLV 是
  「數值右對齊框右緣」的欄式排版,只適用 view_character 那種 stat:value 欄——套到這兩畫面
  反而**偏離 C**。故此項不做(非缺陷)。
- ~~worldmap 拼圖 'p' 觸控出界~~ ✅ 已於三層觸控重構(sidebar 面板隱形熱區
  {256,123,48,34} 在界內)+ commit 82c462b 修復。
- ~~主題切換 toast 提示~~ ✅ 已完成(2026-07-11):`internal/app/toast.go`——切美術主題
  (F8/☰)/切音樂(F9/音樂鈕)顯示置中提示(「主題:Amiga」「音樂:開/關」),深藍底金框,
  ~1.8s 自動消失;-toast 旗標截圖驗證。
- 剩:sidebar 拼圖 piece 疊圖(需 artifact/villain 世界狀態)、主題設定持久化、end_week
  完整化(皆次要)。
- 精確 week_id / opt_* gameplay 旗標、read_signpost 以外的細節音效/動畫。
- RNG 逐 seed parity(演算法忠實,但未從第一個 rand() 起對齊 C,見 NewGame 註記)。
- Android 模擬器整合驗收新系統(桌面 ffmpeg 已驗關鍵畫面)。

> 每項:以 C 源為規格 → 桌面 debug flag 截圖對齊 → gamestate 邏輯旗艦自己做、
> 解碼/render/佈局派便宜 subagent → docker build/test 綠 + 截圖驗收才 commit。

## 已知待打磨(跨畫面,非阻塞)

- [ ] **數字欄右對齊**:C 格式字串用空格 padding 假設每個 CJK 字 = 2 個 ASCII 格,
      但本移植 CJK 渲染是 1 格/字(雙層合成的專案慣例,換來 CJK 更銳利)。故有 `%Nd`
      右對齊數字的畫面(view_character/view_army/recruit)不同長度標籤的數字 x 位置
      略參差、偶有貼近/溢出框邊。**cosmetic、跨畫面一致、非 regression**。正解=寫個
      CJK 格寬感知的欄位對齊 helper(從框右緣算數字 x,不靠空格 padding),之後一次
      套用到所有這類畫面。

## 美術主題切換(F8 / ☰)—— 進行中(2026-07-10)

移植之外的加值功能:執行期切換整套美術主題(sprite+tileset+UI),沿用 C openkb 已抽出的美術。
資料層(邏輯/數值/地圖)不切換,只換美術。詳見 `docs/theme-switching-plan.md`。

### ✅ 已完成(commit 0e111a2 + a528480)
- [x] `internal/screen/loadart.go`:`LoadArt(fsys, dir)` 統一美術載入(取代 main.go/mobile.go 重複段),
      **先鋪 free 當底、再用主題覆蓋** → 版權主題缺的 UI 小件自動用 free。
- [x] ThemeManager:`InitThemes(fsys, order)` 過濾實際內建模組(該 dir 有 tileseta.png)、`CycleTheme`、
      `resolveThemes`(純函式)、`ActiveTheme`/`AvailableThemes`;`loadart_test.go` 單測(xvfb 全過)。
- [x] `render.LoadTilesetFS(fsys, dir)` 加 dir 參數;main.go/mobile.go 改用 `InitThemes([dos,genesis,amiga,free])`。
- [x] `app/game.go` `handleSystem` 攔 F8(ActThemeCycle)→ `CycleTheme`,鍵盤+觸控兩路。
- [x] `.gitignore` 擋 `internal/embedded/data/{dos,genesis,amiga,fmtowns}/`(版權美術本機 build 才有)。
- [x] **Amiga 主題接上並實測**:copy `openkb-code/qa-amiga/all/*.png`→`data/amiga/`;桌面 Xvfb 截圖
      確認 Amiga tileset/sprite 正確渲染、free 缺件回退正常。目前 dos/genesis 未抽 → 預設暫落 amiga。

### ⬜ 剩餘
- [x] **DOS 主題(DOS 當預設)**——✅ 完成(2026-07-10):`scripts/extract-dos-theme.sh`
      (kbcc 拆 256.CC → free 命名 .256;MCGA.DRV@0x032D VGA 調色盤;kbview 轉 PNG)→ `data/dos/`(61 張)。
      桌面 Xvfb 實測世界地圖 + 戰鬥 DOS 美術正確、sprite 去背正確、DOS 為預設。
- [x] **Android 觸控切主題入口**——✅ 完成(2026-07-10):`input/touch.go` 右上全域「主題」鈕
      (ActThemeCycle,觸控無 F8)。桌面 Xvfb + **Android 模擬器實機**驗證:APK 開機為 DOS 版
      NWC logo(DOS 預設 embed),點「主題」鈕 swipe-hold 循環 DOS→Amiga→free(genesis 缺席跳過),
      主題/作弊鈕與新 atlas 字集均正確渲染。
- [ ] **Genesis 主題**:`kb.bin` ROM 走 `md-rom.c`(不同管線,且 C 版僅 troop/villain/world,其餘回退 free)。
- [x] **切換 toast 提示**——✅ 完成(2026-07-11,`internal/app/toast.go`):切主題/音樂顯示置中金框提示。
- [ ]（選用）主題設定持久化(切主題後跨重啟記憶);worldmap 拼圖 p 觸控 —— ✅ 已修(見上,82c462b + 三層觸控)。

## 作弊選單(F12 / 觸控「作弊」)—— ✅ 完成(2026-07-10)

移植 C `debug_cheat_menu`(game.c:5052,F12)到 Go/Ebiten,Android 觸控可用。
- `internal/screen/cheat.go`:CheatMenuScreen,7 項作弊(+5000 金幣/+100 領導/學全法術/學魔法/
  神級隊伍/解鎖下洲+寶珠+飛行/立即勝利),直接改寫 GameState;'w'→WinScreen。LetterRects 觸控熱區。
- `input`:ActCheat + SymF12(keyboard/app 映 F12);`touch.go` 右上「作弊」鈕;worldmap 接 ActCheat。
- `cjk24.bin` 重烤(Noto Sans CJK,收全 Go 字串)1226→1305 字,補缺字(弊/母/鎖/伍)。
- 桌面 Xvfb 截圖驗證選單渲染、字全、版面不溢出。`-startcheat` debug flag。

## 📋 待辦工作項

- [x] **Android 推廣影片**——✅ 完成(2026-07-11):`~/openkb/android-promo/openkb-android-promo-20260711.mp4`
      (1920×1080,39s)。模擬器 adb screenrecord + swipe-hold 觸控錄製,涵蓋:DOS 預設世界地圖 →
      主題切換(主題鈕 DOS→Genesis→Amiga→free)→ 作弊選單(+套用 +5000 金幣)→ 權杖拼圖 →
      戰鬥(走進敵人)→ 戰鬥施法(施法→火球術→選敵→施放)。錄製過程順帶修好 拼圖 觸控不可達
      (commit 82c462b)。demo 用 mobile.go demoWorldMap 已 revert(未 commit)。
      **配樂**:adb screenrecord 不錄音,後製用 ffmpeg 混入 FM Towns 標題主題(`kb02.ogg`,首尾淡入淡出);
      版權音樂、個人 promo 用勿公開散布。
- [x] **音訊 / 音樂移植**——✅ 完成(2026-07-11,commit d947a3d):`internal/bgm`(ebiten/audio + vorbis
      循環)依畫面音樂場景循環播 OGG,對齊 C bgm.c 的 `scenes.ini`。各畫面實作 `bgm.Scener` 宣告場景
      (worldmap=field1/combat=combat/town=town/castle/siege/title/win);app 每幀切曲;F9 + 觸控「音樂」鈕
      靜音。9 首 FM Towns OGG 內建於 `data/music/`(gitignore,版權)。Android 模擬器實測:APK 開機
      無 crash、音樂鈕出現;桌面 headless 亦不崩。⚠ 實機發聲需真裝置(docker 模擬器 -no-audio 聽不到;
      adb screenrecord 也不錄音 → 影片配樂仍為後製混入,但曲目=遊戲實際 title BGM)。
      桌面 build 需 `libasound2-dev`(oto/v3 ALSA 後端);Android 用其自身音訊後端。
- [x] **Genesis 主題抽取**——✅ 完成(2026-07-10):`scripts/extract-genesis-theme.sh`(md-rom.c
      MD_Resolve dumper,SDL2 + env hook stub)→ `data/genesis/`(44 張:tileset+25 兵種+17 惡棍)。
      桌面 Xvfb 驗證世界地圖 + 戰鬥 Genesis 美術正確、兵種白底去背正確。UI/portrait/logo → free 補
      (對齊 C 版部分 Genesis)。四主題齊:dos(預設)/genesis/amiga/free。
