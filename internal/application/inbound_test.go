package application

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

// inboundMemFS is the in-memory FileSystem seam of the inbound tests.
type inboundMemFS struct {
	files    map[string]string
	failPath string
}

func newInboundMemFS(files map[string]string) *inboundMemFS {
	return &inboundMemFS{files: files}
}

func (m *inboundMemFS) ReadFile(path string) ([]byte, error) {
	content, ok := m.files[path]
	if !ok {
		return nil, fmt.Errorf("open %s: %w", path, os.ErrNotExist)
	}
	return []byte(content), nil
}

func (m *inboundMemFS) WriteFile(path string, data []byte) error {
	if m.failPath == path {
		return errors.New("simulated write failure")
	}
	m.files[path] = string(data)
	return nil
}

func (m *inboundMemFS) ListNames(string) ([]string, error) {
	return nil, os.ErrNotExist
}

const (
	inboundOrgDefaults = `{"COPYRIGHT_HOLDER":"Org","GOVERNING_LAW":"Germany","VENUE":"DE","PERMISSION_CONTACT":"https://example"}`
	inboundInstance    = "instance {{PROJECT_NAME}} {{LICENSE_ID}} {{COPYRIGHT_YEAR}} {{COPYRIGHT_HOLDER}} {{CANONICAL_SOURCE_URL}} {{PERMISSION_CONTACT}} {{GOVERNING_LAW}} {{VENUE}}\n"
	inboundPlainValues = `{"PROJECT_NAME":"p","LICENSE_ID":"p-NoRepublish-1.0","COPYRIGHT_YEAR":"2026","CANONICAL_SOURCE_URL":"https://example/p"}`
	inboundMITValues   = `{"PROJECT_NAME":"p","LICENSE_ID":"p-NoRepublish-1.0","COPYRIGHT_YEAR":"2026","CANONICAL_SOURCE_URL":"https://example/p",` +
		`"INBOUND_LICENSES":[{"id":"MIT","template":"mit.hbs","copyright_year":"2017","copyright_holder":"Upstream Author"}]}`
	inboundMITTemplate = "Copyright (c) {{COPYRIGHT_YEAR}} {{COPYRIGHT_HOLDER}}\n"
)

func inboundTestFiles(values string) map[string]string {
	return map[string]string{
		"t.hbs":       inboundInstance,
		"org.json":    inboundOrgDefaults,
		"values.json": values,
		"mit.hbs":     inboundMITTemplate,
	}
}

func inboundRenderRequest() RenderRequest {
	return RenderRequest{TemplatePath: "t.hbs", OrgDefaultsPath: "org.json", ValuesPath: "values.json", OutDir: "out"}
}

func inboundVerifyRequest() VerifyRequest {
	return VerifyRequest{TemplatePath: "t.hbs", OrgDefaultsPath: "org.json", ValuesPath: "values.json", Dir: "out"}
}

func TestRenderWithoutInboundWritesNoInboundFiles(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(inboundPlainValues))
	result, err := NewLicenseService(store).Render(inboundRenderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Written) != 2 {
		t.Fatalf("Render() written = %v", result.Written)
	}
}

func TestRenderWritesInboundFiles(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(inboundMITValues))
	result, err := NewLicenseService(store).Render(inboundRenderRequest())
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(result.Written) != 4 {
		t.Fatalf("Render() written = %v", result.Written)
	}
	if got := store.files["out/LICENSES/MIT.txt"]; got != "Copyright (c) 2017 Upstream Author\n" {
		t.Fatalf("inbound text = %q", got)
	}
	notices := store.files["out/THIRD_PARTY_NOTICES.md"]
	for _, want := range []string{"## MIT", "Copyright (c) 2017 Upstream Author", "LICENSES/MIT.txt"} {
		if !strings.Contains(notices, want) {
			t.Fatalf("notices document misses %q:\n%s", want, notices)
		}
	}
}

func TestRenderInboundTemplateMissingFailsClosed(t *testing.T) {
	files := inboundTestFiles(inboundMITValues)
	delete(files, "mit.hbs")
	store := newInboundMemFS(files)
	if _, err := NewLicenseService(store).Render(inboundRenderRequest()); !errors.Is(err, ErrInvalidInboundDeclaration) {
		t.Fatalf("Render() error = %v, want ErrInvalidInboundDeclaration", err)
	}
}

func TestRenderInboundTemplateUnresolvedAnchorFailsClosed(t *testing.T) {
	files := inboundTestFiles(inboundMITValues)
	files["mit.hbs"] = "Copyright {{COPYRIGHT_YEAR}} {{VENUE}}\n"
	store := newInboundMemFS(files)
	if _, err := NewLicenseService(store).Render(inboundRenderRequest()); !errors.Is(err, ErrInvalidInboundDeclaration) {
		t.Fatalf("Render() error = %v, want ErrInvalidInboundDeclaration", err)
	}
}

