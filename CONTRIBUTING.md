# Contributing to envsecret

Thank you for your interest in contributing. This document covers how to set
up the project locally, the code conventions to follow, and the process for
submitting changes.

---

## Prerequisites

| Tool | Purpose |
|---|---|
| Go 1.21+ | Build and test |
| `gofmt` / `goimports` | Formatting (included with Go) |
| `golangci-lint` | Linting ([install](https://golangci-lint.run/usage/install/)) |
| `libsecret-1-dev` | Linux only — keychain backend |

**macOS** — no extra dependencies beyond Go.

**Linux (Ubuntu/Debian):**
```bash
sudo apt install libsecret-1-dev
```

---

## Local setup

```bash
git clone https://github.com/julianofirme/envsecret.git
cd envsecret

# Verify it builds
go build ./...

# Run vet
go vet ./...

# Run linter
golangci-lint run

# Build the binary
go build -o envs .
```

---

## Project structure

```
envsecret/
├── main.go                   # entry point
├── cmd/                      # one file per CLI subcommand
└── internal/
    ├── keychain/             # OS keychain wrapper
    ├── vault/                # AES-256-GCM encryption
    └── project/              # project name resolution
```

See `AGENTS.md` for full architecture notes and code style guidelines.

---

## Code style

- Format with `gofmt` before every commit — no exceptions.
- Import groups: stdlib → third-party → internal (blank line between each).
- All exported symbols must have a doc comment.
- Errors: wrap with `fmt.Errorf("context: %w", err)`, never `log.Fatal` inside packages.
- Environment variable keys must match `[A-Z_][A-Z0-9_]*`.

See `AGENTS.md` for the full style guide.

---

## Making changes

1. Fork the repository and create a branch from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
2. Make your changes, ensuring `go build ./...` and `go vet ./...` pass.
3. Run `golangci-lint run` and fix any issues.
4. Commit with a clear message describing *why* the change was made.
5. Open a pull request against `main`.

### Security-sensitive changes

If your change touches `internal/vault`, `internal/keychain`, or the `run`
command, please explain the security implications in the PR description. See
[SECURITY.md](SECURITY.md) for the threat model.

---

## Reporting bugs

Use the [bug report template](.github/ISSUE_TEMPLATE/bug_report.md).
For security vulnerabilities, follow the process in [SECURITY.md](SECURITY.md)
instead of opening a public issue.

---

## Commit message style

```
<type>: <short summary>

<optional body explaining why, not what>
```

Types: `feat`, `fix`, `refactor`, `docs`, `test`, `ci`, `chore`.

Example:
```
fix: clear master key bytes from memory after vault load
```
