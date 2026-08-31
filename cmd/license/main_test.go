package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/t33n-software/license-hub/internal/adapters/cli"
	"github.com/t33n-software/license-hub/internal/adapters/fs"
	"github.com/t33n-software/license-hub/internal/application"
	"github.com/t33n-software/license-hub/internal/domain/contract"
)

// Convention: docs/conventions/cli/testing/README.md

const testTemplate = "{{PROJECT_NAME}} (c) {{COPYRIGHT_YEAR}} {{COPYRIGHT_HOLDER}}\n" +
	"{{CANONICAL_SOURCE_URL}} {{PERMISSION_CONTACT}} {{GOVERNING_LAW}} {{VENUE}} {{LICENSE_ID}}\n"

// TestMain pins the environment seam to an empty environment so no developer
// machine variable can leak into a test.
func TestMain(m *testing.M) {
	lookupEnv = func(string) (string, bool) { return "", false }
	stdinIsTerminal = func() bool { return false }
	os.Exit(m.Run())
}

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

// withEnv binds the environment seam for one test.
func withEnv(t *testing.T, env map[string]string) {
	t.Helper()
	t.Cleanup(func() { lookupEnv = func(string) (string, bool) { return "", false } })
	lookupEnv = func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	}
}

// withTerminal binds the terminal seam and the standard input for one test.
func withTerminal(t *testing.T, input string) {
	t.Helper()
	t.Cleanup(func() {
		stdinIsTerminal = func() bool { return false }
		stdinReader = strings.NewReader("")
	})
	stdinIsTerminal = func() bool { return true }
	stdinReader = strings.NewReader(input)
}

func runCLI(arguments ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(arguments, &stdout, &stderr, application.NewLicenseService(fs.New()))
	return code, stdout.String(), stderr.String()
}

func renderArgs(dir, out string) []string {
	return []string{
		"render",
		"--template", filepath.Join(dir, "template.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
		"--out", out,
	}
}

func verifyArgs(dir, target string) []string {
	return []string{
		"verify",
		"--template", filepath.Join(dir, "template.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
		"--dir", target,
	}
}

// --- Root contract ---

func TestRunWithoutArgumentsPrintsRootHelpAsUsage(t *testing.T) {
	code, _, stderr := runCLI()
	if code != contract.ExitUsage {
		t.Fatalf("run() = %d, want %d", code, contract.ExitUsage)
	}
	if !strings.Contains(stderr, "Usage:") || !strings.Contains(stderr, "Discovery:") {
		t.Fatalf("run() stderr = %q", stderr)
	}
}

func TestRunRootHelp(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"-h"}, {"help"}} {
		code, stdout, _ := runCLI(args...)
		if code != contract.ExitSuccess {
			t.Fatalf("run(%v) = %d", args, code)
		}
		for _, command := range contract.Commands() {
			if !strings.Contains(stdout, command.Name) || !strings.Contains(stdout, command.Summary) {
				t.Fatalf("run(%v) root help misses %q:\n%s", args, command.Name, stdout)
			}
		}
		if !strings.Contains(stdout, "Exit codes: 0 success; 1 execution failure; 2 usage error; 3 governance rejection.") {
			t.Fatalf("run(%v) root help misses the exit codes:\n%s", args, stdout)
		}
	}
}

func TestRunHelpCommandRendersTheLeafHelp(t *testing.T) {
	code, stdout, _ := runCLI("help", "render")
	if code != contract.ExitSuccess {
		t.Fatalf("run(help render) = %d", code)
	}
	if !strings.Contains(stdout, "license render") || !strings.Contains(stdout, "--template string") {
		t.Fatalf("run(help render) = %q", stdout)
	}
}

