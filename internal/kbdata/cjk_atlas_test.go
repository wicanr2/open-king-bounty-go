package kbdata

import (
	"path/filepath"
	"testing"
)

func testCJKAtlasPath(t *testing.T) string {
	t.Helper()
	return filepath.Join("testdata", "cjk24.bin")
}

func TestLoadCJKAtlas_NoPanicAndHeader(t *testing.T) {
	atlas, err := LoadCJKAtlas(testCJKAtlasPath(t))
	if err != nil {
		t.Fatalf("LoadCJKAtlas failed: %v", err)
	}
	if atlas.Width != 24 || atlas.Height != 24 {
		t.Fatalf("expected 24x24 glyphs, got %dx%d", atlas.Width, atlas.Height)
	}
	if atlas.Count() != 1226 {
		t.Fatalf("expected 1226 glyphs (0x4ca per header), got %d", atlas.Count())
	}
}

func TestCJKAtlas_KnownGlyphsNonZero(t *testing.T) {
	atlas, err := LoadCJKAtlas(testCJKAtlasPath(t))
	if err != nil {
		t.Fatalf("LoadCJKAtlas failed: %v", err)
	}

	for _, r := range []rune{'農', '王'} {
		g, ok := atlas.Glyph(r)
		if !ok {
			t.Fatalf("expected atlas to contain glyph for %q (U+%04X)", string(r), r)
		}
		if g.Width != 24 || g.Height != 24 {
			t.Fatalf("glyph %q: expected 24x24, got %dx%d", string(r), g.Width, g.Height)
		}
		if len(g.Alpha) != g.Width*g.Height {
			t.Fatalf("glyph %q: alpha length %d != %d*%d", string(r), len(g.Alpha), g.Width, g.Height)
		}
		nonZero := 0
		for _, b := range g.Alpha {
			if b != 0 {
				nonZero++
			}
		}
		if nonZero == 0 {
			t.Fatalf("glyph %q: alpha bitmap is entirely zero, expected visible dots", string(r))
		}
	}
}

func TestCJKAtlas_MissingGlyph(t *testing.T) {
	atlas, err := LoadCJKAtlas(testCJKAtlasPath(t))
	if err != nil {
		t.Fatalf("LoadCJKAtlas failed: %v", err)
	}
	// U+10FFFF is the max valid Unicode code point and should not be a
	// pre-rendered CJK glyph in this small atlas.
	if atlas.Has(0x10FFFF) {
		t.Fatalf("did not expect atlas to contain U+10FFFF")
	}
	if _, ok := atlas.Glyph(0x10FFFF); ok {
		t.Fatalf("Glyph() should report ok=false for an absent code point")
	}
}

func TestParseCJKAtlas_BadMagic(t *testing.T) {
	bad := make([]byte, 20)
	copy(bad, []byte("NOTACJK\x00"))
	if _, err := ParseCJKAtlas(bad); err == nil {
		t.Fatalf("expected error for bad magic")
	}
}

func TestParseCJKAtlas_Truncated(t *testing.T) {
	if _, err := ParseCJKAtlas([]byte{1, 2, 3}); err == nil {
		t.Fatalf("expected error for too-short input")
	}
}
