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

## 5. 各主題原始資源位置與抽取狀態(已確認 2026-07-10)

C 版 openkb(`openkb-code/`)四主題(free/DOS/Genesis/Amiga)**執行期直接解碼原版**;Go port 需**個別 PNG**。
原始資源與已 dump 產物位置(皆在 `openkb-code/`,gitignore):

| 主題 | 原始資源 | C 解碼器 | 已 dump 產物 | Go 用個別 PNG |
|---|---|---|---|---|
| **Amiga** | `amiga-orig/`(unpack GAME 散檔) | `src/lib/amiga-data.c` | ✅ `qa-amiga/all/*.png`(62 張,`tools/amiga_batch.py`+`amiga_decode.py`) | ✅ **已接上**(copy 進 `data/amiga/`) |
| **DOS** | `dos-orig/kings-bounty/256.CC`+`416.CC`+`KB.EXE`(雜湊資源容器) | `DOS_Resolve`(free-data/kbres.c) | 部分:`qa-spr/{troop,villain,portrait}_dos.bmp`、`qa-themes/tiles_dos.bmp`/`iso_dos.bmp`(長條合成,非個別) | ⬜ 待 dump |
| **Genesis** | `genesis-orig/kb.bin`(Sega ROM) | `src/lib/md-rom.c`(MD_Resolve;**僅 troop/villain/world 完成**,tile/UI/cursor/select/title 仍回退 free) | `qa-md/`、`qa-spr/md_*` | ⬜ 待 dump(且本身不完整) |

**關鍵**:DOS/Genesis 在 C 已能解碼(長條 bmp 為證),**不必從零逆向**。產個別 PNG 的兩條路:
1. **C `kbview` 逐資源 dump**(`src/tools/kbview.c` 有 `SDL_SavePNG`):建好 kbview,對 DOS 模組每個 free 同名資源 dump 成 PNG(free 版面),放 `internal/embedded/data/dos/`。這是最忠實、最省的「沿用」。
2. **切既有長條 bmp**:`tiles_dos.bmp`→tileseta/b、`troop_dos.bmp`→各兵種——但長條未涵蓋全部 63 asset(UI/背景/title 等缺),仍需路 1 補齊。

- 漸進上線:§3.0 free 底層回退 → DOS 只要抽好 tileset+主要 sprite 就能先切換看到,缺件自動用 free。
- build 整合:build 前 `data/<theme>` 存在就 embed(`//go:embed all:data`);不存在則 ThemeManager 依 tileseta.png 偵測後自動缺席。
- ⚠ 版權:`internal/embedded/data/{dos,genesis,amiga,fmtowns}/` 已 gitignore;公開 repo 與對外 APK 不得含,個人 build 才內建。

## 6. 分階段(狀態)

- **P1 keystone**:✅ **完成**(commit `0e111a2`+`a528480`)。`LoadArt(fsys, dir)` 統一美術載入(先 free 底再覆蓋)+ ThemeManager(`InitThemes`/`CycleTheme`/`resolveThemes`)+ `app/game.go` 攔 F8。單測 `loadart_test.go`(過濾/預設/循環)xvfb 全過。
- **Amiga 主題**:✅ **完成並實測**。copy `qa-amiga/all/*.png`→`data/amiga/`;桌面 Xvfb 截圖確認 tileset/sprite 正確、free 缺件回退正常。目前 dos/genesis 未抽 → 預設暫落 amiga。
- **P2 DOS(讓 DOS 當預設,下一步)**:用 C `kbview` dump DOS 個別 PNG → `data/dos/`(§5 路 1)。tileset 先(world/combat 立即可見),再 sprite/UI。DOSBox 目視對照。抽好即成預設(偏好序第一)。
- **P3 Genesis**:同法 dump `data/genesis/`(注意 C 版 MD_Resolve 僅 troop/villain/world,tile/UI 本就回退 free → Go 端也接受部分主題)。
- **P4 Android ☰ 觸控入口**:F8 觸控不可達 → 右上 ☰ 系統選單(`Keymap.System`=[ActThemeCycle,…]),手機才點得到切換。順修 worldmap 字母列出界(拼圖 p 觸控不可達)。
- **P5**:切換 toast +（選用）設定持久化 + 文件 + 對外 build 排除版權目錄把關。

## 7. 已定案 / 註記

- 範圍:**整套美術**(art-only module;資料層不切換)。已定案,見開頭。
- 主題集 = C openkb 實際四主題 **free / DOS / Genesis / Amiga**(先前誤植 FM Towns;FM Towns 在 C 只有音樂、圖形未完整 loader)。
- `tilesalt`(9 tile)**不**當可切換主題,維持原用途(神器圖示)。
- Genesis 主題**本身不完整**(C 版 tile/UI 仍回退 free),Go 端沿用此現況(free 底層回退剛好承接)。
