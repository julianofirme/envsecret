package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/julianofirme/envsecret/internal/keychain"
	"github.com/julianofirme/envsecret/internal/project"
	"github.com/julianofirme/envsecret/internal/vault"
)

var runCmd = &cobra.Command{
	Use:   "run [--clean] -- <command> [args...]",
	Short: "Spawn a child process with vault secrets injected into its environment",
	Long: `Spawns a child process with vault secrets injected into its environment.
The parent shell is never modified.

The -- separator is required. Everything after it is the command to run.

--clean strips the parent shell environment entirely. Only vault vars and PATH
are passed to the child.`,
	// DisableFlagParsing lets us receive "--clean" and "--" ourselves,
	// so cobra never swallows the "--" separator before we can see it.
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Manual flag parsing: scan for --clean and --, everything after -- is
		// the child command. --project is a persistent root flag so we handle
		// it here too when DisableFlagParsing is active.
		clean := false
		dashDash := -1
		for i, a := range args {
			switch {
			case a == "--clean":
				clean = true
			case a == "--":
				dashDash = i
			case strings.HasPrefix(a, "--project="):
				projectFlag = strings.TrimPrefix(a, "--project=")
			case a == "--project" && i+1 < len(args):
				projectFlag = args[i+1]
			}
			if dashDash >= 0 {
				break
			}
		}

		if dashDash < 0 {
			return fmt.Errorf("missing -- separator — use: envs run [--clean] -- <command> [args...]")
		}

		childArgs := args[dashDash+1:]
		if len(childArgs) == 0 {
			return fmt.Errorf("no command specified after -- — use: envs run [--clean] -- <command> [args...]")
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

		if len(secrets) == 0 {
			fmt.Fprintf(os.Stderr, "[%s] warning: vault is empty — no secrets injected\n", proj)
		}

		// Build the child environment
		var childEnv []string
		if clean {
			// Only include PATH from parent
			for _, e := range os.Environ() {
				if len(e) >= 5 && e[:5] == "PATH=" {
					childEnv = append(childEnv, e)
					break
				}
			}
		} else {
			childEnv = os.Environ()
		}

		// Inject vault secrets (overrides any existing env vars with same name)
		for k, v := range secrets {
			childEnv = append(childEnv, k+"="+v)
		}

		binary, err := exec.LookPath(childArgs[0])
		if err != nil {
			return fmt.Errorf("command not found: %s", childArgs[0])
		}

		return syscall.Exec(binary, childArgs, childEnv)
	},
}
