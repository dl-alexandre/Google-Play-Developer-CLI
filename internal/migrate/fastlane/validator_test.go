package fastlane

import "testing"

func TestValidateLocale(t *testing.T) {
	meta := &LocaleMetadata{
		Locale:              "en-US",
		Title:               "ok",
		TitleSet:            true,
		ShortDescription:    "ok",
		ShortDescriptionSet: true,
		FullDescription:     "ok",
		FullDescriptionSet:  true,
		Changelogs: map[string]string{
			"100": "ok",
		},
	}
	errs := ValidateLocale(meta)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d", len(errs))
	}

	long := make([]rune, 81)
	for i := range long {
		long[i] = 'a'
	}
	meta.ShortDescription = string(long)
	errs = ValidateLocale(meta)
	if len(errs) == 0 {
		t.Fatalf("expected validation errors")
	}
	if errs[0].Locale != "en-US" {
		t.Fatalf("expected locale set")
	}
}

func TestValidateLocaleNil(t *testing.T) {
	errs := ValidateLocale(nil)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for nil meta")
	}
}

func TestValidateLocaleReleaseNotesMessage(t *testing.T) {
	long := make([]rune, 501)
	for i := range long {
		long[i] = 'a'
	}
	meta := &LocaleMetadata{
		Locale: "en-US",
		Changelogs: map[string]string{
			"123": string(long),
		},
	}
	errs := ValidateLocale(meta)
	if len(errs) == 0 {
		t.Fatalf("expected validation error for release notes")
	}
	if errs[0].Field != "releaseNotes" {
		t.Fatalf("expected releaseNotes field, got %q", errs[0].Field)
	}
	if errs[0].Message != "release notes 123 exceeds limit" {
		t.Fatalf("unexpected message: %q", errs[0].Message)
	}
	if errs[0].Locale != "en-US" {
		t.Fatalf("expected locale set, got %q", errs[0].Locale)
	}
}
