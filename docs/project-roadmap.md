# GoClaw CLI - Project Roadmap

**Last Updated:** 2026-04-15
**Phase Structure:** Legacy Phases 1-9 (bootstrap → CI/CD) + AI-First Expansion Phases 0-5 (2026-04-15)
**Current Status:** Legacy Phases 1-9 ✓ COMPLETE; P0-P4 ✓ COMPLETE; P5 ⏳ DEFERRED to future sprint
**Next Phase:** Phase 10 (Unit Testing & QA) or Phase 5 (Advanced Groups: pair, oauth, packages, users, quota, send)

---

## Project Timeline

### Phase 1: Project Bootstrap ✓ COMPLETE

**Objective:** Set up project structure, tooling, and build infrastructure

**Duration:** ~2 weeks
**Completion Date:** 2026-02-15

**Deliverables:**
- [x] GitHub repository initialized
- [x] Go module setup (go.mod, go.sum)
- [x] Project structure (cmd/, internal/, docs/)
- [x] Makefile with build targets
- [x] .goreleaser.yaml for multi-platform releases
- [x] GitHub Actions workflows (CI and release)
- [x] Initial documentation (README.md)

**Key Files:**
- `go.mod` (1.25.3)
- `Makefile` (build, test, lint, install, clean)
- `.goreleaser.yaml` (release config)
- `.github/workflows/ci.yaml` (build + test)
- `.github/workflows/release.yaml` (multi-platform release)

---

### Phase 2: Core Client & Auth ✓ COMPLETE

**Objective:** Implement HTTP and WebSocket clients, authentication, and configuration

**Duration:** ~2 weeks
**Completion Date:** 2026-02-28

**Deliverables:**
- [x] HTTP client with auth (Bearer token)
- [x] Error handling with wrapped context
- [x] Configuration loading (file → env → flags)
- [x] Multi-profile support
- [x] OS keyring integration
- [x] Device pairing flow
- [x] WebSocket client for streaming
- [x] Output formatters (table, JSON, YAML)

**Key Files:**
- `internal/client/http.go`
- `internal/client/websocket.go`
- `internal/client/auth.go`
- `internal/client/errors.go`
- `internal/config/config.go`
- `internal/output/output.go`
- `internal/tui/prompt.go`
- `cmd/root.go` (global flags)
- `cmd/auth.go` (login, logout, use-context)

---

### Phase 3: Agent & Chat Commands ✓ COMPLETE

**Objective:** Implement agent CRUD operations and interactive chat with streaming

**Duration:** ~2 weeks
**Completion Date:** 2026-03-07

**Deliverables:**
- [x] `goclaw agents list` (table + JSON/YAML output)
- [x] `goclaw agents get <id>`
- [x] `goclaw agents create --name --provider --model`
- [x] `goclaw agents update <id> --field value`
- [x] `goclaw agents delete <id> [-y]`
- [x] `goclaw agents share <id> --user email`
- [x] `goclaw agents delegation-link <id>`
- [x] `goclaw chat <agent-id>` (interactive + streaming)
- [x] Single-shot chat: `goclaw chat <agent-id> -m "message"`
- [x] Pipe input: `echo "message" | goclaw chat <agent-id>`

**Key Files:**
- `cmd/agents.go` (4,048 tokens, 7.2%)
- `cmd/chat.go` (300+ lines)
- `cmd/helpers.go` (shared utilities)

---

### Phase 4: Session & Skill Commands ✓ COMPLETE

**Objective:** Implement session management and skill upload/management

**Duration:** ~1.5 weeks
**Completion Date:** 2026-03-10

**Deliverables:**
- [x] `goclaw sessions list`
- [x] `goclaw sessions get <id>`
- [x] `goclaw sessions delete <id>`
- [x] `goclaw sessions reset <id>`
- [x] `goclaw sessions label <id>`
- [x] `goclaw skills list`
- [x] `goclaw skills upload <file>`
- [x] `goclaw skills delete <id>`
- [x] Grant/revoke skill access

