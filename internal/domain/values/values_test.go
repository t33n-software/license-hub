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

func TestParseSkipsInboundKey(t *testing.T) {
	got, err := Parse([]byte(`{"PROJECT_NAME":"x","INBOUND_LICENSES":[{"id":"MIT"}]}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if _, ok := got["INBOUND_LICENSES"]; ok {
		t.Fatal("Parse() leaked the inbound key into the substitution map")
	}
	if got["PROJECT_NAME"] != "x" {
		t.Fatalf("Parse() = %v", got)
	}
}

func TestInboundRawReturnsArray(t *testing.T) {
	raw, err := InboundRaw([]byte(`{"PROJECT_NAME":"x","INBOUND_LICENSES":[{"id":"MIT"}]}`))
	if err != nil {
		t.Fatalf("InboundRaw() error = %v", err)
	}
	if string(raw) != `[{"id":"MIT"}]` {
		t.Fatalf("InboundRaw() = %q", raw)
	}
}

func TestInboundRawAbsent(t *testing.T) {
	raw, err := InboundRaw([]byte(`{"PROJECT_NAME":"x"}`))
	if err != nil || raw != nil {
		t.Fatalf("InboundRaw() = %q, %v", raw, err)
	}
}

func TestInboundRawInvalidJSON(t *testing.T) {
	if _, err := InboundRaw([]byte(`{`)); err == nil {
		t.Fatal("InboundRaw() expected error for invalid JSON")
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
