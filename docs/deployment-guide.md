# GoClaw CLI - Deployment Guide

## Installation Methods

### 1. From Source (Development)

**Prerequisites:**
- Go 1.25.3+
- Git
- Make

**Installation:**

```bash
# Clone repository
git clone https://github.com/nextlevelbuilder/goclaw-cli.git
cd goclaw-cli

# Build locally
make build
./goclaw --version

# Or install to GOPATH/bin
make install
goclaw --version  # Should work if $GOPATH/bin is in PATH
```

**Verify Installation:**

```bash
$ goclaw version
GoClaw CLI v1.0.0 (commit: abc1234, built: 2026-03-15T10:00:00Z)

$ goclaw --help
Usage:
  goclaw [command]

Available Commands:
  agents      Manage agents
  auth        Authentication and profiles
  ...
```

### 2. From Release (Production)

**Download Latest:**

```bash
# macOS Intel
curl -L https://github.com/nextlevelbuilder/goclaw-cli/releases/download/v1.0.0/goclaw_1.0.0_darwin_amd64.tar.gz | tar xz
mv goclaw /usr/local/bin/

# macOS Apple Silicon
curl -L https://github.com/nextlevelbuilder/goclaw-cli/releases/download/v1.0.0/goclaw_1.0.0_darwin_arm64.tar.gz | tar xz
mv goclaw /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/nextlevelbuilder/goclaw-cli/releases/download/v1.0.0/goclaw_1.0.0_linux_amd64.tar.gz | tar xz
sudo mv goclaw /usr/local/bin/

# Linux (arm64)
curl -L https://github.com/nextlevelbuilder/goclaw-cli/releases/download/v1.0.0/goclaw_1.0.0_linux_arm64.tar.gz | tar xz
sudo mv goclaw /usr/local/bin/

# Windows (amd64)
# Download: goclaw_1.0.0_windows_amd64.zip
# Extract and add to PATH
```

### 3. Via go install (Convenience)

```bash
go install github.com/nextlevelbuilder/goclaw-cli@latest
goclaw --version
```

**Note:** Installs latest from main branch. For stable releases, use release binaries.

### 4. Via Package Manager (Future)

**Homebrew (Planned):**

```bash
brew tap nextlevelbuilder/goclaw
brew install goclaw-cli
```

---

## Initial Configuration

### 1. Create Config Directory

```bash
mkdir -p ~/.goclaw
```

### 2. First Login

**Interactive Login:**

```bash
goclaw auth login --server https://goclaw.example.com --token your-token
# OR
goclaw auth login --server https://goclaw.example.com --pair
# (Enter pairing code from browser)
```

**Automation Login:**

```bash
export GOCLAW_SERVER=https://goclaw.example.com
export GOCLAW_TOKEN=your-token
goclaw status  # Verify connection
```

### 3. Verify Configuration

```bash
$ goclaw status
Server: https://goclaw.example.com
Status: healthy
Version: 1.0.0
Agents: 5
```

**Check Config File:**

```bash
$ cat ~/.goclaw/config.yaml
active_profile: production
profiles:
  - name: production
    server: https://goclaw.example.com
    # token stored in OS keyring
```

---

## Multi-Profile Setup

### Add Staging Profile

```bash
goclaw auth login --profile staging --server https://staging.goclaw.example.com --token staging-token
```

### Switch Profiles

```bash
# Set active profile
goclaw auth use-context staging

# Override per-command
goclaw --profile staging agents list
```

### View All Profiles

```bash
$ cat ~/.goclaw/config.yaml
active_profile: production
profiles:
  - name: production
    server: https://goclaw.example.com
  - name: staging
    server: https://staging.goclaw.example.com
  - name: local
    server: http://localhost:8080
```

---

## Build & Release

### Local Build

```bash
# Build binary
make build
./goclaw --version

# Run tests
make test

# Run linting
make lint

# Install to GOPATH/bin
make install

# Clean build artifacts
make clean
```

### Create Release (Automated via GitHub Actions)

**Trigger Release:**

```bash
# Create and push tag
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions automatically:
# 1. Runs go vet and go test
# 2. Builds for all platforms (GoReleaser)
# 3. Creates checksums
# 4. Publishes release to GitHub
```

**Verify Release:**

```bash
# List available versions
curl https://api.github.com/repos/nextlevelbuilder/goclaw-cli/releases

# Download specific version
curl -L https://github.com/nextlevelbuilder/goclaw-cli/releases/download/v1.0.0/goclaw_1.0.0_linux_amd64.tar.gz
```

### GoReleaser Configuration (.goreleaser.yaml)

