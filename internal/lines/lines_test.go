package lines

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitMatchesScanLines(t *testing.T) {
	cases := []string{
		"",
		"a",
		"a\n",
		"a\nb",
		"a\nb\n",
		"\n",
		"\n\n",
		"a\r\nb\r\n",
		"a\n\n\nb",
	}
	for _, input := range cases {
		got := Split([]byte(input))

		want := []string{}
		scanner := bufio.NewScanner(strings.NewReader(input))
		for scanner.Scan() {
			want = append(want, scanner.Text())
		}

		if len(got) != len(want) {
			t.Fatalf("Split(%q) = %#v, want %#v", input, got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("Split(%q)[%d] = %q, want %q", input, i, got[i], want[i])
			}
		}
	}
}

func TestReadHandlesLineLongerThan64KB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "long.txt")
	long := strings.Repeat("x", 200_000)
	if err := os.WriteFile(path, []byte("first\n"+long+"\nlast\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(got) != 3 || got[0] != "first" || got[1] != long || got[2] != "last" {
		t.Fatalf("Read returned %d lines; middle len=%d", len(got), len(got[1]))
	}
}
