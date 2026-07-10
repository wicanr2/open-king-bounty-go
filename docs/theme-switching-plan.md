# 主題 / tileset 切換 計畫

> 目標:Android(與桌面)可在執行期切換**整套美術主題**(tileset + 所有 sprite + UI),**預設 DOS EGA**。
> 狀態:架構定案(art-only module),分階段實作中。
>
> **範圍定案(2026-07-10)**:使用者選「整套美術都換」。關鍵收斂——DOS / FM Towns / free 三者
> **遊戲資料相同**(兵種數值、城鎮座標、land 地圖 = 邏輯層,維持 free),**只有美術 PNG 不同**。
> 故「主題」= 一組美術 PNG 目錄;切換 = 用新目錄重跑美術載入,`kbdata.Assets`(ini/land)不必重載。

## 1. 現況(2026-07-10)

- **只有骨架、沒有功能**。`input.ActThemeCycle`(對應 F8,`internal/input/keyboard.go` + `internal/app/game.go` 已把按鍵映射成這個 Action)只被**產生**,沒有任何畫面**消費**它——按 F8 目前不會換 tile。
- `Keymap.System []Action`(設計上要把 F8/F9/F10 收進觸控右上的 ☰ 選單)欄位存在,但**沒有畫面填它、也沒有 ☰ 選單 UI**。
- `render.Tileset` 只載入一套 free(`tileseta.png`+`tilesetb.png`),`screen` 用一個全域 `worldTileset`。桌面與 Android 都只有這一套。
- 對照:C openkb 桌面版有啟動時的 module 選擇(`game.c:810 select_module`),Go 移植尚未移植這套。

結論:主題切換要**新建**,不是接線既有功能。

## 2. 可切換的主題(修正版)

「完整世界地形主題」= 一套 72 個 48×34 tile(tile byte 0–71)。實際可行的有三套:

| 主題 | 來源 | 版權 | 打包方式 |
|---|---|---|---|
| **DOS EGA**(預設) | `dos-orig/kings-bounty/KB.EXE` 抽出 | 有 | gitignore,build 時內建;公開 repo / 對外 APK 不含 |
| **FM Towns** | `fmtowns-cd/` 抽出 | 有 | 同上 |
| **free 開放美術** | `free/tileseta.png`+`tilesetb.png`(已內建) | 無 | 隨 repo,公開版一定有 |

> ⚠ **修正**:先前討論提到的 `free/tilesalt.png` 只有 **432×34 = 9 個 tile**(是神器/寶物在地圖上的替代圖示,見 `internal/gamestate/artifact.go` 的 `invert_id`),**不是**完整地形主題,無法當可切換 theme。free 模組只提供**一套**完整世界 tileset。

## 3. 架構設計(art-only module,deep module)

### 3.0 keystone 重構:共用 `loadArt(fsys, dir)`
現況:`cmd/openkb/main.go` 與 `mobile/mobile.go` **各有一段幾乎重複的 ~20 塊美術載入序列**
(SetTileset/SetHero/SetSidebar/SetTroopSprite×N/SetPortrait×4/SetVillainFace×N/…),
main 走磁碟 `render.LoadSprite(dir,…)`、mobile 走 `render.LoadSpriteFS(embed,"free/…")`。
- **統一走 `fs.FS`**(桌面 `os.DirFS`,行動 embed),把整段抽成一個函式
  `screen.LoadArt(fsys fs.FS, dir string)`,`dir` 即模組名(`"free"`/`"dos"`/`"fmtowns"`)。
- main.go / mobile.go 各自只呼叫 `LoadArt(fsys, active)` 一行;**切換主題 = 換 dir 再呼叫一次**,
  所有全域 sprite/tileset setter 被新主題覆寫,一次全畫面生效。
- 缺檔沿用現況 best-effort(個別 asset 缺就 log 跳過),故版權主題只要部分抽好也能漸進上線。

### 3.1 ThemeManager
```
screen(或 app)層:
  type ThemeManager struct { fsys fs.FS; mods []string; active int }
    - New(fsys, order []string)  // order 例:["dos","fmtowns","free"];過濾出真正存在的目錄
    - Active() string
    - Cycle()                    // F8 / ☰:下一套(繞回)並 LoadArt(fsys, mods[active])
    - SetActive(name string)
    - Names() []string
```
- 預設 active = order 第一個**存在**的模組目錄(有 dos 就 dos,否則 free)。
- 「目錄是否存在」:偵測該 dir 下有無 `tileseta.png`(代表該美術模組已內建)。

