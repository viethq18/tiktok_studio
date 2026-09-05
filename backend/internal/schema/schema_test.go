package schema

import (
	"encoding/json"
	"testing"
)

func compile(t *testing.T, schemaJSON, uiJSON string) (Form, error) {
	t.Helper()
	return Compile(json.RawMessage(schemaJSON), json.RawMessage(uiJSON))
}

func TestCompileRejectsUnsupportedComponent(t *testing.T) {
	// The AI asks for a component that is not in the registry; the derived
	// component must fall back rather than pass it through (§11).
	form, err := compile(t,
		`{"type":"object","properties":{"topic":{"type":"string","title":"Topic"}},"required":["topic"]}`,
		`{"fields":{"topic":{"component":"execute_javascript"}}}`)
	if err != nil {
		t.Fatal(err)
	}
	if form.Properties[0].Component == "execute_javascript" {
		t.Error("an unregistered component reached the frontend contract")
	}
}

func TestCompileRejectsNestedObjects(t *testing.T) {
	_, err := compile(t,
		`{"type":"object","properties":{"a":{"type":"object","properties":{"b":{"type":"string"}}}}}`, `{}`)
	if err == nil {
		t.Error("nested objects must be rejected (depth limit)")
	}
}

func TestCompileRejectsRefs(t *testing.T) {
	_, err := compile(t, `{"type":"object","properties":{"a":{"$ref":"#/definitions/x","type":"string"}}}`, `{}`)
	if err == nil {
		t.Error("$ref must be rejected")
	}
}

func TestCompileRejectsOversizedEnums(t *testing.T) {
	enum := `"a"`
	for i := 0; i < MaxEnumValues+2; i++ {
		enum += `,"v` + string(rune('a'+i)) + `"`
	}
	_, err := compile(t, `{"type":"object","properties":{"a":{"type":"string","enum":[`+enum+`]}}}`, `{}`)
	if err == nil {
		t.Error("an oversized enum must be rejected")
	}
}

func TestCompileRejectsUnsafeFieldNames(t *testing.T) {
	_, err := compile(t, `{"type":"object","properties":{"__proto__":{"type":"string"}}}`, `{}`)
	if err == nil {
		t.Error("unsafe field names must be rejected")
	}
}

func TestValidateInputEnforcesRequiredAndRanges(t *testing.T) {
	form, err := compile(t, `{
		"type":"object",
		"properties":{
			"topic":{"type":"string","title":"Topic","maxLength":300},
			"slide_count":{"type":"integer","title":"Slides","minimum":3,"maximum":7,"default":5},
			"cta":{"type":"string","title":"CTA","enum":["save","follow"]}
		},
		"required":["topic","slide_count","cta"]}`, `{}`)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := form.ValidateInput(map[string]any{"slide_count": 5, "cta": "save"}); err == nil {
		t.Error("a missing required field must be rejected")
	}
	if _, err := form.ValidateInput(map[string]any{"topic": "x", "slide_count": 99, "cta": "save"}); err == nil {
		t.Error("a value above maximum must be rejected")
	}
	if _, err := form.ValidateInput(map[string]any{"topic": "x", "slide_count": 5, "cta": "delete_everything"}); err == nil {
		t.Error("a value outside the enum must be rejected")
	}
	clean, err := form.ValidateInput(map[string]any{"topic": " x ", "slide_count": float64(5), "cta": "save"})
	if err != nil {
		t.Fatalf("valid input was rejected: %v", err)
	}
	if clean["topic"] != "x" {
		t.Errorf("input was not normalised: %q", clean["topic"])
	}
}

func TestFallbackSchemaCompiles(t *testing.T) {
	s, u := FallbackSchema(nil, "vi")
	form, err := Compile(s, u)
	if err != nil {
		t.Fatalf("the fallback form must always compile: %v", err)
	}
	if len(form.Properties) != 3 {
		t.Errorf("expected topic, slide_count and cta, got %d fields", len(form.Properties))
	}
}
