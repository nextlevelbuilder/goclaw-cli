# Code Review: GoClaw CLI Feature Parity Update

**Reviewer:** code-reviewer
**Date:** 2026-03-26
**Branch:** claude/agitated-shirley
**Score: 8/10**

## Scope

- **Modified files:** 16, **New files:** 30, **Total LOC changed:** ~1900 (268 added, 1633 removed net)
- **Focus:** Full branch diff against main
- **Build:** PASS (`go build ./...`)
- **Vet:** PASS (`go vet ./...`)
- **Tests:** PASS (all packages)

## Overall Assessment

Strong implementation. Consistent Cobra patterns across all 30+ new command files. The modularization is well-executed -- large files (teams 363 lines, agents 333 lines) were split into focused sub-files all under 200 lines. Tenant header propagation is correctly wired through all HTTP methods. The `tools custom` commands were fully replaced with `tools builtin`. Error handling is consistent throughout. A few issues need attention.

---

## Critical Issues

### C1. `skillsTenantConfigCmd` is orphaned (dead code)

**File:** `cmd/skills_config.go`
**Impact:** The `skills tenant-config set` and `skills tenant-config delete` commands are defined and wired to a parent `skillsTenantConfigCmd`, but that parent is never added to `skillsCmd`. These commands are invisible to users.

**Fix:** Add `skillsCmd.AddCommand(skillsTenantConfigCmd)` at the end of `init()` in `cmd/skills_config.go`.

---

## High Priority

### H1. Missing `url.PathEscape()` on path parameters (30+ locations)

**Files:** `agents.go`, `channels.go`, `channels_writers.go`, `channels_pending.go`, `channels_contacts.go`, `providers.go`, `providers_crud.go`, `skills.go`, `skills_files.go`, `skills_config.go`, `sessions.go`, `mcp.go`, `traces.go`, `admin.go`, `admin_media.go`, `admin_credentials.go`, `agents_ops.go`, `agents_links.go`, `agents_instances.go`, `memory.go`

**Impact:** If any user-supplied ID contains `/`, `%`, `?`, or `#` characters, the URL path will break, potentially causing incorrect API calls or routing to wrong endpoints. While UUIDs are safe, some endpoints take user-defined strings (agent keys, file paths, skill slugs).

**Pattern of concern:**
```go
c.Get("/v1/agents/" + args[0])           // unsafe
c.Get("/v1/agents/" + url.PathEscape(args[0]))  // safe
```

**Already correct in:** `tenants.go`, `system_config.go`, `contacts.go`, `knowledge_graph.go`, `tools.go` (builtin commands). These files show the team knows the pattern -- it just wasn't applied consistently.

**Fix:** Apply `url.PathEscape()` to all path-interpolated `args[N]` values. Also apply `url.QueryEscape()` where values are interpolated into query strings (e.g., `channels_contacts.go:32`, `memory.go:23`).

### H2. Variable shadowing in `configApplyCmd`

**File:** `cmd/config_cmd.go:53`
**Impact:** `var cfg map[string]any` shadows the package-level `cfg *config.Config`. Within this closure scope it's technically fine, but any future modification that references the outer `cfg` after this line would silently use the wrong variable. This is a maintenance hazard.

**Fix:** Rename to `cfgBody` or `configData`.

---

## Medium Priority

### M1. `cronListCmd` creates unused HTTP client

**File:** `cmd/cron.go:16-33`
**Impact:** The function creates an HTTP client (`c, err := newHTTP()`) then only uses WebSocket. The `_ = c` on line 31 is dead code from an incomplete fallback. If auth fails, both HTTP and WS errors fire (wasting one network attempt).

**Fix:** Remove the HTTP client creation. If fallback is needed, implement it properly or document the intent for future work.

### M2. `mediaUploadCmd` is a stub

**File:** `cmd/admin_media.go:13-25`
**Impact:** The upload command authenticates but prints a message telling the user to use the HTTP API directly. This is confusing UX -- users expect the command to work.

**Fix:** Either implement multipart upload (the `PostRaw` method on HTTPClient supports it) or mark the command as hidden/deprecated with a clear error message pointing to the API docs.

### M3. `memory list` and `channels contacts resolve` use raw string concatenation for query params

**Files:** `cmd/memory.go:23`, `cmd/channels_contacts.go:32`
**Impact:** Query parameter values are not URL-encoded. `?user_id=foo+bar` or `?ids=a&b=c` would produce malformed URLs.

**Pattern:**
```go
path += "?user_id=" + v          // unsafe
path += "?user_id=" + url.QueryEscape(v)  // safe
```

