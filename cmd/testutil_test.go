package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// runCmd invokes a command's RunE, captures stdout, and fails the test on any
// unexpected error. Use it for the happy path.
func runCmd(t *testing.T, cmd *cobra.Command, args ...string) string {
	t.Helper()
	var err error
	out := captureOutput(func() { err = cmd.RunE(nil, args) })
	if err != nil {
		t.Fatalf("%s %v: unexpected error: %v", cmd.Name(), args, err)
	}
	return out
}

// runCmdErr invokes a command's RunE and returns its stdout and error, for
// negative-path tests.
func runCmdErr(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var err error
	out := captureOutput(func() { err = cmd.RunE(nil, args) })
	return out, err
}

// setupTestRepo creates a temporary directory, changes into it, initializes a
// MicroGit repository, and registers cleanup to restore the working directory.
func setupTestRepo(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	t.Cleanup(func() { os.Chdir(oldDir) })

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	runCmd(t, initCmd)
}

// captureOutput runs f and returns everything it wrote to os.Stdout.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// statusSection returns the lines printed under a given "=== header ===" block
// of `microgit status` output.
func statusSection(out, header string) string {
	target := "=== " + header + " ==="

	var b strings.Builder
	capturing := false
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "===") {
			capturing = trimmed == target
			continue
		}
		if capturing {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return b.String()
}
