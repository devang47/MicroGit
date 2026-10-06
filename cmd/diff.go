package cmd

import (
	"fmt"
	"microgit/utils"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// diffLabel builds the ---/+++ header path, using /dev/null for absent sides.
func diffLabel(path string, exists bool, prefix string) string {
	if !exists {
		return "/dev/null"
	}
	return prefix + "/" + path
}

// splitLines splits text into lines, dropping the trailing empty element that
// a final newline would otherwise produce.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// diffLines returns a line-level diff of a -> b using a longest-common-
// subsequence walk. Unchanged lines are prefixed with "  ", removals with
// "- " and additions with "+ ".
func diffLines(a, b []string) []string {
	n, m := len(a), len(b)

	// lcs[i][j] = length of the LCS of a[i:] and b[j:].
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	var out []string
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			out = append(out, "  "+a[i])
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, "- "+a[i])
			i++
		default:
			out = append(out, "+ "+b[j])
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, "- "+a[i])
	}
	for ; j < m; j++ {
		out = append(out, "+ "+b[j])
	}
	return out
}

// diffCmd represents the diff command.
var diffCmd = &cobra.Command{
	Use:   "diff [commit]",
	Short: "Show changes between the working tree and a commit",
	Long: `Show line-level differences between the files in your working tree and a
commit. With no argument the current commit (HEAD) is used; you may also pass a
commit hash or 'latest'.`,
	Example: `  microgit diff
  microgit diff latest
  microgit diff 3f8a1c2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRepo(); err != nil {
			return err
		}

		baseHash := getHead()
		if len(args) > 0 {
			baseHash = args[0]
			if baseHash == "latest" {
				baseHash = getLatest()
			}
		}

		baseFiles := map[string]string{}
		if baseHash != "" {
			resolved, err := resolveHash(baseHash)
			if err != nil {
				return fmt.Errorf("commit %s not found", baseHash)
			}

			commit, err := readCommit(resolved)
			if err != nil {
				return fmt.Errorf("commit %s not found", resolved)
			}
			baseFiles = commit.Files
		}

		working, err := getWorkingFiles()
		if err != nil {
			return err
		}

		// Union of paths on both sides, in stable sorted order.
		paths := map[string]bool{}
		for p := range baseFiles {
			paths[p] = true
		}
		for p := range working {
			paths[p] = true
		}
		sorted := make([]string, 0, len(paths))
		for p := range paths {
			sorted = append(sorted, p)
		}
		sort.Strings(sorted)

		changed := false
		for _, path := range sorted {
			baseBlob, inBase := baseFiles[path]
			workHash, inWork := working[path]

			if inBase && inWork && baseBlob == workHash {
				continue // unchanged
			}
			changed = true

			var oldContent, newContent string
			if inBase {
				if data, err := utils.ReadObject(baseBlob); err == nil {
					oldContent = string(data)
				}
			}
			if inWork {
				if data, err := os.ReadFile(path); err == nil {
					newContent = string(data)
				}
			}

			fmt.Printf("--- %s\n", diffLabel(path, inBase, "a"))
			fmt.Printf("+++ %s\n", diffLabel(path, inWork, "b"))
			for _, line := range diffLines(splitLines(oldContent), splitLines(newContent)) {
				fmt.Println(line)
			}
		}

		if !changed {
			fmt.Println("No changes.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
