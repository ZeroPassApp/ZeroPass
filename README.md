# ZeroPass

> The developer's command-line vault: offline-first, zero-knowledge, secrets management that fits your terminal workflow.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-90.1%25-brightgreen)]()

## Features

- **CLI-first** — `zeropass` CLI is the primary interface
- **Offline-first** — full functionality without internet
- **Zero-knowledge** — all encryption client-side, server never sees plaintext
- **Dev secrets native** — API keys, SSH keys, `.env` injection
- **Self-sovereign** — you own 100% of your data

## Quick Start

```bash
# Build
go build -o zeropass ./packages/cli/

# Initialize a new vault
./zeropass init

# Add credentials
./zeropass add --type=login --name="GitHub" --username="user" --generate-password
./zeropass add --type=apikey --name="AWS Prod" --key="AKIA..."

# Search & retrieve
./zeropass search "github"
./zeropass get "GitHub" --copy

# Inject secrets into dev workflow
./zeropass run --env-file=.env -- npm start

# Password health check
./zeropass health
```

## Architecture

```
zeropass/
├── core/
│   ├── crypto/          # Argon2id KDF + AES-256-GCM + BIP-39 recovery
│   ├── vault/           # CRUD, search, versioning, health, import/export
│   └── sync/            # Delta sync client/server, conflict resolution
├── packages/
│   └── cli/             # `zeropass` CLI (cobra-based)
└── services/
    └── syncserver/      # Self-hosted sync server
```

## Security

- **Argon2id** KDF (64MB memory, 3 iterations, 4 parallelism)
- **AES-256-GCM** per-item encryption with unique nonces
- **HKDF-SHA256** per-item key derivation
- **BIP-39** 12-word recovery mnemonic
- **All metadata encrypted** (unlike LastPass)
- **Memory zeroing** for all sensitive data
- **HIBP integration** (opt-in, k-anonymity model)

## CLI Commands

| Command | Description |
|---------|-------------|
| `init` | Create a new vault |
| `unlock/lock` | Vault access control |
| `add` | Add credential (login, apikey, sshkey, note, creditcard, identity) |
| `get` | Retrieve credential (`--copy` to clipboard) |
| `search` | Full-text search |
| `list` | List items with filters |
| `edit` | Edit credential |
| `delete` | Delete credential |
| `generate` | Password/passphrase generator |
| `health` | Password health report |
| `recovery` | Recovery key management |
| `import` | Import from Chrome, Firefox, 1Password, Bitwarden |
| `export` | Export vault (JSON, CSV, encrypted) |
| `run` | Inject secrets via `.env` file (`zp://` references) |
| `env` | Environment management (dev/staging/prod) |

## Sync Server

```bash
# Build and run
go build -o syncserver ./services/syncserver/
./syncserver --port=8443 --api-key=your-secret-key
```

## Development

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Build CLI
go build -o zeropass ./packages/cli/
```

## License

MIT
