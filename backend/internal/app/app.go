// Package app wires every module together once, so the API and the worker
// binaries share exactly the same construction (§51 modular monolith).
package app

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/tks/backend/internal/ai"
	"github.com/tks/backend/internal/asset"
	"github.com/tks/backend/internal/auth"
	"github.com/tks/backend/internal/carousel"
	"github.com/tks/backend/internal/config"
	"github.com/tks/backend/internal/export"
	"github.com/tks/backend/internal/image"
	"github.com/tks/backend/internal/job"
	"github.com/tks/backend/internal/platform"
	"github.com/tks/backend/internal/project"
	"github.com/tks/backend/internal/ratelimit"
	"github.com/tks/backend/internal/registry"
	"github.com/tks/backend/internal/research"
	"github.com/tks/backend/internal/worker"
)

type App struct {
	Cfg     config.Config
	DB      *pgxpool.Pool
	Redis   *redis.Client
	Storage *platform.Storage

	Auth      *auth.Service
	AuthH     *auth.Handler
	Projects  *project.Service
	ProjectH  *project.Handler
	Carousels *carousel.Service
	CarouselH *carousel.Handler
	Assets    *asset.Service
	Images    *image.Service
	ImageH    *image.Handler
	Exports   *export.Service
	ExportH   *export.Handler
	Registry  *registry.Handler
	Worker    *worker.Worker
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	db, err := platform.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := platform.Migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	rdb, err := platform.NewRedis(ctx, cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	storage, err := platform.NewStorage(ctx, cfg)
	if err != nil {
		return nil, err
	}

	aiClient := ai.New(cfg, db)
	searchProvider := research.NewProvider()
	limiter := ratelimit.New(rdb)

	authRepo := auth.NewRepo(db)
	authSvc := auth.NewService(authRepo, cfg)

	assetSvc := asset.NewService(db, storage)

	projectRepo := project.NewRepo(db)
	projectSvc := project.NewService(projectRepo, aiClient, searchProvider, assetSvc)

	carouselRepo := carousel.NewRepo(db)
	jobRepo := job.NewRepo(db)
	queue := job.NewQueue(rdb)
	imageSvc := image.NewService(image.NewProvider(cfg), db, assetSvc, carouselRepo)
	renderer := export.NewRenderer(assetSvc)
	exportSvc := export.NewService(db, storage, renderer, carouselRepo)

	captionSvc := carousel.NewCaptionService(carouselRepo, projectRepo, aiClient, searchProvider)
	carouselSvc := carousel.NewService(carouselRepo, projectSvc, jobRepo, queue, assetSvc, limiter, captionSvc)
	registryRepo := registry.NewRepo(db)

	pipeline := worker.NewPipeline(jobRepo, projectRepo, carouselRepo, registryRepo,
		aiClient, searchProvider, imageSvc, exportSvc, captionSvc)

	return &App{
		Cfg: cfg, DB: db, Redis: rdb, Storage: storage,
		Auth:      authSvc,
		AuthH:     auth.NewHandler(authSvc, cfg),
		Projects:  projectSvc,
		ProjectH:  project.NewHandler(projectSvc),
		Carousels: carouselSvc,
		CarouselH: carousel.NewHandler(carouselSvc, jobRepo),
		Assets:    assetSvc,
		Images:    imageSvc,
		ImageH:    image.NewHandler(imageSvc, carouselRepo, limiter),
		Exports:   exportSvc,
		ExportH:   export.NewHandler(exportSvc, carouselRepo, queue, limiter),
		Registry:  registry.NewHandler(registryRepo),
		Worker:    worker.New(queue, jobRepo, pipeline, exportSvc, cfg.WorkerConcurrency),
	}, nil
}

func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
	if a.Redis != nil {
		_ = a.Redis.Close()
	}
}
