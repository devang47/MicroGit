package cmd

import (
	"fmt"
	"microgit/utils"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func getWorkingFiles() (map[string]string, error) {
	files := make(map[string]string)

	matcher := newIgnoreMatcher()
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if matcher.match(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if matcher.match(path) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		files[path] = utils.HashContent(content)
		return nil
	})

	return files, err
}

func getCommittedFiles() map[string]string {
	head := getHead()
	if head == "" {
		return map[string]string{}
	}

	commit, err := readCommit(head)
	if err != nil {
		return map[string]string{}
	}
	return commit.Files
}

func getStatusData() (map[string]string, map[string]string, map[string]string, error) {
	index, err := readIndex()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read index: %w", err)
	}

	committed := getCommittedFiles()

	working, err := getWorkingFiles()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get working files: %w", err)
	}

	return index, committed, working, nil
}

// workingTreeDirty reports whether the repository has staged changes or
// unsaved working-tree modifications relative to the current commit.
func workingTreeDirty() (bool, error) {
	index, committed, working, err := getStatusData()
	if err != nil {
		return false, err
	}

	// Staged changes not yet committed.
	for path, h := range index {
		if committed[path] != h {
			return true, nil
		}
	}
	// Tracked files modified in the working tree.
	for path, wh := range working {
		base, tracked := index[path]
		if !tracked {
			base, tracked = committed[path]
		}
		if tracked && base != wh {
			return true, nil
		}
	}
	// Tracked files deleted from the working tree.
	for path := range committed {
		if _, ok := working[path]; !ok {
			return true, nil
		}
	}
	return false, nil
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the working tree status",
	Long: `Display the state of the working directory and the staging area.
Shows which files have been staged for the next commit and which files
are untracked. This helps you understand what will be included in your
next commit.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRepo(); err != nil {
			return err
		}

		index, committed, working, err := getStatusData()
		if err != nil {
			return err
		}

		fmt.Println("=== Staged ===")
		for path, hash := range index {
			if committed[path] != hash {
				fmt.Println(path)
			}
		}

		fmt.Println("\n=== Modified but not Staged ===")
		for path, workingHash := range working {
			// Compare the working copy against its baseline: the staged
			// version if present, otherwise the last committed version.
			baseline, tracked := index[path]
			if !tracked {
				baseline, tracked = committed[path]
			}
			if tracked && baseline != workingHash {
				fmt.Println(path)
			}
		}

		fmt.Println("\n=== Untracked Files ===")
		for path := range working {
			_, inIndex := index[path]
			_, inCommit := committed[path]
			if !inIndex && !inCommit {
				fmt.Println(path)
			}
		}

		fmt.Println("\n=== Deleted ===")
		seen := map[string]bool{}

		// Deleted files that were in the last commit
		for path := range committed {
			if _, ok := working[path]; !ok {
				fmt.Println(path + " (was saved)")
				seen[path] = true
			}
		}

		// Deleted files that were staged
		for path := range index {
			if _, ok := working[path]; !ok && !seen[path] {
				fmt.Println(path + " (was staged)")
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
