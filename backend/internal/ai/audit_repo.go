package ai

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PgAuditor persists one row per AI call. Auditing must never break a
// generation, so every failure here is logged and swallowed.
type PgAuditor struct{ db *pgxpool.Pool }

func NewPgAuditor(db *pgxpool.Pool) *PgAuditor { return &PgAuditor{db: db} }

func (a *PgAuditor) Record(ctx context.Context, row AuditRow) {
	input, _ := json.Marshal(redact(row.Input))
	output, _ := json.Marshal(row.Output)
	usage, _ := json.Marshal(row.Usage)

	var jobID, userID any
	if row.JobID != "" {
		jobID = row.JobID
	}
	if row.UserID != "" {
		userID = row.UserID
	}
	// Detached context: the audit should survive a cancelled request.
	if _, err := a.db.Exec(context.WithoutCancel(ctx), `
		INSERT INTO ai_generations
			(job_id, user_id, provider, model, task_type, prompt_version, input_json, output_json, token_usage, latency_ms, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		jobID, userID, row.Provider, row.Model, row.TaskType, row.PromptVersion,
		input, output, usage, row.LatencyMS, row.Status); err != nil {
		slog.Warn("ai audit write failed", "error", err, "task", row.TaskType)
	}
}

// redact keeps the structured inputs but never the rendered prompt text, which
// can carry more context than we want to retain (§70).
func redact(in any) any {
	vars, ok := in.(map[string]any)
	if !ok {
		return in
	}
	out := make(map[string]any, len(vars))
	for k, v := range vars {
		if s, isStr := v.(string); isStr && len([]rune(s)) > 600 {
			out[k] = string([]rune(s)[:600]) + "…[truncated]"
			continue
		}
		out[k] = v
	}
	return out
}
