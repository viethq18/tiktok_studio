package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Providers reject response_format=json_object unless the messages mention JSON,
// and some only inspect the user message. Every prompt must satisfy that.
func TestEveryUserPromptMentionsJSON(t *testing.T) {
	vars := map[string]any{
		"niche_input": "x", "language": "vi", "project_context": "x", "query": "x",
		"web_results": "", "inputs": "x", "research": "x", "cta": "x", "slide_count": 5,
		"disclaimer": "", "max_headline": 90, "max_body": 220, "repair": "",
		"formulas": "x", "content": "x", "slides": "x", "palette": "{}", "fonts": "Inter",
		"canvas_width": 1080, "canvas_height": 1350, "ratio": "4:5",
		"safe_x": 80, "safe_y": 80, "safe_width": 920, "safe_height": 1190,
		"template_style": "minimal", "formula": "listicle",
	}
	for task, prompt := range prompts {
		user, err := prompt.render(vars)
		if err != nil {
			t.Fatalf("%s: %v", task, err)
		}
		if !strings.Contains(strings.ToLower(user), "json") {
			t.Errorf("task %s: the user message never mentions JSON, which json_object mode requires", task)
		}
	}
}

func TestCompleteAddsJSONReminderWhenPromptOmitsIt(t *testing.T) {
	var got struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"}}],"usage":{}}`))
	}))
	defer srv.Close()

	p := NewOpenAICompatible(srv.URL, "test-key", "test-model", 5*time.Second)
	if _, err := p.Complete(context.Background(), Request{
		Task: "custom", System: "be helpful", User: "give me five ideas",
	}); err != nil {
		t.Fatal(err)
	}
	user := got.Messages[1].Content
	if !strings.Contains(strings.ToLower(user), "json") {
		t.Errorf("the provider must guarantee the user message mentions JSON, got %q", user)
	}
}
