package ai

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tks/backend/internal/config"
)

// New builds the AI client for the running process. With no API key configured
// the deterministic mock provider takes over so the product is still fully
// operable offline (§90: never dead-end the user).
func New(cfg config.Config, db *pgxpool.Pool) *Client {
	var provider Provider
	if cfg.AIAPIKey == "" {
		slog.Warn("AI_API_KEY is empty — using the deterministic mock provider")
		provider = NewMockProvider()
	} else {
		provider = NewOpenAICompatible(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel, cfg.AITimeout)
	}
	return NewClient(provider, cfg.AIMaxAttempts, NewPgAuditor(db))
}
