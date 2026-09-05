package project

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tks/backend/internal/httpx"
)

type Repo struct{ db *pgxpool.Pool }

func NewRepo(db *pgxpool.Pool) *Repo { return &Repo{db: db} }

const selectProject = `
	SELECT p.id, p.user_id, p.name, p.niche, p.description, p.language, p.status,
	       p.brand_json, p.created_at, p.updated_at,
	       COALESCE(c.context_json, '{}'::jsonb), COALESCE(c.version, 0),
	       (SELECT count(*) FROM carousels ca WHERE ca.project_id = p.id AND ca.deleted_at IS NULL)
	FROM projects p
	LEFT JOIN LATERAL (
	    SELECT context_json, version FROM project_ai_contexts
	    WHERE project_id = p.id ORDER BY version DESC LIMIT 1
	) c ON true`

func scanProject(row pgx.Row) (Project, error) {
	var p Project
	var brand, ctx []byte
	err := row.Scan(&p.ID, &p.UserID, &p.Name, &p.Niche, &p.Description, &p.Language, &p.Status,
		&brand, &p.CreatedAt, &p.UpdatedAt, &ctx, &p.ContextVersion, &p.CarouselCount)
	if err != nil {
		return p, err
	}
	_ = json.Unmarshal(brand, &p.Brand)
	p.Context = ctx
	return p, nil
}

