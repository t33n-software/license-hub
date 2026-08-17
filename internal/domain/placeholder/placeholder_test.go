package placeholder

import (
	"slices"
	"strings"
	"testing"
)

func TestUnresolvedFindsDistinctPlaceholdersInOrder(t *testing.T) {
	content := "{{VENUE}} then {{PROJECT_NAME}} then {{VENUE}} again"
	want := []string{"{{VENUE}}", "{{PROJECT_NAME}}"}
	if got := Unresolved(content); !slices.Equal(got, want) {
		t.Fatalf("Unresolved() = %v, want %v", got, want)
	}
}

func TestUnresolvedReturnsEmptyWithoutPlaceholders(t *testing.T) {
	if got := Unresolved("plain text {{ lowercase }} and {{}}"); len(got) != 0 {
		t.Fatalf("Unresolved() = %v, want empty", got)
	}
}

func FuzzUnresolved(f *testing.F) {
	f.Add("{{PROJECT_NAME}}")
	f.Add("plain text")
	f.Add("{{lowercase}}")
	f.Add("{{A}} {{A}} {{B_2}}")
	f.Fuzz(func(t *testing.T, input string) {
		for _, match := range Unresolved(input) {
			if !pattern.MatchString(match) {
				t.Fatalf("Unresolved() returned non-placeholder %q", match)
			}
			if !strings.Contains(input, match) {
				t.Fatalf("Unresolved() returned %q not contained in input", match)
			}
		}
	})
}
