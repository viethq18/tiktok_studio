package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tks/backend/internal/content"
	"github.com/tks/backend/internal/design"
)

// MockProvider produces schema-valid, deterministic output for every task so the
// whole pipeline — worker, validation, editor, export — can be run and tested
// with no API key and no network. It is selected automatically when AI_API_KEY
// is empty, and it is the fixture backend for the prompt eval harness.
type MockProvider struct{}

func NewMockProvider() *MockProvider { return &MockProvider{} }

func (p *MockProvider) Name() string  { return "mock" }
func (p *MockProvider) Model() string { return "mock-deterministic" }

func (p *MockProvider) Complete(ctx context.Context, req Request) (Response, error) {
	var (
		out any
		err error
	)
	switch req.Task {
	case TaskProjectAnalysis:
		out = mockProjectAnalysis(req.Vars)
	case TaskCarouselSchema:
		out = mockCarouselSchema(req.Vars)
	case TaskResearch:
		out = mockResearch(req.Vars)
	case TaskContent:
		out = mockContent(req.Vars)
	case TaskFormulaSelection:
		out = mockFormula(req.Vars)
	case TaskDesign:
		out, err = mockDesign(req.Vars)
	case TaskImageIntent:
		out = mockImageIntent(req.Vars)
	case TaskCaption:
		out = mockCaption(req.Vars)
	default:
		return Response{}, fmt.Errorf("mock provider has no fixture for task %s", req.Task)
	}
	if err != nil {
		return Response{}, err
	}
	b, err := json.Marshal(out)
	if err != nil {
		return Response{}, err
	}
	return Response{Text: string(b), Usage: Usage{InputTokens: len(req.User) / 4, OutputTokens: len(b) / 4}}, nil
}

func str(v map[string]any, k string) string {
	if s, ok := v[k].(string); ok {
		return s
	}
	return ""
}

func num(v map[string]any, k string, def int) int {
	switch t := v[k].(type) {
	case int:
		return t
	case float64:
		return int(t)
	}
	return def
}

// leadIns are the phrases a creator naturally opens with; the fixture strips
// them so the generated channel name reads like a name, not a sentence.
var leadIns = []string{
	"tôi muốn xây dựng kênh tiktok chia sẻ kiến thức về",
	"tôi muốn xây dựng kênh tiktok chia sẻ kiến thức",
	"tôi muốn xây kênh tiktok chia sẻ kiến thức về",
	"tôi muốn xây kênh tiktok chia sẻ kiến thức",
	"tôi muốn xây dựng tiktok chia sẻ kiến thức",
	"tôi muốn xây kênh tiktok về",
	"tôi muốn làm kênh về",
	"kênh tiktok về",
	"i want to build a tiktok channel about",
	"a tiktok channel about",
}

// The fixture writes in Vietnamese or English so an offline run of a non-Vietnamese
// project does not come back in the wrong language. It is a fixture, not a
// translation layer: any other language gets the English copy.
func vietnamese(v map[string]any) bool { return str(v, "language_code") == "vi" || str(v, "language") == "Vietnamese" }

func pick(v map[string]any, vi, en string) string {
	if vietnamese(v) {
		return vi
	}
	return en
}

