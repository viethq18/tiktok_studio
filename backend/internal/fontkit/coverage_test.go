package fontkit

import (
	"testing"

	"github.com/tks/backend/internal/language"
)

// A content language is only offerable if every registry font can draw it.
// Otherwise a carousel looks right in the browser and exports as empty boxes —
// the silent failure §157 warns about. This test is what keeps the language
// registry and the font registry honest about each other.
func TestEveryRegisteredFontCoversEveryOfferedLanguage(t *testing.T) {
	for _, lang := range language.Supported {
		for _, fam := range Families {
			for _, weight := range []int{400, 700} {
				face, err := Face(fam.Name, weight, 48)
				if err != nil {
					t.Fatalf("%s %d: %v", fam.Name, weight, err)
				}
				for _, r := range lang.Probe {
					if r == ' ' {
						continue
					}
					if _, ok := face.GlyphAdvance(r); !ok {
						t.Errorf("%s cannot be offered as a content language: %s %d is missing %q",
							lang.EnglishName, fam.Name, weight, r)
					}
				}
			}
		}
	}
}

func TestWrapNeverExceedsMaxWidth(t *testing.T) {
	const maxWidth = 400.0
	text := "5 dấu hiệu cho thấy bé đang thiếu ngủ mà nhiều bố mẹ thường bỏ qua"
	for _, line := range Wrap(text, "Inter", 700, 48, maxWidth) {
		if w := MeasureWidth(line, "Inter", 700, 48); w > maxWidth+0.5 {
			t.Errorf("line %q measured %.1f > %.1f", line, w, maxWidth)
		}
	}
}
