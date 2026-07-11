# 建置說明(Android，Go/Ebiten）

本專案是 **Android-only** 的 Go/Ebiten 移植;全程走 Docker,不污染系統。玩家只需 APK,以下是**開發/建置**用的可重現環境。

## 一、建置映像

工具鏈由 `docker/Dockerfile.android` 完整 pin 死(eclipse-temurin 17 + Go 1.24.13 + Android SDK
platform-34/build-tools 34+35/NDK 26.1 + Gradle 8.7 + ebitenmobile v2.9.9)。build-tools **35** 是
16KB `zipalign -P` 的必要條件;ebitenmobile 必須與 `go.mod` 的 ebiten 同版。

```bash
docker build -f docker/Dockerfile.android -t openkb-android-build .
```

持久 volume(加速、避免每次重抓依賴):
- `openkb-gocache` → `/go`(Go module 快取 + build 快取;首次建置後可離線)
- `openkb-gradle` → `/root/.gradle`(Gradle / AGP plugin 快取)

## 二、版權美術主題(選配,本機打包才有)

`dos` / `genesis` / `amiga` 主題是版權美術,**不入 repo**(`.gitignore`),需自備原始來源後抽出:

```bash
# 需 openkb-code(C 版)與對應原始資料;產物落 internal/embedded/data/<theme>/(gitignored)
OPENKB_CODE=/path/to/openkb-code ./scripts/extract-dos-theme.sh
OPENKB_CODE=/path/to/openkb-code GEN_ROM=/path/to/kb.bin ./scripts/extract-genesis-theme.sh
```

缺主題也能建置(只有 `free`,遊戲仍完整可玩)。`//go:embed all:data` 會把已抽好的主題全部
編進 `libgojni.so`;開機 logcat `art theme: dos (available:[dos genesis amiga free])` 可驗入包。

## 三、Debug APK

```bash
docker run --rm \
  -v openkb-gocache:/go -v openkb-gradle:/root/.gradle \
  -v "$PWD":/src \
  openkb-android-build bash /src/scripts/build-apk.sh
# → android/app/build/outputs/apk/debug/app-debug.apk
```

## 四、正式簽章 Release APK（含 16KB 對齊）

Android 15+ 要求 native `.so` 支援 16KB page。`build-release.sh` 做兩層對齊:
① **ELF LOAD 段 16KB** — bind 帶 `-Wl,-z,max-page-size=16384,-z,common-page-size=16384`;
② **APK 內 16KB** — `zipalign -P 16`（build-tools 35;必須在 apksigner 之前）。

```bash
mkdir -p ~/openkb/android-release          # keystore + APK 放 repo 外
docker run --rm \
  -v openkb-gocache:/go -v openkb-gradle:/root/.gradle \
  -v "$PWD":/src -v ~/openkb/android-release:/out \
  -e KS_PASS='你的keystore密碼' \
  openkb-android-build bash /src/scripts/build-release.sh
# → /out/openkb-release.apk（+ 首次執行產 /out/openkb-release.jks）
```

驗證(腳本自動印):`apksigner verify` 應 v2/v3 scheme = true;`zipalign -c -P 16` 應 OK;
四個 ABI 的 `libgojni.so` LOAD align 應為 `0x4000`(16KB)。

**keystore 紀律(重要)**:
- `openkb-release.jks` 與密碼是**機密**,存 repo 外(如 `~/openkb/android-release/`),**勿 commit**。
- 之後每次 release 用**同一把** keystore 簽,手機才能覆蓋升級不清存檔。
- release 簽章 ≠ debug 簽章 → 手機上首次從 debug 換 release(或反之)須先 `adb uninstall`(會清存檔)。

## 五、裝到實機

```bash
docker run --rm --privileged \
  -v /dev/bus/usb:/dev/bus/usb -v adbkeys:/root/.android \
  -v ~/openkb/android-release:/rel \
  openkb-android-build bash -lc '
    export PATH=$ANDROID_HOME/platform-tools:$PATH
    adb devices -l
    adb uninstall com.wicanr2.openkb 2>/dev/null   # 換簽章時才需要
    adb install /rel/openkb-release.apk
    adb shell am start -n com.wicanr2.openkb/.MainActivity'
```
`adbkeys` volume 持久化 USB 授權金鑰(已授權的手機不必重按對話框)。

## 六、模擬器驗證（選配）

`emu-verify.sh` 需要**含 emulator + system-image + AVD** 的映像。在建置映像上補:

```bash
# 於容器內(或另做一層映像):
sdkmanager --install "emulator" "system-images;android-34;google_apis;x86_64"
avdmanager create avd -n okb -k "system-images;android-34;google_apis;x86_64" --device pixel
```

執行(需 KVM):
```bash
docker run --rm --device /dev/kvm \
  -v "$PWD":/src -v ~/openkb/android-release:/out \
  <含模擬器的映像> bash /src/scripts/emu-verify.sh
# 煙霧:boot → 裝 → 啟動 → 驗 logcat 主題 + 截圖
```

**模擬器實測教訓**(swiftshader):
- `adb input keyevent/tap` **到不了 ebiten**(同 xdotool→SDL);用 `adb shell input swipe x y x y 260`(按住)。
- 觸控導覽 flaky(首觸常漏接):驗行為別靠固定 sleep+單次點,用**重試迴圈**
  (force-stop→重啟→導覽→查 logcat/截圖 byte,≤5 次);讓 app 印 logcat 抓 log 比比對截圖穩。
- 首次 immersive 全螢幕會跳系統「Viewing full screen / Got it」對話框,吃掉第一個 app 觸控:先點掉它
  (**別**用「從頂部下滑」= 會拉出通知欄蓋畫面)。座標基準用**截圖實際尺寸**(app 鎖橫向 1920×1080)換算
  letterbox,不是 `wm size`(可能回直向)。
- 選角畫面用**職業鈕**啟動遊戲(騎士 logical ≈97,160),右下是「返回」不是 Confirm。
- 驗每局隨機:charselect 會印 `new game: seed=.. scepter=..`,兩局比 scepter 是否不同即可
  (確定性主證仍是 `internal/gamestate` 的單元測試)。tileset 切換也可改用**桌面 -shot**(deskgl 映像 +
  `-startclass 0 -theme <dos|genesis|amiga|free> -shot out.png`)決定性截圖,跑的是同一份 LoadArt 碼。

## 版本一覽（改版時同步 Dockerfile）

| 元件 | 版本 |
|---|---|
| base | eclipse-temurin:17-jdk-jammy（Ubuntu 22.04, JDK 17.0.19）|
| Go | 1.24.13 |
| Android platform / build-tools | android-34 / 34.0.0 + 35.0.0 |
| NDK | 26.1.10909125 |
| Gradle / AGP | 8.7 / 8.1.4 |
| ebitenmobile / ebiten | v2.9.9（對齊 go.mod）|