func mockProjectAnalysis(v map[string]any) ProjectIntelligence {
	niche := strings.TrimSpace(str(v, "niche_input"))
	short := niche
	lower := strings.ToLower(short)
	for _, lead := range leadIns {
		if strings.HasPrefix(lower, lead) {
			short = strings.TrimSpace(short[len(lead):])
			break
		}
	}
	if r := []rune(short); len(r) > 48 {
		// Trim at a word boundary so the name never ends mid-word.
		cut := string(r[:48])
		if idx := strings.LastIndexByte(cut, ' '); idx > 20 {
			cut = cut[:idx]
		}
		short = strings.TrimSpace(cut)
	}
	short = strings.ToUpper(short[:1]) + short[1:]
	var p ProjectIntelligence
	p.SchemaVersion = "1.0"
	p.Name = short
	p.Identity.Niche = short
	p.Identity.Language = str(v, "language")
	p.Identity.Description = pick(v, "Kênh chia sẻ kiến thức thực tế về: ", "A channel sharing practical knowledge about: ") + niche
	p.Audience.Description = pick(v, "Người quan tâm tới ", "People interested in ") + short
	p.Audience.Demographics = pick(v, "18–45 tuổi, dùng TikTok hằng ngày", "18–45, on TikTok daily")
	p.Audience.PainPoints = []string{
		pick(v, "Thông tin trên mạng mâu thuẫn nhau", "Advice online contradicts itself"),
		pick(v, "Không có thời gian đọc bài dài", "No time for long-form reading"),
		pick(v, "Khó biết nên bắt đầu từ đâu", "Hard to know where to start"),
	}
	p.Tone = []string{"friendly", "trustworthy", "simple"}
	p.WritingStyle = pick(v, "Câu ngắn, ngôi thứ hai, tránh thuật ngữ.", "Short sentences, second person, no jargon.")
	p.PreferredAngles = []string{
		pick(v, "Sai lầm phổ biến", "Common mistakes"),
		pick(v, "Checklist nhanh", "Quick checklist"),
		pick(v, "Trước và sau", "Before and after"),
	}
	for _, pillar := range []struct{ id, name, desc string }{
		{"basics", pick(v, "Kiến thức nền", "Fundamentals"), pick(v, "Những điều cần biết đầu tiên", "What to know first")},
		{"mistakes", pick(v, "Sai lầm thường gặp", "Common mistakes"), pick(v, "Điều nhiều người làm sai", "What people get wrong")},
		{"how-to", pick(v, "Hướng dẫn", "How-to"), pick(v, "Các bước làm cụ thể", "Concrete steps")},
		{"myths", pick(v, "Hiểu lầm", "Myths"), pick(v, "Bóc tách thông tin sai", "Unpacking bad information")},
	} {
		p.ContentPillars = append(p.ContentPillars, struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}{pillar.id, pillar.name, pillar.desc})
	}
	for _, cta := range []struct{ id, label string }{
		{"save", pick(v, "Lưu lại để dùng sau", "Save this for later")},
		{"follow", pick(v, "Theo dõi để xem thêm", "Follow for more")},
		{"share", pick(v, "Chia sẻ cho người cần", "Share with someone who needs it")},
		{"comment", pick(v, "Bình luận trải nghiệm của bạn", "Tell me your experience")},
	} {
		p.CTAOptions = append(p.CTAOptions, struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		}{cta.id, cta.label})
	}
	p.SuggestedTopics = []string{
		pick(v, "5 sai lầm phổ biến về ", "5 common mistakes about ") + short,
		pick(v, "Checklist bắt đầu với ", "A starter checklist for ") + short,
		pick(v, "3 hiểu lầm về ", "3 myths about ") + short,
		pick(v, "Trước và sau khi hiểu đúng ", "Before and after understanding ") + short,
		pick(v, "Câu hỏi thường gặp về ", "FAQs about ") + short,
	}
	return p
}

