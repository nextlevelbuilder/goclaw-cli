---
phase: 9
title: Testing, CI/CD & Release
status: planned
priority: high
effort: M
depends_on: [phase-03, phase-04, phase-05, phase-06, phase-07, phase-08]
---

# Phase 9 — Testing, CI/CD & Release

## Overview
Unit tests, integration tests, CI/CD pipeline, and cross-platform release automation.

## Requirements

### Testing
- Unit tests for all internal packages (client, config, output, tui)
- Integration tests against a running GoClaw server (test container or mock)
- Table-driven tests for command parsing and flag validation
- Test coverage > 70%

### CI/CD (GitHub Actions)
- On PR: lint, build, test
- On tag: build + release via GoReleaser
- Matrix: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64

### Release
- GoReleaser for cross-platform binaries
- GitHub Releases with checksums
- Homebrew tap formula (optional)
- Docker image (optional)

## Implementation Steps

1. `tests/` directory with integration test suite
2. Unit tests co-located with packages (`*_test.go`)
3. `.github/workflows/ci.yaml` — PR checks (lint, build, test)
4. `.github/workflows/release.yaml` — Tag-triggered release
5. `.goreleaser.yaml` — Multi-platform build config
6. Mock server for integration tests (httptest + gorilla/websocket)
7. Test fixtures for API responses

### Test Categories
- **Client tests:** HTTP request building, auth injection, error parsing
- **WebSocket tests:** Frame encoding/decoding, reconnection logic
- **Output tests:** Table rendering, JSON formatting
- **Command tests:** Flag parsing, argument validation, help text
- **Integration tests:** Full command execution against mock server

## Related Code Files
- Create: `.github/workflows/ci.yaml`, `.github/workflows/release.yaml`
- Create: `tests/integration_test.go`
- Create: `internal/client/http_test.go`, `websocket_test.go`
- Create: `internal/output/formatter_test.go`

## Todo
- [ ] Unit tests for HTTP client
- [ ] Unit tests for WebSocket client
- [ ] Unit tests for output formatters
- [ ] Unit tests for config loader
- [ ] Integration test framework with mock server
- [ ] CI workflow (lint + build + test)
- [ ] Release workflow (GoReleaser)
- [ ] GoReleaser config for 5 platforms
- [ ] README with installation instructions

## Success Criteria
- `go test ./...` passes with > 70% coverage
- CI pipeline green on PRs
- Tagged release produces binaries for linux/mac/windows (amd64+arm64)
- `goclaw version` shows correct version from build tags
- README has install instructions for all platforms

## Risk Assessment
- Integration tests need running server or reliable mock
- Windows path handling may differ — test on Windows CI runner
