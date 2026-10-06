package cmd

import (
	"encoding/json"
	"fmt"
	"microgit/utils"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func getHead() string {
	headPath := filepath.Join(utils.DEFAULT_PATH, "HEAD")

	data, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func setHead(hash string) error {
	headPath := filepath.Join(utils.DEFAULT_PATH, "HEAD")
	latestPath := filepath.Join(utils.DEFAULT_PATH, "LATEST")

	if err := utils.WriteFileAtomic(headPath, []byte(hash)); err != nil {
		return fmt.Errorf("failed to write HEAD: %w", err)
	}
	if err := utils.WriteFileAtomic(latestPath, []byte(hash)); err != nil {
		return fmt.Errorf("failed to write LATEST: %w", err)
	}

	return nil
}

func readIndex() (map[string]string, error) {
	index := make(map[string]string)

	indexPath := filepath.Join(utils.DEFAULT_PATH, "index")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return index, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			index[parts[0]] = parts[1]
		}
	}
	return index, nil
}

func writeSavePointObject(savePoint utils.SavePoint) (string, error) {
	jsonData, err := json.MarshalIndent(savePoint, "", "  ")
	if err != nil {
		return "", err
	}

	// Hash of the entire SavePoint JSON
	hash := utils.HashContent(jsonData)

	err = utils.WriteObject(hash, jsonData)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// saveCmd represents the save command
var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save the current state of staged files",
	Long: `Save the current state of all staged files as a new commit.
This command requires a commit message that describes the changes being saved.
The staged files will be committed and the staging area will be cleared after the save.`,
	Example: `  microgit save "add login form"`,
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRepo(); err != nil {
			return err
		}

		// Join all args so unquoted multi-word messages aren't truncated.
		message := strings.Join(args, " ")

		index, err := readIndex()
		if err != nil {
			return fmt.Errorf("could not read index: %w", err)
		}

		if len(index) == 0 {
			return fmt.Errorf("no files have been added")
		}

		parent := getHead()

		// Warn when committing from a detached (non-tip) checkout: doing so
		// makes any commits after the current HEAD unreachable.
		if latest := getLatest(); parent != "" && latest != "" && parent != latest {
			fmt.Printf("Warning: committing from an older checkout (%s); commits after it will become unreachable.\n", parent)
		}

		timestamp := time.Now().Format(time.RFC3339)

		savePoint := utils.SavePoint{
			Message:   message,
			Timestamp: timestamp,
			Parent:    parent,
			Files:     index,
		}

		hash, err := writeSavePointObject(savePoint)
		if err != nil {
			return fmt.Errorf("failed to write commit: %w", err)
		}

		if err := setHead(hash); err != nil {
			return fmt.Errorf("failed to update HEAD: %w", err)
		}

		fmt.Printf("Saved: %s\n", hash)

		// Clear the staging area.
		indexPath := filepath.Join(utils.DEFAULT_PATH, "index")
		if err := utils.WriteFileAtomic(indexPath, []byte("")); err != nil {
			return fmt.Errorf("failed to clear staging area: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)
}
