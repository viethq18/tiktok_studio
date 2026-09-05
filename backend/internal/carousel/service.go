package carousel

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/tks/backend/internal/ai"

	"github.com/tks/backend/internal/asset"
	"github.com/tks/backend/internal/design"
	"github.com/tks/backend/internal/httpx"
	"github.com/tks/backend/internal/job"
	"github.com/tks/backend/internal/project"
	"github.com/tks/backend/internal/ratelimit"
	"github.com/tks/backend/internal/schema"
)

const thumbnailURLTTL = 6 * time.Hour

type Service struct {
	repo     *Repo
	projects *project.Service
	jobs     *job.Repo
	queue    *job.Queue
	assets   *asset.Service
	limiter  *ratelimit.Limiter
	captions *CaptionService
}

func NewService(repo *Repo, projects *project.Service, jobs *job.Repo, queue *job.Queue,
	assets *asset.Service, limiter *ratelimit.Limiter, captions *CaptionService) *Service {
	return &Service{repo: repo, projects: projects, jobs: jobs, queue: queue,
		assets: assets, limiter: limiter, captions: captions}
}

type GenerateResult struct {
	Carousel Carousel `json:"carousel"`
	JobID    string   `json:"job_id"`
	Status   string   `json:"status"`
}

// Generate validates the submitted form, creates the carousel plus its job in
// one transaction and enqueues it — the HTTP request never waits for AI (§13, §83).
func (s *Service) Generate(ctx context.Context, userID, projectID string, inputs map[string]any, ratio, idempotencyKey string) (GenerateResult, error) {
	// A repeated submit returns the original job instead of a second generation (§159).
	if idempotencyKey != "" {
		if existing, found, err := s.jobs.FindByIdempotencyKey(ctx, userID, idempotencyKey); err == nil && found {
			c, err := s.repo.Get(ctx, userID, existing.CarouselID)
			if err != nil {
				return GenerateResult{}, err
			}
			return GenerateResult{Carousel: c, JobID: existing.ID, Status: string(existing.Status)}, nil
		}
	}
	if err := s.limiter.Take(ctx, userID, ratelimit.ActionGeneration); err != nil {
		return GenerateResult{}, err
	}

	p, err := s.projects.Repo().Get(ctx, userID, projectID)
	if err != nil {
		return GenerateResult{}, err
	}
	form, schemaVersion, err := s.projects.EnsureSchema(ctx, userID, projectID, false)
	if err != nil {
		return GenerateResult{}, err
	}
	// Backend re-validates everything the browser sent (§96).
	clean, err := form.ValidateInput(inputs)
	if err != nil {
		return GenerateResult{}, httpx.BadRequest(err.Error())
	}

	title := strings.TrimSpace(schema.Topic(clean))
	if r := []rune(title); len(r) > 80 {
		title = strings.TrimSpace(string(r[:80]))
	}
	if title == "" {
		title = "Carousel mới"
	}

	c, err := s.repo.Create(ctx, CreateParams{
		ProjectID: projectID, Title: title, Ratio: ratio,
		SchemaVersion: schemaVersion, Input: clean,
	})
	if err != nil {
		return GenerateResult{}, httpx.Internal(err)
	}
	j, err := s.jobs.Create(ctx, userID, projectID, c.ID, job.TypeCarouselGeneration, idempotencyKey)
	if err != nil {
		return GenerateResult{}, httpx.Internal(err)
	}
	// Enqueue only after the transaction is durable (§83).
	if err := s.queue.Enqueue(ctx, job.QueueGeneration, j.ID); err != nil {
		_ = s.jobs.Fail(ctx, j.ID, "queue_unavailable", err.Error())
		_ = s.repo.SetStatus(ctx, c.ID, StatusFailed)
		return GenerateResult{}, httpx.Wrap(503, "queue_unavailable", "Hệ thống đang bận. Vui lòng thử lại.", err)
	}
	s.projects.Repo().TrackEvent(ctx, userID, projectID, c.ID, "carousel_generation_started",
		map[string]any{"ratio": c.CanvasRatio, "language": p.Language})

	return GenerateResult{Carousel: c, JobID: j.ID, Status: string(job.StatusQueued)}, nil
}

