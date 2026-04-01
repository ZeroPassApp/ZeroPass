# ZeroPass: Deployment & Build Guide

## Prerequisites

### System Requirements

- **OS:** macOS, Linux, or Windows (with Git Bash)
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
go build -o zeropass ./packages/cli/

# Verify build
./zeropass --version
./zeropass --help
```

### Build Sync Server

```bash
# Build sync server
go build -o syncserver ./services/syncserver/

# Verify build
./syncserver --version
./syncserver --help
```

### Cross-Compile

```bash
# Build for Linux on macOS
GOOS=linux GOARCH=amd64 go build -o zeropass-linux ./packages/cli/

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o zeropass.exe ./packages/cli/

# Build for ARM (e.g., Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o zeropass-arm64 ./packages/cli/
```

---

## Installation

### macOS

#### Option 1: Build from Source

```bash
go build -o zeropass ./packages/cli/
sudo mv zeropass /usr/local/bin/
chmod +x /usr/local/bin/zeropass

# Verify
which zeropass
zeropass --version
```

#### Option 2: Homebrew (Future)

```bash
# Not yet available; coming in Phase 2
brew install zeropass
```

### Linux

#### Build and Install

```bash
go build -o zeropass ./packages/cli/
sudo install -D zeropass /usr/local/bin/zeropass

# Verify
zeropass --version
```

#### systemd Shell Wrapper (Optional)

Create `/etc/bash_completion.d/zeropass` for CLI autocomplete:

```bash
eval "$(zeropass completion bash)"
```

### Windows

#### PowerShell

```powershell
go build -o zeropass.exe ./packages/cli/
Move-Item zeropass.exe $env:USERPROFILE\AppData\Local\bin\

# Add to PATH if not already set
[Environment]::GetEnvironmentVariable("PATH", "User")
```

---

## Running the CLI

### Initialize a New Vault

```bash
# Interactive setup
zeropass init

# Or with flags
zeropass init --vault-path=~/.zeropass

# Output: New vault created with recovery mnemonic printed
```

### Basic Operations

```bash
# Add a login credential
zeropass add --type=login --name="GitHub"

# Generate and add with password (interactive)
zeropass add --type=login --name="AWS Prod" --generate-password

# Add API key
zeropass add --type=apikey --name="Stripe Live" --key="sk_live_..."

# Retrieve credential
zeropass get "GitHub" --copy

# Search for credentials
zeropass search "github"

# List all items
zeropass list

# Password health report
zeropass health

# Export vault
zeropass export --format=json --output=backup.json

# Generate password
zeropass generate --length=32
```

### Vault Management

```bash
# Unlock vault (required after restart)
zeropass unlock

# Lock vault manually
zeropass lock

# View recovery mnemonic (if lost, vault is unrecoverable)
zeropass recovery show

# Import from Bitwarden export
zeropass import --source=bitwarden --file=bitwarden_export.json

# Export to CSV
zeropass export --format=csv --output=passwords.csv
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
./syncserver --port=8443 --db-path=./sync.db

# With API key authentication
./syncserver --port=8443 --db-path=./sync.db --api-key=my-secret-key-123
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
WorkingDirectory=/opt/zeropass
ExecStart=/opt/zeropass/syncserver --port=8443 --db-path=/var/lib/zeropass/sync.db --api-key=your-secret-key
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

#### 2. Setup User & Directory

```bash
sudo useradd -r -s /bin/false zeropass
sudo mkdir -p /var/lib/zeropass /opt/zeropass
sudo chown -R zeropass:zeropass /var/lib/zeropass /opt/zeropass
```

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

```dockerfile
# Dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o syncserver ./services/syncserver/

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/syncserver /usr/local/bin/
EXPOSE 8443
ENTRYPOINT ["syncserver"]
```

Build and run:

```bash
docker build -t zeropass-syncserver .
docker run -p 8443:8443 \
  -v /var/lib/zeropass:/data \
  -e API_KEY=your-secret-key \
  zeropass-syncserver \
  --port=8443 --db-path=/data/sync.db --api-key=$API_KEY
```

### Configure Client for Sync

