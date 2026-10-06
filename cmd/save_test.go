package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestSaveCreatesCommit(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("hello"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"a.txt"}) })
	captureOutput(func() { saveCmd.RunE(nil, []string{"first", "commit"}) })

	head := getHead()
	if head == "" {
		t.Fatal("HEAD should be set after save")
	}
	if getLatest() != head {
		t.Errorf("LATEST %q should equal HEAD %q", getLatest(), head)
	}

	// The staging area should be cleared after a save.
	idx, _ := readIndex()
	if len(idx) != 0 {
		t.Errorf("index should be cleared after save, got %v", idx)
	}

	commit, err := readCommit(head)
	if err != nil {
		t.Fatalf("readCommit: %v", err)
	}
	// Multi-word message must be preserved (not truncated to "first").
	if commit.Message != "first commit" {
		t.Errorf("message = %q, want %q", commit.Message, "first commit")
	}
	if commit.Parent != "" {
		t.Errorf("parent = %q, want empty", commit.Parent)
	}
	if _, ok := commit.Files["a.txt"]; !ok {
		t.Errorf("commit should contain a.txt, got %v", commit.Files)
	}
}

func TestSaveParentChain(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("v1"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"a.txt"}) })
	captureOutput(func() { saveCmd.RunE(nil, []string{"first"}) })
	first := getHead()

	os.WriteFile("b.txt", []byte("v2"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"b.txt"}) })
	captureOutput(func() { saveCmd.RunE(nil, []string{"second"}) })
	second := getHead()

	if second == first {
		t.Fatal("second commit should differ from first")
	}
	commit, _ := readCommit(second)
	if commit.Parent != first {
		t.Errorf("parent = %q, want %q", commit.Parent, first)
	}
}

func TestSaveNoStagedFiles(t *testing.T) {
	setupTestRepo(t)

	_, err := runCmdErr(t, saveCmd, "nothing")
	if err == nil || !strings.Contains(err.Error(), "no files have been added") {
		t.Errorf("expected 'no files have been added' error, got %v", err)
	}
	if getHead() != "" {
		t.Errorf("HEAD should remain empty, got %q", getHead())
	}
}

func TestSaveWarnsOnNonTipCommit(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("A"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"a.txt"}) })
	captureOutput(func() { saveCmd.RunE(nil, []string{"A"}) })
	first := getHead()

	os.WriteFile("b.txt", []byte("B"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"b.txt"}) })
	captureOutput(func() { saveCmd.RunE(nil, []string{"B"}) })

	// Detach to the older commit, then commit again from there.
	captureOutput(func() { checkoutCmd.RunE(nil, []string{first}) })
	os.WriteFile("c.txt", []byte("C"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"c.txt"}) })
	out := captureOutput(func() { saveCmd.RunE(nil, []string{"C"}) })

	if !strings.Contains(out, "Warning") {
		t.Errorf("expected a non-tip commit warning, got %q", out)
	}
}
