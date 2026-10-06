package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldIgnore(t *testing.T) {
	cases := map[string]bool{
		"normal.txt":      false,
		".":               false,
		"src/main.go":     false,
		".git/config":     true,
		"sub/.git/config": true,
		".microgit/index": true,
		".DS_Store":       true,
		"dir/.DS_Store":   true,
	}
	for path, want := range cases {
		if got := shouldIgnore(filepath.FromSlash(path)); got != want {
			t.Errorf("shouldIgnore(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestEnsureRepo(t *testing.T) {
	dir := t.TempDir()
	oldDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(oldDir) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := ensureRepo(); err == nil {
		t.Error("ensureRepo should return an error before init")
	}

	captureOutput(func() { initCmd.RunE(nil, nil) })

	if err := ensureRepo(); err != nil {
		t.Errorf("ensureRepo should succeed after init, got %v", err)
	}
}

func TestWriteIndexRoundTrip(t *testing.T) {
	setupTestRepo(t)

	want := map[string]string{"a.txt": "h1", "b.txt": "h2"}
	if err := writeIndex(want); err != nil {
		t.Fatalf("writeIndex: %v", err)
	}

	got, err := readIndex()
	if err != nil {
		t.Fatalf("readIndex: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("index length = %d, want %d", len(got), len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("index[%q] = %q, want %q", k, got[k], v)
		}
	}
}

func TestGetLatest(t *testing.T) {
	setupTestRepo(t)

	if getLatest() != "" {
		t.Errorf("expected empty LATEST on fresh repo, got %q", getLatest())
	}

	os.WriteFile("a.txt", []byte("x"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"a.txt"}) })
	captureOutput(func() { saveCmd.RunE(nil, []string{"c"}) })

	if getLatest() != getHead() {
		t.Errorf("LATEST %q should equal HEAD %q after save", getLatest(), getHead())
	}
}
