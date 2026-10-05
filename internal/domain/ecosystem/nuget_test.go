package ecosystem

import (
	"errors"
	"strings"
	"testing"
)

// Convention: docs/conventions/cli/testing/README.md

func TestNuGetProjectionDerivesFromTheTenantValues(t *testing.T) {
	cases := []struct {
		name     string
		merged   map[string]string
		wantForm NuGetLicenseForm
	}{
		{"custom family", map[string]string{"LICENSE_ID": "example-NoRepublish-1.0"}, NuGetLicenseForm{Expression: false, Value: "LICENSE"}},
		{"spdx identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "MIT", "LICENSE_ID": "example-NoRepublish-1.0"}, NuGetLicenseForm{Expression: true, Value: "MIT"}},
		{"padded identifier", map[string]string{"SPDX_LICENSE_IDENTIFIER": "  MIT  "}, NuGetLicenseForm{Expression: true, Value: "MIT"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := NuGetProjection(testCase.merged)
			if got.Expression != testCase.wantForm.Expression || got.Value != testCase.wantForm.Value {
				t.Fatalf("NuGetProjection() = %v, want %v", got, testCase.wantForm)
			}
		})
	}
}

func TestNuGetManifestNamesFiltersTheDirectoryListing(t *testing.T) {
	entries := []string{"readme.md", "B.nuspec", "a.csproj", "App.CSPROJ", "LICENSE", "subdir"}
	got := NuGetManifestNames(entries)
	want := []string{"App.CSPROJ", "B.nuspec", "a.csproj"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("NuGetManifestNames() = %v, want %v", got, want)
	}
	if other := NuGetManifestNames([]string{"readme.md", "LICENSE"}); len(other) != 0 {
		t.Fatalf("NuGetManifestNames() = %v, want no candidates", other)
	}
}

const exampleNuspec = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<package xmlns=\"http://schemas.microsoft.com/packaging/2012/06/nuspec.xsd\">\n    <metadata>\n        <id>example</id>\n        <version>1.0.0</version>\n    </metadata>\n</package>\n"

const nuspecWithLicense = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<package xmlns=\"http://schemas.microsoft.com/packaging/2012/06/nuspec.xsd\">\n    <metadata>\n        <id>example</id>\n        <license type=\"expression\">MIT</license>\n    </metadata>\n</package>\n"

const exampleProject = "<Project Sdk=\"Microsoft.NET.Sdk\">\n    <PropertyGroup>\n        <TargetFramework>net9.0</TargetFramework>\n    </PropertyGroup>\n</Project>\n"

const projectWithLicense = "<Project Sdk=\"Microsoft.NET.Sdk\">\n    <PropertyGroup>\n        <PackageLicenseExpression>MIT</PackageLicenseExpression>\n    </PropertyGroup>\n</Project>\n"

func TestInspectNuGetSurfaceClassifiesTheNuspecSurface(t *testing.T) {
	cases := []struct {
		name           string
		content        string
		wantState      NuGetSurfaceState
		wantExpression bool
		wantValue      string
		wantCount      int
		wantURL        bool
	}{
		{"declared expression", nuspecWithLicense, NuGetSurfaceDeclared, true, "MIT", 1, false},
		{"declared file", strings.Replace(nuspecWithLicense, `type="expression">MIT<`, `type="file">LICENSE<`, 1), NuGetSurfaceDeclared, false, "LICENSE", 1, false},
		{"missing element", exampleNuspec, NuGetSurfaceMissing, false, "", 0, false},
		{"self-closing element", strings.Replace(exampleNuspec, "        <id>example</id>\n", "        <license type=\"file\"/>\n        <id>example</id>\n", 1), NuGetSurfaceDeclared, false, "", 1, false},
		{"element without type", strings.Replace(exampleNuspec, "        <id>example</id>\n", "        <license>MIT</license>\n        <id>example</id>\n", 1), NuGetSurfaceInvalid, false, "", 1, false},
		{"entities decode", strings.Replace(nuspecWithLicense, ">MIT<", ">A &amp; B<", 1), NuGetSurfaceDeclared, true, "A & B", 1, false},
		{"cdata decodes", strings.Replace(nuspecWithLicense, ">MIT<", "><![CDATA[MIT]]><", 1), NuGetSurfaceDeclared, true, "MIT", 1, false},
		{"comment context", strings.Replace(exampleNuspec, "    <metadata>\n", "    <metadata>\n        <!-- the license -->\n", 1), NuGetSurfaceMissing, false, "", 0, false},
		{"foreign elements carry no weight", strings.Replace(exampleNuspec, "    </metadata>", "        <dependencies>\n            <group>\n                <dependency id=\"other\"/>\n            </group>\n        </dependencies>\n    </metadata>", 1), NuGetSurfaceMissing, false, "", 0, false},
		{"empty document", "", NuGetSurfaceMissing, false, "", 0, false},
		{"deprecated licenseUrl present", strings.Replace(nuspecWithLicense, "    </metadata>", "        <licenseUrl>https://example.org</licenseUrl>\n    </metadata>", 1), NuGetSurfaceDeclared, true, "MIT", 1, true},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectNuGetSurface("example.nuspec", testCase.content)
			if err != nil {
				t.Fatalf("InspectNuGetSurface() error = %v", err)
			}
			if surface.State != testCase.wantState || surface.Expression != testCase.wantExpression || surface.Value != testCase.wantValue {
				t.Fatalf("InspectNuGetSurface() = %+v, want state %d expression %v value %q", surface, testCase.wantState, testCase.wantExpression, testCase.wantValue)
			}
			if surface.FormCount != testCase.wantCount {
				t.Fatalf("InspectNuGetSurface() formCount = %d, want %d", surface.FormCount, testCase.wantCount)
			}
			if surface.DeprecatedURLPresent != testCase.wantURL {
				t.Fatalf("InspectNuGetSurface() deprecatedURLPresent = %v, want %v", surface.DeprecatedURLPresent, testCase.wantURL)
			}
		})
	}
}

