package design

import "testing"

func brandDesign(slides int) *Design {
	d := &Design{
		Canvas:  Canvas{Width: 1080, Height: 1350, Ratio: "4:5", SafeArea: SafeAreaFor(1080, 1350)},
		Palette: Palette{Background: "#FFFFFF", Text: "#111111", Accent: "#FF3B5C"},
	}
	for i := 0; i < slides; i++ {
		d.Slides = append(d.Slides, Slide{
			ID: "slide_" + string(rune('1'+i)), Index: i + 1, Background: "#111111",
			Elements: []Element{
				{ID: "bg", Type: "image", X: 0, Y: 0, Width: 1080, Height: 1350, Opacity: 1},
				{ID: "t", Type: "text", Role: "headline", Content: "Tiêu đề", X: 80, Y: 400, Width: 920, Opacity: 1,
					Style: &Style{FontFamily: "Inter", FontSize: 64, FontWeight: 700, Color: "#FFFFFF", TextAlign: "left", LineHeight: 1.2}},
			},
		})
	}
	return d
}

func countRole(d *Design, role string) int {
	n := 0
	for _, s := range d.Slides {
		for _, e := range s.Elements {
			if e.Role == role {
				n++
			}
		}
	}
	return n
}

func TestApplyBrandHonoursScope(t *testing.T) {
	d := brandDesign(5)
	ApplyBrand(d, []BrandMark{
		{Kind: "handle", Text: "@babycare", Scope: "all", Position: "bottom_center"},
		{Kind: "logo", AssetID: "asset-1", Scope: "first_last", Position: "top_left"},
		{Kind: "website", Text: "babycare.vn", Scope: "last", Position: "bottom_center"},
	})
	if got := countRole(d, RoleBrandHandle); got != 5 {
		t.Errorf("handle should be on every slide, got %d", got)
	}
	if got := countRole(d, RoleBrandLogo); got != 2 {
		t.Errorf("logo should be on the first and last slide, got %d", got)
	}
	if got := countRole(d, RoleBrandWebsite); got != 1 {
		t.Errorf("website should be on the last slide only, got %d", got)
	}
}

func TestApplyBrandIsIdempotent(t *testing.T) {
	marks := []BrandMark{{Kind: "handle", Text: "@x", Scope: "all", Position: "bottom_center"}}
	d := brandDesign(3)
	ApplyBrand(d, marks)
	first := len(d.Slides[0].Elements)
	ApplyBrand(d, marks)
	if len(d.Slides[0].Elements) != first {
		t.Errorf("re-applying a brand kit must replace marks, not stack them: %d then %d",
			first, len(d.Slides[0].Elements))
	}
}

func TestBrandMarksSharingACornerDoNotOverlap(t *testing.T) {
	d := brandDesign(2)
	ApplyBrand(d, []BrandMark{
		{Kind: "handle", Text: "@x", Scope: "all", Position: "bottom_center"},
		{Kind: "website", Text: "x.vn", Scope: "all", Position: "bottom_center"},
	})
	var ys []float64
	for _, e := range d.Slides[0].Elements {
		if IsBrandElement(e) {
			ys = append(ys, e.Y)
		}
	}
	if len(ys) != 2 {
		t.Fatalf("expected two marks, got %d", len(ys))
	}
	if ys[0] == ys[1] {
		t.Errorf("two marks in the same corner were placed at the same y (%v)", ys)
	}
}

func TestValidateLeavesBrandMarksInTheMargin(t *testing.T) {
	d := brandDesign(2)
	ApplyBrand(d, []BrandMark{{Kind: "handle", Text: "@x", Scope: "all", Position: "bottom_center"}})
	Validate(d, Palette{})

	safeBottom := float64(d.Canvas.SafeArea.Y + d.Canvas.SafeArea.Height)
	for _, e := range d.Slides[0].Elements {
		if e.Role != RoleBrandHandle {
			continue
		}
		if e.Y < safeBottom {
			t.Errorf("validation pulled the brand mark into the safe area (y=%.0f, safe bottom=%.0f)", e.Y, safeBottom)
		}
		return
	}
	t.Error("the brand mark did not survive validation")
}

func TestBrandMarksNeverEnterTheSafeArea(t *testing.T) {
	marks := []BrandMark{
		{Kind: "handle", Text: "@babycare", Scope: "all", Position: "bottom_center"},
		{Kind: "website", Text: "babycare.vn", Scope: "last", Position: "bottom_center"},
		{Kind: "logo", AssetID: "a1", Scope: "first_last", Position: "top_left"},
	}
	d := brandDesign(4)
	top, bottom := Reserve(marks)
	d.Canvas.ReservedTop, d.Canvas.ReservedBottom = top, bottom
	d.Canvas.RecomputeSafeArea()

	ApplyBrand(d, marks)
	Validate(d, Palette{})

	safe := d.Canvas.SafeArea
	safeTop, safeBottom := float64(safe.Y), float64(safe.Y+safe.Height)
	seen := 0
	for _, s := range d.Slides {
		for _, e := range s.Elements {
			if !IsBrandElement(e) {
				continue
			}
			seen++
			// Every mark must sit entirely above or entirely below the safe area.
			if e.Y < safeBottom && e.Y+e.Height > safeTop {
				t.Errorf("%s on %s overlaps the safe area (y=%.1f..%.1f, safe=%.1f..%.1f)",
					e.Role, s.ID, e.Y, e.Y+e.Height, safeTop, safeBottom)
			}
		}
	}
	if seen == 0 {
		t.Fatal("no brand marks were applied")
	}
}

func TestReserveIsZeroWithoutBrandMarks(t *testing.T) {
	top, bottom := Reserve(nil)
	if top != 0 || bottom != 0 {
		t.Errorf("an empty brand kit must not shrink the canvas, got top=%d bottom=%d", top, bottom)
	}
}
