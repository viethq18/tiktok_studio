package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ db *pgxpool.Pool }

func NewRepo(db *pgxpool.Pool) *Repo { return &Repo{db: db} }

// UpsertUser resolves an identity to a stable user row, keyed on
// (provider, provider_user_id) with email as the secondary identity (§59).
func (r *Repo) UpsertUser(ctx context.Context, provider, providerUserID, email, name, avatar string) (User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (email, name, avatar_url, provider, provider_user_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider, provider_user_id) DO UPDATE
		SET email = EXCLUDED.email,
		    name = CASE WHEN EXCLUDED.name <> '' THEN EXCLUDED.name ELSE users.name END,
		    avatar_url = CASE WHEN EXCLUDED.avatar_url <> '' THEN EXCLUDED.avatar_url ELSE users.avatar_url END,
		    updated_at = now()
		RETURNING id, email, name, avatar_url, provider, created_at`,
		email, name, avatar, provider, providerUserID).
		Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.CreatedAt)
	return u, err
}

func (r *Repo) CreateSession(ctx context.Context, id, userID string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)`, id, userID, expiresAt)
	return err
}

func (r *Repo) UserBySession(ctx context.Context, sessionID string) (User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		SELECT u.id, u.email, u.name, u.avatar_url, u.provider, u.created_at
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.id = $1 AND s.expires_at > now()`, sessionID).
		Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.Provider, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNoSession
	}
	if err == nil {
		// Best-effort last-seen bookkeeping; failure must not break the request.
		_, _ = r.db.Exec(ctx, `UPDATE sessions SET last_seen_at = now() WHERE id = $1`, sessionID)
	}
	return u, err
}

func (r *Repo) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, sessionID)
	return err
}

var ErrNoSession = errors.New("no active session")
