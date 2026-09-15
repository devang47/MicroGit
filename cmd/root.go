// Package cmd implements the MicroGit command-line interface: the cobra
// commands (init, add, remove, status, save, log, diff, checkout) and the
// shared repository helpers they rely on.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is the build version. It defaults to "dev" and is overridden at
// release time via -ldflags "-X microgit/cmd.version=<tag>".
var version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "microgit",
	Short:   "MicroGit - A simple version control system",
	Version: version,
	Long: `MicroGit is a simple version control system that provides basic Git-like functionality.
It allows you to track changes in your files and manage versions of your project.`,
	// Don't print the full usage text on runtime (RunE) errors — only on
	// argument/usage errors.
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to MicroGit! Use --help to see available commands.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Flags and configuration settings are registered on each subcommand's
	// own init(). The root command exposes --version (via the Version field)
	// and --help out of the box.
}
