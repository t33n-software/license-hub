package ecosystem

import (
	"errors"
	"strings"
	"testing"
)

// Convention: docs/conventions/cli/testing/README.md

func TestMavenProjectionDerivesFromTheTenantValues(t *testing.T) {
	cases := []struct {
		name   string
		merged map[string]string
		want   MavenLicenseForm
	}{
		{"custom family", map[string]string{"LICENSE_ID": "example-NoRepublish-1.0", "CANONICAL_SOURCE_URL": "https://github.com/t33n-software/example"}, MavenLicenseForm{Name: "example-NoRepublish-1.0", URL: "https://github.com/t33n-software/example"}},
		{"spdx identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "MIT", "LICENSE_ID": "example-NoRepublish-1.0", "CANONICAL_SOURCE_URL": "https://github.com/t33n-software/example"}, MavenLicenseForm{Name: "MIT", URL: "https://github.com/t33n-software/example"}},
		{"padded identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "  MIT  ", "CANONICAL_SOURCE_URL": "https://github.com/t33n-software/example"}, MavenLicenseForm{Name: "MIT", URL: "https://github.com/t33n-software/example"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := MavenProjection(testCase.merged)
			if got.Name != testCase.want.Name || got.URL != testCase.want.URL {
				t.Fatalf("MavenProjection() = %v, want %v", got, testCase.want)
			}
		})
	}
}

const examplePom = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<project xmlns=\"http://maven.apache.org/POM/4.0.0\">\n    <modelVersion>4.0.0</modelVersion>\n</project>\n"

const pomWithLicense = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<project xmlns=\"http://maven.apache.org/POM/4.0.0\">\n    <licenses>\n        <license>\n            <name>example-NoRepublish-1.0</name>\n            <url>https://github.com/t33n-software/example</url>\n        </license>\n    </licenses>\n</project>\n"

func TestInspectMavenLicenseClassifiesTheSurface(t *testing.T) {
	cases := []struct {
		name         string
		content      string
		wantPresent  bool
		wantCount    int
		wantName     bool
		wantNameText string
		wantURL      bool
		wantURLText  string
	}{
		{"declared surface", pomWithLicense, true, 1, true, "example-NoRepublish-1.0", true, "https://github.com/t33n-software/example"},
		{"missing licenses", examplePom, false, 0, false, "", false, ""},
		{"empty licenses", "<project>\n    <licenses/>\n</project>\n", true, 0, false, "", false, ""},
		{"empty license", "<project>\n    <licenses>\n        <license/>\n    </licenses>\n</project>\n", true, 1, false, "", false, ""},
		{"missing name", "<project>\n    <licenses>\n        <license>\n            <url>https://example.org</url>\n        </license>\n    </licenses>\n</project>\n", true, 1, false, "", true, "https://example.org"},
		{"missing url", "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n", true, 1, true, "MIT", false, ""},
		{"empty document", "", false, 0, false, "", false, ""},
		{"entities decode", "<project>\n    <licenses>\n        <license>\n            <name>A &amp; B</name>\n            <url>https://example.org/?a=1&amp;b=2</url>\n        </license>\n    </licenses>\n</project>\n", true, 1, true, "A & B", true, "https://example.org/?a=1&b=2"},
		{"cdata decodes", "<project>\n    <licenses>\n        <license>\n            <name><![CDATA[MIT]]></name>\n        </license>\n    </licenses>\n</project>\n", true, 1, true, "MIT", false, ""},
		{"comment context", "<project>\n    <!-- the license -->\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n", true, 1, true, "MIT", false, ""},
		{"other elements carry no weight", "<project>\n    <developers>\n        <developer>\n            <name>Developer</name>\n            <url>https://example.org/dev</url>\n        </developer>\n    </developers>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n", true, 1, true, "MIT", false, ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectMavenLicense(testCase.content)
			if err != nil {
				t.Fatalf("InspectMavenLicense() error = %v", err)
			}
			if surface.LicensesPresent != testCase.wantPresent {
				t.Fatalf("InspectMavenLicense() licensesPresent = %v, want %v", surface.LicensesPresent, testCase.wantPresent)
			}
			if surface.LicenseCount != testCase.wantCount {
				t.Fatalf("InspectMavenLicense() licenseCount = %d, want %d", surface.LicenseCount, testCase.wantCount)
			}
			if surface.NamePresent != testCase.wantName || surface.Name != testCase.wantNameText {
				t.Fatalf("InspectMavenLicense() name = (%v, %q), want (%v, %q)", surface.NamePresent, surface.Name, testCase.wantName, testCase.wantNameText)
			}
			if surface.URLPresent != testCase.wantURL || surface.URL != testCase.wantURLText {
				t.Fatalf("InspectMavenLicense() url = (%v, %q), want (%v, %q)", surface.URLPresent, surface.URL, testCase.wantURL, testCase.wantURLText)
			}
		})
	}
}

