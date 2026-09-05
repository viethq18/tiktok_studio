package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/tks/backend/internal/config"
	"github.com/tks/backend/internal/httpx"
)

type Handler struct {
	svc *Service
	cfg config.Config
}

func NewHandler(svc *Service, cfg config.Config) *Handler { return &Handler{svc: svc, cfg: cfg} }

func (h *Handler) Routes(r chi.Router) {
	r.Get("/auth/config", h.config)
	r.Get("/auth/google", h.googleStart)
	r.Get("/auth/callback", h.googleCallback)
	r.Post("/auth/dev-login", h.devLogin)
	r.Post("/auth/logout", h.logout)
	r.With(h.svc.Required).Get("/auth/me", h.me)
}

func (h *Handler) config(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]any{
		"google_enabled":   h.cfg.GoogleEnabled(),
		"dev_login_enabled": h.cfg.AllowDevLogin,
	})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user, _ := CurrentUser(r.Context())
	httpx.JSON(w, http.StatusOK, user)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	h.svc.SignOut(r.Context(), w, r)
	httpx.NoContent(w)
}

// devLogin exists only when ALLOW_DEV_LOGIN=true (§105).
func (h *Handler) devLogin(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.AllowDevLogin {
		httpx.Fail(w, r, httpx.ErrForbidden)
		return
	}
	var body struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if _, err := mail.ParseAddress(body.Email); err != nil {
		httpx.Fail(w, r, httpx.BadRequest("Email không hợp lệ."))
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name, _, _ = strings.Cut(body.Email, "@")
	}
	user, err := h.svc.SignIn(r.Context(), w, "dev", body.Email, body.Email, name, "")
	if err != nil {
		httpx.Fail(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

func (h *Handler) oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.cfg.GoogleClientID,
		ClientSecret: h.cfg.GoogleClientSecret,
		RedirectURL:  h.cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

const oauthStateCookie = "tks_oauth_state"

func (h *Handler) googleStart(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.GoogleEnabled() {
		httpx.Fail(w, r, httpx.BadRequest("Google login chưa được cấu hình."))
		return
	}
	state := newToken()
	http.SetCookie(w, &http.Cookie{
		Name: oauthStateCookie, Value: state, Path: "/", HttpOnly: true,
		Expires: time.Now().Add(10 * time.Minute), SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, h.oauthConfig().AuthCodeURL(state), http.StatusFound)
}

func (h *Handler) googleCallback(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.GoogleEnabled() {
		httpx.Fail(w, r, httpx.BadRequest("Google login chưa được cấu hình."))
		return
	}
	// CSRF: the state in the URL must match the one we planted in the cookie.
	c, err := r.Cookie(oauthStateCookie)
	if err != nil || c.Value == "" || c.Value != r.URL.Query().Get("state") {
		httpx.Fail(w, r, httpx.BadRequest("Phiên đăng nhập không hợp lệ."))
		return
	}
	http.SetCookie(w, &http.Cookie{Name: oauthStateCookie, Value: "", Path: "/", Expires: time.Unix(0, 0)})

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	token, err := h.oauthConfig().Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		httpx.Fail(w, r, httpx.Wrap(http.StatusBadRequest, "oauth_failed", "Đăng nhập Google thất bại.", err))
		return
	}
	profile, err := fetchGoogleProfile(ctx, h.oauthConfig().Client(ctx, token))
	if err != nil {
		httpx.Fail(w, r, httpx.Wrap(http.StatusBadRequest, "oauth_failed", "Đăng nhập Google thất bại.", err))
		return
	}
	if _, err := h.svc.SignIn(r.Context(), w, "google", profile.Sub, profile.Email, profile.Name, profile.Picture); err != nil {
		httpx.Fail(w, r, err)
		return
	}
	http.Redirect(w, r, h.cfg.PublicAppURL+"/projects", http.StatusFound)
}

type googleProfile struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func fetchGoogleProfile(ctx context.Context, client *http.Client) (googleProfile, error) {
	var p googleProfile
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return p, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return p, err
	}
	defer resp.Body.Close()
	return p, json.NewDecoder(resp.Body).Decode(&p)
}
