package ecosystem

import (
	"errors"
	"strings"
	"testing"
)

// Convention: docs/conventions/cli/testing/README.md

func TestSeamLanguageExtractsTheDeclaredLanguage(t *testing.T) {
	cases := []struct {
		name string
		data string
		want string
	}{
		{"node territory", `{"schemaVersion":4,"toolchain":{"language":"node-typescript","version":"26.10.0"}}`, NpmLanguage},
		{"go territory", `{"schemaVersion":4,"toolchain":{"language":"go","version":"1.26.6"}}`, "go"},
		{"missing toolchain", `{"schemaVersion":4,"extends":[]}`, ""},
		{"padded language", `{"toolchain":{"language":"  node-typescript  "}}`, NpmLanguage},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := SeamLanguage([]byte(testCase.data))
			if err != nil {
				t.Fatalf("SeamLanguage() error = %v", err)
			}
			if got != testCase.want {
				t.Fatalf("SeamLanguage() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestSeamLanguageRejectsInvalidDocuments(t *testing.T) {
	if _, err := SeamLanguage([]byte("{")); !errors.Is(err, ErrInvalidSurface) {
		t.Fatalf("SeamLanguage() error = %v, want ErrInvalidSurface", err)
	}
}

func TestNpmProjectionDerivesFromTheTenantValues(t *testing.T) {
	cases := []struct {
		name   string
		merged map[string]string
		want   string
	}{
		{"custom family", map[string]string{"LICENSE_ID": "example-NoRepublish-1.0"}, NpmTargetForm},
		{"spdx identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "MIT"}, "MIT"},
		{"padded identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "  MIT  "}, "MIT"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := NpmProjection(testCase.merged); got != testCase.want {
				t.Fatalf("NpmProjection() = %q, want %q", got, testCase.want)
			}
		})
	}
}

const exampleManifest = "{\n  \"name\": \"example-project\",\n" +
	"  \"version\": \"1.0.0\",\n  \"license\": \"MIT\",\n" +
	"  \"author\": \"t33n Software\"\n}\n"

func TestInspectNpmLicenseClassifiesTheField(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantValue string
		wantState LicenseFieldState
		wantErrIs error
	}{
		{"string field", exampleManifest, "MIT", LicenseFieldString, nil},
		{"missing field", `{"name":"x"}`, "", LicenseFieldMissing, nil},
		{"escaped value", `{"license":"A\u0042C"}`, "ABC", LicenseFieldString, nil},
		{"number field", `{"license":7}`, "", LicenseFieldNonString, nil},
		{"object field", `{"license":{"a":1}}`, "", LicenseFieldNonString, nil},
		{"array field", `{"license":["MIT"]}`, "", LicenseFieldNonString, nil},
		{"boolean field", `{"license":true}`, "", LicenseFieldNonString, nil},
		{"null field", `{"license":null}`, "", LicenseFieldNonString, nil},
		{"empty object", `{}`, "", LicenseFieldMissing, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			value, state, err := InspectNpmLicense(testCase.content)
			if testCase.wantErrIs != nil {
				if !errors.Is(err, testCase.wantErrIs) {
					t.Fatalf("InspectNpmLicense() error = %v, want %v", err, testCase.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("InspectNpmLicense() error = %v", err)
			}
			if state != testCase.wantState {
				t.Fatalf("InspectNpmLicense() state = %d, want %d", state, testCase.wantState)
			}
			if value != testCase.wantValue {
				t.Fatalf("InspectNpmLicense() value = %q, want %q", value, testCase.wantValue)
			}
		})
	}
}

func TestInspectNpmLicenseRejectsUnscannableSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantErrIs error
	}{
		{"malformed json", `{"license":`, ErrInvalidSurface},
		{"root not an object", `[1]`, ErrInvalidSurface},
		{"root not an object string", `"x"`, ErrInvalidSurface},
		{"empty document", ``, ErrInvalidSurface},
		{"whitespace only", `   `, ErrInvalidSurface},
		{"duplicate member", `{"license":"A","license":"B"}`, ErrAmbiguousField},
		{"trailing comma", `{"a":1,}`, ErrInvalidSurface},
		{"missing colon", `{"a" 1}`, ErrInvalidSurface},
		{"key not a string", `{1:2}`, ErrInvalidSurface},
		{"unterminated string", `{"a`, ErrInvalidSurface},
		{"unterminated object", `{"a":1`, ErrInvalidSurface},
		{"unterminated array", `{"a":[1`, ErrInvalidSurface},
		{"bad escape", `{"a":"\x"}`, ErrInvalidSurface},
		{"short unicode escape", `{"a":"\u12"}`, ErrInvalidSurface},
		{"escape at end", `{"a":"\`, ErrInvalidSurface},
		{"bad literal", `{"a":tru}`, ErrInvalidSurface},
		{"leading zero number", `{"a":01}`, ErrInvalidSurface},
		{"plus number", `{"a":+1}`, ErrInvalidSurface},
		{"trailing dot number", `{"a":1.}`, ErrInvalidSurface},
		{"bare exponent", `{"a":1e}`, ErrInvalidSurface},
		{"missing value", `{"a":}`, ErrInvalidSurface},
		{"missing delimiter", `{"a":1 "b":2}`, ErrInvalidSurface},
		{"truncated after comma", `{"a":1,`, ErrInvalidSurface},
		{"array trailing comma", `{"a":[1,]}`, ErrInvalidSurface},
		{"array missing delimiter", `{"a":[1 2]}`, ErrInvalidSurface},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := InspectNpmLicense(testCase.content)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("InspectNpmLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
			}
		})
	}
}

func TestAlignNpmLicenseReplacesADivergingValue(t *testing.T) {
	aligned, changed, err := AlignNpmLicense(exampleManifest, NpmTargetForm)
	if err != nil {
		t.Fatalf("AlignNpmLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignNpmLicense() changed = false, want true")
	}
	want := strings.Replace(exampleManifest, `"license": "MIT"`, `"license": "SEE LICENSE IN LICENSE"`, 1)
	if aligned != want {
		t.Fatalf("AlignNpmLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignNpmLicenseIsIdempotentOnAlignedState(t *testing.T) {
	aligned, changed, err := AlignNpmLicense(exampleManifest, "MIT")
	if err != nil {
		t.Fatalf("AlignNpmLicense() error = %v", err)
	}
	if changed {
		t.Fatal("AlignNpmLicense() changed = true, want false")
	}
	if aligned != exampleManifest {
		t.Fatalf("AlignNpmLicense() mutated an aligned manifest: %q", aligned)
	}
}

func TestAlignNpmLicensePreservesEveryNonLicenseByte(t *testing.T) {
	content := "{\n  \"name\": \"example-project\",\n  \"license\": \"MIT\",\n" +
		"  \"author\": \"t33n Software\",\n  \"private\": true\n}\n"
	aligned, _, err := AlignNpmLicense(content, NpmTargetForm)
	if err != nil {
		t.Fatalf("AlignNpmLicense() error = %v", err)
	}
	before, _, _ := strings.Cut(content, `"license": "MIT"`)
	afterOriginal := content[strings.Index(content, `"license": "MIT"`)+len(`"license": "MIT"`):]
	if !strings.HasPrefix(aligned, before+`"license": "SEE LICENSE IN LICENSE"`) {
		t.Fatalf("AlignNpmLicense() changed bytes before the value: %q", aligned)
	}
	if !strings.HasSuffix(aligned, afterOriginal) {
		t.Fatalf("AlignNpmLicense() changed bytes after the value: %q", aligned)
	}
	if !strings.Contains(aligned, "\"author\": \"t33n Software\"") {
		t.Fatalf("AlignNpmLicense() touched a non-license field: %q", aligned)
	}
}

func TestAlignNpmLicenseInsertsAMissingMember(t *testing.T) {
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
			aligned, changed, err := AlignNpmLicense(testCase.content, NpmTargetForm)
			if err != nil {
				t.Fatalf("AlignNpmLicense() error = %v", err)
			}
			if !changed {
				t.Fatal("AlignNpmLicense() changed = false, want true")
			}
			value, state, err := InspectNpmLicense(aligned)
			if err != nil || state != LicenseFieldString {
				t.Fatalf("AlignNpmLicense() produced an unscannable surface: %q (%v)", aligned, err)
			}
			if value != NpmTargetForm {
				t.Fatalf("AlignNpmLicense() license = %q, want %q", value, NpmTargetForm)
			}
			if testCase.name != "with members" && !strings.Contains(aligned, "license") {
				t.Fatalf("AlignNpmLicense() lost the license member: %q", aligned)
			}
			if testCase.name == "with members" && !strings.Contains(aligned, "\"name\": \"example-project\"") {
				t.Fatalf("AlignNpmLicense() lost the original member: %q", aligned)
			}
		})
	}
}

func TestAlignNpmLicenseRefusesNonStringAndAmbiguousSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantErrIs error
	}{
		{"number field", `{"license":7}`, ErrNonStringField},
		{"object field", `{"license":{"type":"MIT"}}`, ErrNonStringField},
		{"duplicate member", `{"license":"A","license":"B"}`, ErrAmbiguousField},
		{"malformed json", `{"license":`, ErrInvalidSurface},
		{"root not an object", `[1]`, ErrInvalidSurface},
		{"non-utf-8 target", exampleManifest, ErrInvalidTarget},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			target := NpmTargetForm
			if testCase.name == "non-utf-8 target" {
				target = "MIT\xff"
			}
			_, _, err := AlignNpmLicense(testCase.content, target)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("AlignNpmLicense(%q, %q) error = %v, want %v", testCase.content, target, err, testCase.wantErrIs)
			}
		})
	}
}

func TestParseValueValidatesEveryValueKind(t *testing.T) {
	cases := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"string", `{"a":"b"}`, false},
		{"nested object", `{"a":{"b":{"c":1}}}`, false},
		{"nested array", `{"a":[1,[2,{"c":null}]]}`, false},
		{"empty array", `{"a":[]}`, false},
		{"literals", `{"a":true,"b":false,"c":null}`, false},
		{"numbers", `{"a":-1,"b":1.5e3,"c":0}`, false},
		{"escaped strings", `{"a":"tab\tquote\"slash\\solid\/back\b\ff\n\r"}`, false},
		{"unicode escape", `{"a":"\u0041\u00e9"}`, false},
		{"array missing close", `{"a":[1`, true},
		{"array bad literal", `{"a":[tru]}`, true},
		{"array stray delimiter", `{"a":[1 2]}`, true},
		{"object missing close", `{"a":{"b":1`, true},
		{"object missing colon", `{"a":{"b" 1}}`, true},
		{"object key not string", `{"a":{1:2}}`, true},
		{"object trailing comma", `{"a":{"b":1,}}`, true},
		{"array trailing comma", `{"a":[1,]}`, true},
		{"bad literal", `{"a":truex}`, true},
		{"false literal", `{"a":fals}`, true},
		{"null literal", `{"a":nul}`, true},
		{"number missing digits", `{"a":-}`, true},
		{"hex escape short", `{"a":"\u00"}`, true},
		{"escape unknown", `{"a":"\q"}`, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := parseDocument(testCase.content)
			if testCase.wantErr && !errors.Is(err, ErrInvalidSurface) {
				t.Fatalf("parseDocument(%q) error = %v, want ErrInvalidSurface", testCase.content, err)
			}
			if !testCase.wantErr && err != nil {
				t.Fatalf("parseDocument(%q) error = %v", testCase.content, err)
			}
		})
	}
}

func TestParseDocumentRecordsRootSpans(t *testing.T) {
	members, obj, err := parseDocument("  {\"a\": 1, \"b\": \"x\"}  \n")
	if err != nil {
		t.Fatalf("parseDocument() error = %v", err)
	}
	if obj.start != 2 || obj.end != len("  {\"a\": 1, \"b\": \"x\"}") {
		t.Fatalf("parseDocument() span = %+v", obj)
	}
	if len(members) != 2 {
		t.Fatalf("parseDocument() members = %+v", members)
	}
	if members[0].key != "a" || members[0].isString || members[0].valueStart != 8 || members[0].valueEnd != 9 {
		t.Fatalf("parseDocument() member[0] = %+v", members[0])
	}
	if members[1].key != "b" || !members[1].isString || members[1].str != "x" {
		t.Fatalf("parseDocument() member[1] = %+v", members[1])
	}
}

func FuzzAlignNpmLicense(f *testing.F) {
	f.Add(`{"name":"x","license":"MIT"}`, "SEE LICENSE IN LICENSE")
	f.Add(`{}`, "MIT")
	f.Add(`{"license":`, "MIT")
	f.Add(`{"license":7}`, "MIT")
	f.Add(`[1]`, "MIT")
	f.Add(`{"a":01}`, "MIT")
	f.Add(`{"a":"\u12"}`, "MIT")
	f.Add(`{"a":1,}`, "MIT")
	f.Add(`{"license":"A","license":"B"}`, "MIT")
	f.Fuzz(func(t *testing.T, content, target string) {
		aligned, changed, err := AlignNpmLicense(content, target)
		if err != nil {
			return
		}
		value, state, err := InspectNpmLicense(aligned)
		if err != nil || state != LicenseFieldString || value != target {
			t.Fatalf("AlignNpmLicense produced an unprovable surface: %q -> %q (%q, %v)", content, aligned, value, err)
		}
		if !changed {
			return
		}
		again, changedAgain, err := AlignNpmLicense(aligned, target)
		if err != nil || changedAgain || again != aligned {
			t.Fatalf("AlignNpmLicense is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
		}
	})
}