func TestInspectMavenLicenseRefusesUnprovableSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		wantErrIs error
	}{
		{"root is not project", "<distributionManagement>\n</distributionManagement>\n", ErrInvalidMavenSurface},
		{"unclosed element", "<project>\n    <licenses>\n", ErrInvalidMavenSurface},
		{"mismatched end tag", "<project>\n    <licenses>\n    </project>\n", ErrInvalidMavenSurface},
		{"bad entity", "<project>\n    <licenses>&unknown;</licenses>\n</project>\n", ErrInvalidMavenSurface},
		{"duplicate licenses", "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n    <licenses/>\n</project>\n", ErrAmbiguousMavenField},
		{"duplicate name", "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n            <name>Apache-2.0</name>\n        </license>\n    </licenses>\n</project>\n", ErrAmbiguousMavenField},
		{"duplicate url", "<project>\n    <licenses>\n        <license>\n            <url>https://a</url>\n            <url>https://b</url>\n        </license>\n    </licenses>\n</project>\n", ErrAmbiguousMavenField},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := InspectMavenLicense(testCase.content)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("InspectMavenLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
			}
		})
	}
}

func TestAlignMavenLicenseInsertsTheFullBlock(t *testing.T) {
	form := MavenLicenseForm{Name: "license-hub-NoRepublish-1.0", URL: "https://github.com/t33n-software/license-hub"}
	aligned, changed, err := AlignMavenLicense(examplePom, form)
	if err != nil {
		t.Fatalf("AlignMavenLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignMavenLicense() changed = false, want true")
	}
	want := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<project xmlns=\"http://maven.apache.org/POM/4.0.0\">" +
		"\n    <licenses>\n        <license>\n            <name>license-hub-NoRepublish-1.0</name>\n" +
		"            <url>https://github.com/t33n-software/license-hub</url>\n        </license>\n    </licenses>" +
		"\n    <modelVersion>4.0.0</modelVersion>\n</project>\n"
	if aligned != want {
		t.Fatalf("AlignMavenLicense() = %q, want %q", aligned, want)
	}
	surface, err := InspectMavenLicense(aligned)
	if err != nil {
		t.Fatalf("InspectMavenLicense() error = %v", err)
	}
	if !surface.LicensesPresent || surface.LicenseCount != 1 || !surface.NamePresent || surface.Name != form.Name || !surface.URLPresent || surface.URL != form.URL {
		t.Fatalf("AlignMavenLicense() produced an unprovable surface: %+v", surface)
	}
}

func TestAlignMavenLicenseInsertsTheLicenseElement(t *testing.T) {
	form := MavenLicenseForm{Name: "MIT", URL: "https://example.org"}
	aligned, changed, err := AlignMavenLicense("<project>\n    <licenses>\n    </licenses>\n</project>\n", form)
	if err != nil {
		t.Fatalf("AlignMavenLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignMavenLicense() changed = false, want true")
	}
	want := "<project>\n    <licenses>\n        <license><name>MIT</name><url>https://example.org</url></license>\n    </licenses>\n</project>\n"
	if aligned != want {
		t.Fatalf("AlignMavenLicense() = %q, want %q", aligned, want)
	}
	expanded, changed, err := AlignMavenLicense("<project>\n    <licenses/>\n</project>\n", form)
	if err != nil {
		t.Fatalf("AlignMavenLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignMavenLicense() changed = false, want true")
	}
	want = "<project>\n    <licenses>\n        <license><name>MIT</name><url>https://example.org</url></license></licenses>\n</project>\n"
	if expanded != want {
		t.Fatalf("AlignMavenLicense() = %q, want %q", expanded, want)
	}
}

func TestAlignMavenLicenseInsertsTheMissingChildren(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{"both children missing", "<project>\n    <licenses>\n        <license>\n        </license>\n    </licenses>\n</project>\n", "<project>\n    <licenses>\n        <license><name>MIT</name><url>https://example.org</url>\n        </license>\n    </licenses>\n</project>\n"},
		{"self-closing license", "<project>\n    <licenses>\n        <license/>\n    </licenses>\n</project>\n", "<project>\n    <licenses>\n        <license><name>MIT</name><url>https://example.org</url></license>\n    </licenses>\n</project>\n"},
		{"url missing", "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n", "<project>\n    <licenses>\n        <license><url>https://example.org</url>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n"},
		{"self-closing name", "<project>\n    <licenses>\n        <license>\n            <name/>\n            <url>https://example.org</url>\n        </license>\n    </licenses>\n</project>\n", "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n            <url>https://example.org</url>\n        </license>\n    </licenses>\n</project>\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignMavenLicense(testCase.content, MavenLicenseForm{Name: "MIT", URL: "https://example.org"})
			if err != nil {
				t.Fatalf("AlignMavenLicense() error = %v", err)
			}
			if !changed {
				t.Fatal("AlignMavenLicense() changed = false, want true")
			}
			if aligned != testCase.want {
				t.Fatalf("AlignMavenLicense() = %q, want %q", aligned, testCase.want)
			}
		})
	}
}

