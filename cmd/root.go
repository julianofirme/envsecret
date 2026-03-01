package cmd

import (
	"github.com/spf13/cobra"
)

var projectFlag string

var rootCmd = &cobra.Command{
	Use:   "envs",
	Short: "Encrypted environment variable vault for macOS and Linux",
	Long: `envs — per-project encrypted secret storage.

Secrets are stored in AES-256-GCM encrypted files. Each project gets its own
vault and its own master key in the OS keychain. Secrets are injected
exclusively into child process environments.`,
}

// Execute is the entry point called from main.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&projectFlag, "project", "", "Override the project name")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(destroyCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(renameCmd)
	rootCmd.AddCommand(copyCmd)
}
