package project

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tks/backend/internal/auth"
	"github.com/tks/backend/internal/httpx"
	"github.com/tks/backend/internal/language"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes(r chi.Router) {
	r.Get("/projects", h.list)
	r.Post("/projects", h.create)
	r.Get("/projects/{projectID}", h.get)
	r.Patch("/projects/{projectID}", h.update)
	r.Delete("/projects/{projectID}", h.remove)

	// Dynamic schema API (§74)
	r.Get("/projects/{projectID}/schema", h.getSchema)
	r.Post("/projects/{projectID}/generate-schema", h.regenerateSchema)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.repo.List(r.Context(), auth.MustUserID(r.Context()))
	if err != nil {
		httpx.Fail(w, r, httpx.Internal(err))
		return
	}
	for i := range items {
		h.svc.HydrateBrand(r.Context(), &items[i])
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"projects": items})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Niche    string `json:"niche"`
		Language string `json:"language"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	p, err := h.svc.CreateFromNiche(r.Context(), auth.MustUserID(r.Context()), body.Niche, body.Language)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, p)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.repo.Get(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "projectID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	h.svc.HydrateBrand(r.Context(), &p)
	httpx.JSON(w, http.StatusOK, p)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        *string          `json:"name"`
		Niche       *string          `json:"niche"`
		Description *string          `json:"description"`
		Language    *string          `json:"language"`
		Brand       *Brand           `json:"brand"`
		Context     *Context         `json:"context"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	userID := auth.MustUserID(r.Context())
	projectID := chi.URLParam(r, "projectID")

	if body.Language != nil {
		if !language.IsSupported(*body.Language) {
			httpx.Fail(w, r, httpx.BadRequest("Ngôn ngữ này chưa được hỗ trợ."))
			return
		}
		normalized := language.Normalize(*body.Language)
		body.Language = &normalized
	}
	if body.Brand != nil {
		// Normalise before persisting: display rules are a closed set, and a
		// handle should be stored with its "@" whatever the creator typed.
		body.Brand.Normalize()
	}
	p, err := h.svc.repo.Update(r.Context(), userID, projectID, body.Name, body.Niche, body.Description, body.Language, body.Brand)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	// Editing the AI strategy writes a new context version (§107).
	if body.Context != nil {
		version, err := h.svc.repo.SaveContext(r.Context(), projectID, *body.Context)
		if err != nil {
			httpx.Fail(w, r, httpx.Internal(err))
			return
		}
		p.ContextVersion = version
		p.Context, _ = json.Marshal(*body.Context)
	}
	h.svc.repo.TrackEvent(r.Context(), userID, projectID, "", "project_updated", nil)
	h.svc.HydrateBrand(r.Context(), &p)
	httpx.JSON(w, http.StatusOK, p)
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.repo.SoftDelete(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "projectID")); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.NoContent(w)
}

func (h *Handler) getSchema(w http.ResponseWriter, r *http.Request) {
	h.respondSchema(w, r, false)
}

func (h *Handler) regenerateSchema(w http.ResponseWriter, r *http.Request) {
	h.respondSchema(w, r, true)
}

func (h *Handler) respondSchema(w http.ResponseWriter, r *http.Request, force bool) {
	form, version, err := h.svc.EnsureSchema(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "projectID"), force)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"version": version, "form": form})
}
