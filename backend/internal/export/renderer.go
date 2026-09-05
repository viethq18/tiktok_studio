// Package export renders Design JSON to PNG without a browser (§48).
//
// Rendering engine decision (§155): Option B — a native Go renderer. It keeps
// the worker a single Go process with no Node/Chromium sidecar. The WYSIWYG
// risk that comes with Option B is answered by sharing ONE text-layout
// implementation (internal/fontkit) between design validation, this renderer
// and the client's overflow check, plus the golden-slide regression test in
// renderer_test.go.
package export

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"

	"github.com/tks/backend/internal/design"
	"github.com/tks/backend/internal/fontkit"
)

// AssetLoader resolves an asset id to its stored bytes.
type AssetLoader interface {
	Raw(ctx context.Context, assetID string) ([]byte, error)
}

type Renderer struct{ assets AssetLoader }

func NewRenderer(assets AssetLoader) *Renderer { return &Renderer{assets: assets} }

// RenderSlide draws one slide at full canvas resolution.
func (r *Renderer) RenderSlide(ctx context.Context, d design.Design, s design.Slide) (*gg.Context, error) {
	dc := gg.NewContext(d.Canvas.Width, d.Canvas.Height)

	bg := parseHex(s.Background, color.RGBA{255, 255, 255, 255})
	dc.SetColor(bg)
	dc.Clear()

	for _, el := range s.Elements {
		switch el.Type {
		case "image":
			r.drawImage(ctx, dc, d, el)
		case "shape":
			drawShape(dc, el)
		case "text":
			if err := drawText(dc, el); err != nil {
				return nil, err
			}
		}
	}
	return dc, nil
}

func (r *Renderer) drawImage(ctx context.Context, dc *gg.Context, d design.Design, el design.Element) {
	w, h := int(el.Width), int(el.Height)
	if w <= 0 || h <= 0 {
		return
	}
	// A brand logo with no asset is simply absent; it must not paint a block.
	if el.Role == design.RoleBrandLogo && el.AssetID == "" {
		return
	}
	if el.AssetID != "" && r.assets != nil {
		if raw, err := r.assets.Raw(ctx, el.AssetID); err == nil {
			if src, _, err := image.Decode(bytes.NewReader(raw)); err == nil {
				if el.Fit == "contain" {
					// Fit inside the box and centre it — a logo must not be cropped.
					box, offX, offY := containRect(src.Bounds(), w, h)
					dst := image.NewRGBA(image.Rect(0, 0, box.Dx(), box.Dy()))
					draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
					dc.DrawImage(dst, int(el.X)+offX, int(el.Y)+offY)
				} else {
					dst := image.NewRGBA(image.Rect(0, 0, w, h))
					draw.CatmullRom.Scale(dst, dst.Bounds(), src, coverRect(src.Bounds(), w, h), draw.Over, nil)
					dc.DrawImage(dst, int(el.X), int(el.Y))
				}
			}
		}
	} else {
		// No asset chosen yet: paint the palette background so the slide still
		// exports as a usable image rather than a hole.
		dc.SetColor(parseHex(d.Palette.Primary, color.RGBA{17, 17, 17, 255}))
		dc.DrawRectangle(el.X, el.Y, el.Width, el.Height)
		dc.Fill()
	}
	if el.Overlay != nil && el.Overlay.Opacity > 0 {
		c := parseHex(el.Overlay.Color, color.RGBA{0, 0, 0, 255})
		dc.SetRGBA(float64(c.R)/255, float64(c.G)/255, float64(c.B)/255, el.Overlay.Opacity)
		dc.DrawRectangle(el.X, el.Y, el.Width, el.Height)
		dc.Fill()
	}
}

// coverRect picks the source crop that fills w×h without distortion ("cover").
func coverRect(src image.Rectangle, w, h int) image.Rectangle {
	sw, sh := src.Dx(), src.Dy()
	if sw == 0 || sh == 0 {
		return src
	}
	targetRatio := float64(w) / float64(h)
	srcRatio := float64(sw) / float64(sh)
	if srcRatio > targetRatio {
		newW := int(float64(sh) * targetRatio)
		off := (sw - newW) / 2
		return image.Rect(src.Min.X+off, src.Min.Y, src.Min.X+off+newW, src.Max.Y)
	}
	newH := int(float64(sw) / targetRatio)
	off := (sh - newH) / 2
	return image.Rect(src.Min.X, src.Min.Y+off, src.Max.X, src.Min.Y+off+newH)
}

// containRect scales the source to fit entirely inside w×h and returns the
// drawn size plus the offset that centres it.
func containRect(src image.Rectangle, w, h int) (image.Rectangle, int, int) {
	sw, sh := src.Dx(), src.Dy()
	if sw == 0 || sh == 0 {
		return image.Rect(0, 0, w, h), 0, 0
	}
	scale := float64(w) / float64(sw)
	if s := float64(h) / float64(sh); s < scale {
		scale = s
	}
	dw, dh := int(float64(sw)*scale), int(float64(sh)*scale)
	return image.Rect(0, 0, dw, dh), (w - dw) / 2, (h - dh) / 2
}

func drawShape(dc *gg.Context, el design.Element) {
	c := parseHex(el.Fill, color.RGBA{0, 0, 0, 255})
	dc.SetRGBA(float64(c.R)/255, float64(c.G)/255, float64(c.B)/255, el.Opacity)
	if el.Radius > 0 {
		dc.DrawRoundedRectangle(el.X, el.Y, el.Width, el.Height, el.Radius)
	} else {
		dc.DrawRectangle(el.X, el.Y, el.Width, el.Height)
	}
	dc.Fill()
}

func drawText(dc *gg.Context, el design.Element) error {
	st := el.Style
	if st == nil {
		return nil
	}
	face, err := fontkit.Face(st.FontFamily, st.FontWeight, st.FontSize)
	if err != nil {
		return fmt.Errorf("load font %s: %w", st.FontFamily, err)
	}
	dc.SetFontFace(face)

	c := parseHex(st.Color, color.RGBA{0, 0, 0, 255})
	dc.SetRGBA(float64(c.R)/255, float64(c.G)/255, float64(c.B)/255, el.Opacity)

	// Same wrapping the validator and the client use — this is what keeps the
	// editor and the export in agreement.
	lines := fontkit.Wrap(el.Content, st.FontFamily, st.FontWeight, st.FontSize, el.Width)
	lineHeight := st.LineHeight
	if lineHeight <= 0 {
		lineHeight = 1.3
	}
	step := st.FontSize * lineHeight

	// gg anchors by baseline; the first baseline sits one ascent below the top,
	// approximated the same way the browser lays out a line box.
	y := el.Y + (step+st.FontSize*0.72)/2
	for _, line := range lines {
		x := el.X
		switch st.TextAlign {
		case "center":
			x = el.X + (el.Width-fontkit.MeasureWidth(line, st.FontFamily, st.FontWeight, st.FontSize))/2
		case "right":
			x = el.X + el.Width - fontkit.MeasureWidth(line, st.FontFamily, st.FontWeight, st.FontSize)
		}
		dc.DrawString(line, x, y)
		y += step
	}
	return nil
}

func parseHex(s string, fallback color.RGBA) color.RGBA {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 && len(s) != 8 {
		return fallback
	}
	v, err := strconv.ParseUint(s[:6], 16, 32)
	if err != nil {
		return fallback
	}
	out := color.RGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255}
	if len(s) == 8 {
		if a, err := strconv.ParseUint(s[6:8], 16, 8); err == nil {
			out.A = uint8(a)
		}
	}
	return out
}
