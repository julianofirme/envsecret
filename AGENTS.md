# AGENTS.md — envsecret

Guidance for agentic coding assistants working in this repository.

---

## Project overview

`envsecret` is a CLI tool (invoked as `envs`) that stores secrets in
AES-256-GCM encrypted vault files, one per project, with master keys in the
OS keychain (macOS Keychain / libsecret on Linux).

```
envsecret/
├── main.go                   # entry point — calls cmd.Execute()
├── cmd/                      # one file per cobra subcommand
│   ├── root.go               # rootCmd, --project flag, subcommand wiring
│   ├── init.go               # envs init [--force]
│   ├── set.go                # envs set <KEY> <VALUE>
│   ├── get.go                # envs get <KEY>
│   ├── list.go               # envs list
│   ├── delete.go             # envs delete <KEY>
│   ├── run.go                # envs run [--clean] -- <cmd>
│   ├── projects.go           # envs projects
│   └── status.go             # envs status
└── internal/
    ├── keychain/keychain.go  # OS keychain wrapper (go-keyring)
    ├── vault/vault.go        # AES-256-GCM encrypt/decrypt + atomic write
    └── project/project.go   # project name resolution + path helpers
```

---

## Build commands

```bash
# Build the binary
go build -o envs .

# Install to PATH
go build -o envs . && sudo mv envs /usr/local/bin/

# Verify it compiles (no output = success)
go build ./...

# Vet
go vet ./...

# Format (writes in place)
gofmt -w .
# or
goimports -w .
```

---

## Testing

There are currently no test files. When adding tests:

```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./internal/vault/...
go test ./internal/project/...

# Run a single test by name
go test ./internal/vault/ -run TestEncryptDecrypt

# Run with verbose output
go test -v ./...

# Run with race detector
go test -race ./...
```

Test files live next to the code they test (`foo_test.go` beside `foo.go`),
using `package foo` (white-box) or `package foo_test` (black-box) as
appropriate. Prefer black-box tests for `internal/vault` and
`internal/keychain` since their public surface is small and well-defined.

---

## Code style

### Formatting

- All code must be formatted with `gofmt`. No exceptions.
- Import grouping (enforced by `goimports`):
  1. Standard library
  2. _(blank line)_
  3. Third-party modules
  4. _(blank line)_
  5. Internal packages (`github.com/julianofirme/envsecret/internal/...`)
- Use a blank line between each import group. Never merge groups.

```go
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "github.com/julianofirme/envsecret/internal/keychain"
    "github.com/julianofirme/envsecret/internal/project"
)
```

- Import aliases only when a name collision makes them necessary. The
  established alias is `gokeyring` for `github.com/zalando/go-keyring`.

### Naming

- Follow standard Go conventions: `MixedCaps` for exported, `mixedCaps` for
  unexported.
- Acronyms are all-caps: `vaultURL`, `gcmIV`, not `vaultUrl`, `gcmIv`.
- Command-level boolean flags use the pattern `<cmd><Flag>` as package-level
  vars (e.g. `initForce`, `runClean`).
- Cobra command vars are named `<name>Cmd` (e.g. `initCmd`, `runCmd`).

### Error handling

- Always use `fmt.Errorf("context: %w", err)` to wrap errors. Never discard
  errors silently (the one exception is `_ = os.Remove(tmp)` in atomic writes
  where failure is inconsequential).
- Return errors up to the cobra `RunE` handler; cobra prints them to stderr
  automatically. Do not `log.Fatal` or `os.Exit` inside packages.
- User-facing error messages should be lowercase, no trailing period, and
  include the project name in brackets when relevant:
  ```
  [my-app] key "FOO" not found
  no vault found for project "my-app" — run `envs init` first
  ```

### Comments

- Every exported symbol gets a doc comment (`// FuncName ...`).
- Package-level doc comments are required for `internal/` packages; they
  appear at the top of the file before `package <name>`.
- Inline comments explain *why*, not *what*.

### Types

- Secrets are always `map[string]string` in memory. Do not introduce a custom
  `Secret` type unless there is a strong reason.
- Master keys are always `[]byte` (raw 32 bytes). They are encoded as hex
  strings only at the keychain boundary (`keychain.go`).
- File paths are always `string`; use `path/filepath` (not `path`) for
  manipulation.

---

## Architecture rules

### Adding a new command

1. Create `cmd/<name>.go` with `package cmd`.
2. Declare a `var <name>Cmd = &cobra.Command{...}`.
3. Register it in `cmd/root.go`'s `init()` with `rootCmd.AddCommand(<name>Cmd)`.
4. Use `RunE` (not `Run`) so errors propagate correctly.
5. Always resolve the project first: `proj, err := project.Resolve(projectFlag)`.
6. Load the master key via `keychain.Get(proj)` before touching the vault.

### Vault writes

- Always use `vault.Save`, never write to the vault file directly.
- `vault.Save` is atomic (write to `.tmp`, then `os.Rename`) and enforces
  `0600` permissions. Do not change this pattern.
- The vault is re-encrypted on every write with a fresh salt and IV —
  this is by design.

### `run` command

- Must use `syscall.Exec` (process replacement), never `exec.Command`. This
  ensures the `envs` process is gone after exec and cannot be inspected.
- Secrets are injected by appending `KEY=value` strings to the child env
  slice, which overrides any same-named variable already in the environment.

### Security constraints

- Never write a master key or plaintext secret to disk, a log, or stdout.
- The keychain service name is the constant `"envsecret"` (in `keychain.go`).
  Do not change it; existing users' keys are stored under this name.
- Vault directory permissions: `0700`. Vault file permissions: `0600`.
- Do not add a `--output json` or similar flag that would print decrypted
  values in a machine-readable format — the intent is to keep values out of
  shell history and pipes.

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/zalando/go-keyring` | OS keychain (macOS + Linux) |
| `golang.org/x/crypto/scrypt` | Key derivation |

Keep the dependency count minimal. Prefer the standard library. Do not add
logging frameworks, configuration libraries, or ORM-style abstractions.

---

## Module

Module path: `github.com/julianofirme/envsecret`
Binary name: `envs`
Go version: see `go.mod`
