package cmd

import (
	"encoding/json"
	"fmt"
	"microgit/utils"

	"github.com/spf13/cobra"
)

func readCommit(hash string) (utils.SavePoint, error) {
	data, err := utils.ReadObject(hash)
	if err != nil {
		return utils.SavePoint{}, fmt.Errorf("could not read commit object: %w", err)
	}

	var commit utils.SavePoint
	if err := json.Unmarshal(data, &commit); err != nil {
		return utils.SavePoint{}, fmt.Errorf("failed to parse commit JSON: %w", err)
	}
	return commit, nil
}

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show the commit history",
	Long: `Display the commit history in chronological order, starting from the most recent commit.
For each commit, it shows:
- The commit hash
- The timestamp
- The commit message
- The list of files that were modified`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureRepo(); err != nil {
			return err
		}

		head := getHead()
		if head == "" {
			fmt.Println("No commits yet.")
			return nil
		}

		current := head
		for current != "" {
			commit, err := readCommit(current)
			if err != nil {
				return fmt.Errorf("reading commit %s: %w", current, err)
			}

			fmt.Printf("Commit: %s\n", current)
			fmt.Printf("Date: %s\n", commit.Timestamp)
			fmt.Printf("Message: %s\n", commit.Message)
			fmt.Print("Files modified: ")
			for key := range commit.Files {
				fmt.Printf("%s ", key)
			}

			fmt.Print("\n\n")

			current = commit.Parent
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// logCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
