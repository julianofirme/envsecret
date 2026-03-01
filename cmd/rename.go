package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/keychain"
	"github.com/julianofirme/envsecret/internal/project"
	"github.com/julianofirme/envsecret/internal/vault"
)

var renameCmd = &cobra.Command{
	Use:   "rename <OLD> <NEW>",
	Short: "Rename a key in the current project vault",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldKey := args[0]
		newKey := args[1]

		if err := validateKey(newKey); err != nil {
			return err
		}

		proj, err := project.Resolve(projectFlag)
		if err != nil {
			return err
		}

		masterKey, err := keychain.Get(proj)
		if err != nil {
			return fmt.Errorf("no vault found for project %q — run `envs init` first", proj)
		}

		vaultPath, err := project.VaultPath(proj)
		if err != nil {
			return err
		}

		secrets, err := vault.Load(vaultPath, masterKey)
		if err != nil {
			return err
		}

		val, ok := secrets[oldKey]
		if !ok {
			return fmt.Errorf("[%s] key %q not found", proj, oldKey)
		}

		secrets[newKey] = val
		delete(secrets, oldKey)

		if err := vault.Save(vaultPath, masterKey, secrets); err != nil {
			return err
		}

		fmt.Printf("[%s] renamed %s → %s\n", proj, oldKey, newKey)
		return nil
	},
}
