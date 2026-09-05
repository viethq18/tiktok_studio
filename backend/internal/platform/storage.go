package platform

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/tks/backend/internal/config"
)

// Storage wraps MinIO. The bucket stays private (§85); the browser only ever
// receives presigned URLs.
type Storage struct {
	client         *minio.Client
	presignClient  *minio.Client
	bucket         string
}

func NewStorage(ctx context.Context, cfg config.Config) (*Storage, error) {
	creds := credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, "")
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{Creds: creds, Secure: cfg.MinioUseSSL})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}
	presign := client
	if cfg.MinioPublicEndpoint != cfg.MinioEndpoint {
		presign, err = minio.New(cfg.MinioPublicEndpoint, &minio.Options{Creds: creds, Secure: cfg.MinioUseSSL})
		if err != nil {
			return nil, fmt.Errorf("minio presign client: %w", err)
		}
	}
	exists, err := client.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio make bucket: %w", err)
		}
	}
	return &Storage{client: client, presignClient: presign, bucket: cfg.MinioBucket}, nil
}

func (s *Storage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
}

func (s *Storage) Remove(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

// SignedURL returns a time-limited download URL (§85).
func (s *Storage) SignedURL(ctx context.Context, key string, ttl time.Duration, downloadName string) (string, error) {
	if key == "" {
		return "", nil
	}
	params := url.Values{}
	if downloadName != "" {
		params.Set("response-content-disposition", `attachment; filename="`+downloadName+`"`)
	}
	u, err := s.presignClient.PresignedGetObject(ctx, s.bucket, key, ttl, params)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Key layout follows §37.
func AssetKey(userID, projectID, carouselID, assetID, ext string) string {
	return fmt.Sprintf("users/%s/projects/%s/carousels/%s/assets/%s%s", userID, projectID, carouselID, assetID, ext)
}

func ExportKey(userID, projectID, carouselID, exportID string) string {
	return fmt.Sprintf("users/%s/projects/%s/carousels/%s/exports/%s.zip", userID, projectID, carouselID, exportID)
}

func ThumbnailKey(userID, projectID, carouselID string) string {
	return fmt.Sprintf("users/%s/projects/%s/carousels/%s/thumbnail.png", userID, projectID, carouselID)
}
