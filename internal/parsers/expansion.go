package parsers

import (
	"regexp"
	"strings"
)

// POSIX-style parameter expansion shared by the shell and Dockerfile scanners.
var (
	// ${NAME}, ${NAME:-default}, ${NAME:?message}, etc. Group 2 is the operator,
	// group 3 the default/message text.
	dollarBracedPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(:?[-=?+])?([^}]*)\}`)
	// Bare $NAME (not $1, $?, $@ — those don't start with a letter/underscore).
	dollarSimplePattern = regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
)

// expansionRequiredness maps a parameter-expansion operator to requiredness.
// `:-`/`:=` (and the colonless forms) supply a default; `:?` errors when unset
// and `:+` substitutes only when set — both keep the reference required.
func expansionRequiredness(operator string, text string) (string, *string) {
	if strings.ContainsAny(operator, "-=") {
		value := stripMatchingQuotes(text)
		return "optional_with_default", &value
	}
	return "required", nil
}