func TestRunHelpUnknownCommand(t *testing.T) {
	code, _, stderr := runCLI("help", "frobnicate")
	if code != contract.ExitUsage || !strings.Contains(stderr, "USAGE_ERROR") || !strings.Contains(stderr, "frobnicate") {
		t.Fatalf("run(help frobnicate) = %d, %q", code, stderr)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	code, _, stderr := runCLI("frobnicate")
	if code != contract.ExitUsage {
		t.Fatalf("run(frobnicate) = %d, want %d", code, contract.ExitUsage)
	}
	for _, want := range []string{"Error [USAGE_ERROR]", "unknown command", "render, verify, digest, or version"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("run(frobnicate) stderr misses %q:\n%s", want, stderr)
		}
	}
}

func TestRunRootUnknownFlag(t *testing.T) {
	code, _, stderr := runCLI("--bogus")
	if code != contract.ExitUsage || !strings.Contains(stderr, "USAGE_ERROR") {
		t.Fatalf("run(--bogus) = %d, %q", code, stderr)
	}
}

func TestRunRootOutputWithoutValue(t *testing.T) {
	code, _, stderr := runCLI("--output")
	if code != contract.ExitUsage || !strings.Contains(stderr, "USAGE_ERROR") {
		t.Fatalf("run(--output) = %d, %q", code, stderr)
	}
}

func TestRunRootRejectsAnUndocumentedOutputMode(t *testing.T) {
	code, _, stderr := runCLI("--output", "yaml", "--help")
	if code != contract.ExitUsage || !strings.Contains(stderr, "VALUE_INVALID") || !strings.Contains(stderr, "human or json") {
		t.Fatalf("run(--output yaml) = %d, %q", code, stderr)
	}
}

// --- Version contract ---

func TestRunVersion(t *testing.T) {
	code, stdout, _ := runCLI("version")
	if code != contract.ExitSuccess || !strings.Contains(stdout, "version devel") {
		t.Fatalf("run(version) = %d, %q", code, stdout)
	}
}

func TestRunRootVersionIsMachineReadable(t *testing.T) {
	code, stdout, _ := runCLI("--version")
	if code != contract.ExitSuccess || !strings.Contains(stdout, "version devel") {
		t.Fatalf("run(--version) = %d, %q", code, stdout)
	}

	code, stdout, _ = runCLI("--version", "--output", "json")
	if code != contract.ExitSuccess {
		t.Fatalf("run(--version --output json) = %d", code)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("run(--version --output json) does not parse: %v", err)
	}
	if doc["version"] != "devel" || doc["command"] != "version" || doc["status"] != "ok" {
		t.Fatalf("version document = %v", doc)
	}
}

