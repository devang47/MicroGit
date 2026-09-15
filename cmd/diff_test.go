package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestDiffLines(t *testing.T) {
	a := []string{"one", "two", "three"}
	b := []string{"one", "2", "three"}
	got := strings.Join(diffLines(a, b), "\n")

	want := "  one\n- two\n+ 2\n  three"
	if got != want {
		t.Errorf("diffLines mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestSplitLines(t *testing.T) {
	if got := splitLines(""); got != nil {
		t.Errorf("splitLines(\"\") = %v, want nil", got)
	}
	if got := splitLines("a\nb\n"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("splitLines(\"a\\nb\\n\") = %v, want [a b]", got)
	}
}

func TestDiffModifiedFile(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("line1\nline2\n"), 0644)
	runCmd(t, addCmd, "a.txt")
	runCmd(t, saveCmd, "commit")

	os.WriteFile("a.txt", []byte("line1\nCHANGED\n"), 0644)

	out := runCmd(t, diffCmd)
	if !strings.Contains(out, "- line2") {
		t.Errorf("diff should show removed 'line2'; got %q", out)
	}
	if !strings.Contains(out, "+ CHANGED") {
		t.Errorf("diff should show added 'CHANGED'; got %q", out)
	}
}

func TestDiffNewFile(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("hello\n"), 0644)
	runCmd(t, addCmd, "a.txt")
	runCmd(t, saveCmd, "commit")

	// Untracked new file appears as an addition against /dev/null.
	os.WriteFile("b.txt", []byte("brand new\n"), 0644)

	out := runCmd(t, diffCmd)
	if !strings.Contains(out, "--- /dev/null") {
		t.Errorf("new file should diff against /dev/null; got %q", out)
	}
	if !strings.Contains(out, "+ brand new") {
		t.Errorf("new file content should be added; got %q", out)
	}
}

func TestDiffNoChanges(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("stable\n"), 0644)
	runCmd(t, addCmd, "a.txt")
	runCmd(t, saveCmd, "commit")

	out := runCmd(t, diffCmd)
	if !strings.Contains(out, "No changes.") {
		t.Errorf("expected 'No changes.'; got %q", out)
	}
}
