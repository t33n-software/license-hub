package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func fakeRunner(output string, err error) commandRunner {
	return func(context.Context, string, ...string) ([]byte, error) {
		return []byte(output), err
	}
}

func runWith(arguments []string, execute commandRunner) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), arguments, &stdout, &stderr, execute)
	return code, stdout.String(), stderr.String()
}

func TestRunUsageError(t *testing.T) {
	code, _, stderr := runWith([]string{"unexpected"}, fakeRunner("", nil))
	if code != 2 || !strings.Contains(stderr, "usage:") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestRunVersion(t *testing.T) {
	called := false
	runner := func(context.Context, string, ...string) ([]byte, error) {
		called = true
		return nil, nil
	}
	code, stdout, _ := runWith([]string{"--version"}, runner)
	if code != 0 || !strings.Contains(stdout, "check-coverage devel") {
		t.Fatalf("run(--version) = %d, %q", code, stdout)
	}
	if called {
		t.Fatal("run(--version) must not execute the coverage command")
	}
}

func TestRunNilContext(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(testNilContext(), nil, &stdout, &stderr, fakeRunner("ok\n", nil))
	if code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}

func testNilContext() context.Context {
	return nil
}

func TestRunSuccess(t *testing.T) {
	output := "ok  \tgithub.com/t33n-software/license-hub/internal/domain/render\t0.2s\tcoverage: 100.0% of statements\n"
	code, stdout, _ := runWith(nil, fakeRunner(output, nil))
	if code != 0 || !strings.Contains(stdout, "100.0% statement coverage") {
		t.Fatalf("run() = %d, %q", code, stdout)
	}
}

func TestRunCommandError(t *testing.T) {
	code, _, stderr := runWith(nil, fakeRunner("boom", errors.New("exit 1")))
	if code != 1 || !strings.Contains(stderr, "run Go coverage:") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestRunMissingTests(t *testing.T) {
	output := "?   \tgithub.com/t33n-software/license-hub/cmd/license\t[no test files]\n"
	code, _, stderr := runWith(nil, fakeRunner(output, nil))
	if code != 1 || !strings.Contains(stderr, "must contain at least one _test.go file") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestRunIncompleteCoverage(t *testing.T) {
	output := "ok  \tgithub.com/t33n-software/license-hub/internal/domain/render\t0.2s\tcoverage: 85.0% of statements\n"
	code, _, stderr := runWith(nil, fakeRunner(output, nil))
	if code != 1 || !strings.Contains(stderr, "100.0% statement coverage") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestPackagesWithoutTests(t *testing.T) {
	output := "?   \tpkg/a\t[no test files]\r\nok  \tpkg/b\t0.1s\tcoverage: 100.0% of statements\n"
	got := packagesWithoutTests(output)
	if len(got) != 1 || !strings.Contains(got[0], "pkg/a") {
		t.Fatalf("packagesWithoutTests() = %v", got)
	}
}

func TestIncompletePackages(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   int
	}{
		{"complete", "ok  \tpkg/a\t0.1s\tcoverage: 100.0% of statements\n", 0},
		{"incomplete", "ok  \tpkg/a\t0.1s\tcoverage: 85.0% of statements\n", 1},
		{"no statements", "ok  \tpkg/a\t0.1s\tcoverage: [no statements]\n", 0},
		{"no coverage line", "?   \tpkg/a\t[no test files]\n", 0},
		{"crlf", "ok  \tpkg/a\t0.1s\tcoverage: 50.0% of statements\r\n", 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := incompletePackages(test.output); len(got) != test.want {
				t.Fatalf("incompletePackages() = %v, want %d entries", got, test.want)
			}
		})
	}
}

func TestRunGoCommand(t *testing.T) {
	output, err := runGoCommand(context.Background(), "go", "version")
	if err != nil {
		t.Fatalf("runGoCommand() error = %v", err)
	}
	if !strings.Contains(string(output), "go version") {
		t.Fatalf("runGoCommand() = %q", output)
	}
}

func TestMainExitsWithRunResult(t *testing.T) {
	oldExit, oldArgs, oldRunner := exitProcess, commandArgs, runCommand
	defer func() { exitProcess, commandArgs, runCommand = oldExit, oldArgs, oldRunner }()
	commandArgs = []string{"check-coverage"}
	runCommand = fakeRunner("ok\n", nil)
	got := -1
	exitProcess = func(code int) { got = code }
	main()
	if got != 0 {
		t.Fatalf("main() exit = %d, want 0", got)
	}
}
