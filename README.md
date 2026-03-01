# envsecret

[![CI](https://github.com/julianofirme/envsecret/actions/workflows/ci.yml/badge.svg)](https://github.com/julianofirme/envsecret/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/julianofirme/envsecret)](https://github.com/julianofirme/envsecret/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/julianofirme/envsecret)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Encrypted environment variable vault for macOS and Linux. Per-project isolation, zero plaintext on disk.

Secrets are stored in AES-256-GCM encrypted files. Each project gets its own vault and its own master key in the OS keychain. Secrets are injected exclusively into child process environments — invisible to shell history, `printenv`, and AI coding agents.

---

## Install

**One-liner (macOS & Linux):**

```bash
curl -fsSL https://raw.githubusercontent.com/julianofirme/envsecret/main/install.sh | bash
```

**From source:**

```bash
# Linux prerequisite
sudo apt install libsecret-1-dev   # Ubuntu/Debian
sudo dnf install libsecret-devel   # Fedora
sudo pacman -S libsecret           # Arch

git clone https://github.com/julianofirme/envsecret.git
cd envsecret
go build -o envs .
sudo mv envs /usr/local/bin/
```

---

## Quick start

```bash
cd ~/projects/my-app

envs init
envs set DATABASE_URL "postgres://user:pass@host/db"
envs set OPENAI_API_KEY "sk-..."

envs run -- node server.js
```

The project name is auto-detected from the git repository root. No configuration needed.

---

## Project isolation

Each project gets its own vault and its own master key.

```
~/.envsecret/
├── my-app/
│   └── vault.enc
├── api-service/
│   └── vault.enc
└── frontend/
    └── vault.enc
```

**Project resolution order:**

1. `--project` flag (explicit override)
2. Git repository root name (`git rev-parse --show-toplevel`)
3. Current directory name (fallback when not in a git repo)

```bash
# Auto-detected from git root
cd ~/projects/api-service
envs set DB_URL "..."           # writes to api-service vault

# Explicit override
envs --project staging set DB_URL "..."

# Run a command in a specific project context
envs --project api-service run -- node server.js
```

---

## Commands

### `envs init`

Creates a vault for the current project and stores the master key in the OS keychain.

```bash
envs init

# Reinitialize (DESTROYS all stored secrets for this project)
envs init --force
```

You will be prompted to choose between a passphrase or a generated 256-bit key. The key is stored in the OS keychain immediately and never written to disk.

---

### `envs set <KEY> <VALUE>`

Encrypts and stores a variable in the current project vault.

```bash
envs set API_KEY "your-secret"
envs set DATABASE_URL "postgres://..."
```

---

### `envs run -- <command> [args...]`

Spawns a child process with vault secrets injected into its environment. The parent shell is never modified.

```bash
envs run -- node server.js
envs run -- npm run dev
envs run -- make deploy
```

The `--` separator is required. Everything after it is the command.

**`--clean` flag:** strips the parent shell environment entirely. Only vault vars and `PATH` are passed to the child.

```bash
envs run --clean -- node server.js
```

---

### `envs get <KEY>`

Prints a single decrypted value to stdout. Suitable for scripting.

```bash
envs get DATABASE_URL

DB=$(envs get DATABASE_URL)
```

---

### `envs list`

Lists stored key names for the current project. Values are never shown.

```bash
envs list
# [my-app] 3 variable(s):
#
#   DATABASE_URL
#   OPENAI_API_KEY
#   STRIPE_SECRET
```

---

### `envs projects`

Lists all projects with initialized vaults. Marks the currently active project.

```bash
envs projects
# 3 project(s):
#
#   api-service
#   frontend (current)
#   my-app
```

---

### `envs delete <KEY>`

Removes a variable from the current project vault.

```bash
envs delete OLD_SECRET
```

---

### `envs status`

Shows vault diagnostics for the current project without exposing any secrets.

```bash
envs status
# envs status — project: my-app
#
#   Vault path:    /home/user/.envsecret/my-app/vault.enc
#   Vault exists:  yes
#   Keychain key:  found
#   File mode:     600 (ok)
#   Last modified: 2024-11-01T10:22:00Z
#   Variables:     3
#
#   Other projects: [api-service frontend]
```

---

## How secrets stay hidden from AI agents

AI coding agents typically access secrets through three vectors: files on disk, the shell environment, and shell history. envs closes all three.

**Files on disk.** The vault file is AES-256-GCM encrypted. Without the master key it is indistinguishable from random bytes. File permissions are `600`. The master key is in the OS keychain, never on disk.

**Shell environment.** `envs run` replaces the current process with the child via `syscall.Exec`. Vault vars are injected directly into the child's environment. The parent shell is never modified — running `printenv` before or after sees nothing new.

**Shell history.** There is no `export KEY=value` command to record. The secret never appears in a command you type.

### What it does not protect against

If an AI agent has arbitrary code execution inside the process that `envs run` spawned, it can read `os.Getenv("MY_SECRET")`. This tool eliminates ambient exposure — files, shell env, history — not a compromised runtime.

---

## Encryption details

| Property | Value |
|---|---|
| Algorithm | AES-256-GCM (authenticated encryption) |
| Key derivation | scrypt (N=16384, r=8, p=1) |
| Salt | 32 bytes, random per write |
| IV | 12 bytes, random per write |
| Auth tag | 16 bytes — detects tampering |
| Master key storage | OS keychain (macOS Keychain / libsecret on Linux) |
| Keychain account | `project:<name>` — one key per project |
| Vault location | `~/.envsecret/<project>/vault.enc` |
| Vault permissions | `600` |

The vault is re-encrypted on every write with a fresh salt and IV.

---

## Integrating with existing projects

**Replace `.env` + `dotenv`:**

```bash
# Before
node -r dotenv/config server.js

# After
envs run -- node server.js
```

**Next.js:**

```bash
envs run -- npm run dev
envs run -- npm run build
```

**package.json scripts:**

```json
{
  "scripts": {
    "dev": "envs run -- next dev",
    "start": "envs run -- node server.js"
  }
}
```

**Monorepo with multiple services:**

```bash
cd services/api
envs init
envs set DB_URL "postgres://..."

cd services/worker
envs init
envs set QUEUE_URL "redis://..."

cd services/api && envs run -- node index.js
cd services/worker && envs run -- node worker.js
```

---

## Migrating from a `.env` file

```bash
grep -v '^#' .env | grep '=' | while IFS='=' read -r key value; do
  envs set "$key" "$value"
done

envs list
rm .env
echo ".env" >> .gitignore
```

---

## Threat model

| Threat | Protected |
|---|---|
| `.env` files readable on disk | Yes |
| `export KEY=value` in shell history | Yes |
| `printenv` / `env` in sibling processes | Yes |
| Agent reading `~/.bash_history` | Yes |
| Agent scanning filesystem for plaintext secrets | Yes |
| Secrets leaking between projects | Yes — isolated vaults and keys |
| Agent calling `os.Getenv` inside your app | No |
| Root-level `/proc` inspection or `ptrace` | No |
| Compromised OS keychain | No |

---

## License

MIT