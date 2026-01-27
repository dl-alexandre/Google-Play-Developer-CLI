package migrate

import "testing"

func TestValidateText(t *testing.T) {
	if err := ValidateText("unknown", "abc"); err != nil {
		t.Fatalf("expected nil for unknown field")
	}
	if err := ValidateText("title", ""); err != nil {
		t.Fatalf("expected nil for empty text")
	}
	if err := ValidateText("title", "short title"); err != nil {
		t.Fatalf("expected nil within limit")
	}
	long := make([]rune, 31)
	for i := range long {
		long[i] = 'a'
	}
	err := ValidateText("title", string(long))
	if err == nil {
		t.Fatalf("expected error for over limit")
	}
	if err.Limit != 30 || err.Current != 31 {
		t.Fatalf("unexpected limit/current: %v/%v", err.Limit, err.Current)
	}
}

func TestValidateTextUnicode(t *testing.T) {
	text := "é"
	err := ValidateText("shortDescription", text)
	if err != nil {
		t.Fatalf("unexpected error for short unicode")
	}
}
