// Package e2e drives the compiled microgit binary end-to-end as a subprocess,
// exercising the real CLI (argument parsing, exit codes, stdout/stderr) the way
// a user would from a shell.
package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// binPath is the compiled binary built once in TestMain.
var binPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "microgit-e2e-*")
	if err != nil {
		panic(err)
	}

	binName := "microgit"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath = filepath.Join(tmp, binName)
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = ".." // module root, relative to the e2e/ package directory
	if out, err := build.CombinedOutput(); err != nil {
		os.RemoveAll(tmp)
		panic("failed to build microgit binary: " + string(out))
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

type result struct {
	stdout string
	stderr string
	code   int
}

// run executes the binary with args in dir and captures its output/exit code.
func run(t *testing.T, dir string, args ...string) result {
	t.Helper()

	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	code := 0
	if err := cmd.Run(); err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("failed to run %v: %v", args, err)
		}
		code = exitErr.ExitCode()
	}

	return result{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// commitHashes returns commit hashes from `microgit log` output, newest first.
func commitHashes(logOut string) []string {
	var hashes []string
	for _, line := range strings.Split(logOut, "\n") {
		if hash, ok := strings.CutPrefix(line, "Commit: "); ok {
			hashes = append(hashes, strings.TrimSpace(hash))
		}
	}
	return hashes
}

func TestE2EInitAddSaveLog(t *testing.T) {
	dir := t.TempDir()

	if r := run(t, dir, "init"); !strings.Contains(r.stdout, "Initialized") {
		t.Fatalf("init stdout = %q", r.stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, ".microgit", "objects")); err != nil {
		t.Fatalf(".microgit/objects not created: %v", err)
	}

	writeFile(t, dir, "hello.txt", "hello")
	if r := run(t, dir, "add", "hello.txt"); !strings.Contains(r.stdout, "Added hello.txt") {
		t.Fatalf("add stdout = %q", r.stdout)
	}

	// Multi-word message passed as a single shell argument must be preserved.
	if r := run(t, dir, "save", "initial commit"); !strings.Contains(r.stdout, "Saved:") {
		t.Fatalf("save stdout = %q", r.stdout)
	}

	logOut := run(t, dir, "log").stdout
	if !strings.Contains(logOut, "initial commit") {
		t.Errorf("log missing message: %q", logOut)
	}
	if len(commitHashes(logOut)) != 1 {
		t.Errorf("expected 1 commit, got %d: %q", len(commitHashes(logOut)), logOut)
	}
}

func TestE2ECheckoutRestoresAndCleans(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "init")

	writeFile(t, dir, "a.txt", "A")
	run(t, dir, "add", "a.txt")
	run(t, dir, "save", "first")

	writeFile(t, dir, "b.txt", "B")
	run(t, dir, "add", ".")
	run(t, dir, "save", "second")

	hashes := commitHashes(run(t, dir, "log").stdout)
	if len(hashes) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(hashes))
	}
	first := hashes[len(hashes)-1]

	// Checkout the first commit: b.txt must be removed, a.txt kept.
	run(t, dir, "checkout", first)
	if _, err := os.Stat(filepath.Join(dir, "b.txt")); !os.IsNotExist(err) {
		t.Errorf("b.txt should be removed after checkout of first commit")
	}
	if _, err := os.Stat(filepath.Join(dir, "a.txt")); err != nil {
		t.Errorf("a.txt should remain after checkout of first commit: %v", err)
	}

	// Checkout latest: b.txt restored.
	run(t, dir, "checkout", "latest")
	if _, err := os.Stat(filepath.Join(dir, "b.txt")); err != nil {
		t.Errorf("b.txt should be restored after 'checkout latest': %v", err)
	}
}

func TestE2ERemoveExactMatch(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "init")

	writeFile(t, dir, "file.txt", "1")
	writeFile(t, dir, "foobar.txt", "2")
	run(t, dir, "add", "file.txt", "foobar.txt")

	run(t, dir, "remove", "file.txt")

	index, err := os.ReadFile(filepath.Join(dir, ".microgit", "index"))
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if !strings.Contains(string(index), "foobar.txt") {
		t.Errorf("foobar.txt should remain staged; index = %q", index)
	}
	if strings.Contains(string(index), "file.txt ") {
		t.Errorf("file.txt should be un-staged; index = %q", index)
	}
}

func TestE2ERepoGuard(t *testing.T) {
	dir := t.TempDir() // never initialized

	r := run(t, dir, "status")
	if r.code == 0 {
		t.Errorf("status outside a repo should exit non-zero; got code 0")
	}
	if !strings.Contains(r.stderr, "not a MicroGit repository") {
		t.Errorf("expected repo guard error on stderr, got stdout=%q stderr=%q", r.stdout, r.stderr)
	}
}

func TestE2ESaveRequiresMessage(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "init")

	// cobra's MinimumNArgs(1) should reject a message-less save at the CLI.
	r := run(t, dir, "save")
	if r.code == 0 {
		t.Errorf("save with no message should exit non-zero; got code 0, stdout=%q", r.stdout)
	}
	if !strings.Contains(r.stderr, "arg") {
		t.Errorf("expected an argument error on stderr, got %q", r.stderr)
	}
}

func TestE2ECheckoutForce(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "init")

	writeFile(t, dir, "a.txt", "A")
	run(t, dir, "add", "a.txt")
	run(t, dir, "save", "A")
	first := commitHashes(run(t, dir, "log").stdout)[0]

	writeFile(t, dir, "b.txt", "B")
	run(t, dir, "add", ".")
	run(t, dir, "save", "B")

	// Dirty the working tree.
	writeFile(t, dir, "a.txt", "A-modified")

	// Without --force the checkout must be refused.
	if r := run(t, dir, "checkout", first); r.code == 0 {
		t.Errorf("checkout on a dirty tree should fail without --force")
	}

	// With --force it succeeds.
	if r := run(t, dir, "checkout", "--force", first); r.code != 0 {
		t.Errorf("checkout --force should succeed, stderr=%q", r.stderr)
	}
}

func TestE2EDiff(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "init")

	writeFile(t, dir, "a.txt", "line1\nline2\n")
	run(t, dir, "add", "a.txt")
	run(t, dir, "save", "commit")

	writeFile(t, dir, "a.txt", "line1\nchanged\n")

	r := run(t, dir, "diff")
	if !strings.Contains(r.stdout, "- line2") || !strings.Contains(r.stdout, "+ changed") {
		t.Errorf("diff output missing expected changes: %q", r.stdout)
	}
}

func TestE2EVersion(t *testing.T) {
	dir := t.TempDir()

	r := run(t, dir, "--version")
	if !strings.Contains(r.stdout, "microgit version") {
		t.Errorf("--version should report a version line, got %q", r.stdout)
	}
}