func TestRenderInboundWriteFailurePropagates(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(inboundMITValues))
	store.failPath = "out/LICENSES/MIT.txt"
	if _, err := NewLicenseService(store).Render(inboundRenderRequest()); err == nil ||
		!strings.Contains(err.Error(), "write out/LICENSES/MIT.txt") {
		t.Fatalf("Render() error = %v", err)
	}
}

func TestPlanRenderIncludesInboundPaths(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(inboundMITValues))
	plan, err := NewLicenseService(store).PlanRender(inboundRenderRequest())
	if err != nil {
		t.Fatalf("PlanRender() error = %v", err)
	}
	if len(plan.Targets) != 4 {
		t.Fatalf("PlanRender() targets = %v", plan.Targets)
	}
	for path := range store.files {
		if strings.HasPrefix(path, "out/") {
			t.Fatalf("PlanRender() wrote %s", path)
		}
	}
}

func TestInboundEntriesRejectsAnInvalidValuesDocument(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(inboundMITValues))
	if _, err := NewLicenseService(store).inboundEntries([]byte(`{`)); !errors.Is(err, ErrInvalidInboundDeclaration) {
		t.Fatalf("inboundEntries() error = %v, want ErrInvalidInboundDeclaration", err)
	}
}

func TestVerifyInboundSuccess(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(inboundMITValues))
	service := NewLicenseService(store)
	if _, err := service.Render(inboundRenderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	violations, err := service.Verify(inboundVerifyRequest())
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestVerifyInboundMissingAndDriftedArtifacts(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(inboundMITValues))
	service := NewLicenseService(store)
	if _, err := service.Render(inboundRenderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	delete(store.files, "out/LICENSES/MIT.txt")
	store.files["out/THIRD_PARTY_NOTICES.md"] = "hand edited\n"
	violations, err := service.Verify(inboundVerifyRequest())
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !strings.Contains(strings.Join(violations, "\n"), "missing inbound license text: out/LICENSES/MIT.txt") {
		t.Fatalf("Verify() violations = %v", violations)
	}
	if !strings.Contains(strings.Join(violations, "\n"), "inbound notices document drifted from the declaration") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestVerifyInboundDriftedTextAndMissingNotices(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(inboundMITValues))
	service := NewLicenseService(store)
	if _, err := service.Render(inboundRenderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	store.files["out/LICENSES/MIT.txt"] = "Copyright (c) 2017 Someone Else\n"
	delete(store.files, "out/THIRD_PARTY_NOTICES.md")
	violations, err := service.Verify(inboundVerifyRequest())
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	joined := strings.Join(violations, "\n")
	if !strings.Contains(joined, "inbound license text drifted from its canonical template: out/LICENSES/MIT.txt") {
		t.Fatalf("Verify() violations = %v", violations)
	}
	if !strings.Contains(joined, "missing inbound notices document: out/THIRD_PARTY_NOTICES.md") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestVerifyInboundTemplateMissingIsAFinding(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(inboundMITValues))
	service := NewLicenseService(store)
	if _, err := service.Render(inboundRenderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	delete(store.files, "mit.hbs")
	violations, err := service.Verify(inboundVerifyRequest())
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) == 0 || !strings.Contains(violations[0], "read inbound template mit.hbs") {
		t.Fatalf("Verify() violations = %v", violations)
	}
}

func TestInboundInvalidDeclarationFailsClosed(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(`{"PROJECT_NAME":"p","LICENSE_ID":"l","COPYRIGHT_YEAR":"2026","CANONICAL_SOURCE_URL":"u","INBOUND_LICENSES":[]}`))
	if _, err := NewLicenseService(store).Render(inboundRenderRequest()); !errors.Is(err, ErrInvalidInboundDeclaration) {
		t.Fatalf("Render() error = %v, want ErrInvalidInboundDeclaration", err)
	}
}

func TestVerifyInboundInvalidDeclarationIsAnError(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(`{"PROJECT_NAME":"p","LICENSE_ID":"l","COPYRIGHT_YEAR":"2026","CANONICAL_SOURCE_URL":"u","INBOUND_LICENSES":[]}`))
	if _, err := NewLicenseService(store).Verify(inboundVerifyRequest()); !errors.Is(err, ErrInvalidInboundDeclaration) {
		t.Fatalf("Verify() error = %v, want ErrInvalidInboundDeclaration", err)
	}
}

func TestInboundValuesUnreadableFailsClosed(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(inboundMITValues))
	delete(store.files, "values.json")
	if _, err := NewLicenseService(store).Render(inboundRenderRequest()); err == nil ||
		!strings.Contains(err.Error(), "read values values.json") {
		t.Fatalf("Render() error = %v", err)
	}
}

func TestVerifyWithoutInboundHasNoInboundFindings(t *testing.T) {
	store := newInboundMemFS(inboundTestFiles(inboundPlainValues))
	service := NewLicenseService(store)
	if _, err := service.Render(inboundRenderRequest()); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	violations, err := service.Verify(inboundVerifyRequest())
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("Verify() violations = %v", violations)
	}
}
