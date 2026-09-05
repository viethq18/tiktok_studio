package project

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tks/backend/internal/language"

	"github.com/tks/backend/internal/ai"
	"github.com/tks/backend/internal/httpx"
	"github.com/tks/backend/internal/research"
	"github.com/tks/backend/internal/schema"
)

// AssetURLResolver turns asset ids into presigned URLs. Declared here as a
// narrow interface so the project package does not depend on the asset package.
type AssetURLResolver interface {
	SignedURLs(ctx context.Context, ids []string) (map[string]string, error)
}

type Service struct {
	repo     *Repo
	aiClient *ai.Client
	search   research.SearchProvider
	assets   AssetURLResolver
}

func NewService(repo *Repo, aiClient *ai.Client, search research.SearchProvider, assets AssetURLResolver) *Service {
	return &Service{repo: repo, aiClient: aiClient, search: search, assets: assets}
}

// HydrateBrand fills in the logo's presigned URL. Storage keys never leave the
// server (§85), so the browser gets a signed link or nothing.
func (s *Service) HydrateBrand(ctx context.Context, p *Project) {
	p.Brand.Normalize()
	if p.Brand.LogoAssetID == "" || s.assets == nil {
		return
	}
	urls, err := s.assets.SignedURLs(ctx, []string{p.Brand.LogoAssetID})
	if err != nil {
		return
	}
	p.Brand.LogoURL = urls[p.Brand.LogoAssetID]
}

func (s *Service) Repo() *Repo { return s.repo }

// CreateFromNiche is the onboarding pipeline (§6.2): research the niche,
// analyse it, normalise the result into our internal Context and persist.
func (s *Service) CreateFromNiche(ctx context.Context, userID, nicheInput, languageCode string) (Project, error) {
	nicheInput = strings.TrimSpace(nicheInput)
	if len([]rune(nicheInput)) < 8 {
		return Project{}, httpx.BadRequest("Hãy mô tả kênh của bạn rõ hơn một chút (ít nhất 8 ký tự).")
	}
	if len([]rune(nicheInput)) > 800 {
		return Project{}, httpx.BadRequest("Mô tả quá dài, hãy rút ngắn dưới 800 ký tự.")
	}
	// The content language is an explicit choice, not a default: getting it
	// wrong means every carousel in the project is written in the wrong
	// language, and the creator only finds out after generation.
	if strings.TrimSpace(languageCode) == "" {
		return Project{}, httpx.BadRequest("Hãy chọn ngôn ngữ nội dung cho project.")
	}
	if !language.IsSupported(languageCode) {
		return Project{}, httpx.BadRequest("Ngôn ngữ này chưa được hỗ trợ.")
	}
	languageCode = language.Normalize(languageCode)

	meta := ai.CallMeta{UserID: userID}
	intel, err := s.aiClient.AnalyzeProject(ctx, meta, nicheInput, languageCode)
	if err != nil {
		return Project{}, httpx.Wrap(502, "ai_unavailable",
			"Chúng tôi chưa phân tích được kênh của bạn. Vui lòng thử lại.", err)
	}

	p, err := s.repo.Create(ctx, userID, intel.Name, intel.Identity.Niche, intel.Identity.Description, languageCode)
	if err != nil {
		return Project{}, httpx.Internal(err)
	}
	p.Brand.Display = DefaultBrandDisplay()

	pctx := normalizeContext(intel, languageCode)
	version, err := s.repo.SaveContext(ctx, p.ID, pctx)
	if err != nil {
		return Project{}, httpx.Internal(err)
	}
	p.ContextVersion = version
	p.Context, _ = json.Marshal(pctx)

	s.repo.TrackEvent(ctx, userID, p.ID, "", "project_created", map[string]any{"language": languageCode})

	// The input form is generated eagerly so "Create carousel" is instant, but a
	// failure here must not fail project creation.
	go func() {
		bg := context.WithoutCancel(ctx)
		if _, _, err := s.EnsureSchema(bg, userID, p.ID, false); err != nil {
			slog.Warn("eager schema generation failed", "project_id", p.ID, "error", err)
		}
	}()
	return p, nil
}

// normalizeContext converts raw AI output into the internal contract the
// frontend and worker depend on (§7: never expose raw AI output).
func normalizeContext(in ai.ProjectIntelligence, languageCode string) Context {
	c := Context{
		SchemaVersion: "1.0",
		Audience: Audience{
			Description:  in.Audience.Description,
			Demographics: in.Audience.Demographics,
			PainPoints:   in.Audience.PainPoints,
		},
		Tone:         in.Tone,
		WritingStyle: in.WritingStyle,
		Angles:       in.PreferredAngles,
		Topics:       in.SuggestedTopics,
	}
	for _, p := range in.ContentPillars {
		id := slug(p.ID, p.Name)
		if id == "" {
			continue
		}
		c.Pillars = append(c.Pillars, Pillar{ID: id, Name: p.Name, Description: p.Description})
	}
	allowedCTA := map[string]bool{"save": true, "follow": true, "share": true, "comment": true}
	seen := map[string]bool{}
	for _, o := range in.CTAOptions {
		id := strings.ToLower(strings.TrimSpace(o.ID))
		if !allowedCTA[id] || seen[id] {
			continue
		}
		seen[id] = true
		c.CTAOptions = append(c.CTAOptions, CTA{ID: id, Label: strings.TrimSpace(o.Label)})
	}
	if len(c.CTAOptions) == 0 {
		c.CTAOptions = DefaultCTAOptions(languageCode)
	}
	return c
}

