package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"microgit/utils"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new MicroGit repository",
	Long: `Initialize a new MicroGit repository in the current directory.
This creates the necessary directory structure and files for version control.
The repository will be initialized in a .microgit directory.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		repoDir := utils.DEFAULT_PATH
		objectsDir := filepath.Join(repoDir, "objects")

		if utils.IsInitialized() {
			fmt.Println("Repository already initialized.")
			return nil
		}

		if err := os.Mkdir(repoDir, 0755); err != nil {
			return fmt.Errorf("create repository directory: %w", err)
		}
		if err := os.Mkdir(objectsDir, 0755); err != nil {
			return fmt.Errorf("create objects directory: %w", err)
		}

		// index is the staging area; HEAD points at the current commit;
		// LATEST points at the newest commit (the tip).
		for _, name := range []string{"index", "HEAD", "LATEST"} {
			if err := os.WriteFile(filepath.Join(repoDir, name), []byte(""), 0644); err != nil {
				return fmt.Errorf("create %s file: %w", name, err)
			}
		}

		fmt.Printf("Initialized empty SCM repository in %s/\n", utils.DEFAULT_PATH)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
