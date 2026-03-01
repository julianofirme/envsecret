package cmd

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/keychain"
	"github.com/julianofirme/envsecret/internal/project"
	"github.com/julianofirme/envsecret/internal/vault"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a vault for the current project and store the master key in the OS keychain",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := project.Resolve(projectFlag)
		if err != nil {
			return err
		}

		vaultPath, err := project.VaultPath(proj)
		if err != nil {
			return err
		}

		// Check if vault already exists
		if _, err := os.Stat(vaultPath); err == nil {
			if !initForce {
				return fmt.Errorf("vault for project %q already exists. Use --force to reinitialize (DESTROYS all stored secrets)", proj)
			}
			fmt.Fprintf(os.Stderr, "WARNING: reinitializing vault for project %q — all existing secrets will be lost.\n", proj)
		}

		// Generate a new 256-bit master key
		masterKey := make([]byte, 32)
		if _, err := rand.Read(masterKey); err != nil {
			return fmt.Errorf("generate key: %w", err)
		}

		// Store key in keychain
		if err := keychain.Set(proj, masterKey); err != nil {
			return err
		}

		// Create vault directory
		vaultDir, err := project.VaultDir(proj)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(vaultDir, 0700); err != nil {
			return fmt.Errorf("create vault dir: %w", err)
		}

		// Write empty vault
		if err := vault.Save(vaultPath, masterKey, map[string]string{}); err != nil {
			return err
		}

		fmt.Printf("Initialized vault for project %q\n", proj)
		fmt.Printf("  Vault: %s\n", vaultPath)
		fmt.Println("  Master key stored in OS keychain.")
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "Reinitialize vault, destroying all existing secrets")
}