### M4. Files over 200 lines

Per project conventions, code files should be under 200 lines:
- `cmd/mcp.go` (339 lines) -- could split into `mcp_servers.go`, `mcp_grants.go`, `mcp_requests.go`
- `cmd/chat.go` (293 lines) -- could extract `chatInteractive` into `chat_interactive.go`
- `cmd/cron.go` (268 lines) -- could split into `cron_jobs.go` and `cron_runs.go`
- `cmd/auth.go` (254 lines) -- pre-existing, not new in this PR
- `cmd/memory.go` (203 lines) -- marginally over, has both memory + kgCmd definition

### M5. `channels instances list` filter uses raw concatenation

**File:** `cmd/channels.go:25`
```go
path += "?channel_type=" + v  // not URL-encoded
```

**Fix:** Use `url.Values{}` + `.Encode()` like other commands do (e.g., traces, usage, delegations).

---

## Low Priority (Suggestions)

### L1. Inconsistent table output pattern

Some commands output raw JSON for non-table mode and formatted table for table mode (good), while others just call `printer.Print(unmarshalList(data))` for everything. This means table mode for those commands will render a raw map rather than a formatted table.

Affected: `channelsContactsListCmd`, `channelsContactsResolveCmd`, `channelsPendingListCmd`, `channelsPendingRetryCmd` -- all just dump raw data regardless of output format.

Not blocking, but inconsistent with the pattern established in `tenants`, `agents`, `traces`, etc.

### L2. `skills rescan-deps` has optional `--skill-id` but isn't marked required

**File:** `cmd/skills_files.go:99`
The `rescan-deps` command takes `--skill-id` as a flag but it's not marked required. If omitted, the API call sends `{}` body which may or may not be valid server-side.

### L3. `agents get` and `agents delete` accept IDs without `url.PathEscape`

While agent IDs are typically UUIDs (safe), the pattern should be consistent with other commands like `tenants get` which does use `url.PathEscape`.

---

## Positive Observations

1. **Excellent modularization** -- teams split from 363 lines into 5 files (teams.go, teams_members.go, teams_extra.go, teams_tasks.go, teams_tasks_actions.go, teams_workspace.go), all well under 200 lines
2. **Consistent Cobra patterns** -- all commands use `RunE`, proper flag registration in `init()`, required flags marked
3. **`tui.Confirm` on destructive ops** -- every delete command checks `cfg.Yes` for automation mode
4. **Tenant header propagation** -- correctly added to `do()`, `PostRaw()`, and `GetRaw()` in http.go, covering all HTTP methods
5. **Clean dead code removal** -- `tools custom` CRUD (create, update, delete, share, grant, configure) fully replaced with `tools builtin`
6. **Good test coverage** -- `cmd_test.go` validates all 32 expected root commands are registered
7. **Config never persists tenant-id** -- `yaml:"-"` tag prevents accidental serialization

---

## Recommended Actions (Priority Order)

1. **[CRITICAL]** Add `skillsCmd.AddCommand(skillsTenantConfigCmd)` in `cmd/skills_config.go` init()
2. **[HIGH]** Sweep all `cmd/*.go` for `+ args[N]` in URL paths, wrap with `url.PathEscape()`
3. **[HIGH]** Fix query param encoding in `memory.go`, `channels_contacts.go`, `channels.go`
4. **[MEDIUM]** Rename `cfg` variable in `config_cmd.go:53` to avoid shadowing
5. **[MEDIUM]** Remove dead HTTP client from `cronListCmd`
6. **[MEDIUM]** Split `mcp.go` (339 lines) into sub-files per convention
7. **[LOW]** Add table format to commands that currently only dump raw JSON

---

## Metrics

- **Build:** PASS
- **Vet:** PASS
- **Tests:** PASS (4/4 packages)
- **Type Coverage:** N/A (Go is statically typed)
- **Files > 200 lines:** 5 (mcp.go, chat.go, cron.go, auth.go, memory.go)
- **Commands registered:** 32 root commands verified by test
- **Orphaned commands:** 1 (skillsTenantConfigCmd)
- **Missing url.PathEscape:** ~30 locations across ~20 files
- **Missing url.QueryEscape:** 3 locations

---

**Status:** DONE_WITH_CONCERNS
**Summary:** Solid feature parity implementation with consistent patterns. One orphaned command group (skills tenant-config) and widespread missing URL encoding are the key issues.
**Concerns:** The `url.PathEscape` gap affects ~30 call sites. While most IDs are likely UUIDs, this is a correctness and security concern for endpoints accepting user-defined strings.
