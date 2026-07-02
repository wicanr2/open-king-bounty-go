package kbrng

import "testing"

// glibc rand() 在 srand(1) 下的前 10 個輸出(公認值,任何 glibc 系統可複現)。
// 這是 kbrng 對上 C oracle 的第一道釘子:若這組不過,parity 一定崩。
var glibcSeed1 = []int32{
	1804289383, 846930886, 1681692777, 1714636915, 1957747793,
	424238335, 719885386, 1649760492, 596516649, 1189641421,
}

func TestGlibcSeed1(t *testing.T) {
	g := NewGlibc(1)
	for i, want := range glibcSeed1 {
		got := g.next()
		if got != want {
			t.Fatalf("seed=1 第 %d 個: got %d, want %d", i, got, want)
		}
	}
}

// glibc rand() 在 srand(42) 下的前 10 個輸出(由 docker glibc 產生,見 testdata 註解)。
var glibcSeed42 = []int32{
	71876166, 708592740, 1483128881, 907283241, 442951012,
	537146758, 1366999021, 1854614940, 647800535, 53523743,
}

func TestGlibcSeed42(t *testing.T) {
	g := NewGlibc(42)
	for i, want := range glibcSeed42 {
		got := g.next()
		if got != want {
			t.Fatalf("seed=42 第 %d 個: got %d, want %d", i, got, want)
		}
	}
}

// Between 對齊 C 版 KB_rand:rand()%(max-min+1)+min。
func TestBetweenMatchesC(t *testing.T) {
	g := NewGlibc(1)
	// 第一個輸出 1804289383,KB_rand(0,3) = 1804289383 % 4 = 3
	if got := g.Between(0, 3); got != 3 {
		t.Fatalf("KB_rand(0,3) got %d, want 3", got)
	}
	// 第二個 846930886,KB_rand(1,100) = 846930886 % 100 + 1 = 86 + 1 = 87
	if got := g.Between(1, 100); got != 87 {
		t.Fatalf("KB_rand(1,100) got %d, want 87", got)
	}
}

func TestBetweenInRange(t *testing.T) {
	g := NewGlibc(12345)
	for i := 0; i < 10000; i++ {
		v := g.Between(5, 9)
		if v < 5 || v > 9 {
			t.Fatalf("Between(5,9) 越界: %d", v)
		}
	}
}
