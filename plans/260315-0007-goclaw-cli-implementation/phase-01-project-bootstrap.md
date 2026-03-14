---
phase: 1
title: Project Bootstrap
status: completed
priority: critical
effort: S
---

# Phase 1 — Project Bootstrap

## Overview
Initialize Go module, repo, project structure, Makefile, and CI scaffolding.

## Context Links
- [Plan Overview](plan.md)
- [GoClaw Server](../../README.md) (reference)

## Requirements
- Go 1.25 module with proper module path
- Cobra CLI skeleton with root command
- Private GitHub repo on `nextlevelbuilder` org
- Makefile with build/test/lint targets
- GoReleaser config for cross-platform binaries
- Basic README

## Implementation Steps

1. Create private repo: `gh repo create nextlevelbuilder/goclaw-cli --private`
2. Initialize Go module: `go mod init github.com/nextlevelbuilder/goclaw-cli`
3. Create directory structure:
   ```
   cmd/           # Cobra commands
   internal/      # Private packages
     client/      # HTTP + WS client
     config/      # Config loader
     output/      # Formatters
     tui/         # Interactive prompts
   ```
4. Create `main.go` entry point
5. Create `cmd/root.go` with global flags:
   - `--server` / `GOCLAW_SERVER` — server URL
   - `--token` / `GOCLAW_TOKEN` — auth token
   - `--output` / `-o` — output format (table|json|yaml)
   - `--yes` / `-y` — skip confirmations
   - `--insecure` — skip TLS verify
   - `--verbose` / `-v` — debug logging
6. Create `cmd/version.go` — version/build info
7. Create `Makefile`:
   - `build`, `test`, `lint`, `install`, `release`
8. Create `.goreleaser.yaml` for multi-platform builds
9. Add `.gitignore`, `LICENSE` (MIT)
10. Initial commit and push

## Related Code Files
- Create: `main.go`, `cmd/root.go`, `cmd/version.go`
- Create: `Makefile`, `.goreleaser.yaml`, `.gitignore`, `README.md`
- Create: `go.mod`, `go.sum`

## Todo
- [x] Create GitHub repo
- [x] Init Go module
- [x] Create directory structure
- [x] Implement root command with global flags
- [x] Implement version command
- [x] Create Makefile
- [x] Create GoReleaser config
- [x] Initial commit & push

## Success Criteria
- `go build ./...` compiles
- `goclaw --help` shows usage
- `goclaw version` shows build info
- Repo exists on GitHub with proper structure
