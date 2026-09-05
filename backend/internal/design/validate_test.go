package design

import "testing"

func baseDesign(text string, width float64) *Design {
	return &Design{
		Canvas: Canvas{Width: 1080, Height: 1350, Ratio: "4:5"},
		Slides: []Slide{{
			ID: "slide_1",
			Elements: []Element{
				{ID: "bg", Type: "image", X: 0, Y: 0, Width: 1080, Height: 1350, Opacity: 1},
				{ID: "t1", Type: "text", Role: "hook", Content: text, X: 80, Y: 400, Width: width, Opacity: 1,
					Style: &Style{FontFamily: "Inter", FontSize: 72, FontWeight: 700, Color: "#FFFFFF", TextAlign: "center", LineHeight: 1.2}},
			},
		}},
	}
}

func TestValidateClampsTextToSafeArea(t *testing.T) {
	d := baseDesign("5 dấu hiệu bé thiếu ngủ", 2000) // wider than the canvas
	res := Validate(d, Palette{})
	if res.Fatal() {
		t.Fatalf("unexpected fatal issues: %s", res.Error())
	}
	el := d.Slides[0].Elements[1]
	safe := d.Canvas.SafeArea
	if el.Width > float64(safe.Width) {
		t.Errorf("width %.0f was not clamped to safe width %d", el.Width, safe.Width)
	}
	if el.X < float64(safe.X) || el.X+el.Width > float64(safe.X+safe.Width) {
		t.Errorf("element escapes the safe area: x=%.0f w=%.0f safe=%+v", el.X, el.Width, safe)
	}
}

func TestValidateRejectsFontsOutsideTheRegistry(t *testing.T) {
	d := baseDesign("Xin chào", 800)
	d.Slides[0].Elements[1].Style.FontFamily = "Comic Sans MS"
	Validate(d, Palette{})
	if got := d.Slides[0].Elements[1].Style.FontFamily; got != "Inter" {
		t.Errorf("unregistered font survived validation: %q", got)
	}
}

func TestValidateReportsOverflowWithoutBeingFatal(t *testing.T) {
	long := ""
	for i := 0; i < 40; i++ {
		long += "dấu hiệu bé đang thiếu ngủ "
	}
	d := baseDesign(long, 900)
	res := Validate(d, Palette{})
	if res.Fatal() {
		t.Fatalf("overflow should be repairable, not fatal: %s", res.Error())
	}
	if len(res.OverflowReport()) == 0 {
		t.Error("expected an overflow report the worker can feed back to the AI")
	}
}

func TestValidateRejectsUnknownElementTypes(t *testing.T) {
	d := baseDesign("Xin chào", 800)
	d.Slides[0].Elements = append(d.Slides[0].Elements, Element{ID: "x", Type: "execute_javascript"})
	if res := Validate(d, Palette{}); !res.Fatal() {
		t.Error("an unsupported element type must fail validation")
	}
}

func TestValidateRejectsDuplicateSlideIDs(t *testing.T) {
	d := baseDesign("Xin chào", 800)
	d.Slides = append(d.Slides, d.Slides[0])
	if res := Validate(d, Palette{}); !res.Fatal() {
		t.Error("duplicate slide ids must fail validation")
	}
}

func TestBrandPaletteWinsOverAISuggestion(t *testing.T) {
	d := baseDesign("Xin chào", 800)
	d.Palette = Palette{Primary: "#00FF00", Accent: "#00FF00"}
	Validate(d, Palette{Primary: "#123456", Accent: "#654321"})
	if d.Palette.Primary != "#123456" || d.Palette.Accent != "#654321" {
		t.Errorf("brand colours must override AI colours, got %+v", d.Palette)
	}
}
