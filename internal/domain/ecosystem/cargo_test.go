package ecosystem

import (
	"errors"
	"strings"
	"testing"
)

// Convention: docs/conventions/cli/testing/README.md

func TestCargoProjectionDerivesFromTheTenantValues(t *testing.T) {
	cases := []struct {
		name   string
		merged map[string]string
		want   CargoLicenseForm
	}{
		{"custom family", map[string]string{"LICENSE_ID": "example-NoRepublish-1.0"}, CargoLicenseForm{Field: CargoLicenseFileKey, Value: "LICENSE"}},
		{"spdx identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "MIT"}, CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"}},
		{"padded identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "  MIT  "}, CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"}},
		{"spdx identifier wins", map[string]string{"LICENSE_ID": "example-NoRepublish-1.0", "SPDX_LICENSE_IDENTIFIER": "MIT"}, CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := CargoProjection(testCase.merged)
			if got.Field != testCase.want.Field || got.Value != testCase.want.Value {
				t.Fatalf("CargoProjection() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestCargoProjectionRendersTheManifestForm(t *testing.T) {
	if got := (CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"}).String(); got != `license = "MIT"` {
		t.Fatalf("String() = %q", got)
	}
	if got := (CargoLicenseForm{Field: CargoLicenseFileKey, Value: "LICENSE"}).String(); got != `license-file = "LICENSE"` {
		t.Fatalf("String() = %q", got)
	}
}

const exampleCargoToml = "[package]\nname = \"example-project\"\nlicense = \"MIT\"\n"

func TestInspectCargoLicenseClassifiesTheField(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantValue string
		wantState CargoLicenseFieldState
	}{
		{"expression field", exampleCargoToml, "MIT", CargoLicenseExpression},
		{"license-file field", "[package]\nlicense-file = \"LICENSE\"\n", "LICENSE", CargoLicenseFileForm},
		{"missing field", "[package]\nname = \"x\"\n", "", CargoLicenseMissing},
		{"no package table", "tool = \"x\"\n", "", CargoLicenseMissing},
		{"no package table header form", "[tool.x]\nname = \"x\"\n", "", CargoLicenseMissing},
		{"empty document", "", "", CargoLicenseMissing},
		{"array table form", "[[package]]\nlicense = \"x\"\n", "", CargoLicenseMissing},
		{"escaped value", "[package]\nlicense = \"A\\u0042C\"\n", "ABC", CargoLicenseExpression},
		{"literal string value", "[package]\nlicense = 'MIT OR Apache-2.0'\n", "MIT OR Apache-2.0", CargoLicenseExpression},
		{"quoted key", "[package]\n\"license\" = \"MIT\"\n", "MIT", CargoLicenseExpression},
		{"quoted key with dot is a different key", "[package]\n\"license.x\" = \"MIT\"\n", "", CargoLicenseMissing},
		{"unicode escape", "[package]\nlicense = \"\\U0001F600\"\n", "\U0001F600", CargoLicenseExpression},
		{"multiline basic string elsewhere", "[package]\ndesc = \"\"\"a\\\n   b\"\"\"\nlicense = \"MIT\"\n", "MIT", CargoLicenseExpression},
		{"multiline literal string elsewhere", "[package]\ndesc = '''\nraw text\n'''\nlicense = \"MIT\"\n", "MIT", CargoLicenseExpression},
		{"dotted key in another table", "[tool.x]\npackage.name = \"y\"\n\n[package]\nlicense = \"MIT\"\n", "MIT", CargoLicenseExpression},
		{"crlf line endings", "[package]\r\nlicense = \"MIT\"\r\n", "MIT", CargoLicenseExpression},
		{"header comment", "[package] # the metadata\nlicense = \"MIT\"\n", "MIT", CargoLicenseExpression},
		{"padded header", "[ package ]\nlicense = \"MIT\"\n", "MIT", CargoLicenseExpression},
		{"trailing comma array is valid toml", "[package]\nkeywords = [\"x\",]\nlicense = \"MIT\"\n", "MIT", CargoLicenseExpression},
		{"number field", "[package]\nlicense = 7\n", "", CargoLicenseInvalid},
		{"boolean field", "[package]\nlicense = true\n", "", CargoLicenseInvalid},
		{"array field", "[package]\nlicense = [\"MIT\"]\n", "", CargoLicenseInvalid},
		{"number license-file field", "[package]\nlicense-file = 7\n", "", CargoLicenseInvalid},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectCargoLicense(testCase.content)
			if err != nil {
				t.Fatalf("InspectCargoLicense() error = %v", err)
			}
			if surface.State != testCase.wantState {
				t.Fatalf("InspectCargoLicense() state = %d, want %d", surface.State, testCase.wantState)
			}
			if surface.Value != testCase.wantValue {
				t.Fatalf("InspectCargoLicense() value = %q, want %q", surface.Value, testCase.wantValue)
			}
		})
	}
}