func (r *Repo) Create(ctx context.Context, userID, name, niche, description, language string) (Project, error) {
	// A data-modifying CTE is not visible to a SELECT in the same statement,
	// so the write and the read-back are two statements.
	var id string
	if err := r.db.QueryRow(ctx, `
		INSERT INTO projects (user_id, name, niche, description, language)
		VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		userID, name, niche, description, language).Scan(&id); err != nil {
		return Project{}, err
	}
	return r.Get(ctx, userID, id)
}

func (r *Repo) List(ctx context.Context, userID string) ([]Project, error) {
	rows, err := r.db.Query(ctx, selectProject+`
		WHERE p.user_id = $1 AND p.deleted_at IS NULL ORDER BY p.updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Project{}
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Get enforces ownership in the query itself: a wrong user id is a 404, never a
// leak of someone else's project (§86).
func (r *Repo) Get(ctx context.Context, userID, id string) (Project, error) {
	row := r.db.QueryRow(ctx, selectProject+`
		WHERE p.id = $1 AND p.user_id = $2 AND p.deleted_at IS NULL`, id, userID)
	p, err := scanProject(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, httpx.ErrNotFound
	}
	return p, err
}

func (r *Repo) Update(ctx context.Context, userID, id string, name, niche, description, language *string, brand *Brand) (Project, error) {
	var brandJSON []byte
	if brand != nil {
		brandJSON, _ = json.Marshal(brand)
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE projects SET
		    name = COALESCE($3, name),
		    niche = COALESCE($4, niche),
		    description = COALESCE($5, description),
		    language = COALESCE($6, language),
		    brand_json = COALESCE($7::jsonb, brand_json),
		    updated_at = now()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		id, userID, name, niche, description, language, nullableJSON(brandJSON))
	if err != nil {
		return Project{}, err
	}
	if tag.RowsAffected() == 0 {
		return Project{}, httpx.ErrNotFound
	}
	return r.Get(ctx, userID, id)
}

func nullableJSON(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}

func (r *Repo) SoftDelete(ctx context.Context, userID, id string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE projects SET deleted_at = now() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpx.ErrNotFound
	}
	return nil
}

// SaveContext appends a new context version rather than overwriting, so older
// carousels keep the brief they were generated against (§61).
func (r *Repo) SaveContext(ctx context.Context, projectID string, c Context) (int, error) {
	body, err := json.Marshal(c)
	if err != nil {
		return 0, err
	}
	var version int
	err = r.db.QueryRow(ctx, `
		INSERT INTO project_ai_contexts (project_id, version, context_json)
		VALUES ($1, COALESCE((SELECT max(version) FROM project_ai_contexts WHERE project_id=$1), 0) + 1, $2)
		RETURNING version`, projectID, body).Scan(&version)
	return version, err
}

func (r *Repo) LoadContext(ctx context.Context, projectID string) (Context, int, error) {
	var body []byte
	var version int
	err := r.db.QueryRow(ctx,
		`SELECT context_json, version FROM project_ai_contexts WHERE project_id=$1 ORDER BY version DESC LIMIT 1`,
		projectID).Scan(&body, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return Context{}, 0, nil
	}
	if err != nil {
		return Context{}, 0, err
	}
	var c Context
	err = json.Unmarshal(body, &c)
	return c, version, err
}

// ---- carousel input schema versions (§62) ----

type SchemaRecord struct {
	Version int             `json:"version"`
	Form    json.RawMessage `json:"form"`
	Raw     json.RawMessage `json:"-"`
}

func (r *Repo) SaveSchema(ctx context.Context, projectID string, form, raw, ui []byte, contextVersion int) (int, error) {
	stored, _ := json.Marshal(map[string]json.RawMessage{
		"form": form, "schema": raw, "ui": ui,
	})
	var version int
	err := r.db.QueryRow(ctx, `
		INSERT INTO project_schemas (project_id, version, schema_json, ui_json, context_version)
		VALUES ($1, COALESCE((SELECT max(version) FROM project_schemas WHERE project_id=$1), 0) + 1, $2, $3, $4)
		RETURNING version`, projectID, stored, ui, contextVersion).Scan(&version)
	return version, err
}

// LatestSchema returns the newest compiled form for a project, if any.
func (r *Repo) LatestSchema(ctx context.Context, projectID string) (SchemaRecord, bool, error) {
	var body []byte
	var rec SchemaRecord
	err := r.db.QueryRow(ctx,
		`SELECT version, schema_json FROM project_schemas WHERE project_id=$1 ORDER BY version DESC LIMIT 1`,
		projectID).Scan(&rec.Version, &body)
	if errors.Is(err, pgx.ErrNoRows) {
		return rec, false, nil
	}
	if err != nil {
		return rec, false, err
	}
	var stored struct {
		Form json.RawMessage `json:"form"`
	}
	if err := json.Unmarshal(body, &stored); err != nil {
		return rec, false, err
	}
	rec.Form = stored.Form
	return rec, true, nil
}

func (r *Repo) SchemaByVersion(ctx context.Context, projectID string, version int) (SchemaRecord, error) {
	var body []byte
	var rec SchemaRecord
	err := r.db.QueryRow(ctx,
		`SELECT version, schema_json FROM project_schemas WHERE project_id=$1 AND version=$2`,
		projectID, version).Scan(&rec.Version, &body)
	if errors.Is(err, pgx.ErrNoRows) {
		return rec, httpx.ErrNotFound
	}
	if err != nil {
		return rec, err
	}
	var stored struct {
		Form json.RawMessage `json:"form"`
	}
	err = json.Unmarshal(body, &stored)
	rec.Form = stored.Form
	return rec, err
}

// SaveResearchSources records where project/carousel research came from (§68).
func (r *Repo) SaveResearchSources(ctx context.Context, projectID, carouselID string, sources []Source) error {
	for _, s := range sources {
		var pid, cid any
		if projectID != "" {
			pid = projectID
		}
		if carouselID != "" {
			cid = carouselID
		}
		if _, err := r.db.Exec(ctx, `
			INSERT INTO research_sources (project_id, carousel_id, url, title, domain, source_type)
			VALUES ($1,$2,$3,$4,$5,$6)`, pid, cid, s.URL, s.Title, s.Domain, s.SourceType); err != nil {
			return err
		}
	}
	return nil
}

type Source struct {
	URL        string `json:"url"`
	Title      string `json:"title"`
	Domain     string `json:"domain"`
	SourceType string `json:"source_type"`
}

func (r *Repo) SourcesForCarousel(ctx context.Context, carouselID string) ([]Source, error) {
	rows, err := r.db.Query(ctx,
		`SELECT url, title, domain, source_type FROM research_sources WHERE carousel_id=$1 ORDER BY created_at`, carouselID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Source{}
	for rows.Next() {
		var s Source
		if err := rows.Scan(&s.URL, &s.Title, &s.Domain, &s.SourceType); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// TrackEvent records a product analytics event (§115).
func (r *Repo) TrackEvent(ctx context.Context, userID, projectID, carouselID, event string, metadata any) {
	body, _ := json.Marshal(metadata)
	var uid, pid, cid any
	if userID != "" {
		uid = userID
	}
	if projectID != "" {
		pid = projectID
	}
	if carouselID != "" {
		cid = carouselID
	}
	_, _ = r.db.Exec(context.WithoutCancel(ctx),
		`INSERT INTO usage_events (user_id, project_id, carousel_id, event, metadata_json) VALUES ($1,$2,$3,$4,$5)`,
		uid, pid, cid, event, body)
}
