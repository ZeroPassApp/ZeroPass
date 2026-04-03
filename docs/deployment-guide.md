# ZeroPass: Deployment & Build Guide

## Prerequisites

### System Requirements

- **OS (official support):** macOS 14+ (Sonoma and later) or Linux
- **Windows:** best-effort only (not an official support target yet)
- **Go:** 1.26.1 or later (download from [golang.org](https://golang.org/dl))
- **Git:** 2.30+
- **SQLite3:** Bundled with Go driver (no external installation needed)

### Verify Prerequisites

```bash
go version          # Go 1.26+
git --version       # 2.30+
```

---

## Build from Source

### Clone Repository

```bash
git clone https://github.com/zeropass/zeropass.git
cd zeropass
```

### Download Dependencies

```bash
go mod download
go mod verify
```

### Build CLI

```bash
# Build for current OS/architecture
go build -o zp ./packages/cli/

# Verify build
./zp --version
./zp --help
```

### Build Sync Server

```bash
# Build sync server
go build -o syncserver ./services/syncserver/

# Verify build
./syncserver -h
```

### Build macOS app (SwiftUI)

Prereqs: Xcode (`xcodebuild`, macOS SDK tools like `lipo`), Go (per `go.mod`), and `make`.

```bash
# Run macOS app tests (no signing)
xcodebuild test \
  -project apps/macos/ZeroPass/ZeroPass.xcodeproj \
  -scheme ZeroPass \
  -destination 'platform=macOS,arch=arm64' \
  CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""

# Archive (Release) + create artifacts under dist/macos/
# (ZeroPass.xcarchive, ZeroPass.zip, ZeroPass.dmg)
CODE_SIGNING_ALLOWED=NO bash apps/macos/scripts/build.sh

# Signed + notarized release (requires Apple credentials)
# Recommended once-per-machine setup:
#   xcrun notarytool store-credentials "zp-notary" --apple-id <id> --team-id <team> --password <app-specific>
# Then:
#   NOTARY_KEYCHAIN_PROFILE="zp-notary" NOTARIZE=YES bash apps/macos/scripts/build.sh
```

Notes:
- The Xcode project builds the Go bridge as part of the app target via a build phase that runs `make -C bridge build-universal`.
- Optional packaging scripts live in `apps/macos/scripts/` (e.g. `create-dmg.sh`, `notarize.sh`).
- Security tradeoff: the project currently sets `ENABLE_USER_SCRIPT_SANDBOXING=NO` to avoid Go cache/mod-cache sandbox issues. Keep this in mind for tighter build isolation.

### Cross-Compile

```bash
# Build for Linux on macOS
GOOS=linux GOARCH=amd64 go build -o zp-linux ./packages/cli/

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o zp.exe ./packages/cli/

# Build for ARM (e.g., Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o zp-arm64 ./packages/cli/
```

---

## Installation

### macOS

#### Option 1: Build from Source

```bash
go build -o zp ./packages/cli/
sudo mv zp /usr/local/bin/
chmod +x /usr/local/bin/zp

# Verify
which zp
zp --version
```

#### Option 2: Homebrew (Future)

```bash
# Not yet available; coming in Phase 2.
# The formula may still be named `zeropass`, but it should install the `zp` binary.
brew install zeropass
```

### Linux

#### Build and Install

```bash
go build -o zp ./packages/cli/
sudo install -D zp /usr/local/bin/zp

# Verify
zp --version
```

#### systemd Shell Wrapper (Optional)

Create `/etc/bash_completion.d/zp` for CLI autocomplete:

```bash
eval "$(zp completion bash)"
```

### Windows

#### PowerShell

```powershell
go build -o zp.exe ./packages/cli/
Move-Item zp.exe $env:USERPROFILE\AppData\Local\bin\

# Add to PATH if not already set
[Environment]::GetEnvironmentVariable("PATH", "User")
```

---

## Running the CLI

### Initialize a New Vault

```bash
# Interactive setup
zp init

# Or with flags
zp init --vault-path=~/.zeropass

# Output: New vault created with recovery mnemonic printed
```

### Basic Operations

```bash
# Add a login credential
zp add --type=login --name="GitHub"

# Generate and add with password (interactive)
zp add --type=login --name="AWS Prod" --generate-password

# Add API key
zp add --type=apikey --name="Stripe Live" --key="sk_live_..."

# Retrieve credential
zp get "GitHub" --copy

# Search for credentials
zp search "github"

# List all items
zp list

# Password health report
zp health

# Export vault (--force skips plaintext warning for non-encrypted exports)
zp export --format=json --file=backup.json
zp export --format=json --file=backup.json --force

# Generate password
zp generate --length=32
```

### Vault Management

```bash
# Unlock vault (required after restart)
zp unlock

# Lock vault manually
zp lock

# Validate a recovery mnemonic against the current vault
zp recovery --validate "word1 word2 ..."

# Regenerate recovery mnemonic (requires vault unlock)
zp recovery --regenerate

# Import from Bitwarden export
zp import --from=bitwarden --file=bitwarden_export.json

# Import from other sources (safari, lastpass, keepass, 1pux, etc.)
zp import --from=safari --file=safari_passwords.csv
zp import --from=lastpass --file=lastpass_export.csv
zp import --from=keepass --file=keepass_export.csv
zp import --from=1pux --file=1password_export.1pux

# Export to CSV (shows plaintext warning; use --force to skip)
zp export --format=csv --file=passwords.csv
zp export --format=csv --file=passwords.csv --force
```

---

## Building & Running Sync Server

### Build

```bash
go build -o syncserver ./services/syncserver/
```

### Run Locally (Development)

```bash
# Start server on port 8443
./syncserver --port=8443 --db-path=./zeropass-sync.db

# With API key authentication
./syncserver --port=8443 --db-path=./zeropass-sync.db --api-key=my-secret-key-123
```

### Run with systemd (Linux Production)

#### 1. Create Service File

```bash
sudo tee /etc/systemd/system/zeropass-syncserver.service > /dev/null <<EOF
[Unit]
Description=ZeroPass Sync Server
After=network.target

[Service]
Type=simple
User=zeropass
Group=zeropass
WorkingDirectory=/opt/zeropass
StateDirectory=zeropass
ExecStart=/opt/zeropass/syncserver --port=8443 --db-path=/var/lib/zeropass/zeropass-sync.db --api-key=your-secret-key
Restart=on-failure
RestartSec=10s

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true

[Install]
WantedBy=multi-user.target
EOF
```

`StateDirectory=zeropass` keeps `/var/lib/zeropass` writable even with `ProtectSystem=strict`, so the SQLite database can be created without weakening the rest of the filesystem policy.

#### 2. Setup User & Directory

```bash
sudo useradd -r -s /bin/false zeropass
sudo install -d -o zeropass -g zeropass /opt/zeropass
```

You do not need to pre-create `/var/lib/zeropass`; `systemd` creates it on service start because of `StateDirectory=zeropass`.

#### 3. Copy Binary & Enable

```bash
sudo cp syncserver /opt/zeropass/
sudo systemctl daemon-reload
sudo systemctl enable zeropass-syncserver.service
sudo systemctl start zeropass-syncserver.service

# Check status
sudo systemctl status zeropass-syncserver.service
sudo journalctl -u zeropass-syncserver.service -f
```

### Docker (Optional)

The repository includes a real Dockerfile at `services/syncserver/Dockerfile`.

The image now runs as a dedicated non-root user (`uid=10001`, `gid=10001`) with `/var/lib/zeropass` as its writable working directory. Named Docker volumes are the recommended persistence path; for bind mounts, make sure the target directory is writable by `10001:10001`.

Build and run:

```bash
docker build -t zeropass-syncserver -f services/syncserver/Dockerfile .
docker volume create zeropass-sync-data
docker run -p 8443:8443 \
  -v zeropass-sync-data:/var/lib/zeropass \
  zeropass-syncserver \
  --port=8443 --db-path=/var/lib/zeropass/zeropass-sync.db --api-key=your-secret-key
```

If you prefer a bind mount instead of a named volume:

```bash
sudo mkdir -p /var/lib/zeropass
sudo chown -R 10001:10001 /var/lib/zeropass
docker run -p 8443:8443 \
  -v /var/lib/zeropass:/var/lib/zeropass \
  zeropass-syncserver \
  --port=8443 --db-path=/var/lib/zeropass/zeropass-sync.db --api-key=your-secret-key
```

### Preview Smoke Test

The repository includes a repeatable smoke-test script for the verified preview path. It checks:

- health endpoint startup
- bearer-auth enforcement
- malformed request handling
- device registration
- push/pull flow
- non-root Docker runtime configuration
- persistence across restart
- timestamp-first conflict response

The shell script exercises the bearer-auth path. Auth-off preview mode is covered by `go test ./services/syncserver/...` via `services/syncserver/smoke_test.go`.

```bash
# Bare binary mode
bash services/syncserver/smoke-test.sh binary

# Container mode
bash services/syncserver/smoke-test.sh docker

# Run both verified paths
bash services/syncserver/smoke-test.sh both
```

### Configure Client for Sync

The Go CLI does not currently expose first-class `sync` subcommands. Today, sync is configured through the macOS app settings and the native bridge layer, which persist vault-local `sync.json` state. Treat sync as a self-hosted preview feature until broader integration coverage is complete.

### Supported Operating Modes

- **Verified preview path:** SQLite + restart persistence, with bearer-auth exercised by `services/syncserver/smoke-test.sh` in binary/Docker modes and auth-off preview mode covered by `go test ./services/syncserver/...`
- **macOS integration path:** bridge-backed sync configuration persisted in vault-local `sync.json`
- **Advanced preview paths:** PostgreSQL storage and blob-backed payload storage exist, but are not yet the primary documented operator path

---

## Testing

### Run Unit Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./core/crypto/...
go test ./packages/cli/cmd/...
```

### Run Tests with Coverage

```bash
# Generate coverage
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Get coverage percentage
go tool cover -func=coverage.out | tail -1
```

### Run Benchmarks

```bash
# Benchmark encryption/decryption
go test -bench=. -benchmem ./core/crypto/

# Benchmark specific function
go test -bench=BenchmarkArgon2id -benchmem ./core/crypto/kdf/
```

### Integration Tests

```bash
# Run syncserver integration + smoke tests
go test ./services/syncserver/...

# Run the operator smoke test script
bash services/syncserver/smoke-test.sh both
```

### Code Quality Checks

```bash
# Format code
go fmt ./...

# Run linter
go vet ./...

# Race detector (concurrent safety)
go test -race ./...

# Optional: golangci-lint (install via golangci-lint.com)
golangci-lint run ./...
```

---

## Configuration

### CLI Configuration

The CLI currently uses command flags and vault-local metadata rather than a standalone user config file. The primary runtime knobs are:

- `--vault-path`
- `--output`
- `--no-color`

### Sync Server Configuration

```bash
# Flags (see syncserver --help)
--port              HTTP listen port (default: 8443)
--db-path           SQLite database path (default: zeropass-sync.db)
--api-key           Bearer token for authentication (optional)
--storage           Storage backend: sqlite or postgres
--postgres-url      PostgreSQL connection string (required with --storage=postgres)
--blob-enabled      Enable S3/MinIO blob storage
--blob-endpoint     S3/MinIO endpoint
--blob-bucket       S3/MinIO bucket name
--blob-access-key   S3/MinIO access key
--blob-secret-key   S3/MinIO secret key
```

For release readiness, the verified deployment path is SQLite. PostgreSQL and blob flags should be treated as advanced preview options unless you validate them in your environment.

---

## Troubleshooting

### Issue: "Command not found: zp"

**Solution:**
```bash
# Verify binary is in PATH
which zp

# Add to PATH if needed
export PATH="$PATH:/path/to/zp/binary"

# Or move to standard location
sudo mv zp /usr/local/bin/
```

### Issue: "Vault locked" error

**Solution:**
```bash
# Unlock vault
zp unlock

# Lock manually when finished
zp lock
```

### Issue: "Checksum mismatch" error

**Scenario:** Corrupted encrypted file or disk error

**Solution:**
```bash
# Verify vault integrity
zp health

# Check vault directory
ls -la ~/.zeropass/

# Recover from backup or re-import if needed
```

### Issue: Import fails with "unsupported format"

**Solution:**
```bash
# Verify export file is correct format
file bitwarden_export.json

# Check supported sources
# Supported: chrome, firefox, safari, 1password, 1pux, bitwarden, lastpass, keepass, csv
zp import --from=bitwarden --file=bitwarden_export.json
```

### Issue: Sync server won't start

**Solution:**
```bash
# Check port availability
lsof -i :8443

# Verify database permissions
ls -la /var/lib/zeropass/
sudo chown -R zeropass:zeropass /var/lib/zeropass

# Check service logs
sudo journalctl -u zeropass-syncserver.service -n 50
```

### Issue: High memory usage

**Scenario:** Argon2id KDF consuming too much RAM

**Solution:**
```bash
# This is expected (64MB) during unlock
# Memory is freed after unlock completes

# Monitor process
top -p $(pgrep zp)
```

---

## Performance Optimization

### CLI Performance

```bash
# Disable color for faster output redirection
zp list --no-color

# Use JSON output for parsing
zp list --output=json
```

### Sync Server Optimization

For larger deployments, prefer PostgreSQL over SQLite:

```bash
./syncserver --storage=postgres --postgres-url='postgresql://user:pass@host/zeropass?sslmode=disable'
```

Treat this as an advanced preview/operator path until you have validated it locally; the default smoke-tested deployment path remains SQLite.

---

## Release Gates

Use this checklist before tagging or packaging a release candidate:

- [ ] `go test ./...`
- [ ] `xcodebuild test -project apps/macos/ZeroPass/ZeroPass.xcodeproj -scheme ZeroPass -destination 'platform=macOS,arch=arm64' CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO CODE_SIGN_IDENTITY=""`
- [ ] `go build -o zp ./packages/cli/`
- [ ] `go build -o syncserver ./services/syncserver/`
- [ ] `docker build -t zeropass-syncserver -f services/syncserver/Dockerfile .`
- [ ] `bash services/syncserver/smoke-test.sh both`
- [ ] Docs still describe sync as preview, not GA

---

## Security Checklist

- [ ] Run `go vet ./...` before committing
- [ ] Close vault after sensitive operations (`zp lock`)
- [ ] Store recovery mnemonic securely (printed, not file)
- [ ] Enable API key on sync server (`--api-key`)
- [ ] Use HTTPS/TLS for sync server (configure reverse proxy)
- [ ] Rotate API keys periodically
- [ ] Monitor sync server logs for failed sync attempts
- [ ] Run security audit: `go test -race ./... && go test -coverprofile=coverage.out ./...`

---

## Next Steps

1. **Initialize vault:** `zp init`
2. **Add first credential:** `zp add --type=login`
3. **Test retrieval:** `zp get --copy`
4. **Setup sync (Phase 2):** Deploy sync server and configure client
5. **Join community:** GitHub discussions, Discord

---

**Document Version:** 1.2  
**Last Updated:** April 2026  
**Owner:** DevOps & Infrastructure
