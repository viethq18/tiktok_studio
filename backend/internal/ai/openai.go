package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAICompatible speaks the /chat/completions dialect, which covers OpenAI,
// Azure OpenAI, Together, Groq, OpenRouter, vLLM and friends.
type OpenAICompatible struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewOpenAICompatible(baseURL, apiKey, model string, timeout time.Duration) *OpenAICompatible {
	return &OpenAICompatible{
		baseURL: baseURL, apiKey: apiKey, model: model,
		client: &http.Client{Timeout: timeout},
	}
}

func (p *OpenAICompatible) Name() string  { return "openai_compatible" }
func (p *OpenAICompatible) Model() string { return p.model }

// jsonModeReminder satisfies the provider-side rule that a request using
// response_format=json_object must mention JSON in the messages. Prompts already
// say so, but some endpoints only inspect the user message — this guarantees it
// for any prompt, including ones added later.
const jsonModeReminder = "\n\nReturn the answer as a single JSON object."

func (p *OpenAICompatible) Complete(ctx context.Context, req Request) (Response, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4000
	}
	user := req.User
	if !strings.Contains(strings.ToLower(user), "json") {
		user += jsonModeReminder
	}
	payload := map[string]any{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": req.System},
			{"role": "user", "content": user},
		},
		"temperature":     0.7,
		"max_tokens":      maxTokens,
		"response_format": map[string]string{"type": "json_object"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("ai request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode >= 300 {
		return Response{}, fmt.Errorf("ai provider returned %d: %s", resp.StatusCode, truncate(string(raw), 400))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Response{}, fmt.Errorf("ai response decode: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return Response{}, fmt.Errorf("ai response had no choices")
	}
	return Response{
		Text:  parsed.Choices[0].Message.Content,
		Usage: Usage{InputTokens: parsed.Usage.PromptTokens, OutputTokens: parsed.Usage.CompletionTokens},
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
