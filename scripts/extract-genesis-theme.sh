#!/usr/bin/env bash
# extract-genesis-theme.sh — 從原版 Genesis(Mega Drive)King's Bounty ROM(kb.bin)抽出
# 美術主題 PNG → internal/embedded/data/genesis/*.png(free 版面,gitignore、勿散布)。
#
# 沿用 C openkb 的 md-rom.c 解碼器(MD_Resolve):C 版 Genesis 主題本身僅實作
# tileset / troop / villain(其餘 UI/portrait/背景/title 由引擎 fallback 到 free);
# 本腳本亦只抽這三類,其餘在 Go 端由 LoadArt 的 free 底層自動補上。
#
# 管線:內嵌一支 dump_md.c,連結引擎 lib(含 md-rom.c),對 kb.bin 呼叫 MD_Resolve:
#   - GR_TILE i (0..71) → 逐格排成 free 的 tileseta.png(0-35)/ tilesetb.png(36-71)
#   - GR_TROOP i → <troopname>.png(id 序同 free troops)
#   - GR_VILLAIN i → <villainname>.png(id 序同 free villains)
# 每張以 RGB surface 存(各資源自帶 CRAM palette,blit 轉 RGB 避免調色盤衝突)。
#
# 用法:OPENKB_CODE=/path/to/openkb-code GEN_ROM=/path/to/kb.bin ./scripts/extract-genesis-theme.sh
set -euo pipefail
OPENKB_CODE="${OPENKB_CODE:-/home/anr2/openkb/openkb-code}"
GEN_ROM="${GEN_ROM:-$OPENKB_CODE/genesis-orig/kb.bin}"
GO_REPO="$(cd "$(dirname "$0")/.." && pwd)"
OUT="$GO_REPO/internal/embedded/data/genesis"
IMG="${BUILD_IMAGE:-openkb-build-sdl2}"
[ -f "$GEN_ROM" ] || { echo "找不到 kb.bin: $GEN_ROM"; exit 1; }
mkdir -p "$OUT"

# 內嵌 dump_md.c(寫進 openkb-code/src/tools 供連結引擎相對路徑)
cat > "$OPENKB_CODE/src/tools/dump_md.c" <<'CSRC'
#include <SDL.h>
#include <string.h>
#include <stdio.h>
#include "../lib/kbconf.h"
#include "../lib/kbres.h"
#include "../../vendor/savepng.h"

void* MD_Resolve(KBmodule *mod, int id, int sub_id);

static const char *TROOPS[] = {"peas","spri","mili","wolf","skel","zomb","gnom","orcs",
 "arcr","elfs","pike","noma","dwar","ghos","kght","ogre","brbn","trol","cavl","drui",
 "arcm","vamp","gian","demo","drag"};
static const char *VILLAINS[] = {"mury","hack","ammi","baro","drea","cane","mora","barr",
 "barg","rina","ragf","mahk","auri","czar","magu","urth","arec"};

/* sprite(troop/villain):輸出 24-bit 不透明 PNG、magenta 底、pad 到高 34(對齊 free)。
 * MD index 0 = 透明底 → 設 colorkey 後 blit 到 magenta 底;Go colorKeyTopLeft 認左上 magenta
 * 乾淨去背(同 DOS 灰底作法,避免用黑色當 key 而挖穿黑輪廓)。 */
#define FREE_TROOP_H 34
static void save_sprite(SDL_Surface *src, const char *path) {
    if (!src) return;
    SDL_SetColorKey(src, SDL_FALSE, 0); /* 關 colorkey,整張 blit(MD 自帶白底 = 透明色) */
    SDL_Surface *out = SDL_CreateRGBSurface(0, src->w, FREE_TROOP_H, 24, 0xFF0000,0xFF00,0xFF,0);
    SDL_FillRect(out, NULL, SDL_MapRGB(out->format, 255, 255, 255)); /* 白底,與 MD 底色一致 → pad 也白 */
    SDL_Rect d = { 0, 0, (Uint16)src->w, (Uint16)src->h };
    SDL_UpperBlit(src, NULL, out, &d);
    SDL_SavePNG(out, path);
    SDL_FreeSurface(out);
}