```yaml
version: 2
project_name: goclaw

builds:
  - main: .                          # Build entry point
    binary: goclaw                   # Output binary name
    ldflags:                         # Inject version at build-time
      - -s -w
      - -X github.com/nextlevelbuilder/goclaw-cli/cmd.Version={{.Version}}
      - -X github.com/nextlevelbuilder/goclaw-cli/cmd.Commit={{.Commit}}
      - -X github.com/nextlevelbuilder/goclaw-cli/cmd.BuildDate={{.Date}}
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz                  # Compression format
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip                 # Windows gets .zip instead of .tar.gz

checksum:
  name_template: checksums.txt      # SHA256 checksums

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"                    # Exclude doc commits
      - "^test:"
      - "^ci:"
```

**Generated Artifacts:**

```
dist/
├── goclaw_1.0.0_darwin_amd64.tar.gz      # macOS Intel
├── goclaw_1.0.0_darwin_arm64.tar.gz      # macOS Apple Silicon
├── goclaw_1.0.0_linux_amd64.tar.gz       # Linux Intel
├── goclaw_1.0.0_linux_arm64.tar.gz       # Linux ARM
├── goclaw_1.0.0_windows_amd64.zip        # Windows Intel
├── goclaw_1.0.0_windows_arm64.zip        # Windows ARM
└── checksums.txt                          # SHA256 checksums
```

---

## CI/CD Pipeline

### GitHub Actions Workflows

#### ci.yaml (Test & Build on Push)

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.25']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - name: Build
        run: go build ./...
      - name: Vet
        run: go vet ./...
      - name: Test
        run: go test -race ./...
```

**Triggers:**
- On push to `main` branch
- On pull requests to `main` branch

**Steps:**
1. Checkout code
2. Setup Go 1.25
3. Build all packages
4. Run `go vet` (linting)
5. Run tests with race detector

#### release.yaml (Build & Release on Tag)

```yaml
name: Release

on:
  push:
    tags: ['v*']

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Triggers:**
- On push of tags matching `v*` (e.g., v1.0.0)

**Steps:**
1. Checkout full history
2. Setup Go 1.25
3. Run GoReleaser v2
4. Build and publish to GitHub Releases

---

## Docker Deployment

### Dockerfile Example

```dockerfile
# Build stage
FROM golang:1.25 as builder

WORKDIR /build
COPY . .

RUN go build -o goclaw .

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates bash

COPY --from=builder /build/goclaw /usr/local/bin/

ENV GOCLAW_SERVER=https://goclaw.example.com
ENV GOCLAW_OUTPUT=json

ENTRYPOINT ["goclaw"]
CMD ["agents", "list"]
```

### Build & Push

```bash
# Build image
docker build -t nextlevelbuilder/goclaw-cli:latest .

# Push to registry
docker push nextlevelbuilder/goclaw-cli:latest
```

### Run in Container

```bash
# Interactive
docker run -it \
  -e GOCLAW_SERVER=https://goclaw.example.com \
  -e GOCLAW_TOKEN=your-token \
  nextlevelbuilder/goclaw-cli:latest \
  agents list -o json

# With volume mount for config
docker run -it \
  -v ~/.goclaw:/root/.goclaw \
  nextlevelbuilder/goclaw-cli:latest \
  status
```

---

## Environment Variables Reference

### Core Configuration

| Variable | Purpose | Example | Required |
|----------|---------|---------|----------|
| `GOCLAW_SERVER` | GoClaw server URL | `https://goclaw.example.com` | Yes |
| `GOCLAW_TOKEN` | Authentication token | `sk_prod_abc123xyz` | Yes* |
| `GOCLAW_OUTPUT` | Output format | `json`, `table`, `yaml` | No (default: table) |

*Token stored in OS keyring if using `auth login` interactively.

### Optional Flags

| Flag | Env Var (if any) | Purpose | Example |
|------|------------------|---------|---------|
| `--server` | `GOCLAW_SERVER` | Override server URL | `--server https://staging.example.com` |
| `--token` | `GOCLAW_TOKEN` | Override token | `--token new-token` |
| `--output, -o` | `GOCLAW_OUTPUT` | Output format | `--output json` |
| `--profile` | — | Select config profile | `--profile staging` |
| `--yes, -y` | — | Skip confirmation prompts | `--yes` |
| `--verbose, -v` | — | Enable debug logging | `--verbose` |
| `--insecure` | — | Skip TLS verification | `--insecure` |

---

## Configuration Precedence (Detailed)

**Resolution Order (highest to lowest priority):**

```
1. CLI Flags
   goclaw --server https://custom.com agents list

2. Environment Variables
   export GOCLAW_SERVER=https://staging.com
   goclaw agents list

3. Config File (~/.goclaw/config.yaml)
   active_profile: production
   profiles:
     - name: production
       server: https://goclaw.example.com

4. Profile Defaults
   If --profile staging specified and exists in config

5. Built-in Defaults
   OutputFormat: "table"
   Insecure: false
```

**Example Resolution:**

