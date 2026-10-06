package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestObjectPath(t *testing.T) {
	got := ObjectPath("abc123")
	want := filepath.Join(DEFAULT_PATH, "objects", "ab", "c123")
	if got != want {
		t.Errorf("ObjectPath() = %q, want %q", got, want)
	}
}

func TestWriteReadObjectRoundTrip(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(oldDir) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(DEFAULT_PATH, "objects"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content := []byte("some file content\nwith two lines\n")
	hash := HashContent(content)

	if err := WriteObject(hash, content); err != nil {
		t.Fatalf("WriteObject: %v", err)
	}

	// Stored bytes must be compressed (different from the raw content).
	stored, err := os.ReadFile(ObjectPath(hash))
	if err != nil {
		t.Fatalf("read stored object: %v", err)
	}
	if string(stored) == string(content) {
		t.Error("stored object should be compressed, not raw content")
	}

	got, err := ReadObject(hash)
	if err != nil {
		t.Fatalf("ReadObject: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("round-trip mismatch: got %q, want %q", got, content)
	}
}

func TestIsInitialized(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(oldDir) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if IsInitialized() {
		t.Error("IsInitialized should be false in an empty directory")
	}

	if err := os.Mkdir(DEFAULT_PATH, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if !IsInitialized() {
		t.Error("IsInitialized should be true after creating the repo directory")
	}
}