func TestInspectNuGetSurfaceClassifiesTheMSBuildSurface(t *testing.T) {
	cases := []struct {
		name           string
		content        string
		wantState      NuGetSurfaceState
		wantExpression bool
		wantValue      string
		wantCount      int
		wantURL        bool
	}{
		{"declared expression", projectWithLicense, NuGetSurfaceDeclared, true, "MIT", 1, false},
		{"declared file", strings.Replace(projectWithLicense, "PackageLicenseExpression", "PackageLicenseFile", -1), NuGetSurfaceDeclared, false, "MIT", 1, false},
		{"missing property", exampleProject, NuGetSurfaceMissing, false, "", 0, false},
		{"second property group", strings.Replace(projectWithLicense, "</Project>", "    <PropertyGroup>\n        <PackageLicenseExpression>Apache-2.0</PackageLicenseExpression>\n    </PropertyGroup>\n</Project>", 1), NuGetSurfaceDeclared, true, "MIT", 2, false},
		{"empty document", "", NuGetSurfaceMissing, false, "", 0, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			surface, err := InspectNuGetSurface("example.csproj", testCase.content)
			if err != nil {
				t.Fatalf("InspectNuGetSurface() error = %v", err)
			}
			if surface.State != testCase.wantState || surface.Expression != testCase.wantExpression || surface.Value != testCase.wantValue {
				t.Fatalf("InspectNuGetSurface() = %+v, want state %d expression %v value %q", surface, testCase.wantState, testCase.wantExpression, testCase.wantValue)
			}
			if surface.FormCount != testCase.wantCount {
				t.Fatalf("InspectNuGetSurface() formCount = %d, want %d", surface.FormCount, testCase.wantCount)
			}
		})
	}
}

