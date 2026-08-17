package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type recordedStep struct {
	executable string
	arguments  []string
}

func fakeExecute(failOn string) (commandRunner, *[]recordedStep) {
	recorded := make([]recordedStep, 0)
	return func(_ context.Context, executable string, arguments ...string) ([]byte, error) {
		recorded = append(recorded, recordedStep{executable: executable, arguments: arguments})
		name := executable + " " + strings.Join(arguments, " ")
		if failOn != "" && strings.Contains(name, failOn) {
			return []byte("failure output"), errors.New("exit status 1")
		}
		return nil, nil
	}, &recorded
}

func fakeFinder(files []string, err error) goFileFinder {
	return func(string) ([]string, error) {
		return files, err
	}
}

func runBuild(execute commandRunner, locate goFileFinder, makeDir directoryCreator) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), nil, &stdout, &stderr, execute, locate, makeDir)
	return code, stdout.String(), stderr.String()
}

func okDir(string, os.FileMode) error { return nil }

func TestRunUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	execute, _ := fakeExecute("")
	code := run(context.Background(), []string{"unexpected"}, &stdout, &stderr, execute, fakeFinder(nil, nil), okDir)
	if code != 2 || !strings.Contains(stderr.String(), "usage:") {
		t.Fatalf("run() = %d, %q", code, stderr.String())
	}
}

func TestRunNilContext(t *testing.T) {
	execute, _ := fakeExecute("")
	code, _, _ := runBuildWithNilContext(execute)
	if code != 0 {
		t.Fatalf("run() = %d, want 0", code)
	}
}

func runBuildWithNilContext(execute commandRunner) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := run(testNilContext(), nil, &stdout, &stderr, execute, fakeFinder(nil, nil), okDir)
	return code, stdout.String(), stderr.String()
}

func testNilContext() context.Context {
	return nil
}

func TestRunSuccess(t *testing.T) {
	execute, recorded := fakeExecute("")
	code, stdout, _ := runBuild(execute, fakeFinder(nil, nil), okDir)
	if code != 0 || !strings.Contains(stdout, "Build completed successfully.") {
		t.Fatalf("run() = %d, %q", code, stdout)
	}
	joined := ""
	for _, step := range *recorded {
		joined += step.executable + " " + strings.Join(step.arguments, " ") + "\n"
	}
	for _, want := range []string{"mod tidy -diff", "staticcheck", "govulncheck", "lefthook validate", "check-coverage", "build -mod=readonly -trimpath"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected step %q in:\n%s", want, joined)
		}
	}
}

func TestRunDependencyStepFailure(t *testing.T) {
	execute, _ := fakeExecute("mod download")
	code, _, stderr := runBuild(execute, fakeFinder(nil, nil), okDir)
	if code != 1 || !strings.Contains(stderr, "download module dependencies") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestRunFormattingListError(t *testing.T) {
	execute, _ := fakeExecute("")
	code, _, stderr := runBuild(execute, fakeFinder(nil, errors.New("walk failed")), okDir)
	if code != 1 || !strings.Contains(stderr, "list Go files:") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestRunFormattingCommandError(t *testing.T) {
	execute, _ := fakeExecute("gofmt")
	code, _, stderr := runBuild(execute, fakeFinder(nil, nil), okDir)
	if code != 1 || !strings.Contains(stderr, "check Go formatting:") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestRunUnformattedFiles(t *testing.T) {
	execute := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return []byte("cmd/license/main.go\n"), nil
	}
	code, _, stderr := runBuild(execute, fakeFinder([]string{"cmd/license/main.go"}, nil), okDir)
	if code != 1 || !strings.Contains(stderr, "require gofmt") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestRunQualityStepFailure(t *testing.T) {
	execute, _ := fakeExecute("staticcheck")
	code, _, stderr := runBuild(execute, fakeFinder(nil, nil), okDir)
	if code != 1 || !strings.Contains(stderr, "run lint") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestRunMakeDirectoryError(t *testing.T) {
	execute, _ := fakeExecute("")
	code, _, stderr := runBuild(execute, fakeFinder(nil, nil), func(string, os.FileMode) error {
		return errors.New("denied")
	})
	if code != 1 || !strings.Contains(stderr, "create build directory:") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestRunBuildStepFailure(t *testing.T) {
	execute, _ := fakeExecute("build -mod=readonly")
	code, _, stderr := runBuild(execute, fakeFinder(nil, nil), okDir)
	if code != 1 || !strings.Contains(stderr, "build native binary") {
		t.Fatalf("run() = %d, %q", code, stderr)
	}
}

func TestGoFilesWalksAndSorts(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		filepath.Join(dir, "b.go"),
		filepath.Join(dir, "a.go"),
		filepath.Join(dir, "note.txt"),
		filepath.Join(dir, "vendor", "skipped.go"),
		filepath.Join(dir, ".build", "skipped.go"),
	}
	for _, file := range files {
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatalf("setup mkdir: %v", err)
		}
		if err := os.WriteFile(file, []byte("package x"), 0o644); err != nil {
			t.Fatalf("setup write: %v", err)
		}
	}
	got, err := goFiles(dir)
	if err != nil {
		t.Fatalf("goFiles() error = %v", err)
	}
	if len(got) != 2 || !strings.HasSuffix(got[0], "a.go") || !strings.HasSuffix(got[1], "b.go") {
		t.Fatalf("goFiles() = %v", got)
	}
}

func TestGoFilesMissingRoot(t *testing.T) {
	if _, err := goFiles(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("goFiles() expected error for missing root")
	}
}

func TestIgnoredDirectory(t *testing.T) {
	for _, name := range []string{".build", ".git", ".cache", "coverage", "dist", "vendor"} {
		if !ignoredDirectory(name) {
			t.Fatalf("ignoredDirectory(%q) = false", name)
		}
	}
	if ignoredDirectory("internal") {
		t.Fatal("ignoredDirectory(internal) = true")
	}
}

func TestBinaryPathFor(t *testing.T) {
	if got := binaryPathFor("linux"); got != filepath.Join(".build", "bin", "license") {
		t.Fatalf("binaryPathFor(linux) = %q", got)
	}
	if got := binaryPathFor("windows"); !strings.HasSuffix(got, "license.exe") {
		t.Fatalf("binaryPathFor(windows) = %q", got)
	}
	if got := binaryPath(); !strings.Contains(got, "license") {
		t.Fatalf("binaryPath() = %q", got)
	}
}

func TestRunCommand(t *testing.T) {
	output, err := runCommand(context.Background(), "go", "version")
	if err != nil {
		t.Fatalf("runCommand() error = %v", err)
	}
	if !strings.Contains(string(output), "go version") {
		t.Fatalf("runCommand() = %q", output)
	}
}

func TestMainExitsWithRunResult(t *testing.T) {
	oldExit, oldArgs, oldRunner, oldFinder, oldMaker := exitProcess, commandArgs, runExternalCommand, findGoFiles, createDirectory
	defer func() {
		exitProcess, commandArgs, runExternalCommand, findGoFiles, createDirectory = oldExit, oldArgs, oldRunner, oldFinder, oldMaker
	}()
	commandArgs = []string{"build"}
	runExternalCommand, _ = fakeExecute("")
	findGoFiles = fakeFinder(nil, nil)
	createDirectory = okDir
	got := -1
	exitProcess = func(code int) { got = code }
	main()
	if got != 0 {
		t.Fatalf("main() exit = %d, want 0", got)
	}
}
