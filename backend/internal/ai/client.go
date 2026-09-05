package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Auditor records every AI call so a bad carousel can be traced to the exact
// generation that produced it (§70, §92).
type Auditor interface {
	Record(ctx context.Context, row AuditRow)
}

type AuditRow struct {
	JobID         string
	UserID        string
	Provider      string
	Model         string
	TaskType      string
	PromptVersion string
	Input         any
	Output        any
	Usage         Usage
	LatencyMS     int
	Status        string
}

type Client struct {
	provider    Provider
	maxAttempts int
	auditor     Auditor
}

func NewClient(p Provider, maxAttempts int, auditor Auditor) *Client {
	if maxAttempts < 1 {
		maxAttempts = 3
	}
	return &Client{provider: p, maxAttempts: maxAttempts, auditor: auditor}
}

func (c *Client) Provider() Provider { return c.provider }

// Call context carried through a pipeline run, used only for audit rows.
type CallMeta struct {
	JobID  string
	UserID string
}

// ValidationError signals output that parsed but failed a business rule. It is
// fed back to the model as a repair instruction rather than surfaced (§55).
type ValidationError struct{ Instruction string }

func (e *ValidationError) Error() string { return "ai output rejected: " + e.Instruction }

func Reject(format string, args ...any) error {
	return &ValidationError{Instruction: fmt.Sprintf(format, args...)}
}

// run executes one AI task with the generate → parse → validate → repair loop.
// semantic may mutate out (normalisation) and may return a ValidationError to
// request a repair attempt. Attempts are capped (§55).
func run[T any](ctx context.Context, c *Client, meta CallMeta, task TaskType, vars map[string]any, semantic func(*T) error) (T, error) {
	var zero T
	prompt, err := promptFor(task)
	if err != nil {
		return zero, err
	}
	user, err := prompt.render(vars)
	if err != nil {
		return zero, fmt.Errorf("render prompt %s: %w", task, err)
	}

	var lastErr error
	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		userMsg := user
		if lastErr != nil {
			userMsg = user + "\n\nYour previous answer was rejected. Fix exactly this and return the full JSON again:\n" + lastErr.Error()
		}

		start := time.Now()
		resp, err := c.provider.Complete(ctx, Request{
			Task: task, System: prompt.System, User: userMsg, Vars: vars, MaxTokens: maxTokensFor(task),
		})
		latency := int(time.Since(start).Milliseconds())

		if err != nil {
			lastErr = err
			c.audit(ctx, meta, task, prompt.Version, vars, nil, Usage{}, latency, "provider_error")
			slog.WarnContext(ctx, "ai call failed", "task", task, "attempt", attempt, "error", err)
			if ctx.Err() != nil {
				return zero, ctx.Err()
			}
			backoff(ctx, attempt)
			continue
		}

		var out T
		if err := json.Unmarshal([]byte(extractJSON(resp.Text)), &out); err != nil {
			lastErr = Reject("the response was not valid JSON (%v)", err)
			c.audit(ctx, meta, task, prompt.Version, vars, resp.Text, resp.Usage, latency, "invalid_json")
			continue
		}
		if semantic != nil {
			if err := semantic(&out); err != nil {
				var ve *ValidationError
				if errors.As(err, &ve) {
					lastErr = ve
					c.audit(ctx, meta, task, prompt.Version, vars, out, resp.Usage, latency, "rejected")
					continue
				}
				return zero, err
			}
		}
		c.audit(ctx, meta, task, prompt.Version, vars, out, resp.Usage, latency, "ok")
		return out, nil
	}
	return zero, fmt.Errorf("ai task %s failed after %d attempts: %w", task, c.maxAttempts, lastErr)
}

func backoff(ctx context.Context, attempt int) {
	select {
	case <-ctx.Done():
	case <-time.After(time.Duration(attempt*attempt) * time.Second):
	}
}

func maxTokensFor(task TaskType) int {
	switch task {
	case TaskDesign, TaskContent:
		return 8000
	case TaskProjectAnalysis, TaskCarouselSchema, TaskResearch:
		return 3000
	default:
		return 1500
	}
}

func (c *Client) audit(ctx context.Context, meta CallMeta, task TaskType, version string, in, out any, usage Usage, latency int, status string) {
	if c.auditor == nil {
		return
	}
	c.auditor.Record(ctx, AuditRow{
		JobID: meta.JobID, UserID: meta.UserID,
		Provider: c.provider.Name(), Model: c.provider.Model(),
		TaskType: string(task), PromptVersion: version,
		Input: in, Output: out, Usage: usage, LatencyMS: latency, Status: status,
	})
}

// extractJSON tolerates models that wrap JSON in prose or markdown fences.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if fence := strings.Index(s, "```"); fence >= 0 {
		rest := s[fence+3:]
		if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
			rest = rest[nl+1:]
		}
		if end := strings.Index(rest, "```"); end >= 0 {
			rest = rest[:end]
		}
		s = strings.TrimSpace(rest)
	}
	start := strings.IndexByte(s, '{')
	end := strings.LastIndexByte(s, '}')
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
