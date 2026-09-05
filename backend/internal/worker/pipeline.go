// Package worker runs the async generation and export pipelines (§12, §82).
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/tks/backend/internal/ai"
	"github.com/tks/backend/internal/carousel"
	"github.com/tks/backend/internal/content"
	"github.com/tks/backend/internal/design"
	"github.com/tks/backend/internal/export"
	"github.com/tks/backend/internal/image"
	"github.com/tks/backend/internal/job"
	"github.com/tks/backend/internal/project"
	"github.com/tks/backend/internal/registry"
	"github.com/tks/backend/internal/research"
	"github.com/tks/backend/internal/schema"
)

type Pipeline struct {
	jobs      *job.Repo
	projects  *project.Repo
	carousels *carousel.Repo
	registry  *registry.Repo
	ai        *ai.Client
	search    research.SearchProvider
	images    *image.Service
	exports   *export.Service
	captions  *carousel.CaptionService
}

func NewPipeline(jobs *job.Repo, projects *project.Repo, carousels *carousel.Repo, reg *registry.Repo,
	aiClient *ai.Client, search research.SearchProvider, images *image.Service, exports *export.Service,
	captions *carousel.CaptionService) *Pipeline {
	return &Pipeline{jobs: jobs, projects: projects, carousels: carousels, registry: reg,
		ai: aiClient, search: search, images: images, exports: exports, captions: captions}
}

// Run executes the generation pipeline for one job, resuming from the last
// completed step when this is a retry (§158).
func (p *Pipeline) Run(ctx context.Context, jobID string) error {
	j, err := p.jobs.GetForWorker(ctx, jobID)
	if err != nil {
		return err
	}
	if err := p.jobs.MarkStarted(ctx, jobID); err != nil {
		return err
	}
	cp := p.jobs.Checkpoint(j)
	meta := ai.CallMeta{JobID: j.ID, UserID: j.UserID}
	log := slog.With("job_id", j.ID, "carousel_id", j.CarouselID)

	if err := p.run(ctx, j, cp, meta, log); err != nil {
		code, message := classify(err)
		log.Error("generation failed", "error", err, "code", code)
		_ = p.jobs.Fail(ctx, j.ID, code, err.Error())
		_ = p.carousels.SetStatus(ctx, j.CarouselID, carousel.StatusFailed)
		p.projects.TrackEvent(ctx, j.UserID, j.ProjectID, j.CarouselID, "carousel_generation_failed",
			map[string]any{"code": code, "message": message})
		return err
	}

	if err := p.jobs.Complete(ctx, j.ID); err != nil {
		return err
	}
	if err := p.carousels.SetStatus(ctx, j.CarouselID, carousel.StatusReady); err != nil {
		return err
	}
	p.projects.TrackEvent(ctx, j.UserID, j.ProjectID, j.CarouselID, "carousel_generation_completed", nil)
	log.Info("generation completed")
	return nil
}

