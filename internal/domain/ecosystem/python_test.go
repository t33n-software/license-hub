package ecosystem

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// Convention: docs/conventions/cli/testing/README.md

func TestPythonProjectionDerivesFromTheTenantValues(t *testing.T) {
	cases := []struct {
		name   string
		merged map[string]string
		want   string
	}{
		{"custom family", map[string]string{"LICENSE_ID": "example-NoRepublish-1.0"}, "LicenseRef-example-NoRepublish-1.0"},
		{"spdx identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "MIT"}, "MIT"},
		{"padded identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "  MIT  "}, "MIT"},
		{"spdx identifier wins", map[string]string{"LICENSE_ID": "example-NoRepublish-1.0", "SPDX_LICENSE_IDENTIFIER": "MIT"}, "MIT"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := PythonProjection(testCase.merged); got != testCase.want {
				t.Fatalf("PythonProjection() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestPythonLicenseFilesFormRendersTheManifestForm(t *testing.T) {
	if got := PythonLicenseFilesForm(PythonLicenseFilesTarget); got != `["LICENSE"]` {
		t.Fatalf("PythonLicenseFilesForm() = %q", got)
	}
	if got := PythonLicenseFilesForm([]string{"A", "B"}); got != `["A", "B"]` {
		t.Fatalf("PythonLicenseFilesForm() = %q", got)
	}
	if got := PythonLicenseFilesForm(nil); got != "[]" {
		t.Fatalf("PythonLicenseFilesForm() = %q", got)
	}
}

const examplePyproject = "[project]\nname = \"example-project\"\nlicense = \"MIT\"\n"

func TestInspectPythonLicenseClassifiesTheField(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantValue string
		wantState PythonLicenseFieldState
	}{
		{"string field", examplePyproject, "MIT", PythonLicenseString},
		{"missing field", "[project]\nname = \"x\"\n", "", PythonLicenseMissing},
		{"no project table", "tool = \"x\"\n", "", PythonLicenseMissing},
		{"no project table header form", "[tool.x]\nname = \"x\"\n", "", PythonLicenseMissing},
		{"empty document", "", "", PythonLicenseMissing},
		{"array table form", "[[project]]\nlicense = \"x\"\n", "", PythonLicenseMissing},
		{"escaped value", "[project]\nlicense = \"A\\u0042C\"\n", "ABC", PythonLicenseString},
		{"literal string value", "[project]\nlicense = 'LicenseRef-x'\n", "LicenseRef-x", PythonLicenseString},
		{"quoted key", "[project]\n\"license\" = \"MIT\"\n", "MIT", PythonLicenseString},
		{"quoted key with dot is a different key", "[project]\n\"license.x\" = \"MIT\"\n", "", PythonLicenseMissing},
		{"unicode escape", "[project]\nlicense = \"\\U0001F600\"\n", "\U0001F600", PythonLicenseString},
		{"multiline basic string elsewhere", "[project]\ndesc = \"\"\"a\\\n   b\"\"\"\nlicense = \"MIT\"\n", "MIT", PythonLicenseString},
		{"multiline basic with escapes", "[project]\ndesc = \"\"\"a\\tb\"\"\"\nlicense = \"MIT\"\n", "MIT", PythonLicenseString},
		{"multiline basic immediate newline", "[project]\ndesc = \"\"\"\nabc\n\"\"\"\nlicense = \"MIT\"\n", "MIT", PythonLicenseString},
		{"multiline basic immediate crlf", "[project]\ndesc = \"\"\"\r\nabc\r\n\"\"\"\nlicense = \"MIT\"\n", "MIT", PythonLicenseString},
		{"multiline literal immediate newline", "[project]\ndesc = '''\nabc\n'''\nlicense = \"MIT\"\n", "MIT", PythonLicenseString},
		{"multiline literal immediate crlf", "[project]\ndesc = '''\r\nabc\r\n'''\nlicense = \"MIT\"\n", "MIT", PythonLicenseString},
		{"lowercase unicode escape", "[project]\nlicense = \"\\u00e9\"\n", "é", PythonLicenseString},
		{"literal quoted key", "[project]\n'license' = \"MIT\"\n", "MIT", PythonLicenseString},
		{"trailing comment", "[project]\nlicense = \"MIT\"\n# trailing\n", "MIT", PythonLicenseString},
		{"multiline literal string elsewhere", "[project]\ndesc = '''\nraw text\n'''\nlicense = \"MIT\"\n", "MIT", PythonLicenseString},
		{"dotted key in another table", "[tool.x]\nproject.name = \"y\"\n\n[project]\nlicense = \"MIT\"\n", "MIT", PythonLicenseString},
		{"crlf line endings", "[project]\r\nlicense = \"MIT\"\r\n", "MIT", PythonLicenseString},
		{"header comment", "[project] # the metadata\nlicense = \"MIT\"\n", "MIT", PythonLicenseString},
		{"padded header", "[ project ]\nlicense = \"MIT\"\n", "MIT", PythonLicenseString},
		{"trailing comma array is valid toml", "[project]\nlicense-files = [\"LICENSE\",]\nlicense = \"MIT\"\n", "MIT", PythonLicenseString},
		{"table form", "[project]\nlicense = { text = \"MIT\" }\n", "", PythonLicenseTable},
		{"empty table form", "[project]\nlicense = { }\n", "", PythonLicenseTable},
		{"file table form", "[project]\nlicense = { file = \"LICENSE\" }\n", "", PythonLicenseTable},
		{"number field", "[project]\nlicense = 7\n", "", PythonLicenseInvalid},
		{"boolean field", "[project]\nlicense = true\n", "", PythonLicenseInvalid},
		{"array field", "[project]\nlicense = [\"MIT\"]\n", "", PythonLicenseInvalid},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectPythonLicense(testCase.content)
			if err != nil {
				t.Fatalf("InspectPythonLicense() error = %v", err)
			}
			if surface.State != testCase.wantState {
				t.Fatalf("InspectPythonLicense() state = %d, want %d", surface.State, testCase.wantState)
			}
			if surface.Value != testCase.wantValue {
				t.Fatalf("InspectPythonLicense() value = %q, want %q", surface.Value, testCase.wantValue)
			}
		})
	}
}

func TestInspectPythonLicenseClassifiesTheLicenseFilesField(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantSet   bool
		wantValid bool
		wantFiles []string
	}{
		{"absent", "[project]\nlicense = \"MIT\"\n", false, false, nil},
		{"target globs", "[project]\nlicense = \"MIT\"\nlicense-files = [\"LICENSE\"]\n", true, true, []string{"LICENSE"}},
		{"diverging globs", "[project]\nlicense-files = [\"A\", \"B\"]\n", true, true, []string{"A", "B"}},
		{"empty array", "[project]\nlicense-files = []\n", true, true, []string{}},
		{"multiline array with comments", "[project]\nlicense-files = [\n    \"LICENSE\",  # the root text\n    \"LICENSES/*\",\n]\n", true, true, []string{"LICENSE", "LICENSES/*"}},
		{"string form", "[project]\nlicense-files = \"LICENSE\"\n", true, false, nil},
		{"non-string item", "[project]\nlicense-files = [\"A\", 7]\n", true, false, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectPythonLicense(testCase.content)
			if err != nil {
				t.Fatalf("InspectPythonLicense() error = %v", err)
			}
			if surface.LicenseFilesSet != testCase.wantSet {
				t.Fatalf("InspectPythonLicense() licenseFilesSet = %v, want %v", surface.LicenseFilesSet, testCase.wantSet)
			}
			if surface.LicenseFilesValid != testCase.wantValid {
				t.Fatalf("InspectPythonLicense() licenseFilesValid = %v, want %v", surface.LicenseFilesValid, testCase.wantValid)
			}
			if !slices.Equal(surface.LicenseFiles, testCase.wantFiles) {
				t.Fatalf("InspectPythonLicense() licenseFiles = %v, want %v", surface.LicenseFiles, testCase.wantFiles)
			}
		})
	}
}

func TestInspectPythonLicenseGuardsTheDeprecatedClassifiers(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    []string
	}{
		{"absent", "[project]\nlicense = \"MIT\"\n", nil},
		{"without license entries", "[project]\nclassifiers = [\"Programming Language :: Python :: 3\"]\n", nil},
		{"with license entries", "[project]\nclassifiers = [\n    \"Programming Language :: Python :: 3\",\n    \"License :: OSI Approved :: MIT License\",\n]\n", []string{"License :: OSI Approved :: MIT License"}},
		{"non-array classifiers", "[project]\nclassifiers = \"x\"\n", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectPythonLicense(testCase.content)
			if err != nil {
				t.Fatalf("InspectPythonLicense() error = %v", err)
			}
			if !slices.Equal(surface.DeprecatedClassifiers, testCase.want) {
				t.Fatalf("InspectPythonLicense() deprecatedClassifiers = %v, want %v", surface.DeprecatedClassifiers, testCase.want)
			}
		})
	}
}