func mockCarouselSchema(v map[string]any) SchemaOutput {
	type labels struct {
		topic, topicHelp, topicPlaceholder, angle, slideCount, slideHelp, cta string
		angleLabels, ctaLabels                                                map[string]string
	}
	l := labels{
		topic: "Carousel topic", topicHelp: "What should this carousel be about?",
		topicPlaceholder: "e.g. 5 signs your baby is overtired",
		angle:            "Angle", slideCount: "Number of slides", slideHelp: "Between 3 and 7",
		cta: "Call to action",
		angleLabels: map[string]string{
			"mistakes": "Common mistakes", "how_to": "Step by step",
			"myths": "Myth busting", "checklist": "Quick checklist",
		},
		ctaLabels: map[string]string{
			"save": "Save this for later", "follow": "Follow for more",
			"share": "Share with someone", "comment": "Tell me what you think",
		},
	}
	if vietnamese(v) {
		l = labels{
			topic: "Chủ đề carousel", topicHelp: "Carousel này nói về điều gì?",
			topicPlaceholder: "VD: 5 dấu hiệu bé đang thiếu ngủ",
			angle:            "Góc tiếp cận", slideCount: "Số slide", slideHelp: "3 tới 7 slide",
			cta: "Kêu gọi hành động",
			angleLabels: map[string]string{
				"mistakes": "Sai lầm phổ biến", "how_to": "Hướng dẫn từng bước",
				"myths": "Bóc tách hiểu lầm", "checklist": "Checklist nhanh",
			},
			ctaLabels: map[string]string{
				"save": "Lưu lại để dùng sau", "follow": "Theo dõi để xem thêm",
				"share": "Chia sẻ cho người cần", "comment": "Bình luận trải nghiệm",
			},
		}
	}

	schemaDoc := map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
		"properties": map[string]any{
			"topic": map[string]any{
				"type": "string", "title": l.topic, "description": l.topicHelp, "maxLength": 300,
			},
			"angle": map[string]any{
				"type": "string", "title": l.angle,
				"enum": []string{"mistakes", "how_to", "myths", "checklist"},
			},
			"slide_count": map[string]any{
				"type": "integer", "title": l.slideCount, "minimum": 3, "maximum": 7, "default": 5,
			},
			"cta": map[string]any{
				"type": "string", "title": l.cta,
				"enum": []string{"save", "follow", "share", "comment"},
			},
		},
		"required": []string{"topic", "slide_count", "cta"},
	}
	uiDoc := map[string]any{
		"order": []string{"topic", "angle", "slide_count", "cta"},
		"fields": map[string]any{
			"topic":       map[string]any{"component": "textarea", "placeholder": l.topicPlaceholder},
			"angle":       map[string]any{"component": "radio", "enum_labels": l.angleLabels},
			"slide_count": map[string]any{"component": "slider", "help": l.slideHelp},
			"cta":         map[string]any{"component": "radio", "enum_labels": l.ctaLabels},
		},
	}
	schemaJSON, _ := json.Marshal(schemaDoc)
	uiJSON, _ := json.Marshal(uiDoc)
	return SchemaOutput{Schema: schemaJSON, UI: uiJSON}
}

func mockResearch(v map[string]any) ResearchOutput {
	q := str(v, "query")
	return ResearchOutput{
		Summary: pick(v,
			"Tổng hợp nhanh về “"+q+"”: đây là chủ đề nhiều người quan tâm, thông tin phổ biến thường thiếu ngữ cảnh và dễ bị hiểu sai.",
			"A quick synthesis on “"+q+"”: widely searched, and the common advice usually lacks the context that makes it correct."),
		KeyFacts: []string{
			pick(v, "Phần lớn người mới tiếp cận “"+q+"” bắt đầu từ nguồn không kiểm chứng.", "Most newcomers to “"+q+"” start from unverified sources."),
			pick(v, "Có ít nhất ba dấu hiệu dễ nhận biết liên quan tới "+q+".", "There are at least three easily observed signals around "+q+"."),
			pick(v, "Sai lầm phổ biến nhất là áp dụng lời khuyên chung cho hoàn cảnh riêng.", "The commonest mistake is applying generic advice to a specific situation."),
			pick(v, "Thay đổi nhỏ, lặp lại đều đặn cho kết quả tốt hơn thay đổi lớn một lần.", "Small repeated changes beat one large change."),
			pick(v, "Nên theo dõi kết quả trong 1–2 tuần trước khi kết luận.", "Track results for one to two weeks before concluding anything."),
		},
		Angles: []string{
			pick(v, "Liệt kê dấu hiệu", "List the signals"),
			pick(v, "Bóc tách hiểu lầm", "Unpack the myths"),
			pick(v, "Checklist hành động", "An action checklist"),
		},
		Cautions: []string{
			pick(v, "Không đưa ra cam kết tuyệt đối", "Avoid absolute promises"),
			pick(v, "Luôn nhắc bối cảnh cá nhân khác nhau", "Always note that context differs per person"),
		},
		Sources: nil,
	}
}

