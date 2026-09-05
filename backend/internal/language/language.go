// Package language is the registry of content languages a project may use.
//
// The list is deliberately short and deliberately Latin-script: the export
// renderer draws with the fonts in internal/fontkit, so a language whose glyphs
// those fonts lack would export as empty boxes while looking fine in the
// browser — the same silent failure §157 warns about. Every entry here is
// glyph-tested against every registry font (see fontkit/coverage_test.go).
package language

import "strings"

type Language struct {
	Code       string `json:"code"`
	NativeName string `json:"native_name"`
	EnglishName string `json:"english_name"`
	// Probe carries the characters this language needs beyond plain ASCII.
	Probe string `json:"-"`
	// Disclaimer is appended to sensitive-niche carousels (§160).
	Disclaimer string `json:"-"`
}

var Supported = []Language{
	{
		Code: "vi", NativeName: "Tiếng Việt", EnglishName: "Vietnamese",
		Probe:      "ăâđêôơư ĂÂĐÊÔƠƯ ệ ộ ữ ỹ ặ ẫ ợ ử Nguyễn Huệ Đường",
		Disclaimer: "Nội dung mang tính tham khảo, không thay thế tư vấn chuyên môn.",
	},
	{
		Code: "en", NativeName: "English", EnglishName: "English",
		Probe:      "naive café résumé",
		Disclaimer: "For reference only — not a substitute for professional advice.",
	},
	{
		Code: "id", NativeName: "Bahasa Indonesia", EnglishName: "Indonesian",
		Probe:      "Ekonomi keluarga",
		Disclaimer: "Hanya sebagai referensi — bukan pengganti saran profesional.",
	},
	{
		Code: "ms", NativeName: "Bahasa Melayu", EnglishName: "Malay",
		Probe:      "Kesihatan keluarga",
		Disclaimer: "Sebagai rujukan sahaja — bukan pengganti nasihat profesional.",
	},
	{
		Code: "tl", NativeName: "Filipino", EnglishName: "Filipino",
		Probe:      "Ang pamilya ñ",
		Disclaimer: "Para sa sanggunian lamang — hindi kapalit ng payo ng propesyonal.",
	},
	{
		Code: "es", NativeName: "Español", EnglishName: "Spanish",
		Probe:      "áéíóú ñ ü ¿¡ Añadir",
		Disclaimer: "Solo a modo de referencia; no sustituye el consejo profesional.",
	},
	{
		Code: "pt", NativeName: "Português", EnglishName: "Portuguese",
		Probe:      "ãõ ç áéíóú âêô Coração",
		Disclaimer: "Apenas para referência — não substitui aconselhamento profissional.",
	},
	{
		Code: "fr", NativeName: "Français", EnglishName: "French",
		Probe:      "àâçéèêëîïôùûü œ Français",
		Disclaimer: "À titre indicatif — ne remplace pas un avis professionnel.",
	},
	{
		Code: "de", NativeName: "Deutsch", EnglishName: "German",
		Probe:      "äöü ÄÖÜ ß Grüße",
		Disclaimer: "Nur zur Orientierung — ersetzt keine fachliche Beratung.",
	},
}

const Default = "vi"

var byCode = func() map[string]Language {
	m := make(map[string]Language, len(Supported))
	for _, l := range Supported {
		m[l.Code] = l
	}
	return m
}()

func IsSupported(code string) bool {
	_, ok := byCode[normalize(code)]
	return ok
}

// Get returns the language, falling back to the default for anything unknown.
func Get(code string) Language {
	if l, ok := byCode[normalize(code)]; ok {
		return l
	}
	return byCode[Default]
}

// Normalize maps a request value onto a supported code.
func Normalize(code string) string { return Get(code).Code }

// Name is what the AI is told to write in. The English name is used because
// that is what a model reliably understands as an instruction.
func Name(code string) string { return Get(code).EnglishName }

func Disclaimer(code string) string { return Get(code).Disclaimer }

func Codes() []string {
	out := make([]string, 0, len(Supported))
	for _, l := range Supported {
		out = append(out, l.Code)
	}
	return out
}

func normalize(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if i := strings.IndexAny(code, "-_"); i > 0 {
		code = code[:i]
	}
	return code
}
