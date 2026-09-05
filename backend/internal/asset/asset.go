// Package asset owns binary media: validation, storage in MinIO and the
// database record that ties an object to a user, project and carousel (§67).
package asset

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tks/backend/internal/httpx"
	"github.com/tks/backend/internal/platform"
)

const (
	MaxUploadBytes = 50 << 20 // §163
	MinDimension   = 200
	MaxDimension   = 8000
	signedURLTTL   = 6 * time.Hour
)

type Asset struct {
	ID         string          `json:"id"`
	UserID     string          `json:"user_id"`
	ProjectID  string          `json:"project_id,omitempty"`
	CarouselID string          `json:"carousel_id,omitempty"`
	Source     string          `json:"source"`
	SourceID   string          `json:"source_id,omitempty"`
	StorageKey string          `json:"-"`
	MimeType   string          `json:"mime_type"`
	Width      int             `json:"width"`
	Height     int             `json:"height"`
	FileSize   int64           `json:"file_size"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	URL        string          `json:"url,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type Service struct {
	db      *pgxpool.Pool
	storage *platform.Storage
}

func NewService(db *pgxpool.Pool, storage *platform.Storage) *Service {
	return &Service{db: db, storage: storage}
}

// allowedMIME is checked against the sniffed bytes, never the client's
// extension or Content-Type header (§38).
var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type StoreParams struct {
	UserID     string
	ProjectID  string
	CarouselID string
	Source     string
	SourceID   string
	Metadata   any
}

// Store validates the bytes, uploads them and records the asset.
func (s *Service) Store(ctx context.Context, data []byte, p StoreParams) (Asset, error) {
	if len(data) == 0 {
		return Asset{}, httpx.BadRequest("File rỗng.")
	}
	if len(data) > MaxUploadBytes {
		return Asset{}, httpx.BadRequest("File vượt quá 50MB.")
	}
	cfg, format, err := image.DecodeConfig(newReader(data))
	if err != nil {
		return Asset{}, httpx.BadRequest("File không phải là ảnh hợp lệ.")
	}
	mime := "image/" + format
	if format == "jpeg" {
		mime = "image/jpeg"
	}
	ext, ok := allowedMIME[mime]
	if !ok {
		return Asset{}, httpx.BadRequest("Định dạng ảnh không được hỗ trợ.")
	}
	if cfg.Width < MinDimension || cfg.Height < MinDimension {
		return Asset{}, httpx.BadRequest(fmt.Sprintf("Ảnh quá nhỏ (tối thiểu %dpx mỗi cạnh).", MinDimension))
	}
	if cfg.Width > MaxDimension || cfg.Height > MaxDimension {
		return Asset{}, httpx.BadRequest("Ảnh quá lớn.")
	}

	id := newUUID()
	key := platform.AssetKey(p.UserID, orNone(p.ProjectID), orNone(p.CarouselID), id, ext)
	if err := s.storage.Put(ctx, key, newReader(data), int64(len(data)), mime); err != nil {
		return Asset{}, httpx.Wrap(500, "storage_error", "Không lưu được ảnh. Vui lòng thử lại.", err)
	}

	metadata, _ := json.Marshal(p.Metadata)
	if len(metadata) == 0 || string(metadata) == "null" {
		metadata = []byte("{}")
	}
	a := Asset{
		ID: id, UserID: p.UserID, ProjectID: p.ProjectID, CarouselID: p.CarouselID,
		Source: p.Source, SourceID: p.SourceID, StorageKey: key, MimeType: mime,
		Width: cfg.Width, Height: cfg.Height, FileSize: int64(len(data)), Metadata: metadata,
	}
	if err := s.db.QueryRow(ctx, `
		INSERT INTO assets (id, user_id, project_id, carousel_id, source, source_id, storage_key,
		                    mime_type, width, height, file_size, metadata_json)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING created_at`,
		a.ID, a.UserID, nullable(a.ProjectID), nullable(a.CarouselID), a.Source, a.SourceID, a.StorageKey,
		a.MimeType, a.Width, a.Height, a.FileSize, a.Metadata).Scan(&a.CreatedAt); err != nil {
		_ = s.storage.Remove(ctx, key)
		return Asset{}, httpx.Internal(err)
	}
	a.URL, _ = s.storage.SignedURL(ctx, key, signedURLTTL, "")
	return a, nil
}

func (s *Service) Get(ctx context.Context, userID, id string) (Asset, error) {
	var a Asset
	var projectID, carouselID *string
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, project_id, carousel_id, source, source_id, storage_key, mime_type,
		       width, height, file_size, metadata_json, created_at
		FROM assets WHERE id=$1 AND user_id=$2`, id, userID).
		Scan(&a.ID, &a.UserID, &projectID, &carouselID, &a.Source, &a.SourceID, &a.StorageKey,
			&a.MimeType, &a.Width, &a.Height, &a.FileSize, &a.Metadata, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, httpx.ErrNotFound
	}
	if err != nil {
		return a, err
	}
	a.ProjectID, a.CarouselID = deref(projectID), deref(carouselID)
	a.URL, _ = s.storage.SignedURL(ctx, a.StorageKey, signedURLTTL, "")
	return a, nil
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	var key string
	err := s.db.QueryRow(ctx, `DELETE FROM assets WHERE id=$1 AND user_id=$2 RETURNING storage_key`, id, userID).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.ErrNotFound
	}
	if err != nil {
		return err
	}
	return s.storage.Remove(ctx, key)
}

// SignedURLs resolves asset ids to presigned URLs in one round trip. The
// browser never sees a storage key and the bucket stays private (§85).
func (s *Service) SignedURLs(ctx context.Context, ids []string) (map[string]string, error) {
	out := map[string]string{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.db.Query(ctx, `SELECT id, storage_key FROM assets WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type pair struct{ id, key string }
	var pairs []pair
	for rows.Next() {
		var p pair
		if err := rows.Scan(&p.id, &p.key); err != nil {
			return nil, err
		}
		pairs = append(pairs, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, p := range pairs {
		if url, err := s.storage.SignedURL(ctx, p.key, signedURLTTL, ""); err == nil {
			out[p.id] = url
		}
	}
	return out, nil
}

// Attach links an asset to a slide so a carousel's media is queryable (§67).
func (s *Service) Attach(ctx context.Context, carouselID, assetID, slideID, role string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO carousel_assets (carousel_id, asset_id, slide_id, role) VALUES ($1,$2,$3,$4)
		ON CONFLICT DO NOTHING`, carouselID, assetID, slideID, role)
	return err
}

// Raw returns the stored bytes, used by the export renderer.
func (s *Service) Raw(ctx context.Context, assetID string) ([]byte, error) {
	var key string
	if err := s.db.QueryRow(ctx, `SELECT storage_key FROM assets WHERE id=$1`, assetID).Scan(&key); err != nil {
		return nil, err
	}
	obj, err := s.storage.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(io.LimitReader(obj, MaxUploadBytes))
}

func (s *Service) Storage() *platform.Storage { return s.storage }

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func orNone(s string) string {
	if s == "" {
		return "_none"
	}
	return s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
