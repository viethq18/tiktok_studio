package image

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tks/backend/internal/asset"
	"github.com/tks/backend/internal/carousel"
	"github.com/tks/backend/internal/config"
	"github.com/tks/backend/internal/httpx"
)

const CandidatesPerSlide = 3

type Service struct {
	provider Provider
	db       *pgxpool.Pool
	assets   *asset.Service
	carousel *carousel.Repo
}

func NewService(p Provider, db *pgxpool.Pool, assets *asset.Service, carousels *carousel.Repo) *Service {
	return &Service{provider: p, db: db, assets: assets, carousel: carousels}
}

// NewProvider selects the configured provider, falling back to the offline stub
// when no Unsplash key is present.
func NewProvider(cfg config.Config) Provider {
	if cfg.ImageProvider == "unsplash" && cfg.UnsplashAccessKey != "" {
		return NewUnsplash(cfg.UnsplashAccessKey)
	}
	slog.Warn("UNSPLASH_ACCESS_KEY is empty — using the picsum placeholder provider")
	return NewPicsum()
}

func (s *Service) ProviderName() string { return s.provider.Name() }

// Search returns candidates for a slide, reading through a cache so repeated
// visits to the editor do not burn the provider's hourly quota (§71, §156).
func (s *Service) Search(ctx context.Context, carouselID, slideID, query string, page int) ([]Candidate, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, httpx.BadRequest("Chưa có từ khoá tìm ảnh cho slide này.")
	}
	if page < 1 {
		page = 1
	}
	if cached, ok := s.cached(ctx, carouselID, slideID, query, page); ok {
		return cached, nil
	}
	results, err := s.provider.Search(ctx, query, page, CandidatesPerSlide)
	if err != nil {
		return nil, httpx.Wrap(502, "image_search_failed", "Chưa tìm được ảnh. Vui lòng thử lại.", err)
	}
	s.cache(ctx, carouselID, slideID, query, page, results)
	return results, nil
}

func (s *Service) cached(ctx context.Context, carouselID, slideID, query string, page int) ([]Candidate, bool) {
	var body []byte
	err := s.db.QueryRow(ctx, `
		SELECT results_json FROM image_searches
		WHERE carousel_id=$1 AND slide_id=$2 AND query=$3 AND page=$4 AND provider=$5
		  AND created_at > now() - interval '7 days'
		ORDER BY created_at DESC LIMIT 1`,
		carouselID, slideID, query, page, s.provider.Name()).Scan(&body)
	if err != nil {
		return nil, false
	}
	var out []Candidate
	if json.Unmarshal(body, &out) != nil || len(out) == 0 {
		return nil, false
	}
	return out, true
}

func (s *Service) cache(ctx context.Context, carouselID, slideID, query string, page int, results []Candidate) {
	body, err := json.Marshal(results)
	if err != nil {
		return
	}
	var cid any
	if carouselID != "" {
		cid = carouselID
	}
	if _, err := s.db.Exec(context.WithoutCancel(ctx), `
		INSERT INTO image_searches (carousel_id, slide_id, query, provider, page, results_json)
		VALUES ($1,$2,$3,$4,$5,$6)`, cid, slideID, query, s.provider.Name(), page, body); err != nil {
		slog.Warn("image search cache write failed", "error", err)
	}
}

// Ingest downloads a candidate and stores it as an owned asset (§36).
func (s *Service) Ingest(ctx context.Context, userID, projectID, carouselID, slideID string, c Candidate) (asset.Asset, error) {
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	data, err := s.provider.Download(ctx, c)
	if err != nil {
		return asset.Asset{}, httpx.Wrap(502, "image_download_failed", "Không tải được ảnh. Vui lòng chọn ảnh khác.", err)
	}
	a, err := s.assets.Store(ctx, data, asset.StoreParams{
		UserID: userID, ProjectID: projectID, CarouselID: carouselID,
		Source: c.Provider, SourceID: c.ID,
		// Attribution is retained even though the MVP has no UI for it (§156).
		Metadata: map[string]any{
			"photographer_name": c.PhotographerName,
			"photographer_url":  c.PhotographerURL,
			"source_url":        c.SourceURL,
			"description":       c.Description,
			"slide_id":          slideID,
		},
	})
	if err != nil {
		return asset.Asset{}, err
	}
	if err := s.assets.Attach(ctx, carouselID, a.ID, slideID, "background"); err != nil {
		slog.Warn("attach asset failed", "error", err, "asset_id", a.ID)
	}
	return a, nil
}

// ResolveCandidate looks a candidate id up in this carousel's own cached search
// results. The client sends an id, never a URL: the server must never fetch a
// URL supplied by the browser (§84).
func (s *Service) ResolveCandidate(ctx context.Context, carouselID, slideID, candidateID string) (Candidate, error) {
	rows, err := s.db.Query(ctx, `
		SELECT results_json FROM image_searches
		WHERE carousel_id=$1 AND slide_id=$2 AND provider=$3
		  AND created_at > now() - interval '7 days'
		ORDER BY created_at DESC LIMIT 20`, carouselID, slideID, s.provider.Name())
	if err != nil {
		return Candidate{}, httpx.Internal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var body []byte
		if err := rows.Scan(&body); err != nil {
			return Candidate{}, httpx.Internal(err)
		}
		var candidates []Candidate
		if json.Unmarshal(body, &candidates) != nil {
			continue
		}
		for _, c := range candidates {
			if c.ID == candidateID {
				return c, nil
			}
		}
	}
	return Candidate{}, httpx.BadRequest("Ảnh này không còn trong kết quả tìm kiếm. Hãy tìm lại.")
}

// SelectForSlide ingests a candidate and writes it into the slide's background
// image element, producing a new design version.
func (s *Service) SelectForSlide(ctx context.Context, userID, projectID, carouselID, slideID, candidateID string, baseVersion int) (int, asset.Asset, error) {
	c, err := s.ResolveCandidate(ctx, carouselID, slideID, candidateID)
	if err != nil {
		return 0, asset.Asset{}, err
	}
	a, err := s.Ingest(ctx, userID, projectID, carouselID, slideID, c)
	if err != nil {
		return 0, asset.Asset{}, err
	}
	rec, err := s.carousel.LatestDesign(ctx, carouselID)
	if err != nil {
		return 0, asset.Asset{}, err
	}
	if baseVersion > 0 && baseVersion != rec.Version {
		return rec.Version, asset.Asset{}, httpx.ErrConflict
	}
	found := false
	for i := range rec.Design.Slides {
		if rec.Design.Slides[i].ID != slideID {
			continue
		}
		for j := range rec.Design.Slides[i].Elements {
			if rec.Design.Slides[i].Elements[j].Type == "image" {
				rec.Design.Slides[i].Elements[j].AssetID = a.ID
				found = true
				break
			}
		}
	}
	if !found {
		return rec.Version, asset.Asset{}, httpx.BadRequest("Slide này không có vùng ảnh nền.")
	}
	version, err := s.carousel.UpdateDesign(ctx, carouselID, rec.Version, rec.Design)
	if err != nil {
		return version, asset.Asset{}, err
	}
	return version, a, nil
}