// Regenerate reuses the stored inputs to run the pipeline again (§45).
func (s *Service) Regenerate(ctx context.Context, userID, carouselID string) (GenerateResult, error) {
	c, err := s.repo.Get(ctx, userID, carouselID)
	if err != nil {
		return GenerateResult{}, err
	}
	if err := s.limiter.Take(ctx, userID, ratelimit.ActionGeneration); err != nil {
		return GenerateResult{}, err
	}
	if err := s.repo.SetStatus(ctx, c.ID, StatusGenerating); err != nil {
		return GenerateResult{}, httpx.Internal(err)
	}
	j, err := s.jobs.Create(ctx, userID, c.ProjectID, c.ID, job.TypeCarouselGeneration, "")
	if err != nil {
		return GenerateResult{}, httpx.Internal(err)
	}
	if err := s.queue.Enqueue(ctx, job.QueueGeneration, j.ID); err != nil {
		return GenerateResult{}, httpx.Wrap(503, "queue_unavailable", "Hệ thống đang bận. Vui lòng thử lại.", err)
	}
	c.Status = StatusGenerating
	return GenerateResult{Carousel: c, JobID: j.ID, Status: string(job.StatusQueued)}, nil
}

// Duplicate copies the carousel, its input and its current design (§42).
func (s *Service) Duplicate(ctx context.Context, userID, carouselID string) (Carousel, error) {
	src, err := s.repo.Get(ctx, userID, carouselID)
	if err != nil {
		return Carousel{}, err
	}
	var input map[string]any
	_ = json.Unmarshal(src.Input, &input)

	copyC, err := s.repo.Create(ctx, CreateParams{
		ProjectID: src.ProjectID, Title: src.Title + " (bản sao)", Ratio: src.CanvasRatio,
		SchemaVersion: src.SchemaVersion, Input: input,
	})
	if err != nil {
		return Carousel{}, httpx.Internal(err)
	}
	status := StatusDraft
	if rec, err := s.repo.LatestDesign(ctx, src.ID); err == nil {
		if _, err := s.repo.SaveDesign(ctx, copyC.ID, rec.Design); err != nil {
			return Carousel{}, httpx.Internal(err)
		}
		status = StatusReady
	}
	if c, err := s.repo.LatestContent(ctx, src.ID); err == nil {
		_, _ = s.repo.SaveContent(ctx, copyC.ID, c)
	}
	if err := s.repo.SetStatus(ctx, copyC.ID, status); err != nil {
		return Carousel{}, httpx.Internal(err)
	}
	_ = s.repo.SetFormula(ctx, copyC.ID, src.FormulaID, "")
	return s.repo.Get(ctx, userID, copyC.ID)
}

// DesignView is the design plus the presigned URLs its images need. Storage
// keys never leave the server (§85).
type DesignView struct {
	Version int           `json:"version"`
	Design  design.Design `json:"design"`
}

func (s *Service) LoadDesign(ctx context.Context, userID, carouselID string) (DesignView, error) {
	if _, err := s.repo.Get(ctx, userID, carouselID); err != nil {
		return DesignView{}, err
	}
	rec, err := s.repo.LatestDesign(ctx, carouselID)
	if err != nil {
		return DesignView{}, err
	}
	if err := s.hydrateAssetURLs(ctx, &rec.Design); err != nil {
		return DesignView{}, httpx.Internal(err)
	}
	return DesignView{Version: rec.Version, Design: rec.Design}, nil
}

func (s *Service) hydrateAssetURLs(ctx context.Context, d *design.Design) error {
	var ids []string
	for _, slide := range d.Slides {
		for _, el := range slide.Elements {
			if el.Type == "image" && el.AssetID != "" {
				ids = append(ids, el.AssetID)
			}
		}
	}
	urls, err := s.assets.SignedURLs(ctx, ids)
	if err != nil {
		return err
	}
	for si := range d.Slides {
		for ei := range d.Slides[si].Elements {
			el := &d.Slides[si].Elements[ei]
			if el.Type == "image" {
				el.URL = urls[el.AssetID]
			}
		}
	}
	return nil
}

// SaveDesign is the editor autosave path (§32, §79, §80).
func (s *Service) SaveDesign(ctx context.Context, userID, carouselID string, baseVersion int, d design.Design) (DesignView, error) {
	c, err := s.repo.Get(ctx, userID, carouselID)
	if err != nil {
		return DesignView{}, err
	}
	// Drop presigned URLs before persisting: they expire, so they are a view
	// concern, never part of the stored document.
	for si := range d.Slides {
		for ei := range d.Slides[si].Elements {
			d.Slides[si].Elements[ei].URL = ""
		}
	}
	// Canvas geometry is owned by the server, but the brand reservation the
	// design was generated with must survive an autosave.
	preset := design.PresetFor(c.CanvasRatio)
	d.Canvas = design.Canvas{
		Width: preset.Width, Height: preset.Height, Ratio: preset.Ratio,
		ReservedTop: d.Canvas.ReservedTop, ReservedBottom: d.Canvas.ReservedBottom,
	}
	d.Canvas.RecomputeSafeArea()
	if res := design.Validate(&d, d.Palette); res.Fatal() {
		return DesignView{}, httpx.BadRequest("Thiết kế không hợp lệ: " + res.Error())
	}
	version, err := s.repo.UpdateDesign(ctx, carouselID, baseVersion, d)
	if err != nil {
		if err == httpx.ErrConflict {
			// Hand back the winning version so the client can reload (§80).
			current, loadErr := s.LoadDesign(ctx, userID, carouselID)
			if loadErr == nil {
				return current, httpx.ErrConflict
			}
		}
		return DesignView{}, err
	}
	if err := s.hydrateAssetURLs(ctx, &d); err != nil {
		return DesignView{}, httpx.Internal(err)
	}
	return DesignView{Version: version, Design: d}, nil
}

