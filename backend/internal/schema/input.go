package schema

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ValidateInput checks user-submitted values against the compiled form.
// Frontend validation is for UX only; this is the decision (§96).
func (f Form) ValidateInput(input map[string]any) (map[string]any, error) {
	clean := make(map[string]any, len(f.Properties))
	var problems []string

	for _, p := range f.Properties {
		raw, present := input[p.Name]
		if !present || isEmpty(raw) {
			if p.Required {
				problems = append(problems, fmt.Sprintf("%s là bắt buộc", p.Title))
			} else if p.Default != nil {
				clean[p.Name] = p.Default
			}
			continue
		}
		value, err := p.coerce(raw)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		clean[p.Name] = value
	}
	if len(problems) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(problems, "; "))
	}
	return clean, nil
}

func (p Property) coerce(raw any) (any, error) {
	switch p.Type {
	case "string":
		s, ok := raw.(string)
		if !ok {
			return nil, fmt.Errorf("%s phải là văn bản", p.Title)
		}
		s = strings.TrimSpace(s)
		limit := MaxStringLength
		if p.MaxLength != nil && *p.MaxLength < limit {
			limit = *p.MaxLength
		}
		if len([]rune(s)) > limit {
			return nil, fmt.Errorf("%s dài quá %d ký tự", p.Title, limit)
		}
		if len(p.Enum) > 0 && !contains(p.Enum, s) {
			return nil, fmt.Errorf("%s có giá trị không hợp lệ", p.Title)
		}
		if p.Component == "url" && s != "" && !strings.HasPrefix(s, "https://") && !strings.HasPrefix(s, "http://") {
			return nil, fmt.Errorf("%s phải là một URL hợp lệ", p.Title)
		}
		return s, nil

	case "integer", "number":
		n, ok := toFloat(raw)
		if !ok {
			return nil, fmt.Errorf("%s phải là một số", p.Title)
		}
		if p.Minimum != nil && n < *p.Minimum {
			return nil, fmt.Errorf("%s phải >= %g", p.Title, *p.Minimum)
		}
		if p.Maximum != nil && n > *p.Maximum {
			return nil, fmt.Errorf("%s phải <= %g", p.Title, *p.Maximum)
		}
		if p.Type == "integer" {
			return int(n), nil
		}
		return n, nil

	case "boolean":
		b, ok := raw.(bool)
		if !ok {
			return nil, fmt.Errorf("%s phải là true/false", p.Title)
		}
		return b, nil

	case "array":
		items, ok := raw.([]any)
		if !ok {
			return nil, fmt.Errorf("%s phải là một danh sách", p.Title)
		}
		if len(items) > MaxEnumValues {
			return nil, fmt.Errorf("%s chọn quá nhiều giá trị", p.Title)
		}
		out := make([]string, 0, len(items))
		for _, it := range items {
			s, ok := it.(string)
			if !ok {
				return nil, fmt.Errorf("%s chứa giá trị không hợp lệ", p.Title)
			}
			if len(p.Enum) > 0 && !contains(p.Enum, s) {
				return nil, fmt.Errorf("%s có giá trị không hợp lệ", p.Title)
			}
			out = append(out, s)
		}
		return out, nil
	}
	return nil, fmt.Errorf("%s có kiểu dữ liệu không được hỗ trợ", p.Title)
}

// SlideCount reads the canonical slide_count field, falling back to 5.
func SlideCount(input map[string]any) int {
	if n, ok := toFloat(input["slide_count"]); ok && n >= 3 && n <= 10 {
		return int(n)
	}
	return 5
}

func Topic(input map[string]any) string {
	if s, ok := input["topic"].(string); ok {
		return s
	}
	return ""
}

func CTA(input map[string]any) string {
	if s, ok := input["cta"].(string); ok {
		return s
	}
	return "save"
}

func isEmpty(v any) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(t) == ""
	case []any:
		return len(t) == 0
	}
	return false
}

func toFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	}
	return 0, false
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
