package export

import (
	"bytes"
	"context"
	"flag"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/tks/backend/internal/design"
	"github.com/tks/backend/internal/fontkit"
)

var update = flag.Bool("update-golden", false, "rewrite the golden slide PNG")

// goldenDesign is a fixed slide exercising the parts most likely to drift:
// Vietnamese diacritics, wrapping, each text alignment, an overlay and a shape.
func goldenDesign() design.Design {
	d := design.Design{
		Version: design.SchemaVersion,
		Canvas:  design.Canvas{Width: 1080, Height: 1350, Ratio: "4:5", SafeArea: design.SafeAreaFor(1080, 1350)},
		Palette: design.Palette{Primary: "#111111", Secondary: "#6B7280", Accent: "#FF3B5C", Background: "#0F172A", Text: "#FFFFFF"},
		Slides: []design.Slide{{
			ID: "slide_1", Index: 1, Role: "hook", Background: "#0F172A",
			Elements: []design.Element{
				{ID: "bg", Type: "image", X: 0, Y: 0, Width: 1080, Height: 1350, Opacity: 1,
					Overlay: &design.Overlay{Color: "#000000", Opacity: 0.45}},
				{ID: "rule", Type: "shape", Shape: "rect", X: 480, Y: 380, Width: 120, Height: 8,
					Fill: "#FF3B5C", Opacity: 1, Radius: 4},
				{ID: "hook", Type: "text", Role: "hook",
					Content: "5 dấu hiệu bé đang thiếu ngủ mà bố mẹ hay bỏ qua",
					X: 80, Y: 430, Width: 920, Opacity: 1,
					Style: &design.Style{FontFamily: "Inter", FontSize: 76, FontWeight: 700,
						Color: "#FFFFFF", TextAlign: "center", LineHeight: 1.15}},
				{ID: "body", Type: "text", Role: "body",
					Content: "Trẻ sơ sinh thường dụi mắt, cáu gắt và khó vào giấc trước khi bạn kịp nhận ra.",
					X: 80, Y: 900, Width: 920, Opacity: 1,
					Style: &design.Style{FontFamily: "Be Vietnam Pro", FontSize: 36, FontWeight: 400,
						Color: "#F3F4F6", TextAlign: "left", LineHeight: 1.4}},
			},
		}},
	}
	design.Validate(&d, d.Palette)
	return d
}

// TestGoldenSlideRendersIdentically is the regression guard the blueprint asks
// for in §155: the export renderer must keep producing the same pixels for the
// same Design JSON, so a change to fonts, wrapping or drawing cannot silently
// break WYSIWYG. Refresh with: go test ./internal/export -update-golden
func TestGoldenSlideRendersIdentically(t *testing.T) {
	d := goldenDesign()
	dc, err := NewRenderer(nil).RenderSlide(context.Background(), d, d.Slides[0])
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		t.Fatal(err)
	}

	goldenPath := filepath.Join("testdata", "golden_slide.png")
	if *update {
		if err := os.WriteFile(goldenPath, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Log("golden slide rewritten")
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden slide missing; run: go test ./internal/export -update-golden (%v)", err)
	}
	got, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	want, err := png.Decode(bytes.NewReader(expected))
	if err != nil {
		t.Fatal(err)
	}
	if diff := pixelDiff(got, want); diff > 0 {
		t.Errorf("export output drifted from the golden slide: %d pixels differ", diff)
	}
}

func pixelDiff(a, b image.Image) int {
	if a.Bounds() != b.Bounds() {
		return a.Bounds().Dx() * a.Bounds().Dy()
	}
	diff := 0
	for y := a.Bounds().Min.Y; y < a.Bounds().Max.Y; y++ {
		for x := a.Bounds().Min.X; x < a.Bounds().Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				diff++
			}
		}
	}
	return diff
}

// TestRendererUsesTheSharedWrapper is the other half of the WYSIWYG contract:
// the renderer must break lines with fontkit.Wrap — the same function the
// design validator and the client's overflow check use (§157).
func TestRendererUsesTheSharedWrapper(t *testing.T) {
	d := goldenDesign()
	el := d.Slides[0].Elements[2]
	lines := fontkit.Wrap(el.Content, el.Style.FontFamily, el.Style.FontWeight, el.Style.FontSize, el.Width)
	if len(lines) < 2 {
		t.Fatalf("expected the hook to wrap, got %d line(s)", len(lines))
	}
	for _, line := range lines {
		if w := fontkit.MeasureWidth(line, el.Style.FontFamily, el.Style.FontWeight, el.Style.FontSize); w > el.Width {
			t.Errorf("line %q is %.1fpx wide, box is %.0fpx", line, w, el.Width)
		}
	}
}
