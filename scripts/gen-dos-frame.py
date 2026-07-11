#!/usr/bin/env python3
# 生成 DOS 主題「金框」邊框圖 frame.png(320×200,play area 透明 + 四周金色浮雕邊)。
#
# 色階直接對齊 C 版 openkb DOS 世界地圖截圖(docs/screenshots/4-world-map.png)實測的
# 金框浮雕:黑外緣 → 暗金 → 亮金(內側受光)→ 深藍內線。play area = (16,8)-(304,192),
# 對齊 Go internal/screen 的 worldContentL/T/R + chromeContentB。距內容邊緣的距離場
# (轉角取 max = 方形斜接)決定每個 margin 像素取哪一階色。
#
# 產物 frame.png 屬版權美術衍生 → 放 internal/embedded/data/dos/(gitignore),個人 build
# 才內建;公開/free 版無此檔 → Go drawChromeFrame 自動退回平面金框帶。
#
# 用法:python3 scripts/gen-dos-frame.py [輸出路徑，預設 internal/embedded/data/dos/frame.png]
import sys
from PIL import Image

PROFILE = {                 # d = 距內容邊緣往外的像素數(1 = 緊貼內容)
    1: (0x00, 0x02, 0x66),  # 深藍內線
    2: (0xf9, 0xf0, 0x00),  # 最亮金(內側受光)
    3: (0xf9, 0xf0, 0x00),
    4: (0xe6, 0xcf, 0x00),  # 亮金
    5: (0xe6, 0xcf, 0x00),
    6: (0xd8, 0xc2, 0x00),  # 中金
    7: (0x82, 0x75, 0x00),  # 暗金(外側陰影)
    8: (0x00, 0x00, 0x00),  # 黑外緣
}
BLACK = (0, 0, 0)
L, T, R, B = 16, 8, 304, 192
W, H = 320, 200


def main():
    out = sys.argv[1] if len(sys.argv) > 1 else "internal/embedded/data/dos/frame.png"
    img = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    px = img.load()
    for y in range(H):
        for x in range(W):
            dx = (L - x) if x < L else (x - (R - 1)) if x >= R else 0
            dy = (T - y) if y < T else (y - (B - 1)) if y >= B else 0
            d = max(dx, dy)
            if d <= 0:
                continue  # play area 透明
            c = PROFILE.get(d, BLACK)
            px[x, y] = (c[0], c[1], c[2], 255)
    img.save(out)
    print("wrote", out, img.size)


if __name__ == "__main__":
    main()