// hydrateThumbnails fills ThumbnailURL for a page of carousels.
func (s *Service) hydrateThumbnails(ctx context.Context, items []Carousel) error {
	for i := range items {
		if items[i].thumbnailKey == "" {
			continue
		}
		url, err := s.assets.Storage().SignedURL(ctx, items[i].thumbnailKey, thumbnailURLTTL, "")
		if err != nil {
			return err
		}
		items[i].ThumbnailURL = url
	}
	return nil
}

// ---- caption (§ post copy that ships with the carousel) ----

// Caption returns the stored caption, drafting one on first read so opening the
// editor is enough to get post copy.
func (s *Service) Caption(ctx context.Context, userID, carouselID string) (Caption, error) {
	c, err := s.repo.Get(ctx, userID, carouselID)
	if err != nil {
		return Caption{}, err
	}
	caption, err := s.repo.LoadCaption(ctx, c.ID)
	if err == nil {
		return caption, nil
	}
	if err != httpx.ErrNotFound {
		return Caption{}, httpx.Internal(err)
	}
	return s.captions.Generate(ctx, ai.CallMeta{UserID: userID}, c.ID, false)
}

func (s *Service) RegenerateCaption(ctx context.Context, userID, carouselID string) (Caption, error) {
	c, err := s.repo.Get(ctx, userID, carouselID)
	if err != nil {
		return Caption{}, err
	}
	if err := s.limiter.Take(ctx, userID, ratelimit.ActionGeneration); err != nil {
		return Caption{}, err
	}
	return s.captions.Generate(ctx, ai.CallMeta{UserID: userID}, c.ID, true)
}

func (s *Service) UpdateCaption(ctx context.Context, userID, carouselID, caption string, hashtags []string) (Caption, error) {
	c, err := s.repo.Get(ctx, userID, carouselID)
	if err != nil {
		return Caption{}, err
	}
	return s.repo.SaveEditedCaption(ctx, c.ID, caption, ai.NormalizeHashtags(hashtags))
}

// ---- brand marks (§39) ----

// BrandMarks projects a project's brand kit onto the design layer's flat mark
// description. Kept here so the design package stays free of project types.
func BrandMarks(b project.Brand) []design.BrandMark {
	b.Normalize()
	font := b.FontFamily
	return []design.BrandMark{
		{Kind: "handle", Text: b.Handle, Scope: b.Display.Handle.Scope, Position: b.Display.Handle.Position, Font: font},
		{Kind: "website", Text: b.Website, Scope: b.Display.Website.Scope, Position: b.Display.Website.Position, Font: font},
		{Kind: "logo", AssetID: b.LogoAssetID, Scope: b.Display.Logo.Scope, Position: b.Display.Logo.Position},
	}
}

// ApplyBrand re-stamps the project's brand kit onto an existing carousel, so a
// brand set up after the fact can still be applied without regenerating.
func (s *Service) ApplyBrand(ctx context.Context, userID, carouselID string) (DesignView, error) {
	c, err := s.repo.Get(ctx, userID, carouselID)
	if err != nil {
		return DesignView{}, err
	}
	proj, err := s.projects.Repo().Get(ctx, userID, c.ProjectID)
	if err != nil {
		return DesignView{}, err
	}
	rec, err := s.repo.LatestDesign(ctx, c.ID)
	if err != nil {
		return DesignView{}, err
	}
	design.ApplyBrand(&rec.Design, BrandMarks(proj.Brand))
	if res := design.Validate(&rec.Design, proj.Brand.Palette()); res.Fatal() {
		return DesignView{}, httpx.BadRequest("Không áp dụng được thương hiệu: " + res.Error())
	}
	version, err := s.repo.UpdateDesign(ctx, c.ID, rec.Version, rec.Design)
	if err != nil {
		return DesignView{}, err
	}
	if err := s.hydrateAssetURLs(ctx, &rec.Design); err != nil {
		return DesignView{}, httpx.Internal(err)
	}
	return DesignView{Version: version, Design: rec.Design}, nil
}