func slug(id, fallback string) string {
	s := strings.ToLower(strings.TrimSpace(id))
	if s == "" {
		s = strings.ToLower(strings.TrimSpace(fallback))
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '_' || r == '-':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// EnsureSchema returns the project's dynamic form, generating it if missing.
// force regenerates even when one exists (§74).
func (s *Service) EnsureSchema(ctx context.Context, userID, projectID string, force bool) (schema.Form, int, error) {
	p, err := s.repo.Get(ctx, userID, projectID)
	if err != nil {
		return schema.Form{}, 0, err
	}
	if !force {
		if rec, ok, err := s.repo.LatestSchema(ctx, projectID); err == nil && ok {
			var form schema.Form
			if json.Unmarshal(rec.Form, &form) == nil && len(form.Properties) > 0 {
				return form, rec.Version, nil
			}
		}
	}

	pctx, ctxVersion, err := s.repo.LoadContext(ctx, projectID)
	if err != nil {
		return schema.Form{}, 0, httpx.Internal(err)
	}

	out, form, aiErr := s.aiClient.GenerateCarouselSchema(ctx,
		ai.CallMeta{UserID: userID}, ContextPrompt(p, pctx), p.Language)
	if aiErr != nil {
		// A form the creator can fill in beats no form at all (§90).
		slog.Warn("schema generation fell back to the default form", "project_id", projectID, "error", aiErr)
		ctaOpts := make([]schema.CTAOption, 0, len(pctx.CTAOptions))
		for _, o := range pctx.CTAOptions {
			ctaOpts = append(ctaOpts, schema.CTAOption{ID: o.ID, Label: o.Label})
		}
		rawSchema, rawUI := schema.FallbackSchema(ctaOpts, p.Language)
		form, err = schema.Compile(rawSchema, rawUI)
		if err != nil {
			return schema.Form{}, 0, httpx.Internal(err)
		}
		out = ai.SchemaOutput{Schema: rawSchema, UI: rawUI}
	}

	formJSON, _ := json.Marshal(form)
	version, err := s.repo.SaveSchema(ctx, projectID, formJSON, out.Schema, out.UI, ctxVersion)
	if err != nil {
		return schema.Form{}, 0, httpx.Internal(err)
	}
	return form, version, nil
}

// FormByVersion loads the exact form a carousel was created against, so an old
// carousel keeps rendering even after the schema evolves (§62).
func (s *Service) FormByVersion(ctx context.Context, projectID string, version int) (schema.Form, error) {
	rec, err := s.repo.SchemaByVersion(ctx, projectID, version)
	if err != nil {
		return schema.Form{}, err
	}
	var form schema.Form
	err = json.Unmarshal(rec.Form, &form)
	return form, err
}

// ContextPrompt renders the channel brief for any AI task that needs it.
func ContextPrompt(p Project, c Context) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Channel: %s\nNiche: %s\nLanguage: %s\n", p.Name, p.Niche, p.Language)
	if p.Description != "" {
		fmt.Fprintf(&b, "Description: %s\n", p.Description)
	}
	if c.Audience.Description != "" {
		fmt.Fprintf(&b, "Audience: %s (%s)\n", c.Audience.Description, c.Audience.Demographics)
	}
	if len(c.Audience.PainPoints) > 0 {
		fmt.Fprintf(&b, "Pain points: %s\n", strings.Join(c.Audience.PainPoints, "; "))
	}
	if len(c.Tone) > 0 {
		fmt.Fprintf(&b, "Tone: %s\n", strings.Join(c.Tone, ", "))
	}
	if c.WritingStyle != "" {
		fmt.Fprintf(&b, "Writing style: %s\n", c.WritingStyle)
	}
	if len(c.Angles) > 0 {
		fmt.Fprintf(&b, "Preferred angles: %s\n", strings.Join(c.Angles, "; "))
	}
	if len(c.Pillars) > 0 {
		names := make([]string, 0, len(c.Pillars))
		for _, p := range c.Pillars {
			names = append(names, p.Name)
		}
		fmt.Fprintf(&b, "Content pillars: %s\n", strings.Join(names, ", "))
	}
	if len(c.CTAOptions) > 0 {
		parts := make([]string, 0, len(c.CTAOptions))
		for _, o := range c.CTAOptions {
			parts = append(parts, o.ID+" = "+o.Label)
		}
		fmt.Fprintf(&b, "CTA options: %s\n", strings.Join(parts, "; "))
	}
	if p.Brand.PrimaryColor != "" {
		fmt.Fprintf(&b, "Brand colors: primary %s, secondary %s, accent %s\n",
			p.Brand.PrimaryColor, p.Brand.SecondaryColor, p.Brand.AccentColor)
	}
	return b.String()
}

// DefaultCTAOptions is the fallback when the model returns none. Only the two
// languages the product itself is translated into get hand-written labels;
// anything else falls back to English, which is honest — a machine-translated
// CTA in a language nobody here can check would be worse.
func DefaultCTAOptions(languageCode string) []CTA {
	if language.Normalize(languageCode) == "vi" {
		return []CTA{
			{"save", "Lưu bài này"}, {"follow", "Theo dõi để xem thêm"},
			{"share", "Chia sẻ cho bạn bè"}, {"comment", "Bình luận suy nghĩ của bạn"},
		}
	}
	return []CTA{
		{"save", "Save this post"}, {"follow", "Follow for more"},
		{"share", "Share with a friend"}, {"comment", "Tell me what you think"},
	}
}