func TestInspectNuGetSurfaceRefusesUnprovableSurfaces(t *testing.T) {
	cases := []struct {
		name         string
		manifestName string
		content      string
		wantErrIs    error
	}{
		{"root is not package", "example.nuspec", "<notpackage>\n    <metadata>\n    </metadata>\n</notpackage>\n", ErrInvalidNuspecSurface},
		{"unclosed metadata", "example.nuspec", "<package>\n    <metadata>\n", ErrInvalidNuspecSurface},
		{"bad entity", "example.nuspec", "<package>\n    <metadata>&unknown;</metadata>\n</package>\n", ErrInvalidNuspecSurface},
		{"duplicate metadata", "example.nuspec", "<package>\n    <metadata>\n    </metadata>\n    <metadata/>\n</package>\n", ErrInvalidNuspecSurface},
		{"duplicate licenseUrl", "example.nuspec", "<package>\n    <metadata>\n        <licenseUrl>https://a</licenseUrl>\n        <licenseUrl>https://b</licenseUrl>\n    </metadata>\n</package>\n", ErrInvalidNuspecSurface},
		{"root is not Project", "example.csproj", "<NotProject/>\n", ErrInvalidMSBuildSurface},
		{"unclosed project", "example.csproj", "<Project>\n", ErrInvalidMSBuildSurface},
		{"duplicate PackageLicenseUrl", "example.csproj", "<Project>\n    <PropertyGroup>\n        <PackageLicenseUrl>https://a</PackageLicenseUrl>\n        <PackageLicenseUrl>https://b</PackageLicenseUrl>\n    </PropertyGroup>\n</Project>\n", ErrInvalidMSBuildSurface},
		{"unattributable name", "other.xml", exampleNuspec, ErrInvalidNuGetName},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := InspectNuGetSurface(testCase.manifestName, testCase.content)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("InspectNuGetSurface(%q) error = %v, want %v", testCase.manifestName, err, testCase.wantErrIs)
			}
		})
	}
}

func TestAlignNuGetSurfaceInsertsTheMissingElements(t *testing.T) {
	cases := []struct {
		name         string
		manifestName string
		content      string
		form         NuGetLicenseForm
		want         string
	}{
		{"nuspec insertion", "example.nuspec", exampleNuspec, NuGetLicenseForm{Expression: false, Value: "LICENSE"},
			"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<package xmlns=\"http://schemas.microsoft.com/packaging/2012/06/nuspec.xsd\">\n    <metadata>\n        <license type=\"file\">LICENSE</license>\n        <id>example</id>\n        <version>1.0.0</version>\n    </metadata>\n</package>\n"},
		{"nuspec self-closing metadata", "example.nuspec", "<package>\n    <metadata/>\n</package>\n", NuGetLicenseForm{Expression: true, Value: "MIT"},
			"<package>\n    <metadata>\n        <license type=\"expression\">MIT</license></metadata>\n</package>\n"},
		{"msbuild insertion", "example.csproj", exampleProject, NuGetLicenseForm{Expression: false, Value: "LICENSE"},
			"<Project Sdk=\"Microsoft.NET.Sdk\">\n    <PropertyGroup>\n        <PackageLicenseFile>LICENSE</PackageLicenseFile>\n        <TargetFramework>net9.0</TargetFramework>\n    </PropertyGroup>\n</Project>\n"},
		{"msbuild without property group", "example.csproj", "<Project Sdk=\"Microsoft.NET.Sdk\">\n</Project>\n", NuGetLicenseForm{Expression: true, Value: "MIT"},
			"<Project Sdk=\"Microsoft.NET.Sdk\">\n    <PropertyGroup>\n        <PackageLicenseExpression>MIT</PackageLicenseExpression>\n    </PropertyGroup>\n</Project>\n"},
		{"msbuild self-closing project", "example.csproj", "<Project Sdk=\"Microsoft.NET.Sdk\"/>", NuGetLicenseForm{Expression: true, Value: "MIT"},
			"<Project Sdk=\"Microsoft.NET.Sdk\">\n    <PropertyGroup>\n        <PackageLicenseExpression>MIT</PackageLicenseExpression>\n    </PropertyGroup></Project>"},
		{"msbuild self-closing property group", "example.csproj", "<Project>\n    <PropertyGroup/>\n</Project>\n", NuGetLicenseForm{Expression: false, Value: "LICENSE"},
			"<Project>\n    <PropertyGroup>\n        <PackageLicenseFile>LICENSE</PackageLicenseFile></PropertyGroup>\n</Project>\n"},
		{"msbuild self-closing property", "example.csproj", "<Project>\n    <PropertyGroup>\n        <PackageLicenseFile/>\n    </PropertyGroup>\n</Project>\n", NuGetLicenseForm{Expression: false, Value: "LICENSE"},
			"<Project>\n    <PropertyGroup>\n        <PackageLicenseFile>LICENSE</PackageLicenseFile>\n    </PropertyGroup>\n</Project>\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignNuGetSurface(testCase.manifestName, testCase.content, testCase.form)
			if err != nil {
				t.Fatalf("AlignNuGetSurface() error = %v", err)
			}
			if !changed {
				t.Fatal("AlignNuGetSurface() changed = false, want true")
			}
			if aligned != testCase.want {
				t.Fatalf("AlignNuGetSurface() = %q, want %q", aligned, testCase.want)
			}
		})
	}
}

