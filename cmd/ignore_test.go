package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestIgnoreMatcherPatterns(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile(".gitignore", []byte("# comment\n*.log\nbuild/\nsecret.txt\n"), 0644)
	m := newIgnoreMatcher()

	cases := map[string]bool{
		"app.log":         true,  // glob
		"logs/app.log":    true,  // glob on basename
		"secret.txt":      true,  // exact
		"build/output.o":  true,  // directory pattern
		"main.go":         false, // not ignored
		".gitignore":      false, // tracked by default
		".git/config":     true,  // built-in
		".microgit/index": true,  // built-in
	}
	for path, want := range cases {
		if got := m.match(path); got != want {
			t.Errorf("match(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestAddDotRespectsGitignore(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile(".gitignore", []byte("*.log\n"), 0644)
	os.WriteFile("keep.txt", []byte("keep"), 0644)
	os.WriteFile("skip.log", []byte("skip"), 0644)

	runCmd(t, addCmd, ".")

	index, err := os.ReadFile(".microgit/index")
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if !strings.Contains(string(index), "keep.txt") {
		t.Errorf("keep.txt should be staged; index = %q", index)
	}
	if strings.Contains(string(index), "skip.log") {
		t.Errorf("skip.log should be ignored; index = %q", index)
	}
}

func TestStatusRespectsGitignore(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile(".gitignore", []byte("ignored.txt\n"), 0644)
	os.WriteFile("ignored.txt", []byte("x"), 0644)
	os.WriteFile("visible.txt", []byte("y"), 0644)

	out := runCmd(t, statusCmd)
	if strings.Contains(out, "ignored.txt") {
		t.Errorf("ignored.txt should not appear in status; got %q", out)
	}
	if !strings.Contains(out, "visible.txt") {
		t.Errorf("visible.txt should appear in status; got %q", out)
	}
}
