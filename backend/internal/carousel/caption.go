package carousel

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/tks/backend/internal/httpx"
)

// Caption is the TikTok post copy that ships with the carousel. The AI drafts
// it; the creator owns the final text, so `Edited` records who wrote what and a
// regeneration never silently discards a hand-written caption.
type Caption struct {
	Caption     string     `json:"caption"`
	Hashtags    []string   `json:"hashtags"`
	Edited      bool       `json:"edited"`
	GeneratedAt *time.Time `json:"generated_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

const MaxCaptionChars = 2200 // TikTok's limit

func (r *Repo) LoadCaption(ctx context.Context, carouselID string) (Caption, error) {
	var c Caption
	err := r.db.QueryRow(ctx, `
		SELECT caption, hashtags, edited, generated_at, updated_at
		FROM carousel_captions WHERE carousel_id = $1`, carouselID).
		Scan(&c.Caption, &c.Hashtags, &c.Edited, &c.GeneratedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, httpx.ErrNotFound
	}
	return c, err
}

// SaveGeneratedCaption stores an AI draft. A caption the creator already edited
// is left alone unless force is set.
func (r *Repo) SaveGeneratedCaption(ctx context.Context, carouselID, caption string, hashtags []string, force bool) error {
	if force {
		_, err := r.db.Exec(ctx, `
			INSERT INTO carousel_captions (carousel_id, caption, hashtags, edited, generated_at, updated_at)
			VALUES ($1,$2,$3,false,now(),now())
			ON CONFLICT (carousel_id) DO UPDATE
			SET caption = EXCLUDED.caption, hashtags = EXCLUDED.hashtags,
			    edited = false, generated_at = now(), updated_at = now()`,
			carouselID, caption, hashtags)
		return err
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO carousel_captions (carousel_id, caption, hashtags, generated_at, updated_at)
		VALUES ($1,$2,$3,now(),now())
		ON CONFLICT (carousel_id) DO UPDATE
		SET caption = EXCLUDED.caption, hashtags = EXCLUDED.hashtags,
		    generated_at = now(), updated_at = now()
		WHERE NOT carousel_captions.edited`,
		carouselID, caption, hashtags)
	return err
}

func (r *Repo) SaveEditedCaption(ctx context.Context, carouselID, caption string, hashtags []string) (Caption, error) {
	caption = strings.TrimSpace(caption)
	if len([]rune(caption)) > MaxCaptionChars {
		return Caption{}, httpx.BadRequest("Caption vượt quá 2200 ký tự — giới hạn của TikTok.")
	}
	if _, err := r.db.Exec(ctx, `
		INSERT INTO carousel_captions (carousel_id, caption, hashtags, edited, updated_at)
		VALUES ($1,$2,$3,true,now())
		ON CONFLICT (carousel_id) DO UPDATE
		SET caption = EXCLUDED.caption, hashtags = EXCLUDED.hashtags,
		    edited = true, updated_at = now()`,
		carouselID, caption, hashtags); err != nil {
		return Caption{}, err
	}
	return r.LoadCaption(ctx, carouselID)
}