**Key Files:**
- `cmd/sessions.go` (200+ lines)
- `cmd/skills.go` (2,635 tokens, 4.7%)

---

### Phase 5: MCP, Provider & Tool Commands ✓ COMPLETE

**Objective:** Implement MCP server management, LLM provider CRUD, and custom tools

**Duration:** ~2 weeks
**Completion Date:** 2026-03-12

**Deliverables:**
- [x] `goclaw mcp list` (MCP servers)
- [x] `goclaw mcp add`
- [x] `goclaw mcp remove`
- [x] `goclaw mcp grants` (access control)
- [x] `goclaw mcp access-requests`
- [x] `goclaw providers list` (LLM providers)
- [x] `goclaw providers create`
- [x] `goclaw providers update`
- [x] `goclaw providers delete`
- [x] `goclaw providers models` (list available models)
- [x] `goclaw providers verify` (test connection)
- [x] `goclaw tools list` (custom tools)
- [x] `goclaw tools invoke <id>`
- [x] `goclaw tools delete <id>`

**Key Files:**
- `cmd/mcp.go` (2,940 tokens, 5.2%)
- `cmd/providers.go` (200+ lines)
- `cmd/tools.go` (180+ lines)

---

### Phase 6: Team, Channel & Cron Commands ✓ COMPLETE

**Objective:** Implement team management, channels, and scheduled jobs

**Duration:** ~2 weeks
**Completion Date:** 2026-03-13

**Deliverables:**
- [x] `goclaw teams list`
- [x] `goclaw teams create`
- [x] `goclaw teams members` (list, add, remove)
- [x] `goclaw teams task-board`
- [x] `goclaw teams workspace`
- [x] `goclaw channels list`
- [x] `goclaw channels contacts`
- [x] `goclaw channels pending-messages`
- [x] `goclaw cron list` (scheduled jobs)
- [x] `goclaw cron create`
- [x] `goclaw cron update`
- [x] `goclaw cron delete`
- [x] `goclaw cron trigger` (run manually)
- [x] `goclaw cron history` (execution history)

**Key Files:**
- `cmd/teams.go` (4,075 tokens, 7.3% — largest file)
- `cmd/channels.go` (200+ lines)
- `cmd/cron.go` (220+ lines)

---

### Phase 7: Trace, Memory & Utility Commands ✓ COMPLETE

**Objective:** Implement LLM trace viewer, memory documents, and utility operations

**Duration:** ~1.5 weeks
**Completion Date:** 2026-03-14

**Deliverables:**
- [x] `goclaw traces list` (LLM traces)
- [x] `goclaw traces export` (export traces)
- [x] `goclaw memory list` (memory documents)
- [x] `goclaw memory search` (semantic search)
- [x] `goclaw memory upsert` (create/update)
- [x] `goclaw knowledge-graph entities`
- [x] `goclaw knowledge-graph links`
- [x] `goclaw knowledge-graph query`
- [x] `goclaw usage summary` (analytics)
- [x] `goclaw usage cost-breakdown`
- [x] `goclaw activity` (audit log)

**Key Files:**
- `cmd/traces.go` (180+ lines)
- `cmd/memory.go` (180+ lines)

---

### Phase 8: Config, Logs, Storage & Admin Commands ✓ COMPLETE

**Objective:** Implement config management, log streaming, file browser, and admin operations

**Duration:** ~2 weeks
**Completion Date:** 2026-03-14

**Deliverables:**
- [x] `goclaw config get` (server config)
- [x] `goclaw config apply` (apply config)
- [x] `goclaw config patch` (partial update)
- [x] `goclaw logs` (real-time log streaming with -f)
- [x] `goclaw storage list` (workspace file browser)
- [x] `goclaw storage download`
- [x] `goclaw approvals list` (execution approvals)
- [x] `goclaw approvals approve`
- [x] `goclaw approvals deny`
- [x] `goclaw delegations` (delegation history)
- [x] `goclaw credentials get/set` (CLI credential store)
- [x] `goclaw tts synthesize` (text-to-speech)
- [x] `goclaw tts list-voices`
- [x] `goclaw media upload/download`
- [x] `goclaw admin` (admin-specific operations)
- [x] `goclaw status` (server health)

