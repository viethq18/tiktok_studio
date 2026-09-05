// Package schema validates AI-generated form schemas and the user input filled
// against them. AI proposes; this package decides (§11, §145).
package schema

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Component registry — the complete set the frontend can render (§9.2 MVP).
// A schema referencing anything else is rejected outright.
var allowedComponents = map[string]bool{
	"text": true, "textarea": true, "number": true, "select": true,
	"multi_select": true, "radio": true, "checkbox": true, "slider": true,
	"image": true, "url": true,
}

// Complexity caps (§11).
const (
	MaxProperties    = 12
	MaxEnumValues    = 12
	MaxStringLength  = 2000
	MaxTitleLength   = 120
	MaxRequiredCount = 12
	MaxRawSchemaSize = 64 << 10
)

var allowedTypes = map[string]bool{
	"string": true, "integer": true, "number": true, "boolean": true, "array": true,
}

// Property is the normalized, safe form of one AI-proposed field.
type Property struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	EnumLabels  map[string]string `json:"enum_labels,omitempty"`
	Minimum     *float64 `json:"minimum,omitempty"`
	Maximum     *float64 `json:"maximum,omitempty"`
	MaxLength   *int     `json:"max_length,omitempty"`
	Default     any      `json:"default,omitempty"`
	Component   string   `json:"component"`
	Placeholder string   `json:"placeholder,omitempty"`
	Help        string   `json:"help,omitempty"`
	Required    bool     `json:"required"`
	Items       string   `json:"items,omitempty"` // element type for arrays
}

// Form is what the frontend renders. It is derived from the AI schema but is
// itself a closed, typed contract — the frontend never sees raw AI output (§7).
type Form struct {
	Version    int        `json:"version"`
	Properties []Property `json:"properties"`
}

type rawSchema struct {
	Type       string                     `json:"type"`
	Properties map[string]json.RawMessage `json:"properties"`
	Required   []string                   `json:"required"`
}

type rawProperty struct {
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Enum        []any    `json:"enum"`
	Minimum     *float64 `json:"minimum"`
	Maximum     *float64 `json:"maximum"`
	MaxLength   *int     `json:"maxLength"`
	Default     any      `json:"default"`
	Items       *struct {
		Type string `json:"type"`
		Enum []any  `json:"enum"`
	} `json:"items"`
	// Anything that would let the schema nest or reference is rejected.
	Ref        json.RawMessage `json:"$ref"`
	Properties json.RawMessage `json:"properties"`
	AllOf      json.RawMessage `json:"allOf"`
	AnyOf      json.RawMessage `json:"anyOf"`
	OneOf      json.RawMessage `json:"oneOf"`
}

type rawUI struct {
	Order  []string `json:"order"`
	Fields map[string]struct {
		Component   string            `json:"component"`
		Placeholder string            `json:"placeholder"`
		Help        string            `json:"help"`
		EnumLabels  map[string]string `json:"enum_labels"`
	} `json:"fields"`
}

// Compile turns an AI-proposed schema + ui hint into a safe Form, or fails.
func Compile(schemaJSON, uiJSON json.RawMessage) (Form, error) {
	if len(schemaJSON) > MaxRawSchemaSize {
		return Form{}, fmt.Errorf("schema is too large (%d bytes)", len(schemaJSON))
	}
	var rs rawSchema
	if err := json.Unmarshal(schemaJSON, &rs); err != nil {
		return Form{}, fmt.Errorf("schema is not valid JSON: %w", err)
	}
	if rs.Type != "object" {
		return Form{}, fmt.Errorf(`schema root type must be "object", got %q`, rs.Type)
	}
	if len(rs.Properties) == 0 {
		return Form{}, fmt.Errorf("schema has no properties")
	}
	if len(rs.Properties) > MaxProperties {
		return Form{}, fmt.Errorf("schema has %d properties, limit is %d", len(rs.Properties), MaxProperties)
	}
	if len(rs.Required) > MaxRequiredCount {
		return Form{}, fmt.Errorf("schema marks %d fields required, limit is %d", len(rs.Required), MaxRequiredCount)
	}

	var ui rawUI
	if len(uiJSON) > 0 {
		_ = json.Unmarshal(uiJSON, &ui) // ui hints are advisory; bad hints degrade, never fail
	}

	required := map[string]bool{}
	for _, name := range rs.Required {
		required[name] = true
	}

	props := make(map[string]Property, len(rs.Properties))
	for name, raw := range rs.Properties {
		p, err := compileProperty(name, raw, ui)
		if err != nil {
			return Form{}, err
		}
		p.Required = required[name]
		props[name] = p
	}

	// Ordering: honour the AI's ui.order for known fields, then the rest by name.
	ordered := make([]Property, 0, len(props))
	seen := map[string]bool{}
	for _, name := range ui.Order {
		if p, ok := props[name]; ok && !seen[name] {
			ordered = append(ordered, p)
			seen[name] = true
		}
	}
	rest := make([]string, 0, len(props))
	for name := range props {
		if !seen[name] {
			rest = append(rest, name)
		}
	}
	sort.Strings(rest)
	for _, name := range rest {
		ordered = append(ordered, props[name])
	}
	return Form{Version: 1, Properties: ordered}, nil
}