int main(int argc, char **argv) {
    if (argc < 3) { fprintf(stderr,"usage: dump_md kb.bin OUTDIR\n"); return 1; }
    SDL_Init(SDL_INIT_VIDEO);
    KBmodule mod; memset(&mod,0,sizeof(mod));
    mod.kb_family = KBFAMILY_MD;
    strncpy(mod.slotA_name, argv[1], sizeof(mod.slotA_name)-1);
    char p[1024];
    const int TW=48, TH=34;
    /* tileset:GR_TILE 0..71 → tileseta(0-35)/tilesetb(36-71),各 36*48 x 34 RGB */
    SDL_Surface *a = SDL_CreateRGBSurface(0, 36*TW, TH, 24, 0xFF0000,0xFF00,0xFF,0);
    SDL_Surface *b = SDL_CreateRGBSurface(0, 36*TW, TH, 24, 0xFF0000,0xFF00,0xFF,0);
    for (int i=0;i<72;i++) {
        SDL_Surface *t = (SDL_Surface*)MD_Resolve(&mod, GR_TILE, i);
        if (!t) continue;
        SDL_Surface *dst = (i<36)?a:b;
        SDL_Rect d = { (Sint16)((i%36)*TW), 0, TW, TH };
        SDL_UpperBlit(t, NULL, dst, &d); /* tile 不透明,直接 blit 成 RGB */
        SDL_FreeSurface(t);
    }
    snprintf(p,sizeof(p),"%s/tileseta.png",argv[2]); SDL_SavePNG(a,p);
    snprintf(p,sizeof(p),"%s/tilesetb.png",argv[2]); SDL_SavePNG(b,p);
    SDL_FreeSurface(a); SDL_FreeSurface(b);
    /* troops */
    for (int i=0;i<25;i++) {
        SDL_Surface *t = (SDL_Surface*)MD_Resolve(&mod, GR_TROOP, i);
        if (!t) continue;
        snprintf(p,sizeof(p),"%s/%s.png",argv[2],TROOPS[i]); save_sprite(t,p);
        SDL_FreeSurface(t);
    }
    /* villains */
    for (int i=0;i<17;i++) {
        SDL_Surface *t = (SDL_Surface*)MD_Resolve(&mod, GR_VILLAIN, i);
        if (!t) continue;
        snprintf(p,sizeof(p),"%s/%s.png",argv[2],VILLAINS[i]); save_sprite(t,p);
        SDL_FreeSurface(t);
    }
    printf("genesis dump done -> %s\n", argv[2]);
    return 0;
}
CSRC

# env-sdl.c 的兩個 hook stub(只含 SDL2,不引 sdlcompat.h,故 SDL_UpperBlit/SDL_FillRect
# 為真 SDL2 函式而非遞迴)。md-rom.c/kbres.c 經 sdlcompat 把 blit/fill 導到這兩個 hook。
cat > "$OPENKB_CODE/src/tools/dump_md_stubs.c" <<'CSTUB'
#include <SDL.h>
int KB_BlitSurface_hook(SDL_Surface *src, SDL_Rect *sr, SDL_Surface *dst, SDL_Rect *dr) {
    return SDL_UpperBlit(src, sr, dst, dr);
}
int KB_FillRect_hook(SDL_Surface *s, const SDL_Rect *r, Uint32 c) {
    return SDL_FillRect(s, (SDL_Rect*)r, c);
}
CSTUB

docker run --rm -v "$OPENKB_CODE:/src" -v "$OUT:/out" -v "$(dirname "$GEN_ROM"):/rom" \
  --entrypoint bash "$IMG" -lc '
set -e
export DEBIAN_FRONTEND=noninteractive SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy
apt-get update >/dev/null 2>&1
apt-get install -y libsdl2-dev libsdl2-image-dev libpng-dev gcc >/dev/null 2>&1
cd /src/src/tools
LIB="../lib/kbdir.c ../lib/kbfile.c ../lib/md-rom.c ../lib/dos-cc.c ../lib/dos-img.c ../lib/kbstd.c ../lib/kbres.c ../../vendor/strlcat.c ../../vendor/strlcpy.c"
gcc -g `sdl2-config --cflags` -DHAVE_LIBSDL -DHAVE_LIBPNG dump_md.c dump_md_stubs.c ../../vendor/savepng.c $LIB \
    -o dump_md `sdl2-config --libs` -lSDL2_image -lpng 2>&1 | grep -iE "error|undefined" | head -20 || true
[ -x dump_md ] || { echo "BUILD FAILED"; exit 1; }
./dump_md /rom/kb.bin /out
echo "抽出 $(ls /out/*.png 2>/dev/null | wc -l) 張 → /out"
'
echo "完成:$OUT"
rm -f "$OPENKB_CODE/src/tools/dump_md.c" "$OPENKB_CODE/src/tools/dump_md_stubs.c"