**Key Files:**
- `cmd/config_cmd.go` (150+ lines)
- `cmd/logs.go` (120+ lines)
- `cmd/storage.go` (150+ lines)
- `cmd/admin.go` (2,949 tokens, 5.3%)

---

### Phase 9: Testing, CI/CD & Release ✓ COMPLETE

**Objective:** Set up comprehensive testing, CI/CD automation, and release process

**Duration:** ~2 weeks
**Completion Date:** 2026-03-15

**Deliverables:**
- [x] GitHub Actions CI pipeline (build, vet, test)
- [x] GoReleaser configuration (multi-platform builds)
- [x] Release workflow (automated on tag push)
- [x] Artifact generation (linux, darwin, windows; amd64, arm64)
- [x] Checksum generation
- [x] Documentation (README, API reference)
- [x] Version injection via ldflags
- [x] Smoke tests (basic command execution)

**Key Files:**
- `.github/workflows/ci.yaml`
- `.github/workflows/release.yaml`
- `.goreleaser.yaml`
- `Makefile` (build, test, lint, install)

**Artifacts:**
- goclaw_X.X.X_darwin_amd64.tar.gz
- goclaw_X.X.X_darwin_arm64.tar.gz
- goclaw_X.X.X_linux_amd64.tar.gz
- goclaw_X.X.X_linux_arm64.tar.gz
- goclaw_X.X.X_windows_amd64.zip
- goclaw_X.X.X_windows_arm64.zip

---

## Phase 0: AI Ergonomics Foundation ✓ COMPLETE

**Objective:** Implement exit codes, TTY-aware output, structured error handling, and streaming reconnect for AI/automation consumers

**Duration:** ~3 days
**Completion Date:** 2026-04-15

**Deliverables:**
- [x] Exit code mapping (server codes → 0-6)
- [x] TTY detection + format auto-resolution
- [x] Structured error output (JSON envelope with code/message/details)
- [x] FollowStream with exponential backoff reconnect
- [x] Central error handler in cmd.Execute()
- [x] --quiet flag for non-TTY contexts
- [x] Updated CHANGELOG, README, CLAUDE.md

**Status:** COMPLETE with 1 HIGH finding (H1: handler error retry semantics) + 3 MEDIUM findings (M1: MaxRetries=0 override, M3: --output validation)

**Key Files:**
- `internal/output/exit.go`, `error.go`, `tty.go`
- `internal/client/follow.go`
- `cmd/root.go` (error handler)
- `CHANGELOG.md` (breaking change doc)

**Note:** Phase 0 is foundational AI ergonomics. Recommended fixes for H1/M1/M3 should be addressed before merging Phase 1+ features.

---

## Phases 1-4: AI-First CLI Expansion (2026-04-15)

**Context:** After completing legacy Phases 1-9 (bootstrap through CI/CD), this expansion (P0-P5) adds AI-agent-centric ergonomics and advanced CLI features.

### P1: Admin/Ops Foundation ✓ COMPLETE

**Deliverables:**
- [x] `tenants` group (CRUD, user membership)
- [x] `heartbeat` group (agent health, monitoring, logs with WS streaming)
- [x] `system-configs` group (key-value server configuration)
- [x] `edition` group (server edition info, no auth)
- [x] `config` extensions (permissions CRUD via HTTP + WS)

**Key Files:** `cmd/tenants.go`, `cmd/heartbeat.go`, `cmd/system_configs.go`, `cmd/edition.go`, `cmd/config_cmd.go` extensions

**Status:** COMPLETE; 1 critical fix applied (config permissions revoke gated with `tui.Confirm`)

---

### P2: Migration (Backup/Restore + Export/Import) ✓ COMPLETE

