package ecosystem

import (
	"errors"
	"strings"
	"testing"
)

// Convention: docs/conventions/cli/testing/README.md

func TestComposerProjectionDerivesFromTheTenantValues(t *testing.T) {
	cases := []struct {
		name   string
		merged map[string]string
		want   string
	}{
		{"custom family", map[string]string{"LICENSE_ID": "example-NoRepublish-1.0"}, "LicenseRef-example-NoRepublish-1.0"},
		{"spdx identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "MIT"}, "MIT"},
		{"padded identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "  MIT  "}, "MIT"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := ComposerProjection(testCase.merged); got != testCase.want {
				t.Fatalf("ComposerProjection() = %q, want %q", got, testCase.want)
			}
		})
	}
}

const composerExampleManifest = "{\n  \"name\": \"example-project\",\n" +
	"  \"description\": \"example\",\n  \"license\": \"MIT\",\n" +
	"  \"type\": \"library\"\n}\n"

func TestInspectComposerLicenseClassifiesTheField(t *testing.T) {
	cases := []struct {
		name       string
		content    string
		wantValue  string
		wantState  ComposerLicenseState
		wantErrIs  error
		wantErr    bool
		wantValues []string
	}{
		{"string field", composerExampleManifest, "MIT", ComposerLicenseString, nil, false, nil},
		{"missing field", `{"name":"x"}`, "", ComposerLicenseMissing, nil, false, nil},
		{"escaped value", `{"license":"A\u0042C"}`, "ABC", ComposerLicenseString, nil, false, nil},
		{"compound expression", `{"license":"(MIT OR Apache-2.0)"}`, "(MIT OR Apache-2.0)", ComposerLicenseString, nil, false, nil},
		{"proprietary form", `{"license":"proprietary"}`, "proprietary", ComposerLicenseString, nil, false, nil},
		{"one-element array", `{"license":["MIT"]}`, "MIT", ComposerLicenseArray, nil, false, []string{"MIT"}},
		{"multi-element array", `{"license":["MIT","Apache-2.0"]}`, "", ComposerLicenseArray, nil, false, []string{"MIT", "Apache-2.0"}},
		{"number field", `{"license":7}`, "", ComposerLicenseInvalid, nil, false, nil},
		{"object field", `{"license":{"a":1}}`, "", ComposerLicenseInvalid, nil, false, nil},
		{"boolean field", `{"license":true}`, "", ComposerLicenseInvalid, nil, false, nil},
		{"null field", `{"license":null}`, "", ComposerLicenseInvalid, nil, false, nil},
		{"empty object", `{}`, "", ComposerLicenseMissing, nil, false, nil},
		{"duplicate member", `{"license":"A","license":"B"}`, "", 0, ErrAmbiguousField, true, nil},
		{"malformed json", `{"license":`, "", 0, ErrInvalidSurface, true, nil},
		{"root not an object", `[1]`, "", 0, ErrInvalidSurface, true, nil},
		{"empty array", `{"license":[]}`, "", 0, ErrInvalidComposerLicenseValue, true, nil},
		{"non-string entry", `{"license":[7]}`, "", 0, ErrInvalidComposerLicenseValue, true, nil},
		{"mixed entries", `{"license":["MIT",7]}`, "", 0, ErrInvalidComposerLicenseValue, true, nil},
		{"array trailing comma", `{"license":["MIT",]}`, "", 0, ErrInvalidSurface, true, nil},
		{"unterminated array", `{"license":["MIT"`, "", 0, ErrInvalidSurface, true, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectComposerLicense(testCase.content)
			if testCase.wantErr {
				if !errors.Is(err, testCase.wantErrIs) {
					t.Fatalf("InspectComposerLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("InspectComposerLicense() error = %v", err)
			}
			if surface.State != testCase.wantState {
				t.Fatalf("InspectComposerLicense() state = %d, want %d", surface.State, testCase.wantState)
			}
			if surface.Value != testCase.wantValue {
				t.Fatalf("InspectComposerLicense() value = %q, want %q", surface.Value, testCase.wantValue)
			}
			if len(surface.Entries) != len(testCase.wantValues) {
				t.Fatalf("InspectComposerLicense() entries = %v, want %v", surface.Entries, testCase.wantValues)
			}
			for i, want := range testCase.wantValues {
				if surface.Entries[i] != want {
					t.Fatalf("InspectComposerLicense() entries = %v, want %v", surface.Entries, testCase.wantValues)
				}
			}
		})
	}
}

func TestAlignComposerLicenseReplacesADivergingValue(t *testing.T) {
	aligned, changed, err := AlignComposerLicense(composerExampleManifest, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignComposerLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignComposerLicense() changed = false, want true")
	}
	want := strings.Replace(composerExampleManifest, `"license": "MIT"`, `"license": "LicenseRef-example-NoRepublish-1.0"`, 1)
	if aligned != want {
		t.Fatalf("AlignComposerLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignComposerLicenseNormalizesADivergingOneElementArray(t *testing.T) {
	aligned, changed, err := AlignComposerLicense(`{"name":"x","license":["MIT"]}`, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignComposerLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignComposerLicense() changed = false, want true")
	}
	want := `{"name":"x","license":"LicenseRef-example-NoRepublish-1.0"}`
	if aligned != want {
		t.Fatalf("AlignComposerLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignComposerLicenseIsIdempotentOnAnAlignedArraySpelling(t *testing.T) {
	content := `{"license":["LicenseRef-example-NoRepublish-1.0"]}`
	aligned, changed, err := AlignComposerLicense(content, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignComposerLicense() error = %v", err)
	}
	if changed {
		t.Fatal("AlignComposerLicense() changed = true, want false")
	}
	if aligned != content {
		t.Fatalf("AlignComposerLicense() mutated an aligned manifest: %q", aligned)
	}
}

func TestAlignComposerLicenseIsIdempotentOnAlignedState(t *testing.T) {
	aligned, changed, err := AlignComposerLicense(composerExampleManifest, "MIT")
	if err != nil {
		t.Fatalf("AlignComposerLicense() error = %v", err)
	}
	if changed {
		t.Fatal("AlignComposerLicense() changed = true, want false")
	}
	if aligned != composerExampleManifest {
		t.Fatalf("AlignComposerLicense() mutated an aligned manifest: %q", aligned)
	}
}

func TestAlignComposerLicensePreservesEveryNonLicenseByte(t *testing.T) {
	content := "{\n  \"name\": \"example-project\",\n  \"license\": \"MIT\",\n" +
		"  \"type\": \"library\"\n}\n"
	aligned, _, err := AlignComposerLicense(content, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignComposerLicense() error = %v", err)
	}
	before, _, _ := strings.Cut(content, `"license": "MIT"`)
	afterOriginal := content[strings.Index(content, `"license": "MIT"`)+len(`"license": "MIT"`):]
	if !strings.HasPrefix(aligned, before+`"license": "LicenseRef-example-NoRepublish-1.0"`) {
		t.Fatalf("AlignComposerLicense() changed bytes before the value: %q", aligned)
	}
	if !strings.HasSuffix(aligned, afterOriginal) {
		t.Fatalf("AlignComposerLicense() changed bytes after the value: %q", aligned)
	}
	if !strings.Contains(aligned, "\"type\": \"library\"") {
		t.Fatalf("AlignComposerLicense() touched a non-license field: %q", aligned)
	}
}

func TestAlignComposerLicenseInsertsAMissingMember(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"with members", "{\n  \"name\": \"example-project\"\n}\n"},
		{"empty object", `{}`},
		{"whitespace object", "{\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignComposerLicense(testCase.content, "LicenseRef-example-NoRepublish-1.0")
			if err != nil {
				t.Fatalf("AlignComposerLicense() error = %v", err)
			}
			if !changed {
				t.Fatal("AlignComposerLicense() changed = false, want true")
			}
			surface, err := InspectComposerLicense(aligned)
			if err != nil || surface.State != ComposerLicenseString {
				t.Fatalf("AlignComposerLicense() produced an unscannable surface: %q (%v)", aligned, err)
			}
			if surface.Value != "LicenseRef-example-NoRepublish-1.0" {
				t.Fatalf("AlignComposerLicense() license = %q", surface.Value)
			}
			if testCase.name == "with members" && !strings.Contains(aligned, "\"name\": \"example-project\"") {
				t.Fatalf("AlignComposerLicense() lost the original member: %q", aligned)
			}
		})
	}
}

func TestAlignComposerLicenseRefusesAmbiguousAndInvalidSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantErrIs error
	}{
		{"multiple entries", `{"license":["MIT","Apache-2.0"]}`, ErrMultipleComposerLicenses},
		{"object field", `{"license":{"type":"MIT"}}`, ErrInvalidComposerLicenseValue},
		{"number field", `{"license":7}`, ErrInvalidComposerLicenseValue},
		{"duplicate member", `{"license":"A","license":"B"}`, ErrAmbiguousField},
		{"malformed json", `{"license":`, ErrInvalidSurface},
		{"root not an object", `[1]`, ErrInvalidSurface},
		{"non-utf-8 target", composerExampleManifest, ErrInvalidTarget},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			target := "LicenseRef-example-NoRepublish-1.0"
			if testCase.name == "non-utf-8 target" {
				target = "MIT\xff"
			}
			_, _, err := AlignComposerLicense(testCase.content, target)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("AlignComposerLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
			}
		})
	}
}

