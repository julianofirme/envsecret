package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/keychain"
	"github.com/julianofirme/envsecret/internal/project"
)

var destroyForce bool

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Permanently delete the current project vault and remove its key from the OS keychain",
	Long: `Removes the vault file and the master key from the OS keychain for the current
project. This operation is irreversible — all stored secrets are lost.

A confirmation prompt is shown unless --yes is passed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := project.Resolve(projectFlag)
		if err != nil {
			return err
		}

		vaultPath, err := project.VaultPath(proj)
		if err != nil {
			return err
		}

		if !destroyForce {
			fmt.Fprintf(os.Stderr, "WARNING: this will permanently delete the vault and keychain key for project %q.\n", proj)
			fmt.Fprintf(os.Stderr, "All stored secrets will be lost and cannot be recovered.\n\n")
			fmt.Fprintf(os.Stderr, "Type the project name to confirm: ")

			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("read confirmation: %w", err)
			}
			input = strings.TrimSpace(input)
			if input != proj {
				return fmt.Errorf("confirmation did not match — aborting")
			}
		}

		// Remove vault file
		vaultErr := os.Remove(vaultPath)
		if vaultErr != nil && !os.IsNotExist(vaultErr) {
			return fmt.Errorf("remove vault: %w", vaultErr)
		}

		// Remove vault directory if empty
		vaultDir, err := project.VaultDir(proj)
		if err == nil {
			_ = os.Remove(vaultDir) // ignore error — directory may not be empty
		}

		// Remove key from keychain
		keychainErr := keychain.Delete(proj)
		if keychainErr != nil {
			fmt.Fprintf(os.Stderr, "[%s] warning: could not remove keychain key: %v\n", proj, keychainErr)
		}

		fmt.Printf("[%s] destroyed — vault and keychain key removed\n", proj)
		return nil
	},
}

func init() {
	destroyCmd.Flags().BoolVar(&destroyForce, "yes", false, "Skip confirmation prompt")
}
