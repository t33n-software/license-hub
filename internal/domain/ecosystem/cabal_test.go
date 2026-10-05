package ecosystem

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// Convention: docs/conventions/cli/testing/README.md

func TestCabalProjectionDerivesFromTheTenantValues(t *testing.T) {
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
			if got := CabalProjection(testCase.merged); got != testCase.want {
				t.Fatalf("CabalProjection() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestCabalManifestNamesDiscoversTheSuffix(t *testing.T) {
	names := CabalManifestNames([]string{"B.cabal", "a.CABAL", "readme.md", "x.cabal.bak"})
	if want := []string{"B.cabal", "a.CABAL"}; !slices.Equal(names, want) {
		t.Fatalf("CabalManifestNames() = %v, want %v", names, want)
	}
}

const cabalExample = "cabal-version:      2.2\n" +
	"name:               example\n" +
	"license:            MIT\n" +
	"license-file:       LICENSE\n" +
	"\n" +
	"library\n" +
	"    exposed-modules: Example\n"

func TestInspectCabalLicenseClassifiesTheSurface(t *testing.T) {
	cases := []struct {
		name           string
		content        string
		wantValue      string
		wantState      CabalLicenseState
		wantFileState  CabalFileState
		wantFileValue  string
		wantErrIs      error
		wantErr        bool
		wantFileEnties []string
	}{
		{"full surface", cabalExample, "MIT", CabalLicenseValue, CabalFileSingle, "LICENSE", nil, false, nil},
		{"missing fields", "name: x\n", "", CabalLicenseMissing, CabalFileMissing, "", nil, false, nil},
		{"empty content", "", "", CabalLicenseMissing, CabalFileMissing, "", nil, false, nil},
		{"case-insensitive keys", "LICENSE: MIT\nLicense-File: LICENSE\n", "MIT", CabalLicenseValue, CabalFileSingle, "LICENSE", nil, false, nil},
		{"quoted values", "license: \"MIT\"\nlicense-file: \"LICENSE\"\n", "MIT", CabalLicenseValue, CabalFileSingle, "LICENSE", nil, false, nil},
		{"compound value", "license: MIT OR Apache-2.0\n", "MIT OR Apache-2.0", CabalLicenseValue, CabalFileMissing, "", nil, false, nil},
		{"comment after the value", "license: MIT -- the license\n", "MIT", CabalLicenseValue, CabalFileMissing, "", nil, false, nil},
		{"continuation value", "license:\n  MIT\n", "MIT", CabalLicenseValue, CabalFileMissing, "", nil, false, nil},
		{"blank line inside the value", "license-files:\n\n  LICENSE\n", "", CabalLicenseMissing, CabalFileList, "", nil, false, []string{"LICENSE"}},
		{"comment inside the value", "license-files: LICENSE\n  -- note\n  NOTICE\n", "", CabalLicenseMissing, CabalFileList, "", nil, false, []string{"LICENSE", "NOTICE"}},
		{"brace list", "license-files: { LICENSE }\n", "", CabalLicenseMissing, CabalFileList, "", nil, false, []string{"LICENSE"}},
		{"multi-entry list", "license-files: LICENSE NOTICE\n", "", CabalLicenseMissing, CabalFileList, "", nil, false, []string{"LICENSE", "NOTICE"}},
		{"continuation list at eof", "license-files:\n  LICENSE", "", CabalLicenseMissing, CabalFileList, "", nil, false, []string{"LICENSE"}},
		{"empty license field", "license:\nname: x\n", "", CabalLicenseInvalid, CabalFileMissing, "", nil, false, nil},
		{"whitespace-only license field", "license:   \nname: x\n", "", CabalLicenseInvalid, CabalFileMissing, "", nil, false, nil},
		{"comment-only license field", "license: -- note\nname: x\n", "", CabalLicenseInvalid, CabalFileMissing, "", nil, false, nil},
		{"empty license-files field", "license-files:\nname: x\n", "", CabalLicenseMissing, CabalFileInvalid, "", nil, false, nil},
		{"empty license-file field", "license-file:\nname: x\n", "", CabalLicenseMissing, CabalFileInvalid, "", nil, false, nil},
		{"unterminated brace in a continuation", "license-files:\n  { LICENSE\n", "", 0, 0, "", ErrInvalidCabalSurface, true, nil},
		{"empty brace list", "license-files: { }\nname: x\n", "", CabalLicenseMissing, CabalFileInvalid, "", nil, false, nil},
		{"value on the next line only", "license-files:\n  LICENSE\n  NOTICE\n", "", CabalLicenseMissing, CabalFileList, "", nil, false, []string{"LICENSE", "NOTICE"}},
		{"crlf endings", "license: MIT\r\nlicense-file: LICENSE\r\n", "MIT", CabalLicenseValue, CabalFileSingle, "LICENSE", nil, false, nil},
		{"section shadow", "library\n  license: MIT\n", "", 0, 0, "", ErrInvalidCabalSurface, true, nil},
		{"duplicate license field", "license: MIT\nlicense: BSD3\n", "", 0, 0, "", ErrAmbiguousCabalLicense, true, nil},
		{"duplicate license-file field", "license-file: A\nlicense-file: B\n", "", 0, 0, "", ErrAmbiguousCabalLicenseFile, true, nil},
		{"duplicate license-files field", "license-files: A\nlicense-files: B\n", "", 0, 0, "", ErrAmbiguousCabalLicenseFile, true, nil},
		{"both file forms", "license-file: A\nlicense-files: B\n", "", 0, 0, "", ErrAmbiguousCabalLicenseFile, true, nil},
		{"unterminated quoted value", "license: \"MIT\n", "", 0, 0, "", ErrInvalidCabalSurface, true, nil},
		{"quoted escape", "license: \"M\\IT\"\n", "", 0, 0, "", ErrInvalidCabalSurface, true, nil},
		{"quoted trailing tokens", "license: \"MIT\" x\n", "", 0, 0, "", ErrInvalidCabalSurface, true, nil},
		{"unterminated brace list", "license-files: { LICENSE\n", "", 0, 0, "", ErrInvalidCabalSurface, true, nil},
		{"brace trailing tokens", "license-files: { LICENSE } x\n", "", 0, 0, "", ErrInvalidCabalSurface, true, nil},
		{"indented content before any property", "  license: MIT\n", "", 0, 0, "", ErrInvalidCabalSurface, true, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectCabalLicense(testCase.content)
			if testCase.wantErr {
				if !errors.Is(err, testCase.wantErrIs) {
					t.Fatalf("InspectCabalLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("InspectCabalLicense() error = %v", err)
			}
			if surface.State != testCase.wantState {
				t.Fatalf("InspectCabalLicense() state = %d, want %d", surface.State, testCase.wantState)
			}
			if surface.Value != testCase.wantValue {
				t.Fatalf("InspectCabalLicense() value = %q, want %q", surface.Value, testCase.wantValue)
			}
			if surface.FileState != testCase.wantFileState {
				t.Fatalf("InspectCabalLicense() file state = %d, want %d", surface.FileState, testCase.wantFileState)
			}
			if surface.FileValue != testCase.wantFileValue {
				t.Fatalf("InspectCabalLicense() file value = %q, want %q", surface.FileValue, testCase.wantFileValue)
			}
			if len(surface.FileEntries) != len(testCase.wantFileEnties) {
				t.Fatalf("InspectCabalLicense() file entries = %v, want %v", surface.FileEntries, testCase.wantFileEnties)
			}
			for i, want := range testCase.wantFileEnties {
				if surface.FileEntries[i] != want {
					t.Fatalf("InspectCabalLicense() file entries = %v, want %v", surface.FileEntries, testCase.wantFileEnties)
				}
			}
		})
	}
}

func TestAlignCabalLicenseReplacesADivergingLicenseValue(t *testing.T) {
	aligned, changed, err := AlignCabalLicense(cabalExample, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignCabalLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignCabalLicense() changed = false, want true")
	}
	want := strings.Replace(cabalExample, "license:            MIT", "license:            LicenseRef-example-NoRepublish-1.0", 1)
	if aligned != want {
		t.Fatalf("AlignCabalLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignCabalLicenseIsIdempotentOnAlignedState(t *testing.T) {
	aligned, changed, err := AlignCabalLicense(cabalExample, "MIT")
	if err != nil {
		t.Fatalf("AlignCabalLicense() error = %v", err)
	}
	if changed {
		t.Fatal("AlignCabalLicense() changed = true, want false")
	}
	if aligned != cabalExample {
		t.Fatalf("AlignCabalLicense() mutated an aligned manifest: %q", aligned)
	}
}

func TestAlignCabalLicenseAlignsTheFileSurface(t *testing.T) {
	target := "LicenseRef-example-NoRepublish-1.0"
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{"diverging license-file value",
			"license: " + target + "\nlicense-file: COPYING\n",
			"license: " + target + "\nlicense-file: LICENSE\n"},
		{"quoted diverging license-file value",
			"license: " + target + "\nlicense-file: \"COPYING\"\n",
			"license: " + target + "\nlicense-file: LICENSE\n"},
		{"diverging license-files list",
			"license: " + target + "\nlicense-files: COPYING\n",
			"license: " + target + "\nlicense-file: LICENSE\n"},
		{"diverging license-files continuation list",
			"license: " + target + "\nlicense-files:\n  COPYING\n",
			"license: " + target + "\nlicense-file: LICENSE\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignCabalLicense(testCase.content, target)
			if err != nil {
				t.Fatalf("AlignCabalLicense() error = %v", err)
			}
			if !changed {
				t.Fatal("AlignCabalLicense() changed = false, want true")
			}
			if aligned != testCase.want {
				t.Fatalf("AlignCabalLicense() = %q, want %q", aligned, testCase.want)
			}
			surface, err := InspectCabalLicense(aligned)
			if err != nil || surface.FileState != CabalFileSingle || surface.FileValue != CabalFileTarget {
				t.Fatalf("AlignCabalLicense() produced an unprovable file surface: %q (%v)", aligned, err)
			}
			again, changedAgain, err := AlignCabalLicense(aligned, target)
			if err != nil || changedAgain || again != aligned {
				t.Fatalf("AlignCabalLicense() file alignment is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
			}
		})
	}
}

func TestAlignCabalLicenseKeepsTheValueEqualListSpelling(t *testing.T) {
	content := "license: LicenseRef-example-NoRepublish-1.0\nlicense-files: LICENSE\n"
	aligned, changed, err := AlignCabalLicense(content, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignCabalLicense() error = %v", err)
	}
	if changed {
		t.Fatal("AlignCabalLicense() changed = true, want false")
	}
	if aligned != content {
		t.Fatalf("AlignCabalLicense() mutated the value-equal spelling: %q", aligned)
	}
}

func TestAlignCabalLicenseInsertsMissingFields(t *testing.T) {
	target := "LicenseRef-example-NoRepublish-1.0"
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{"both fields missing",
			"cabal-version: 2.2\nname: x\n",
			"cabal-version: 2.2\nlicense: " + target + "\nlicense-file: LICENSE\nname: x\n"},
		{"license field missing",
			"cabal-version: 2.2\nlicense-file: LICENSE\n",
			"cabal-version: 2.2\nlicense: " + target + "\nlicense-file: LICENSE\n"},
		{"license-file field missing",
			"cabal-version: 2.2\nlicense: " + target + "\n",
			"cabal-version: 2.2\nlicense-file: LICENSE\nlicense: " + target + "\n"},
		{"anchor after a continued first property",
			"synopsis:\n  a text\nlicense-file: LICENSE\n",
			"synopsis:\n  a text\nlicense: " + target + "\nlicense-file: LICENSE\n"},
		{"anchor after comments before the first property",
			"-- header\n\nname: x\n",
			"-- header\n\nname: x\nlicense: " + target + "\nlicense-file: LICENSE\n"},
		{"anchor at eof without a newline",
			"name: x",
			"name: x\nlicense: " + target + "\nlicense-file: LICENSE"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignCabalLicense(testCase.content, target)
			if err != nil {
				t.Fatalf("AlignCabalLicense() error = %v", err)
			}
			if !changed {
				t.Fatal("AlignCabalLicense() changed = false, want true")
			}
			if aligned != testCase.want {
				t.Fatalf("AlignCabalLicense() = %q, want %q", aligned, testCase.want)
			}
			surface, err := InspectCabalLicense(aligned)
			if err != nil || surface.State != CabalLicenseValue || surface.Value != target {
				t.Fatalf("AlignCabalLicense() produced an unprovable surface: %q (%v)", aligned, err)
			}
			again, changedAgain, err := AlignCabalLicense(aligned, target)
			if err != nil || changedAgain || again != aligned {
				t.Fatalf("AlignCabalLicense() insertion is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
			}
		})
	}
}

func TestAlignCabalLicensePreservesEveryNonLicenseByte(t *testing.T) {
	content := "cabal-version: 2.2 -- pinned\nname: x\nlicense: MIT -- the license\nlicense-file: OTHER -- text\n"
	aligned, _, err := AlignCabalLicense(content, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignCabalLicense() error = %v", err)
	}
	want := "cabal-version: 2.2 -- pinned\nname: x\nlicense: LicenseRef-example-NoRepublish-1.0 -- the license\nlicense-file: LICENSE -- text\n"
	if aligned != want {
		t.Fatalf("AlignCabalLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignCabalLicenseRefusesAmbiguousAndInvalidSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		target    string
		wantErrIs error
	}{
		{"multiple file entries", "license: MIT\nlicense-files: LICENSE NOTICE\n", "MIT", ErrMultipleCabalLicenseFiles},
		{"both file forms", "license: MIT\nlicense-file: LICENSE\nlicense-files: LICENSE\n", "MIT", ErrAmbiguousCabalLicenseFile},
		{"duplicate license field", "license: MIT\nlicense: BSD3\n", "MIT", ErrAmbiguousCabalLicense},
		{"empty license field", "license:\nname: x\n", "MIT", ErrInvalidCabalLicenseValue},
		{"empty license-files field", "license-files:\nname: x\n", "MIT", ErrInvalidCabalLicenseValue},
		{"empty license-file field", "license-file:\nname: x\n", "MIT", ErrInvalidCabalLicenseValue},
		{"unterminated brace in a continuation", "license-files:\n  { LICENSE\n", "MIT", ErrInvalidCabalSurface},
		{"license surface in a section", "library\n  license: MIT\n", "MIT", ErrInvalidCabalSurface},
		{"no anchor", "library\n  exposed-modules: X\n", "MIT", ErrInvalidCabalSurface},
		{"empty document", "", "MIT", ErrInvalidCabalSurface},
		{"unscannable surface", "license: \"MIT\n", "MIT", ErrInvalidCabalSurface},
		{"non-utf-8 target", "name: x\n", "MIT\xff", ErrInvalidTarget},
		{"empty target", "name: x\n", "", ErrInvalidTarget},
		{"whitespace target", "name: x\n", "MIT OR Apache-2.0", ErrInvalidTarget},
		{"quote target", "name: x\n", "MI\"T", ErrInvalidTarget},
		{"backslash target", "name: x\n", "MI\\T", ErrInvalidTarget},
		{"brace target", "name: x\n", "MI{T}", ErrInvalidTarget},
		{"comment target", "name: x\n", "--MIT", ErrInvalidTarget},
		{"vertical-tab target", "name: x\n", "MI\vT", ErrInvalidTarget},
		{"form-feed target", "name: x\n", "MI\fT", ErrInvalidTarget},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := AlignCabalLicense(testCase.content, testCase.target)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("AlignCabalLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
			}
		})
	}
}

func TestCabalSkeletonLexerContextForms(t *testing.T) {
	cases := []struct {
		name          string
		content       string
		wantValue     string
		wantFileState CabalFileState
	}{
		{"line comment", "-- license: MIT\nname: x\n", "", CabalFileMissing},
		{"indented comment inside the value", "license-files:\n  -- note\n  LICENSE\n", "", CabalFileList},
		{"section header forms", "library\nexecutable foo\nif flag(x)\nelse\nlicense: MIT\n", "MIT", CabalFileMissing},
		{"column-0 dash line", "-\nlicense: MIT\n", "MIT", CabalFileMissing},
		{"double dash inside the token", "license: MIT--x\n", "MIT--x", CabalFileMissing},
		{"column-0 digit line", "1 flag\nlicense: MIT\n", "MIT", CabalFileMissing},
		{"section content is not a continuation", "license: MIT\nlibrary\n  exposed-modules: X\n", "MIT", CabalFileMissing},
		{"property after a section", "library\n  x: 1\nlicense: MIT\n", "MIT", CabalFileMissing},
		{"value with padded key spacing", "license:MIT\n", "MIT", CabalFileMissing},
		{"trailing comment in the list", "license-files: LICENSE -- text\n", "", CabalFileList},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectCabalLicense(testCase.content)
			if err != nil {
				t.Fatalf("InspectCabalLicense(%q) error = %v", testCase.content, err)
			}
			if surface.Value != testCase.wantValue {
				t.Fatalf("InspectCabalLicense(%q) value = %q, want %q", testCase.content, surface.Value, testCase.wantValue)
			}
			if surface.FileState != testCase.wantFileState {
				t.Fatalf("InspectCabalLicense(%q) file state = %d, want %d", testCase.content, surface.FileState, testCase.wantFileState)
			}
		})
	}
}