**Deliverables:**
- [x] `backup` group (system/tenant, preflight, signed download, S3 integration)
- [x] `restore` group (system/tenant with typed confirmation safety)
- [x] Export/import for agents, teams, skills, mcp (preview-first)
- [x] Signed download flow (unauthenticated binary via token)
- [x] Multipart streaming upload (no RAM buffering)

**Key Files:** `cmd/backup.go`, `cmd/backup_s3.go`, `cmd/restore.go`, `cmd/*_export.go`, `internal/client/signed_download.go`, `internal/client/multipart_upload.go`

**Status:** COMPLETE; 4 critical fixes applied (S3 masking, URL escaping, error propagation, MkdirAll)

---

### P3: Vault (Knowledge Vault / RAG) ✓ COMPLETE

**Deliverables:**
- [x] `vault` group (documents CRUD, links management, upload, search, tree view, graph, enrichment)
- [x] Document metadata + links (knowledge graph edges)
- [x] Streaming multipart file upload
- [x] Semantic + full-text search (RAG)
- [x] Directory tree browser (TTY: ASCII, piped: JSON)
- [x] Graph visualization (JSON or Graphviz DOT format)
- [x] Background enrichment pipeline control

**Key Files:** `cmd/vault.go`, `cmd/vault_documents.go`, `cmd/vault_links.go`, `cmd/vault_upload.go`, `cmd/vault_enrichment.go`, `internal/output/tree.go`

**Status:** COMPLETE; 2 critical fixes applied (documents create --file handling, URL query escaping)

**Deferred:** vault_documents.go split (303 LoC overage; refactoring only, low priority)

---

### P4: Agent Lifecycle + Chat + Teams + Memory KG ✓ COMPLETE

**Deliverables:**
- [x] `agents` extensions (lifecycle: wake/wait/identity; admin ops; sharing; instances; links; evolution; episodic; v3-flags; misc)
- [x] `chat` extensions (history, inject, session-status — AI-critical MAX POLISH)
- [x] `teams` extensions (members, tasks: CRUD + review + advanced + delete-bulk, workspace, events streaming, scopes)
- [x] `memory kg` subsystem (entities CRUD, traversal, stats, graph, deduplication, legacy compat)
- [x] `memory` extensions (index, chunks, global documents)
- [x] AI-critical commands with ≥80% test coverage

**Key Files:** 26 new files (agents_*, chat_ai_commands.go, teams_*, memory_kg_*) + 4 modified (agents.go trimmed to 196 LoC, chat.go to 214 LoC, teams.go to 150 LoC, memory.go to 147 LoC)

**Status:** COMPLETE with modularization; 2 critical fixes applied (strict JSON validation, WS cleanup on timeout)

**Note:** chat.go and chat_ai_commands.go are 214 LoC (14 lines over limit) — overage is entirely docstrings for AI-critical help text (MAX POLISH requirement)

---

### P5: Advanced Groups (pair, oauth, packages, users, quota, send) ⏳ DEFERRED

**Status:** Not started; deferred to future sprint

**Scope:**
- pair — Device pairing CLI flow
- oauth — OAuth authorization endpoints
- packages — Package management (skills, tools, etc.)
- users — User account management
- quota — Usage quota/limits
- send — Message broadcasting

**Also includes subcommand extensions:**
- channels pending extensions
- mcp reconnect
- skills install-dep
- providers verify-embedding / claude-cli
- tools builtin tenant-config
- admin credentials extensions

---

## Phase 10: Testing & Quality Assurance (PLANNED)

**Objective:** Achieve high code coverage with table-driven tests and integration tests

**Status:** Not Started
**Target Duration:** 2-3 weeks
**Target Completion:** 2026-04-15

**Planned Deliverables:**
- [ ] Unit tests for all command handlers (table-driven)
- [ ] Mock HTTP responses for testing
- [ ] Integration tests (auth flow, agent CRUD, chat)
- [ ] Error scenario testing
- [ ] Config loading precedence tests
- [ ] Output formatting tests (table, JSON, YAML)
- [ ] Code coverage >80%
- [ ] Race detector passes: `go test -race ./...`

