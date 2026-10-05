package application

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/t33n-software/license-hub/internal/domain/digest"
)

// Convention: docs/conventions/cli/testing/README.md

type fakeFS struct {
	files    map[string][]byte
	readErr  map[string]error
	writeErr map[string]error
}

func newFakeFS() *fakeFS {
	return &fakeFS{
		files:    make(map[string][]byte),
		readErr:  make(map[string]error),
		writeErr: make(map[string]error),
	}
}

func (f *fakeFS) ReadFile(path string) ([]byte, error) {
	if err, blocked := f.readErr[path]; blocked {
		return nil, err
	}
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("not found: %s: %w", path, os.ErrNotExist)
	}
	return data, nil
}

func (f *fakeFS) WriteFile(path string, data []byte) error {
	if err, blocked := f.writeErr[path]; blocked {
		return err
	}
	f.files[path] = data
	return nil
}

const testTemplate = "{{PROJECT_NAME}} (c) {{COPYRIGHT_YEAR}} {{COPYRIGHT_HOLDER}}\n" +
	"source: {{CANONICAL_SOURCE_URL}} contact: {{PERMISSION_CONTACT}}\n" +
	"law: {{GOVERNING_LAW}} venue: {{VENUE}} id: {{LICENSE_ID}}\n"

var (
	licensePath   = filepath.Join("out", "LICENSE")
	canonicalPath = filepath.Join("out", "LICENSES", "LicenseRef-license-hub-NoRepublish-1.0.txt")
)

func valuesJSON(t *testing.T, pairs map[string]string) []byte {
	t.Helper()
	parts := make([]string, 0, len(pairs))
	for key, value := range pairs {
		parts = append(parts, fmt.Sprintf("%q:%q", key, value))
	}
	return []byte("{" + strings.Join(parts, ",") + "}")
}

// seededFS returns a fake filesystem with the canonical template, complete
// organization defaults, and complete tenant values.
func seededFS(t *testing.T) *fakeFS {
	t.Helper()
	f := newFakeFS()
	f.files["template.hbs"] = []byte(testTemplate)
	f.files["org.json"] = valuesJSON(t, map[string]string{
		"COPYRIGHT_HOLDER":   "CyberT33N",
		"GOVERNING_LAW":      "the Federal Republic of Germany",
		"VENUE":              "Germany",
		"PERMISSION_CONTACT": "https://github.com/t33n-software",
	})
	f.files["values.json"] = valuesJSON(t, map[string]string{
		"PROJECT_NAME":         "license-hub",
		"LICENSE_ID":           "license-hub-NoRepublish-1.0",
		"COPYRIGHT_YEAR":       "2026",
		"CANONICAL_SOURCE_URL": "https://github.com/t33n-software/license-hub",
	})
	return f
}

func renderRequest() RenderRequest {
	return RenderRequest{
		TemplatePath:    "template.hbs",
		OrgDefaultsPath: "org.json",
		ValuesPath:      "values.json",
		OutDir:          "out",
	}
}

func TestRenderSuccess(t *testing.T) {
	f := seededFS(t)
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Written) != 2 {
		t.Fatalf("Render() wrote %v", result.Written)
	}
	if result.Digest != digest.SHA256([]byte(testTemplate)) {
		t.Fatalf("Render() digest = %q", result.Digest)
	}
	content := string(f.files[licensePath])
	if strings.Contains(content, "{{") {
		t.Fatalf("Render() left placeholders: %q", content)
	}
	if !strings.Contains(content, "license-hub (c) 2026 CyberT33N") {
		t.Fatalf("Render() content = %q", content)
	}
	if string(f.files[canonicalPath]) != content {
		t.Fatalf("Render() canonical file mismatch")
	}
}

func TestRenderTemplateReadError(t *testing.T) {
	service := NewLicenseService(newFakeFS())
	if _, err := service.Render(renderRequest()); err == nil {
		t.Fatal("Render() expected template read error")
	}
}

func TestRenderOrgDefaultsReadError(t *testing.T) {
	f := seededFS(t)
	delete(f.files, "org.json")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err == nil {
		t.Fatal("Render() expected org defaults read error")
	}
}

func TestRenderValuesReadError(t *testing.T) {
	f := seededFS(t)
	delete(f.files, "values.json")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err == nil {
		t.Fatal("Render() expected values read error")
	}
}

func TestRenderOrgDefaultsParseError(t *testing.T) {
	f := seededFS(t)
	f.files["org.json"] = []byte("{")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err == nil {
		t.Fatal("Render() expected org defaults parse error")
	}
}

func TestRenderValuesParseError(t *testing.T) {
	f := seededFS(t)
	f.files["values.json"] = []byte("{")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err == nil {
		t.Fatal("Render() expected values parse error")
	}
}

func TestRenderMissingRequiredKeys(t *testing.T) {
	f := seededFS(t)
	f.files["values.json"] = valuesJSON(t, map[string]string{"PROJECT_NAME": "x"})
	service := NewLicenseService(f)
	_, err := service.Render(renderRequest())
	if err == nil || !strings.Contains(err.Error(), "missing required values") {
		t.Fatalf("Render() error = %v", err)
	}
}

func TestRenderUnresolvedPlaceholders(t *testing.T) {
	f := seededFS(t)
	f.files["template.hbs"] = []byte(testTemplate + "{{UNKNOWN_ANCHOR}}\n")
	service := NewLicenseService(f)
	_, err := service.Render(renderRequest())
	if err == nil || !strings.Contains(err.Error(), "{{UNKNOWN_ANCHOR}}") {
		t.Fatalf("Render() error = %v", err)
	}
}

func TestRenderWriteError(t *testing.T) {
	f := seededFS(t)
	f.writeErr[licensePath] = fmt.Errorf("disk full")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err == nil {
		t.Fatal("Render() expected write error")
	}
}