```bash
# Config file
$ cat ~/.goclaw/config.yaml
active_profile: production
profiles:
  - name: production
    server: https://goclaw.example.com
    output: table

# Environment
$ export GOCLAW_OUTPUT=json

# Command with flag
$ goclaw --output yaml agents list

# Resolution:
Server: https://goclaw.example.com (from config file)
Output: yaml (from CLI flag, overrides env and config)
Token: {from keyring}
```

---

## Troubleshooting

### Connection Issues

**Error: Connection refused**

```bash
# Check server is running
curl https://goclaw.example.com/health

# Check network connectivity
ping goclaw.example.com

# Check firewall rules
sudo iptables -L (Linux)
sudo pf -sr cat /etc/pf.conf (macOS)
```

**Error: Certificate verification failed**

```bash
# For testing only (not production):
goclaw --insecure agents list

# Or set env var:
export GOCLAW_INSECURE=true
```

### Authentication Issues

**Error: Invalid token**

```bash
# Re-login
goclaw auth login --server https://goclaw.example.com --token new-token

# Or logout and re-auth
goclaw auth logout
goclaw auth login --server https://goclaw.example.com --pair
```

**Error: Keyring not available (Linux)**

```bash
# Install secret service
sudo apt-get install gnome-keyring  # Ubuntu
sudo dnf install gnome-keyring      # Fedora

# Or use env var instead
export GOCLAW_TOKEN=your-token
goclaw agents list
```

### Output Issues

**Error: Invalid output format**

```bash
# Valid formats: table, json, yaml
goclaw agents list -o json

# Check current setting
cat ~/.goclaw/config.yaml | grep output
```

### Debugging

**Enable verbose output:**

```bash
goclaw --verbose agents list
# Shows request/response details, error stack traces
```

**Check config:**

```bash
$ cat ~/.goclaw/config.yaml
$ env | grep GOCLAW
```

---

## Upgrade & Rollback

### Upgrade to New Version

```bash
# From release binary
curl -L https://github.com/nextlevelbuilder/goclaw-cli/releases/download/v1.1.0/goclaw_1.1.0_linux_amd64.tar.gz | tar xz
sudo mv goclaw /usr/local/bin/

# Or via go install
go install github.com/nextlevelbuilder/goclaw-cli@latest

# Verify
goclaw version
```

### Rollback to Previous Version

```bash
# Download previous release
curl -L https://github.com/nextlevelbuilder/goclaw-cli/releases/download/v1.0.0/goclaw_1.0.0_linux_amd64.tar.gz | tar xz
sudo mv goclaw /usr/local/bin/

# Verify
goclaw version
```

**Note:** Configuration is backward compatible. No migration needed.

---

## Production Checklist

- [ ] Go 1.25.3+ installed
- [ ] Binary downloaded or built from official source
- [ ] Configuration file created (~/.goclaw/config.yaml)
- [ ] Credentials stored in OS keyring (not as plaintext)
- [ ] Server URL verified (HTTPS in production)
- [ ] Network connectivity tested (`goclaw status`)
- [ ] Automation scripts use env vars (not flags for secrets)
- [ ] Verbose logging disabled (unless debugging)
- [ ] Insecure mode disabled (`--insecure` removed)
- [ ] Backups of config and credentials created

---

## Security Deployment Notes

### Credential Management

**Best Practice:**
```bash
# Use environment variables for CI/CD
export GOCLAW_SERVER=https://goclaw.example.com
export GOCLAW_TOKEN=sk_prod_xxx  # From secrets manager

# Never in config file:
# ❌ config.yaml should NOT contain token
# ✓ config.yaml stores server URL and profile metadata only
# ✓ Credentials in OS keyring or environment
```

### TLS & HTTPS

```bash
# Production (always HTTPS)
goclaw --server https://goclaw.example.com agents list

# Testing/Local (HTTP with --insecure)
goclaw --insecure --server http://localhost:8080 agents list

# Certificate pinning (future enhancement)
# Currently uses system CA bundle
```

### Access Control

```bash
# Restrict config directory permissions
chmod 700 ~/.goclaw
ls -la ~/.goclaw  # Should show drwx------

# Restrict binary permissions
chmod 755 /usr/local/bin/goclaw
ls -la /usr/local/bin/goclaw  # Should show -rwxr-xr-x
```

---

## Performance Tuning

### Connection Reuse

Default HTTP client reuses TCP connections (connection pooling). No tuning needed.

### Timeout Configuration

Currently hardcoded to 30 seconds per request. Override in future versions if needed.

```go
// internal/client/http.go
HTTPClient: &http.Client{
	Timeout: 30 * time.Second,  // Can be made configurable
}
```

---

## Last Updated

- **Date:** 2026-03-15
- **Go Version:** 1.25.3+
- **Status:** Production Ready
- **Release Process:** Automated via GitHub Actions + GoReleaser