func FuzzAlignCabalLicense(f *testing.F) {
	f.Add(cabalExample, "LicenseRef-example-NoRepublish-1.0")
	f.Add("name: x\n", "MIT")
	f.Add("license: MIT\nlicense-file: LICENSE\n", "Apache-2.0")
	f.Add("license: MIT\nlicense-file: LICENSE\nlicense-files: LICENSE\n", "MIT")
	f.Add("license-files: LICENSE NOTICE\n", "MIT")
	f.Add("license:\n", "MIT")
	f.Add("license-files:\n", "MIT")
	f.Add("library\n  license: MIT\n", "MIT")
	f.Add("license: \"MIT\n", "MIT")
	f.Add("license: \"M\\IT\"\n", "MIT")
	f.Add("library\n  x: 1\n", "MIT")
	f.Add("", "MIT")
	f.Add("-- c\n\nname: x\n", "MIT")
	f.Add("license-files:\n  LICENSE\n  NOTICE\n", "MIT")
	f.Add("license: MIT -- c\n", "MIT")
	f.Add("license: { MIT }\n", "MIT")
	f.Add("  license: MIT\n", "MIT")
	f.Add("license: LicenseRef-license-hub-NoRepublish-1.0\nlicense-files: LICENSE\n", "LicenseRef-license-hub-NoRepublish-1.0")
	f.Add("synopsis:\n  a text\nname: x\n", "MIT")
	f.Add("name: x", "MIT")
	f.Add("license-files: { LICENSE\n", "MIT")
	f.Add("license: MIT--x\n", "MIT")
	f.Fuzz(func(t *testing.T, content, target string) {
		aligned, changed, err := AlignCabalLicense(content, target)
		if err != nil {
			return
		}
		surface, err := InspectCabalLicense(aligned)
		if err != nil || surface.State != CabalLicenseValue || surface.Value != target {
			t.Fatalf("AlignCabalLicense produced an unprovable license surface: %q -> %q (%q, %v)", content, aligned, surface.Value, err)
		}
		provesFile := surface.FileState == CabalFileSingle && surface.FileValue == CabalFileTarget
		if surface.FileState == CabalFileList {
			provesFile = len(surface.FileEntries) == 1 && surface.FileEntries[0] == CabalFileTarget
		}
		if !provesFile {
			t.Fatalf("AlignCabalLicense produced an unprovable file surface: %q -> %q", content, aligned)
		}
		if !changed {
			return
		}
		again, changedAgain, err := AlignCabalLicense(aligned, target)
		if err != nil || changedAgain || again != aligned {
			t.Fatalf("AlignCabalLicense is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
		}
	})
}
