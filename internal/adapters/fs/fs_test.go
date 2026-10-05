package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndReadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "nested", "deep", "file.txt")
	s := New()
	if err := s.WriteFile(target, []byte("content")); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	got, err := s.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != "content" {
		t.Fatalf("ReadFile() = %q, want %q", got, "content")
	}
}

func TestReadFileMissing(t *testing.T) {
	s := New()
	if _, err := s.ReadFile(filepath.Join(t.TempDir(), "missing.txt")); err == nil {
		t.Fatal("ReadFile() expected error for missing file")
	}
}

func TestWriteFileFailsWhenParentIsFile(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	s := New()
	if err := s.WriteFile(filepath.Join(blocker, "file.txt"), []byte("x")); err == nil {
		t.Fatal("WriteFile() expected error when parent path is a file")
	}
}

func TestListNamesListsNonDirectoryEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.csproj"), []byte("<Project/>"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.nuspec"), []byte("<package/>"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	s := New()
	names, err := s.ListNames(dir)
	if err != nil {
		t.Fatalf("ListNames() error = %v", err)
	}
	want := []string{"a.csproj", "b.nuspec"}
	if len(names) != len(want) || names[0] != want[0] || names[1] != want[1] {
		t.Fatalf("ListNames() = %v, want %v", names, want)
	}
}

func TestListNamesMissingDirectory(t *testing.T) {
	s := New()
	if _, err := s.ListNames(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("ListNames() expected error for missing directory")
	}
}