```bash
# Set sync server URL (optional, defaults to local-only)
zeropass config set sync-url https://your-server.com:8443
zeropass config set sync-key your-secret-key

# Sync with server
zeropass sync pull
zeropass sync push
zeropass sync full
```

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
# Run integration tests (marked with build tag)
go test -tags=integration ./...
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

Configuration is stored in `~/.zeropass/config.json`:

```json
{
  "vault_path": "~/.zeropass",
  "auto_lock_minutes": 30,
  "clipboard_clear_seconds": 30,
  "color_output": true,
  "sync_url": "https://sync.example.com:8443",
  "sync_api_key": "your-bearer-token",
  "sync_interval_seconds": 300
}
```

### Environment Variables

```bash
# Override vault path
export ZEROPASS_VAULT_PATH=/custom/path

# Set sync server
export ZEROPASS_SYNC_URL=https://sync.example.com:8443
export ZEROPASS_SYNC_KEY=your-api-key

# Disable color output
export ZEROPASS_NO_COLOR=1

# Set auto-lock timeout (minutes)
export ZEROPASS_AUTO_LOCK_MINUTES=15
```

### Sync Server Configuration

```bash
# Flags (see syncserver --help)
--port              HTTP listen port (default: 8443)
--db-path           SQLite database path (default: ./sync.db)
--api-key           Bearer token for authentication (optional)
--max-devices       Max devices per user (default: 10)
--max-retries       Max sync retries (default: 3)
```

---

## Troubleshooting

### Issue: "Command not found: zeropass"

**Solution:**
```bash
# Verify binary is in PATH
which zeropass

# Add to PATH if needed
export PATH="$PATH:/path/to/zeropass/binary"

# Or move to standard location
sudo mv zeropass /usr/local/bin/
```

### Issue: "Vault locked" error

**Solution:**
```bash
# Unlock vault
zeropass unlock

# Or check if already unlocked
zeropass status
```

### Issue: "Checksum mismatch" error

**Scenario:** Corrupted encrypted file or disk error

**Solution:**
```bash
# Verify vault integrity
zeropass health

# Check vault directory
ls -la ~/.zeropass/

# Recover from backup or sync source (Phase 2+)
zeropass sync pull
```

### Issue: Import fails with "unsupported format"

**Solution:**
```bash
# Verify export file is correct format
file bitwarden_export.json

# Try importing with explicit source
zeropass import --source=bitwarden --file=bitwarden_export.json

# Check logs for specific error
zeropass import ... --verbose
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
top -p $(pgrep zeropass)
```

---

## Performance Optimization

### CLI Performance

```bash
# Disable color for faster output redirection
zeropass list --no-color

# Use JSON output for parsing (faster than table)
zeropass list --output=json
```

### FTS5 Index Optimization

```bash
# Rebuild index (if search becomes slow)
zeropass admin rebuild-index

# This is automatic on vault creation/import
```

### Sync Server Optimization

For large deployments:

```bash
# Use PostgreSQL instead of SQLite (Phase 2+)
export ZEROPASS_DB_URL=postgresql://user:pass@host/zeropass

# Enable query caching
--cache-seconds=300

# Increase worker threads
--workers=8
```

---

## Security Checklist

- [ ] Run `go vet ./...` before committing
- [ ] Close vault after sensitive operations (`zeropass lock`)
- [ ] Store recovery mnemonic securely (printed, not file)
- [ ] Enable API key on sync server (`--api-key`)
- [ ] Use HTTPS/TLS for sync server (configure reverse proxy)
- [ ] Rotate API keys periodically
- [ ] Monitor sync server logs for failed sync attempts
- [ ] Run security audit: `go test -race ./... && go test -coverprofile=coverage.out ./...`

---

## Next Steps

1. **Initialize vault:** `zeropass init`
2. **Add first credential:** `zeropass add --type=login`
3. **Test retrieval:** `zeropass get --copy`
4. **Setup sync (Phase 2):** Deploy sync server and configure client
5. **Join community:** GitHub discussions, Discord

---

**Document Version:** 1.0  
**Last Updated:** April 1, 2026  
**Owner:** DevOps & Infrastructure
