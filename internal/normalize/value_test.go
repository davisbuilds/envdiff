package normalize

import "testing"

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
