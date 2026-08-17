package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/t33n-software/license-hub/internal/adapters/fs"
	"github.com/t33n-software/license-hub/internal/application"
)

const testTemplate = "{{PROJECT_NAME}} (c) {{COPYRIGHT_YEAR}} {{COPYRIGHT_HOLDER}}\n" +
	"{{CANONICAL_SOURCE_URL}} {{PERMISSION_CONTACT}} {{GOVERNING_LAW}} {{VENUE}} {{LICENSE_ID}}\n"

func seedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"template.hbs": testTemplate,
		"org.json": `{"COPYRIGHT_HOLDER":"CyberT33N","GOVERNING_LAW":"the Federal Republic of Germany",` +
			`"VENUE":"Germany","PERMISSION_CONTACT":"https://github.com/t33n-software"}`,
		"values.json": `{"PROJECT_NAME":"license-hub","LICENSE_ID":"license-hub-NoRepublish-1.0",` +
			`"COPYRIGHT_YEAR":"2026","CANONICAL_SOURCE_URL":"https://github.com/t33n-software/license-hub"}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}
	return dir
}

func runCLI(arguments ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(arguments, &stdout, &stderr, application.NewLicenseService(fs.New()))
	return code, stdout.String(), stderr.String()
}

func TestRunWithoutArguments(t *testing.T) {
	code, _, stderr := runCLI()
	if code != 2 || !strings.Contains(stderr, "usage:") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	code, _, stderr := runCLI("frobnicate")
	if code != 2 || !strings.Contains(stderr, "unknown command") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestRunVersion(t *testing.T) {
	code, stdout, _ := runCLI("version")
	if code != 0 || !strings.Contains(stdout, "license devel") {
		t.Fatalf("run(version) = %d, %q", code, stdout)
	}
}

func TestRunRenderSuccess(t *testing.T) {
	dir := seedRepo(t)
	out := filepath.Join(dir, "out")
	code, stdout, _ := runCLI("render",
		"--template", filepath.Join(dir, "template.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
		"--out", out)
	if code != 0 || !strings.Contains(stdout, "template digest") {
		t.Fatalf("run(render) = %d, %q", code, stdout)
	}
	content, err := os.ReadFile(filepath.Join(out, "LICENSE"))
	if err != nil {
		t.Fatalf("read rendered LICENSE: %v", err)
	}
	if !strings.Contains(string(content), "license-hub (c) 2026 CyberT33N") {
		t.Fatalf("rendered content = %q", content)
	}
}

func TestRunRenderMissingFlags(t *testing.T) {
	code, _, stderr := runCLI("render")
	if code != 2 || !strings.Contains(stderr, "requires") {
		t.Fatalf("run(render) = %d, %q", code, stderr)
	}
}

func TestRunRenderFlagParseError(t *testing.T) {
	code, _, _ := runCLI("render", "--bogus")
	if code != 2 {
		t.Fatalf("run(render --bogus) = %d", code)
	}
}

func TestRunRenderServiceError(t *testing.T) {
	dir := seedRepo(t)
	code, _, stderr := runCLI("render",
		"--template", filepath.Join(dir, "missing.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
		"--out", filepath.Join(dir, "out"))
	if code != 1 || !strings.Contains(stderr, "render:") {
		t.Fatalf("run(render) = %d, %q", code, stderr)
	}
}

func TestRunVerifySuccess(t *testing.T) {
	dir := seedRepo(t)
	out := filepath.Join(dir, "out")
	if code, _, _ := runCLI("render",
		"--template", filepath.Join(dir, "template.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
		"--out", out); code != 0 {
		t.Fatal("render failed")
	}
	code, stdout, _ := runCLI("verify",
		"--template", filepath.Join(dir, "template.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
		"--dir", out)
	if code != 0 || !strings.Contains(stdout, "matches the canonical render") {
		t.Fatalf("run(verify) = %d, %q", code, stdout)
	}
}

func TestRunVerifyViolations(t *testing.T) {
	dir := seedRepo(t)
	code, _, stderr := runCLI("verify",
		"--template", filepath.Join(dir, "template.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
		"--dir", filepath.Join(dir, "out"))
	if code != 1 || !strings.Contains(stderr, "violation:") {
		t.Fatalf("run(verify) = %d, %q", code, stderr)
	}
}

func TestRunVerifyMissingFlags(t *testing.T) {
	code, _, stderr := runCLI("verify")
	if code != 2 || !strings.Contains(stderr, "requires") {
		t.Fatalf("run(verify) = %d, %q", code, stderr)
	}
}

func TestRunVerifyFlagParseError(t *testing.T) {
	code, _, _ := runCLI("verify", "--bogus")
	if code != 2 {
		t.Fatalf("run(verify --bogus) = %d", code)
	}
}

func TestRunVerifyServiceError(t *testing.T) {
	dir := seedRepo(t)
	code, _, stderr := runCLI("verify",
		"--template", filepath.Join(dir, "missing.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
		"--dir", dir)
	if code != 1 || !strings.Contains(stderr, "verify:") {
		t.Fatalf("run(verify) = %d, %q", code, stderr)
	}
}

func TestRunDigestSuccess(t *testing.T) {
	dir := seedRepo(t)
	code, stdout, _ := runCLI("digest", "--template", filepath.Join(dir, "template.hbs"))
	if code != 0 || !strings.Contains(stdout, "sha256:") {
		t.Fatalf("run(digest) = %d, %q", code, stdout)
	}
}

func TestRunDigestMissingFlag(t *testing.T) {
	code, _, stderr := runCLI("digest")
	if code != 2 || !strings.Contains(stderr, "requires") {
		t.Fatalf("run(digest) = %d, %q", code, stderr)
	}
}

func TestRunDigestFlagParseError(t *testing.T) {
	code, _, _ := runCLI("digest", "--bogus")
	if code != 2 {
		t.Fatalf("run(digest --bogus) = %d", code)
	}
}

func TestRunDigestServiceError(t *testing.T) {
	code, _, stderr := runCLI("digest", "--template", filepath.Join(t.TempDir(), "missing.hbs"))
	if code != 1 || !strings.Contains(stderr, "digest:") {
		t.Fatalf("run(digest) = %d, %q", code, stderr)
	}
}

func TestMainExitsWithRunResult(t *testing.T) {
	oldExit, oldArgs := exitProcess, commandArgs
	defer func() { exitProcess, commandArgs = oldExit, oldArgs }()
	commandArgs = []string{"license", "version"}
	got := -1
	exitProcess = func(code int) { got = code }
	main()
	if got != 0 {
		t.Fatalf("main() exit = %d, want 0", got)
	}
}
