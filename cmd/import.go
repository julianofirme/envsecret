package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/keychain"
	"github.com/julianofirme/envsecret/internal/project"
	"github.com/julianofirme/envsecret/internal/vault"
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import KEY=VALUE pairs from a .env file or stdin into the current project vault",
	Long: `Reads KEY=VALUE pairs from a .env file (or stdin if no file is given) and
stores them in the current project vault. Existing keys are overwritten.

Blank lines and lines starting with # are ignored.
Surrounding quotes (single or double) are stripped from values.
The "export " prefix is also accepted.

Examples:
  envs import .env
  envs import < .env
  cat .env | envs import`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := ""
		if len(args) == 1 {
			filePath = args[0]
		}

		f, shouldClose, err := openEnvFile(filePath)
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		if shouldClose {
			defer func() { _ = f.Close() }()
		}

		pairs, err := readEnvLines(bufio.NewReader(f))
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		// Validate all keys before touching the vault
		for _, p := range pairs {
			if err := validateKey(p[0]); err != nil {
				return err
			}
		}

		if len(pairs) == 0 {
			fmt.Println("no variables found to import")
			return nil
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

		for _, p := range pairs {
			secrets[p[0]] = p[1]
		}

		if err := vault.Save(vaultPath, masterKey, secrets); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "[%s] imported %d variable(s)\n", proj, len(pairs))
		return nil
	},
}
