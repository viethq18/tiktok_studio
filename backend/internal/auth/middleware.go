package auth

import (
	"context"
	"net/http"

	"github.com/tks/backend/internal/httpx"
)

type ctxKey string

const userKey ctxKey = "current_user"

// Required rejects anonymous requests before any handler runs (§86).
func (s *Service) Required(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := s.UserFromRequest(r)
		if err != nil {
			httpx.Fail(w, r, httpx.ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
	})
}

// CurrentUser returns the authenticated user placed by Required.
func CurrentUser(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey).(User)
	return u, ok
}

// MustUserID is safe inside handlers mounted behind Required.
func MustUserID(ctx context.Context) string {
	u, _ := CurrentUser(ctx)
	return u.ID
}
