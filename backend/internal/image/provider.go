// Package image is the stock-photo layer. Domain code depends on the
// ImageProvider interface, never on Unsplash directly (§35), and the browser
// never talks to the provider itself (§34).
package image

import "context"

// Candidate is one search result. It is persisted in the search cache in full,
// including the URLs only the server may fetch (§84) — those are stripped by
// Public() before anything is sent to the browser.
type Candidate struct {
	ID               string `json:"id"`
	Provider         string `json:"provider"`
	ThumbURL         string `json:"thumb_url"`
	PreviewURL       string `json:"preview_url"`
	DownloadURL      string `json:"download_url"`
	TriggerURL       string `json:"trigger_url"` // §156 download-trigger endpoint
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	Color            string `json:"color,omitempty"`
	Description      string `json:"description,omitempty"`
	PhotographerName string `json:"photographer_name,omitempty"`
	PhotographerURL  string `json:"photographer_url,omitempty"`
	SourceURL        string `json:"source_url,omitempty"`
}

// PublicCandidate is what the browser sees. The client selects by id, so it
// never needs — and never gets — a URL the server would fetch on its behalf.
type PublicCandidate struct {
	ID               string `json:"id"`
	Provider         string `json:"provider"`
	ThumbURL         string `json:"thumb_url"`
	PreviewURL       string `json:"preview_url"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	Color            string `json:"color,omitempty"`
	Description      string `json:"description,omitempty"`
	PhotographerName string `json:"photographer_name,omitempty"`
	PhotographerURL  string `json:"photographer_url,omitempty"`
	SourceURL        string `json:"source_url,omitempty"`
}

func (c Candidate) Public() PublicCandidate {
	return PublicCandidate{
		ID: c.ID, Provider: c.Provider, ThumbURL: c.ThumbURL, PreviewURL: c.PreviewURL,
		Width: c.Width, Height: c.Height, Color: c.Color, Description: c.Description,
		PhotographerName: c.PhotographerName, PhotographerURL: c.PhotographerURL, SourceURL: c.SourceURL,
	}
}

func PublicList(in []Candidate) []PublicCandidate {
	out := make([]PublicCandidate, 0, len(in))
	for _, c := range in {
		out = append(out, c.Public())
	}
	return out
}

type Provider interface {
	Name() string
	Search(ctx context.Context, query string, page, perPage int) ([]Candidate, error)
	// Download fetches the image bytes. Implementations must honour any
	// provider-mandated download trigger before returning.
	Download(ctx context.Context, c Candidate) ([]byte, error)
}
