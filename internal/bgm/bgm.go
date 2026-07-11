// Package bgm 是背景音樂子系統:依當前畫面的「音樂場景」循環播放對應 OGG。
//
// 對齊 C 版 bgm.c(SDL2_mixer 依場景循環播 OGG,F9 切換);Go 版改用 ebiten/audio + vorbis。
// 場景→檔名對照讀 data/music/scenes.ini(同 C 格式:key = kbNN.ogg)。音樂為 FM Towns CDDA
// 抽出的版權素材 → data/music/ 一律 gitignore、只在本機 build 內建;公開版無此目錄 → 自動靜音。
//
// 用法:進入點(main/mobile 的 boot)呼叫 Init(fsys);app 每幀對當前畫面呼叫 PlayScene(scene)。
package bgm

import (
	"bytes"
	"io/fs"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

// sampleRate 對齊 FM Towns OGG 的 44100(vorbis.DecodeWithSampleRate 會重採樣到此值)。
const sampleRate = 44100

// Scener 是可選介面:畫面實作它以宣告自己的音樂場景(如世界地圖="field1"、戰鬥="combat")。
// 未實作的畫面沿用目前音樂(不打斷)。
type Scener interface {
	MusicScene() string
}

var (
	mu      sync.Mutex
	ctx     *audio.Context
	fsys    fs.FS
	scenes  map[string]string // 場景 → 檔名(如 "field1" → "kb06.ogg")
	cache   map[string][]byte // 檔名 → OGG bytes(懶載,避免一次解全部)
	current string            // 目前播放中的場景
	player  *audio.Player
	enabled = true
)

// Init 建立音訊 context 並讀 scenes.ini。fsys 缺 music/scenes.ini(公開版無版權音樂)即靜音。
// 建 audio.Context 一程式僅能一次,故以 ctx==nil 守衛;需在遊戲迴圈啟動後呼叫(音訊裝置就緒)。
func Init(f fs.FS) {
	mu.Lock()
	defer mu.Unlock()
	fsys = f
	scenes = map[string]string{}
	cache = map[string][]byte{}
	b, err := fs.ReadFile(f, "music/scenes.ini")
	if err != nil {
		return // 無音樂資料 → 靜音
	}
	parseScenes(string(b))
	if ctx == nil {
		ctx = audio.NewContext(sampleRate)
	}
}

// parseScenes 解析 scenes.ini:每行 `key = value.ogg`,';' / '#' 開頭為註解。
func parseScenes(s string) {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == ';' || line[0] == '#' {
			continue
		}
		i := strings.IndexByte(line, '=')
		if i < 0 {
			continue
		}
		k := strings.TrimSpace(line[:i])
		v := strings.TrimSpace(line[i+1:])
		if k != "" && v != "" {
			scenes[k] = v
		}
	}
}

// PlayScene 切到指定音樂場景並循環播放。與目前相同(或 scene=="" / 未對照 / 靜音資料缺)時
// 不動作,故可每幀安全呼叫。
func PlayScene(scene string) {
	mu.Lock()
	defer mu.Unlock()
	if ctx == nil || scene == "" || scene == current {
		return
	}
	file, ok := scenes[scene]
	if !ok {
		return // 未對照的場景 → 維持目前音樂,不打斷
	}
	data := load(file)
	if data == nil {
		return
	}
	st, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return
	}
	loop := audio.NewInfiniteLoop(st, st.Length())
	p, err := ctx.NewPlayer(loop)
	if err != nil {
		return
	}
	if player != nil {
		player.Close()
	}
	player = p
	current = scene
	if enabled {
		p.Play()
	}
}

// load 懶載並快取 OGG bytes(缺檔回 nil)。
func load(file string) []byte {
	if b, ok := cache[file]; ok {
		return b
	}
	b, err := fs.ReadFile(fsys, "music/"+file)
	if err != nil {
		return nil
	}
	cache[file] = b
	return b
}

// SetEnabled 開關 BGM(對齊 C F9 靜音),暫停/續播目前曲目。
func SetEnabled(on bool) {
	mu.Lock()
	defer mu.Unlock()
	enabled = on
	if player == nil {
		return
	}
	if on {
		player.Play()
	} else {
		player.Pause()
	}
}

// Toggle 反轉 BGM 開關並回傳新狀態。
func Toggle() bool {
	on := !Enabled()
	SetEnabled(on)
	return on
}

// Enabled 回傳目前 BGM 是否開啟。
func Enabled() bool {
	mu.Lock()
	defer mu.Unlock()
	return enabled
}