func TestScanComposerArrayValidatesTheRawSpan(t *testing.T) {
	cases := []struct {
		name      string
		raw       string
		want      []string
		wantErrIs error
	}{
		{"spaced entries", `[ "MIT" , "Apache-2.0" ]`, []string{"MIT", "Apache-2.0"}, nil},
		{"escaped entry", `["A\u0042C"]`, []string{"ABC"}, nil},
		{"not an array", `x`, nil, ErrInvalidComposerLicenseValue},
		{"empty array", `[]`, nil, ErrInvalidComposerLicenseValue},
		{"non-string entry", `[7]`, nil, ErrInvalidComposerLicenseValue},
		{"trailing comma", `["a",]`, nil, ErrInvalidComposerLicenseValue},
		{"missing delimiter", `["a" "b"]`, nil, ErrInvalidComposerLicenseValue},
		{"unterminated array", `["a"`, nil, ErrInvalidComposerLicenseValue},
		{"unterminated string", `["a`, nil, ErrInvalidSurface},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			entries, err := scanComposerArray(testCase.raw)
			if testCase.wantErrIs != nil {
				if !errors.Is(err, testCase.wantErrIs) {
					t.Fatalf("scanComposerArray(%q) error = %v, want %v", testCase.raw, err, testCase.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("scanComposerArray(%q) error = %v", testCase.raw, err)
			}
			if strings.Join(entries, "|") != strings.Join(testCase.want, "|") {
				t.Fatalf("scanComposerArray(%q) = %v, want %v", testCase.raw, entries, testCase.want)
			}
		})
	}
}

