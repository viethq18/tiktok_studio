package image

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

// Picsum is the no-key fallback provider. It returns deterministic placeholder
// photography so the full image pipeline — candidates, selection, download,
// MinIO, background render, export — is exercisable without an Unsplash key.
// It is not a stand-in for Unsplash in production.
type Picsum struct{ client *http.Client }

func NewPicsum() *Picsum { return &Picsum{client: &http.Client{Timeout: 25 * time.Second}} }

func (p *Picsum) Name() string { return "picsum" }

func (p *Picsum) Search(ctx context.Context, query string, page, perPage int) ([]Candidate, error) {
	if perPage <= 0 {
		perPage = 3
	}
	out := make([]Candidate, 0, perPage)
	for i := 0; i < perPage; i++ {
		seed := seedFor(fmt.Sprintf("%s|%d|%d", query, page, i))
		out = append(out, Candidate{
			ID:          seed,
			Provider:    "picsum",
			ThumbURL:    fmt.Sprintf("https://picsum.photos/seed/%s/400/500", seed),
			PreviewURL:  fmt.Sprintf("https://picsum.photos/seed/%s/800/1000", seed),
			DownloadURL: fmt.Sprintf("https://picsum.photos/seed/%s/1080/1350", seed),
			Width:       1080, Height: 1350,
			Description:      query,
			PhotographerName: "Lorem Picsum",
			SourceURL:        "https://picsum.photos",
		})
	}
	return out, nil
}

func (p *Picsum) Download(ctx context.Context, c Candidate) ([]byte, error) {
	return fetchBytes(ctx, p.client, c.DownloadURL)
}

func seedFor(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])[:10]
}