### 3.2 消費 ActThemeCycle
- 在共用 action 分派點(app 層攔 `ActThemeCycle`,先於畫面 Update)→ `themes.Cycle()` → toast「主題:DOS EGA」。
- 持久化:獨立設定檔(Android `getFilesDir`,桌面 XDG),下次啟動沿用。

### 3.3 資料層不動
- `kbdata.Assets`(ini/land.org/strings/font)維持只從 free 載入一次;主題不影響遊戲邏輯,只影響美術。
- (若日後想要「DOS 版兵種名/文字」也切換,再擴 ThemeManager 到重載 Strings;目前不做。)

## 4. Android / 桌面 UI

- **桌面**:F8 直接切(已有按鍵映射)。
- **Android(觸控)**:實作右上 **☰ 系統選單**——`Keymap.System` 填入 `[ActThemeCycle, ActMusicToggle, ActQuitSave]`,`input.TouchLayout` 在右上角放一顆 ☰ 熱區,點開列出這些系統項。主題切換就從這裡進。
  - 順帶解決另一個實機發現的缺陷:worldmap 觸控字母列 `x=82+i*32` 在第 8 顆就出界,**拼圖(p)按鈕觸控不可達**——系統/次要動作移進 ☰ 可一併釋放底列空間。

## 5. 版權主題的抽取管線(DOS / FM Towns)— 最大工作量

**抽取面積**:free/ 有 **63 個 PNG**。整套美術主題要每張都有對應品 → DOS 與 FM Towns **各約 63 張、合計 ~126 張**。
分兩類:
- **tileset 類**(tileseta/b/comtiles/view/select 等圖集):固定格點,可程式化切割 dump。
- **sprite 類**(cursor/各兵種/各惡棍臉/職業立繪/棲地背景/title/logo/endpic…):數量多、尺寸各異。

做法:**用 C openkb 當抽取 oracle**(kbres.c 已能讀原版 EGA / FM Towns 資源)。指向 `dos-orig/kings-bounty/KB.EXE`
與 `fmtowns-cd/`,把各資源 render 成 free/ 同名 PNG(同尺寸/同 frame 佈局),產物落在 gitignore 的
`internal/embedded/data/dos/`、`.../fmtowns/`。

- 平行化:抽取是大量機械工作,可用 subagent fan-out(見 rule 45)分批抽 sprite,逐張目視對照 DOSBox / FM Towns 原版。
- 漸進上線:§3.0 缺檔 best-effort → 只要抽好一部分就能先切換看到(缺的 asset 暫用 free 或 log 跳過)。
- build 整合:build 前若 `data/dos` 存在就 embed;不存在則該主題自動缺席(ThemeManager 依 tileseta.png 偵測)。
- ⚠ 版權:`data/dos`、`data/fmtowns` 一律 gitignore;公開 repo 與對外散布的 APK 不得含。個人 build 才內建。

## 6. 分階段

- **P1(keystone,零版權、可先驗)**:§3.0 `LoadArt(fsys, dir)` 統一重構 + ThemeManager + F8 消費 + ☰ 觸控選單。
  用 free +（暫時複製一份調色的 `free2`）證明「執行期整套換」端到端可行。單測涵蓋 Cycle / 缺目錄過濾 / 預設選擇。
- **P2**:DOS EGA 抽取管線(tileset 類先,~10 張圖集)→ 內建 → 設為預設,world/combat 目視對照 DOSBox。
- **P3**:DOS EGA sprite 類抽完(兵種/惡棍/立繪/UI,~50 張,subagent fan-out)。
- **P4**:FM Towns 抽取(tileset + sprite)。
- **P5**:設定持久化 + toast + 文件 + 對外 build 排除版權目錄的把關。

## 7. 已定案 / 註記

- 範圍:**整套美術**(art-only module;資料層不切換)。已定案,見開頭。
- `tilesalt`(9 tile)**不**當可切換主題,維持原用途(神器圖示)。
- 實機附帶發現(P1 一併處理):worldmap 觸控字母列出界 → **拼圖(p)按鈕觸控不可達**,系統/次要動作移進 ☰ 釋放空間。
