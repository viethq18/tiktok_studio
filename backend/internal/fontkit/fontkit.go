// Package fontkit is the single implementation of font lookup, text
// measurement and line wrapping used by BOTH design validation and the export
// renderer. Client-side canvas measurement mirrors the same greedy algorithm
// (frontend/lib/design/measure.ts); keeping one algorithm on both sides is what
// makes overflow detection trustworthy (§121, §157).
package fontkit

import (
	"embed"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/*.ttf
var fontFS embed.FS

// Registry entry (§40). AI may only pick a family listed here.
type Family struct {
	ID         string
	Name       string
	Slug       string
	Vietnamese bool
}

var Families = []Family{
	{"inter", "Inter", "inter", true},
	{"roboto", "Roboto", "roboto", true},
	{"montserrat", "Montserrat", "montserrat", true},
	{"nunito", "Nunito", "nunito", true},
	{"playfair", "Playfair Display", "playfair", true},
	{"be-vietnam", "Be Vietnam Pro", "be-vietnam", true},
}

const DefaultFamily = "Inter"

var byName = func() map[string]Family {
	m := map[string]Family{}
	for _, f := range Families {
		m[strings.ToLower(f.Name)] = f
		m[f.ID] = f
	}
	return m
}()

// Allowed reports whether an AI-proposed font family is in the registry.
func Allowed(name string) bool {
	_, ok := byName[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

func Normalize(name string) string {
	if f, ok := byName[strings.ToLower(strings.TrimSpace(name))]; ok {
		return f.Name
	}
	return DefaultFamily
}

func AllowedNames() []string {
	out := make([]string, 0, len(Families))
	for _, f := range Families {
		out = append(out, f.Name)
	}
	return out
}

// Only two static weights ship per family; 500/600 round down to regular and
// 800/900 round up to bold. Documented so design output stays predictable.
func weightSlug(weight int) string {
	if weight >= 700 {
		return "700"
	}
	return "400"
}

type faceKey struct {
	slug   string
	weight string
	size   float64
}

var (
	mu      sync.Mutex
	sfnts   = map[string]*sfnt.Font{}
	faces   = map[faceKey]font.Face{}
)

func loadSfnt(slug, weight string) (*sfnt.Font, error) {
	key := slug + "-" + weight
	if f, ok := sfnts[key]; ok {
		return f, nil
	}
	raw, err := fontFS.ReadFile("fonts/" + key + ".ttf")
	if err != nil {
		return nil, fmt.Errorf("font %s not embedded: %w", key, err)
	}
	f, err := sfnt.Parse(raw)
	if err != nil {
		return nil, err
	}
	sfnts[key] = f
	return f, nil
}

// Face returns a cached font.Face for the given family/weight/size.
func Face(family string, weight int, size float64) (font.Face, error) {
	fam, ok := byName[strings.ToLower(strings.TrimSpace(family))]
	if !ok {
		fam = byName[strings.ToLower(DefaultFamily)]
	}
	ws := weightSlug(weight)
	key := faceKey{fam.Slug, ws, size}

	mu.Lock()
	defer mu.Unlock()
	if f, ok := faces[key]; ok {
		return f, nil
	}
	sf, err := loadSfnt(fam.Slug, ws)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(sf, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		return nil, err
	}
	faces[key] = face
	return face, nil
}

// RawTTF exposes the embedded bytes (used by the export renderer).
func RawTTF(family string, weight int) ([]byte, error) {
	fam, ok := byName[strings.ToLower(strings.TrimSpace(family))]
	if !ok {
		fam = byName[strings.ToLower(DefaultFamily)]
	}
	return fontFS.ReadFile("fonts/" + fam.Slug + "-" + weightSlug(weight) + ".ttf")
}

// MeasureWidth returns the advance width of s in pixels.
func MeasureWidth(s, family string, weight int, size float64) float64 {
	face, err := Face(family, weight, size)
	if err != nil {
		// Fall back to a coarse estimate rather than failing validation outright.
		return float64(len([]rune(s))) * size * 0.5
	}
	return fixedToFloat(font.MeasureString(face, s))
}

func fixedToFloat(v fixed.Int26_6) float64 { return float64(v) / 64.0 }

// Wrap greedily breaks s into lines that fit maxWidth. Words longer than the
// line are split by rune. This exact algorithm is mirrored on the client.
func Wrap(s, family string, weight int, size, maxWidth float64) []string {
	var lines []string
	for _, paragraph := range strings.Split(s, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		current := ""
		for _, word := range words {
			candidate := word
			if current != "" {
				candidate = current + " " + word
			}
			if MeasureWidth(candidate, family, weight, size) <= maxWidth || current == "" {
				if MeasureWidth(candidate, family, weight, size) > maxWidth && current == "" {
					// A single word wider than the box: hard-split by rune.
					chunk := ""
					for _, r := range word {
						next := chunk + string(r)
						if MeasureWidth(next, family, weight, size) > maxWidth && chunk != "" {
							lines = append(lines, chunk)
							chunk = string(r)
						} else {
							chunk = next
						}
					}
					current = chunk
					continue
				}
				current = candidate
				continue
			}
			lines = append(lines, current)
			current = word
		}
		if current != "" {
			lines = append(lines, current)
		}
	}
	return lines
}

// MeasureBlock returns the rendered width and height of wrapped text.
func MeasureBlock(s, family string, weight int, size, lineHeight, maxWidth float64) (float64, float64) {
	if lineHeight <= 0 {
		lineHeight = 1.3
	}
	lines := Wrap(s, family, weight, size, maxWidth)
	widest := 0.0
	for _, ln := range lines {
		if w := MeasureWidth(ln, family, weight, size); w > widest {
			widest = w
		}
	}
	return widest, float64(len(lines)) * size * lineHeight
}