func TestAlignNuGetSurfaceSwitchesTheExclusiveForm(t *testing.T) {
	cases := []struct {
		name         string
		manifestName string
		content      string
		form         NuGetLicenseForm
		want         string
	}{
		{"nuspec expression to file", "example.nuspec", nuspecWithLicense, NuGetLicenseForm{Expression: false, Value: "LICENSE"},
			strings.Replace(nuspecWithLicense, `<license type="expression">MIT</license>`, `<license type="file">LICENSE</license>`, 1)},
		{"nuspec file to expression", "example.nuspec", strings.Replace(nuspecWithLicense, `type="expression">MIT<`, `type="file">LICENSE<`, 1), NuGetLicenseForm{Expression: true, Value: "Apache-2.0"},
			strings.Replace(nuspecWithLicense, `<license type="expression">MIT</license>`, `<license type="expression">Apache-2.0</license>`, 1)},
		{"msbuild expression to file", "example.csproj", projectWithLicense, NuGetLicenseForm{Expression: false, Value: "LICENSE"},
			strings.Replace(projectWithLicense, "<PackageLicenseExpression>MIT</PackageLicenseExpression>", "<PackageLicenseFile>LICENSE</PackageLicenseFile>", 1)},
		{"msbuild file to expression", "example.csproj", strings.Replace(projectWithLicense, "PackageLicenseExpression", "PackageLicenseFile", -1), NuGetLicenseForm{Expression: true, Value: "MIT"},
			projectWithLicense},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignNuGetSurface(testCase.manifestName, testCase.content, testCase.form)
			if err != nil {
				t.Fatalf("AlignNuGetSurface() error = %v", err)
			}
			if !changed {
				t.Fatal("AlignNuGetSurface() changed = false, want true")
			}
			if aligned != testCase.want {
				t.Fatalf("AlignNuGetSurface() = %q, want %q", aligned, testCase.want)
			}
		})
	}
}

func TestAlignNuGetSurfaceReplacesADivergingValue(t *testing.T) {
	aligned, changed, err := AlignNuGetSurface("example.nuspec", nuspecWithLicense, NuGetLicenseForm{Expression: true, Value: "Apache-2.0"})
	if err != nil {
		t.Fatalf("AlignNuGetSurface() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignNuGetSurface() changed = false, want true")
	}
	if !strings.Contains(aligned, `<license type="expression">Apache-2.0</license>`) {
		t.Fatalf("AlignNuGetSurface() = %q", aligned)
	}
}

