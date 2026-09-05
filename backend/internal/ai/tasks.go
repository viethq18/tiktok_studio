package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tks/backend/internal/content"
	"github.com/tks/backend/internal/design"
	"github.com/tks/backend/internal/fontkit"
	"github.com/tks/backend/internal/language"
	"github.com/tks/backend/internal/schema"
)

// AnalyzeProject turns a raw niche sentence into the channel brief (§6.2, §7).
func (c *Client) AnalyzeProject(ctx context.Context, meta CallMeta, nicheInput, languageCode string) (ProjectIntelligence, error) {
	return run(ctx, c, meta, TaskProjectAnalysis, map[string]any{
		"niche_input":   nicheInput,
		"language":      language.Name(languageCode),
		"language_code": languageCode,
	}, func(p *ProjectIntelligence) error {
		p.SchemaVersion = "1.0"
		p.Name = strings.TrimSpace(p.Name)
		if p.Name == "" {
			return Reject("field \"name\" is missing")
		}
		if p.Identity.Niche == "" {
			return Reject("field \"identity.niche\" is missing")
		}
		p.Identity.Language = languageCode
		if len(p.ContentPillars) < 3 {
			return Reject("give at least 3 content_pillars, you gave %d", len(p.ContentPillars))
		}
		if len(p.CTAOptions) < 2 {
			return Reject("give at least 2 cta_options, you gave %d", len(p.CTAOptions))
		}
		if len(p.Tone) == 0 {
			return Reject("field \"tone\" must list at least one adjective")
		}
		return nil
	})
}

// GenerateCarouselSchema asks for a dynamic form and only accepts one that
// survives the schema safety compiler (§11).
func (c *Client) GenerateCarouselSchema(ctx context.Context, meta CallMeta, projectContext, languageCode string) (SchemaOutput, schema.Form, error) {
	var form schema.Form
	out, err := run(ctx, c, meta, TaskCarouselSchema, map[string]any{
		"project_context": projectContext,
		"language":        language.Name(languageCode),
		"language_code":   languageCode,
	}, func(o *SchemaOutput) error {
		f, err := schema.Compile(o.Schema, o.UI)
		if err != nil {
			return Reject("the schema was rejected by validation: %v", err)
		}
		if !hasField(f, "topic") || !hasField(f, "slide_count") || !hasField(f, "cta") {
			return Reject("the schema must define topic, slide_count and cta")
		}
		form = f
		return nil
	})
	return out, form, err
}

func hasField(f schema.Form, name string) bool {
	for _, p := range f.Properties {
		if p.Name == name {
			return true
		}
	}
	return false
}

type ResearchArgs struct {
	ProjectContext string
	Query          string
	Language       string
	WebResults     string
}

func (c *Client) Research(ctx context.Context, meta CallMeta, a ResearchArgs) (ResearchOutput, error) {
	return run(ctx, c, meta, TaskResearch, map[string]any{
		"project_context": a.ProjectContext,
		"query":           a.Query,
		"language":        language.Name(a.Language),
		"web_results":     a.WebResults,
	}, func(o *ResearchOutput) error {
		if strings.TrimSpace(o.Summary) == "" {
			return Reject("field \"summary\" is empty")
		}
		if len(o.KeyFacts) < 3 {
			return Reject("give at least 3 key_facts, you gave %d", len(o.KeyFacts))
		}
		// Never let a model-invented URL reach the database.
		kept := o.Sources[:0]
		for _, s := range o.Sources {
			if strings.HasPrefix(s.URL, "http://") || strings.HasPrefix(s.URL, "https://") {
				kept = append(kept, s)
			}
		}
		o.Sources = kept
		return nil
	})
}

type ContentArgs struct {
	ProjectContext string
	Inputs         string
	Research       string
	CTA            string
	SlideCount     int
	Language       string
	Disclaimer     string
}

// GenerateContent writes the slides and enforces the semantic quality gate,
// feeding any failure back as a targeted repair instruction (§119, §121).
func (c *Client) GenerateContent(ctx context.Context, meta CallMeta, a ContentArgs) (ContentOutput, error) {
	return run(ctx, c, meta, TaskContent, map[string]any{
		"project_context": a.ProjectContext,
		"inputs":          a.Inputs,
		"research":        a.Research,
		"cta":             a.CTA,
		"slide_count":     a.SlideCount,
		"language":        language.Name(a.Language),
		"disclaimer":      a.Disclaimer,
		"max_headline":    content.MaxHeadlineChars,
		"max_body":        content.MaxBodyChars,
		"repair":          "",
	}, func(o *ContentOutput) error {
		rep := content.Validate(o, a.SlideCount)
		if !rep.OK() {
			return Reject("the copy failed review:\n%s", rep.Instruction())
		}
		return nil
	})
}