func mockContent(v map[string]any) ContentOutput {
	n := num(v, "slide_count", 5)
	topic := str(v, "inputs")
	if i := strings.Index(topic, "topic:"); i >= 0 {
		topic = strings.TrimSpace(strings.SplitN(topic[i+6:], "\n", 2)[0])
	}
	if topic == "" {
		topic = "chủ đề của bạn"
	}
	title := topic
	if r := []rune(title); len(r) > 58 {
		title = string(r[:58])
	}

	out := ContentOutput{SchemaVersion: "1.0", Title: title}
	out.Slides = append(out.Slides, content.Slide{
		Index: 1, Role: "hook", Headline: clipRunes(topic, content.MaxHeadlineChars),
		ImageIntent: "A calm, human scene introducing the topic", ImageQuery: "calm lifestyle portrait",
	})
	points := []struct{ h, b, q string }{
		{pick(v, "Điều đầu tiên cần biết", "The first thing to know"),
			pick(v, "Bắt đầu từ những dấu hiệu dễ quan sát nhất trước khi thay đổi bất cứ điều gì.", "Start from the signs you can actually observe, before changing anything."), "notebook checklist desk"},
		{pick(v, "Sai lầm nhiều người mắc", "The mistake most people make"),
			pick(v, "Áp dụng lời khuyên chung mà bỏ qua bối cảnh riêng của mình.", "Applying generic advice while ignoring their own context."), "person thinking window"},
		{pick(v, "Cách làm đúng", "What to do instead"),
			pick(v, "Chọn một thay đổi nhỏ, giữ đều trong hai tuần rồi mới đánh giá lại.", "Pick one small change, hold it for two weeks, then judge."), "morning routine light"},
		{pick(v, "Dấu hiệu bạn đang đi đúng hướng", "Signs it is working"),
			pick(v, "Kết quả ổn định hơn, ít phải điều chỉnh gấp hơn trước.", "Steadier results, fewer last-minute corrections."), "sunrise calm nature"},
		{pick(v, "Việc nên làm ngay hôm nay", "Do this today"),
			pick(v, "Ghi lại một quan sát mỗi ngày — dữ liệu của bạn đáng tin hơn lời khuyên chung.", "Write down one observation a day — your own data beats generic advice."), "journal writing hands"},
	}
	for i := 2; i < n; i++ {
		p := points[(i-2)%len(points)]
		out.Slides = append(out.Slides, content.Slide{
			Index: i, Role: "point", Headline: fmt.Sprintf("%d. %s", i-1, p.h), Body: p.b,
			ImageIntent: "Supporting photo for: " + p.h, ImageQuery: p.q,
		})
	}
	ctaBody := pick(v, "Lưu lại để xem khi cần — bạn sẽ không phải tìm lại từ đầu.",
		"Save this so you do not have to search for it again.")
	if d := str(v, "disclaimer"); d != "" {
		ctaBody = clipRunes(ctaBody+" "+d, content.MaxBodyChars)
	}
	out.Slides = append(out.Slides, content.Slide{
		Index: n, Role: "cta", Headline: pick(v, "Giữ lại cho lần sau", "Keep this for later"), Body: ctaBody,
		ImageIntent: "Warm closing image", ImageQuery: "warm cozy hands cup",
	})
	return out
}

func clipRunes(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return strings.TrimSpace(string(r[:n]))
}

func mockFormula(v map[string]any) FormulaSelection {
	body := str(v, "content")
	switch {
	case strings.Contains(body, "hiểu lầm"):
		return FormulaSelection{FormulaID: "myth_fact", Reason: "Content contrasts belief with reality."}
	case strings.Contains(body, "Sai lầm"):
		return FormulaSelection{FormulaID: "mistake_fix", Reason: "Content names mistakes and corrects them."}
	default:
		return FormulaSelection{FormulaID: "listicle", Reason: "Content is a numbered set of points."}
	}
}

