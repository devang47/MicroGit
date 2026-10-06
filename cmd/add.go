package cmd

import (
	"fmt"
	"microgit/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// updateIndex writes or updates the index file with path -> hash
func updateIndex(filePath, hash string) error {
	indexPath := filepath.Join(utils.DEFAULT_PATH, "index")
	existing := ""

	if data, err := os.ReadFile(indexPath); err == nil {
		existing = string(data)
	}

	lines := strings.Split(existing, "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, filePath+" ") {
			lines[i] = filePath + " " + hash
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, filePath+" "+hash)
	}

	newIndex := strings.Trim(strings.Join(lines, "\n"), "\n")
	return utils.WriteFileAtomic(indexPath, []byte(newIndex))
}

func stageFile(path, fileName string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file %q: %w", fileName, err)
	}

	// Calculate hash
	hash := utils.HashContent(content)

	// Write object
	if err := utils.WriteObject(hash, content); err != nil {
		return fmt.Errorf("writing object for %q: %w", fileName, err)
	}

	// Update index
	if err := updateIndex(path, hash); err != nil {
		return fmt.Errorf("updating index for %q: %w", fileName, err)
	}

	fmt.Printf("Added %s (hash: %s)\n", path, hash)

	return nil
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [files...]",
	Short: "Add files to the staging area",
	Long: `Add files to the staging area for the next commit.

Usage:
  microgit add <file1> [file2 ...]  - Stage specific files
  microgit add .                    - Stage all files in current directory

The add command will:
1. Calculate a SHA-256 hash of the file content
2. Store the file content in the objects directory
3. Update the index with the file path and corresponding hash

Files in the .microgit/ and .git/ directories are automatically ignored, as
are paths matched by a .gitignore file.`,
	Example: `  microgit add main.go utils.go
  microgit add .`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRepo(); err != nil {
			return err
		}

		if args[0] == "." {
			matcher := newIgnoreMatcher()
			return filepath.WalkDir(".", func(path string, file os.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if file.IsDir() {
					if matcher.match(path) {
						return filepath.SkipDir
					}
					return nil
				}

				if matcher.match(path) {
					return nil
				}

				return stageFile(path, file.Name())
			})
		}

		for _, file := range args {
			if err := stageFile(file, file); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