func TestInspectCargoLicenseRefusesUnprovableSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantErrIs error
	}{
		{"both exclusive keys", "[package]\nlicense = \"MIT\"\nlicense-file = \"LICENSE\"\n", ErrExclusiveCargoLicense},
		{"malformed header", "[package", ErrInvalidTOMLSurface},
		{"trailing header content", "[package]]\n", ErrInvalidTOMLSurface},
		{"duplicate table", "[package]\n[package]\n", ErrInvalidTOMLSurface},
		{"duplicate key", "[package]\nlicense = \"a\"\nlicense = \"b\"\n", ErrAmbiguousCargoField},
		{"duplicate other key", "[package]\nname = \"a\"\nname = \"b\"\n", ErrAmbiguousCargoField},
		{"root package key", "package = \"x\"\n[package]\n", ErrInvalidCargoSurface},
		{"root dotted package key", "package.license = \"x\"\n", ErrInvalidCargoSurface},
		{"dotted license shadow", "[package]\nlicense.x = 1\n", ErrInvalidCargoSurface},
		{"dotted license-file shadow", "[package]\nlicense-file.x = 1\n", ErrInvalidCargoSurface},
		{"unterminated string", "[package]\nlicense = \"unterminated\n", ErrInvalidTOMLSurface},
		{"missing equals", "[package]\nlicense \"MIT\"\n", ErrInvalidTOMLSurface},
		{"trailing value content", "[package]\nlicense = \"MIT\" junk\n", ErrInvalidTOMLSurface},
		{"multiline inline table", "[package]\nlicense = {\n    text = \"MIT\"\n}\n", ErrInvalidTOMLSurface},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := InspectCargoLicense(testCase.content)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("InspectCargoLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
			}
		})
	}
}

