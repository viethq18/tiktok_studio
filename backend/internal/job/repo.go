package job

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tks/backend/internal/httpx"
)

type Repo struct{ db *pgxpool.Pool }

func NewRepo(db *pgxpool.Pool) *Repo { return &Repo{db: db} }

const selectJob = `
	SELECT id, user_id, project_id, COALESCE(carousel_id::text,''), type, status, progress,
	       current_step, last_completed_step, step_outputs, attempt, error_code, error_message,
	       created_at, completed_at
	FROM generation_jobs`

func scan(row pgx.Row) (Job, error) {
	var j Job
	var status string
	var outputs []byte
	err := row.Scan(&j.ID, &j.UserID, &j.ProjectID, &j.CarouselID, &j.Type, &status, &j.Progress,
		&j.CurrentStep, &j.LastCompletedStep, &outputs, &j.Attempt, &j.ErrorCode, &j.ErrorMessage,
		&j.CreatedAt, &j.CompletedAt)
	j.Status = Status(status)
	j.StepOutputs = outputs
	return j, err
}

// FindByIdempotencyKey lets a repeated submit return the original job instead
// of starting a second generation (§159).
func (r *Repo) FindByIdempotencyKey(ctx context.Context, userID, key string) (Job, bool, error) {
	j, err := scan(r.db.QueryRow(ctx, selectJob+` WHERE user_id=$1 AND idempotency_key=$2`, userID, key))
	if errors.Is(err, pgx.ErrNoRows) {
		return Job{}, false, nil
	}
	return j, err == nil, err
}

func (r *Repo) Create(ctx context.Context, userID, projectID, carouselID, jobType, idempotencyKey string) (Job, error) {
	var key any
	if idempotencyKey != "" {
		key = idempotencyKey
	}
	step := StepFor(StatusQueued)
	var id string
	if err := r.db.QueryRow(ctx, `
		INSERT INTO generation_jobs (user_id, project_id, carousel_id, type, status, progress, current_step, idempotency_key)
		VALUES ($1,$2,$3,$4,'queued',$5,$6,$7) RETURNING id`,
		userID, projectID, carouselID, jobType, step.Progress, step.Label, key).Scan(&id); err != nil {
		return Job{}, err
	}
	return r.GetForWorker(ctx, id)
}

func (r *Repo) Get(ctx context.Context, userID, id string) (Job, error) {
	j, err := scan(r.db.QueryRow(ctx, selectJob+` WHERE id=$1 AND user_id=$2`, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return j, httpx.ErrNotFound
	}
	return j, err
}

func (r *Repo) GetForWorker(ctx context.Context, id string) (Job, error) {
	j, err := scan(r.db.QueryRow(ctx, selectJob+` WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return j, httpx.ErrNotFound
	}
	return j, err
}

// LatestForCarousel powers the editor's "is this still generating?" check.
func (r *Repo) LatestForCarousel(ctx context.Context, userID, carouselID string) (Job, bool, error) {
	j, err := scan(r.db.QueryRow(ctx,
		selectJob+` WHERE carousel_id=$1 AND user_id=$2 ORDER BY created_at DESC LIMIT 1`, carouselID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Job{}, false, nil
	}
	return j, err == nil, err
}

func (r *Repo) MarkStarted(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE generation_jobs SET attempt = attempt + 1, started_at = COALESCE(started_at, now()),
		 error_code='', error_message='' WHERE id=$1`, id)
	return err
}

// Advance moves the job to a stage and records it as the resume point (§158).
func (r *Repo) Advance(ctx context.Context, id string, status Status, cp Checkpoint) error {
	step := StepFor(status)
	outputs, _ := json.Marshal(cp)
	_, err := r.db.Exec(ctx, `
		UPDATE generation_jobs
		SET status=$2, progress=$3, current_step=$4, last_completed_step=$5, step_outputs=$6
		WHERE id=$1`, id, string(status), step.Progress, step.Label, string(status), outputs)
	return err
}

func (r *Repo) Complete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE generation_jobs SET status='completed', progress=100, current_step='Hoàn tất',
		completed_at=now() WHERE id=$1`, id)
	return err
}

func (r *Repo) Fail(ctx context.Context, id, code, message string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE generation_jobs SET status='failed', error_code=$2, error_message=$3, completed_at=now()
		WHERE id=$1`, id, code, message)
	return err
}

// ResetForRetry puts a failed job back to its checkpoint so the retry skips the
// stages that already succeeded.
func (r *Repo) ResetForRetry(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE generation_jobs SET status='queued', error_code='', error_message='', completed_at=NULL
		WHERE id=$1`, id)
	return err
}

func (r *Repo) Checkpoint(j Job) Checkpoint {
	var cp Checkpoint
	if len(j.StepOutputs) > 0 {
		_ = json.Unmarshal(j.StepOutputs, &cp)
	}
	return cp
}

// StaleRunning finds jobs a crashed worker left behind.
func (r *Repo) StaleRunning(ctx context.Context, olderThan time.Duration) ([]Job, error) {
	rows, err := r.db.Query(ctx, selectJob+`
		WHERE status NOT IN ('completed','failed','queued') AND started_at < now() - $1::interval`,
		olderThan.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Job
	for rows.Next() {
		j, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}