func TestInspectPythonLicenseRejectsUnscannableSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantErrIs error
	}{
		{"malformed header", "[project", ErrInvalidPythonSurface},
		{"trailing header content", "[project]]\n", ErrInvalidPythonSurface},
		{"malformed array header", "[[project]\n", ErrInvalidPythonSurface},
		{"empty header", "[]\n", ErrInvalidPythonSurface},
		{"header concatenation", "[\"pro\"+\"ject\"]\n", ErrInvalidPythonSurface},
		{"duplicate table", "[project]\n[project]\n", ErrInvalidPythonSurface},
		{"duplicate key", "[project]\nlicense = \"a\"\nlicense = \"b\"\n", ErrAmbiguousPythonField},
		{"duplicate other key", "[project]\nname = \"a\"\nname = \"b\"\n", ErrAmbiguousPythonField},
		{"root project key", "project = \"x\"\n[project]\n", ErrInvalidPythonSurface},
		{"root dotted project key", "project.license = \"x\"\n", ErrInvalidPythonSurface},
		{"dotted license shadow", "[project]\nlicense.x = 1\n", ErrInvalidPythonSurface},
		{"dotted files shadow", "[project]\nlicense-files.x = 1\n", ErrInvalidPythonSurface},
		{"unterminated string", "[project]\nlicense = \"unterminated\n", ErrInvalidPythonSurface},
		{"unterminated string at eof", "[project]\nlicense = \"x", ErrInvalidPythonSurface},
		{"control character", "[project]\nlicense = \"a\x00b\"\n", ErrInvalidPythonSurface},
		{"unknown escape", "[project]\nlicense = \"bad \\q escape\"\n", ErrInvalidPythonSurface},
		{"short unicode escape", "[project]\nlicense = \"\\u12\"\n", ErrInvalidPythonSurface},
		{"invalid unicode escape", "[project]\nlicense = \"\\uDEAD\"\n", ErrInvalidPythonSurface},
		{"short long unicode escape", "[project]\nlicense = \"\\U0001F6\"\n", ErrInvalidPythonSurface},
		{"escape at end", "[project]\nlicense = \"x\\\n", ErrInvalidPythonSurface},
		{"unterminated literal string", "[project]\nlicense = 'unterminated\n", ErrInvalidPythonSurface},
		{"unterminated literal string at eof", "[project]\nlicense = 'x", ErrInvalidPythonSurface},
		{"short unicode escape at eof", "[project]\nlicense = \"\\u1", ErrInvalidPythonSurface},
		{"unterminated multiline basic", "[project]\ndesc = \"\"\"abc\n", ErrInvalidPythonSurface},
		{"unterminated multiline basic at escape", "[project]\ndesc = \"\"\"a\\\n", ErrInvalidPythonSurface},
		{"unterminated multiline literal", "[project]\ndesc = '''abc\n", ErrInvalidPythonSurface},
		{"unterminated array", "[project]\nlicense-files = [\"a\",\n", ErrInvalidPythonSurface},
		{"array bad delimiter", "[project]\nlicense-files = [\"a\" \"b\"]\n", ErrInvalidPythonSurface},
		{"array unterminated string", "[project]\nlicense-files = [\"a\n", ErrInvalidPythonSurface},
		{"multiline inline table", "[project]\nlicense = {\n    text = \"MIT\"\n}\n", ErrInvalidPythonSurface},
		{"newline in nested inline value", "[project]\nlicense = { text = [\"a\",\n\"b\"] }\n", ErrInvalidPythonSurface},
		{"missing equals", "[project]\nlicense \"MIT\"\n", ErrInvalidPythonSurface},
		{"missing key", "[project]\n= \"MIT\"\n", ErrInvalidPythonSurface},
		{"trailing value content", "[project]\nlicense = \"MIT\" junk\n", ErrInvalidPythonSurface},
		{"missing value", "[project]\nlicense =\n", ErrInvalidPythonSurface},
		{"value at eof", "[project]\nlicense =", ErrInvalidPythonSurface},
		{"dotted key without value", "[project]\nlicense.\n", ErrInvalidPythonSurface},
		{"key at eof", "[project]\nkey.", ErrInvalidPythonSurface},
		{"unterminated quoted basic key", "[project]\n\"abc = 1\n", ErrInvalidPythonSurface},
		{"unterminated quoted literal key", "['abc\n", ErrInvalidPythonSurface},
		{"array unterminated after item", "[project]\nlicense-files = [\"a\"", ErrInvalidPythonSurface},
		{"inline table bad delimiter", "[project]\nlicense = { text = \"a\" junk }\n", ErrInvalidPythonSurface},
		{"inline table missing equals", "[project]\nlicense = { text \"a\" }\n", ErrInvalidPythonSurface},
		{"inline table missing value", "[project]\nlicense = { text = }\n", ErrInvalidPythonSurface},
		{"inline table bad key", "[project]\nlicense = { \"abc }\n", ErrInvalidPythonSurface},
		{"inline table eof after value", "[project]\nlicense = { text = \"a\"", ErrInvalidPythonSurface},
		{"inline table eof after comma", "[project]\nlicense = { text = \"a\",", ErrInvalidPythonSurface},
		{"inline table newline after comma", "[project]\nlicense = { text = \"a\",\n b = 1 }\n", ErrInvalidPythonSurface},
		{"inline table crlf after comma", "[project]\nlicense = { text = \"a\",\r\n b = 1 }\n", ErrInvalidPythonSurface},
		{"multiline basic unknown escape", "[project]\ndesc = \"\"\"a\\q\"\"\"\n", ErrInvalidPythonSurface},
		{"escape at eof", "[project]\nlicense = \"x\\", ErrInvalidPythonSurface},
		{"long unicode overflow", "[project]\nlicense = \"\\UFFFFFFFF\"\n", ErrInvalidPythonSurface},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := InspectPythonLicense(testCase.content)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("InspectPythonLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
			}
		})
	}
}