func verifyRequest(lockPath string) VerifyRequest {
	return VerifyRequest{
		TemplatePath:    "template.hbs",
		OrgDefaultsPath: "org.json",
		ValuesPath:      "values.json",
		LockPath:        lockPath,
		Dir:             "out",
	}
}

func TestVerifySuccessAfterRender(t *testing.T) {
	f := seededFS(t)
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestVerifyTemplateReadError(t *testing.T) {
	service := NewLicenseService(newFakeFS())
	if _, err := service.Verify(verifyRequest("")); err == nil {
		t.Fatal("Verify() expected template read error")
	}
}

func TestVerifyValuesError(t *testing.T) {
	f := seededFS(t)
	delete(f.files, "values.json")
	service := NewLicenseService(f)
	if _, err := service.Verify(verifyRequest("")); err == nil {
		t.Fatal("Verify() expected values error")
	}
}

func TestVerifyLockReadError(t *testing.T) {
	f := seededFS(t)
	service := NewLicenseService(f)
	if _, err := service.Verify(verifyRequest("missing.lock.json")); err == nil {
		t.Fatal("Verify() expected lock read error")
	}
}

func TestVerifyLockParseError(t *testing.T) {
	f := seededFS(t)
	f.files["lock.json"] = []byte("{")
	service := NewLicenseService(f)
	if _, err := service.Verify(verifyRequest("lock.json")); err == nil {
		t.Fatal("Verify() expected lock parse error")
	}
}

func TestVerifyLockDigestMatch(t *testing.T) {
	f := seededFS(t)
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	f.files["lock.json"] = []byte(fmt.Sprintf(
		`{"template":"template.hbs","version":"1.0.0","digest":%q}`,
		digest.SHA256([]byte(testTemplate)),
	))
	violations, err := service.Verify(verifyRequest("lock.json"))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestVerifyLockDigestMismatch(t *testing.T) {
	f := seededFS(t)
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	f.files["lock.json"] = []byte(`{"template":"template.hbs","version":"1.0.0","digest":"sha256:00"}`)
	violations, err := service.Verify(verifyRequest("lock.json"))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "digest") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestVerifyUnresolvedPlaceholders(t *testing.T) {
	f := seededFS(t)
	f.files["template.hbs"] = []byte(testTemplate + "{{UNKNOWN_ANCHOR}}\n")
	service := NewLicenseService(f)
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	found := false
	for _, violation := range violations {
		if strings.Contains(violation, "{{UNKNOWN_ANCHOR}}") {
			found = true
		}
	}
	if !found {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestVerifyMissingCommittedFile(t *testing.T) {
	f := seededFS(t)
	service := NewLicenseService(f)
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 2 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestVerifyDriftedCommittedFile(t *testing.T) {
	f := seededFS(t)
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	f.files[licensePath] = []byte("hand edited")
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "drifted") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestTemplateDigest(t *testing.T) {
	f := seededFS(t)
	service := NewLicenseService(f)
	got, err := service.TemplateDigest("template.hbs")
	if err != nil {
		t.Fatalf("TemplateDigest() error = %v", err)
	}
	if got != digest.SHA256([]byte(testTemplate)) {
		t.Fatalf("TemplateDigest() = %q", got)
	}
}

func TestTemplateDigestReadError(t *testing.T) {
	service := NewLicenseService(newFakeFS())
	if _, err := service.TemplateDigest("missing.hbs"); err == nil {
		t.Fatal("TemplateDigest() expected error")
	}
}

func TestPlanRenderReportsTargetsAndDigestWithoutWriting(t *testing.T) {
	f := seededFS(t)
	service := NewLicenseService(f)
	plan, err := service.PlanRender(renderRequest())
	if err != nil {
		t.Fatalf("PlanRender() error = %v", err)
	}
	if len(plan.Targets) != 2 || plan.Targets[0] != licensePath || plan.Targets[1] != canonicalPath {
		t.Fatalf("PlanRender() targets = %v", plan.Targets)
	}
	if plan.Digest != digest.SHA256([]byte(testTemplate)) {
		t.Fatalf("PlanRender() digest = %q", plan.Digest)
	}
	if len(f.files) != 3 {
		t.Fatalf("PlanRender() must not write; files = %v", len(f.files))
	}
}

func TestPlanRenderTemplateReadError(t *testing.T) {
	service := NewLicenseService(newFakeFS())
	if _, err := service.PlanRender(renderRequest()); err == nil {
		t.Fatal("PlanRender() expected template read error")
	}
}

func TestPlanRenderValuesError(t *testing.T) {
	f := seededFS(t)
	delete(f.files, "values.json")
	service := NewLicenseService(f)
	if _, err := service.PlanRender(renderRequest()); err == nil {
		t.Fatal("PlanRender() expected values error")
	}
}

func TestPlanRenderUnresolvedPlaceholders(t *testing.T) {
	f := seededFS(t)
	f.files["template.hbs"] = []byte(testTemplate + "{{UNKNOWN_ANCHOR}}\n")
	service := NewLicenseService(f)
	if _, err := service.PlanRender(renderRequest()); err == nil {
		t.Fatal("PlanRender() expected unresolved placeholders error")
	}
}

func TestMissingValuesErrorCarriesTheSentinel(t *testing.T) {
	f := seededFS(t)
	f.files["values.json"] = valuesJSON(t, map[string]string{"PROJECT_NAME": "x"})
	service := NewLicenseService(f)
	_, err := service.Render(renderRequest())
	if !errors.Is(err, ErrMissingValues) {
		t.Fatalf("Render() error = %v, want ErrMissingValues", err)
	}
}

func TestUnresolvedPlaceholdersErrorCarriesTheSentinel(t *testing.T) {
	f := seededFS(t)
	f.files["template.hbs"] = []byte(testTemplate + "{{UNKNOWN_ANCHOR}}\n")
	service := NewLicenseService(f)
	_, err := service.Render(renderRequest())
	if !errors.Is(err, ErrUnresolvedPlaceholders) {
		t.Fatalf("Render() error = %v, want ErrUnresolvedPlaceholders", err)
	}
}

func TestInstancePathsLegacyStemWithoutSpdxIdentifier(t *testing.T) {
	got := instancePaths("out", map[string]string{"LICENSE_ID": "example-NoRepublish-1.0"})
	want := []string{
		filepath.Join("out", "LICENSE"),
		filepath.Join("out", "LICENSES", "LicenseRef-example-NoRepublish-1.0.txt"),
	}
	if !slices.Equal(got, want) {
		t.Fatalf("instancePaths() = %v, want %v", got, want)
	}
}

func TestInstancePathsSpdxIdentifierWins(t *testing.T) {
	got := instancePaths("out", map[string]string{
		"LICENSE_ID":              "example-MIT",
		"SPDX_LICENSE_IDENTIFIER": "MIT",
	})
	want := []string{
		filepath.Join("out", "LICENSE"),
		filepath.Join("out", "LICENSES", "MIT.txt"),
	}
	if !slices.Equal(got, want) {
		t.Fatalf("instancePaths() = %v, want %v", got, want)
	}
}

func TestInstancePathsBlankSpdxIdentifierFallsBack(t *testing.T) {
	got := instancePaths("out", map[string]string{
		"LICENSE_ID":              "example-MIT",
		"SPDX_LICENSE_IDENTIFIER": "   ",
	})
	want := filepath.Join("out", "LICENSES", "LicenseRef-example-MIT.txt")
	if got[1] != want {
		t.Fatalf("instancePaths() = %v, want second path %v", got, want)
	}
}

func TestRenderAndVerifyWithSpdxIdentifier(t *testing.T) {
	f := seededFS(t)
	f.files["values.json"] = valuesJSON(t, map[string]string{
		"PROJECT_NAME":            "example-project",
		"LICENSE_ID":              "example-project-MIT",
		"COPYRIGHT_YEAR":          "2026",
		"CANONICAL_SOURCE_URL":    "https://github.com/t33n-software/example-project",
		"SPDX_LICENSE_IDENTIFIER": "MIT",
	})
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	spdxPath := filepath.Join("out", "LICENSES", "MIT.txt")
	if len(result.Written) != 2 || result.Written[1] != spdxPath {
		t.Fatalf("Render() wrote %v, want second path %s", result.Written, spdxPath)
	}
	legacyPath := filepath.Join("out", "LICENSES", "LicenseRef-example-project-MIT.txt")
	if _, ok := f.files[legacyPath]; ok {
		t.Fatalf("Render() wrote %s despite SPDX_LICENSE_IDENTIFIER", legacyPath)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

// --- Ecosystem license surfaces ---

const nodeSeam = `{"schemaVersion":4,"toolchain":{"language":"node-typescript","version":"26.10.0"}}`
const goSeam = `{"schemaVersion":4,"toolchain":{"language":"go","version":"1.26.6"}}`

var seamPath = filepath.Join("out", "git-governance.quality.json")
var manifestPath = filepath.Join("out", "package.json")

func seedNpmSurface(f *fakeFS, seam, manifest string) {
	if seam != "" {
		f.files[seamPath] = []byte(seam)
	}
	if manifest != "" {
		f.files[manifestPath] = []byte(manifest)
	}
}

func TestRenderAlignsTheDeclaredNpmSurface(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, nodeSeam, "{\n  \"name\": \"example-project\",\n  \"license\": \"MIT\"\n}\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !slices.Equal(result.Aligned, []string{manifestPath}) {
		t.Fatalf("Render() aligned = %v", result.Aligned)
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("Render() skipped = %v", result.Skipped)
	}
	content := string(f.files[manifestPath])
	if !strings.Contains(content, `"license": "SEE LICENSE IN LICENSE"`) {
		t.Fatalf("Render() manifest = %q", content)
	}
	if !strings.Contains(content, "\"name\": \"example-project\"") {
		t.Fatalf("Render() touched a non-license field: %q", content)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderNpmAlignmentIsIdempotent(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, nodeSeam, "{\n  \"name\": \"example-project\",\n  \"license\": \"SEE LICENSE IN LICENSE\"\n}\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
}

func TestPlanRenderPreviewsTheNpmAlignment(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, nodeSeam, "{\n  \"license\": \"MIT\"\n}\n")
	service := NewLicenseService(f)
	plan, err := service.PlanRender(renderRequest())
	if err != nil {
		t.Fatalf("PlanRender() error = %v", err)
	}
	if len(plan.Alignments) != 1 || !strings.Contains(plan.Alignments[0], manifestPath) {
		t.Fatalf("PlanRender() alignments = %v", plan.Alignments)
	}
	if string(f.files[manifestPath]) != "{\n  \"license\": \"MIT\"\n}\n" {
		t.Fatalf("PlanRender() mutated the manifest: %q", f.files[manifestPath])
	}
}

func TestRenderSkipsTheAbsentDeclaredManifest(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, nodeSeam, "")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 {
		t.Fatalf("Render() aligned = %v", result.Aligned)
	}
	if len(result.Skipped) != 1 || !strings.Contains(result.Skipped[0], "absent") {
		t.Fatalf("Render() skipped = %v", result.Skipped)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "manifest missing") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderWithoutDeclarationLeavesNpmSurfacesUntouched(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, "", "{\n  \"license\": \"MIT\"\n}\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
}

func TestRenderWithNonNpmDeclarationLeavesNpmSurfacesUntouched(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, goSeam, "{\n  \"license\": \"MIT\"\n}\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "without covering declaration") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderRejectsAnInvalidNpmSurface(t *testing.T) {
	cases := []struct {
		name     string
		seam     string
		manifest string
	}{
		{"invalid seam", "{", ""},
		{"invalid manifest", nodeSeam, `{"license":`},
		{"non-string field", nodeSeam, `{"license":7}`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedNpmSurface(f, testCase.seam, testCase.manifest)
			service := NewLicenseService(f)
			_, err := service.Render(renderRequest())
			if !errors.Is(err, ErrInvalidEcosystemSurface) {
				t.Fatalf("Render() error = %v, want ErrInvalidEcosystemSurface", err)
			}
		})
	}
}

func TestRenderReportsEcosystemReadFailures(t *testing.T) {
	cases := []struct {
		name    string
		blocked string
		wantIs  error
		wantErr string
	}{
		{"seam read failure", seamPath, os.ErrPermission, "read ecosystem declaration"},
		{"manifest read failure", manifestPath, os.ErrPermission, "read npm manifest"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedNpmSurface(f, nodeSeam, "{\n  \"license\": \"MIT\"\n}\n")
			f.readErr[testCase.blocked] = testCase.wantIs
			service := NewLicenseService(f)
			_, err := service.Render(renderRequest())
			if !errors.Is(err, testCase.wantIs) || !strings.Contains(err.Error(), testCase.wantErr) {
				t.Fatalf("Render() error = %v, want %v with %q", err, testCase.wantIs, testCase.wantErr)
			}
		})
	}
}

func TestRenderNpmAlignWriteError(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, nodeSeam, "{\n  \"license\": \"MIT\"\n}\n")
	f.writeErr[manifestPath] = fmt.Errorf("disk full")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err == nil {
		t.Fatal("Render() expected npm manifest write error")
	}
}

func TestPlanRenderSkipsTheAbsentDeclaredManifest(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, nodeSeam, "")
	service := NewLicenseService(f)
	plan, err := service.PlanRender(renderRequest())
	if err != nil {
		t.Fatalf("PlanRender() error = %v", err)
	}
	if len(plan.Skips) != 1 || !strings.Contains(plan.Skips[0], "absent") {
		t.Fatalf("PlanRender() skips = %v", plan.Skips)
	}
	if len(plan.Alignments) != 0 {
		t.Fatalf("PlanRender() alignments = %v", plan.Alignments)
	}
}

func TestVerifyReportsManifestReadFailure(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, nodeSeam, "")
	f.readErr[manifestPath] = os.ErrPermission
	service := NewLicenseService(f)
	if _, err := service.Verify(verifyRequest("")); !errors.Is(err, os.ErrPermission) {
		t.Fatalf("Verify() error = %v, want os.ErrPermission", err)
	}
}

func TestVerifyNpmFindingsFailClosed(t *testing.T) {
	cases := []struct {
		name     string
		seam     string
		manifest string
		want     string
	}{
		{"diverging field", nodeSeam, "{\n  \"license\": \"MIT\"\n}\n", "diverges from the lock projection"},
		{"missing field", nodeSeam, "{\n  \"name\": \"x\"\n}\n", "license field is missing"},
		{"non-string field", nodeSeam, "{\n  \"license\": 7\n}\n", "not a JSON string"},
		{"unscannable manifest", nodeSeam, `{"license":`, "cannot be proven"},
		{"shadow without declaration", "", "{\n  \"license\": \"MIT\"\n}\n", "without covering declaration"},
		{"shadow with other language", goSeam, "{\n  \"license\": \"MIT\"\n}\n", "without covering declaration"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedNpmSurface(f, testCase.seam, testCase.manifest)
			service := NewLicenseService(f)
			violations, err := service.Verify(verifyRequest(""))
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			found := false
			for _, violation := range violations {
				if strings.Contains(violation, testCase.want) {
					found = true
				}
			}
			if !found {
				t.Fatalf("Verify() violations = %v, want %q", violations, testCase.want)
			}
		})
	}
}

func TestVerifyNpmShadowDetailNamesTheDeclarationState(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, "", "{\n  \"license\": \"MIT\"\n}\n")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "is absent") {
		t.Fatalf("Verify() violations = %v", violations)
	}
	f2 := seededFS(t)
	seedNpmSurface(f2, goSeam, "{\n  \"license\": \"MIT\"\n}\n")
	service = NewLicenseService(f2)
	if _, err := service.Render(renderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	violations, err = service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], `declared language: "go"`) {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestVerifyRejectsAnUnscannableDeclaration(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, "{", "")
	service := NewLicenseService(f)
	if _, err := service.Verify(verifyRequest("")); !errors.Is(err, ErrInvalidEcosystemSurface) {
		t.Fatalf("Verify() error = %v, want ErrInvalidEcosystemSurface", err)
	}
}

