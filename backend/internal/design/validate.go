package design

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tks/backend/internal/fontkit"
)

// Character budgets handed to the AI and enforced here (§122).
const (
	MaxHeadlineChars = 90
	MaxBodyChars     = 220
	MinFontSize      = 18
	MaxFontSize      = 140
	MaxElements      = 12
	MaxSlides        = 10
)

var hexColor = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)

// Issue is one design-quality problem. Fatal issues fail validation; the rest
// are auto-corrected in place so a near-miss from the AI still ships.
type Issue struct {
	SlideID   string `json:"slide_id"`
	ElementID string `json:"element_id,omitempty"`
	Rule      string `json:"rule"`
	Detail    string `json:"detail"`
	Fatal     bool   `json:"fatal"`
}

type Result struct {
	Issues []Issue
}

func (r Result) Fatal() bool {
	for _, i := range r.Issues {
		if i.Fatal {
			return true
		}
	}
	return false
}

func (r Result) Error() string {
	parts := make([]string, 0, len(r.Issues))
	for _, i := range r.Issues {
		if i.Fatal {
			parts = append(parts, fmt.Sprintf("%s/%s: %s", i.SlideID, i.Rule, i.Detail))
		}
	}
	return strings.Join(parts, "; ")
}

// OverflowReport lists text that does not fit its box after wrapping. The
// worker feeds this back to the AI as a "rewrite shorter" instruction (§121).
func (r Result) OverflowReport() []string {
	var out []string
	for _, i := range r.Issues {
		if i.Rule == "text_overflow" {
			out = append(out, i.Detail)
		}
	}
	return out
}

// Validate checks and normalizes a Design JSON document in place (§120–§123).
// Anything the AI got mildly wrong (bad font, out-of-range size, element
// nudged outside the safe area) is clamped; only structural breakage is fatal.
func Validate(d *Design, palette Palette) Result {
	var res Result
	add := func(slide, elem, rule, detail string, fatal bool) {
		res.Issues = append(res.Issues, Issue{SlideID: slide, ElementID: elem, Rule: rule, Detail: detail, Fatal: fatal})
	}

	if d.Version == "" {
		d.Version = SchemaVersion
	}
	if d.Canvas.Width <= 0 || d.Canvas.Height <= 0 {
		add("", "", "canvas", "canvas dimensions missing", true)
		return res
	}
	d.Canvas.RecomputeSafeArea()
	d.Palette = normalizePalette(d.Palette, palette)

	if len(d.Slides) == 0 {
		add("", "", "slides", "design has no slides", true)
		return res
	}
	if len(d.Slides) > MaxSlides {
		add("", "", "slides", fmt.Sprintf("%d slides exceeds the %d cap", len(d.Slides), MaxSlides), true)
		return res
	}

	seenSlides := map[string]bool{}
	for si := range d.Slides {
		s := &d.Slides[si]
		if s.ID == "" {
			s.ID = fmt.Sprintf("slide_%d", si+1)
		}
		if seenSlides[s.ID] {
			add(s.ID, "", "duplicate_slide_id", "slide id is not unique", true)
		}
		seenSlides[s.ID] = true
		s.Index = si + 1
		if !hexColor.MatchString(s.Background) {
			s.Background = d.Palette.Background
		}
		if len(s.Elements) == 0 {
			add(s.ID, "", "empty_slide", "slide has no elements", true)
			continue
		}
		if len(s.Elements) > MaxElements {
			s.Elements = s.Elements[:MaxElements]
		}

		seenElems := map[string]bool{}
		for ei := range s.Elements {
			e := &s.Elements[ei]
			if e.ID == "" {
				e.ID = fmt.Sprintf("%s_el_%d", s.ID, ei+1)
			}
			if seenElems[e.ID] {
				e.ID = fmt.Sprintf("%s_el_%d", s.ID, ei+1)
			}
			seenElems[e.ID] = true
			if e.Opacity <= 0 || e.Opacity > 1 {
				e.Opacity = 1
			}
			e.Z = ei

			switch e.Type {
			case "text":
				validateText(d, s, e, &res, add)
			case "image":
				validateImage(d, s, e)
			case "shape":
				validateShape(d, e)
			default:
				add(s.ID, e.ID, "element_type", "unsupported element type "+e.Type, true)
			}
		}
	}
	return res
}

