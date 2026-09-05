package carousel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tks/backend/internal/content"
	"github.com/tks/backend/internal/design"
	"github.com/tks/backend/internal/httpx"
)

type Repo struct{ db *pgxpool.Pool }

func NewRepo(db *pgxpool.Pool) *Repo { return &Repo{db: db} }

func (r *Repo) DB() *pgxpool.Pool { return r.db }

// Every read joins projects so ownership is enforced in SQL, not in a handler
// that someone might forget to guard (§86).
const selectCarousel = `
	SELECT c.id, c.project_id, c.title, c.status, c.platform, c.canvas_ratio,
	       c.canvas_width, c.canvas_height, c.formula_id, c.thumbnail_key,
	       c.created_at, c.updated_at,
	       COALESCE(i.schema_version, 0), COALESCE(i.input_json, '{}'::jsonb),
	       COALESCE(d.version, 0)
	FROM carousels c
	JOIN projects p ON p.id = c.project_id AND p.deleted_at IS NULL
	LEFT JOIN LATERAL (
	    SELECT schema_version, input_json FROM carousel_inputs
	    WHERE carousel_id = c.id ORDER BY created_at DESC LIMIT 1
	) i ON true
	LEFT JOIN LATERAL (
	    SELECT version FROM carousel_designs WHERE carousel_id = c.id ORDER BY version DESC LIMIT 1
	) d ON true`

func scan(row pgx.Row) (Carousel, error) {
	var c Carousel
	var input []byte
	err := row.Scan(&c.ID, &c.ProjectID, &c.Title, &c.Status, &c.Platform, &c.CanvasRatio,
		&c.CanvasWidth, &c.CanvasHeight, &c.FormulaID, &c.thumbnailKey,
		&c.CreatedAt, &c.UpdatedAt, &c.SchemaVersion, &input, &c.DesignVersion)
	if err != nil {
		return c, err
	}
	c.Input = input
	return c, nil
}

type CreateParams struct {
	ProjectID     string
	Title         string
	Ratio         string
	SchemaVersion int
	Input         map[string]any
}

// Create writes the carousel and its input in one transaction (§83).
func (r *Repo) Create(ctx context.Context, p CreateParams) (Carousel, error) {
	preset := design.PresetFor(p.Ratio)
	inputJSON, err := json.Marshal(p.Input)
	if err != nil {
		return Carousel{}, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Carousel{}, err
	}
	defer tx.Rollback(ctx)

	var id string
	if err := tx.QueryRow(ctx, `
		INSERT INTO carousels (project_id, title, status, canvas_ratio, canvas_width, canvas_height)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		p.ProjectID, p.Title, StatusGenerating, preset.Ratio, preset.Width, preset.Height).Scan(&id); err != nil {
		return Carousel{}, err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO carousel_inputs (carousel_id, schema_version, input_json) VALUES ($1,$2,$3)`,
		id, p.SchemaVersion, inputJSON); err != nil {
		return Carousel{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Carousel{}, err
	}
	return r.getByID(ctx, id)
}

func (r *Repo) getByID(ctx context.Context, id string) (Carousel, error) {
	return scan(r.db.QueryRow(ctx, selectCarousel+` WHERE c.id = $1 AND c.deleted_at IS NULL`, id))
}

func (r *Repo) Get(ctx context.Context, userID, id string) (Carousel, error) {
	c, err := scan(r.db.QueryRow(ctx, selectCarousel+`
		WHERE c.id = $1 AND p.user_id = $2 AND c.deleted_at IS NULL`, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return c, httpx.ErrNotFound
	}
	return c, err
}

// GetForWorker skips the user check — the worker already validated ownership
// when it created the job.
func (r *Repo) GetForWorker(ctx context.Context, id string) (Carousel, error) {
	c, err := r.getByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, httpx.ErrNotFound
	}
	return c, err
}

type ListFilter struct {
	Status string
	Query  string
}

func (r *Repo) List(ctx context.Context, userID, projectID string, f ListFilter) ([]Carousel, error) {
	sql := selectCarousel + ` WHERE c.project_id = $1 AND p.user_id = $2 AND c.deleted_at IS NULL`
	args := []any{projectID, userID}
	if f.Status != "" && f.Status != "all" {
		args = append(args, f.Status)
		sql += fmt.Sprintf(" AND c.status = $%d", len(args))
	}
	if f.Query != "" {
		args = append(args, "%"+f.Query+"%")
		sql += fmt.Sprintf(" AND c.title ILIKE $%d", len(args))
	}
	sql += " ORDER BY c.updated_at DESC LIMIT 200"

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Carousel{}
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repo) UpdateMeta(ctx context.Context, userID, id string, title, status *string) (Carousel, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE carousels c SET
		    title = COALESCE($3, c.title),
		    status = COALESCE($4, c.status),
		    updated_at = now()
		FROM projects p
		WHERE c.id = $1 AND c.project_id = p.id AND p.user_id = $2 AND c.deleted_at IS NULL`,
		id, userID, title, status)
	if err != nil {
		return Carousel{}, err
	}
	if tag.RowsAffected() == 0 {
		return Carousel{}, httpx.ErrNotFound
	}
	return r.Get(ctx, userID, id)
}

// SetStatus is the worker's path; it does not need a user.
func (r *Repo) SetStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE carousels SET status=$2, updated_at=now() WHERE id=$1`, id, status)
	return err
}

