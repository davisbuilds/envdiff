// Package lines reads files into lines using the same splitting rules as
// bufio.ScanLines but without its 64 KB per-line limit, so files containing
// very long lines are scanned intact instead of aborting the whole run.
package lines

import (
	"os"
	"strings"
)

// Read returns the lines of the file at path.
func Read(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Split(data), nil
}

// Split divides data into lines on "\n", trimming a trailing "\r" from each
// line, and (like bufio.ScanLines) does not emit a trailing empty line for
// content that ends with a newline.
func Split(data []byte) []string {
	if len(data) == 0 {
		return []string{}
	}
	text := strings.TrimSuffix(string(data), "\n")
	parts := strings.Split(text, "\n")
	for index := range parts {
		parts[index] = strings.TrimSuffix(parts[index], "\r")
	}
	return parts
}
