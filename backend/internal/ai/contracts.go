package ai

import (
	"encoding/json"

	"github.com/tks/backend/internal/content"
	"github.com/tks/backend/internal/design"
)

// Each AI task has its own typed output contract (§57). No task returns a bare
// map — except the dynamic form schema, which has its own safety validator.

type ProjectIntelligence struct {
	SchemaVersion string `json:"schema_version"`
	Name          string `json:"name"`
	Identity      struct {
		Niche       string `json:"niche"`
		Language    string `json:"language"`
		Description string `json:"description"`
	} `json:"identity"`
	Audience struct {
		Description  string   `json:"description"`
		Demographics string   `json:"demographics"`
		PainPoints   []string `json:"pain_points"`
	} `json:"audience"`
	Tone            []string `json:"tone"`
	WritingStyle    string   `json:"writing_style"`
	PreferredAngles []string `json:"preferred_angles"`
	ContentPillars  []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"content_pillars"`
	CTAOptions []struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	} `json:"cta_options"`
	SuggestedTopics []string `json:"suggested_topics"`
}

// SchemaOutput is the one place raw JSON survives — guarded by schema.Validate.
type SchemaOutput struct {
	Schema json.RawMessage `json:"schema"`
	UI     json.RawMessage `json:"ui"`
}

type ResearchOutput struct {
	Summary  string   `json:"summary"`
	KeyFacts []string `json:"key_facts"`
	Angles   []string `json:"angles"`
	Cautions []string `json:"cautions"`
	Sources  []Source `json:"sources"`
}

type Source struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
	Domain string `json:"domain"`
}

type FormulaSelection struct {
	FormulaID string `json:"formula_id"`
	Reason    string `json:"reason"`
}

type ImageIntentOutput struct {
	Slides []struct {
		SlideID     string `json:"slide_id"`
		ImageQuery  string `json:"image_query"`
		ImageIntent string `json:"image_intent"`
	} `json:"slides"`
}

type CaptionOutput struct {
	Caption  string   `json:"caption"`
	Hashtags []string `json:"hashtags"`
}

// Aliases so callers read naturally.
type ContentOutput = content.Content
type DesignOutput = design.Design
