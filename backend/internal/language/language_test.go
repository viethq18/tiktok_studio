package language

import "testing"

func TestEveryLanguageHasADisclaimerInItsOwnLanguage(t *testing.T) {
	for _, l := range Supported {
		if l.Disclaimer == "" {
			t.Errorf("%s has no disclaimer; a sensitive niche in this language would ship without one", l.EnglishName)
		}
		if l.Code != "en" && l.Disclaimer == Get("en").Disclaimer {
			t.Errorf("%s reuses the English disclaimer verbatim", l.EnglishName)
		}
		if l.NativeName == "" || l.EnglishName == "" || l.Probe == "" {
			t.Errorf("%s is missing registry metadata", l.Code)
		}
	}
}

func TestUnsupportedLanguagesAreRejected(t *testing.T) {
	for _, code := range []string{"", "zz", "ja", "th", "ar", "zh"} {
		if IsSupported(code) {
			t.Errorf("%q must not be offered: the export fonts cannot draw it", code)
		}
	}
	for _, code := range Codes() {
		if !IsSupported(code) {
			t.Errorf("registered language %q reports as unsupported", code)
		}
	}
}

func TestNormalizeAcceptsRegionalTags(t *testing.T) {
	for input, want := range map[string]string{
		"vi-VN": "vi", "en_US": "en", "PT-br": "pt", " es ": "es", "unknown": Default,
	} {
		if got := Normalize(input); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNameIsWhatAModelIsTold(t *testing.T) {
	// Prompts say `write in "{{.language}}"`; "vi" is ambiguous, "Vietnamese" is not.
	if got := Name("vi"); got != "Vietnamese" {
		t.Errorf("Name(\"vi\") = %q", got)
	}
}
