#!/usr/bin/env bash
# build-apk.sh — debug APK(ebitenmobile bind → gradle assembleDebug)。
# 在 openkb-android-build 映像內執行(見 docs/BUILD.md)。
#
# 預期掛載:repo → /src、go 快取 volume → /go。
# 輸出:    android/app/build/outputs/apk/debug/app-debug.apk
#
# 版權主題(dos/genesis/amiga)須先抽到 internal/embedded/data/(scripts/extract-*-theme.sh);
# 缺就只有 free(APK 仍可建、可玩)。
set -euo pipefail
cd /src
export GOPATH="${GOPATH:-/go}" GOCACHE="${GOCACHE:-/go/gocache}" GOFLAGS=-mod=mod
export GOPROXY="${GOPROXY:-off}"   # 快取齊時離線建置;首次填包改設 GOPROXY=https://proxy.golang.org
git config --global --add safe.directory /src 2>/dev/null || true

echo "=== ebitenmobile bind → openkb.aar ==="
ebitenmobile bind -target android -javapkg com.wicanr2.openkb -o mobile/openkb.aar ./mobile
cp mobile/openkb.aar android/app/libs/openkb.aar

echo "=== gradle assembleDebug ==="
( cd android && gradle assembleDebug )
ls -la android/app/build/outputs/apk/debug/*.apk
