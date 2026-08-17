package lockfile

import "testing"

func TestParseValidLock(t *testing.T) {
	lock, err := Parse([]byte(`{"template":"t.hbs","version":"1.0.0","digest":"sha256:abc"}`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if lock.Template != "t.hbs" || lock.Version != "1.0.0" || lock.Digest != "sha256:abc" {
		t.Fatalf("Parse() = %+v", lock)
	}
}

func TestParseRejectsInvalidJSON(t *testing.T) {
	if _, err := Parse([]byte(`{`)); err == nil {
		t.Fatal("Parse() expected error for invalid JSON")
	}
}

func TestParseRejectsMissingFields(t *testing.T) {
	if _, err := Parse([]byte(`{"template":"t.hbs"}`)); err == nil {
		t.Fatal("Parse() expected error for missing fields")
	}
}

func FuzzParse(f *testing.F) {
	f.Add(`{"template":"t.hbs","version":"1.0.0","digest":"sha256:abc"}`)
	f.Add(`{}`)
	f.Add(`{`)
	f.Fuzz(func(t *testing.T, input string) {
		lock, err := Parse([]byte(input))
		if err == nil && (lock.Template == "" || lock.Version == "" || lock.Digest == "") {
			t.Fatal("Parse() succeeded with incomplete lock")
		}
	})
}
