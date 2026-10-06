package cmd

import (
	"os"
	"strings"
	"testing"
)

// TestCheckoutRemovesStaleFilesAndResetsIndex covers two fixes: checking out an
// older commit must delete files not present in it and reset the staging area.
func TestCheckoutRemovesStaleFilesAndResetsIndex(t *testing.T) {
	setupTestRepo(t)

	// Commit A: a.txt only.
	os.WriteFile("a.txt", []byte("A"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"a.txt"}) })
	captureOutput(func() { saveCmd.RunE(nil, []string{"A"}) })
	commitA := getHead()

	// Commit B: a.txt + b.txt.
	os.WriteFile("b.txt", []byte("B"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"a.txt", "b.txt"}) })
	captureOutput(func() { saveCmd.RunE(nil, []string{"B"}) })

	captureOutput(func() { checkoutCmd.RunE(nil, []string{commitA}) })

	if _, err := os.Stat("b.txt"); !os.IsNotExist(err) {
		t.Errorf("b.txt should be removed after checking out A")
	}
	if _, err := os.Stat("a.txt"); err != nil {
		t.Errorf("a.txt should still exist after checking out A: %v", err)
	}
	if getHead() != commitA {
		t.Errorf("HEAD = %q, want %q", getHead(), commitA)
	}

	idx, _ := readIndex()
	if _, ok := idx["b.txt"]; ok {
		t.Errorf("index should not contain b.txt after checkout A, got %v", idx)
	}
	if _, ok := idx["a.txt"]; !ok {
		t.Errorf("index should contain a.txt after checkout A, got %v", idx)
	}
}

func TestCheckoutLatestEmpty(t *testing.T) {
	setupTestRepo(t)

	_, err := runCmdErr(t, checkoutCmd, "latest")
	if err == nil || !strings.Contains(err.Error(), "no commits yet") {
		t.Errorf("expected 'no commits yet' error, got %v", err)
	}
}

func TestCheckoutNonexistent(t *testing.T) {
	setupTestRepo(t)

	_, err := runCmdErr(t, checkoutCmd, "deadbeef")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got %v", err)
	}
}

func TestCheckoutRefusesDirtyTree(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("A"), 0644)
	runCmd(t, addCmd, "a.txt")
	runCmd(t, saveCmd, "A")
	first := getHead()

	os.WriteFile("b.txt", []byte("B"), 0644)
	runCmd(t, addCmd, ".")
	runCmd(t, saveCmd, "B")

	// Dirty the tree: modify a committed file without staging.
	os.WriteFile("a.txt", []byte("A-modified"), 0644)

	_, err := runCmdErr(t, checkoutCmd, first)
	if err == nil || !strings.Contains(err.Error(), "uncommitted changes") {
		t.Errorf("expected refusal on dirty tree, got %v", err)
	}
}

func TestCheckoutForceDiscardsChanges(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("A"), 0644)
	runCmd(t, addCmd, "a.txt")
	runCmd(t, saveCmd, "A")
	first := getHead()

	os.WriteFile("b.txt", []byte("B"), 0644)
	runCmd(t, addCmd, ".")
	runCmd(t, saveCmd, "B")

	os.WriteFile("a.txt", []byte("A-modified"), 0644)

	checkoutForce = true
	defer func() { checkoutForce = false }()
	runCmd(t, checkoutCmd, first)

	if getHead() != first {
		t.Errorf("checkout --force should switch HEAD to %q, got %q", first, getHead())
	}
	if data, _ := os.ReadFile("a.txt"); string(data) != "A" {
		t.Errorf("a.txt should be restored to 'A', got %q", data)
	}
}

func TestCheckoutLatestRestoresTip(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("A"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"a.txt"}) })
	captureOutput(func() { saveCmd.RunE(nil, []string{"A"}) })

	os.WriteFile("b.txt", []byte("B"), 0644)
	captureOutput(func() { addCmd.RunE(nil, []string{"a.txt", "b.txt"}) })
	captureOutput(func() { saveCmd.RunE(nil, []string{"B"}) })
	tip := getHead()

	commit, _ := readCommit(tip)
	commitA := commit.Parent

	captureOutput(func() { checkoutCmd.RunE(nil, []string{commitA}) })
	captureOutput(func() { checkoutCmd.RunE(nil, []string{"latest"}) })

	if _, err := os.Stat("b.txt"); err != nil {
		t.Errorf("b.txt should be restored after 'checkout latest': %v", err)
	}
	if getHead() != tip {
		t.Errorf("HEAD = %q, want tip %q after 'checkout latest'", getHead(), tip)
	}
}
