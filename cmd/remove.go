package cmd

import (
	"fmt"
	"microgit/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove files from the staging area",
	Long: `Remove files from the staging area, effectively un-staging them.

Usage:
  microgit remove <file1> [file2 ...]  - Remove specific files from staging
  microgit remove .                    - Remove all files from staging

This command will:
1. Remove the specified files from the index
2. Keep the files in your working directory
3. Allow you to re-stage them later if needed`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRepo(); err != nil {
			return err
		}

		indexPath := filepath.Join(utils.DEFAULT_PATH, "index")

		if args[0] == "." {
			if err := utils.WriteFileAtomic(indexPath, []byte("")); err != nil {
				return fmt.Errorf("unstage all files: %w", err)
			}
			return nil
		}

		for _, file := range args {
			data, err := os.ReadFile(indexPath)
			if err != nil {
				return fmt.Errorf("read index: %w", err)
			}

			lines := strings.Split(string(data), "\n")

			var stagedFiles []string
			for _, line := range lines {
				if line == "" {
					continue
				}
				// The index stores "path hash" (or just "path"); match on the
				// exact path field so 'remove foo' doesn't drop 'foobar'.
				path := strings.SplitN(line, " ", 2)[0]
				if path != file {
					stagedFiles = append(stagedFiles, line)
				}
			}

			newIndex := strings.Trim(strings.Join(stagedFiles, "\n"), "\n")
			if err := utils.WriteFileAtomic(indexPath, []byte(newIndex)); err != nil {
				return fmt.Errorf("update index: %w", err)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
