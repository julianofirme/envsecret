package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/keychain"
	"github.com/julianofirme/envsecret/internal/project"
	"github.com/julianofirme/envsecret/internal/vault"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored key names for the current project (values are never shown)",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		keys := make([]string, 0, len(secrets))
		for k := range secrets {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		fmt.Fprintf(os.Stdout, "[%s] %d variable(s):\n", proj, len(keys))
		if len(keys) > 0 {
			fmt.Println()
			for _, k := range keys {
				fmt.Printf("  %s\n", k)
			}
		}
		return nil
	},
}
