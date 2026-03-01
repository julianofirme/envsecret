package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/keychain"
	"github.com/julianofirme/envsecret/internal/project"
	"github.com/julianofirme/envsecret/internal/vault"
)

var copyCmd = &cobra.Command{
	Use:   "copy <SRC> <DST>",
	Short: "Copy a key to a new name within the current project vault",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcKey := args[0]
		dstKey := args[1]

		if err := validateKey(dstKey); err != nil {
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

		val, ok := secrets[srcKey]
		if !ok {
			return fmt.Errorf("[%s] key %q not found", proj, srcKey)
		}

		secrets[dstKey] = val

		if err := vault.Save(vaultPath, masterKey, secrets); err != nil {
			return err
		}

		fmt.Printf("[%s] copied %s → %s\n", proj, srcKey, dstKey)
		return nil
	},
}