func TestAlignMavenLicenseReplacesADivergingValue(t *testing.T) {
	form := MavenLicenseForm{Name: "license-hub-NoRepublish-1.0", URL: "https://github.com/t33n-software/license-hub"}
	aligned, changed, err := AlignMavenLicense(pomWithLicense, form)
	if err != nil {
		t.Fatalf("AlignMavenLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignMavenLicense() changed = false, want true")
	}
	if !strings.Contains(aligned, "<name>license-hub-NoRepublish-1.0</name>") || !strings.Contains(aligned, "<url>https://github.com/t33n-software/license-hub</url>") {
		t.Fatalf("AlignMavenLicense() = %q", aligned)
	}
}

func TestAlignMavenLicenseExpandsASelfClosingProjectRoot(t *testing.T) {
	form := MavenLicenseForm{Name: "MIT", URL: "https://example.org"}
	aligned, changed, err := AlignMavenLicense("<project/>", form)
	if err != nil {
		t.Fatalf("AlignMavenLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignMavenLicense() changed = false, want true")
	}
	want := "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n            <url>https://example.org</url>\n        </license>\n    </licenses></project>"
	if aligned != want {
		t.Fatalf("AlignMavenLicense() = %q, want %q", aligned, want)
	}
	surface, err := InspectMavenLicense(aligned)
	if err != nil {
		t.Fatalf("InspectMavenLicense() error = %v", err)
	}
	if !surface.LicensesPresent || surface.Name != "MIT" {
		t.Fatalf("AlignMavenLicense() produced an unprovable surface: %+v", surface)
	}
}

func TestAlignMavenLicenseExpansionKeepsTheEndTagNameClean(t *testing.T) {
	// Regression: the expansion of a self-closing start tag that carries
	// attributes must reconstruct the end tag from the element name alone.
	form := MavenLicenseForm{Name: "MIT", URL: "https://example.org"}
	aligned, changed, err := AlignMavenLicense("<project xmlns=\"http://maven.apache.org/POM/4.0.0\"/>", form)
	if err != nil {
		t.Fatalf("AlignMavenLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignMavenLicense() changed = false, want true")
	}
	want := "<project xmlns=\"http://maven.apache.org/POM/4.0.0\">\n    <licenses>\n        <license>\n            <name>MIT</name>\n            <url>https://example.org</url>\n        </license>\n    </licenses></project>"
	if aligned != want {
		t.Fatalf("AlignMavenLicense() = %q, want %q", aligned, want)
	}
	if _, err := InspectMavenLicense(aligned); err != nil {
		t.Fatalf("AlignMavenLicense() produced an unprovable surface: %q (%v)", aligned, err)
	}
}

func TestAlignMavenLicenseIsIdempotentOnAlignedState(t *testing.T) {
	form := MavenLicenseForm{Name: "example-NoRepublish-1.0", URL: "https://github.com/t33n-software/example"}
	aligned, changed, err := AlignMavenLicense(pomWithLicense, form)
	if err != nil {
		t.Fatalf("AlignMavenLicense() error = %v", err)
	}
	if changed {
		t.Fatal("AlignMavenLicense() changed = true, want false")
	}
	if aligned != pomWithLicense {
		t.Fatalf("AlignMavenLicense() mutated an aligned manifest: %q", aligned)
	}
}

func TestAlignMavenLicensePreservesEveryNonLicenseByte(t *testing.T) {
	content := "<?xml version=\"1.0\"?>\n<project>\n    <!-- the metadata -->\n    <modelVersion>4.0.0</modelVersion>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n"
	aligned, _, err := AlignMavenLicense(content, MavenLicenseForm{Name: "MIT", URL: "https://example.org"})
	if err != nil {
		t.Fatalf("AlignMavenLicense() error = %v", err)
	}
	if !strings.Contains(aligned, "<!-- the metadata -->") {
		t.Fatalf("AlignMavenLicense() touched a comment: %q", aligned)
	}
	if !strings.Contains(aligned, "<modelVersion>4.0.0</modelVersion>") {
		t.Fatalf("AlignMavenLicense() touched a foreign element: %q", aligned)
	}
	if !strings.Contains(aligned, "<url>https://example.org</url>") {
		t.Fatalf("AlignMavenLicense() lost the inserted url: %q", aligned)
	}
}

