# Changelog

All notable changes to envsecret are documented here.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versioning follows [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

---

## [0.1.0] — 2026-03-01

### Added
- `envs init [--force]` — create a vault and store the master key in the OS keychain
- `envs set <KEY> <VALUE>` — encrypt and store a variable; validates key format `[A-Z_][A-Z0-9_]*`
- `envs get <KEY>` — print a single decrypted value to stdout
- `envs list` — list stored key names (values never shown)
- `envs delete <KEY>` — remove a variable from the vault
- `envs run [--clean] -- <cmd>` — spawn a child process with secrets injected; uses `syscall.Exec` so the parent process is replaced
- `envs import [file]` — import KEY=VALUE pairs from a `.env` file or stdin
- `envs rename <OLD> <NEW>` — rename a key within the vault
- `envs copy <SRC> <DST>` — duplicate a key within the vault
- `envs destroy [--yes]` — permanently remove vault file and keychain key
- `envs projects` — list all initialized projects, marking the current one
- `envs status` — show vault diagnostics without exposing secrets
- AES-256-GCM encryption with scrypt key derivation (N=16384, r=8, p=1)
- Per-write random 32-byte salt and 12-byte nonce
- Atomic vault writes via `.tmp` + `os.Rename`
- Vault and directory permissions enforced at `0600` / `0700`
- macOS Keychain and Linux libsecret support via `go-keyring`
- Project resolution: `--project` flag → git root name → current directory name

[Unreleased]: https://github.com/julianofirme/envsecret/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/julianofirme/envsecret/releases/tag/v0.1.0