func FuzzAlignComposerLicense(f *testing.F) {
	f.Add(`{"name":"x","license":"MIT"}`, "LicenseRef-example-NoRepublish-1.0")
	f.Add(`{}`, "proprietary")
	f.Add(`{"license":["MIT"]}`, "MIT")
	f.Add(`{"license":["MIT","Apache-2.0"]}`, "MIT")
	f.Add(`{"license":[]}`, "MIT")
	f.Add(`{"license":`, "MIT")
	f.Add(`{"license":7}`, "MIT")
	f.Add(`[1]`, "MIT")
	f.Add(`{"a":01}`, "MIT")
	f.Add(`{"a":"\u12"}`, "MIT")
	f.Add(`{"a":1,}`, "MIT")
	f.Add(`{"license":"A","license":"B"}`, "MIT")
	f.Fuzz(func(t *testing.T, content, target string) {
		aligned, changed, err := AlignComposerLicense(content, target)
		if err != nil {
			return
		}
		surface, err := InspectComposerLicense(aligned)
		if err != nil || (surface.State != ComposerLicenseString && surface.State != ComposerLicenseArray) ||
			surface.Value != target || len(surface.Entries) > 1 {
			t.Fatalf("AlignComposerLicense produced an unprovable surface: %q -> %q (%q, %v)", content, aligned, surface.Value, err)
		}
		if !changed {
			return
		}
		again, changedAgain, err := AlignComposerLicense(aligned, target)
		if err != nil || changedAgain || again != aligned {
			t.Fatalf("AlignComposerLicense is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
		}
	})
}
