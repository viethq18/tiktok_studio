package design

import (
	"math"
	"strings"
)

// BrandMark is one piece of channel identity to stamp onto slides. The design
// package takes a flat description rather than importing the project package,
// so the layout rules stay testable in isolation.
type BrandMark struct {
	Kind     string // handle | website | logo
	Text     string
	AssetID  string
	Scope    string // off | first | last | first_last | all
	Position string
	Color    string
	Font     string
}

// Element roles reserved for brand marks. The design generator never emits
// these, so re-applying a brand kit can safely replace them wholesale.
const (
	RoleBrandHandle  = "brand_handle"
	RoleBrandWebsite = "brand_website"
	RoleBrandLogo    = "brand_logo"
)

func IsBrandElement(e Element) bool {
	switch e.Role {
	case RoleBrandHandle, RoleBrandWebsite, RoleBrandLogo:
		return true
	}
	return false
}

const (
	brandMargin     = 36 // sits in the outer margin, outside the AI's safe area
	brandFontSize   = 26
	brandLineHeight = 1.3
	brandLogoBox    = 96
	brandLineGap    = 8
)

// brandTextHeight is used by BOTH the reservation and the placement. Computing
// it twice with different rounding is how a mark ends up a pixel or two inside
// the safe area it was supposed to stay out of.
const brandTextHeight = float64(brandFontSize) * brandLineHeight

func markHeight(m BrandMark) float64 {
	if m.Kind == "logo" {
		return brandLogoBox
	}
	return brandTextHeight
}

// Reserve reports how much vertical space the brand marks need beyond the
// standard margin, at the top and at the bottom. The worker applies this to the
// canvas BEFORE the design is generated, so the AI lays copy out inside a safe
// area that already excludes the brand band.
func Reserve(marks []BrandMark) (top, bottom int) {
	stack := map[string]float64{}
	for _, m := range marks {
		if m.empty() || m.Scope == "off" {
			continue
		}
		stack[m.Position] += markHeight(m) + brandLineGap
	}
	for position, height := range stack {
		band := brandMargin + height - brandLineGap
		reserved := int(math.Ceil(band - SafeMargin))
		if reserved < 0 {
			reserved = 0
		}
		if strings.HasPrefix(position, "top_") {
			top = max(top, reserved)
		} else {
			bottom = max(bottom, reserved)
		}
	}
	return top, bottom
}

// ApplyBrand stamps the brand marks onto a design, replacing any marks from a
// previous run. Marks live in the outer margin band so they can never collide
// with AI-placed copy, and several marks sharing a corner are stacked.
func ApplyBrand(d *Design, marks []BrandMark) {
	d.Canvas.ReservedTop, d.Canvas.ReservedBottom = Reserve(marks)
	d.Canvas.RecomputeSafeArea()

	// Clear previous marks first so re-applying is idempotent.
	for si := range d.Slides {
		kept := d.Slides[si].Elements[:0]
		for _, e := range d.Slides[si].Elements {
			if !IsBrandElement(e) {
				kept = append(kept, e)
			}
		}
		d.Slides[si].Elements = kept
	}

	last := len(d.Slides) - 1
	for si := range d.Slides {
		slide := &d.Slides[si]
		// Group this slide's marks by corner so overlapping ones can stack.
		byPosition := map[string][]BrandMark{}
		for _, m := range marks {
			if !appliesTo(m.Scope, si, last) || m.empty() {
				continue
			}
			byPosition[m.Position] = append(byPosition[m.Position], m)
		}
		for position, group := range byPosition {
			offset := 0.0
			for _, m := range group {
				el, height := brandElement(d, slide, m, position, offset)
				slide.Elements = append(slide.Elements, el)
				offset += height + brandLineGap
			}
		}
		for i := range slide.Elements {
			slide.Elements[i].Z = i
		}
	}
}

func (m BrandMark) empty() bool {
	if m.Kind == "logo" {
		return m.AssetID == ""
	}
	return strings.TrimSpace(m.Text) == ""
}

func appliesTo(scope string, index, last int) bool {
	switch scope {
	case "all":
		return true
	case "first":
		return index == 0
	case "last":
		return index == last
	case "first_last":
		return index == 0 || index == last
	default:
		return false
	}
}

// brandElement places one mark and reports the vertical space it consumed.
func brandElement(d *Design, slide *Slide, m BrandMark, position string, offset float64) (Element, float64) {
	canvasW := float64(d.Canvas.Width)
	canvasH := float64(d.Canvas.Height)
	top := strings.HasPrefix(position, "top_")

	if m.Kind == "logo" {
		size := float64(brandLogoBox)
		x := float64(brandMargin)
		switch {
		case strings.HasSuffix(position, "_center"):
			x = (canvasW - size) / 2
		case strings.HasSuffix(position, "_right"):
			x = canvasW - size - brandMargin
		}
		y := canvasH - size - brandMargin - offset
		if top {
			y = brandMargin + offset
		}
		return Element{
			ID: slide.ID + "_brand_logo", Type: "image", Role: RoleBrandLogo,
			X: x, Y: y, Width: size, Height: size,
			Opacity: 1, Fit: "contain", AssetID: m.AssetID,
		}, size
	}

	width := canvasW - 2*brandMargin
	height := brandTextHeight
	align := "left"
	switch {
	case strings.HasSuffix(position, "_center"):
		align = "center"
	case strings.HasSuffix(position, "_right"):
		align = "right"
	}
	y := canvasH - height - brandMargin - offset
	if top {
		y = brandMargin + offset
	}

	role := RoleBrandHandle
	id := slide.ID + "_brand_handle"
	if m.Kind == "website" {
		role, id = RoleBrandWebsite, slide.ID+"_brand_website"
	}
	color := m.Color
	if color == "" {
		color = "#FFFFFF"
	}
	font := m.Font
	if font == "" {
		font = "Inter"
	}
	return Element{
		ID: id, Type: "text", Role: role, Content: m.Text,
		X: brandMargin, Y: y, Width: width, Height: height, Opacity: 0.85,
		Style: &Style{
			FontFamily: font, FontSize: brandFontSize, FontWeight: 600,
			Color: color, TextAlign: align, LineHeight: brandLineHeight,
		},
	}, height
}

// CountBrandElements reports how many brand marks a design carries. Used by the
// worker's log line so an operator can see at a glance whether a project's
// brand kit actually reached the slides.
func CountBrandElements(d *Design) int {
	n := 0
	for _, s := range d.Slides {
		for _, e := range s.Elements {
			if IsBrandElement(e) {
				n++
			}
		}
	}
	return n
}
