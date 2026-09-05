// Package content owns the Content JSON contract (§18) and the semantic quality
// gate that sits between "the AI returned valid JSON" and "we persist it" (§119).
package content

type Content struct {
	SchemaVersion string  `json:"schema_version"`
	Title         string  `json:"title"`
	FormulaID     string  `json:"formula_id"`
	FormulaReason string  `json:"formula_reason"`
	Slides        []Slide `json:"slides"`
}

type Slide struct {
	Index       int    `json:"index"`
	Role        string `json:"role"` // hook | point | cta (plus formula-specific roles)
	Headline    string `json:"headline"`
	Body        string `json:"body,omitempty"`
	ImageIntent string `json:"image_intent,omitempty"`
	ImageQuery  string `json:"image_query,omitempty"`
}
