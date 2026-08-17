package values

import (
	"slices"
	"testing"
)

func TestParseValidObject(t *testing.T) {
	got, err := Parse([]byte(`{"PROJECT_NAME":"license-hub"}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got["PROJECT_NAME"] != "license-hub" {
		t.Fatalf("Parse() = %v", got)
	}
}

func TestParseRejectsInvalidJSON(t *testing.T) {
	if _, err := Parse([]byte(`{`)); err == nil {
		t.Fatal("Parse() expected error for invalid JSON")
	}
}

func TestParseRejectsNonStringValues(t *testing.T) {
	if _, err := Parse([]byte(`{"PROJECT_NAME":1}`)); err == nil {
		t.Fatal("Parse() expected error for non-string value")
	}
}

func TestMergeProjectWinsOnConflict(t *testing.T) {
	org := map[string]string{"COPYRIGHT_HOLDER": "Org", "VENUE": "Germany"}
	project := map[string]string{"VENUE": "Berlin, Germany"}
	got := Merge(org, project)
	if got["COPYRIGHT_HOLDER"] != "Org" {
		t.Fatalf("Merge() lost org default: %v", got)
	}
	if got["VENUE"] != "Berlin, Germany" {
		t.Fatalf("Merge() did not let project win: %v", got)
	}
}

func TestMissingRequiredListsAbsentAnchorsSorted(t *testing.T) {
	got := MissingRequired(map[string]string{"PROJECT_NAME": "x"})
	if len(got) != len(RequiredKeys())-1 {
		t.Fatalf("MissingRequired() = %v", got)
	}
	if !slices.IsSorted(got) {
		t.Fatalf("MissingRequired() not sorted: %v", got)
	}
	if slices.Contains(got, "PROJECT_NAME") {
		t.Fatalf("MissingRequired() contains present key: %v", got)
	}
}

func TestMissingRequiredComplete(t *testing.T) {
	full := make(map[string]string)
	for _, key := range RequiredKeys() {
		full[key] = "x"
	}
	if got := MissingRequired(full); len(got) != 0 {
		t.Fatalf("MissingRequired() = %v, want empty", got)
	}
}

func FuzzParse(f *testing.F) {
	f.Add(`{"PROJECT_NAME":"license-hub"}`)
	f.Add(`{}`)
	f.Add(`{`)
	f.Add(`{"A":1}`)
	f.Add(`"string"`)
	f.Fuzz(func(t *testing.T, input string) {
		got, err := Parse([]byte(input))
		if err == nil && got == nil {
			t.Fatal("Parse() succeeded with nil map")
		}
	})
}
