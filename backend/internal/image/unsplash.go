package image

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Unsplash implements Provider. Two API rules are non-negotiable and are
// implemented here rather than left to callers (§156):
//   1. hitting the photo's download_location endpoint before using an image;
//   2. carrying photographer attribution alongside the stored asset.
type Unsplash struct {
	accessKey string
	client    *http.Client
}

func NewUnsplash(accessKey string) *Unsplash {
	return &Unsplash{accessKey: accessKey, client: &http.Client{Timeout: 25 * time.Second}}
}

func (u *Unsplash) Name() string { return "unsplash" }

type unsplashPhoto struct {
	ID          string `json:"id"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Color       string `json:"color"`
	Description string `json:"description"`
	AltDesc     string `json:"alt_description"`
	URLs        struct {
		Raw     string `json:"raw"`
		Full    string `json:"full"`
		Regular string `json:"regular"`
		Small   string `json:"small"`
		Thumb   string `json:"thumb"`
	} `json:"urls"`
	Links struct {
		HTML             string `json:"html"`
		DownloadLocation string `json:"download_location"`
	} `json:"links"`
	User struct {
		Name  string `json:"name"`
		Links struct {
			HTML string `json:"html"`
		} `json:"links"`
	} `json:"user"`
}

func (u *Unsplash) Search(ctx context.Context, query string, page, perPage int) ([]Candidate, error) {
	q := url.Values{}
	q.Set("query", query)
	q.Set("page", fmt.Sprint(max(page, 1)))
	q.Set("per_page", fmt.Sprint(max(perPage, 3)))
	q.Set("orientation", "portrait")
	q.Set("content_filter", "high")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.unsplash.com/search/photos?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Client-ID "+u.accessKey)
	req.Header.Set("Accept-Version", "v1")

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("unsplash rate limit reached")
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unsplash search returned %d", resp.StatusCode)
	}
	var parsed struct {
		Results []unsplashPhoto `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	out := make([]Candidate, 0, len(parsed.Results))
	for _, p := range parsed.Results {
		desc := p.Description
		if desc == "" {
			desc = p.AltDesc
		}
		out = append(out, Candidate{
			ID: p.ID, Provider: "unsplash",
			ThumbURL: p.URLs.Small, PreviewURL: p.URLs.Regular,
			DownloadURL: p.URLs.Raw + "&w=1600&fit=max&fm=jpg&q=85",
			TriggerURL:  p.Links.DownloadLocation,
			Width:       p.Width, Height: p.Height, Color: p.Color, Description: desc,
			PhotographerName: p.User.Name, PhotographerURL: p.User.Links.HTML,
			SourceURL:        p.Links.HTML,
		})
	}
	return out, nil
}

func (u *Unsplash) Download(ctx context.Context, c Candidate) ([]byte, error) {
	// Required by the Unsplash API Guidelines before the image is used.
	if c.TriggerURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.TriggerURL, nil)
		if err == nil {
			req.Header.Set("Authorization", "Client-ID "+u.accessKey)
			if resp, err := u.client.Do(req); err == nil {
				_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
				resp.Body.Close()
			}
		}
	}
	return fetchBytes(ctx, u.client, c.DownloadURL)
}

func fetchBytes(ctx context.Context, client *http.Client, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("image download returned %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 25<<20))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
