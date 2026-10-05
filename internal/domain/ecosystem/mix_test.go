package ecosystem

import (
	"errors"
	"strings"
	"testing"
)

// Convention: docs/conventions/cli/testing/README.md

func TestElixirProjectionDerivesFromTheTenantValues(t *testing.T) {
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
			if got := ElixirProjection(testCase.merged); got != testCase.want {
				t.Fatalf("ElixirProjection() = %q, want %q", got, testCase.want)
			}
		})
	}
}

const mixExample = "defmodule Example.MixProject do\n" +
	"  use Mix.Project\n\n" +
	"  def project do\n" +
	"    [\n" +
	"      app: :example_project,\n" +
	"      version: \"0.1.0\",\n" +
	"      package: [\n" +
	"        name: :example_project,\n" +
	"        licenses: [\"MIT\"]\n" +
	"      ]\n" +
	"    ]\n" +
	"  end\n" +
	"end\n"

func TestInspectElixirLicenseClassifiesTheEntry(t *testing.T) {
	cases := []struct {
		name       string
		content    string
		wantValue  string
		wantState  ElixirLicenseState
		wantErrIs  error
		wantErr    bool
		wantValues []string
	}{
		{"licenses entry", mixExample, "MIT", ElixirLicenseList, nil, false, []string{"MIT"}},
		{"missing licenses entry", "def project do\n  [package: [name: :x]]\nend\n", "", ElixirLicenseMissing, nil, false, nil},
		{"missing package configuration", "defmodule X.MixProject do\n  def project do\n    [app: :x]\n  end\nend\n", "", ElixirLicenseMissing, nil, false, nil},
		{"empty module", "defmodule X.MixProject do\nend\n", "", ElixirLicenseMissing, nil, false, nil},
		{"inline package", "def project do\n  [app: :x, package: [licenses: [\"MIT\"]]]\nend\n", "MIT", ElixirLicenseList, nil, false, []string{"MIT"}},
		{"trailing comma", "def project do\n  [package: [licenses: [\"MIT\",]]]\nend\n", "MIT", ElixirLicenseList, nil, false, []string{"MIT"}},
		{"padded entries", "def project do\n  [package: [licenses: [ \"MIT\" ]]]\nend\n", "MIT", ElixirLicenseList, nil, false, []string{"MIT"}},
		{"multiline list", "def project do\n  [\n    package: [\n      licenses: [\n        \"MIT\"\n      ]\n    ]\n  ]\nend\n", "MIT", ElixirLicenseList, nil, false, []string{"MIT"}},
		{"package with other entries", "def project do\n  [package: [name: :x, links: %{\"GitHub\" => \"url\"}, licenses: [\"MIT\"], files: [\"lib\"]]]\nend\n", "MIT", ElixirLicenseList, nil, false, []string{"MIT"}},
		{"comment before the value", "def project do\n  [package: [# the package\n    licenses: [\"MIT\"]]]\nend\n", "MIT", ElixirLicenseList, nil, false, []string{"MIT"}},
		{"compound expression entry", "def project do\n  [package: [licenses: [\"(MIT OR Apache-2.0)\"]]]\nend\n", "(MIT OR Apache-2.0)", ElixirLicenseList, nil, false, []string{"(MIT OR Apache-2.0)"}},
		{"multi-element list", "def project do\n  [package: [licenses: [\"MIT\", \"Apache-2.0\"]]]\nend\n", "", ElixirLicenseList, nil, false, []string{"MIT", "Apache-2.0"}},
		{"string key map context", "def project do\n  [links: %{\"licenses\" => [\"MIT\"]}]\nend\n", "", ElixirLicenseMissing, nil, false, nil},
		{"atom value context", "x = :licenses\n", "", ElixirLicenseMissing, nil, false, nil},
		{"module attribute context", "@licenses [\"MIT\"]\n", "", ElixirLicenseMissing, nil, false, nil},
		{"comment context", "# licenses: [\"MIT\"]\n", "", ElixirLicenseMissing, nil, false, nil},
		{"string context", "summary = \"licenses: [\\\"MIT\\\"]\"\n", "", ElixirLicenseMissing, nil, false, nil},
		{"interpolated string context", "summary = \"prefix #{name} suffix\"\n", "", ElixirLicenseMissing, nil, false, nil},
		{"interpolated string with nested string context", "summary = \"#{Map.get(m, \"k\")}\"\n", "", ElixirLicenseMissing, nil, false, nil},
		{"heredoc docstring context", "@moduledoc \"\"\"\nlicenses: [\"MIT\"]\n\"\"\"\n", "", ElixirLicenseMissing, nil, false, nil},
		{"sigil context", "summary = ~s(licenses: [\"MIT\"])\n", "", ElixirLicenseMissing, nil, false, nil},
		{"uppercase sigil context", "summary = ~S(licenses: [\"MIT\"])\n", "", ElixirLicenseMissing, nil, false, nil},
		{"heredoc sigil context", "summary = ~s\"\"\"\nlicenses: [\"MIT\"]\n\"\"\"\n", "", ElixirLicenseMissing, nil, false, nil},
		{"charlist context", "summary = 'licenses'\n", "", ElixirLicenseMissing, nil, false, nil},
		{"char literal context", "x = ?a\n", "", ElixirLicenseMissing, nil, false, nil},
		{"non-keyword ident context", "licenseds = 1\n", "", ElixirLicenseMissing, nil, false, nil},
		{"double colon not a keyword", "x = a::b\n", "", ElixirLicenseMissing, nil, false, nil},
		{"def package form is outside the canonical surface", "def package do\n  [licenses: [\"MIT\"]]\nend\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"shadow licenses entry", "def project do\n  [licenses: [\"MIT\"]]\nend\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"nested licenses entry in a map", "def project do\n  [package: [links: %{licenses: [\"MIT\"]}]]\nend\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"nested licenses entry in a list", "def project do\n  [package: [links: [licenses: [\"MIT\"]]]]\nend\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"two package configurations", "def project do\n  [package: [licenses: [\"MIT\"]], package: [licenses: [\"Apache-2.0\"]]]\nend\n", "", 0, ErrAmbiguousElixirPackage, true, nil},
		{"two licenses entries", "def project do\n  [package: [licenses: [\"MIT\"], licenses: [\"Apache-2.0\"]]]\nend\n", "", 0, ErrAmbiguousElixirLicense, true, nil},
		{"package configuration is a call", "def project do\n  [package: package_config()]\nend\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"package configuration is a map", "def project do\n  [package: %{licenses: [\"MIT\"]}]\nend\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"bare string value", "def project do\n  [package: [licenses: \"MIT\"]]\nend\n", "", ElixirLicenseInvalid, nil, false, nil},
		{"atom value", "def project do\n  [package: [licenses: :mit]]\nend\n", "", ElixirLicenseInvalid, nil, false, nil},
		{"number value", "def project do\n  [package: [licenses: 7]]\nend\n", "", ElixirLicenseInvalid, nil, false, nil},
		{"sigil word list value", "def project do\n  [package: [licenses: ~w(MIT)]]\nend\n", "", ElixirLicenseInvalid, nil, false, nil},
		{"empty list", "def project do\n  [package: [licenses: []]]\nend\n", "", 0, ErrInvalidElixirLicenseValue, true, nil},
		{"non-string entry", "def project do\n  [package: [licenses: [:mit]]]\nend\n", "", 0, ErrInvalidElixirLicenseValue, true, nil},
		{"mixed entries", "def project do\n  [package: [licenses: [\"MIT\", :mit]]]\nend\n", "", 0, ErrInvalidElixirLicenseValue, true, nil},
		{"entry with escape", "def project do\n  [package: [licenses: [\"M\\\\IT\"]]]\nend\n", "", 0, ErrInvalidElixirLicenseValue, true, nil},
		{"interpolated entry", "def project do\n  [package: [licenses: [\"MIT#{x}\"]]]\nend\n", "", 0, ErrInvalidElixirLicenseValue, true, nil},
		{"malformed delimiter", "def project do\n  [package: [licenses: [\"MIT\" \"Apache-2.0\"]]]\nend\n", "", 0, ErrInvalidElixirLicenseValue, true, nil},
		{"unterminated licenses list", "def project do\n  [package: [licenses: [\"MIT\"]]\nend\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"unterminated string context", "summary = \"unterminated\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"unterminated heredoc context", "@moduledoc \"\"\"\ntext\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"unterminated sigil context", "summary = ~s(text\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"unterminated charlist context", "x = 'text\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"unterminated interpolation context", "summary = \"a #{b\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"unterminated bracket scope", "def project do\n  [package: [licenses: [\"MIT\"]]\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"unbalanced bracket", "def project do\n  ]\nend\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"interpolation nesting beyond the scanner bound", "summary = \"" + strings.Repeat("#{", 70) + "\"\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"interpolation at the scanner bound", "summary = \"" + strings.Repeat("#{", 64) + strings.Repeat("}", 64) + "\"\n", "", ElixirLicenseMissing, nil, false, nil},
		{"string nesting beyond the scanner bound", "summary = \"" + strings.Repeat("#{", 64) + "\"\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"top-level licenses entry outside every scope", "licenses: [\"MIT\"]\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"unterminated charlist inside interpolation", "summary = \"a #{'b\n", "", 0, ErrInvalidElixirSurface, true, nil},
		{"list opens at eof", "def project do\n  [package: [licenses: [", "", 0, ErrInvalidElixirSurface, true, nil},
		{"entry literal opens at eof", "def project do\n  [package: [licenses: [\"MIT", "", 0, ErrInvalidElixirSurface, true, nil},
		{"entry without list close at eof", "def project do\n  [package: [licenses: [\"MIT\"", "", 0, ErrInvalidElixirSurface, true, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectElixirLicense(testCase.content)
			if testCase.wantErr {
				if !errors.Is(err, testCase.wantErrIs) {
					t.Fatalf("InspectElixirLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("InspectElixirLicense() error = %v", err)
			}
			if surface.State != testCase.wantState {
				t.Fatalf("InspectElixirLicense() state = %d, want %d", surface.State, testCase.wantState)
			}
			if surface.Value != testCase.wantValue {
				t.Fatalf("InspectElixirLicense() value = %q, want %q", surface.Value, testCase.wantValue)
			}
			if len(surface.Entries) != len(testCase.wantValues) {
				t.Fatalf("InspectElixirLicense() entries = %v, want %v", surface.Entries, testCase.wantValues)
			}
			for i, want := range testCase.wantValues {
				if surface.Entries[i] != want {
					t.Fatalf("InspectElixirLicense() entries = %v, want %v", surface.Entries, testCase.wantValues)
				}
			}
		})
	}
}

func TestAlignElixirLicenseReplacesADivergingEntry(t *testing.T) {
	aligned, changed, err := AlignElixirLicense(mixExample, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignElixirLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignElixirLicense() changed = false, want true")
	}
	want := strings.Replace(mixExample, `licenses: ["MIT"]`, `licenses: ["LicenseRef-example-NoRepublish-1.0"]`, 1)
	if aligned != want {
		t.Fatalf("AlignElixirLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignElixirLicenseIsIdempotentOnAlignedState(t *testing.T) {
	aligned, changed, err := AlignElixirLicense(mixExample, "MIT")
	if err != nil {
		t.Fatalf("AlignElixirLicense() error = %v", err)
	}
	if changed {
		t.Fatal("AlignElixirLicense() changed = true, want false")
	}
	if aligned != mixExample {
		t.Fatalf("AlignElixirLicense() mutated an aligned manifest: %q", aligned)
	}
}

func TestAlignElixirLicensePreservesEveryNonLicenseByte(t *testing.T) {
	content := "def project do\n  [package: [name: :x, licenses: [\"MIT\"], files: [\"lib\"]]] # the package\nend\n"
	aligned, _, err := AlignElixirLicense(content, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignElixirLicense() error = %v", err)
	}
	before, _, _ := strings.Cut(content, `licenses: ["MIT"]`)
	afterOriginal := content[strings.Index(content, `licenses: ["MIT"]`)+len(`licenses: ["MIT"]`):]
	if !strings.HasPrefix(aligned, before+`licenses: ["LicenseRef-example-NoRepublish-1.0"]`) {
		t.Fatalf("AlignElixirLicense() changed bytes before the entry: %q", aligned)
	}
	if !strings.HasSuffix(aligned, afterOriginal) {
		t.Fatalf("AlignElixirLicense() changed bytes after the entry: %q", aligned)
	}
	if !strings.Contains(aligned, "name: :x") || !strings.Contains(aligned, "files: [\"lib\"]") {
		t.Fatalf("AlignElixirLicense() touched a non-license entry: %q", aligned)
	}
}

func TestAlignElixirLicensePreservesTheMultilineListFormatting(t *testing.T) {
	content := "def project do\n  [\n    package: [\n      licenses: [\n        \"MIT\"\n      ]\n    ]\n  ]\nend\n"
	aligned, changed, err := AlignElixirLicense(content, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignElixirLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignElixirLicense() changed = false, want true")
	}
	want := strings.Replace(content, "\"MIT\"", "\"LicenseRef-example-NoRepublish-1.0\"", 1)
	if aligned != want {
		t.Fatalf("AlignElixirLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignElixirLicenseInsertsAMissingEntry(t *testing.T) {
	target := "LicenseRef-example-NoRepublish-1.0"
	cases := []struct {
		name       string
		content    string
		wantSubstr string
	}{
		{"inline package with entries", "def project do\n  [package: [name: :x]]\nend\n", "[licenses: [\"" + target + "\"], name: :x]"},
		{"multiline package with entries", "def project do\n  [\n    package: [\n      name: :x\n    ]\n  ]\nend\n", "[licenses: [\"" + target + "\"],\n      name: :x"},
		{"empty package list", "def project do\n  [package: []]\nend\n", "[package: [licenses: [\"" + target + "\"]]]"},
		{"blank package list", "def project do\n  [package: [\n  ]]\nend\n", "[licenses: [\"" + target + "\"]"},
		{"comment right after the bracket", "def project do\n  [package: [# the package\n    name: :x]]\nend\n", "[licenses: [\"" + target + "\"],# the package"},
		{"string entry follows inline", "def project do\n  [package: [\"name\"]]\nend\n", "[licenses: [\"" + target + "\"], \"name\"]"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignElixirLicense(testCase.content, target)
			if err != nil {
				t.Fatalf("AlignElixirLicense() error = %v", err)
			}
			if !changed {
				t.Fatal("AlignElixirLicense() changed = false, want true")
			}
			if !strings.Contains(aligned, testCase.wantSubstr) {
				t.Fatalf("AlignElixirLicense() = %q, want it to carry %q", aligned, testCase.wantSubstr)
			}
			surface, err := InspectElixirLicense(aligned)
			if err != nil || surface.State != ElixirLicenseList {
				t.Fatalf("AlignElixirLicense() produced an unscannable surface: %q (%v)", aligned, err)
			}
			if surface.Value != target {
				t.Fatalf("AlignElixirLicense() licenses = %q", surface.Value)
			}
			again, changedAgain, err := AlignElixirLicense(aligned, target)
			if err != nil || changedAgain || again != aligned {
				t.Fatalf("AlignElixirLicense() insertion is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
			}
		})
	}
}

func TestAlignElixirLicenseRefusesAmbiguousAndInvalidSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		target    string
		wantErrIs error
	}{
		{"multiple entries", "def project do\n  [package: [licenses: [\"MIT\", \"Apache-2.0\"]]]\nend\n", "MIT", ErrMultipleElixirLicenses},
		{"invalid value", "def project do\n  [package: [licenses: \"MIT\"]]\nend\n", "MIT", ErrInvalidElixirLicenseValue},
		{"empty list", "def project do\n  [package: [licenses: []]]\nend\n", "MIT", ErrInvalidElixirLicenseValue},
		{"ambiguous package configurations", "def project do\n  [package: [licenses: [\"MIT\"]], package: []]\nend\n", "MIT", ErrAmbiguousElixirPackage},
		{"ambiguous licenses entries", "def project do\n  [package: [licenses: [\"MIT\"], licenses: [\"MIT\"]]]\nend\n", "MIT", ErrAmbiguousElixirLicense},
		{"malformed surface", "def project do\n  [package: [licenses: [\"MIT\"]]\nend\n", "MIT", ErrInvalidElixirSurface},
		{"no package configuration", "defmodule X.MixProject do\nend\n", "MIT", ErrInvalidElixirSurface},
		{"def package form", "def package do\n  [licenses: [\"MIT\"]]\nend\n", "MIT", ErrInvalidElixirSurface},
		{"shadow licenses entry", "def project do\n  [licenses: [\"MIT\"]]\nend\n", "MIT", ErrInvalidElixirSurface},
		{"package configuration is not a keyword list", "def project do\n  [package: package_config()]\nend\n", "MIT", ErrInvalidElixirSurface},
		{"non-utf-8 target", mixExample, "MIT\xff", ErrInvalidTarget},
		{"quote target", mixExample, "MI\"T", ErrInvalidTarget},
		{"backslash target", mixExample, "MI\\T", ErrInvalidTarget},
		{"hash target", mixExample, "MI#T", ErrInvalidTarget},
		{"empty target", mixExample, "", ErrInvalidTarget},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := AlignElixirLicense(testCase.content, testCase.target)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("AlignElixirLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
			}
		})
	}
}

func TestElixirSkeletonLexerContextForms(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{"paren sigil", "summary = ~s(licenses: [\"MIT\"])\npackage: []\n", ""},
		{"bracket sigil", "summary = ~w[MIT Apache]\npackage: []\n", ""},
		{"brace sigil", "summary = ~s{licenses: [\"MIT\"]}\npackage: []\n", ""},
		{"angle sigil", "summary = ~w<MIT>\npackage: []\n", ""},
		{"same delimiter sigil", "summary = ~s/licenses: [\"MIT\"]/\npackage: []\n", ""},
		{"question delimiter sigil", "summary = ~s?x?\npackage: []\n", ""},
		{"escaped sigil delimiter", "summary = ~s(a\\)b)\npackage: []\n", ""},
		{"nested paired sigil", "summary = ~s(a(b)c)\npackage: []\n", ""},
		{"escaped string quote", "summary = \"a \\\" b\"\npackage: []\n", ""},
		{"escaped charlist quote", "x = 'a \\' b'\npackage: []\n", ""},
		{"interpolation with nested braces", "summary = \"#{Map.get(m, %{a: 1})}\"\npackage: []\n", ""},
		{"interpolation with nested charlist", "summary = \"#{f('x')}\"\npackage: []\n", ""},
		{"interpolation with nested interpolation", "summary = \"#{\"#{a}\"}\"\npackage: []\n", ""},
		{"heredoc with escapes", "@moduledoc \"\"\"\nline \\\" one\n\"\"\"\npackage: []\n", ""},
		{"heredoc sigil", "summary = ~S\"\"\"\nlicenses: [\"MIT\"]\n\"\"\"\npackage: []\n", ""},
		{"empty string then string", "x = \"\" <> \"y\"\npackage: []\n", ""},
		{"char literal at eof", "x = ?a", ""},
		{"tilde without letter", "x = a ~ b\npackage: []\n", ""},
		{"tilde at eof", "x = a ~", ""},
		{"sigil without delimiter", "x = ~s", ""},
		{"sigil kind at eof", "x = ~s", ""},
		{"percent without brace", "x = a % b\npackage: []\n", ""},
		{"percent at eof", "x = a %", ""},
		{"question mark without operand", "x = a ? b\npackage: []\n", ""},
		{"keyword with question suffix", "def empty?, do: true\npackage: []\n", ""},
		{"bang keyword context", "def run!, do: :ok\npackage: []\n", ""},
		{"module attribute atom", "@attr :package\npackage: []\n", ""},
		{"unquote call", "x = unquote(:licenses)\npackage: []\n", ""},
		{"tuple scope", "x = {1, 2}\npackage: []\n", ""},
		{"paren scope", "x = f(1, 2)\npackage: []\n", ""},
		{"map scope", "x = %{a: 1}\npackage: []\n", ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectElixirLicense(testCase.content)
			if err != nil {
				t.Fatalf("InspectElixirLicense(%q) error = %v", testCase.content, err)
			}
			if surface.State != ElixirLicenseMissing {
				t.Fatalf("InspectElixirLicense(%q) state = %d, want %d", testCase.content, surface.State, ElixirLicenseMissing)
			}
			if surface.Value != testCase.want {
				t.Fatalf("InspectElixirLicense(%q) value = %q, want %q", testCase.content, surface.Value, testCase.want)
			}
		})
	}
}

func TestElixirSkeletonLexerRefusesStructuralBreaks(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"unterminated same-delimiter sigil", "summary = ~s/text\n"},
		{"unterminated bracket sigil", "summary = ~s[text\n"},
		{"unterminated brace sigil", "summary = ~s{text\n"},
		{"unterminated angle sigil", "summary = ~s<text\n"},
		{"unterminated heredoc sigil", "summary = ~s\"\"\"\ntext\n"},
		{"unterminated interpolation in charlist context", "x = 'a\n"},
		{"closing bracket without scope", "]"},
		{"closing brace without scope", "}"},
		{"closing paren without scope", ")"},
		{"empty list at eof inside package", "def project do\n  [package: [licenses: [\"MIT\"], x: ["},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := InspectElixirLicense(testCase.content); !errors.Is(err, ErrInvalidElixirSurface) {
				t.Fatalf("InspectElixirLicense(%q) error = %v, want %v", testCase.content, err, ErrInvalidElixirSurface)
			}
		})
	}
}

func FuzzAlignElixirLicense(f *testing.F) {
	f.Add(mixExample, "LicenseRef-example-NoRepublish-1.0")
	f.Add("def project do\n  [package: []]\nend\n", "MIT")
	f.Add("def project do\n  [package: [licenses: [\"MIT\"]]]\nend\n", "Apache-2.0")
	f.Add("def project do\n  [package: [licenses: [\"MIT\", \"Apache-2.0\"]]]\nend\n", "MIT")
	f.Add("def project do\n  [package: [licenses: []]]\nend\n", "MIT")
	f.Add("def project do\n  [package: [licenses: \"MIT\"]]\nend\n", "MIT")
	f.Add("def project do\n  [package: [licenses: ~w(MIT)]]\nend\n", "MIT")
	f.Add("def package do\n  [licenses: [\"MIT\"]]\nend\n", "MIT")
	f.Add("def project do\n  [licenses: [\"MIT\"]]\nend\n", "MIT")
	f.Add("def project do\n  [package: %{licenses: [\"MIT\"]}]\nend\n", "MIT")
	f.Add("def project do\n  [package: [links: %{licenses: [\"MIT\"]}]]\nend\n", "MIT")
	f.Add("def project do\n  [package: [licenses: [\"MIT\"], licenses: [\"MIT\"]]]\nend\n", "MIT")
	f.Add("summary = \"a #{b\"\n", "MIT")
	f.Add("@moduledoc \"\"\"\nlicenses: [\"MIT\"]\n\"\"\"\n", "MIT")
	f.Add("summary = ~s(licenses: [\"MIT\"])\n", "MIT")
	f.Add("def project do\n  [package: [licenses: [\"M\\\\IT\"]]]\nend\n", "MIT")
	f.Add("def project do\n  [package: [licenses: [:mit]]]\nend\n", "MIT")
	f.Add("defmodule X do\nend\n", "MIT")
	f.Add("]\n", "MIT")
	f.Add("def project do\n  [package: [# c\n  ]]\nend\n", "MIT")
	f.Fuzz(func(t *testing.T, content, target string) {
		aligned, changed, err := AlignElixirLicense(content, target)
		if err != nil {
			return
		}
		surface, err := InspectElixirLicense(aligned)
		if err != nil || surface.State != ElixirLicenseList ||
			surface.Value != target || len(surface.Entries) > 1 {
			t.Fatalf("AlignElixirLicense produced an unprovable surface: %q -> %q (%q, %v)", content, aligned, surface.Value, err)
		}
		if !changed {
			return
		}
		again, changedAgain, err := AlignElixirLicense(aligned, target)
		if err != nil || changedAgain || again != aligned {
			t.Fatalf("AlignElixirLicense is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
		}
	})
}
