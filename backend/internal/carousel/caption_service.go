package carousel

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/tks/backend/internal/ai"
	"github.com/tks/backend/internal/httpx"
	"github.com/tks/backend/internal/project"
	"github.com/tks/backend/internal/research"
	"github.com/tks/backend/internal/schema"
)

// CaptionService drafts and stores the TikTok post copy. It is shared by the
// generation worker (which drafts one automatically) and the editor (which can
// regenerate on demand), so both paths produce captions the same way.
type CaptionService struct {
	repo     *Repo
	projects *project.Repo
	ai       *ai.Client
	search   research.SearchProvider
}

func NewCaptionService(repo *Repo, projects *project.Repo, aiClient *ai.Client, search research.SearchProvider) *CaptionService {
	return &CaptionService{repo: repo, projects: projects, ai: aiClient, search: search}
}

// Generate drafts a caption for a carousel. force overwrites a caption the
// creator has already edited; the worker never forces.
func (s *CaptionService) Generate(ctx context.Context, meta ai.CallMeta, carouselID string, force bool) (Caption, error) {
	c, err := s.repo.GetForWorker(ctx, carouselID)
	if err != nil {
		return Caption{}, err
	}
	body, err := s.repo.LatestContent(ctx, carouselID)
	if err != nil {
		return Caption{}, httpx.BadRequest("Carousel này chưa có nội dung để viết caption.")
	}
	proj, err := s.projects.Get(ctx, meta.UserID, c.ProjectID)
	if err != nil {
		return Caption{}, err
	}
	pctx, _, err := s.projects.LoadContext(ctx, c.ProjectID)
	if err != nil {
		return Caption{}, httpx.Internal(err)
	}

	// Hashtag reach is a moving target and we have no TikTok trends API. When a
	// search provider is configured we ground the broad picks in what it finds;
	// otherwise the model's suggestions stand on their own, and the UI labels
	// them as suggestions rather than live trending data.
	webResults := ""
	if hits, err := s.search.Search(ctx, proj.Niche+" tiktok hashtags", 4); err == nil && len(hits) > 0 {
		webResults = research.Format(hits)
	}

	ctaID := "save"
	if len(c.Input) > 0 {
		var inputs map[string]any
		if err := jsonUnmarshal(c.Input, &inputs); err == nil {
			ctaID = schema.CTA(inputs)
		}
	}

	out, err := s.ai.GenerateCaption(ctx, meta, ai.CaptionArgs{
		ProjectContext: project.ContextPrompt(proj, pctx),
		ContentJSON:    ai.MustJSON(body),
		CTA:            pctx.CTALabel(ctaID),
		Language:       proj.Language,
		WebResults:     webResults,
	})
	if err != nil {
		return Caption{}, httpx.Wrap(502, "caption_failed", "Chưa viết được caption. Vui lòng thử lại.", err)
	}
	if err := s.repo.SaveGeneratedCaption(ctx, carouselID, out.Caption, out.Hashtags, force); err != nil {
		return Caption{}, httpx.Internal(err)
	}
	return s.repo.LoadCaption(ctx, carouselID)
}

// GenerateQuietly is the worker's entry point: a caption is a nice-to-have, so a
// failure here is logged and the carousel still ships.
func (s *CaptionService) GenerateQuietly(ctx context.Context, meta ai.CallMeta, carouselID string) {
	if _, err := s.Generate(ctx, meta, carouselID, false); err != nil {
		slog.WarnContext(ctx, "caption generation failed", "carousel_id", carouselID, "error", err)
	}
}

func jsonUnmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