func TestAlignNuGetSurfaceReplacesADivergingMSBuildValue(t *testing.T) {
	aligned, changed, err := AlignNuGetSurface("example.csproj",
		"<Project>\n    <PropertyGroup>\n        <PackageLicenseFile>COPYING</PackageLicenseFile>\n    </PropertyGroup>\n</Project>\n",
		NuGetLicenseForm{Expression: false, Value: "LICENSE"})
	if err != nil {
		t.Fatalf("AlignNuGetSurface() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignNuGetSurface() changed = false, want true")
	}
	if !strings.Contains(aligned, "<PackageLicenseFile>LICENSE</PackageLicenseFile>") || strings.Contains(aligned, "COPYING") {
		t.Fatalf("AlignNuGetSurface() = %q", aligned)
	}
}

func TestAlignNuGetSurfaceIsIdempotentOnAlignedState(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		manifestName string
		content      string
		form         NuGetLicenseForm
	}{
		{"nuspec", "example.nuspec", nuspecWithLicense, NuGetLicenseForm{Expression: true, Value: "MIT"}},
		{"msbuild", "example.csproj", projectWithLicense, NuGetLicenseForm{Expression: true, Value: "MIT"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			aligned, changed, err := AlignNuGetSurface(testCase.manifestName, testCase.content, testCase.form)
			if err != nil {
				t.Fatalf("AlignNuGetSurface() error = %v", err)
			}
			if changed {
				t.Fatal("AlignNuGetSurface() changed = true, want false")
			}
			if aligned != testCase.content {
				t.Fatalf("AlignNuGetSurface() mutated an aligned manifest: %q", aligned)
			}
		})
	}
}

func TestAlignNuGetSurfacePreservesEveryNonLicenseByte(t *testing.T) {
	aligned, _, err := AlignNuGetSurface("example.nuspec", nuspecWithLicense, NuGetLicenseForm{Expression: true, Value: "Apache-2.0"})
	if err != nil {
		t.Fatalf("AlignNuGetSurface() error = %v", err)
	}
	if !strings.Contains(aligned, "<id>example</id>") {
		t.Fatalf("AlignNuGetSurface() touched a foreign element: %q", aligned)
	}
}

func TestAlignNuGetSurfaceEscapesTheProjectedValues(t *testing.T) {
	aligned, changed, err := AlignNuGetSurface("example.nuspec", nuspecWithLicense, NuGetLicenseForm{Expression: true, Value: "A & B <C>"})
	if err != nil {
		t.Fatalf("AlignNuGetSurface() error = %v", err)
	}
	if !changed {
		t.Fatal("AlignNuGetSurface() changed = false, want true")
	}
	surface, err := InspectNuGetSurface("example.nuspec", aligned)
	if err != nil {
		t.Fatalf("InspectNuGetSurface() error = %v", err)
	}
	if surface.Value != "A & B <C>" {
		t.Fatalf("InspectNuGetSurface() value = %q, want the decoded round trip", surface.Value)
	}
}

