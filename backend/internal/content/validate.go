package content

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tks/backend/internal/language"
)

const (
	MaxHeadlineChars = 90
	MaxBodyChars     = 220
)

// Problem carries enough detail to become a repair instruction for the AI.
type Problem struct {
	Rule   string `json:"rule"`
	Detail string `json:"detail"`
}

type Report struct {
	Problems []Problem `json:"problems"`
}

func (r Report) OK() bool { return len(r.Problems) == 0 }

func (r Report) Instruction() string {
	parts := make([]string, 0, len(r.Problems))
	for _, p := range r.Problems {
		parts = append(parts, "- "+p.Detail)
	}
	return strings.Join(parts, "\n")
}

// Validate runs the semantic gate: structure alone is not success (§119).
// wantSlides is the slide count the user asked for.
func Validate(c *Content, wantSlides int) Report {
	var rep Report
	fail := func(rule, format string, args ...any) {
		rep.Problems = append(rep.Problems, Problem{Rule: rule, Detail: fmt.Sprintf(format, args...)})
	}

	c.Title = strings.TrimSpace(c.Title)
	if c.Title == "" && len(c.Slides) > 0 {
		c.Title = strings.TrimSpace(c.Slides[0].Headline)
	}

	if wantSlides > 0 && len(c.Slides) != wantSlides {
		fail("slide_count", "carousel must have exactly %d slides, got %d", wantSlides, len(c.Slides))
	}
	if len(c.Slides) == 0 {
		return rep
	}

	seen := map[string]bool{}
	hasHook, hasCTA := false, false

	for i := range c.Slides {
		s := &c.Slides[i]
		s.Index = i + 1
		s.Headline = collapse(s.Headline)
		s.Body = collapse(s.Body)
		s.Role = strings.ToLower(strings.TrimSpace(s.Role))

		if s.Headline == "" {
			fail("empty_headline", "slide %d has an empty headline", s.Index)
			continue
		}
		if n := len([]rune(s.Headline)); n > MaxHeadlineChars {
			fail("headline_too_long", "slide %d headline is %d characters, rewrite it to at most %d", s.Index, n, MaxHeadlineChars)
		}
		if n := len([]rune(s.Body)); n > MaxBodyChars {
			fail("body_too_long", "slide %d body is %d characters, rewrite it to at most %d", s.Index, n, MaxBodyChars)
		}
		key := strings.ToLower(s.Headline)
		if seen[key] {
			fail("duplicate_slide", "slide %d repeats an earlier headline, write something new", s.Index)
		}
		seen[key] = true

		if i == 0 {
			hasHook = true
			if s.Role == "" {
				s.Role = "hook"
			}
		}
		if s.Role == "cta" || i == len(c.Slides)-1 {
			hasCTA = true
			if s.Role == "" {
				s.Role = "cta"
			}
		}
		if s.Role == "" {
			s.Role = "point"
		}
		if s.ImageQuery == "" {
			s.ImageQuery = s.ImageIntent
		}
	}

	if !hasHook {
		fail("missing_hook", "the first slide must be a hook")
	}
	if !hasCTA {
		fail("missing_cta", "the last slide must contain the call to action")
	}
	rep.Problems = append(rep.Problems, moderate(c)...)
	return rep
}

func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }

// Absolute health/finance claims are the failure mode most likely to hurt a
// creator in a sensitive niche, so they are blocked, not warned about (§160).
var bannedClaims = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(chữa khỏi hoàn toàn|khỏi 100%|trị dứt điểm|cam kết khỏi)\b`),
	regexp.MustCompile(`(?i)\b(guaranteed cure|cures? (all|any)|100% safe|no risk)\b`),
	regexp.MustCompile(`(?i)\b(lãi suất cam kết|chắc chắn x[0-9]+ tài khoản|lợi nhuận đảm bảo|không bao giờ lỗ)\b`),
	regexp.MustCompile(`(?i)\b(guaranteed returns?|risk[- ]free profit|cannot lose)\b`),
}

func moderate(c *Content) []Problem {
	var out []Problem
	for _, s := range c.Slides {
		text := s.Headline + " " + s.Body
		for _, re := range bannedClaims {
			if m := re.FindString(text); m != "" {
				out = append(out, Problem{
					Rule:   "unsupported_claim",
					Detail: fmt.Sprintf("slide %d makes an absolute claim (%q); rewrite it as careful, non-guaranteed advice", s.Index, m),
				})
			}
		}
	}
	return out
}

// SensitiveNiche flags niches that should carry a disclaimer slide (§160).
var sensitivePatterns = regexp.MustCompile(`(?i)(sức khỏe|y tế|thuốc|bệnh|trẻ sơ sinh|dinh dưỡng|tài chính|đầu tư|chứng khoán|crypto|health|medical|medicine|financ|invest|nutrition|baby|infant)`)

func SensitiveNiche(niche string) bool { return sensitivePatterns.MatchString(niche) }

// Disclaimer text lives in the language registry so every offered content
// language has one written in that language, not an English stand-in.
func Disclaimer(languageCode string) string { return language.Disclaimer(languageCode) }
