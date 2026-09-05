package schema

import "encoding/json"

type fallbackCopy struct {
	topicTitle, topicHelp, topicPlaceholder string
	slideCountTitle, ctaTitle              string
}

// Only the languages the product itself is translated into get hand-written
// labels here; everything else falls back to English. This form is the rare
// path (§90) — it appears only when schema generation fails.
var fallbackStrings = map[string]fallbackCopy{
	"vi": {
		topicTitle: "Chủ đề carousel", topicHelp: "Bạn muốn carousel này nói về điều gì?",
		topicPlaceholder: "VD: 5 dấu hiệu bé đang thiếu ngủ",
		slideCountTitle:  "Số slide", ctaTitle: "Kêu gọi hành động",
	},
	"en": {
		topicTitle: "Carousel topic", topicHelp: "What should this carousel be about?",
		topicPlaceholder: "e.g. 5 signs your baby is overtired",
		slideCountTitle:  "Number of slides", ctaTitle: "Call to action",
	},
}

// FallbackSchema is used when the AI schema generator is unavailable or its
// output is rejected. The product must never dead-end at "create carousel" (§90).
func FallbackSchema(ctaOptions []CTAOption, languageCode string) (json.RawMessage, json.RawMessage) {
	copy, ok := fallbackStrings[languageCode]
	if !ok {
		copy = fallbackStrings["en"]
	}
	if len(ctaOptions) == 0 {
		if languageCode == "vi" {
			ctaOptions = []CTAOption{
				{"save", "Lưu bài này"}, {"follow", "Theo dõi để xem thêm"},
				{"share", "Chia sẻ cho bạn bè"}, {"comment", "Bình luận suy nghĩ của bạn"},
			}
		} else {
			ctaOptions = []CTAOption{
				{"save", "Save this post"}, {"follow", "Follow for more"},
				{"share", "Share with a friend"}, {"comment", "Tell me what you think"},
			}
		}
	}
	ids := make([]string, 0, len(ctaOptions))
	labels := map[string]string{}
	for _, o := range ctaOptions {
		ids = append(ids, o.ID)
		labels[o.ID] = o.Label
	}
	schema := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
		"properties": map[string]any{
			"topic": map[string]any{
				"type": "string", "title": copy.topicTitle, "maxLength": 300,
				"description": copy.topicHelp,
			},
			"slide_count": map[string]any{
				"type": "integer", "title": copy.slideCountTitle, "minimum": 3, "maximum": 7, "default": 5,
			},
			"cta": map[string]any{"type": "string", "title": copy.ctaTitle, "enum": ids},
		},
		"required": []string{"topic", "slide_count", "cta"},
	}
	ui := map[string]any{
		"order": []string{"topic", "slide_count", "cta"},
		"fields": map[string]any{
			"topic":       map[string]any{"component": "textarea", "placeholder": copy.topicPlaceholder},
			"slide_count": map[string]any{"component": "slider"},
			"cta":         map[string]any{"component": "radio", "enum_labels": labels},
		},
	}
	s, _ := json.Marshal(schema)
	u, _ := json.Marshal(ui)
	return s, u
}

type CTAOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
