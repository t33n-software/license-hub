package digest

import "testing"

func TestSHA256KnownVector(t *testing.T) {
	want := "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got := SHA256([]byte{}); got != want {
		t.Fatalf("SHA256(empty) = %q, want %q", got, want)
	}
}

func TestSHA256HasPrefixAndHexLength(t *testing.T) {
	got := SHA256([]byte("license-hub"))
	if len(got) != len("sha256:")+64 {
		t.Fatalf("SHA256() length = %d, want %d", len(got), len("sha256:")+64)
	}
	if got[:len("sha256:")] != "sha256:" {
		t.Fatalf("SHA256() = %q, want sha256: prefix", got)
	}
}
