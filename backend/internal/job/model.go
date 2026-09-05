// Package job models the async generation pipeline: its state machine (§14),
// its checkpoints (§158) and its Redis queue (§81).
package job

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusQueued            Status = "queued"
	StatusResearching       Status = "researching"
	StatusGeneratingContent Status = "generating_content"
	StatusValidatingContent Status = "validating_content"
	StatusSelectingFormula  Status = "selecting_formula"
	StatusGeneratingDesign  Status = "generating_design"
	StatusSearchingImages   Status = "searching_images"
	StatusDownloadingAssets Status = "downloading_assets"
	StatusFinalizing        Status = "finalizing"
	StatusCompleted         Status = "completed"
	StatusFailed            Status = "failed"
)

// Step describes one stage for the progress UI (§15). Keeping the labels here
// means the frontend renders whatever the backend reports rather than
// hard-coding its own copy of the pipeline.
type Step struct {
	Status   Status `json:"status"`
	Label    string `json:"label"`
	Progress int    `json:"progress"`
}

var Pipeline = []Step{
	{StatusResearching, "Đang nghiên cứu chủ đề", 12},
	{StatusGeneratingContent, "Đang viết nội dung", 34},
	{StatusValidatingContent, "Đang kiểm tra nội dung", 46},
	{StatusSelectingFormula, "Đang chọn công thức carousel", 54},
	{StatusGeneratingDesign, "Đang thiết kế slide", 72},
	{StatusSearchingImages, "Đang tìm hình ảnh", 86},
	{StatusFinalizing, "Đang hoàn thiện", 96},
}

func StepFor(s Status) Step {
	for _, step := range Pipeline {
		if step.Status == s {
			return step
		}
	}
	switch s {
	case StatusCompleted:
		return Step{StatusCompleted, "Hoàn tất", 100}
	case StatusFailed:
		return Step{StatusFailed, "Thất bại", 100}
	}
	return Step{StatusQueued, "Đang xếp hàng", 4}
}

type Job struct {
	ID                string          `json:"id"`
	UserID            string          `json:"user_id"`
	ProjectID         string          `json:"project_id"`
	CarouselID        string          `json:"carousel_id"`
	Type              string          `json:"type"`
	Status            Status          `json:"status"`
	Progress          int             `json:"progress"`
	CurrentStep       string          `json:"current_step"`
	LastCompletedStep string          `json:"last_completed_step"`
	StepOutputs       json.RawMessage `json:"-"`
	Attempt           int             `json:"attempt"`
	ErrorCode         string          `json:"error_code,omitempty"`
	ErrorMessage      string          `json:"error_message,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	CompletedAt       *time.Time      `json:"completed_at,omitempty"`
}

// Steps renders the checklist the progress screen draws (§15, §102).
func (j Job) Steps() []map[string]any {
	reached := map[Status]bool{}
	done := true
	for _, s := range Pipeline {
		if s.Status == j.Status {
			done = false
		}
		reached[s.Status] = done
	}
	out := make([]map[string]any, 0, len(Pipeline))
	for _, s := range Pipeline {
		state := "pending"
		switch {
		case j.Status == StatusCompleted:
			state = "done"
		case s.Status == j.Status:
			state = "active"
		case reached[s.Status]:
			state = "done"
		}
		out = append(out, map[string]any{"status": string(s.Status), "label": s.Label, "state": state})
	}
	return out
}

// Checkpoint is what a retry resumes from (§158).
type Checkpoint struct {
	ResearchJSON string `json:"research_json,omitempty"`
	ContentSaved bool   `json:"content_saved,omitempty"`
	FormulaID    string `json:"formula_id,omitempty"`
	DesignSaved  bool   `json:"design_saved,omitempty"`
	ImagesDone   bool   `json:"images_done,omitempty"`
}