func TestAlignMavenLicenseEscapesTheProjectedValues(t *testing.T) {
	form := MavenLicenseForm{Name: "A & B <C>", URL: "https://example.org/?a=1&b=2"}
	aligned, changed, err := AlignMavenLicense("<project>\n    <licenses>\n        <license>\n            <name>x</name>\n            <url>y</url>\n        </license>\n    </licenses>\n</project>\n", form)
	if err != nil {
		t.Fatalf("AlignMavenLicense() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignMavenLicense() changed = false, want true")
	}
	surface, err := InspectMavenLicense(aligned)
	if err != nil {
		t.Fatalf("InspectMavenLicense() error = %v", err)
	}
	if surface.Name != form.Name || surface.URL != form.URL {
		t.Fatalf("InspectMavenLicense() = (%q, %q), want the decoded round trip (%q, %q)", surface.Name, surface.URL, form.Name, form.URL)
	}
}

func TestAlignMavenLicenseRefusesInvalidSurfaces(t *testing.T) {
	cases := []struct {
		name      string
		content   string
		form      MavenLicenseForm
		wantErrIs error
	}{
		{"multiple license elements", "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n        <license>\n            <name>Apache-2.0</name>\n        </license>\n    </licenses>\n</project>\n", MavenLicenseForm{Name: "MIT", URL: "https://example.org"}, ErrMultipleMavenLicenses},
		{"malformed surface", "<project>\n    <licenses>\n", MavenLicenseForm{Name: "MIT", URL: "https://example.org"}, ErrInvalidMavenSurface},
		{"root is not project", "<notaproject/>\n", MavenLicenseForm{Name: "MIT", URL: "https://example.org"}, ErrInvalidMavenSurface},
		{"no project root", "", MavenLicenseForm{Name: "MIT", URL: "https://example.org"}, ErrInvalidMavenSurface},
		{"non-utf-8 name", pomWithLicense, MavenLicenseForm{Name: "MIT\xff", URL: "https://example.org"}, ErrInvalidTarget},
		{"non-utf-8 url", pomWithLicense, MavenLicenseForm{Name: "MIT", URL: "https://example.org/\xff"}, ErrInvalidTarget},
		{"unrepresentable name", pomWithLicense, MavenLicenseForm{Name: "MIT\x00", URL: "https://example.org"}, ErrInvalidTarget},
		{"unrepresentable url", pomWithLicense, MavenLicenseForm{Name: "MIT", URL: "https://example.org/\x00"}, ErrInvalidTarget},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := AlignMavenLicense(testCase.content, testCase.form)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("AlignMavenLicense(%q) error = %v, want %v", testCase.content, err, testCase.wantErrIs)
			}
		})
	}
}

func FuzzAlignMavenLicense(f *testing.F) {
	f.Add(examplePom, "license-hub-NoRepublish-1.0", "https://github.com/t33n-software/license-hub")
	f.Add(pomWithLicense, "MIT", "https://example.org")
	f.Add("<project>\n    <licenses/>\n</project>\n", "MIT", "https://example.org")
	f.Add("<project>\n    <licenses>\n        <license/>\n    </licenses>\n</project>\n", "MIT", "https://example.org")
	f.Add("<project/>", "MIT", "https://example.org")
	f.Add("<project xmlns=\"http://maven.apache.org/POM/4.0.0\"/>", "MIT", "https://example.org")
	f.Add("", "MIT", "https://example.org")
	f.Add("<project>\n    <licenses>\n", "MIT", "https://example.org")
	f.Add("<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n        <license/>\n    </licenses>\n</project>\n", "MIT", "https://example.org")
	f.Fuzz(func(t *testing.T, content, name, url string) {
		form := MavenLicenseForm{Name: name, URL: url}
		aligned, changed, err := AlignMavenLicense(content, form)
		if err != nil {
			return
		}
		surface, err := InspectMavenLicense(aligned)
		if err != nil {
			t.Fatalf("AlignMavenLicense produced an unprovable surface: %q -> %q (%v)", content, aligned, err)
		}
		if !surface.LicensesPresent || surface.LicenseCount != 1 || !surface.NamePresent || !surface.URLPresent {
			t.Fatalf("AlignMavenLicense did not align the license surface: %q -> %q (%+v)", content, aligned, surface)
		}
		if surface.Name != name || surface.URL != url {
			t.Fatalf("AlignMavenLicense did not align the license values: %q -> %q (%q, %q)", content, aligned, surface.Name, surface.URL)
		}
		if !changed {
			return
		}
		again, changedAgain, err := AlignMavenLicense(aligned, form)
		if err != nil || changedAgain || again != aligned {
			t.Fatalf("AlignMavenLicense is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
		}
	})
}
