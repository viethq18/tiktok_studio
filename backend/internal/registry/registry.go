// Package registry exposes the internal font and formula tables. AI selects
// from these; it never invents entries (§40, §118).
package registry

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tks/backend/internal/design"
	"github.com/tks/backend/internal/language"
	"github.com/tks/backend/internal/httpx"
)

type Font struct {
	ID         string `json:"id"`
	Family     string `json:"family"`
	CSSStack   string `json:"css_stack"`
	Weights    []int  `json:"weights"`
	Vietnamese bool   `json:"vietnamese"`
}

type Formula struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Structure      []string `json:"structure"`
	RecommendedFor string   `json:"recommended_for"`
}

type Repo struct{ db *pgxpool.Pool }

func NewRepo(db *pgxpool.Pool) *Repo { return &Repo{db: db} }

func (r *Repo) Fonts(ctx context.Context) ([]Font, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, family, css_stack, weights, vietnamese FROM font_registry WHERE enabled ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Font{}
	for rows.Next() {
		var f Font
		if err := rows.Scan(&f.ID, &f.Family, &f.CSSStack, &f.Weights, &f.Vietnamese); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *Repo) Formulas(ctx context.Context) ([]Formula, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, structure, recommended_for FROM formula_registry WHERE enabled ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Formula{}
	for rows.Next() {
		var f Formula
		if err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.Structure, &f.RecommendedFor); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// FormulaPrompt renders the registry for the formula-selection prompt.
func FormulaPrompt(formulas []Formula) (string, map[string]bool) {
	allowed := map[string]bool{}
	var b strings.Builder
	for _, f := range formulas {
		allowed[f.ID] = true
		fmt.Fprintf(&b, "- %s (%s): %s | structure: %s | best for: %s\n",
			f.ID, f.Name, f.Description, strings.Join(f.Structure, " → "), f.RecommendedFor)
	}
	return b.String(), allowed
}

func (r *Repo) Formula(ctx context.Context, id string) (Formula, error) {
	var f Formula
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, structure, recommended_for FROM formula_registry WHERE id=$1`, id).
		Scan(&f.ID, &f.Name, &f.Description, &f.Structure, &f.RecommendedFor)
	return f, err
}

type Handler struct{ repo *Repo }

func NewHandler(repo *Repo) *Handler { return &Handler{repo: repo} }

func (h *Handler) Routes(r chi.Router) {
	r.Get("/registry", h.all)
}

func (h *Handler) all(w http.ResponseWriter, r *http.Request) {
	fonts, err := h.repo.Fonts(r.Context())
	if err != nil {
		httpx.Fail(w, r, httpx.Internal(err))
		return
	}
	formulas, err := h.repo.Formulas(r.Context())
	if err != nil {
		httpx.Fail(w, r, httpx.Internal(err))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"fonts":               fonts,
		"formulas":            formulas,
		"presets":             design.Presets,
		"languages":           language.Supported,
		"typography_defaults": design.TypographyDefaults,
	})
}