func TestRunVersionCommandJSON(t *testing.T) {
	code, stdout, _ := runCLI("version", "--output=json")
	if code != contract.ExitSuccess {
		t.Fatalf("run(version --output=json) = %d", code)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil || doc["version"] != "devel" {
		t.Fatalf("run(version --output=json) = %q, %v", stdout, err)
	}
}

// --- Help contract on every command node ---

func TestRunLeafHelpOnEveryCommand(t *testing.T) {
	for _, command := range contract.Commands() {
		for _, form := range [][]string{{command.Name, "--help"}, {command.Name, "-h"}, {"--help", command.Name}} {
			code, stdout, _ := runCLI(form...)
			if code != contract.ExitSuccess {
				t.Fatalf("run(%v) = %d", form, code)
			}
			for _, flag := range command.Flags {
				if !strings.Contains(stdout, "--"+flag.Name) || !strings.Contains(stdout, flag.Env) {
					t.Fatalf("run(%v) leaf help misses flag %q:\n%s", form, flag.Name, stdout)
				}
			}
			if !strings.Contains(stdout, "Examples:") || !strings.Contains(stdout, "Stability: stable") {
				t.Fatalf("run(%v) leaf help misses examples or stability:\n%s", form, stdout)
			}
		}
	}
}

// --- Render ---

func TestRunRenderSuccess(t *testing.T) {
	dir := seedRepo(t)
	out := filepath.Join(dir, "out")
	args := append(renderArgs(dir, out), "--yes")
	code, stdout, _ := runCLI(args...)
	if code != contract.ExitSuccess || !strings.Contains(stdout, "template digest") {
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
	if code != contract.ExitUsage || !strings.Contains(stderr, "VALUE_MISSING") {
		t.Fatalf("run(render) = %d, %q", code, stderr)
	}
	if !strings.Contains(stderr, "Field: template") || !strings.Contains(stderr, "LICENSE_TEMPLATE") {
		t.Fatalf("run(render) error record incomplete:\n%s", stderr)
	}
}

func TestRunRenderFlagParseError(t *testing.T) {
	code, _, _ := runCLI("render", "--bogus")
	if code != contract.ExitUsage {
		t.Fatalf("run(render --bogus) = %d", code)
	}
}

func TestRunRenderServiceError(t *testing.T) {
	dir := seedRepo(t)
	args := []string{
		"render",
		"--template", filepath.Join(dir, "missing.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
		"--out", filepath.Join(dir, "out"),
		"--yes",
	}
	code, _, stderr := runCLI(args...)
	if code != contract.ExitExecution || !strings.Contains(stderr, "EXECUTION_FAILED") {
		t.Fatalf("run(render) = %d, %q", code, stderr)
	}
}

func TestRunRenderMissingValuesIsAUsageError(t *testing.T) {
	dir := seedRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "values.json"), []byte(`{"PROJECT_NAME":"x"}`), 0o644); err != nil {
		t.Fatalf("seed values: %v", err)
	}
	args := append(renderArgs(dir, filepath.Join(dir, "out")), "--yes")
	code, _, stderr := runCLI(args...)
	if code != contract.ExitUsage || !strings.Contains(stderr, "VALUE_INVALID") || !strings.Contains(stderr, "missing required values") {
		t.Fatalf("run(render) = %d, %q", code, stderr)
	}
}

func TestRunRenderUnresolvedPlaceholdersIsAUsageError(t *testing.T) {
	dir := seedRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "template.hbs"), []byte(testTemplate+"{{UNKNOWN_ANCHOR}}\n"), 0o644); err != nil {
		t.Fatalf("seed template: %v", err)
	}
	args := append(renderArgs(dir, filepath.Join(dir, "out")), "--yes")
	code, _, stderr := runCLI(args...)
	if code != contract.ExitUsage || !strings.Contains(stderr, "VALUE_INVALID") || !strings.Contains(stderr, "{{UNKNOWN_ANCHOR}}") {
		t.Fatalf("run(render) = %d, %q", code, stderr)
	}
}

func TestRunRenderDryRunWritesNothing(t *testing.T) {
	dir := seedRepo(t)
	out := filepath.Join(dir, "out")
	args := append(renderArgs(dir, out), "--dry-run")
	code, stdout, _ := runCLI(args...)
	if code != contract.ExitSuccess {
		t.Fatalf("run(render --dry-run) = %d", code)
	}
	if !strings.Contains(stdout, "dry-run plan: no files written") || !strings.Contains(stdout, "would write") {
		t.Fatalf("run(render --dry-run) = %q", stdout)
	}
	if _, err := os.Stat(filepath.Join(out, "LICENSE")); !os.IsNotExist(err) {
		t.Fatal("run(render --dry-run) must not write any file")
	}
}