func (r *Repo) SetFormula(ctx context.Context, id, formulaID, title string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE carousels SET formula_id=$2, title=COALESCE(NULLIF($3,''), title), updated_at=now() WHERE id=$1`,
		id, formulaID, title)
	return err
}

func (r *Repo) SetThumbnailKey(ctx context.Context, id, key string) error {
	_, err := r.db.Exec(ctx, `UPDATE carousels SET thumbnail_key=$2 WHERE id=$1`, id, key)
	return err
}

func (r *Repo) SoftDelete(ctx context.Context, userID, id string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE carousels c SET deleted_at = now()
		FROM projects p
		WHERE c.id=$1 AND c.project_id=p.id AND p.user_id=$2 AND c.deleted_at IS NULL`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpx.ErrNotFound
	}
	return nil
}

// ---- content ----

func (r *Repo) SaveContent(ctx context.Context, carouselID string, c content.Content) (int, error) {
	body, err := json.Marshal(c)
	if err != nil {
		return 0, err
	}
	var version int
	err = r.db.QueryRow(ctx, `
		INSERT INTO carousel_content (carousel_id, version, content_json)
		VALUES ($1, COALESCE((SELECT max(version) FROM carousel_content WHERE carousel_id=$1),0)+1, $2)
		RETURNING version`, carouselID, body).Scan(&version)
	return version, err
}

func (r *Repo) LatestContent(ctx context.Context, carouselID string) (content.Content, error) {
	var body []byte
	err := r.db.QueryRow(ctx,
		`SELECT content_json FROM carousel_content WHERE carousel_id=$1 ORDER BY version DESC LIMIT 1`,
		carouselID).Scan(&body)
	if errors.Is(err, pgx.ErrNoRows) {
		return content.Content{}, httpx.ErrNotFound
	}
	if err != nil {
		return content.Content{}, err
	}
	var c content.Content
	err = json.Unmarshal(body, &c)
	return c, err
}

// ---- design ----

func (r *Repo) SaveDesign(ctx context.Context, carouselID string, d design.Design) (int, error) {
	body, err := json.Marshal(d)
	if err != nil {
		return 0, err
	}
	var version int
	err = r.db.QueryRow(ctx, `
		INSERT INTO carousel_designs (carousel_id, version, design_json)
		VALUES ($1, COALESCE((SELECT max(version) FROM carousel_designs WHERE carousel_id=$1),0)+1, $2)
		RETURNING version`, carouselID, body).Scan(&version)
	if err != nil {
		return 0, err
	}
	// Snapshot every generated version so history exists from day one (§44).
	_, _ = r.db.Exec(ctx,
		`INSERT INTO carousel_versions (carousel_id, version, label, design_json) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (carousel_id, version) DO NOTHING`, carouselID, version, "generated", body)
	return version, nil
}

type DesignRecord struct {
	Version int           `json:"version"`
	Design  design.Design `json:"design"`
}

func (r *Repo) LatestDesign(ctx context.Context, carouselID string) (DesignRecord, error) {
	var rec DesignRecord
	var body []byte
	err := r.db.QueryRow(ctx,
		`SELECT version, design_json FROM carousel_designs WHERE carousel_id=$1 ORDER BY version DESC LIMIT 1`,
		carouselID).Scan(&rec.Version, &body)
	if errors.Is(err, pgx.ErrNoRows) {
		return rec, httpx.ErrNotFound
	}
	if err != nil {
		return rec, err
	}
	err = json.Unmarshal(body, &rec.Design)
	return rec, err
}

// UpdateDesign applies an autosave with optimistic concurrency: the caller must
// hold the version currently stored, otherwise it gets a 409 and reloads (§80).
func (r *Repo) UpdateDesign(ctx context.Context, carouselID string, baseVersion int, d design.Design) (int, error) {
	body, err := json.Marshal(d)
	if err != nil {
		return 0, err
	}
	var current int
	if err := r.db.QueryRow(ctx,
		`SELECT COALESCE(max(version),0) FROM carousel_designs WHERE carousel_id=$1`, carouselID).Scan(&current); err != nil {
		return 0, err
	}
	if current != baseVersion {
		return current, httpx.ErrConflict
	}
	var version int
	err = r.db.QueryRow(ctx, `
		INSERT INTO carousel_designs (carousel_id, version, design_json)
		VALUES ($1, $2, $3) RETURNING version`, carouselID, current+1, body).Scan(&version)
	if err != nil {
		// A concurrent writer won the race and took version current+1.
		return current, httpx.ErrConflict
	}
	_, _ = r.db.Exec(ctx, `UPDATE carousels SET updated_at = now() WHERE id = $1`, carouselID)
	return version, nil
}

func (r *Repo) SnapshotVersion(ctx context.Context, carouselID string, version int, label string, d design.Design) error {
	body, _ := json.Marshal(d)
	_, err := r.db.Exec(ctx,
		`INSERT INTO carousel_versions (carousel_id, version, label, design_json) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (carousel_id, version) DO UPDATE SET label = EXCLUDED.label`,
		carouselID, version, label, body)
	return err
}
