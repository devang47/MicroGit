package cmd

import (
	"fmt"
	"microgit/utils"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// checkoutForce, when set via --force, discards uncommitted changes on checkout.
var checkoutForce bool

// checkoutCmd represents the checkout command
var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Switch to a specific commit",
	Long: `Switch to a specific commit in the repository history.

Usage:
  microgit checkout <commit-hash>  - Switch to a specific commit
  microgit checkout latest        - Switch to the most recent commit

This command will:
1. Restore all files to their state at the specified commit
2. Update the HEAD reference to point to the checked out commit
3. Preserve the commit history for future operations`,
	Example: `  microgit checkout latest
  microgit checkout 3f8a1c2
  microgit checkout --force 3f8a1c2`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRepo(); err != nil {
			return err
		}

		savePointHash := args[0]

		if savePointHash == "latest" {
			savePointHash = getLatest()
		}
		if savePointHash == "" {
			return fmt.Errorf("no commits yet")
		}

		resolved, err := resolveHash(savePointHash)
		if err != nil {
			return fmt.Errorf("commit %s not found: %w", savePointHash, err)
		}
		savePointHash = resolved

		savePoint, err := readCommit(savePointHash)
		if err != nil {
			return fmt.Errorf("commit %s not found", savePointHash)
		}

		// Refuse to overwrite uncommitted work unless --force is given.
		if !checkoutForce {
			dirty, err := workingTreeDirty()
			if err != nil {
				return err
			}
			if dirty {
				return fmt.Errorf("you have uncommitted changes; commit them or use --force to discard")
			}
		}

		// Files tracked by the commit we're leaving that are absent from the
		// target must be removed so the working tree matches the target.
		current := getCommittedFiles()
		for path := range current {
			if _, ok := savePoint.Files[path]; !ok {
				_ = os.Remove(path)
			}
		}

		for path, hash := range savePoint.Files {
			content, err := utils.ReadObject(hash)
			if err != nil {
				return fmt.Errorf("missing object for file %s", path)
			}

			if err := os.WriteFile(path, content, 0644); err != nil {
				return fmt.Errorf("failed to restore file %s: %w", path, err)
			}
		}

		// Reset the staging area to match the checked-out commit so status
		// and a subsequent save operate from a consistent baseline.
		if err := writeIndex(savePoint.Files); err != nil {
			return fmt.Errorf("failed to reset staging area: %w", err)
		}

		headPath := filepath.Join(utils.DEFAULT_PATH, "HEAD")
		if err := utils.WriteFileAtomic(headPath, []byte(savePointHash)); err != nil {
			return fmt.Errorf("failed to update HEAD: %w", err)
		}

		fmt.Printf("Successfully checked out commit %s\n", savePointHash)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkoutCmd)
	checkoutCmd.Flags().BoolVarP(&checkoutForce, "force", "f", false,
		"Discard uncommitted changes when checking out")
}
