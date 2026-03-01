package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/keychain"
	"github.com/julianofirme/envsecret/internal/project"
	"github.com/julianofirme/envsecret/internal/vault"
)

var setCmd = &cobra.Command{
	Use:   "set <KEY> <VALUE>",
	Short: "Encrypt and store a variable in the current project vault",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if err := validateKey(key); err != nil {
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

		secrets[key] = value

		if err := vault.Save(vaultPath, masterKey, secrets); err != nil {
			return err
		}

		fmt.Printf("[%s] set %s\n", proj, key)
		return nil
	},
}

// validateKey returns an error if k is not a valid environment variable name.
// Valid names match [A-Z_][A-Z0-9_]* (POSIX convention).
func validateKey(k string) error {
	if k == "" {
		return fmt.Errorf("key must not be empty")
	}
	for i, c := range k {
		switch {
		case c >= 'A' && c <= 'Z':
			// ok
		case c == '_':
			// ok
		case c >= '0' && c <= '9' && i > 0:
			// ok (not first char)
		default:
			return fmt.Errorf("invalid key %q: must match [A-Z_][A-Z0-9_]* (uppercase letters, digits, underscores only; cannot start with a digit)", k)
		}
	}
	return nil
}

// parseEnvLine parses a single .env line into (key, value, ok).
// Blank lines and lines starting with # are skipped (ok=false).
// Supports: KEY=VALUE, KEY="VALUE", KEY='VALUE', export KEY=VALUE.
func parseEnvLine(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	// strip optional "export " prefix
	line = strings.TrimPrefix(line, "export ")
	line = strings.TrimSpace(line)

	idx := strings.IndexByte(line, '=')
	if idx < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	val := line[idx+1:]

	// strip surrounding quotes (single or double)
	if len(val) >= 2 {
		if (val[0] == '"' && val[len(val)-1] == '"') ||
			(val[0] == '\'' && val[len(val)-1] == '\'') {
			val = val[1 : len(val)-1]
		}
	}

	return key, val, true
}

// readEnvLines reads KEY=VALUE pairs from r, returning them in order.
// Lines that cannot be parsed are silently skipped.
func readEnvLines(r *bufio.Reader) ([][2]string, error) {
	var pairs [][2]string
	for {
		line, err := r.ReadString('\n')
		if line != "" {
			if k, v, ok := parseEnvLine(line); ok {
				pairs = append(pairs, [2]string{k, v})
			}
		}
		if err != nil {
			break
		}
	}
	return pairs, nil
}

// openEnvFile opens path for reading, or returns os.Stdin if path is "-" or empty.
func openEnvFile(path string) (*os.File, bool, error) {
	if path == "" || path == "-" {
		return os.Stdin, false, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	return f, true, nil
}