func TestVerifyReportsDeclarationReadFailure(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, nodeSeam, "")
	f.readErr[seamPath] = os.ErrPermission
	service := NewLicenseService(f)
	if _, err := service.Verify(verifyRequest("")); !errors.Is(err, os.ErrPermission) {
		t.Fatalf("Verify() error = %v, want os.ErrPermission", err)
	}
}

func TestVerifyNpmCleanWithoutNpmSurfaces(t *testing.T) {
	f := seededFS(t)
	seedNpmSurface(f, goSeam, "")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

// --- Python ecosystem surfaces ---

const pythonSeam = `{"schemaVersion":4,"toolchain":{"language":"python","version":"3.14.0"}}`

var pythonManifestPath = filepath.Join("out", "pyproject.toml")

func seedPythonSurface(f *fakeFS, seam, manifest string) {
	if seam != "" {
		f.files[seamPath] = []byte(seam)
	}
	if manifest != "" {
		f.files[pythonManifestPath] = []byte(manifest)
	}
}

func TestRenderAlignsTheDeclaredPythonSurface(t *testing.T) {
	f := seededFS(t)
	seedPythonSurface(f, pythonSeam, "[project]\nname = \"example-project\"\nlicense = \"MIT\"\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !slices.Equal(result.Aligned, []string{pythonManifestPath}) {
		t.Fatalf("Render() aligned = %v", result.Aligned)
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("Render() skipped = %v", result.Skipped)
	}
	content := string(f.files[pythonManifestPath])
	want := "[project]\nname = \"example-project\"\nlicense = \"LicenseRef-license-hub-NoRepublish-1.0\"\nlicense-files = [\"LICENSE\"]\n"
	if content != want {
		t.Fatalf("Render() manifest = %q, want %q", content, want)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderPythonAlignmentIsIdempotent(t *testing.T) {
	f := seededFS(t)
	seedPythonSurface(f, pythonSeam, "[project]\nlicense = \"LicenseRef-license-hub-NoRepublish-1.0\"\nlicense-files = [\"LICENSE\"]\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
}

func TestPlanRenderPreviewsThePythonAlignment(t *testing.T) {
	f := seededFS(t)
	seedPythonSurface(f, pythonSeam, "[project]\nlicense = \"MIT\"\n")
	service := NewLicenseService(f)
	plan, err := service.PlanRender(renderRequest())
	if err != nil {
		t.Fatalf("PlanRender() error = %v", err)
	}
	if len(plan.Alignments) != 1 || !strings.Contains(plan.Alignments[0], pythonManifestPath) {
		t.Fatalf("PlanRender() alignments = %v", plan.Alignments)
	}
	if string(f.files[pythonManifestPath]) != "[project]\nlicense = \"MIT\"\n" {
		t.Fatalf("PlanRender() mutated the manifest: %q", f.files[pythonManifestPath])
	}
}

func TestRenderSkipsTheAbsentDeclaredPythonManifest(t *testing.T) {
	f := seededFS(t)
	seedPythonSurface(f, pythonSeam, "")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 {
		t.Fatalf("Render() aligned = %v", result.Aligned)
	}
	if len(result.Skipped) != 1 || !strings.Contains(result.Skipped[0], "absent") {
		t.Fatalf("Render() skipped = %v", result.Skipped)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "manifest missing") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderWithoutDeclarationLeavesPythonSurfacesUntouched(t *testing.T) {
	f := seededFS(t)
	seedPythonSurface(f, "", "[project]\nlicense = \"MIT\"\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
}

func TestRenderWithNonPythonDeclarationLeavesPythonSurfacesUntouched(t *testing.T) {
	f := seededFS(t)
	seedPythonSurface(f, goSeam, "[project]\nlicense = \"MIT\"\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "without covering declaration") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderRejectsAnInvalidPythonSurface(t *testing.T) {
	cases := []struct {
		name     string
		seam     string
		manifest string
	}{
		{"invalid seam", "{", ""},
		{"invalid manifest", pythonSeam, "[project"},
		{"no project table", pythonSeam, "tool = \"x\"\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedPythonSurface(f, testCase.seam, testCase.manifest)
			service := NewLicenseService(f)
			_, err := service.Render(renderRequest())
			if !errors.Is(err, ErrInvalidEcosystemSurface) {
				t.Fatalf("Render() error = %v, want ErrInvalidEcosystemSurface", err)
			}
		})
	}
}

func TestRenderReportsPythonReadFailures(t *testing.T) {
	cases := []struct {
		name    string
		blocked string
		wantIs  error
		wantErr string
	}{
		{"seam read failure", seamPath, os.ErrPermission, "read ecosystem declaration"},
		{"manifest read failure", pythonManifestPath, os.ErrPermission, "read python manifest"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedPythonSurface(f, pythonSeam, "[project]\nlicense = \"MIT\"\n")
			f.readErr[testCase.blocked] = testCase.wantIs
			service := NewLicenseService(f)
			_, err := service.Render(renderRequest())
			if !errors.Is(err, testCase.wantIs) || !strings.Contains(err.Error(), testCase.wantErr) {
				t.Fatalf("Render() error = %v, want %v with %q", err, testCase.wantIs, testCase.wantErr)
			}
		})
	}
}

func TestRenderPythonAlignWriteError(t *testing.T) {
	f := seededFS(t)
	seedPythonSurface(f, pythonSeam, "[project]\nlicense = \"MIT\"\n")
	f.writeErr[pythonManifestPath] = fmt.Errorf("disk full")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err == nil {
		t.Fatal("Render() expected python manifest write error")
	}
}

func TestPlanRenderSkipsTheAbsentDeclaredPythonManifest(t *testing.T) {
	f := seededFS(t)
	seedPythonSurface(f, pythonSeam, "")
	service := NewLicenseService(f)
	plan, err := service.PlanRender(renderRequest())
	if err != nil {
		t.Fatalf("PlanRender() error = %v", err)
	}
	if len(plan.Skips) != 1 || !strings.Contains(plan.Skips[0], "absent") {
		t.Fatalf("PlanRender() skips = %v", plan.Skips)
	}
	if len(plan.Alignments) != 0 {
		t.Fatalf("PlanRender() alignments = %v", plan.Alignments)
	}
}

func TestVerifyPythonFindingsFailClosed(t *testing.T) {
	cases := []struct {
		name     string
		seam     string
		manifest string
		want     string
	}{
		{"diverging field", pythonSeam, "[project]\nlicense = \"MIT\"\n", "diverges from the lock projection"},
		{"missing field", pythonSeam, "[project]\nname = \"x\"\n", "license field is missing"},
		{"deprecated table form", pythonSeam, "[project]\nlicense = { text = \"MIT\" }\n", "deprecated license table form"},
		{"invalid field", pythonSeam, "[project]\nlicense = 7\n", "not a license expression string"},
		{"files missing", pythonSeam, "[project]\nlicense = \"LicenseRef-license-hub-NoRepublish-1.0\"\n", "license-files field is missing"},
		{"files diverging", pythonSeam, "[project]\nlicense = \"LicenseRef-license-hub-NoRepublish-1.0\"\nlicense-files = [\"A\"]\n", "license-files field diverges"},
		{"files invalid form", pythonSeam, "[project]\nlicense = \"LicenseRef-license-hub-NoRepublish-1.0\"\nlicense-files = \"LICENSE\"\n", "not an array of strings"},
		{"deprecated classifiers", pythonSeam, "[project]\nlicense = \"LicenseRef-license-hub-NoRepublish-1.0\"\nlicense-files = [\"LICENSE\"]\nclassifiers = [\"License :: OSI Approved :: MIT License\"]\n", "deprecated License :: classifiers"},
		{"unscannable manifest", pythonSeam, "[project", "cannot be proven"},
		{"shadow without declaration", "", "[project]\nlicense = \"MIT\"\n", "without covering declaration"},
		{"shadow with other language", goSeam, "[project]\nlicense = \"MIT\"\n", "without covering declaration"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedPythonSurface(f, testCase.seam, testCase.manifest)
			service := NewLicenseService(f)
			violations, err := service.Verify(verifyRequest(""))
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			found := false
			for _, violation := range violations {
				if strings.Contains(violation, testCase.want) {
					found = true
				}
			}
			if !found {
				t.Fatalf("Verify() violations = %v, want %q", violations, testCase.want)
			}
		})
	}
}

func TestVerifyPythonReportsDeprecatedClassifiersAfterAlign(t *testing.T) {
	f := seededFS(t)
	seedPythonSurface(f, pythonSeam, "[project]\nlicense = \"MIT\"\nclassifiers = [\"License :: OSI Approved :: MIT License\"]\n")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "deprecated License :: classifiers") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestVerifyPythonCleanWithoutPythonSurfaces(t *testing.T) {
	f := seededFS(t)
	seedPythonSurface(f, goSeam, "")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

// --- Rust ecosystem surfaces ---

const rustSeam = `{"schemaVersion":4,"toolchain":{"language":"rust","version":"1.90.0"}}`

var cargoManifestPath = filepath.Join("out", "Cargo.toml")

func seedCargoSurface(f *fakeFS, seam, manifest string) {
	if seam != "" {
		f.files[seamPath] = []byte(seam)
	}
	if manifest != "" {
		f.files[cargoManifestPath] = []byte(manifest)
	}
}

func TestRenderAlignsTheDeclaredCargoSurface(t *testing.T) {
	f := seededFS(t)
	seedCargoSurface(f, rustSeam, "[package]\nname = \"example-project\"\nlicense = \"MIT\"\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !slices.Equal(result.Aligned, []string{cargoManifestPath}) {
		t.Fatalf("Render() aligned = %v", result.Aligned)
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("Render() skipped = %v", result.Skipped)
	}
	content := string(f.files[cargoManifestPath])
	want := "[package]\nname = \"example-project\"\nlicense-file = \"LICENSE\"\n"
	if content != want {
		t.Fatalf("Render() manifest = %q, want %q", content, want)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderAlignsTheDeclaredCargoSpdxSurface(t *testing.T) {
	f := seededFS(t)
	f.files["values.json"] = valuesJSON(t, map[string]string{
		"PROJECT_NAME":            "example-project",
		"LICENSE_ID":              "example-project-NoRepublish-1.0",
		"COPYRIGHT_YEAR":          "2026",
		"CANONICAL_SOURCE_URL":    "https://github.com/t33n-software/example-project",
		"SPDX_LICENSE_IDENTIFIER": "MIT",
	})
	seedCargoSurface(f, rustSeam, "[package]\nname = \"example-project\"\nlicense-file = \"LICENSE\"\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !slices.Equal(result.Aligned, []string{cargoManifestPath}) {
		t.Fatalf("Render() aligned = %v", result.Aligned)
	}
	content := string(f.files[cargoManifestPath])
	want := "[package]\nname = \"example-project\"\nlicense = \"MIT\"\n"
	if content != want {
		t.Fatalf("Render() manifest = %q, want %q", content, want)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderCargoAlignmentIsIdempotent(t *testing.T) {
	f := seededFS(t)
	seedCargoSurface(f, rustSeam, "[package]\nlicense-file = \"LICENSE\"\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
}

func TestPlanRenderPreviewsTheCargoAlignment(t *testing.T) {
	f := seededFS(t)
	seedCargoSurface(f, rustSeam, "[package]\nlicense = \"MIT\"\n")
	service := NewLicenseService(f)
	plan, err := service.PlanRender(renderRequest())
	if err != nil {
		t.Fatalf("PlanRender() error = %v", err)
	}
	if len(plan.Alignments) != 1 || !strings.Contains(plan.Alignments[0], cargoManifestPath) {
		t.Fatalf("PlanRender() alignments = %v", plan.Alignments)
	}
	if string(f.files[cargoManifestPath]) != "[package]\nlicense = \"MIT\"\n" {
		t.Fatalf("PlanRender() mutated the manifest: %q", f.files[cargoManifestPath])
	}
}

func TestRenderSkipsTheAbsentDeclaredCargoManifest(t *testing.T) {
	f := seededFS(t)
	seedCargoSurface(f, rustSeam, "")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 {
		t.Fatalf("Render() aligned = %v", result.Aligned)
	}
	if len(result.Skipped) != 1 || !strings.Contains(result.Skipped[0], "absent") {
		t.Fatalf("Render() skipped = %v", result.Skipped)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "manifest missing") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderWithoutDeclarationLeavesCargoSurfacesUntouched(t *testing.T) {
	f := seededFS(t)
	seedCargoSurface(f, "", "[package]\nlicense = \"MIT\"\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
}

func TestRenderWithNonCargoDeclarationLeavesCargoSurfacesUntouched(t *testing.T) {
	f := seededFS(t)
	seedCargoSurface(f, goSeam, "[package]\nlicense = \"MIT\"\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "without covering declaration") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderRejectsAnInvalidCargoSurface(t *testing.T) {
	cases := []struct {
		name     string
		seam     string
		manifest string
	}{
		{"invalid seam", "{", ""},
		{"invalid manifest", rustSeam, "[package"},
		{"no package table", rustSeam, "tool = \"x\"\n"},
		{"exclusive keys", rustSeam, "[package]\nlicense = \"MIT\"\nlicense-file = \"LICENSE\"\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedCargoSurface(f, testCase.seam, testCase.manifest)
			service := NewLicenseService(f)
			_, err := service.Render(renderRequest())
			if !errors.Is(err, ErrInvalidEcosystemSurface) {
				t.Fatalf("Render() error = %v, want ErrInvalidEcosystemSurface", err)
			}
		})
	}
}

func TestRenderReportsCargoReadFailures(t *testing.T) {
	cases := []struct {
		name    string
		blocked string
		wantIs  error
		wantErr string
	}{
		{"seam read failure", seamPath, os.ErrPermission, "read ecosystem declaration"},
		{"manifest read failure", cargoManifestPath, os.ErrPermission, "read cargo manifest"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedCargoSurface(f, rustSeam, "[package]\nlicense = \"MIT\"\n")
			f.readErr[testCase.blocked] = testCase.wantIs
			service := NewLicenseService(f)
			_, err := service.Render(renderRequest())
			if !errors.Is(err, testCase.wantIs) || !strings.Contains(err.Error(), testCase.wantErr) {
				t.Fatalf("Render() error = %v, want %v with %q", err, testCase.wantIs, testCase.wantErr)
			}
		})
	}
}

func TestRenderCargoAlignWriteError(t *testing.T) {
	f := seededFS(t)
	seedCargoSurface(f, rustSeam, "[package]\nlicense = \"MIT\"\n")
	f.writeErr[cargoManifestPath] = fmt.Errorf("disk full")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err == nil {
		t.Fatal("Render() expected cargo manifest write error")
	}
}

func TestPlanRenderSkipsTheAbsentDeclaredCargoManifest(t *testing.T) {
	f := seededFS(t)
	seedCargoSurface(f, rustSeam, "")
	service := NewLicenseService(f)
	plan, err := service.PlanRender(renderRequest())
	if err != nil {
		t.Fatalf("PlanRender() error = %v", err)
	}
	if len(plan.Skips) != 1 || !strings.Contains(plan.Skips[0], "absent") {
		t.Fatalf("PlanRender() skips = %v", plan.Skips)
	}
	if len(plan.Alignments) != 0 {
		t.Fatalf("PlanRender() alignments = %v", plan.Alignments)
	}
}

func TestVerifyCargoFindingsFailClosed(t *testing.T) {
	cases := []struct {
		name     string
		seam     string
		manifest string
		want     string
	}{
		{"diverging expression", rustSeam, "[package]\nlicense = \"MIT\"\n", "diverges from the lock projection"},
		{"diverging form", rustSeam, "[package]\nlicense-file = \"COPYING\"\n", "diverges from the lock projection"},
		{"missing field", rustSeam, "[package]\nname = \"x\"\n", "license field is missing"},
		{"invalid field", rustSeam, "[package]\nlicense = 7\n", "not a license expression string"},
		{"unscannable manifest", rustSeam, "[package", "cannot be proven"},
		{"exclusive keys", rustSeam, "[package]\nlicense = \"MIT\"\nlicense-file = \"LICENSE\"\n", "cannot be proven"},
		{"shadow without declaration", "", "[package]\nlicense = \"MIT\"\n", "without covering declaration"},
		{"shadow with other language", goSeam, "[package]\nlicense = \"MIT\"\n", "without covering declaration"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedCargoSurface(f, testCase.seam, testCase.manifest)
			service := NewLicenseService(f)
			violations, err := service.Verify(verifyRequest(""))
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			found := false
			for _, violation := range violations {
				if strings.Contains(violation, testCase.want) {
					found = true
				}
			}
			if !found {
				t.Fatalf("Verify() violations = %v, want %q", violations, testCase.want)
			}
		})
	}
}

func TestVerifyCargoCleanWithoutCargoSurfaces(t *testing.T) {
	f := seededFS(t)
	seedCargoSurface(f, goSeam, "")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

// --- Maven ecosystem surfaces ---

const mavenSeam = `{"schemaVersion":4,"toolchain":{"language":"maven","version":"4.0.0"}}`

var mavenManifestPath = filepath.Join("out", "pom.xml")

func seedMavenSurface(f *fakeFS, seam, manifest string) {
	if seam != "" {
		f.files[seamPath] = []byte(seam)
	}
	if manifest != "" {
		f.files[mavenManifestPath] = []byte(manifest)
	}
}

func TestRenderAlignsTheDeclaredMavenSurface(t *testing.T) {
	f := seededFS(t)
	seedMavenSurface(f, mavenSeam, "<project>\n    <modelVersion>4.0.0</modelVersion>\n</project>\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !slices.Equal(result.Aligned, []string{mavenManifestPath}) {
		t.Fatalf("Render() aligned = %v", result.Aligned)
	}
	if len(result.Skipped) != 0 {
		t.Fatalf("Render() skipped = %v", result.Skipped)
	}
	content := string(f.files[mavenManifestPath])
	if !strings.Contains(content, "<name>license-hub-NoRepublish-1.0</name>") || !strings.Contains(content, "<url>https://github.com/t33n-software/license-hub</url>") {
		t.Fatalf("Render() manifest = %q", content)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderMavenAlignmentIsIdempotent(t *testing.T) {
	f := seededFS(t)
	seedMavenSurface(f, mavenSeam, "<project>\n    <licenses>\n        <license>\n            <name>license-hub-NoRepublish-1.0</name>\n            <url>https://github.com/t33n-software/license-hub</url>\n        </license>\n    </licenses>\n</project>\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
}

func TestPlanRenderPreviewsTheMavenAlignment(t *testing.T) {
	f := seededFS(t)
	seedMavenSurface(f, mavenSeam, "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n")
	service := NewLicenseService(f)
	plan, err := service.PlanRender(renderRequest())
	if err != nil {
		t.Fatalf("PlanRender() error = %v", err)
	}
	if len(plan.Alignments) != 1 || !strings.Contains(plan.Alignments[0], mavenManifestPath) {
		t.Fatalf("PlanRender() alignments = %v", plan.Alignments)
	}
	if !strings.Contains(string(f.files[mavenManifestPath]), "<name>MIT</name>") {
		t.Fatalf("PlanRender() mutated the manifest: %q", f.files[mavenManifestPath])
	}
}

func TestRenderSkipsTheAbsentDeclaredMavenManifest(t *testing.T) {
	f := seededFS(t)
	seedMavenSurface(f, mavenSeam, "")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 {
		t.Fatalf("Render() aligned = %v", result.Aligned)
	}
	if len(result.Skipped) != 1 || !strings.Contains(result.Skipped[0], "absent") {
		t.Fatalf("Render() skipped = %v", result.Skipped)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "manifest missing") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderWithoutDeclarationLeavesMavenSurfacesUntouched(t *testing.T) {
	f := seededFS(t)
	seedMavenSurface(f, "", "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
}

func TestRenderWithNonMavenDeclarationLeavesMavenSurfacesUntouched(t *testing.T) {
	f := seededFS(t)
	seedMavenSurface(f, goSeam, "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n")
	service := NewLicenseService(f)
	result, err := service.Render(renderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Aligned) != 0 || len(result.Skipped) != 0 {
		t.Fatalf("Render() aligned = %v, skipped = %v", result.Aligned, result.Skipped)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 1 || !strings.Contains(violations[0], "without covering declaration") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestRenderRejectsAnInvalidMavenSurface(t *testing.T) {
	cases := []struct {
		name     string
		seam     string
		manifest string
	}{
		{"invalid seam", "{", ""},
		{"invalid manifest", mavenSeam, "<project>\n    <licenses>\n"},
		{"multiple license elements", mavenSeam, "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n        <license/>\n    </licenses>\n</project>\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedMavenSurface(f, testCase.seam, testCase.manifest)
			service := NewLicenseService(f)
			_, err := service.Render(renderRequest())
			if !errors.Is(err, ErrInvalidEcosystemSurface) {
				t.Fatalf("Render() error = %v, want ErrInvalidEcosystemSurface", err)
			}
		})
	}
}

func TestRenderReportsMavenReadFailures(t *testing.T) {
	cases := []struct {
		name    string
		blocked string
		wantIs  error
		wantErr string
	}{
		{"seam read failure", seamPath, os.ErrPermission, "read ecosystem declaration"},
		{"manifest read failure", mavenManifestPath, os.ErrPermission, "read maven manifest"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedMavenSurface(f, mavenSeam, "<project>\n</project>\n")
			f.readErr[testCase.blocked] = testCase.wantIs
			service := NewLicenseService(f)
			_, err := service.Render(renderRequest())
			if !errors.Is(err, testCase.wantIs) || !strings.Contains(err.Error(), testCase.wantErr) {
				t.Fatalf("Render() error = %v, want %v with %q", err, testCase.wantIs, testCase.wantErr)
			}
		})
	}
}

func TestRenderMavenAlignWriteError(t *testing.T) {
	f := seededFS(t)
	seedMavenSurface(f, mavenSeam, "<project>\n</project>\n")
	f.writeErr[mavenManifestPath] = fmt.Errorf("disk full")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err == nil {
		t.Fatal("Render() expected maven manifest write error")
	}
}

func TestPlanRenderSkipsTheAbsentDeclaredMavenManifest(t *testing.T) {
	f := seededFS(t)
	seedMavenSurface(f, mavenSeam, "")
	service := NewLicenseService(f)
	plan, err := service.PlanRender(renderRequest())
	if err != nil {
		t.Fatalf("PlanRender() error = %v", err)
	}
	if len(plan.Skips) != 1 || !strings.Contains(plan.Skips[0], "absent") {
		t.Fatalf("PlanRender() skips = %v", plan.Skips)
	}
	if len(plan.Alignments) != 0 {
		t.Fatalf("PlanRender() alignments = %v", plan.Alignments)
	}
}

func TestVerifyMavenFindingsFailClosed(t *testing.T) {
	cases := []struct {
		name     string
		seam     string
		manifest string
		want     string
	}{
		{"licenses missing", mavenSeam, "<project>\n</project>\n", "licenses element is missing"},
		{"multiple license elements", mavenSeam, "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n        <license/>\n    </licenses>\n</project>\n", "multiple license elements"},
		{"name missing", mavenSeam, "<project>\n    <licenses>\n        <license>\n            <url>https://example.org</url>\n        </license>\n    </licenses>\n</project>\n", "name element is missing"},
		{"name diverging", mavenSeam, "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n", "name diverges from the lock projection"},
		{"url missing", mavenSeam, "<project>\n    <licenses>\n        <license>\n            <name>license-hub-NoRepublish-1.0</name>\n        </license>\n    </licenses>\n</project>\n", "url element is missing"},
		{"url diverging", mavenSeam, "<project>\n    <licenses>\n        <license>\n            <name>license-hub-NoRepublish-1.0</name>\n            <url>https://example.org</url>\n        </license>\n    </licenses>\n</project>\n", "url diverges from the lock projection"},
		{"unscannable manifest", mavenSeam, "<project>\n    <licenses>\n", "cannot be proven"},
		{"shadow without declaration", "", "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n", "without covering declaration"},
		{"shadow with other language", goSeam, "<project>\n    <licenses>\n        <license>\n            <name>MIT</name>\n        </license>\n    </licenses>\n</project>\n", "without covering declaration"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			f := seededFS(t)
			seedMavenSurface(f, testCase.seam, testCase.manifest)
			service := NewLicenseService(f)
			violations, err := service.Verify(verifyRequest(""))
			if err != nil {
				t.Fatalf("Verify() error = %v", err)
			}
			found := false
			for _, violation := range violations {
				if strings.Contains(violation, testCase.want) {
					found = true
				}
			}
			if !found {
				t.Fatalf("Verify() violations = %v, want %q", violations, testCase.want)
			}
		})
	}
}

func TestVerifyMavenCleanWithoutMavenSurfaces(t *testing.T) {
	f := seededFS(t)
	seedMavenSurface(f, goSeam, "")
	service := NewLicenseService(f)
	if _, err := service.Render(renderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	violations, err := service.Verify(verifyRequest(""))
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}
