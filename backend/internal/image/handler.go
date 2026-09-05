package image

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/tks/backend/internal/auth"
	"github.com/tks/backend/internal/carousel"
	"github.com/tks/backend/internal/httpx"
	"github.com/tks/backend/internal/ratelimit"
)

type Handler struct {
	svc      *Service
	carousel *carousel.Repo
	limiter  *ratelimit.Limiter
}

func NewHandler(svc *Service, carousels *carousel.Repo, limiter *ratelimit.Limiter) *Handler {
	return &Handler{svc: svc, carousel: carousels, limiter: limiter}
}

// Routes mounts the image API (§77). The browser never calls the provider
// directly, so keys, quota and caching stay server-side (§34).
func (h *Handler) Routes(r chi.Router) {
	r.Get("/carousels/{carouselID}/slides/{slideID}/images", h.candidates)
	r.Post("/carousels/{carouselID}/slides/{slideID}/images/search", h.searchMore)
	r.Post("/carousels/{carouselID}/slides/{slideID}/images/select", h.selectImage)
}

// resolve loads the carousel with an ownership check and returns the slide's
// stored search query.
func (h *Handler) resolve(r *http.Request) (carousel.Carousel, string, error) {
	userID := auth.MustUserID(r.Context())
	c, err := h.carousel.Get(r.Context(), userID, chi.URLParam(r, "carouselID"))
	if err != nil {
		return c, "", err
	}
	slideID := chi.URLParam(r, "slideID")
	rec, err := h.carousel.LatestDesign(r.Context(), c.ID)
	if err != nil {
		return c, "", err
	}
	for _, s := range rec.Design.Slides {
		if s.ID == slideID {
			query := s.ImageQuery
			if query == "" {
				query = s.ImageIntent
			}
			return c, query, nil
		}
	}
	return c, "", httpx.ErrNotFound
}

func (h *Handler) candidates(w http.ResponseWriter, r *http.Request) {
	c, query, err := h.resolve(r)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	results, err := h.svc.Search(r.Context(), c.ID, chi.URLParam(r, "slideID"), query, page)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"query": query, "candidates": PublicList(results)})
}

// searchMore is the "Research more" button (§33).
func (h *Handler) searchMore(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserID(r.Context())
	if err := h.limiter.Take(r.Context(), userID, ratelimit.ActionImageSearch); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	c, storedQuery, err := h.resolve(r)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var body struct {
		Query string `json:"query"`
		Page  int    `json:"page"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	query := body.Query
	if query == "" {
		query = storedQuery
	}
	if body.Page < 1 {
		body.Page = 2
	}
	results, err := h.svc.Search(r.Context(), c.ID, chi.URLParam(r, "slideID"), query, body.Page)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"query": query, "page": body.Page, "candidates": PublicList(results)})
}

func (h *Handler) selectImage(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserID(r.Context())
	c, err := h.carousel.Get(r.Context(), userID, chi.URLParam(r, "carouselID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	var body struct {
		CandidateID string `json:"candidate_id"`
		Version     int    `json:"version"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if body.CandidateID == "" {
		httpx.Fail(w, r, httpx.BadRequest("Ảnh được chọn không hợp lệ."))
		return
	}
	version, a, err := h.svc.SelectForSlide(r.Context(), userID, c.ProjectID, c.ID,
		chi.URLParam(r, "slideID"), body.CandidateID, body.Version)
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"version": version, "asset": a})
}