func TestRunRenderDryRunServiceError(t *testing.T) {
	dir := seedRepo(t)
	args := []string{
		"render",
		"--template", filepath.Join(dir, "missing.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
		"--dry-run",
	}
	code, _, stderr := runCLI(args...)
	if code != contract.ExitExecution || !strings.Contains(stderr, "EXECUTION_FAILED") {
		t.Fatalf("run(render --dry-run) = %d, %q", code, stderr)
	}
}

func TestRunRenderWithoutYesFailsClosedNonInteractively(t *testing.T) {
	dir := seedRepo(t)
	out := filepath.Join(dir, "out")
	code, _, stderr := runCLI(renderArgs(dir, out)...)
	if code != contract.ExitUsage || !strings.Contains(stderr, "CONFIRMATION_REQUIRED") {
		t.Fatalf("run(render) = %d, %q", code, stderr)
	}
	if !strings.Contains(stderr, "--yes") {
		t.Fatalf("run(render) remediation must name --yes:\n%s", stderr)
	}
	if _, err := os.Stat(filepath.Join(out, "LICENSE")); !os.IsNotExist(err) {
		t.Fatal("run(render) without confirmation must not write any file")
	}
}

func TestRunRenderInteractiveConfirmationAccepts(t *testing.T) {
	dir := seedRepo(t)
	out := filepath.Join(dir, "out")
	withTerminal(t, "y\n")
	code, stdout, _ := runCLI(renderArgs(dir, out)...)
	if code != contract.ExitSuccess {
		t.Fatalf("run(render) interactive = %d", code)
	}
	if !strings.Contains(stdout, "Apply the render? [y/N]") || !strings.Contains(stdout, "would write") {
		t.Fatalf("run(render) interactive output = %q", stdout)
	}
	if _, err := os.Stat(filepath.Join(out, "LICENSE")); err != nil {
		t.Fatalf("run(render) interactive did not write: %v", err)
	}
}

func TestRunRenderInteractiveConfirmationDeclines(t *testing.T) {
	dir := seedRepo(t)
	out := filepath.Join(dir, "out")
	withTerminal(t, "n\n")
	code, _, stderr := runCLI(renderArgs(dir, out)...)
	if code != contract.ExitUsage || !strings.Contains(stderr, "CONFIRMATION_REQUIRED") {
		t.Fatalf("run(render) interactive decline = %d, %q", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(out, "LICENSE")); !os.IsNotExist(err) {
		t.Fatal("a declined confirmation must not write any file")
	}
}

func TestRunRenderInteractivePlanError(t *testing.T) {
	dir := seedRepo(t)
	withTerminal(t, "y\n")
	args := []string{
		"render",
		"--template", filepath.Join(dir, "missing.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
	}
	code, _, stderr := runCLI(args...)
	if code != contract.ExitExecution || !strings.Contains(stderr, "EXECUTION_FAILED") {
		t.Fatalf("run(render) interactive plan error = %d, %q", code, stderr)
	}
}

func TestRunRenderJSON(t *testing.T) {
	dir := seedRepo(t)
	out := filepath.Join(dir, "out")
	args := append(renderArgs(dir, out), "--yes", "--output", "json")
	code, stdout, stderr := runCLI(args...)
	if code != contract.ExitSuccess {
		t.Fatalf("run(render --output json) = %d, %q", code, stderr)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("run(render --output json) does not parse: %v", err)
	}
	if doc["command"] != "render" || doc["status"] != "ok" {
		t.Fatalf("render document = %v", doc)
	}
	if written, ok := doc["written"].([]any); !ok || len(written) != 2 {
		t.Fatalf("render document written = %v", doc["written"])
	}
	if digest, ok := doc["digest"].(string); !ok || !strings.HasPrefix(digest, "sha256:") {
		t.Fatalf("render document digest = %v", doc["digest"])
	}
}

// --- Verify ---

func TestRunVerifySuccess(t *testing.T) {
	dir := seedRepo(t)
	out := filepath.Join(dir, "out")
	render := append(renderArgs(dir, out), "--yes")
	if code, _, _ := runCLI(render...); code != contract.ExitSuccess {
		t.Fatal("render failed")
	}
	code, stdout, _ := runCLI(verifyArgs(dir, out)...)
	if code != contract.ExitSuccess || !strings.Contains(stdout, "matches the canonical render") {
		t.Fatalf("run(verify) = %d, %q", code, stdout)
	}
}

func TestRunVerifyViolationsAreAGovernanceRejection(t *testing.T) {
	dir := seedRepo(t)
	code, _, stderr := runCLI(verifyArgs(dir, filepath.Join(dir, "out"))...)
	if code != contract.ExitGovernance {
		t.Fatalf("run(verify) = %d, want %d", code, contract.ExitGovernance)
	}
	if !strings.Contains(stderr, "violation:") {
		t.Fatalf("run(verify) stderr = %q", stderr)
	}
}

func TestRunVerifyViolationsJSON(t *testing.T) {
	dir := seedRepo(t)
	args := append(verifyArgs(dir, filepath.Join(dir, "out")), "--output", "json")
	code, stdout, stderr := runCLI(args...)
	if code != contract.ExitGovernance {
		t.Fatalf("run(verify --output json) = %d", code)
	}
	if stderr != "" {
		t.Fatalf("run(verify --output json) must not write to stderr: %q", stderr)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("run(verify --output json) does not parse: %v", err)
	}
	if doc["status"] != "governance_rejected" {
		t.Fatalf("verify document = %v", doc)
	}
	if violations, ok := doc["violations"].([]any); !ok || len(violations) == 0 {
		t.Fatalf("verify document violations = %v", doc["violations"])
	}
}

func TestRunVerifyQuietSuppressesTheSuccessOutput(t *testing.T) {
	dir := seedRepo(t)
	out := filepath.Join(dir, "out")
	render := append(renderArgs(dir, out), "--yes")
	if code, _, _ := runCLI(render...); code != contract.ExitSuccess {
		t.Fatal("render failed")
	}
	code, stdout, _ := runCLI(append(verifyArgs(dir, out), "--quiet")...)
	if code != contract.ExitSuccess || stdout != "" {
		t.Fatalf("run(verify --quiet) = %d, %q", code, stdout)
	}
}

func TestRunVerifyMissingFlags(t *testing.T) {
	code, _, stderr := runCLI("verify")
	if code != contract.ExitUsage || !strings.Contains(stderr, "VALUE_MISSING") {
		t.Fatalf("run(verify) = %d, %q", code, stderr)
	}
}

func TestRunVerifyFlagParseError(t *testing.T) {
	code, _, _ := runCLI("verify", "--bogus")
	if code != contract.ExitUsage {
		t.Fatalf("run(verify --bogus) = %d", code)
	}
}

func TestRunVerifyServiceError(t *testing.T) {
	dir := seedRepo(t)
	args := []string{
		"verify",
		"--template", filepath.Join(dir, "missing.hbs"),
		"--org-defaults", filepath.Join(dir, "org.json"),
		"--values", filepath.Join(dir, "values.json"),
		"--dir", dir,
	}
	code, _, stderr := runCLI(args...)
	if code != contract.ExitExecution || !strings.Contains(stderr, "EXECUTION_FAILED") {
		t.Fatalf("run(verify) = %d, %q", code, stderr)
	}
}

// --- Digest ---

func TestRunDigestSuccess(t *testing.T) {
	dir := seedRepo(t)
	code, stdout, _ := runCLI("digest", "--template", filepath.Join(dir, "template.hbs"))
	if code != contract.ExitSuccess || !strings.Contains(stdout, "sha256:") {
		t.Fatalf("run(digest) = %d, %q", code, stdout)
	}
}

func TestRunDigestMissingFlag(t *testing.T) {
	code, _, stderr := runCLI("digest")
	if code != contract.ExitUsage || !strings.Contains(stderr, "VALUE_MISSING") {
		t.Fatalf("run(digest) = %d, %q", code, stderr)
	}
}

func TestRunDigestFlagParseError(t *testing.T) {
	code, _, _ := runCLI("digest", "--bogus")
	if code != contract.ExitUsage {
		t.Fatalf("run(digest --bogus) = %d", code)
	}
}

func TestRunDigestServiceError(t *testing.T) {
	code, _, stderr := runCLI("digest", "--template", filepath.Join(t.TempDir(), "missing.hbs"))
	if code != contract.ExitExecution || !strings.Contains(stderr, "EXECUTION_FAILED") {
		t.Fatalf("run(digest) = %d, %q", code, stderr)
	}
}

func TestRunDigestParseErrorHonorsTheRequestedJSONMode(t *testing.T) {
	code, stdout, _ := runCLI("digest", "--bogus", "--output", "json")
	if code != contract.ExitUsage {
		t.Fatalf("run(digest --bogus --output json) = %d", code)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("the usage error must render in the requested JSON mode: %q", stdout)
	}
	if doc["status"] != "error" {
		t.Fatalf("error document = %v", doc)
	}
}

func TestRunDigestParseErrorHonorsTheRequestedJSONModeWithEqualsForm(t *testing.T) {
	code, stdout, _ := runCLI("digest", "--bogus", "--output=json")
	if code != contract.ExitUsage {
		t.Fatalf("run(digest --bogus --output=json) = %d", code)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil || doc["status"] != "error" {
		t.Fatalf("run(digest --bogus --output=json) = %q, %v", stdout, err)
	}
}

// --- Environment precedence ---

func TestRunDigestReadsTheTemplateFromTheEnvironment(t *testing.T) {
	dir := seedRepo(t)
	withEnv(t, map[string]string{"LICENSE_TEMPLATE": filepath.Join(dir, "template.hbs")})
	code, stdout, _ := runCLI("digest")
	if code != contract.ExitSuccess || !strings.Contains(stdout, "sha256:") {
		t.Fatalf("run(digest) with LICENSE_TEMPLATE = %d, %q", code, stdout)
	}
}

func TestRunDigestReadsTheOutputModeFromTheEnvironment(t *testing.T) {
	dir := seedRepo(t)
	withEnv(t, map[string]string{"LICENSE_OUTPUT": "json"})
	code, stdout, _ := runCLI("digest", "--template", filepath.Join(dir, "template.hbs"))
	if code != contract.ExitSuccess {
		t.Fatalf("run(digest) with LICENSE_OUTPUT = %d", code)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil || doc["command"] != "digest" {
		t.Fatalf("run(digest) with LICENSE_OUTPUT=json = %q, %v", stdout, err)
	}
}

// --- Help-first consumer proof ---

func TestHelpFirstConsumerDerivesAValidCallFromTheHelpAlone(t *testing.T) {
	render, ok := contract.CommandByName("render")
	if !ok {
		t.Fatal("render not registered")
	}
	var help bytes.Buffer
	cli.WriteLeafHelp(&help, render)
	var example string
	inExamples := false
	for _, line := range strings.Split(help.String(), "\n") {
		if line == "Examples:" {
			inExamples = true
			continue
		}
		if inExamples && strings.HasPrefix(line, "  license render ") {
			example = strings.TrimSpace(line)
			break
		}
	}
	if example == "" {
		t.Fatal("the leaf help carries no canonical render example")
	}

	// The consumer materializes exactly the files the example references.
	dir := t.TempDir()
	files := map[string]string{
		"templates/custom/norepublish/NoRepublish-1.0.0.hbs": testTemplate,
		"org-defaults.json": `{"COPYRIGHT_HOLDER":"CyberT33N","GOVERNING_LAW":"the Federal Republic of Germany",` +
			`"VENUE":"Germany","PERMISSION_CONTACT":"https://github.com/t33n-software"}`,
		"license.values.json": `{"PROJECT_NAME":"license-hub","LICENSE_ID":"license-hub-NoRepublish-1.0",` +
			`"COPYRIGHT_YEAR":"2026","CANONICAL_SOURCE_URL":"https://github.com/t33n-software/license-hub"}`,
	}
	for name, content := range files {
		target := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}
	t.Chdir(dir)

	args := strings.Fields(example)[1:]
	code, stdout, stderr := runCLI(args...)
	if code != contract.ExitSuccess {
		t.Fatalf("the canonical help example must validate and succeed: %d, %q, %q", code, stdout, stderr)
	}
	if _, err := os.Stat("LICENSE"); err != nil {
		t.Fatalf("the canonical help example did not render: %v", err)
	}
}

func TestCharDeviceClassifiesTheFileMode(t *testing.T) {
	if !charDevice(os.ModeCharDevice) {
		t.Fatal("charDevice(ModeCharDevice) = false, want true")
	}
	if charDevice(0) {
		t.Fatal("charDevice(0) = true, want false")
	}
}

func TestDetectTerminalAnswersOnEveryHost(t *testing.T) {
	// The assertion is the call itself: the detection must answer without a
	// panic on every host, whether the test process owns a terminal or not.
	_ = detectTerminal()
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