func TestAlignNuGetSurfaceRefusesInvalidSurfaces(t *testing.T) {
	cases := []struct {
		name         string
		manifestName string
		content      string
		form         NuGetLicenseForm
		wantErrIs    error
	}{
		{"multiple nuspec elements", "example.nuspec", "<package>\n    <metadata>\n        <license type=\"expression\">MIT</license>\n        <license type=\"file\">LICENSE</license>\n    </metadata>\n</package>\n", NuGetLicenseForm{Expression: true, Value: "MIT"}, ErrMultipleNuspecLicenses},
		{"malformed nuspec", "example.nuspec", "<package>\n    <metadata>\n", NuGetLicenseForm{Expression: true, Value: "MIT"}, ErrInvalidNuspecSurface},
		{"nuspec without metadata", "example.nuspec", "<package>\n</package>\n", NuGetLicenseForm{Expression: true, Value: "MIT"}, ErrInvalidNuspecSurface},
		{"invalid license type", "example.nuspec", "<package>\n    <metadata>\n        <license>MIT</license>\n    </metadata>\n</package>\n", NuGetLicenseForm{Expression: true, Value: "MIT"}, ErrInvalidNuspecLicenseType},
		{"empty nuspec", "example.nuspec", "", NuGetLicenseForm{Expression: true, Value: "MIT"}, ErrInvalidNuspecSurface},
		{"multiple msbuild properties", "example.csproj", "<Project>\n    <PropertyGroup>\n        <PackageLicenseExpression>MIT</PackageLicenseExpression>\n        <PackageLicenseFile>LICENSE</PackageLicenseFile>\n    </PropertyGroup>\n</Project>\n", NuGetLicenseForm{Expression: true, Value: "MIT"}, ErrMultipleMSBuildLicenses},
		{"malformed msbuild", "example.csproj", "<Project>\n", NuGetLicenseForm{Expression: true, Value: "MIT"}, ErrInvalidMSBuildSurface},
		{"empty msbuild", "example.csproj", "", NuGetLicenseForm{Expression: true, Value: "MIT"}, ErrInvalidMSBuildSurface},
		{"non-utf-8 value", "example.nuspec", nuspecWithLicense, NuGetLicenseForm{Expression: true, Value: "MIT\xff"}, ErrInvalidTarget},
		{"unrepresentable value", "example.nuspec", nuspecWithLicense, NuGetLicenseForm{Expression: true, Value: "MIT\x00"}, ErrInvalidTarget},
		{"unattributable name", "other.xml", nuspecWithLicense, NuGetLicenseForm{Expression: true, Value: "MIT"}, ErrInvalidNuGetName},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := AlignNuGetSurface(testCase.manifestName, testCase.content, testCase.form)
			if !errors.Is(err, testCase.wantErrIs) {
				t.Fatalf("AlignNuGetSurface(%q) error = %v, want %v", testCase.manifestName, err, testCase.wantErrIs)
			}
		})
	}
}

func FuzzAlignNuGetLicense(f *testing.F) {
	f.Add("example.nuspec", exampleNuspec, "MIT", true)
	f.Add("example.nuspec", nuspecWithLicense, "LICENSE", false)
	f.Add("example.nuspec", "<package>\n    <metadata/>\n</package>\n", "MIT", true)
	f.Add("example.nuspec", "<package>\n    <metadata>\n        <license type=\"file\"/>\n    </metadata>\n</package>\n", "LICENSE", false)
	f.Add("example.nuspec", "", "MIT", true)
	f.Add("example.csproj", exampleProject, "MIT", true)
	f.Add("example.csproj", projectWithLicense, "LICENSE", false)
	f.Add("example.csproj", "<Project Sdk=\"Microsoft.NET.Sdk\"/>", "LICENSE", false)
	f.Add("example.csproj", "<Project>\n    <PropertyGroup/>\n</Project>\n", "MIT", true)
	f.Add("other.xml", nuspecWithLicense, "MIT", true)
	f.Fuzz(func(t *testing.T, name, content, value string, expression bool) {
		form := NuGetLicenseForm{Expression: expression, Value: value}
		aligned, changed, err := AlignNuGetSurface(name, content, form)
		if err != nil {
			return
		}
		surface, err := InspectNuGetSurface(name, aligned)
		if err != nil {
			t.Fatalf("AlignNuGetSurface produced an unprovable surface: %q -> %q (%v)", content, aligned, err)
		}
		if surface.State != NuGetSurfaceDeclared || surface.Expression != form.Expression || surface.Value != form.Value {
			t.Fatalf("AlignNuGetSurface did not align the license surface: %q -> %q (%+v)", content, aligned, surface)
		}
		if !changed {
			return
		}
		again, changedAgain, err := AlignNuGetSurface(name, aligned, form)
		if err != nil || changedAgain || again != aligned {
			t.Fatalf("AlignNuGetSurface is not idempotent: %q -> %q (%v, %v)", aligned, again, changedAgain, err)
		}
	})
}
