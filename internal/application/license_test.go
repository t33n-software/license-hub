package application

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/t33n-software/license-hub/internal/domain/digest"
)

// Convention: docs/conventions/cli/testing/README.md

type fakeFS struct {
	files    map[string][]byte
	writeErr map[string]error
}

func newFakeFS() *fakeFS {
	return &fakeFS{files: make(map[string][]byte), writeErr: make(map[string]error)}
}

func (f *fakeFS) ReadFile(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("not found: %s", path)
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