**Estimated Effort:** 40-60 hours

**Key Areas:**
- `cmd/*_test.go` (command handlers)
- `internal/client/*_test.go` (HTTP, WebSocket)
- `internal/config/*_test.go` (configuration loading)
- `internal/output/*_test.go` (formatters)

---

## Phase 11: Shell Completions (PLANNED)

**Objective:** Generate shell completions for bash, zsh, and fish

**Status:** Not Started
**Target Duration:** 1 week
**Target Completion:** 2026-04-30

**Planned Deliverables:**
- [ ] Bash completion script
- [ ] Zsh completion script
- [ ] Fish completion script
- [ ] Installation instructions
- [ ] Dynamic completion support (agents, skills, etc.)

**Implementation:**
- Use Cobra's built-in completion generation
- Include in release archives
- Installation via package managers

---

## Phase 12: Homebrew Tap (PLANNED)

**Objective:** Publish GoClaw CLI to Homebrew for easy macOS installation

**Status:** Not Started
**Target Duration:** 1 week
**Target Completion:** 2026-05-15

**Planned Deliverables:**
- [ ] Create Homebrew tap repository
- [ ] Formula for goclaw-cli
- [ ] Automated updates on release
- [ ] Installation via: `brew tap nextlevelbuilder/goclaw` + `brew install goclaw-cli`

**Tap Repository:**
```
https://github.com/nextlevelbuilder/homebrew-goclaw
```

---

## Phase 13: Man Pages & Advanced Documentation (PLANNED)

**Objective:** Create comprehensive man pages and advanced user guides

**Status:** Not Started
**Target Duration:** 1-2 weeks
**Target Completion:** 2026-06-01

**Planned Deliverables:**
- [ ] Man page for `goclaw(1)` (main command)
- [ ] Individual man pages for command groups
- [ ] Advanced usage guide
- [ ] Integration examples (CI/CD, scripts)
- [ ] Troubleshooting guide expansion

**Distribution:**
- Include in release archives
- Install via `goclaw completion install` or package manager

---

## Completed Features Summary (Phase 1-9)

| Feature | Status | Notes |
|---------|--------|-------|
| **28 Command Groups** | ✓ Complete | All 28 command groups implemented |
| **Full API Coverage** | ✓ Complete | 100% of dashboard features |
| **Dual Mode** | ✓ Complete | Interactive + automation |
| **Multi-Profile** | ✓ Complete | Config file + CLI override |
| **Output Formats** | ✓ Complete | Table, JSON, YAML |
| **WebSocket Streaming** | ✓ Complete | Chat, logs, traces |
| **Authentication** | ✓ Complete | Keyring, device pairing, token |
| **Error Handling** | ✓ Complete | Wrapped with context |
| **Build Automation** | ✓ Complete | Makefile, GoReleaser, CI/CD |
| **Multi-Platform** | ✓ Complete | Linux, macOS, Windows; amd64, arm64 |
| **Documentation** | ✓ In Progress | PDR, architecture, deployment guide |

---

## Metrics & KPIs

### Phase 1-9 Metrics (Actual)

| Metric | Target | Achieved | Status |
|--------|--------|----------|--------|
| Command Coverage | 100% (28/28) | 28/28 | ✓ Complete |
| Build Time | <5s | ~2s | ✓ Exceeds |
| Binary Size | <15MB | ~8MB | ✓ Exceeds |
| Test Coverage | >80% (Phase 10) | TBD | In Progress |
| Platforms | 6 (3 OS × 2 arch) | 6 | ✓ Complete |
| Code Quality | 0 vet warnings | 0 | ✓ Complete |

### Phase 10+ Targets

| Metric | Target | Current |
|--------|--------|---------|
| Test Coverage | >80% | 0% (TBD) |
| Integration Tests | 20+ critical paths | 0 |
| Documentation Pages | 10+ guides | 5 |
| Shell Completions | 3 (bash, zsh, fish) | 0 |

---

## Dependency Management

