package ecosystem

import (
	"errors"
	"strings"
	"testing"
)

// Convention: docs/conventions/cli/testing/README.md

func TestGemspecProjectionDerivesFromTheTenantValues(t *testing.T) {
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
			if got := GemspecProjection(testCase.merged); got != testCase.want {
				t.Fatalf("GemspecProjection() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestGemspecManifestNamesFiltersTheDirectoryListing(t *testing.T) {
	entries := []string{"readme.md", "B.gemspec", "example.gemspec", "App.GEMSPEC", "LICENSE", "subdir"}
	got := GemspecManifestNames(entries)
	want := []string{"App.GEMSPEC", "B.gemspec", "example.gemspec"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("GemspecManifestNames() = %v, want %v", got, want)
	}
	if other := GemspecManifestNames([]string{"readme.md", "LICENSE"}); len(other) != 0 {
		t.Fatalf("GemspecManifestNames() = %v, want no candidates", other)
	}
}

const gemspecExample = "Gem::Specification.new do |spec|\n" +
	"  spec.name = \"example-project\"\n" +
	"  spec.version = \"1.0.0\"\n" +
	"  spec.license = \"MIT\"\n" +
	"  spec.summary = \"example\"\n" +
	"end\n"

func TestInspectGemspecLicenseClassifiesTheAssignment(t *testing.T) {
	cases := []struct {
		name       string
		content    string
		wantValue  string
		wantState  GemspecLicenseState
		wantErrIs  error
		wantErr    bool
		wantValues []string
	}{
		{"string field", gemspecExample, "MIT", GemspecLicenseString, nil, false, nil},
		{"missing assignment", "Gem::Specification.new do |spec|\n  spec.name = \"x\"\nend\n", "", GemspecLicenseMissing, nil, false, nil},
		{"single-quoted field", "spec.license = 'MIT'\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"plural bracket array", "spec.licenses = [\"MIT\"]\n", "MIT", GemspecLicenseArray, nil, false, []string{"MIT"}},
		{"multi-element bracket array", "spec.licenses = [\"MIT\", \"Apache-2.0\"]\n", "", GemspecLicenseArray, nil, false, []string{"MIT", "Apache-2.0"}},
		{"plural percent array", "spec.licenses = %w[MIT]\n", "MIT", GemspecLicenseArray, nil, false, []string{"MIT"}},
		{"multi-word percent array", "spec.licenses = %w[MIT Apache-2.0]\n", "", GemspecLicenseArray, nil, false, []string{"MIT", "Apache-2.0"}},
		{"paren percent array", "spec.licenses = %w(MIT)\n", "MIT", GemspecLicenseArray, nil, false, []string{"MIT"}},
		{"brace percent array", "spec.licenses = %w{MIT}\n", "MIT", GemspecLicenseArray, nil, false, []string{"MIT"}},
		{"angle percent array", "spec.licenses = %w<MIT>\n", "MIT", GemspecLicenseArray, nil, false, []string{"MIT"}},
		{"receiver-agnostic", "gem.license = \"MIT\"\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"instance variable receiver", "@spec.license = \"MIT\"\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"padded value", "spec.license =   \"MIT\"\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"value on the next line", "spec.license =\n  \"MIT\"\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"comment before the value", "spec.license = # the license\n  \"MIT\"\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"freeze suffix", "spec.license = \"MIT\".freeze\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"spaced freeze suffix", "spec.license = \"MIT\" . freeze\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"crlf line ending", "spec.license = \"MIT\"\r\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"escaped context string", "spec.summary = \"a \\\"quoted\\\" note\"\n", "", GemspecLicenseMissing, nil, false, nil},
		{"comment context", "# spec.license = \"MIT\"\n", "", GemspecLicenseMissing, nil, false, nil},
		{"heredoc context", "spec.description = <<~DESC\nlicense = \"MIT\"\nDESC\n", "", GemspecLicenseMissing, nil, false, nil},
		{"dash heredoc context", "spec.description = <<-DESC\nlicense = \"MIT\"\n  DESC\n", "", GemspecLicenseMissing, nil, false, nil},
		{"quoted heredoc context", "spec.description = <<~\"DESC\"\nlicense = \"MIT\"\nDESC\n", "", GemspecLicenseMissing, nil, false, nil},
		{"paren percent literal context", "spec.summary = %q(license = \"MIT\")\n", "", GemspecLicenseMissing, nil, false, nil},
		{"angle percent literal context", "spec.summary = %q<license = \"MIT\">\n", "", GemspecLicenseMissing, nil, false, nil},
		{"non-paired percent literal context", "spec.summary = %q!license = \"MIT\"!\n", "", GemspecLicenseMissing, nil, false, nil},
		{"generic percent literal context", "spec.summary = %{license = \"MIT\"}\n", "", GemspecLicenseMissing, nil, false, nil},
		{"escaped percent literal context", "spec.summary = %q{a \\} b}\n", "", GemspecLicenseMissing, nil, false, nil},
		{"nested paired percent literal", "spec.summary = %q(a(b)c)\n", "", GemspecLicenseMissing, nil, false, nil},
		{"block comment context", "=begin\nspec.license = \"MIT\"\n=end\n", "", GemspecLicenseMissing, nil, false, nil},
		{"line starting with equals", "= x\nspec.license = \"MIT\"\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"modulo not a literal", "x = y % 2\nspec.license = \"MIT\"\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"less-than comparison context", "x = y < z\nspec.license = \"MIT\"\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"shift operator context", "x = y << 2\nspec.license = \"MIT\"\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"number field", "spec.license = 7\n", "", GemspecLicenseInvalid, nil, false, nil},
		{"symbol field", "spec.license = :mit\n", "", GemspecLicenseInvalid, nil, false, nil},
		{"nil field", "spec.license = nil\n", "", GemspecLicenseInvalid, nil, false, nil},
		{"hash field", "spec.license = { a: 1 }\n", "", GemspecLicenseInvalid, nil, false, nil},
		{"paren group field", "spec.license = (\"MIT\")\n", "", GemspecLicenseInvalid, nil, false, nil},
		{"comparison not assignment", "spec.license == \"MIT\"\n", "", GemspecLicenseMissing, nil, false, nil},
		{"regex match not assignment", "spec.license =~ /MIT/\n", "", GemspecLicenseMissing, nil, false, nil},
		{"hashrocket not assignment", "h = { spec.license => 1 }\n", "", GemspecLicenseMissing, nil, false, nil},
		{"method read not assignment", "puts spec.license\n", "", GemspecLicenseMissing, nil, false, nil},
		{"empty array", "spec.licenses = []\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"non-string entry", "spec.licenses = [\"MIT\", 7]\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"symbol entry", "spec.licenses = [:mit]\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"nested array entry", "spec.licenses = [[\"MIT\"]]\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"entry with escape", "spec.licenses = [\"M\\IT\"]\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"interpolated value", "spec.license = \"MIT#{x}\"\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"interpolated percent array", "spec.licenses = %W[MIT #{x}]\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"dynamic suffix", "spec.license = \"MIT\".upcase\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"double freeze", "spec.license = \"MIT\".freeze.freeze\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"trailing expression", "spec.license = \"MIT\" or \"Apache-2.0\"\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"star field", "spec.license = *x\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"malformed delimiter in array", "spec.licenses = [\"MIT\" \"Apache-2.0\"]\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"percent array without delimiter", "spec.licenses = %wx\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"truncated percent array", "spec.licenses = %w", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"chained receiver", "a.b.license = \"MIT\"\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"unattributable receiver", ".license = \"MIT\"\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"assignment without value", "spec.license =\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"duplicate assignments", "spec.license = \"MIT\"\nspec.licenses = [\"MIT\"]\n", "", 0, ErrAmbiguousGemspecLicense, true, nil},
		{"two singular assignments", "spec.license = \"MIT\"\nspec.license = \"Apache-2.0\"\n", "", 0, ErrAmbiguousGemspecLicense, true, nil},
		{"malformed string", "spec.license = \"MIT\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"unterminated array", "spec.licenses = [\"MIT\"\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"array open at eof", "spec.licenses = [", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"unterminated percent array", "spec.licenses = %w[MIT\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"unterminated heredoc", "spec.description = <<~DESC\ntext\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"unterminated quoted terminator", "spec.description = <<\"EOS\ntext\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"unterminated percent literal", "spec.summary = %q{ text\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"unterminated non-paired percent literal", "spec.summary = %q! text\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"unterminated block comment", "=begin\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"bracket percent literal context", "spec.summary = %q[license = \"MIT\"]\n", "", GemspecLicenseMissing, nil, false, nil},
		{"unterminated context string", "spec.summary = \"unterminated\n", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"escaped percent array", "spec.licenses = %w[M\\IT]\n", "", 0, ErrInvalidGemspecLicenseValue, true, nil},
		{"nested percent array", "spec.licenses = %w[a[b] c]\n", "", GemspecLicenseArray, nil, false, []string{"a[b]", "c"}},
		{"unterminated paren group field", "spec.license = (\"MIT\"\n", "", GemspecLicenseInvalid, nil, false, nil},
		{"percent at eof", "spec.license = \"MIT\"\nx = y %", "MIT", GemspecLicenseString, nil, false, nil},
		{"unknown percent kind", "spec.summary = %a{x}\nspec.license = \"MIT\"\n", "MIT", GemspecLicenseString, nil, false, nil},
		{"percent kind at eof", "spec.license = \"MIT\"\nx = y %w", "MIT", GemspecLicenseString, nil, false, nil},
		{"tilde at eof", "spec.license = \"MIT\"\nx = y <<~", "MIT", GemspecLicenseString, nil, false, nil},
		{"heredoc with trailing method", "spec.description = <<~EOS.strip\nlicense = \"MIT\"\nEOS\n", "", GemspecLicenseMissing, nil, false, nil},
		{"heredoc opener at eof", "spec.description = <<EOS", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"unterminated block comment without newline", "=begin text", "", 0, ErrInvalidGemspecSurface, true, nil},
		{"unterminated block comment body", "=begin\nbody text", "", 0, ErrInvalidGemspecSurface, true, nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectGemspecLicense(testCase.content)
			if testCase.wantErr {
				if !errors.Is(err, testCase.wantErrIs) {
					t.Fatalf("InspectGemspecLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("InspectGemspecLicense() error = %v", err)
			}
			if surface.State != testCase.wantState {
				t.Fatalf("InspectGemspecLicense() state = %d, want %d", surface.State, testCase.wantState)
			}
			if surface.Value != testCase.wantValue {
				t.Fatalf("InspectGemspecLicense() value = %q, want %q", surface.Value, testCase.wantValue)
			}
			if len(surface.Entries) != len(testCase.wantValues) {
				t.Fatalf("InspectGemspecLicense() entries = %v, want %v", surface.Entries, testCase.wantValues)
			}
			for i, want := range testCase.wantValues {
				if surface.Entries[i] != want {
					t.Fatalf("InspectGemspecLicense() entries = %v, want %v", surface.Entries, testCase.wantValues)
				}
			}
		})
	}
}

func TestAlignGemspecLicenseReplacesADivergingValue(t *testing.T) {
	aligned, changed, err := AlignGemspecLicense(gemspecExample, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignGemspecLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignGemspecLicense() changed = false, want true")
	}
	want := strings.Replace(gemspecExample, `spec.license = "MIT"`, `spec.license = "LicenseRef-example-NoRepublish-1.0"`, 1)
	if aligned != want {
		t.Fatalf("AlignGemspecLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignGemspecLicenseNormalizesADivergingOneElementArray(t *testing.T) {
	aligned, changed, err := AlignGemspecLicense("spec.licenses = [\"MIT\"]\n", "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignGemspecLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignGemspecLicense() changed = false, want true")
	}
	want := "spec.license = \"LicenseRef-example-NoRepublish-1.0\"\n"
	if aligned != want {
		t.Fatalf("AlignGemspecLicense() = %q, want %q", aligned, want)
	}
}

func TestAlignGemspecLicenseIsIdempotentOnAnAlignedArraySpelling(t *testing.T) {
	content := "spec.licenses = [\"LicenseRef-example-NoRepublish-1.0\"]\n"
	aligned, changed, err := AlignGemspecLicense(content, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignGemspecLicense() error = %v", err)
	}
	if changed {
		t.Fatal("AlignGemspecLicense() changed = true, want false")
	}
	if aligned != content {
		t.Fatalf("AlignGemspecLicense() mutated an aligned gemspec: %q", aligned)
	}
}

func TestAlignGemspecLicenseIsIdempotentOnAlignedState(t *testing.T) {
	aligned, changed, err := AlignGemspecLicense(gemspecExample, "MIT")
	if err != nil {
		t.Fatalf("AlignGemspecLicense() error = %v", err)
	}
	if changed {
		t.Fatal("AlignGemspecLicense() changed = true, want false")
	}
	if aligned != gemspecExample {
		t.Fatalf("AlignGemspecLicense() mutated an aligned gemspec: %q", aligned)
	}
}

func TestAlignGemspecLicensePreservesEveryNonLicenseByte(t *testing.T) {
	content := "Gem::Specification.new do |spec|\n  spec.name = \"x\"\n  spec.license = \"MIT\" # the license\nend\n"
	aligned, _, err := AlignGemspecLicense(content, "LicenseRef-example-NoRepublish-1.0")
	if err != nil {
		t.Fatalf("AlignGemspecLicense() error = %v", err)
	}
	before, _, _ := strings.Cut(content, `spec.license = "MIT"`)
	afterOriginal := content[strings.Index(content, `spec.license = "MIT"`)+len(`spec.license = "MIT"`):]
	if !strings.HasPrefix(aligned, before+`spec.license = "LicenseRef-example-NoRepublish-1.0"`) {
		t.Fatalf("AlignGemspecLicense() changed bytes before the value: %q", aligned)
	}
	if !strings.HasSuffix(aligned, afterOriginal) {
		t.Fatalf("AlignGemspecLicense() changed bytes after the value: %q", aligned)
	}
	if !strings.Contains(aligned, "spec.name = \"x\"") {
		t.Fatalf("AlignGemspecLicense() touched a non-license assignment: %q", aligned)
	}
}

func TestAlignGemspecLicenseInsertsAMissingAssignment(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"with body", "Gem::Specification.new do |spec|\n  spec.name = \"x\"\nend\n"},
		{"empty block", "Gem::Specification.new do |spec|\nend\n"},
		{"blank body lines", "Gem::Specification.new do |spec|\n\n\nend\n"},
		{"no trailing newline", "Gem::Specification.new do |spec|"},
		{"trailing do-line text", "Gem::Specification.new do |spec|   \n  spec.name = \"x\"\nend\n"},
		{"trailing blank body without newline", "Gem::Specification.new do |spec|\n   "},
		{"indented do-line without body", "  Gem::Specification.new do |spec|"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignGemspecLicense(testCase.content, "LicenseRef-example-NoRepublish-1.0")
			if err != nil {
				t.Fatalf("AlignGemspecLicense() error = %v", err)
			}
			if !changed {
				t.Fatal("AlignGemspecLicense() changed = false, want true")
			}
			surface, err := InspectGemspecLicense(aligned)
			if err != nil || surface.State != GemspecLicenseString {
				t.Fatalf("AlignGemspecLicense() produced an unscannable surface: %q (%v)", aligned, err)
			}
			if surface.Value != "LicenseRef-example-NoRepublish-1.0" {
				t.Fatalf("AlignGemspecLicense() license = %q", surface.Value)
			}
			if testCase.name == "with body" && !strings.Contains(aligned, "spec.name = \"x\"") {
				t.Fatalf("AlignGemspecLicense() lost the original body: %q", aligned)
			}
		})
	}
}

func TestAlignGemspecLicenseRefusesAmbiguousAndInvalidSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		target    string
		wantErrIs error
	}{
		{"multiple entries", "spec.licenses = [\"MIT\", \"Apache-2.0\"]\n", "MIT", ErrMultipleGemspecLicenses},
		{"invalid value", "spec.license = 7\n", "MIT", ErrInvalidGemspecLicenseValue},
		{"empty array", "spec.licenses = []\n", "MIT", ErrInvalidGemspecLicenseValue},
		{"duplicate assignments", "spec.license = \"MIT\"\nspec.licenses = [\"MIT\"]\n", "MIT", ErrAmbiguousGemspecLicense},
		{"malformed gemspec", "spec.license = \"MIT\n", "MIT", ErrInvalidGemspecSurface},
		{"no specification block", "x = 1\n", "MIT", ErrInvalidGemspecSurface},
		{"non-utf-8 target", gemspecExample, "MIT\xff", ErrInvalidTarget},
		{"whitespace target", gemspecExample, "MIT OR Apache-2.0", ErrInvalidTarget},
		{"overlong target", gemspecExample, strings.Repeat("A", 65), ErrInvalidTarget},
		{"quote target", gemspecExample, "MI\"T", ErrInvalidTarget},
		{"hash target", gemspecExample, "MI#T", ErrInvalidTarget},
		{"empty target", gemspecExample, "", ErrInvalidTarget},
		{"anchor without do block", "Gem::Specification.new(1) do |spec|\nend\n", "MIT", ErrInvalidGemspecSurface},
		{"anchor without block parameter", "Gem::Specification.new do x\n", "MIT", ErrInvalidGemspecSurface},
		{"anchor without closing pipe", "Gem::Specification.new do |spec\n", "MIT", ErrInvalidGemspecSurface},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := AlignGemspecLicense(testCase.content, testCase.target)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("AlignGemspecLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
			}
		})
	}
}

func TestAlignGemspecLicenseRefusesAnUnprovableInsertionAnchor(t *testing.T) {
	content := "=begin\nGem::Specification.new do |spec|\n=end\n"
	_, _, err := AlignGemspecLicense(content, "LicenseRef-example-NoRepublish-1.0")
	if !errors.Is(err, ErrInvalidGemspecSurface) {
		t.Fatalf("AlignGemspecLicense(%q) error = %v, want ErrInvalidGemspecSurface", content, err)
	}
}

func FuzzAlignGemspecLicense(f *testing.F) {
	f.Add(gemspecExample, "LicenseRef-example-NoRepublish-1.0")
	f.Add("Gem::Specification.new do |spec|\nend\n", "MIT")
	f.Add("spec.license = \"MIT\"\n", "Apache-2.0")
	f.Add("spec.licenses = [\"MIT\", \"Apache-2.0\"]\n", "MIT")
	f.Add("spec.licenses = %w[MIT Apache-2.0]\n", "MIT")
	f.Add("spec.license = \"MIT\".freeze\n", "MIT")
	f.Add("spec.license = \"MIT#{x}\"\n", "MIT")
	f.Add("spec.license = \"MIT\n", "MIT")
	f.Add("spec.licenses = []\n", "MIT")
	f.Add("a.b.license = \"MIT\"\n", "MIT")
	f.Add("Gem::Specification.new(1) do |spec|\nspec.license = \"MIT\"\nend\n", "MIT")
	f.Add("spec.license = 7\n", "MIT")
	f.Add("=begin\nGem::Specification.new do |spec|\n=end\n", "MIT")
	f.Add("# Gem::Specification.new do |spec|\n", "MIT")
	f.Fuzz(func(t *testing.T, content, target string) {
		aligned, changed, err := AlignGemspecLicense(content, target)
		if err != nil {
			return
		}
		surface, err := InspectGemspecLicense(aligned)
		if err != nil || (surface.State != GemspecLicenseString && surface.State != GemspecLicenseArray) ||
			surface.Value != target || len(surface.Entries) > 1 {
			t.Fatalf("AlignGemspecLicense produced an unprovable surface: %q -> %q (%q, %v)", content, aligned, surface.Value, err)
		}
		if !changed {
			return
		}
		again, changedAgain, err := AlignGemspecLicense(aligned, target)
		if err != nil || changedAgain || again != aligned {
			t.Fatalf("AlignGemspecLicense is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
		}
	})
}
