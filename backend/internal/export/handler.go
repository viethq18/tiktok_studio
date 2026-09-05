package export

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tks/backend/internal/auth"
	"github.com/tks/backend/internal/carousel"
	"github.com/tks/backend/internal/httpx"
	"github.com/tks/backend/internal/job"
	"github.com/tks/backend/internal/ratelimit"
)

type Handler struct {
	svc      *Service
	carousel *carousel.Repo
	queue    *job.Queue
	limiter  *ratelimit.Limiter
}

func NewHandler(svc *Service, carousels *carousel.Repo, queue *job.Queue, limiter *ratelimit.Limiter) *Handler {
	return &Handler{svc: svc, carousel: carousels, queue: queue, limiter: limiter}
}

func (h *Handler) Routes(r chi.Router) {
	r.Post("/carousels/{carouselID}/export", h.create)
	r.Get("/exports/{exportID}", h.get)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserID(r.Context())
	if err := h.limiter.Take(r.Context(), userID, ratelimit.ActionExport); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	c, err := h.carousel.Get(r.Context(), userID, chi.URLParam(r, "carouselID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	if _, err := h.carousel.LatestDesign(r.Context(), c.ID); err != nil {
		httpx.Fail(w, r, httpx.BadRequest("Carousel này chưa có thiết kế để export."))
		return
	}
	j, err := h.svc.Create(r.Context(), userID, c.ID)
	if err != nil {
		httpx.Fail(w, r, httpx.Internal(err))
		return
	}
	if err := h.queue.Enqueue(r.Context(), job.QueueExport, j.ID); err != nil {
		httpx.Fail(w, r, httpx.Wrap(503, "queue_unavailable", "Hệ thống đang bận. Vui lòng thử lại.", err))
		return
	}
	httpx.JSON(w, http.StatusAccepted, j)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	j, err := h.svc.Get(r.Context(), auth.MustUserID(r.Context()), chi.URLParam(r, "exportID"))
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, j)
}
