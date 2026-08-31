package normalize

import (
	"strings"
	"testing"
)

func TestValueKindClassifiesPlaceholdersBooleansNumbersURLsAndSecrets(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{"", "placeholder"},
		{"replace_me", "placeholder"},
		{"false", "boolean"},
		{"-42", "integer"},
		{"+3.14", "float"},
		{"postgres://localhost/app", "url"},
		{"sk-proj-1234567890abcdef1234567890", "secret_like"},
		{"hello world", "string"},
	}

	for _, test := range tests {
		got := ValueKind(test.value)
		if got != test.want {
			t.Fatalf("ValueKind(%q) = %q, want %q", test.value, got, test.want)
		}
	}
}

func TestLooksLikeSecretCountsCodePointsNotBytes(t *testing.T) {
	// "ééééééééé12" is 9 two-byte "é" runes plus two ASCII digits: 20 bytes
	// but only 11 code points. A byte-length gate of 20 would wrongly treat
	// this as long enough to be secret-like; a rune-length gate correctly
	// treats it as too short.
	short := strings.Repeat("é", 9) + "12"
	if got := ValueKind(short); got != "string" {
		t.Fatalf("ValueKind(%q) = %q, want %q (too short in code points to be secret-like)", short, got, "string")
	}

	// The equivalent value padded with ASCII letters up to 20 code points
	// should classify as secret-like.
	long := strings.Repeat("é", 18) + "12"
	if got := ValueKind(long); got != "secret_like" {
		t.Fatalf("ValueKind(%q) = %q, want %q", long, got, "secret_like")
	}
}

func TestLooksLikeSecretUsesASCIIDigitSemantics(t *testing.T) {
	// A Devanagari digit (U+0966) satisfies unicode.IsDigit but is not an
	// ASCII digit. A value that only reaches "digit" status via non-ASCII
	// digit code points should not classify as secret-like, matching the
	// ASCII-only digit semantics used elsewhere in this package (allDigits).
	value := strings.Repeat("a", 19) + "०"
	if got := ValueKind(value); got != "string" {
		t.Fatalf("ValueKind(%q) = %q, want %q (non-ASCII digit should not count)", value, got, "string")
	}
}

func TestPlaceholderHelpersMatchPythonBehavior(t *testing.T) {
	if !IsPlaceholder("your_key_here") {
		t.Fatal("your_key_here should be placeholder")
	}
	if IsNonEmptyPlaceholder("") {
		t.Fatal("empty value should not be a non-empty placeholder")
	}
	if !IsNonEmptyPlaceholder("changeme") {
		t.Fatal("changeme should be a non-empty placeholder")
	}
}