func (p *Pipeline) run(ctx context.Context, j job.Job, cp job.Checkpoint, meta ai.CallMeta, log *slog.Logger) error {
	c, err := p.carousels.GetForWorker(ctx, j.CarouselID)
	if err != nil {
		return err
	}
	proj, err := p.projects.Get(ctx, j.UserID, j.ProjectID)
	if err != nil {
		return err
	}
	pctx, _, err := p.projects.LoadContext(ctx, j.ProjectID)
	if err != nil {
		return err
	}
	brief := project.ContextPrompt(proj, pctx)

	var inputs map[string]any
	_ = json.Unmarshal(c.Input, &inputs)
	topic := schema.Topic(inputs)
	slideCount := schema.SlideCount(inputs)
	ctaID := schema.CTA(inputs)

	// ---- 1. Research (§16.2) ----
	if cp.ResearchJSON == "" {
		if err := p.jobs.Advance(ctx, j.ID, job.StatusResearching, cp); err != nil {
			return err
		}
		webResults := ""
		if hits, err := p.search.Search(ctx, topic, 5); err != nil {
			log.Warn("web search failed, continuing without live sources", "error", err)
		} else if len(hits) > 0 {
			webResults = research.Format(hits)
		}
		out, err := p.ai.Research(ctx, meta, ai.ResearchArgs{
			ProjectContext: brief, Query: topic, Language: proj.Language, WebResults: webResults,
		})
		if err != nil {
			return fmt.Errorf("research: %w", err)
		}
		if len(out.Sources) > 0 {
			sources := make([]project.Source, 0, len(out.Sources))
			for _, s := range out.Sources {
				domain := s.Domain
				if domain == "" {
					domain = research.Domain(s.URL)
				}
				sources = append(sources, project.Source{URL: s.URL, Title: s.Title, Domain: domain, SourceType: "web"})
			}
			if err := p.projects.SaveResearchSources(ctx, j.ProjectID, j.CarouselID, sources); err != nil {
				log.Warn("saving research sources failed", "error", err)
			}
		}
		cp.ResearchJSON = ai.MustJSON(out)
	}

	// ---- 2. Content (§18) + 3. semantic validation (§119) ----
	var body content.Content
	if !cp.ContentSaved {
		if err := p.jobs.Advance(ctx, j.ID, job.StatusGeneratingContent, cp); err != nil {
			return err
		}
		disclaimer := ""
		if content.SensitiveNiche(proj.Niche + " " + topic) {
			disclaimer = content.Disclaimer(proj.Language)
		}
		body, err = p.ai.GenerateContent(ctx, meta, ai.ContentArgs{
			ProjectContext: brief,
			Inputs:         renderInputs(inputs),
			Research:       cp.ResearchJSON,
			CTA:            pctx.CTALabel(ctaID),
			SlideCount:     slideCount,
			Language:       proj.Language,
			Disclaimer:     disclaimer,
		})
		if err != nil {
			return fmt.Errorf("content generation: %w", err)
		}
		if err := p.jobs.Advance(ctx, j.ID, job.StatusValidatingContent, cp); err != nil {
			return err
		}
		if rep := content.Validate(&body, slideCount); !rep.OK() {
			return fmt.Errorf("content rejected after retries: %s", rep.Instruction())
		}
		if _, err := p.carousels.SaveContent(ctx, j.CarouselID, body); err != nil {
			return err
		}
		cp.ContentSaved = true
	} else {
		if body, err = p.carousels.LatestContent(ctx, j.CarouselID); err != nil {
			return err
		}
	}

	// ---- 4. Formula selection (§19) ----
	if cp.FormulaID == "" {
		if err := p.jobs.Advance(ctx, j.ID, job.StatusSelectingFormula, cp); err != nil {
			return err
		}
		formulas, err := p.registry.Formulas(ctx)
		if err != nil {
			return err
		}
		list, allowed := registry.FormulaPrompt(formulas)
		sel, err := p.ai.SelectFormula(ctx, meta, list, ai.MustJSON(body), allowed)
		if err != nil {
			log.Warn("formula selection failed, defaulting to listicle", "error", err)
			sel = ai.FormulaSelection{FormulaID: "listicle", Reason: "default"}
		}
		cp.FormulaID = sel.FormulaID
		body.FormulaID = sel.FormulaID
		body.FormulaReason = sel.Reason
		if _, err := p.carousels.SaveContent(ctx, j.CarouselID, body); err != nil {
			return err
		}
		if err := p.carousels.SetFormula(ctx, j.CarouselID, sel.FormulaID, body.Title); err != nil {
			return err
		}
	}

	// ---- 5. Design (§21) ----
	var d design.Design
	if !cp.DesignSaved {
		if err := p.jobs.Advance(ctx, j.ID, job.StatusGeneratingDesign, cp); err != nil {
			return err
		}
		formula, _ := p.registry.Formula(ctx, cp.FormulaID)
		marks := carousel.BrandMarks(proj.Brand)
		canvas := design.Canvas{Width: c.CanvasWidth, Height: c.CanvasHeight, Ratio: c.CanvasRatio}
		canvas.ReservedTop, canvas.ReservedBottom = design.Reserve(marks)
		canvas.RecomputeSafeArea()
		d, err = p.ai.GenerateDesign(ctx, meta, ai.DesignArgs{
			ProjectContext: brief,
			ContentJSON:    ai.MustJSON(body),
			Formula:        formula.Name + " → " + fmt.Sprint(formula.Structure),
			TemplateStyle:  ai.TemplateStyleFor(j.CarouselID),
			Palette:        proj.Brand.Palette(),
			Canvas:         canvas,
		})
		if err != nil {
			return fmt.Errorf("design generation: %w", err)
		}
		// Carry each content slide's image intent onto the design slide so the
		// image step has a query even if the model omitted one.
		for i := range d.Slides {
			if i < len(body.Slides) {
				if d.Slides[i].ImageQuery == "" {
					d.Slides[i].ImageQuery = body.Slides[i].ImageQuery
				}
				if d.Slides[i].ImageIntent == "" {
					d.Slides[i].ImageIntent = body.Slides[i].ImageIntent
				}
			}
		}
		// Brand marks are stamped after validation so they keep their place in
		// the outer margin instead of being clamped into the safe area (§39).
		design.ApplyBrand(&d, marks)
		log.Info("brand marks applied", "count", design.CountBrandElements(&d),
			"reserved_top", d.Canvas.ReservedTop, "reserved_bottom", d.Canvas.ReservedBottom)
		if _, err := p.carousels.SaveDesign(ctx, j.CarouselID, d); err != nil {
			return err
		}
		cp.DesignSaved = true
	} else {
		rec, err := p.carousels.LatestDesign(ctx, j.CarouselID)
		if err != nil {
			return err
		}
		d = rec.Design
	}

	// ---- 6. Images (§33) ----
	if !cp.ImagesDone {
		if err := p.jobs.Advance(ctx, j.ID, job.StatusSearchingImages, cp); err != nil {
			return err
		}
		p.attachImages(ctx, j, &d, log)
		rec, err := p.carousels.LatestDesign(ctx, j.CarouselID)
		if err != nil {
			return err
		}
		if _, err := p.carousels.UpdateDesign(ctx, j.CarouselID, rec.Version, d); err != nil {
			return err
		}
		cp.ImagesDone = true
	}

	// ---- 7. Finalize ----
	if err := p.jobs.Advance(ctx, j.ID, job.StatusFinalizing, cp); err != nil {
		return err
	}
	// Post copy is a nice-to-have: it must never fail the carousel.
	p.captions.GenerateQuietly(ctx, meta, j.CarouselID)
	if err := p.exports.RenderThumbnail(ctx, j.UserID, j.ProjectID, j.CarouselID); err != nil {
		log.Warn("thumbnail render failed", "error", err)
	}
	return nil
}

