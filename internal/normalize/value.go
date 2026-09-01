package normalize

import (
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

var placeholderValues = map[string]struct{}{
	"":              {},
	"changeme":      {},
	"your_key_here": {},
	"example":       {},
	"example_value": {},
	"replace_me":    {},
}

func ValueKind(value string) string {
	stripped := strings.TrimSpace(value)
	lowered := strings.ToLower(stripped)

	if _, ok := placeholderValues[lowered]; ok {
		return "placeholder"
	}
	if lowered == "true" || lowered == "false" {
		return "boolean"
	}
	if isInt(stripped) {
		return "integer"
	}
	if isFloat(stripped) {
		return "float"
	}
	if looksLikeURL(stripped) {
		return "url"
	}
	if LooksLikeSecret(stripped) {
		return "secret_like"
	}
	return "string"
}

func IsPlaceholder(value string) bool {
	return ValueKind(value) == "placeholder"
}

func IsNonEmptyPlaceholder(value string) bool {
	return strings.TrimSpace(value) != "" && IsPlaceholder(value)
}

func LooksLikeSecret(value string) bool {
	// Count code points, not bytes: a byte-length gate would classify a
	// multibyte value differently from an equivalent-length ASCII one.
	if utf8.RuneCountInString(value) < 20 {
		return false
	}
	letters := 0
	digits := 0
	for _, character := range value {
		if unicode.IsLetter(character) {
			letters++
		}
		// ASCII-only, matching allDigits below: a non-ASCII digit code
		// point (e.g. a Devanagari digit) should not by itself make a
		// value look secret-like.
		if character >= '0' && character <= '9' {
			digits++
		}
	}
	return letters > 0 && digits > 0
}

func isInt(value string) bool {
	if value == "" {
		return false
	}
	if strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") {
		value = value[1:]
	}
	return value != "" && allDigits(value)
}

func isFloat(value string) bool {
	if value == "" || strings.Count(value, ".") != 1 {
		return false
	}
	parts := strings.SplitN(value, ".", 2)
	left := parts[0]
	right := parts[1]
	if strings.HasPrefix(left, "+") || strings.HasPrefix(left, "-") {
		left = left[1:]
	}
	return left != "" && right != "" && allDigits(left) && allDigits(right)
}

func looksLikeURL(value string) bool {
	if !strings.Contains(value, "://") {
		return false
	}
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

func allDigits(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
