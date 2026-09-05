package content

import "testing"

func sample(n int) *Content {
	c := &Content{Title: "T"}
	for i := 1; i <= n; i++ {
		role := "point"
		if i == 1 {
			role = "hook"
		} else if i == n {
			role = "cta"
		}
		c.Slides = append(c.Slides, Slide{Index: i, Role: role,
			Headline: "Tiêu đề slide số " + string(rune('0'+i)), Body: "Nội dung."})
	}
	return c
}

func TestValidateAcceptsWellFormedContent(t *testing.T) {
	if rep := Validate(sample(5), 5); !rep.OK() {
		t.Errorf("valid content was rejected: %s", rep.Instruction())
	}
}

func TestValidateRejectsWrongSlideCount(t *testing.T) {
	if rep := Validate(sample(4), 5); rep.OK() {
		t.Error("a slide-count mismatch must be reported")
	}
}

func TestValidateRejectsDuplicateSlides(t *testing.T) {
	c := sample(5)
	c.Slides[2].Headline = c.Slides[1].Headline
	if rep := Validate(c, 5); rep.OK() {
		t.Error("duplicate headlines must be reported")
	}
}

func TestValidateRejectsOverlongHeadline(t *testing.T) {
	c := sample(5)
	long := ""
	for i := 0; i < MaxHeadlineChars+20; i++ {
		long += "a"
	}
	c.Slides[1].Headline = long
	if rep := Validate(c, 5); rep.OK() {
		t.Error("an overlong headline must be reported so the AI can shorten it")
	}
}

func TestModerationBlocksAbsoluteClaims(t *testing.T) {
	c := sample(5)
	c.Slides[1].Body = "Phương pháp này trị dứt điểm mọi vấn đề."
	rep := Validate(c, 5)
	if rep.OK() {
		return
	}
	found := false
	for _, p := range rep.Problems {
		if p.Rule == "unsupported_claim" {
			found = true
		}
	}
	if !found {
		t.Error("an absolute medical claim must be flagged for rewrite")
	}
}

func TestSensitiveNicheDetection(t *testing.T) {
	if !SensitiveNiche("chăm sóc trẻ sơ sinh") {
		t.Error("a baby-care niche should require a disclaimer")
	}
	if SensitiveNiche("mẹo trang trí bàn làm việc") {
		t.Error("a decor niche should not require a disclaimer")
	}
}