### Core Dependencies (Stable)

| Package | Version | Status |
|---------|---------|--------|
| cobra | v1.10.2 | Stable, active maintenance |
| gorilla/websocket | v1.5.3 | Stable, widely used |
| golang.org/x/term | v0.41.0 | Stable (stdlib subset) |
| yaml.v3 | v3.0.1 | Stable, mature |

### Update Strategy

- **Major Updates:** Evaluate before upgrade, test thoroughly
- **Minor Updates:** Apply if compatible
- **Security Patches:** Apply immediately

**Current Go Version:** 1.25.3+ (latest stable)

---

## Risk Assessment

### Phase 10 (Testing)

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Low test coverage | Medium | Medium | Allocate dedicated testing phase |
| Flaky tests | Low | Medium | Use table-driven, deterministic tests |
| Integration issues | Low | High | Test with real server instance |

### Phase 11-12 (Future)

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Shell completion complexity | Medium | Low | Use Cobra's built-in support |
| Homebrew formula approval | Low | Low | Leverage existing tap infrastructure |

---

## Success Criteria

### Phase 1-9 (Completed)
- [x] All 28 command groups implemented and functional
- [x] Full API coverage verified
- [x] Dual mode (interactive + automation) operational
- [x] Multi-profile support working
- [x] WebSocket streaming functional
- [x] CI/CD automated (build, test, release)
- [x] Multi-platform binaries generated
- [x] No security vulnerabilities
- [x] Project documentation started

### Phase 10 (Planned)
- [ ] Unit test coverage >80%
- [ ] Integration tests for critical paths
- [ ] All edge cases documented
- [ ] No flaky tests (100% deterministic)
- [ ] Performance benchmarks established

### Phase 11-12 (Planned)
- [ ] Shell completions working in bash/zsh/fish
- [ ] Homebrew installation successful
- [ ] First-time user can install and authenticate in <5 minutes

---

## Release Timeline

| Version | Date | Phase | Status |
|---------|------|-------|--------|
| v1.0.0 | 2026-03-15 | 1-9 | Released |
| v1.1.0 | 2026-04-15 (est) | 10 | Planning |
| v1.2.0 | 2026-05-15 (est) | 11-12 | Planning |
| v2.0.0 | TBD | Major features | Not planned |

---

## Open Questions & Blockers

### Current
- None identified; all Phases 1-9 complete

### Phase 10+
- What test infrastructure should be used? (mock server, real server, containers?)
- Should completion generation be automated in CI/CD?
- What's the priority: Homebrew tap or man pages?

---

## Communication & Stakeholder Updates

**Status Report Frequency:** Weekly
**Last Update:** 2026-03-15
**Next Update:** 2026-03-22

---

## Appendix: Command Inventory

### All 28 Command Groups

1. **auth** — Authentication & profiles
2. **agents** — Agent CRUD + sharing
3. **chat** — Interactive + streaming chat
4. **sessions** — Session management
5. **skills** — Skill upload + management
6. **mcp** — MCP server management
7. **providers** — LLM provider CRUD
8. **tools** — Custom tool management
9. **cron** — Scheduled job management
10. **teams** — Team management
11. **channels** — Channel management
12. **traces** — LLM trace viewing
13. **memory** — Memory document management
14. **knowledge-graph** — Entity extraction & linking
15. **usage** — Usage analytics
16. **config** — Server config management
17. **logs** — Real-time log streaming
18. **storage** — Workspace file browser
19. **approvals** — Execution approvals
20. **delegations** — Delegation history
21. **credentials** — CLI credential store
22. **tts** — Text-to-speech
23. **media** — Media upload/download
24. **activity** — Audit log
25. **admin** — Admin operations
26. **status** — Server health check
27. **version** — Version display

Plus 1 root command: **goclaw**

**Total: 28 command groups**

---

## Last Updated

- **Date:** 2026-03-15
- **By:** Documentation Team
- **Status:** Production Ready (Phases 1-9 Complete)
- **Next Review:** 2026-04-15