func TestAlignPythonLicenseReplacesADivergingValue(t *testing.T) {
	target := "LicenseRef-example-NoRepublish-1.0"
	aligned, changed, err := AlignPythonLicense(examplePyproject, target, PythonLicenseFilesTarget)
	if err != nil {
		t.Fatalf("AlignPythonLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignPythonLicense() changed = false, want true")
	}
	want := "[project]\nname = \"example-project\"\nlicense = \"" + target + "\"\nlicense-files = [\"LICENSE\"]\n"
	if aligned != want {
		t.Fatalf("AlignPythonLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignPythonLicenseAlignsTheDeprecatedTableForm(t *testing.T) {
	aligned, changed, err := AlignPythonLicense("[project]\nlicense = { text = \"MIT\" }\n", "MIT", PythonLicenseFilesTarget)
	if err != nil {
		t.Fatalf("AlignPythonLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignPythonLicense() changed = false, want true")
	}
	want := "[project]\nlicense = \"MIT\"\nlicense-files = [\"LICENSE\"]\n"
	if aligned != want {
		t.Fatalf("AlignPythonLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignPythonLicenseInsertsBothMissingFields(t *testing.T) {
	aligned, changed, err := AlignPythonLicense("[project]\nname = \"example-project\"\n", "MIT", PythonLicenseFilesTarget)
	if err != nil {
		t.Fatalf("AlignPythonLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignPythonLicense() changed = false, want true")
	}
	want := "[project]\nlicense = \"MIT\"\nlicense-files = [\"LICENSE\"]\nname = \"example-project\"\n"
	if aligned != want {
		t.Fatalf("AlignPythonLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignPythonLicenseInsertsOnlyTheMissingField(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{"license present", "[project]\nlicense = \"MIT\"\nname = \"x\"\n", "[project]\nlicense = \"MIT\"\nlicense-files = [\"LICENSE\"]\nname = \"x\"\n"},
		{"files present", "[project]\nlicense-files = [\"LICENSE\"]\nname = \"x\"\n", "[project]\nlicense = \"MIT\"\nlicense-files = [\"LICENSE\"]\nname = \"x\"\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignPythonLicense(testCase.content, "MIT", PythonLicenseFilesTarget)
			if err != nil {
				t.Fatalf("AlignPythonLicense() error = %v", err)
			}
			if !changed {
				t.Fatal("AlignPythonLicense() changed = false, want true")
			}
			if aligned != testCase.want {
				t.Fatalf("AlignPythonLicense() = %q, want %q", aligned, testCase.want)
			}
			surface, err := InspectPythonLicense(aligned)
			if err != nil || surface.State != PythonLicenseString || surface.Value != "MIT" {
				t.Fatalf("AlignPythonLicense() produced an unprovable surface: %q (%v)", aligned, err)
			}
		})
	}
}

func TestAlignPythonLicenseInsertsAfterAnUnterminatedHeader(t *testing.T) {
	aligned, changed, err := AlignPythonLicense("[project]", "MIT", PythonLicenseFilesTarget)
	if err != nil {
		t.Fatalf("AlignPythonLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignPythonLicense() changed = false, want true")
	}
	want := "[project]\nlicense = \"MIT\"\nlicense-files = [\"LICENSE\"]\n"
	if aligned != want {
		t.Fatalf("AlignPythonLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignPythonLicenseInsertsAfterAnUnterminatedLicenseLine(t *testing.T) {
	aligned, changed, err := AlignPythonLicense("[project]\nlicense = \"MIT\"", "MIT", PythonLicenseFilesTarget)
	if err != nil {
		t.Fatalf("AlignPythonLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignPythonLicense() changed = false, want true")
	}
	want := "[project]\nlicense = \"MIT\"\nlicense-files = [\"LICENSE\"]\n"
	if aligned != want {
		t.Fatalf("AlignPythonLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignPythonLicenseIsIdempotentOnAlignedState(t *testing.T) {
	alignedInput := "[project]\nname = \"x\"\nlicense = \"MIT\"\nlicense-files = [\"LICENSE\"]\n"
	aligned, changed, err := AlignPythonLicense(alignedInput, "MIT", PythonLicenseFilesTarget)
	if err != nil {
		t.Fatalf("AlignPythonLicense() error = %v", err)
	}
	if changed {
		t.Fatal("AlignPythonLicense() changed = true, want false")
	}
	if aligned != alignedInput {
		t.Fatalf("AlignPythonLicense() mutated an aligned manifest: %q", aligned)
	}
}

func TestAlignPythonLicensePreservesEveryNonLicenseByte(t *testing.T) {
	content := "[project]\nname = \"example-project\"\nlicense = \"MIT\"  # trailing\nauthors = [\"x\"]\n" +
		"classifiers = [\n    \"License :: OSI Approved :: MIT License\",\n]\n"
	aligned, _, err := AlignPythonLicense(content, "MIT", PythonLicenseFilesTarget)
	if err != nil {
		t.Fatalf("AlignPythonLicense() error = %v", err)
	}
	if !strings.Contains(aligned, "name = \"example-project\"\n") {
		t.Fatalf("AlignPythonLicense() touched a non-license field: %q", aligned)
	}
	if !strings.Contains(aligned, "license = \"MIT\"  # trailing\n") {
		t.Fatalf("AlignPythonLicense() lost the trailing comment: %q", aligned)
	}
	if !strings.Contains(aligned, "authors = [\"x\"]\n") {
		t.Fatalf("AlignPythonLicense() touched the authors field: %q", aligned)
	}
	if !strings.Contains(aligned, "classifiers = [\n    \"License :: OSI Approved :: MIT License\",\n]\n") {
		t.Fatalf("AlignPythonLicense() rewrote the classifiers array: %q", aligned)
	}
	if !strings.Contains(aligned, "license-files = [\"LICENSE\"]\n") {
		t.Fatalf("AlignPythonLicense() lost the inserted files field: %q", aligned)
	}
}

func TestAlignPythonLicenseQuotesEveryEscapeForm(t *testing.T) {
	target := "a\"b\\c\bd\te\nf\fg\rh\u0001i"
	aligned, changed, err := AlignPythonLicense("[project]\nlicense = \"MIT\"\n", target, PythonLicenseFilesTarget)
	if err != nil {
		t.Fatalf("AlignPythonLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignPythonLicense() changed = false, want true")
	}
	want := "[project]\nlicense = \"a\\\"b\\\\c\\bd\\te\\nf\\fg\\rh\\u0001i\"\nlicense-files = [\"LICENSE\"]\n"
	if aligned != want {
		t.Fatalf("AlignPythonLicense() = %q, want %q", aligned, want)
	}
	surface, err := InspectPythonLicense(aligned)
	if err != nil {
		t.Fatalf("InspectPythonLicense() error = %v", err)
	}
	if surface.State != PythonLicenseString || surface.Value != target {
		t.Fatalf("InspectPythonLicense() = (%d, %q), want the decoded round trip", surface.State, surface.Value)
	}
}

func TestAlignPythonLicenseRefusesInvalidSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		target    string
		files     []string
		wantErrIs error
	}{
		{"no project table", "tool = \"x\"\n", "MIT", PythonLicenseFilesTarget, ErrNoProjectTable},
		{"array table only", "[[project]]\nlicense = \"x\"\n", "MIT", PythonLicenseFilesTarget, ErrNoProjectTable},
		{"unscannable surface", "[project", "MIT", PythonLicenseFilesTarget, ErrInvalidPythonSurface},
		{"invalid license value", "[project]\nlicense = 7\n", "MIT", PythonLicenseFilesTarget, ErrInvalidPythonLicenseValue},
		{"array license value", "[project]\nlicense = [\"MIT\"]\n", "MIT", PythonLicenseFilesTarget, ErrInvalidPythonLicenseValue},
		{"files not an array", "[project]\nlicense-files = \"LICENSE\"\n", "MIT", PythonLicenseFilesTarget, ErrInvalidPythonLicenseFiles},
		{"files non-string item", "[project]\nlicense-files = [\"A\", 7]\n", "MIT", PythonLicenseFilesTarget, ErrInvalidPythonLicenseFiles},
		{"duplicate license key", "[project]\nlicense = \"a\"\nlicense = \"b\"\n", "MIT", PythonLicenseFilesTarget, ErrAmbiguousPythonField},
		{"non-utf-8 target", examplePyproject, "MIT\xff", PythonLicenseFilesTarget, ErrInvalidTarget},
		{"non-utf-8 file", examplePyproject, "MIT", []string{"LICENSE\xff"}, ErrInvalidTarget},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := AlignPythonLicense(testCase.content, testCase.target, testCase.files)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("AlignPythonLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
			}
		})
	}
}

func FuzzAlignPythonLicense(f *testing.F) {
	f.Add(examplePyproject, "LicenseRef-example-NoRepublish-1.0", "LICENSE")
	f.Add("[project]\nname = \"x\"\n", "MIT", "LICENSE")
	f.Add("[project]", "MIT", "LICENSE")
	f.Add("[project\n", "MIT", "LICENSE")
	f.Add("[project]\nlicense = {text=\"x\"}\n", "MIT", "LICENSE")
	f.Add("[project]\nlicense = 7\n", "MIT", "LICENSE")
	f.Add("[project]\nlicense-files = [\"A\"]\n", "MIT", "LICENSE,LICENSES/*")
	f.Add("[[project]]\nlicense = \"x\"\n", "MIT", "LICENSE")
	f.Add("[project]\nlicense = \"a\"\nlicense = \"b\"\n", "MIT", "LICENSE")
	f.Add("", "MIT", "LICENSE")
	f.Add("[project]\nclassifiers = [\"License :: X\"]\n", "MIT", "LICENSE")
	f.Add("[project]\nlicense = \"\"\"multi\nline\"\"\"\n", "MIT", "LICENSE")
	f.Fuzz(func(t *testing.T, content, target, files string) {
		list := strings.Split(files, ",")
		aligned, changed, err := AlignPythonLicense(content, target, list)
		if err != nil {
			return
		}
		surface, err := InspectPythonLicense(aligned)
		if err != nil {
			t.Fatalf("AlignPythonLicense produced an unprovable surface: %q -> %q (%v)", content, aligned, err)
		}
		if surface.State != PythonLicenseString || surface.Value != target {
			t.Fatalf("AlignPythonLicense did not align the license field: %q -> %q (%q, %d)", content, aligned, surface.Value, surface.State)
		}
		if !surface.LicenseFilesSet || !surface.LicenseFilesValid || !slices.Equal(surface.LicenseFiles, list) {
			t.Fatalf("AlignPythonLicense did not align the license-files field: %q -> %q", content, aligned)
		}
		if !changed {
			return
		}
		again, changedAgain, err := AlignPythonLicense(aligned, target, list)
		if err != nil || changedAgain || again != aligned {
			t.Fatalf("AlignPythonLicense is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
		}
	})
}