// mockDesign lays out slides with the same rules the prompt asks the model for,
// so the offline path exercises the real validator rather than bypassing it.
func mockDesign(v map[string]any) (DesignOutput, error) {
	var c ContentOutput
	if err := json.Unmarshal([]byte(str(v, "content")), &c); err != nil {
		return DesignOutput{}, fmt.Errorf("mock design: %w", err)
	}
	w := num(v, "canvas_width", 1080)
	h := num(v, "canvas_height", 1350)
	safe := design.SafeAreaFor(w, h)

	var palette design.Palette
	_ = json.Unmarshal([]byte(str(v, "palette")), &palette)

	d := DesignOutput{
		Version: design.SchemaVersion,
		Canvas:  design.Canvas{Width: w, Height: h, Ratio: str(v, "ratio"), SafeArea: safe},
		Palette: palette,
	}
	style := str(v, "template_style")

	for i, cs := range c.Slides {
		id := fmt.Sprintf("slide_%d", i+1)
		slide := design.Slide{
			ID: id, Index: i + 1, Role: cs.Role, Template: style,
			Background: palette.Background, ImageIntent: cs.ImageIntent, ImageQuery: cs.ImageQuery,
			Elements: []design.Element{{
				ID: id + "_bg", Type: "image", X: 0, Y: 0,
				Width: float64(w), Height: float64(h), Opacity: 1, Fit: "cover",
				Overlay: &design.Overlay{Color: "#000000", Opacity: 0.48},
			}},
		}

		switch cs.Role {
		case "hook":
			slide.Elements = append(slide.Elements,
				textEl(id+"_hook", "hook", cs.Headline, float64(safe.X), float64(h)*0.34, float64(safe.Width), 78, 700, "#FFFFFF", "center", 1.15),
				shapeEl(id+"_rule", float64(w)/2-60, float64(h)*0.34-44, 120, 8, palette.Accent),
			)
		case "cta":
			slide.Elements = append(slide.Elements,
				textEl(id+"_cta", "cta", cs.Headline, float64(safe.X), float64(h)*0.38, float64(safe.Width), 58, 700, "#FFFFFF", "center", 1.2),
			)
			if cs.Body != "" {
				slide.Elements = append(slide.Elements,
					textEl(id+"_body", "body", cs.Body, float64(safe.X), float64(h)*0.52, float64(safe.Width), 34, 400, "#FFFFFF", "center", 1.4))
			}
		default:
			// Alternate top- and bottom-weighted layouts so the set feels designed.
			top := float64(safe.Y) + 40
			if i%2 == 1 {
				top = float64(h) * 0.42
			}
			slide.Elements = append(slide.Elements,
				textEl(id+"_headline", "headline", cs.Headline, float64(safe.X), top, float64(safe.Width), 62, 700, "#FFFFFF", "left", 1.2),
			)
			if cs.Body != "" {
				slide.Elements = append(slide.Elements,
					textEl(id+"_body", "body", cs.Body, float64(safe.X), top+190, float64(safe.Width), 36, 400, "#F3F4F6", "left", 1.4))
			}
		}
		d.Slides = append(d.Slides, slide)
	}
	return d, nil
}

func textEl(id, role, contentText string, x, y, width, size float64, weight int, color, align string, lh float64) design.Element {
	return design.Element{
		ID: id, Type: "text", Role: role, Content: contentText,
		X: x, Y: y, Width: width, Opacity: 1,
		Style: &design.Style{
			FontFamily: "Inter", FontSize: size, FontWeight: weight,
			Color: color, TextAlign: align, LineHeight: lh,
		},
	}
}

func shapeEl(id string, x, y, w, h float64, fill string) design.Element {
	return design.Element{
		ID: id, Type: "shape", Shape: "rect", X: x, Y: y, Width: w, Height: h,
		Fill: fill, Opacity: 1, Radius: 4,
	}
}

func mockCaption(v map[string]any) CaptionOutput {
	var c ContentOutput
	_ = json.Unmarshal([]byte(str(v, "content")), &c)
	topic := c.Title
	if topic == "" {
		topic = "chủ đề này"
	}
	caption := pick(v,
		"Nhiều người bỏ qua "+topic+" cho tới khi nó thành vấn đề.\nLướt hết bộ ảnh để xem những dấu hiệu dễ nhận ra nhất.\n",
		"Most people ignore "+topic+" until it becomes a problem.\nSwipe through for the signals that are easiest to spot.\n")
	tags := []string{"#fyp", "#learnontiktok", "#tips", "#mindset", "#howto", "#advice", "#dailyhabits", "#creators"}
	if vietnamese(v) {
		tags = []string{"#fyp", "#xuhuong", "#learnontiktok", "#mindset", "#kienthuc", "#chiaseketnoi", "#mucsongtot", "#tips"}
	}
	return CaptionOutput{Caption: caption + str(v, "cta"), Hashtags: tags}
}

func mockImageIntent(v map[string]any) ImageIntentOutput {
	var slides []struct {
		SlideID  string `json:"slide_id"`
		Headline string `json:"headline"`
	}
	_ = json.Unmarshal([]byte(str(v, "slides")), &slides)
	var out ImageIntentOutput
	for _, s := range slides {
		out.Slides = append(out.Slides, struct {
			SlideID     string `json:"slide_id"`
			ImageQuery  string `json:"image_query"`
			ImageIntent string `json:"image_intent"`
		}{s.SlideID, "minimal lifestyle background", "Neutral supporting photo"})
	}
	return out
}
