package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/tks/backend/internal/export"
	"github.com/tks/backend/internal/job"
)

// Worker drains the Redis queues. It is a separate process from the API so a
// long AI run never blocks an HTTP request (§82).
type Worker struct {
	queue       *job.Queue
	jobs        *job.Repo
	pipeline    *Pipeline
	exports     *export.Service
	concurrency int
}

func New(queue *job.Queue, jobs *job.Repo, pipeline *Pipeline, exports *export.Service, concurrency int) *Worker {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Worker{queue: queue, jobs: jobs, pipeline: pipeline, exports: exports, concurrency: concurrency}
}

func (w *Worker) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			w.loop(ctx, job.QueueGeneration, n, w.handleGeneration)
		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.loop(ctx, job.QueueExport, 0, w.handleExport)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		w.reapStale(ctx)
	}()

	slog.Info("worker started", "generation_workers", w.concurrency)
	wg.Wait()
}

func (w *Worker) loop(ctx context.Context, queue string, n int, handle func(context.Context, string)) {
	for {
		if ctx.Err() != nil {
			return
		}
		id, err := w.queue.Dequeue(ctx, queue, 5*time.Second)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("dequeue failed", "queue", queue, "error", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if id == "" {
			continue
		}
		slog.Info("job picked up", "queue", queue, "job_id", id, "worker", n)
		handle(ctx, id)
	}
}

func (w *Worker) handleGeneration(ctx context.Context, jobID string) {
	// A generation gets its own deadline so one stuck AI call cannot pin a worker.
	runCtx, cancel := context.WithTimeout(ctx, 12*time.Minute)
	defer cancel()
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("generation panicked", "job_id", jobID, "recover", rec)
			_ = w.jobs.Fail(context.WithoutCancel(ctx), jobID, "internal_error", "worker panic")
		}
	}()
	if err := w.pipeline.Run(runCtx, jobID); err != nil {
		slog.Error("generation job failed", "job_id", jobID, "error", err)
	}
}

func (w *Worker) handleExport(ctx context.Context, exportID string) {
	runCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	if err := w.exports.Run(runCtx, exportID); err != nil {
		slog.Error("export job failed", "export_id", exportID, "error", err)
	}
}

// reapStale fails jobs a crashed worker abandoned, so the UI never spins forever.
func (w *Worker) reapStale(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stale, err := w.jobs.StaleRunning(ctx, 15*time.Minute)
			if err != nil {
				continue
			}
			for _, j := range stale {
				slog.Warn("reaping stale job", "job_id", j.ID, "status", j.Status)
				_ = w.jobs.Fail(ctx, j.ID, "timeout", "generation timed out")
			}
		}
	}
}
