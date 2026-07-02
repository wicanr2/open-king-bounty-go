# 輸入接縫:一套 Action,鍵盤與觸控共用

Go 重寫相對 C 版最重要的一個架構決定,不是畫面或音樂,而是**輸入怎麼進遊戲**。這頁說明為什麼要在寫任何畫面前先把它定死。

## C 版的問題:觸控是疊上去的補丁

openkb(C/SDL2)每個畫面用一張 `KB_keymap` 表描述當下有效的按鍵,交給 `KB_event()` 等輸入。Android 版的觸控,本質是「手指事件 → 合成一個 `SDLK_*` → 餵回 `KB_event()`」——**疊在鍵盤引擎上的 overlay**。能動,但觸控始終是二等公民,與核心邏輯纏在一起。

## Go 版的做法:語義 Action 當一等公民

把「輸入的語義」抽出來。鍵盤與觸控都只是**來源**,收斂成同一套 `Action`;畫面與遊戲狀態只認 `Action`,永遠不碰 raw key / raw touch。

```
鍵盤 (keyboard.go)  ─┐
                      ├─▶ Action ──▶ screen 狀態機 ──▶ gamestate
觸控 (touch.go, P7) ─┘
```

- `Action`:遊戲邏輯真正理解的語義(方向 / 確認 / 取消 / 字母快捷 / 數字 / 是非 / 捲動 / 系統)。
- `keyboard.go`:中性 `Key`(Sym + 字元)→ `Action` 的純函式。Ebiten 的 `ebiten.Key` 由 `cmd`/`render` 層轉成中性 `Key` 再進來——所以 `input` 套件**零引擎依賴、可 headless 測**。
- `touch.go`(P7):手指點 → `Action`,實作留到 P7,但介面現在就在。

## 另一半:每個畫面「宣告」自己的 Keymap

沿用 openkb「每畫面一張 keymap」的精神,但抽成語義宣告。每個 screen 提供一個 `Keymap`:當下接受哪些 Action、字母選項各是什麼標籤。

```go
// 城鎮選單宣告(對照 ui-design.md 的情境快捷列)
Keymap{
    Confirm: "選擇", Cancel: "離開",
    Letters: []LetterItem{{'a',"接任務"},{'b',"租船"},{'c',"情報"},{'d',"造橋"},{'e',"買攻城"}},
}
```

觸控層讀這張 `Keymap` 決定浮出哪些控制:有 `Directions` 就畫左下 D-pad;有 `Confirm/Cancel` 就畫右下 A/B;有 `Letters` 就把每顆做成帶標籤的按鈕(`[A 接任務]`);切畫面時控制列自動跟著換,不會一堆無用按鈕擋畫面。完整視覺設計見 [`android/ui-design.md`](android/ui-design.md) 與 6 張 mockup。

## 為什麼「趁現在」

`screen` 狀態機即將開工。若各畫面直接吃 Ebiten raw key,之後 touch 又得像 C 版那樣 overlay 補回去。**先把 Action + 每畫面 Keymap 宣告的接縫定死,`screen` 一開工就走對的接線,觸控到 P7 只是「照 Keymap 填按鈕」,不用重接。** 這正是深模組(窄介面、把複雜度收在模組內)與「別重踩 C 版覆轍」的落點。

## 現況(本次落地)

- `internal/input/action.go`:`Action` / `ActionKind` / `Keymap` / `LetterItem`(純語義)。
- `internal/input/keyboard.go`:中性 `Key` → `Action`(純函式,已測)。
- `internal/input/touch.go`:P7 觸控介面佔位(`TouchLayout` 讀 `Keymap`)。
- 測試 headless 全綠(無 GL 依賴)。

下一步(P2/P4 寫 screen 時):每個 screen 實作 `CurrentKeymap() Keymap` + 在 Update 消費 `Action`。
