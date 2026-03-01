# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| latest (`main`) | Yes |

---

## Reporting a vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Please report them privately via GitHub's built-in security advisory feature:

1. Go to https://github.com/julianofirme/envsecret/security/advisories
2. Click **"New draft security advisory"**
3. Describe the vulnerability, steps to reproduce, and potential impact

You will receive a response within **72 hours**. If the vulnerability is
confirmed, a patched release will be published and you will be credited in the
advisory (unless you prefer to remain anonymous).

---

## Threat model

envsecret is designed to eliminate *ambient* secret exposure — files on disk,
shell history, and environment variables visible to sibling processes. It is
not designed to protect against a compromised runtime or OS.

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

## Cryptographic details

| Property | Value |
|---|---|
| Algorithm | AES-256-GCM (authenticated encryption) |
| Key derivation | scrypt (N=16384, r=8, p=1) |
| Salt | 32 bytes, random per write |
| IV | 12 bytes, random per write |
| Auth tag | 16 bytes — detects tampering |
| Master key storage | OS keychain (macOS Keychain / libsecret on Linux) |

The vault is re-encrypted on every write with a fresh salt and IV.
Master keys are stored only in the OS keychain, never written to disk.
