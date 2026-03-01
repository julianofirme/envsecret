package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/keychain"
	"github.com/julianofirme/envsecret/internal/project"
	"github.com/julianofirme/envsecret/internal/vault"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show vault diagnostics for the current project without exposing any secrets",
	RunE: func(cmd *cobra.Command, args []string) error {
		proj, err := project.Resolve(projectFlag)
		if err != nil {
			return err
		}

		vaultPath, err := project.VaultPath(proj)
		if err != nil {
			return err
		}

		fmt.Printf("envs status — project: %s\n\n", proj)
		fmt.Printf("  Vault path:    %s\n", vaultPath)

		// Check vault existence
		info, statErr := os.Stat(vaultPath)
		if statErr != nil {
			fmt.Println("  Vault exists:  no")
			fmt.Println("  Keychain key:  (vault not initialized)")
			return nil
		}
		fmt.Println("  Vault exists:  yes")

		// Keychain
		keychainOK := keychain.Exists(proj)
		if keychainOK {
			fmt.Println("  Keychain key:  found")
		} else {
			fmt.Println("  Keychain key:  NOT FOUND")
		}

		// File permissions
		mode := info.Mode().Perm()
		modeStr := fmt.Sprintf("%o", mode)
		modeNote := ""
		if mode == 0600 {
			modeNote = " (ok)"
		} else {
			modeNote = " (WARNING: expected 600)"
		}
		fmt.Printf("  File mode:     %s%s\n", modeStr, modeNote)

		// Last modified
		fmt.Printf("  Last modified: %s\n", info.ModTime().UTC().Format(time.RFC3339))

		// Variable count (only if we can decrypt)
		if keychainOK {
			masterKey, err := keychain.Get(proj)
			if err == nil {
				secrets, err := vault.Load(vaultPath, masterKey)
				if err == nil {
					fmt.Printf("  Variables:     %d\n", len(secrets))
				}
			}
		}

		// Other projects
		baseDir, err := project.BaseDir()
		if err == nil {
			entries, err := os.ReadDir(baseDir)
			if err == nil {
				var others []string
				for _, e := range entries {
					if !e.IsDir() || e.Name() == proj {
						continue
					}
					vaultFile := filepath.Join(baseDir, e.Name(), "vault.enc")
					if _, err := os.Stat(vaultFile); err == nil {
						others = append(others, e.Name())
					}
				}
				sort.Strings(others)
				if len(others) > 0 {
					fmt.Printf("\n  Other projects: %v\n", others)
				}
			}
		}

		return nil
	},
}
