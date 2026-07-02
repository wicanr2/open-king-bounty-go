# open-king-bounty-go

《御封戰將》(King's Bounty, 1990)以 **Go + [Ebiten](https://ebitengine.org)** 乾淨重寫,**Android 為第一目標**,桌面(Windows/macOS/Linux)與 Web(WASM)一併帶出。

C 版 [`openkb-cht`](https://github.com/wicanr2/open-king-bounty-cht) 作為**行為真值 oracle**:資料格式與遊戲公式都已破解摸透,這裡是「把已知邏輯翻成 Go」,不移植 C runtime。重寫計畫全文見 openkb-cht 的 `docs/rewrite-go-ebiten.md`。

## 為什麼重寫

- **一份碼全平台**:Ebiten 純 Go,同碼出 Android / Windows / macOS / Linux / Web。手機用 `ebitenmobile bind`,不再維護「SDL2 桌面 + NDK + 觸控 overlay」雙軌。**Android APK 是純 Go,零 C/glibc 依賴。**
- **記憶體安全**:C 版半年修的 NULL deref、`KB_fgets` 文字模式錯位、路徑雙 data 等,在 Go 結構上不會發生。
- **好測**:邏輯與渲染分離,邏輯層 headless 對 C oracle 做 parity。

## 架構(`internal/`,按功能垂直切)

| 套件 | 職責 |
|---|---|
| `kbrng` | glibc `rand()` 的純 Go 重現(parity 錨,seed 對得上 C) |
| `kbdata` | 資料層:讀解所有格式 → 唯讀 `Assets`(窄介面 `Load(dir)`) |
| `gamestate` | 純遊戲狀態與規則(領導力/寶箱/招兵/晉升),無渲染、可 headless |
| `combat` | 戰鬥狀態機 + AI,無渲染 |
| `render` | Ebiten 繪製(worldmap/combat/ui/cjktext),只讀狀態 |
| `input` | 桌面鍵盤 + 手機觸控 → 同一套 Action |
| `screen` | 畫面流程狀態機 | 
| `save` `audio` | 跨平台存讀檔 / BGM |

## 開發

```sh
go test ./internal/...      # 邏輯層(無 GL,快)
go build ./cmd/openkb       # 桌面(需 GL/X11 開發標頭)
ebitenmobile bind -target android -o openkb.aar ./mobile   # Android AAR
```

## 版權

同 openkb-cht 政策:free 自由美術/資料公開;原版美術與 FM-Towns 音樂只供正版擁有者的完整版,不散布。