func TestAlignCargoLicenseReplacesADivergingValue(t *testing.T) {
	aligned, changed, err := AlignCargoLicense(exampleCargoToml, CargoLicenseForm{Field: CargoLicenseField, Value: "LicenseRef-example-NoRepublish-1.0"})
	if err != nil {
		t.Fatalf("AlignCargoLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignCargoLicense() changed = false, want true")
	}
	want := "[package]\nname = \"example-project\"\nlicense = \"LicenseRef-example-NoRepublish-1.0\"\n"
	if aligned != want {
		t.Fatalf("AlignCargoLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignCargoLicenseReplacesADivergingFileValue(t *testing.T) {
	aligned, changed, err := AlignCargoLicense("[package]\nname = \"x\"\nlicense-file = \"COPYING\"\n", CargoLicenseForm{Field: CargoLicenseFileKey, Value: "LICENSE"})
	if err != nil {
		t.Fatalf("AlignCargoLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignCargoLicense() changed = false, want true")
	}
	want := "[package]\nname = \"x\"\nlicense-file = \"LICENSE\"\n"
	if aligned != want {
		t.Fatalf("AlignCargoLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignCargoLicenseSwitchesTheExpressionFormToTheFileForm(t *testing.T) {
	aligned, changed, err := AlignCargoLicense(exampleCargoToml, CargoLicenseForm{Field: CargoLicenseFileKey, Value: "LICENSE"})
	if err != nil {
		t.Fatalf("AlignCargoLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignCargoLicense() changed = false, want true")
	}
	want := "[package]\nname = \"example-project\"\nlicense-file = \"LICENSE\"\n"
	if aligned != want {
		t.Fatalf("AlignCargoLicense() = %q, want %q", aligned, want)
	}
	surface, err := InspectCargoLicense(aligned)
	if err != nil || surface.State != CargoLicenseFileForm || surface.Value != "LICENSE" {
		t.Fatalf("AlignCargoLicense() produced an unprovable surface: %q (%v)", aligned, err)
	}
}

func TestAlignCargoLicenseSwitchesTheFileFormToTheExpressionForm(t *testing.T) {
	aligned, changed, err := AlignCargoLicense("[package]\nname = \"x\"\nlicense-file = \"LICENSE\"  # the text\n", CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"})
	if err != nil {
		t.Fatalf("AlignCargoLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignCargoLicense() changed = false, want true")
	}
	want := "[package]\nname = \"x\"\nlicense = \"MIT\"  # the text\n"
	if aligned != want {
		t.Fatalf("AlignCargoLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignCargoLicensePreservesTheKeySpacingOnAValueChange(t *testing.T) {
	aligned, changed, err := AlignCargoLicense("[package]\nlicense=\"Apache-2.0\"\n", CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"})
	if err != nil {
		t.Fatalf("AlignCargoLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignCargoLicense() changed = false, want true")
	}
	want := "[package]\nlicense=\"MIT\"\n"
	if aligned != want {
		t.Fatalf("AlignCargoLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignCargoLicenseInsertsTheMissingField(t *testing.T) {
	cases := []struct {
		name    string
		content string
		form    CargoLicenseForm
		want    string
	}{
		{"expression form after header", "[package]\nname = \"example-project\"\n", CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"}, "[package]\nlicense = \"MIT\"\nname = \"example-project\"\n"},
		{"file form after header", "[package]\nname = \"example-project\"\n", CargoLicenseForm{Field: CargoLicenseFileKey, Value: "LICENSE"}, "[package]\nlicense-file = \"LICENSE\"\nname = \"example-project\"\n"},
		{"unterminated header", "[package]", CargoLicenseForm{Field: CargoLicenseFileKey, Value: "LICENSE"}, "[package]\nlicense-file = \"LICENSE\"\n"},
		{"unterminated license line", "[package]\nlicense = \"MIT\"", CargoLicenseForm{Field: CargoLicenseFileKey, Value: "LICENSE"}, "[package]\nlicense-file = \"LICENSE\""},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignCargoLicense(testCase.content, testCase.form)
			if err != nil {
				t.Fatalf("AlignCargoLicense() error = %v", err)
			}
			if !changed {
				t.Fatal("AlignCargoLicense() changed = false, want true")
			}
			if aligned != testCase.want {
				t.Fatalf("AlignCargoLicense() = %q, want %q", aligned, testCase.want)
			}
		})
	}
}

func TestAlignCargoLicenseIsIdempotentOnAlignedState(t *testing.T) {
	cases := []struct {
		name    string
		content string
		form    CargoLicenseForm
	}{
		{"expression form", "[package]\nname = \"x\"\nlicense = \"MIT\"\n", CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"}},
		{"file form", "[package]\nname = \"x\"\nlicense-file = \"LICENSE\"\n", CargoLicenseForm{Field: CargoLicenseFileKey, Value: "LICENSE"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignCargoLicense(testCase.content, testCase.form)
			if err != nil {
				t.Fatalf("AlignCargoLicense() error = %v", err)
			}
			if changed {
				t.Fatal("AlignCargoLicense() changed = true, want false")
			}
			if aligned != testCase.content {
				t.Fatalf("AlignCargoLicense() mutated an aligned manifest: %q", aligned)
			}
		})
	}
}

func TestAlignCargoLicensePreservesEveryNonLicenseByte(t *testing.T) {
	content := "[package]\nname = \"example-project\"\nlicense = \"MIT\"  # trailing\nauthors = [\"x\"]\n"
	aligned, _, err := AlignCargoLicense(content, CargoLicenseForm{Field: CargoLicenseFileKey, Value: "LICENSE"})
	if err != nil {
		t.Fatalf("AlignCargoLicense() error = %v", err)
	}
	if !strings.Contains(aligned, "name = \"example-project\"\n") {
		t.Fatalf("AlignCargoLicense() touched a non-license field: %q", aligned)
	}
	if !strings.Contains(aligned, "authors = [\"x\"]\n") {
		t.Fatalf("AlignCargoLicense() touched the authors field: %q", aligned)
	}
	if !strings.Contains(aligned, "license-file = \"LICENSE\"  # trailing\n") {
		t.Fatalf("AlignCargoLicense() lost the trailing comment: %q", aligned)
	}
}

func TestAlignCargoLicenseQuotesEveryEscapeForm(t *testing.T) {
	target := "a\"b\\c\bd\te\nf\fg\rh\u0001i"
	aligned, changed, err := AlignCargoLicense(exampleCargoToml, CargoLicenseForm{Field: CargoLicenseField, Value: target})
	if err != nil {
		t.Fatalf("AlignCargoLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignCargoLicense() changed = false, want true")
	}
	want := "[package]\nname = \"example-project\"\nlicense = \"a\\\"b\\\\c\\bd\\te\\nf\\fg\\rh\\u0001i\"\n"
	if aligned != want {
		t.Fatalf("AlignCargoLicense() = %q, want %q", aligned, want)
	}
	surface, err := InspectCargoLicense(aligned)
	if err != nil {
		t.Fatalf("InspectCargoLicense() error = %v", err)
	}
	if surface.State != CargoLicenseExpression || surface.Value != target {
		t.Fatalf("InspectCargoLicense() = (%d, %q), want the decoded round trip", surface.State, surface.Value)
	}
}

func TestAlignCargoLicenseRefusesInvalidSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		form      CargoLicenseForm
		wantErrIs error
	}{
		{"no package table", "tool = \"x\"\n", CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"}, ErrNoPackageTable},
		{"array table only", "[[package]]\nlicense = \"x\"\n", CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"}, ErrNoPackageTable},
		{"unscannable surface", "[package", CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"}, ErrInvalidTOMLSurface},
		{"invalid license value", "[package]\nlicense = 7\n", CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"}, ErrInvalidCargoLicenseValue},
		{"invalid license-file value", "[package]\nlicense-file = 7\n", CargoLicenseForm{Field: CargoLicenseFileKey, Value: "LICENSE"}, ErrInvalidCargoLicenseFile},
		{"exclusive state", "[package]\nlicense = \"MIT\"\nlicense-file = \"LICENSE\"\n", CargoLicenseForm{Field: CargoLicenseField, Value: "MIT"}, ErrExclusiveCargoLicense},
		{"non-utf-8 target", exampleCargoToml, CargoLicenseForm{Field: CargoLicenseField, Value: "MIT\xff"}, ErrInvalidTarget},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := AlignCargoLicense(testCase.content, testCase.form)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("AlignCargoLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
			}
		})
	}
}

func FuzzAlignCargoLicense(f *testing.F) {
	f.Add(exampleCargoToml, "license", "MIT")
	f.Add(exampleCargoToml, "license-file", "LICENSE")
	f.Add("[package]\nname = \"x\"\n", "license-file", "LICENSE")
	f.Add("[package]", "license", "MIT")
	f.Add("[package\n", "license", "MIT")
	f.Add("[package]\nlicense = \"a\"\nlicense-file = \"b\"\n", "license", "MIT")
	f.Add("[package]\nlicense = 7\n", "license", "MIT")
	f.Add("[[package]]\nlicense = \"x\"\n", "license", "MIT")
	f.Add("", "license", "MIT")
	f.Add("[package]\nlicense = \"\"\"multi\nline\"\"\"\n", "license-file", "LICENSE")
	f.Fuzz(func(t *testing.T, content, field, value string) {
		if field != CargoLicenseField && field != CargoLicenseFileKey {
			return
		}
		form := CargoLicenseForm{Field: field, Value: value}
		aligned, changed, err := AlignCargoLicense(content, form)
		if err != nil {
			return
		}
		surface, err := InspectCargoLicense(aligned)
		if err != nil {
			t.Fatalf("AlignCargoLicense produced an unprovable surface: %q -> %q (%v)", content, aligned, err)
		}
		if surface.State != CargoLicenseExpression && surface.State != CargoLicenseFileForm {
			t.Fatalf("AlignCargoLicense did not align the license field: %q -> %q (%d)", content, aligned, surface.State)
		}
		if surface.Value != value {
			t.Fatalf("AlignCargoLicense did not align the license value: %q -> %q (%q)", content, aligned, surface.Value)
		}
		if !changed {
			return
		}
		again, changedAgain, err := AlignCargoLicense(aligned, form)
		if err != nil || changedAgain || again != aligned {
			t.Fatalf("AlignCargoLicense is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
		}
	})
}
