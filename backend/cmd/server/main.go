// Command server runs the HTTP API.
package main

import (
	"context"
	"log/slog"
	"net/http"
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

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           application.Router(),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	go func() {
		slog.Info("api listening", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}