func compileProperty(name string, raw json.RawMessage, ui rawUI) (Property, error) {
	if !validFieldName(name) {
		return Property{}, fmt.Errorf("field name %q is not a safe identifier", name)
	}
	var rp rawProperty
	if err := json.Unmarshal(raw, &rp); err != nil {
		return Property{}, fmt.Errorf("field %q is not a valid schema: %w", name, err)
	}
	// Depth limit: a property may not itself be an object or a reference (§11).
	if len(rp.Ref) > 0 || len(rp.Properties) > 0 || len(rp.AllOf) > 0 || len(rp.AnyOf) > 0 || len(rp.OneOf) > 0 {
		return Property{}, fmt.Errorf("field %q uses nesting or references, which is not allowed", name)
	}
	if !allowedTypes[rp.Type] {
		return Property{}, fmt.Errorf("field %q has unsupported type %q", name, rp.Type)
	}

	p := Property{
		Name:        name,
		Type:        rp.Type,
		Title:       clip(firstNonEmpty(rp.Title, humanize(name)), MaxTitleLength),
		Description: clip(rp.Description, MaxTitleLength*3),
		Minimum:     rp.Minimum,
		Maximum:     rp.Maximum,
		MaxLength:   rp.MaxLength,
		Default:     rp.Default,
	}

	enumSource := rp.Enum
	if rp.Type == "array" {
		if rp.Items == nil || rp.Items.Type != "string" {
			return Property{}, fmt.Errorf("field %q must be an array of strings", name)
		}
		p.Items = "string"
		enumSource = rp.Items.Enum
	}
	for _, v := range enumSource {
		s, ok := v.(string)
		if !ok {
			s = fmt.Sprint(v)
		}
		p.Enum = append(p.Enum, clip(s, MaxTitleLength))
	}
	if len(p.Enum) > MaxEnumValues {
		return Property{}, fmt.Errorf("field %q has %d enum values, limit is %d", name, len(p.Enum), MaxEnumValues)
	}
	if p.MaxLength != nil && *p.MaxLength > MaxStringLength {
		v := MaxStringLength
		p.MaxLength = &v
	}
	if rp.Minimum != nil && rp.Maximum != nil && *rp.Minimum > *rp.Maximum {
		return Property{}, fmt.Errorf("field %q has minimum greater than maximum", name)
	}

	// Component: take the AI's hint only if it is in the registry, otherwise
	// derive one from the type. AI never widens the component surface (§11).
	hint := ui.Fields[name]
	if hint.Component != "" && allowedComponents[hint.Component] && componentFits(hint.Component, p) {
		p.Component = hint.Component
	} else {
		p.Component = deriveComponent(p)
	}
	p.Placeholder = clip(hint.Placeholder, MaxTitleLength)
	p.Help = clip(hint.Help, MaxTitleLength*2)
	if len(hint.EnumLabels) > 0 {
		p.EnumLabels = map[string]string{}
		for _, v := range p.Enum {
			if label, ok := hint.EnumLabels[v]; ok {
				p.EnumLabels[v] = clip(label, MaxTitleLength)
			}
		}
	}
	return p, nil
}

func componentFits(component string, p Property) bool {
	switch component {
	case "select", "radio":
		return len(p.Enum) > 0 && p.Type == "string"
	case "multi_select", "checkbox":
		return p.Type == "array" || p.Type == "boolean"
	case "slider", "number":
		return p.Type == "integer" || p.Type == "number"
	case "text", "textarea", "url", "image":
		return p.Type == "string"
	}
	return false
}

func deriveComponent(p Property) string {
	switch p.Type {
	case "boolean":
		return "checkbox"
	case "integer", "number":
		if p.Minimum != nil && p.Maximum != nil {
			return "slider"
		}
		return "number"
	case "array":
		return "multi_select"
	default:
		if len(p.Enum) > 0 {
			if len(p.Enum) <= 4 {
				return "radio"
			}
			return "select"
		}
		if p.MaxLength != nil && *p.MaxLength > 120 {
			return "textarea"
		}
		return "text"
	}
}

// validFieldName requires a leading letter, which also keeps names like
// "__proto__" out of the object the frontend builds from this form.
func validFieldName(name string) bool {
	if name == "" || len(name) > 40 {
		return false
	}
	for i, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case (r == '_' || (r >= '0' && r <= '9')) && i > 0:
		default:
			return false
		}
	}
	return true
}

func humanize(name string) string {
	parts := strings.Split(strings.ReplaceAll(name, "-", "_"), "_")
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
