// Command worker drains the generation and export queues.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tks/backend/internal/app"
	"github.com/tks/backend/internal/config"
)

func main() {
	config.LoadDotEnv()
	cfg := config.Load()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bootCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	application, err := app.New(bootCtx, cfg)
	if err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
	defer application.Close()

	application.Worker.Run(ctx)
	slog.Info("worker stopped")
	time.Sleep(200 * time.Millisecond)
}
