package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/tks/backend/internal/config"
	"github.com/tks/backend/internal/httpx"
)

type Service struct {
	repo *Repo
	cfg  config.Config
}

func NewService(repo *Repo, cfg config.Config) *Service { return &Service{repo: repo, cfg: cfg} }

func newToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// SignIn upserts the identity, opens a server-side session and sets the cookie.
func (s *Service) SignIn(ctx context.Context, w http.ResponseWriter, provider, providerUserID, email, name, avatar string) (User, error) {
	user, err := s.repo.UpsertUser(ctx, provider, providerUserID, email, name, avatar)
	if err != nil {
		return User{}, httpx.Wrap(http.StatusInternalServerError, "auth_failed", "Không thể đăng nhập.", err)
	}
	sid := newToken()
	expires := time.Now().Add(s.cfg.SessionTTL)
	if err := s.repo.CreateSession(ctx, sid, user.ID, expires); err != nil {
		return User{}, httpx.Wrap(http.StatusInternalServerError, "auth_failed", "Không thể đăng nhập.", err)
	}
	s.setCookie(w, sid, expires)
	return user, nil
}

func (s *Service) setCookie(w http.ResponseWriter, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   strings.HasPrefix(s.cfg.PublicAPIURL, "https://"),
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Service) SignOut(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(s.cfg.SessionCookieName); err == nil {
		_ = s.repo.DeleteSession(ctx, c.Value)
	}
	s.setCookie(w, "", time.Unix(0, 0))
}

func (s *Service) UserFromRequest(r *http.Request) (User, error) {
	c, err := r.Cookie(s.cfg.SessionCookieName)
	if err != nil || c.Value == "" {
		return User{}, httpx.ErrUnauthorized
	}
	user, err := s.repo.UserBySession(r.Context(), c.Value)
	if err != nil {
		return User{}, httpx.ErrUnauthorized
	}
	return user, nil
}
