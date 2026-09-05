// Package carousel owns the carousel lifecycle: inputs, content, design
// versions and status (§42, §43).
package carousel

import (
	"encoding/json"
	"time"
)

const (
	StatusDraft      = "draft"
	StatusGenerating = "generating"
	StatusReady      = "ready"
	StatusArchived   = "archived"
	StatusFailed     = "failed"
)

type Carousel struct {
	ID            string          `json:"id"`
	ProjectID     string          `json:"project_id"`
	Title         string          `json:"title"`
	Status        string          `json:"status"`
	Platform      string          `json:"platform"`
	CanvasRatio   string          `json:"canvas_ratio"`
	CanvasWidth   int             `json:"canvas_width"`
	CanvasHeight  int             `json:"canvas_height"`
	FormulaID     string          `json:"formula_id"`
	ThumbnailURL  string          `json:"thumbnail_url,omitempty"`
	SchemaVersion int             `json:"schema_version"`
	Input         json.RawMessage `json:"input,omitempty"`
	DesignVersion int             `json:"design_version"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`

	thumbnailKey string
}

func (c Carousel) ThumbnailKey() string { return c.thumbnailKey }
