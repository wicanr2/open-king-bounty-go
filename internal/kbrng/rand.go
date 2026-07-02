// Package kbrng 重現 C 版 openkb 用的亂數,確保與 C oracle 對得上 seed。
//
// C 版 KB_rand(min,max) = rand() % (max-min+1) + min (src/lib/kbstd.c),
// 底層是 glibc 的 rand()/random()。glibc 預設 TYPE_3:31 個 int32 狀態、
// SEP=3 的加法回饋(additive feedback)。用 Go math/rand 對不上,故在此逐位重現。
//
// 演算法來源:glibc stdlib/random_r.c(srandom_r / random_r,TYPE_3)。
// parity 由 P3 的黃金樣本(固定 seed → 期望序列)釘死。
package kbrng

// glibc TYPE_3 參數
const (
	deg3 = 31 // 狀態陣列長度 (DEG_3)
	sep3 = 3  // fptr 與 rptr 的間距 (SEP_3)
)

// Rand 是可注入的亂數來源,讓 gamestate/combat 能對 seed 做 parity 測試。
type Rand interface {
	// Between 回傳 [min,max] 內的整數,語意同 C 版 KB_rand(min,max)。
	Between(min, max int) int
	// Seed 重設種子,語意同 C 版 srand(seed)。
	Seed(seed uint32)
}

// GlibcRand 是 glibc rand()/random() 的精確重現。
type GlibcRand struct {
	r    [deg3]int32
	fptr int
	rptr int
}

// NewGlibc 建立並以 seed 初始化(等同 C 版 srand(seed) 後可 rand())。
func NewGlibc(seed uint32) *GlibcRand {
	g := &GlibcRand{}
	g.Seed(seed)
	return g
}

// Seed 重現 glibc srandom_r 的初始化(TYPE_3)。
func (g *GlibcRand) Seed(seed uint32) {
	if seed == 0 {
		seed = 1 // glibc:seed 0 視為 1
	}
	g.r[0] = int32(seed)
	// r[i] = (16807 * r[i-1]) mod 2147483647,用 Schrage 法避免溢位(同 glibc)
	word := int32(seed)
	for i := 1; i < deg3; i++ {
		hi := word / 127773
		lo := word % 127773
		word = 16807*lo - 2836*hi
		if word < 0 {
			word += 2147483647
		}
		g.r[i] = word
	}
	g.fptr = sep3
	g.rptr = 0
	// glibc 丟棄前 10*DEG_3 個輸出以打散初始狀態
	for i := 0; i < 10*deg3; i++ {
		g.next()
	}
}

// next 重現 glibc random_r 的單步(TYPE_3 加法回饋),回傳 [0, 2^31-1]。
func (g *GlibcRand) next() int32 {
	// 以 uint32 溢位相加(同 C 的 int32 wraparound 語意)
	g.r[g.fptr] = int32(uint32(g.r[g.fptr]) + uint32(g.r[g.rptr]))
	result := int32(uint32(g.r[g.fptr]) >> 1) // 取高位,去掉最低位
	g.fptr++
	if g.fptr >= deg3 {
		g.fptr = 0
	}
	g.rptr++
	if g.rptr >= deg3 {
		g.rptr = 0
	}
	return result
}

// Between 回傳 [min,max],語意同 C 版 KB_rand(min,max) = rand()%(max-min+1)+min。
func (g *GlibcRand) Between(min, max int) int {
	span := max - min + 1
	if span <= 0 {
		return min // 與 C 對齊防呆:非法區間回 min(C 版會 UB,這裡收斂)
	}
	return int(g.next())%span + min
}

// 確保 GlibcRand 實作 Rand 介面。
var _ Rand = (*GlibcRand)(nil)
