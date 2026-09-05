// Package research separates *finding* evidence from *synthesising* it.
// The search provider is an interface so a real search API can be dropped in
// without touching the AI or worker code (§16, §35).
package research

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"bytes"
	"encoding/json"
	"net/http"
)

type Result struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	Domain  string `json:"domain"`
}

// SearchProvider fetches evidence for a query. Implementations must never
// follow user-supplied URLs — they only query a fixed search endpoint (§84).
type SearchProvider interface {
	Name() string
	Search(ctx context.Context, query string, limit int) ([]Result, error)
}

// NoopProvider is the default: no live search is configured, so the AI is told
// to rely on established knowledge and to return no sources rather than
// inventing them. This is an honest degradation, not a silent one.
type NoopProvider struct{}

func (NoopProvider) Name() string { return "none" }
func (NoopProvider) Search(context.Context, string, int) ([]Result, error) { return nil, nil }

// TavilyProvider is a thin client for a search API that returns clean snippets.
type TavilyProvider struct {
	apiKey string
	client *http.Client
}

func (p *TavilyProvider) Name() string { return "tavily" }

func (p *TavilyProvider) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	body, _ := json.Marshal(map[string]any{
		"api_key": p.apiKey, "query": query, "max_results": limit, "search_depth": "basic",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.tavily.com/search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("search provider returned %d", resp.StatusCode)
	}
	var parsed struct {
		Results []struct {
			URL     string `json:"url"`
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	out := make([]Result, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		out = append(out, Result{URL: r.URL, Title: r.Title, Snippet: r.Content, Domain: domainOf(r.URL)})
	}
	return out, nil
}

// NewProvider picks a provider from the environment. Search stays optional so
// the platform runs without a paid search key.
func NewProvider() SearchProvider {
	if key := os.Getenv("SEARCH_API_KEY"); key != "" {
		return &TavilyProvider{apiKey: key, client: &http.Client{Timeout: 20 * time.Second}}
	}
	return NoopProvider{}
}

// Format renders results for the research prompt.
func Format(results []Result) string {
	if len(results) == 0 {
		return ""
	}
	var b strings.Builder
	for i, r := range results {
		fmt.Fprintf(&b, "%d. %s (%s)\n   %s\n   %s\n", i+1, r.Title, r.Domain, r.URL, truncate(r.Snippet, 500))
	}
	return b.String()
}

func domainOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Host, "www.")
}

// Domain is exported for callers persisting AI-cited sources.
func Domain(raw string) string { return domainOf(raw) }

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
