# ZeroPass

> The developer's command-line vault: offline-first, zero-knowledge, secrets management that fits your terminal workflow.

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://golang.org)
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
- **Crash-safe vault metadata** — `vault.json` persisted via atomic replace writes
- **Forward-compatible vault metadata** — unknown `vault.json` fields preserved on read/write

## Platform Support

- **CLI:** macOS, Linux (x86_64, arm64)
- **macOS app:** macOS 14+ (Sonoma and later)

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

### macOS app (SwiftUI)

Prereqs: Xcode (for `xcodebuild`/`lipo`), Go (per `go.mod`), and `make`.

Note: Xcode GUI builds may run with a minimal `PATH`. The ZeroPass Xcode target’s build phase exports `PATH` to include common Go install locations (`/opt/homebrew/bin`, `/usr/local/bin`, `/usr/local/go/bin`) so `go` is found. If your Go is installed elsewhere, update the build phase script accordingly.

Security note: the project currently sets `ENABLE_USER_SCRIPT_SANDBOXING=NO` so the bridge build script can access Go’s caches/module downloads. This improves dev UX but increases the blast radius of scripts; revisit if you need stricter build isolation.

```bash
# Run macOS app unit tests (no signing)
xcodebuild test \
  -project apps/macos/ZeroPass/ZeroPass.xcodeproj \
  -scheme ZeroPass \
  -destination 'platform=macOS,arch=arm64' \
  CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""

# Create an unsigned Release archive + artifacts under dist/macos/
# (ZeroPass.xcarchive, ZeroPass.zip, ZeroPass.dmg)
CODE_SIGNING_ALLOWED=NO bash apps/macos/scripts/build.sh

# Signed + notarized release (requires Apple credentials)
# Recommended once-per-machine setup:
#   xcrun notarytool store-credentials "zp-notary" --apple-id <id> --team-id <team> --password <app-specific>
# Then:
#   NOTARY_KEYCHAIN_PROFILE="zp-notary" NOTARIZE=YES bash apps/macos/scripts/build.sh
```

## CI

GitHub Actions workflow: [`.github/workflows/macos-build.yml`](.github/workflows/macos-build.yml)
- runs `go test ./...`
- runs the macOS app tests via `xcodebuild` (no signing)
- on tags `v*`, archives the macOS app and uploads `dist/macos/ZeroPass.xcarchive` as an artifact

## Documentation

- [Project Overview & PDR](docs/project-overview-pdr.md) — Vision, target users, scope, roadmap
- [Codebase Summary](docs/codebase-summary.md) — Architecture, module breakdown, dependencies
- [Code Standards](docs/code-standards.md) — Go conventions, patterns, testing strategy
- [System Architecture](docs/system-architecture.md) — Key hierarchy, encryption flow, sync protocol
- [Project Roadmap](docs/project-roadmap.md) — Phased development plan (MVP, Sync, Extensions, Mobile)
- [Deployment Guide](docs/deployment-guide.md) — Build, install, run, troubleshoot

## Contributing

Contributions welcome! Please read our code standards and submit PRs with test coverage.

## License

MIT
