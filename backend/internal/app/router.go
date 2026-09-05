package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tks/backend/internal/asset"
	"github.com/tks/backend/internal/httpx"
	"github.com/tks/backend/internal/ratelimit"
)

// Router builds the HTTP surface (§73). Everything under /api/v1 except auth
// and health sits behind the session middleware.
func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(httpx.WithRequestID, httpx.Recoverer, httpx.Logger, httpx.CORS(a.Cfg.PublicAppURL))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := a.DB.Ping(r.Context()); err != nil {
			httpx.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded"})
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		a.AuthH.Routes(r)

		r.Group(func(r chi.Router) {
			r.Use(a.Auth.Required)
			a.ProjectH.Routes(r)
			a.CarouselH.Routes(r)
			a.ImageH.Routes(r)
			a.ExportH.Routes(r)
			a.Registry.Routes(r)
			a.assetRoutes(r)
		})
	})
	return r
}

// assetRoutes covers the upload API (§78). Multipart is handled here rather
// than in the asset package so that package stays transport-agnostic.
func (a *App) assetRoutes(r chi.Router) {
	limiter := ratelimit.New(a.Redis)

	r.Post("/assets/upload", func(w http.ResponseWriter, r *http.Request) {
		userID := mustUserID(r)
		if err := limiter.Take(r.Context(), userID, ratelimit.ActionUpload); err != nil {
			httpx.Fail(w, r, err)
			return
		}
		if err := r.ParseMultipartForm(asset.MaxUploadBytes); err != nil {
			httpx.Fail(w, r, httpx.BadRequest("File tải lên không hợp lệ hoặc quá lớn."))
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			httpx.Fail(w, r, httpx.BadRequest("Thiếu file."))
			return
		}
		defer file.Close()

		data, err := readLimited(file, asset.MaxUploadBytes)
		if err != nil {
			httpx.Fail(w, r, httpx.BadRequest("File vượt quá 50MB."))
			return
		}
		out, err := a.Assets.Store(r.Context(), data, asset.StoreParams{
			UserID:     userID,
			ProjectID:  r.FormValue("project_id"),
			CarouselID: r.FormValue("carousel_id"),
			Source:     "upload",
		})
		if err != nil {
			httpx.Fail(w, r, err)
			return
		}
		httpx.JSON(w, http.StatusCreated, out)
	})

	r.Get("/assets/{assetID}", func(w http.ResponseWriter, r *http.Request) {
		out, err := a.Assets.Get(r.Context(), mustUserID(r), chi.URLParam(r, "assetID"))
		if err != nil {
			httpx.Fail(w, r, err)
			return
		}
		httpx.JSON(w, http.StatusOK, out)
	})

	r.Delete("/assets/{assetID}", func(w http.ResponseWriter, r *http.Request) {
		if err := a.Assets.Delete(r.Context(), mustUserID(r), chi.URLParam(r, "assetID")); err != nil {
			httpx.Fail(w, r, err)
			return
		}
		httpx.NoContent(w)
	})
}
