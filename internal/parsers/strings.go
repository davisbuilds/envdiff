package parsers

// stripMatchingQuotes removes a single pair of matching surrounding single or
// double quotes, if present. Used to normalize captured default literals.
func stripMatchingQuotes(value string) string {
	if len(value) >= 2 {
		first := value[0]
		last := value[len(value)-1]
		if first == last && (first == '"' || first == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}
