package carousel

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tks/backend/internal/auth"
	"github.com/tks/backend/internal/design"
	"github.com/tks/backend/internal/httpx"
	"github.com/tks/backend/internal/job"
)

type Handler struct {
	svc  *Service
	jobs *job.Repo
}

func NewHandler(svc *Service, jobs *job.Repo) *Handler { return &Handler{svc: svc, jobs: jobs} }

func (h *Handler) Routes(r chi.Router) {
	r.Get("/projects/{projectID}/carousels", h.list)
	r.Post("/projects/{projectID}/carousels/generate", h.generate)

	r.Get("/carousels/{carouselID}", h.get)
	r.Patch("/carousels/{carouselID}", h.update)
	r.Delete("/carousels/{carouselID}", h.remove)
	r.Post("/carousels/{carouselID}/duplicate", h.duplicate)
	r.Post("/carousels/{carouselID}/archive", h.archive)
	r.Post("/carousels/{carouselID}/regenerate", h.regenerate)

	r.Get("/carousels/{carouselID}/content", h.content)
	r.Get("/carousels/{carouselID}/sources", h.sources)

	r.Get("/carousels/{carouselID}/caption", h.getCaption)
	r.Patch("/carousels/{carouselID}/caption", h.patchCaption)
	r.Post("/carousels/{carouselID}/caption/generate", h.regenerateCaption)
	r.Post("/carousels/{carouselID}/apply-brand", h.applyBrand)

	// Design API (§79)
	r.Get("/carousels/{carouselID}/design", h.getDesign)
	r.Patch("/carousels/{carouselID}/design", h.patchDesign)

	// Job API (§76)
	r.Get("/jobs/{jobID}", h.getJob)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.repo.List(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "projectID"),
		ListFilter{Status: r.URL.Query().Get("status"), Query: r.URL.Query().Get("q")})
	if err != nil {
		httpx.Fail(w, r, httpx.Internal(err))
		return
	}
	if err := h.svc.hydrateThumbnails(r.Context(), items); err != nil {
		httpx.Fail(w, r, httpx.Internal(err))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"carousels": items})
}

func (h *Handler) generate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Inputs map[string]any `json:"inputs"`
		Ratio  string         `json:"ratio"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	res, err := h.svc.Generate(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "projectID"),
		body.Inputs, body.Ratio, r.Header.Get("Idempotency-Key"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusAccepted, map[string]any{
		"carousel_id": res.Carousel.ID, "carousel": res.Carousel,
		"job_id": res.JobID, "status": res.Status,
	})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserID(r.Context())
	c, err := h.svc.repo.Get(r.Context(), userID, chi.URLParam(r, "carouselID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	one := []Carousel{c}
	if err := h.svc.hydrateThumbnails(r.Context(), one); err != nil {
		httpx.Fail(w, r, httpx.Internal(err))
		return
	}
	out := map[string]any{"carousel": one[0]}
	if j, ok, err := h.jobs.LatestForCarousel(r.Context(), userID, c.ID); err == nil && ok {
		out["job"] = map[string]any{
			"id": j.ID, "status": j.Status, "progress": j.Progress,
			"current_step": j.CurrentStep, "error_message": j.ErrorMessage, "steps": j.Steps(),
		}
	}
	httpx.JSON(w, http.StatusOK, out)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title  *string `json:"title"`
		Status *string `json:"status"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if body.Status != nil && !validStatus(*body.Status) {
		httpx.Fail(w, r, httpx.BadRequest("Trạng thái không hợp lệ."))
		return
	}
	c, err := h.svc.repo.UpdateMeta(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "carouselID"), body.Title, body.Status)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, c)
}

func validStatus(s string) bool {
	switch s {
	case StatusDraft, StatusReady, StatusArchived:
		return true
	}
	return false
}

func (h *Handler) remove(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.repo.SoftDelete(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "carouselID")); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.NoContent(w)
}

func (h *Handler) duplicate(w http.ResponseWriter, r *http.Request) {
	c, err := h.svc.Duplicate(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "carouselID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, c)
}

func (h *Handler) archive(w http.ResponseWriter, r *http.Request) {
	status := StatusArchived
	c, err := h.svc.repo.UpdateMeta(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "carouselID"), nil, &status)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, c)
}

func (h *Handler) regenerate(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.Regenerate(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "carouselID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusAccepted, map[string]any{
		"carousel_id": res.Carousel.ID, "job_id": res.JobID, "status": res.Status,
	})
}

func (h *Handler) content(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserID(r.Context())
	c, err := h.svc.repo.Get(r.Context(), userID, chi.URLParam(r, "carouselID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	body, err := h.svc.repo.LatestContent(r.Context(), c.ID)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, body)
}

func (h *Handler) sources(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserID(r.Context())
	c, err := h.svc.repo.Get(r.Context(), userID, chi.URLParam(r, "carouselID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	sources, err := h.svc.projects.Repo().SourcesForCarousel(r.Context(), c.ID)
	if err != nil {
		httpx.Fail(w, r, httpx.Internal(err))
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"sources": sources})
}

func (h *Handler) getDesign(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.LoadDesign(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "carouselID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, view)
}

func (h *Handler) patchDesign(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Version int           `json:"version"`
		Design  design.Design `json:"design"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	view, err := h.svc.SaveDesign(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "carouselID"), body.Version, body.Design)
	if err == httpx.ErrConflict {
		// 409 carries the server's current version so the editor can recover.
		httpx.JSON(w, http.StatusConflict, map[string]any{
			"error":   map[string]string{"code": "conflict", "message": httpx.ErrConflict.Message},
			"current": view,
		})
		return
	}
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, view)
}

func (h *Handler) getCaption(w http.ResponseWriter, r *http.Request) {
	caption, err := h.svc.Caption(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "carouselID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, caption)
}

func (h *Handler) patchCaption(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Caption  string   `json:"caption"`
		Hashtags []string `json:"hashtags"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	caption, err := h.svc.UpdateCaption(r.Context(), auth.MustUserID(r.Context()),
		chi.URLParam(r, "carouselID"), body.Caption, body.Hashtags)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, caption)
}

func (h *Handler) regenerateCaption(w http.ResponseWriter, r *http.Request) {
	caption, err := h.svc.RegenerateCaption(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "carouselID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, caption)
}

func (h *Handler) applyBrand(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.ApplyBrand(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "carouselID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, view)
}

func (h *Handler) getJob(w http.ResponseWriter, r *http.Request) {
	j, err := h.jobs.Get(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "jobID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"id": j.ID, "carousel_id": j.CarouselID, "status": j.Status, "progress": j.Progress,
		"current_step": j.CurrentStep, "error_message": j.ErrorMessage, "steps": j.Steps(),
	})
}