func (c *Client) SelectFormula(ctx context.Context, meta CallMeta, formulas, contentJSON string, allowed map[string]bool) (FormulaSelection, error) {
	return run(ctx, c, meta, TaskFormulaSelection, map[string]any{
		"formulas": formulas,
		"content":  contentJSON,
	}, func(o *FormulaSelection) error {
		if !allowed[o.FormulaID] {
			return Reject("formula_id %q is not in the registry; pick one of the listed ids", o.FormulaID)
		}
		return nil
	})
}

type DesignArgs struct {
	ProjectContext string
	ContentJSON    string
	Formula        string
	TemplateStyle  string
	Palette        design.Palette
	Canvas         design.Canvas
}

// GenerateDesign lays out the slides and re-runs design validation until the
// layout is clean, or gives up after the attempt cap (§55, §120).
func (c *Client) GenerateDesign(ctx context.Context, meta CallMeta, a DesignArgs) (DesignOutput, error) {
	paletteJSON, _ := json.Marshal(a.Palette)
	safe := a.Canvas.SafeArea
	return run(ctx, c, meta, TaskDesign, map[string]any{
		"project_context": a.ProjectContext,
		"content":         a.ContentJSON,
		"formula":         a.Formula,
		"template_style":  a.TemplateStyle,
		"palette":         string(paletteJSON),
		"fonts":           strings.Join(fontkit.AllowedNames(), ", "),
		"canvas_width":    a.Canvas.Width,
		"canvas_height":   a.Canvas.Height,
		"ratio":           a.Canvas.Ratio,
		"safe_x":          safe.X,
		"safe_y":          safe.Y,
		"safe_width":      safe.Width,
		"safe_height":     safe.Height,
	}, func(d *DesignOutput) error {
		d.Canvas = a.Canvas
		res := design.Validate(d, a.Palette)
		if res.Fatal() {
			return Reject("the layout was rejected: %s", res.Error())
		}
		if overflow := res.OverflowReport(); len(overflow) > 0 {
			return Reject("text does not fit its box — shorten the copy or enlarge the box:\n- %s",
				strings.Join(overflow, "\n- "))
		}
		return nil
	})
}

type CaptionArgs struct {
	ProjectContext string
	ContentJSON    string
	CTA            string
	Language       string
	WebResults     string
}

// GenerateCaption writes the post copy that ships alongside the carousel.
func (c *Client) GenerateCaption(ctx context.Context, meta CallMeta, a CaptionArgs) (CaptionOutput, error) {
	return run(ctx, c, meta, TaskCaption, map[string]any{
		"project_context": a.ProjectContext,
		"content":         a.ContentJSON,
		"cta":             a.CTA,
		"language":        language.Name(a.Language),
		"web_results":     a.WebResults,
	}, func(o *CaptionOutput) error {
		o.Caption = strings.TrimSpace(o.Caption)
		if o.Caption == "" {
			return Reject("field \"caption\" is empty")
		}
		if n := len([]rune(o.Caption)); n > 400 {
			return Reject("the caption is %d characters; keep it under 300", n)
		}
		o.Hashtags = NormalizeHashtags(o.Hashtags)
		if len(o.Hashtags) < 5 {
			return Reject("give at least 8 hashtags, you gave %d usable ones", len(o.Hashtags))
		}
		return nil
	})
}

// NormalizeHashtags lowercases, de-duplicates, strips whitespace and enforces a
// leading "#". Exported because the edit endpoint applies the same rules to
// whatever the creator types.
func NormalizeHashtags(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, raw := range in {
		for _, part := range strings.Fields(raw) {
			tag := strings.ToLower(strings.TrimSpace(part))
			tag = strings.TrimPrefix(tag, "#")
			tag = strings.Map(func(r rune) rune {
				switch {
				case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_':
					return r
				case r > 127: // keep non-ASCII letters, e.g. Vietnamese
					return r
				}
				return -1
			}, tag)
			if tag == "" || seen[tag] || len(out) >= 15 {
				continue
			}
			seen[tag] = true
			out = append(out, "#"+tag)
		}
	}
	return out
}

func (c *Client) GenerateImageIntents(ctx context.Context, meta CallMeta, slidesJSON string) (ImageIntentOutput, error) {
	return run(ctx, c, meta, TaskImageIntent, map[string]any{"slides": slidesJSON},
		func(o *ImageIntentOutput) error {
			if len(o.Slides) == 0 {
				return Reject("field \"slides\" is empty")
			}
			return nil
		})
}

// TemplateStyles are the internal design languages the planner picks from (§126).
var TemplateStyles = []string{"minimal", "editorial", "bold", "photo-heavy", "educational"}

func TemplateStyleFor(seed string) string {
	if seed == "" {
		return TemplateStyles[0]
	}
	sum := 0
	for _, r := range seed {
		sum += int(r)
	}
	return TemplateStyles[sum%len(TemplateStyles)]
}

func MustJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
