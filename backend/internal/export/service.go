package export

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/png"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tks/backend/internal/carousel"
	"github.com/tks/backend/internal/httpx"
	"github.com/tks/backend/internal/platform"
)

const downloadTTL = 2 * time.Hour

type Job struct {
	ID          string     `json:"id"`
	CarouselID  string     `json:"carousel_id"`
	Format      string     `json:"format"`
	Status      string     `json:"status"`
	Progress    int        `json:"progress"`
	Error       string     `json:"error,omitempty"`
	DownloadURL string     `json:"download_url,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type Service struct {
	db       *pgxpool.Pool
	storage  *platform.Storage
	renderer *Renderer
	carousel *carousel.Repo
}

func NewService(db *pgxpool.Pool, storage *platform.Storage, renderer *Renderer, carousels *carousel.Repo) *Service {
	return &Service{db: db, storage: storage, renderer: renderer, carousel: carousels}
}

func (s *Service) Create(ctx context.Context, userID, carouselID string) (Job, error) {
	var j Job
	err := s.db.QueryRow(ctx, `
		INSERT INTO export_jobs (user_id, carousel_id, format, status)
		VALUES ($1,$2,'png_zip','queued')
		RETURNING id, carousel_id, format, status, progress, error_message, created_at, completed_at`,
		userID, carouselID).
		Scan(&j.ID, &j.CarouselID, &j.Format, &j.Status, &j.Progress, &j.Error, &j.CreatedAt, &j.CompletedAt)
	return j, err
}

func (s *Service) Get(ctx context.Context, userID, id string) (Job, error) {
	var j Job
	var key string
	err := s.db.QueryRow(ctx, `
		SELECT id, carousel_id, format, status, progress, error_message, storage_key, created_at, completed_at
		FROM export_jobs WHERE id=$1 AND user_id=$2`, id, userID).
		Scan(&j.ID, &j.CarouselID, &j.Format, &j.Status, &j.Progress, &j.Error, &key, &j.CreatedAt, &j.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return j, httpx.ErrNotFound
	}
	if err != nil {
		return j, err
	}
	if j.Status == "completed" && key != "" {
		j.DownloadURL, _ = s.storage.SignedURL(ctx, key, downloadTTL, "carousel.zip")
	}
	return j, nil
}

func (s *Service) GetForWorker(ctx context.Context, id string) (string, string, error) {
	var userID, carouselID string
	err := s.db.QueryRow(ctx, `SELECT user_id, carousel_id FROM export_jobs WHERE id=$1`, id).Scan(&userID, &carouselID)
	return userID, carouselID, err
}

func (s *Service) setProgress(ctx context.Context, id string, progress int) {
	_, _ = s.db.Exec(ctx, `UPDATE export_jobs SET status='rendering', progress=$2 WHERE id=$1`, id, progress)
}

func (s *Service) fail(ctx context.Context, id string, err error) {
	_, _ = s.db.Exec(context.WithoutCancel(ctx),
		`UPDATE export_jobs SET status='failed', error_message=$2, completed_at=now() WHERE id=$1`,
		id, err.Error())
}

// Run renders every slide to PNG, zips them and stores the archive (§47, §49).
func (s *Service) Run(ctx context.Context, exportID string) error {
	userID, carouselID, err := s.GetForWorker(ctx, exportID)
	if err != nil {
		return err
	}
	c, err := s.carousel.GetForWorker(ctx, carouselID)
	if err != nil {
		s.fail(ctx, exportID, err)
		return err
	}
	rec, err := s.carousel.LatestDesign(ctx, carouselID)
	if err != nil {
		s.fail(ctx, exportID, fmt.Errorf("carousel has no design yet"))
		return err
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for i, slide := range rec.Design.Slides {
		dc, err := s.renderer.RenderSlide(ctx, rec.Design, slide)
		if err != nil {
			s.fail(ctx, exportID, err)
			return err
		}
		f, err := zw.Create(fmt.Sprintf("%02d.png", i+1))
		if err != nil {
			s.fail(ctx, exportID, err)
			return err
		}
		if err := png.Encode(f, dc.Image()); err != nil {
			s.fail(ctx, exportID, err)
			return err
		}
		s.setProgress(ctx, exportID, int(float64(i+1)/float64(len(rec.Design.Slides))*90))
	}
	if err := zw.Close(); err != nil {
		s.fail(ctx, exportID, err)
		return err
	}

	key := platform.ExportKey(userID, c.ProjectID, carouselID, exportID)
	if err := s.storage.Put(ctx, key, bytes.NewReader(buf.Bytes()), int64(buf.Len()), "application/zip"); err != nil {
		s.fail(ctx, exportID, err)
		return err
	}
	_, err = s.db.Exec(ctx,
		`UPDATE export_jobs SET status='completed', progress=100, storage_key=$2, completed_at=now() WHERE id=$1`,
		exportID, key)
	return err
}

// RenderThumbnail stores a downscaled first slide for the dashboard grid (§41).
func (s *Service) RenderThumbnail(ctx context.Context, userID, projectID, carouselID string) error {
	rec, err := s.carousel.LatestDesign(ctx, carouselID)
	if err != nil || len(rec.Design.Slides) == 0 {
		return err
	}
	dc, err := s.renderer.RenderSlide(ctx, rec.Design, rec.Design.Slides[0])
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return err
	}
	key := platform.ThumbnailKey(userID, projectID, carouselID)
	if err := s.storage.Put(ctx, key, bytes.NewReader(buf.Bytes()), int64(buf.Len()), "image/png"); err != nil {
		return err
	}
	return s.carousel.SetThumbnailKey(ctx, carouselID, key)
}
