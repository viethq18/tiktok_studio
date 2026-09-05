// Package project owns the "AI Content Brain" of a channel (§5): identity,
// audience, strategy, CTA options and brand kit.
package project

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/tks/backend/internal/design"
)

type Project struct {
	ID             string          `json:"id"`
	UserID         string          `json:"user_id"`
	Name           string          `json:"name"`
	Niche          string          `json:"niche"`
	Description    string          `json:"description"`
	Language       string          `json:"language"`
	Status         string          `json:"status"`
	Brand          Brand           `json:"brand"`
	Context        json.RawMessage `json:"context,omitempty"`
	ContextVersion int             `json:"context_version"`
	CarouselCount  int             `json:"carousel_count"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// Brand is the MVP brand-kit foundation (§39). AI must respect it when it
// generates a design.
type Brand struct {
	LogoAssetID    string `json:"logo_asset_id,omitempty"`
	LogoURL        string `json:"logo_url,omitempty"` // presigned, derived per request
	PrimaryColor   string `json:"primary_color,omitempty"`
	SecondaryColor string `json:"secondary_color,omitempty"`
	AccentColor    string `json:"accent_color,omitempty"`
	FontFamily     string `json:"font_family,omitempty"`

	// Optional identity shown on the carousel itself.
	Website string `json:"website,omitempty"`
	Handle  string `json:"handle,omitempty"` // @channel

	Display BrandDisplay `json:"display"`
}

// BrandDisplay says where each piece of identity appears. The defaults encode a
// recommendation rather than leaving it to the creator (§2.1):
//
//   - handle  — every slide. It is one small line, it is what converts a
//     viewer into a follower, and a viewer can enter the carousel on any slide.
//   - logo    — first and last slide only. It carries more visual weight than a
//     handle, so it belongs at the entry and the exit, not in between.
//   - website — last slide only. It is a conversion element and belongs beside
//     the call to action, where the viewer has already decided to act.
type BrandDisplay struct {
	Handle  Placement `json:"handle"`
	Logo    Placement `json:"logo"`
	Website Placement `json:"website"`
}

type Placement struct {
	Scope    string `json:"scope"`    // off | first | last | first_last | all
	Position string `json:"position"` // top_left | top_center | top_right | bottom_left | bottom_center | bottom_right
}

func DefaultBrandDisplay() BrandDisplay {
	return BrandDisplay{
		Handle:  Placement{Scope: "all", Position: "bottom_center"},
		Logo:    Placement{Scope: "first_last", Position: "top_left"},
		Website: Placement{Scope: "last", Position: "bottom_center"},
	}
}

var validScopes = map[string]bool{"off": true, "first": true, "last": true, "first_last": true, "all": true}
var validPositions = map[string]bool{
	"top_left": true, "top_center": true, "top_right": true,
	"bottom_left": true, "bottom_center": true, "bottom_right": true,
}

// Normalize fills in defaults and rejects values outside the allowed sets, so a
// hand-edited brand document can never produce an unplaceable element.
func (b *Brand) Normalize() {
	def := DefaultBrandDisplay()
	fix := func(p *Placement, fallback Placement) {
		if !validScopes[p.Scope] {
			p.Scope = fallback.Scope
		}
		if !validPositions[p.Position] {
			p.Position = fallback.Position
		}
	}
	fix(&b.Display.Handle, def.Handle)
	fix(&b.Display.Logo, def.Logo)
	fix(&b.Display.Website, def.Website)

	b.Handle = strings.TrimSpace(b.Handle)
	if b.Handle != "" && !strings.HasPrefix(b.Handle, "@") {
		b.Handle = "@" + b.Handle
	}
	b.Website = strings.TrimSpace(b.Website)
	b.Website = strings.TrimPrefix(strings.TrimPrefix(b.Website, "https://"), "http://")
	b.Website = strings.TrimSuffix(b.Website, "/")
}

// Palette projects the brand kit onto the design palette contract (§125).
func (b Brand) Palette() design.Palette {
	return design.Palette{
		Primary:   b.PrimaryColor,
		Secondary: b.SecondaryColor,
		Accent:    b.AccentColor,
	}
}

// Context is the normalized internal shape of project intelligence. The
// frontend depends on this, never on raw AI output (§7).
type Context struct {
	SchemaVersion string   `json:"schema_version"`
	Audience      Audience `json:"audience"`
	Tone          []string `json:"tone"`
	WritingStyle  string   `json:"writing_style"`
	Angles        []string `json:"preferred_angles"`
	Pillars       []Pillar `json:"content_pillars"`
	CTAOptions    []CTA    `json:"cta_options"`
	Topics        []string `json:"suggested_topics"`
}

type Audience struct {
	Description  string   `json:"description"`
	Demographics string   `json:"demographics"`
	PainPoints   []string `json:"pain_points"`
}

type Pillar struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CTA struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

func (c Context) CTALabel(id string) string {
	for _, o := range c.CTAOptions {
		if o.ID == id {
			return o.Label
		}
	}
	return id
}
