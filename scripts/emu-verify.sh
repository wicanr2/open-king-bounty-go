#!/usr/bin/env bash
# emu-verify.sh — 模擬器煙霧測試:boot AVD → 裝 APK → 啟動 → 驗 logcat 主題 + 截圖。
#
# 需要「含 emulator + system-image + AVD」的映像(比 build 映像多,見 docs/BUILD.md「模擬器驗證」);
# 執行需 --device /dev/kvm。預期掛載:repo → /src、輸出 → /out。
# 環境變數:AVD(預設 okb)、APK(預設 debug APK 路徑)。
#
# 注意(swiftshader 模擬器實測教訓):adb input tap 到不了 ebiten;要用 `input swipe`(按住)。
# 且觸控導覽 flaky —— 本煙霧只驗「開機 + 渲染 + 主題載入」,不做深層導覽;要驗每局隨機/tileset
# 切換,參考 docs/BUILD.md 的重試迴圈作法或改用桌面 -shot(deskgl 映像)。
set -uo pipefail
export ANDROID_HOME=/opt/android-sdk
export PATH=$ANDROID_HOME/platform-tools:$ANDROID_HOME/emulator:$PATH
: "${AVD:=okb}"
: "${APK:=/src/android/app/build/outputs/apk/debug/app-debug.apk}"
: "${OUT_DIR:=/out}"
PKG=com.wicanr2.openkb
mkdir -p "$OUT_DIR"

echo "=== boot emulator ($AVD) ==="
emulator -avd "$AVD" -no-window -no-audio -no-snapshot -no-boot-anim -gpu swiftshader_indirect -no-metrics >"$OUT_DIR/emu.log" 2>&1 &
adb start-server >/dev/null 2>&1; adb wait-for-device
B=""; for i in $(seq 1 80); do B=$(adb shell getprop sys.boot_completed 2>/dev/null | tr -d '\r'); [ "$B" = 1 ] && break; sleep 3; done
[ "$B" = 1 ] || { echo "BOOT FAILED"; tail -20 "$OUT_DIR/emu.log"; exit 1; }

echo "=== install + launch ==="
adb install -r "$APK" 2>&1 | tail -1
adb shell pm clear $PKG >/dev/null 2>&1
adb logcat -c
adb shell am start -n $PKG/.MainActivity >/dev/null 2>&1; sleep 8

echo "=== 驗證:主題載入(四主題入包)==="
THEME=$(adb logcat -d -s GoLog 2>/dev/null | grep -a "art theme" | tail -1)
echo "  $THEME"
echo "$THEME" | grep -q "available: \[dos genesis amiga free\]" && echo "  四主題入包: OK" || echo "  (主題行未如預期;確認 internal/embedded/data 已抽版權主題)"

echo "=== 截圖 ==="
adb exec-out screencap -p > "$OUT_DIR/emu-smoke.png"
echo "  $OUT_DIR/emu-smoke.png ($(wc -c <"$OUT_DIR/emu-smoke.png")B)"

adb emu kill >/dev/null 2>&1 || true
echo DONE
