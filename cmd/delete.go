package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/keychain"
	"github.com/julianofirme/envsecret/internal/project"
	"github.com/julianofirme/envsecret/internal/vault"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <KEY>",
	Short: "Remove a variable from the current project vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

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

		if _, ok := secrets[key]; !ok {
			return fmt.Errorf("[%s] key %q not found", proj, key)
		}

		delete(secrets, key)

		if err := vault.Save(vaultPath, masterKey, secrets); err != nil {
			return err
		}

		fmt.Printf("[%s] deleted %s\n", proj, key)
		return nil
	},
}
