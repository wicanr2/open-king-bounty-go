#!/usr/bin/env bash
# extract-dos-theme.sh — 從原版 DOS King's Bounty 的 256.CC 抽出美術主題 PNG。
#
# 產物:internal/embedded/data/dos/*.png(free 版面同名),供 Go port 的 F8/☰ 主題切換
# 當「DOS」主題(偏好序第一 → 內建時即為預設)。**版權美術,gitignore、勿散布公開包。**
#
# 原理(沿用 C openkb 的解碼器,不重寫):
#   1. kbcc  -x 256.CC          → 拆出個別 <name>.256(容器內名 = free 名 + .256)。
#   2. MCGA.DRV 偏移 0x032D      → 256 色 VGA 調色盤(6-bit ×255/63);做成 palette.png。
#   3. kbview <name>.256 -p palette.png -o <name>.png  → 轉 PNG(VGA,水平 frame,保留
#      灰底 colorkey 供 Go colorKeyTopLeft 去背)。
#   4. select.png → select-0.png(Go LoadArt 讀 select-0)。
#
# 需求:openkb-code(C 版,提供 src/tools/kbcc.c + kbview.c);Docker(SDL1.2 + libpng)。
# 用法:OPENKB_CODE=/path/to/openkb-code DOS_CC=/path/to/256.CC ./scripts/extract-dos-theme.sh
set -euo pipefail

OPENKB_CODE="${OPENKB_CODE:-/home/anr2/openkb/openkb-code}"
DOS_CC="${DOS_CC:-$OPENKB_CODE/dos-orig/kings-bounty/256.CC}"
GO_REPO="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$GO_REPO/internal/embedded/data/dos"
IMG="${BUILD_IMAGE:-openkb-build-sdl2}"

[ -f "$DOS_CC" ] || { echo "找不到 256.CC: $DOS_CC"; exit 1; }
mkdir -p "$OUT"

docker run --rm \
  -v "$OPENKB_CODE:/src" -v "$OUT:/out" -v "$(dirname "$DOS_CC"):/dos" \
  --entrypoint bash "$IMG" -lc '
set -e
export DEBIAN_FRONTEND=noninteractive SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy
apt-get update >/dev/null 2>&1
apt-get install -y libsdl1.2-dev libsdl-image1.2-dev libpng-dev gcc python3-pil >/dev/null 2>&1
cd /src/src/tools
gcc -g kbcc.c -o kbcc
gcc -g `sdl-config --cflags` -DHAVE_LIBSDL -DHAVE_LIBPNG kbview.c ../../vendor/savepng.c \
    -o kbview `sdl-config --libs` -lSDL_image -lpng
W=/tmp/dosdump; rm -rf "$W"; mkdir -p "$W"; cd "$W"
cp /dos/256.CC .
/src/src/tools/kbcc -x 256.CC >/dev/null 2>&1
# VGA 調色盤:MCGA.DRV @ 0x032D,256 色 6-bit → 8-bit
python3 - <<PY
from PIL import Image
d=open("MCGA.DRV","rb").read(); off=0x032D; pal=[]
for i in range(256):
    pal += [(d[off+i*3+c]*255)//63 for c in range(3)]
im=Image.new("P",(16,16)); im.putpalette(pal); im.save("palette.png")
PY
for f in *.256; do n="${f%.256}"; /src/src/tools/kbview "$f" -p palette.png -o "/out/$n.png" >/dev/null 2>&1 || true; done
[ -f /out/select.png ] && mv -f /out/select.png /out/select-0.png || true
echo "抽出 $(ls /out/*.png | wc -l) 張 → /out"
'
echo "完成:$OUT"
echo "DOS 有 tileseta.png → Go InitThemes 偏好序 [dos,genesis,amiga,free] 會以 DOS 為預設。"
