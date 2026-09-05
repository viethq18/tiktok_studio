// Package ai turns AI calls into typed, validated, audited operations.
// Nothing outside this package ever sees a raw model response (§2.4, §57).
package ai

import "context"

type TaskType string

const (
	TaskProjectAnalysis  TaskType = "project_analysis"
	TaskCarouselSchema   TaskType = "carousel_schema"
	TaskResearch         TaskType = "research"
	TaskContent          TaskType = "content_generation"
	TaskFormulaSelection TaskType = "formula_selection"
	TaskDesign           TaskType = "design_generation"
	TaskImageIntent      TaskType = "image_intent"
	TaskCaption          TaskType = "caption_generation"
)

type Request struct {
	Task      TaskType
	System    string
	User      string
	MaxTokens int
	// Vars carries the structured inputs behind User. The real provider ignores
	// it; the offline mock provider uses it to synthesise a plausible answer.
	Vars map[string]any
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type Response struct {
	Text  string
	Usage Usage
}

// Provider is the seam between the domain and any OpenAI-compatible endpoint (§54).
type Provider interface {
	Name() string
	Model() string
	Complete(ctx context.Context, req Request) (Response, error)
}