// attachImages searches candidates for every slide in parallel and ingests the
// first one as a starting background; the creator can swap it in the editor.
func (p *Pipeline) attachImages(ctx context.Context, j job.Job, d *design.Design, log *slog.Logger) {
	type result struct {
		slideIdx int
		assetID  string
	}
	results := make(chan result, len(d.Slides))
	var wg sync.WaitGroup

	for i, slide := range d.Slides {
		query := slide.ImageQuery
		if query == "" {
			query = slide.ImageIntent
		}
		if query == "" {
			continue
		}
		wg.Add(1)
		go func(idx int, slideID, query string) {
			defer wg.Done()
			candidates, err := p.images.Search(ctx, j.CarouselID, slideID, query, 1)
			if err != nil || len(candidates) == 0 {
				log.Warn("image search returned nothing", "slide", slideID, "query", query, "error", err)
				return
			}
			a, err := p.images.Ingest(ctx, j.UserID, j.ProjectID, j.CarouselID, slideID, candidates[0])
			if err != nil {
				log.Warn("image ingest failed", "slide", slideID, "error", err)
				return
			}
			results <- result{idx, a.ID}
		}(i, slide.ID, query)
	}
	wg.Wait()
	close(results)

	for r := range results {
		for ei := range d.Slides[r.slideIdx].Elements {
			if d.Slides[r.slideIdx].Elements[ei].Type == "image" {
				d.Slides[r.slideIdx].Elements[ei].AssetID = r.assetID
				break
			}
		}
	}
}

func renderInputs(inputs map[string]any) string {
	body, _ := json.MarshalIndent(inputs, "", "  ")
	return "topic: " + schema.Topic(inputs) + "\n" + string(body)
}

// classify maps an error to a stable code plus a user-safe message (§90, §91).
func classify(err error) (string, string) {
	msg := err.Error()
	switch {
	case containsAny(msg, "research:"):
		return "research_failed", "Chúng tôi chưa nghiên cứu được chủ đề này."
	case containsAny(msg, "content generation:", "content rejected"):
		return "content_failed", "Chúng tôi chưa viết được nội dung phù hợp."
	case containsAny(msg, "design generation:"):
		return "design_failed", "Chúng tôi chưa dựng được thiết kế."
	default:
		return "generation_failed", "Chúng tôi chưa tạo được carousel này."
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
