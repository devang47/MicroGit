package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestStatusStaged(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("hi"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"a.txt"}) })

	out := captureOutput(func() { statusCmd.RunE(nil, nil) })
	if !strings.Contains(statusSection(out, "Staged"), "a.txt") {
		t.Errorf("a.txt should be staged; got %q", out)
	}
}

// TestStatusModifiedTracked covers the fix: a committed file edited in the
// working tree (without re-adding) must be reported as modified.
func TestStatusModifiedTracked(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("v1"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"a.txt"}) })
	captureOutput(func() { saveCmd.RunE(nil, []string{"commit"}) })

	// Modify the committed file but do NOT re-stage it.
	os.WriteFile("a.txt", []byte("v2"), 0644)

	out := captureOutput(func() { statusCmd.RunE(nil, nil) })
	if !strings.Contains(statusSection(out, "Modified but not Staged"), "a.txt") {
		t.Errorf("modified committed file should appear as modified; got %q", out)
	}
	if strings.Contains(statusSection(out, "Untracked Files"), "a.txt") {
		t.Errorf("committed file should not be reported as untracked; got %q", out)
	}
}

func TestStatusUntracked(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("new.txt", []byte("x"), 0644)
	out := captureOutput(func() { statusCmd.RunE(nil, nil) })
	if !strings.Contains(statusSection(out, "Untracked Files"), "new.txt") {
		t.Errorf("new.txt should be untracked; got %q", out)
	}
}

func TestStatusIgnoresInternalDirs(t *testing.T) {
	setupTestRepo(t)

	os.MkdirAll(".git", 0755)
	os.WriteFile(".git/config", []byte("x"), 0644)
	os.WriteFile(".DS_Store", []byte("x"), 0644)
	os.WriteFile("real.txt", []byte("x"), 0644)

	out := captureOutput(func() { statusCmd.RunE(nil, nil) })
	if strings.Contains(out, ".git") {
		t.Errorf(".git contents should be ignored; got %q", out)
	}
	if strings.Contains(out, ".DS_Store") {
		t.Errorf(".DS_Store should be ignored; got %q", out)
	}
	if !strings.Contains(out, "real.txt") {
		t.Errorf("real.txt should appear; got %q", out)
	}
}
