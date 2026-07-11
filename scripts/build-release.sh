#!/usr/bin/env bash
# build-release.sh — 16KB 對齊 + 正式簽章 release APK。
# 在 openkb-android-build 映像內執行(見 docs/BUILD.md)。
#
# 預期掛載:repo → /src、go 快取 volume → /go、輸出目錄 → /out(keystore 也存這,repo 外)。
# 必要環境變數:KS_PASS(keystore 密碼)。選填:KEYSTORE / KS_ALIAS / OUT_DIR。
# 輸出:$OUT_DIR/openkb-release.apk(+ 首次執行會產 $KEYSTORE)。
#
# 兩層 16KB 對齊(Android 15+ 必備):
#   ① ELF LOAD 段 16KB —— bind 帶 -Wl,-z,max-page-size=16384,-z,common-page-size=16384
#   ② APK 內 16KB      —— zipalign -P 16(需 build-tools 35;必須在 apksigner 之前)
set -euo pipefail
cd /src
export GOPATH="${GOPATH:-/go}" GOCACHE="${GOCACHE:-/go/gocache}" GOFLAGS=-mod=mod
export GOPROXY="${GOPROXY:-off}"
: "${OUT_DIR:=/out}"
: "${KEYSTORE:=$OUT_DIR/openkb-release.jks}"
: "${KS_ALIAS:=openkb}"
: "${KS_PASS:?請設 KS_PASS(keystore 密碼)}"
: "${KS_DNAME:=CN=openkb, O=open-king-bounty, C=TW}"
BT="$ANDROID_HOME/build-tools/35.0.0"
[ -x "$BT/zipalign" ] || { echo "缺 build-tools 35(16KB zipalign)"; exit 1; }
git config --global --add safe.directory /src 2>/dev/null || true
mkdir -p "$OUT_DIR"

echo "=== [1] bind(16KB ELF 對齊)==="
ebitenmobile bind -target android -javapkg com.wicanr2.openkb \
  -ldflags "-extldflags=-Wl,-z,max-page-size=16384,-z,common-page-size=16384" \
  -o mobile/openkb.aar ./mobile
cp mobile/openkb.aar android/app/libs/openkb.aar

echo "=== [2] gradle assembleRelease(unsigned;不加 --offline,AGP plugin 需線上/快取解析)==="
( cd android && gradle assembleRelease )
UNSIGNED=$(ls android/app/build/outputs/apk/release/*-unsigned.apk 2>/dev/null | head -1)
[ -f "$UNSIGNED" ] || UNSIGNED=$(ls android/app/build/outputs/apk/release/*.apk | head -1)

echo "=== [3] release keystore(不存在才產;存 $OUT_DIR,勿進 repo)==="
if [ ! -f "$KEYSTORE" ]; then
  keytool -genkeypair -keystore "$KEYSTORE" -alias "$KS_ALIAS" -keyalg RSA -keysize 2048 \
    -validity 10000 -storepass "$KS_PASS" -keypass "$KS_PASS" -dname "$KS_DNAME"
  echo "  已產生 $KEYSTORE(請妥善保管:日後同 keystore 簽才能覆蓋升級不清存檔)"
fi

echo "=== [4] zipalign -P 16(16KB)→ apksigner 簽章 ==="
"$BT/zipalign" -f -P 16 -v 4 "$UNSIGNED" "$OUT_DIR/.aligned.apk" >/dev/null
"$BT/apksigner" sign --ks "$KEYSTORE" --ks-key-alias "$KS_ALIAS" \
  --ks-pass pass:"$KS_PASS" --key-pass pass:"$KS_PASS" \
  --out "$OUT_DIR/openkb-release.apk" "$OUT_DIR/.aligned.apk"
rm -f "$OUT_DIR/.aligned.apk" "$OUT_DIR/openkb-release.apk.idsig"

echo "=== [5] 驗證 ==="
"$BT/apksigner" verify -v --print-certs "$OUT_DIR/openkb-release.apk" | grep -iE "verified|scheme" || true
"$BT/zipalign" -c -P 16 -v 4 "$OUT_DIR/openkb-release.apk" >/dev/null && echo "  16KB APK 對齊: OK"
RE=$(ls "$ANDROID_NDK_HOME"/toolchains/llvm/prebuilt/*/bin/llvm-readelf 2>/dev/null | head -1)
if [ -n "$RE" ]; then
  tmp=$(mktemp -d); ( cd "$tmp" && unzip -oq "$OUT_DIR/openkb-release.apk" "lib/*" )
  for so in $(find "$tmp/lib" -name '*.so'); do
    al=$("$RE" -l "$so" 2>/dev/null | awk '/LOAD/{print $NF; exit}')
    echo "  $(echo "$so" | sed "s#$tmp/##") LOAD align=$al $([ "$al" = 0x4000 ] && echo '(16KB OK)' || echo '(!! 非16KB)')"
  done
  rm -rf "$tmp"
fi
echo "=== 完成: $OUT_DIR/openkb-release.apk ==="
