// Package design owns the Design JSON contract — the source of truth for what a
// carousel looks like (§2.5, §22). Fabric.js on the client and the Go export
// renderer are two views of this same document, never the other way round.
package design

const SchemaVersion = "1.0"

type Design struct {
	Version string  `json:"version"`
	Canvas  Canvas  `json:"canvas"`
	Palette Palette `json:"palette"`
	Slides  []Slide `json:"slides"`
}

type Canvas struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Ratio  string `json:"ratio"`

	// Bands reserved for brand marks. They shrink the safe area so AI-placed
	// copy can never collide with a handle, website or logo.
	ReservedTop    int `json:"reserved_top,omitempty"`
	ReservedBottom int `json:"reserved_bottom,omitempty"`

	SafeArea SafeArea `json:"safe_area"`
}

// RecomputeSafeArea derives the safe area from the margin plus any brand
// reservation. It is the single place the safe area is calculated.
func (c *Canvas) RecomputeSafeArea() {
	c.SafeArea = SafeArea{
		X:      SafeMargin,
		Y:      SafeMargin + c.ReservedTop,
		Width:  c.Width - 2*SafeMargin,
		Height: c.Height - 2*SafeMargin - c.ReservedTop - c.ReservedBottom,
	}
}

// SafeArea is the region AI must keep text inside (§123).
type SafeArea struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Palette keeps colours to a named system instead of arbitrary hex (§125).
type Palette struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Accent     string `json:"accent"`
	Background string `json:"background"`
	Text       string `json:"text"`
}

type Slide struct {
	ID          string    `json:"id"`
	Index       int       `json:"index"`
	Role        string    `json:"role"`
	Template    string    `json:"template,omitempty"`
	Background  string    `json:"background"`
	ImageIntent string    `json:"image_intent,omitempty"`
	ImageQuery  string    `json:"image_query,omitempty"`
	Elements    []Element `json:"elements"`
}

// Element is a flat union of the three allowed element types (§23). A single
// struct keeps the JSON schema simple and round-trips cleanly through Fabric.
type Element struct {
	ID      string  `json:"id"`
	Type    string  `json:"type"` // text | image | shape
	Z       int     `json:"z"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Opacity float64 `json:"opacity"`
	Locked  bool    `json:"locked,omitempty"` // §46

	// text
	Content string `json:"content,omitempty"`
	Role    string `json:"role,omitempty"` // §124 hook|headline|subheadline|body|caption|cta
	Style   *Style `json:"style,omitempty"`

	// image
	AssetID string   `json:"asset_id,omitempty"`
	URL     string   `json:"url,omitempty"` // presigned, derived server-side, never persisted
	Fit     string   `json:"fit,omitempty"`
	Overlay *Overlay `json:"overlay,omitempty"`

	// shape
	Shape  string  `json:"shape,omitempty"`
	Fill   string  `json:"fill,omitempty"`
	Radius float64 `json:"radius,omitempty"`
}

type Style struct {
	FontFamily string  `json:"fontFamily"`
	FontSize   float64 `json:"fontSize"`
	FontWeight int     `json:"fontWeight"`
	Color      string  `json:"color"`
	TextAlign  string  `json:"textAlign"`
	LineHeight float64 `json:"lineHeight"`
}

type Overlay struct {
	Color   string  `json:"color"`
	Opacity float64 `json:"opacity"`
}

// Ratio presets (§24). Keeping them in one table means adding Instagram later
// is a data change, not a code change (§50).
type Preset struct {
	Ratio  string `json:"ratio"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Label  string `json:"label"`
}

var Presets = []Preset{
	{"4:5", 1080, 1350, "Recommended"},
	{"1:1", 1080, 1080, "Square"},
	{"9:16", 1080, 1920, "Full-screen"},
}

func PresetFor(ratio string) Preset {
	for _, p := range Presets {
		if p.Ratio == ratio {
			return p
		}
	}
	return Presets[0]
}

const SafeMargin = 80

func SafeAreaFor(w, h int) SafeArea {
	return SafeArea{X: SafeMargin, Y: SafeMargin, Width: w - 2*SafeMargin, Height: h - 2*SafeMargin}
}

// TypographyDefaults per role (§124).
var TypographyDefaults = map[string]Style{
	"hook":         {FontSize: 76, FontWeight: 800, TextAlign: "center", LineHeight: 1.15},
	"headline":     {FontSize: 64, FontWeight: 700, TextAlign: "left", LineHeight: 1.2},
	"subheadline":  {FontSize: 44, FontWeight: 600, TextAlign: "left", LineHeight: 1.25},
	"body":         {FontSize: 36, FontWeight: 400, TextAlign: "left", LineHeight: 1.4},
	"caption":      {FontSize: 24, FontWeight: 400, TextAlign: "left", LineHeight: 1.4},
	"cta":          {FontSize: 56, FontWeight: 700, TextAlign: "center", LineHeight: 1.2},
}
