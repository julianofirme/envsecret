package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/keychain"
	"github.com/julianofirme/envsecret/internal/project"
	"github.com/julianofirme/envsecret/internal/vault"
)

var getCmd = &cobra.Command{
	Use:   "get <KEY>",
	Short: "Print a single decrypted value to stdout",
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

		val, ok := secrets[key]
		if !ok {
			return fmt.Errorf("[%s] key %q not found", proj, key)
		}

		fmt.Println(val)
		return nil
	},
}
