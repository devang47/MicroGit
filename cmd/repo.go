package cmd

import (
	"fmt"
	"microgit/utils"
	"os"
	"path/filepath"
	"strings"
)

// ensureRepo returns an error if the current directory is not a MicroGit
// repository, so callers can propagate it and exit non-zero.
func ensureRepo() error {
	if !utils.IsInitialized() {
		return fmt.Errorf("not a MicroGit repository (run 'microgit init' first)")
	}
	return nil
}

// shouldIgnore reports whether a path is always excluded from staging and
// status: the internal repo directory, any (possibly nested) .git directory,
// and common OS cruft.
func shouldIgnore(path string) bool {
	for _, part := range strings.Split(path, string(filepath.Separator)) {
		switch part {
		case utils.DEFAULT_PATH, ".git", ".DS_Store":
			return true
		}
	}
	return false
}

// ignoreMatcher decides whether a path is ignored, combining the always-ignored
// built-ins (shouldIgnore) with patterns loaded from a .gitignore file.
type ignoreMatcher struct {
	patterns []string
}

// newIgnoreMatcher loads .gitignore from the current directory (if present).
// Blank lines and comments (#) are skipped.
func newIgnoreMatcher() *ignoreMatcher {
	m := &ignoreMatcher{}

	data, err := os.ReadFile(".gitignore")
	if err != nil {
		return m
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m.patterns = append(m.patterns, line)
	}
	return m
}

// match reports whether path should be ignored.
func (m *ignoreMatcher) match(path string) bool {
	if shouldIgnore(path) {
		return true
	}

	base := filepath.Base(path)
	for _, pat := range m.patterns {
		p := strings.TrimSuffix(pat, "/")
		p = strings.TrimPrefix(p, "/")

		if ok, _ := filepath.Match(p, base); ok {
			return true
		}
		if ok, _ := filepath.Match(p, path); ok {
			return true
		}
		// A bare directory pattern ("build") ignores everything under it.
		if strings.HasPrefix(path, p+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// resolveHash expands a (possibly abbreviated) object hash to a full hash by
// unique-prefix match, mirroring git's short-hash behavior.
func resolveHash(prefix string) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("empty commit reference")
	}

	// Fast path: an exact object file already exists (guard against a 2-char
	// prefix resolving to a shard directory).
	if info, err := os.Stat(utils.ObjectPath(prefix)); err == nil && !info.IsDir() {
		return prefix, nil
	}

	// Objects are sharded as objects/<xx>/<rest>; reconstruct each full hash
	// (shard + filename) and collect those that share the prefix.
	objectsDir := filepath.Join(utils.DEFAULT_PATH, "objects")
	shards, err := os.ReadDir(objectsDir)
	if err != nil {
		return "", err
	}

	var matches []string
	for _, shard := range shards {
		if !shard.IsDir() {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(objectsDir, shard.Name()))
		if err != nil {
			return "", err
		}
		for _, e := range entries {
			full := shard.Name() + e.Name()
			if strings.HasPrefix(full, prefix) {
				matches = append(matches, full)
			}
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no object matches %q", prefix)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("ambiguous prefix %q matches %d objects", prefix, len(matches))
	}
}

// getLatest returns the hash of the newest commit (the tip), or "" if none.
func getLatest() string {
	latestPath := filepath.Join(utils.DEFAULT_PATH, "LATEST")
	data, err := os.ReadFile(latestPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// writeIndex overwrites the staging area with the given path -> hash entries.
func writeIndex(index map[string]string) error {
	indexPath := filepath.Join(utils.DEFAULT_PATH, "index")

	lines := make([]string, 0, len(index))
	for path, hash := range index {
		lines = append(lines, path+" "+hash)
	}

	return utils.WriteFileAtomic(indexPath, []byte(strings.Join(lines, "\n")))
}