func validateText(d *Design, s *Slide, e *Element, res *Result, add func(string, string, string, string, bool)) {
	e.Content = strings.TrimSpace(e.Content)
	if e.Content == "" {
		add(s.ID, e.ID, "empty_text", "text element has no content", true)
		return
	}
	if e.Style == nil {
		e.Style = &Style{}
	}
	if e.Role == "" {
		e.Role = "body"
	}
	def, ok := TypographyDefaults[e.Role]
	if !ok {
		if IsBrandElement(*e) {
			// Brand marks are placed by ApplyBrand in the outer margin, on
			// purpose. They keep their own role, sizing and position.
			def = Style{FontSize: brandFontSize, FontWeight: 600, TextAlign: e.Style.TextAlign, LineHeight: brandLineHeight}
		} else {
			def = TypographyDefaults["body"]
			e.Role = "body"
		}
	}

	// Font must come from the registry — AI never picks an arbitrary face (§40).
	if !fontkit.Allowed(e.Style.FontFamily) {
		e.Style.FontFamily = fontkit.DefaultFamily
	} else {
		e.Style.FontFamily = fontkit.Normalize(e.Style.FontFamily)
	}
	if e.Style.FontSize < MinFontSize || e.Style.FontSize > MaxFontSize {
		e.Style.FontSize = def.FontSize
	}
	if e.Style.FontWeight < 100 || e.Style.FontWeight > 900 {
		e.Style.FontWeight = def.FontWeight
	}
	if e.Style.LineHeight < 1 || e.Style.LineHeight > 2.5 {
		e.Style.LineHeight = def.LineHeight
	}
	switch e.Style.TextAlign {
	case "left", "center", "right":
	default:
		e.Style.TextAlign = def.TextAlign
	}
	if !hexColor.MatchString(e.Style.Color) {
		e.Style.Color = d.Palette.Text
	}

	if IsBrandElement(*e) {
		// Deliberately outside the safe area, with a height fixed by ApplyBrand:
		// re-measuring here could grow the box past the band reserved for it.
		return
	}

	// Keep the box inside the safe area (§123).
	safe := d.Canvas.SafeArea
	minW := 120.0
	if e.Width < minW {
		e.Width = float64(safe.Width)
	}
	if e.Width > float64(safe.Width) {
		e.Width = float64(safe.Width)
	}
	if e.X < float64(safe.X) {
		e.X = float64(safe.X)
	}
	if e.X+e.Width > float64(safe.X+safe.Width) {
		e.X = float64(safe.X+safe.Width) - e.Width
	}
	if e.Y < float64(safe.Y) {
		e.Y = float64(safe.Y)
	}

	// Measure with the same engine the exporter uses, then record overflow (§121).
	_, height := fontkit.MeasureBlock(e.Content, e.Style.FontFamily, e.Style.FontWeight,
		e.Style.FontSize, e.Style.LineHeight, e.Width)
	e.Height = height

	limit := MaxBodyChars
	if e.Role == "hook" || e.Role == "headline" || e.Role == "cta" {
		limit = MaxHeadlineChars
	}
	if n := len([]rune(e.Content)); n > limit {
		add(s.ID, e.ID, "text_overflow",
			fmt.Sprintf("slide %s %s is %d characters, limit is %d", s.ID, e.Role, n, limit), false)
	}

	bottom := e.Y + e.Height
	safeBottom := float64(safe.Y + safe.Height)
	if bottom > safeBottom {
		// Try to pull the block up before declaring overflow.
		if e.Height <= float64(safe.Height) {
			e.Y = safeBottom - e.Height
			if e.Y < float64(safe.Y) {
				e.Y = float64(safe.Y)
			}
		} else {
			add(s.ID, e.ID, "text_overflow",
				fmt.Sprintf("slide %s %s does not fit: needs %.0fpx, safe area is %dpx", s.ID, e.Role, e.Height, safe.Height), false)
		}
	}
}

func validateImage(d *Design, s *Slide, e *Element) {
	// Images may bleed to the canvas edge; only clamp to the canvas itself.
	if e.Width <= 0 || e.Height <= 0 {
		e.X, e.Y = 0, 0
		e.Width, e.Height = float64(d.Canvas.Width), float64(d.Canvas.Height)
	}
	if e.Fit != "cover" && e.Fit != "contain" {
		e.Fit = "cover"
	}
	if e.Overlay != nil {
		if !hexColor.MatchString(e.Overlay.Color) {
			e.Overlay.Color = "#000000"
		}
		if e.Overlay.Opacity < 0 || e.Overlay.Opacity > 0.85 {
			e.Overlay.Opacity = 0.35
		}
	}
	if s.ImageQuery == "" && e.Content != "" {
		s.ImageQuery = e.Content
	}
}

func validateShape(d *Design, e *Element) {
	if !hexColor.MatchString(e.Fill) {
		e.Fill = d.Palette.Accent
	}
	if e.Width <= 0 {
		e.Width = 120
	}
	if e.Height <= 0 {
		e.Height = 12
	}
	if e.Radius < 0 {
		e.Radius = 0
	}
}

func normalizePalette(p, brand Palette) Palette {
	pick := func(candidates ...string) string {
		for _, c := range candidates {
			if hexColor.MatchString(c) {
				return c
			}
		}
		return "#111111"
	}
	return Palette{
		Primary:    pick(brand.Primary, p.Primary, "#111111"),
		Secondary:  pick(brand.Secondary, p.Secondary, "#6B7280"),
		Accent:     pick(brand.Accent, p.Accent, "#FF3B5C"),
		Background: pick(p.Background, brand.Background, "#FFFFFF"),
		Text:       pick(p.Text, brand.Text, "#111111"),
	}
}
