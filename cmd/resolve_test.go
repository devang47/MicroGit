package cmd

import (
	"os"
	"strings"
	"testing"

	"microgit/utils"
)

func TestResolveHashPrefix(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("x"), 0644)
	runCmd(t, addCmd, "a.txt")
	runCmd(t, saveCmd, "commit")
	full := getHead()

	got, err := resolveHash(full[:8])
	if err != nil {
		t.Fatalf("resolveHash(prefix): %v", err)
	}
	if got != full {
		t.Errorf("resolveHash(%q) = %q, want %q", full[:8], got, full)
	}
}

func TestResolveHashAmbiguous(t *testing.T) {
	setupTestRepo(t)

	if err := utils.WriteObject("abcd0000", []byte("x")); err != nil {
		t.Fatalf("write object: %v", err)
	}
	if err := utils.WriteObject("abce1111", []byte("y")); err != nil {
		t.Fatalf("write object: %v", err)
	}

	if _, err := resolveHash("ab"); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("expected ambiguous error, got %v", err)
	}
}

func TestResolveHashNoMatch(t *testing.T) {
	setupTestRepo(t)

	if _, err := resolveHash("zzzz"); err == nil {
		t.Error("expected error for non-matching prefix")
	}
}

func TestCheckoutByShortHash(t *testing.T) {
	setupTestRepo(t)

	os.WriteFile("a.txt", []byte("A"), 0644)
	runCmd(t, addCmd, "a.txt")
	runCmd(t, saveCmd, "A")
	first := getHead()

	os.WriteFile("b.txt", []byte("B"), 0644)
	runCmd(t, addCmd, ".")
	runCmd(t, saveCmd, "B")

	// Checkout the first commit using only a short prefix.
	runCmd(t, checkoutCmd, first[:8])
	if getHead() != first {
		t.Errorf("HEAD = %q, want %q after short-hash checkout", getHead(), first)
	}
}
