# Phase 3 — Vault: Code Review

**Reviewer:** code-reviewer
**Date:** 2026-04-15
**Scope:** cmd/vault*.go, internal/output/tree.go, all vault *_test.go
**Build/vet/test status:** all green (`go vet`, `go test ./...` pass)

---

## Scope

| File | LoC | Note |
|---|---|---|
| cmd/vault.go | 191 | OK |
| cmd/vault_documents.go | 303 | **OVER 200 LoC limit** |
| cmd/vault_links.go | 106 | OK |
| cmd/vault_upload.go | 98 | OK |
| cmd/vault_enrichment.go | 53 | OK |
| cmd/vault_multipart_helper.go | 48 | OK |
| internal/output/tree.go | 58 | OK |

Tests: 5 test files, all green. Coverage ≥ 60% per report.

---

## Overall Assessment

Solid, idiomatic Go. The streaming multipart upload is well engineered (pipe-based, content-type captured before goroutine, `CloseWithError` propagation). `tui.Confirm` is correctly invoked for all destructive ops — non-interactive without `--yes` = refuse. DOT transform is defensive (skips incomplete edges, handles empty graph). Tree rendering is clean.

**However**: one **P0 bug in `documents create`** silently drops `--file` content, plus two **high-priority** URL injection / escaping issues. Plus the 200 LoC rule violation.

---

## Critical Issues

### C1 — `documents create --file` silently discards file content  (BLOCKING)
**File:** `cmd/vault_documents.go:98-108`

```go
if fileVal != "" {
    content, err := readFileOrStdin(fileVal)
    if err != nil {
        return err
    }
    if docPath == "" {
        docPath = fileVal
    }
    _ = content // content written to path on server side via upload; create just registers metadata
}
```

Problem:
1. `--file` reads the content, then **throws it away**. User intent is clearly "use this file's content as the document body" — silently discarded with zero signal to the user.
2. `buildBody(...)` never gets a `content` key. Server sees only `{path, title, doc_type, scope}` — document created but body is empty.
3. Flag `--content` exists and is also never passed through — `contentVal` is read, mutex-checked against `fileVal`, but never placed into the body.
4. `--path` is `MarkFlagRequired` (line 288), so the `if docPath == ""` recovery at line 104 is dead code: cobra rejects the command before `RunE` runs.
5. The comment "content written to path on server side via upload" is misleading — the `create` endpoint is not an upload path; `vault upload` is.

**Impact:** Every user who does `vault documents create --title=X --path=... --file=doc.md` gets a silent empty document. Data-loss pattern, highest severity.

**Fix:** Two directions, pick one:
- (a) If the create endpoint accepts inline `content`: add it to `buildBody("content", contentVal)` when `--content` is set, and set `contentVal = content` when `--file` is used. Remove the dead `docPath == ""` block.
- (b) If create is metadata-only: remove `--file` / `--content` flags entirely and point users at `vault upload`. Update help text accordingly.

Check server `vault_handler_documents.go` create handler for the contract, then align.

---

### C2 — URL query values not escaped  (HIGH, injection-class)
**Files:**
- `cmd/vault.go:31` — `url += "?path=" + path`
- `cmd/vault_documents.go:35` — `url += "&q=" + q`

Both directly concatenate user input into URL query without `url.QueryEscape`.

**Break cases:**
- `vault tree --path="notes/one two"` → produces `?path=notes/one two` (space is illegal in URLs; some servers 400, some silently truncate).
- `vault documents list --q="a&b"` → `&q=a&b` — `&b` becomes an additional query parameter.
- `--q="foo#bar"` → fragment injection, `#bar` stripped from query.
- Harder: `--q="foo&admin=true"` lets a user inject extra query params.

**Fix:** use `net/url`:

```go
v := url.Values{}
v.Set("limit", strconv.Itoa(limit))
v.Set("offset", strconv.Itoa(offset))
if q != "" { v.Set("q", q) }
url := "/v1/vault/documents?" + v.Encode()
```

---

### C3 — `vault_documents.go` exceeds 200 LoC hard limit (303 LoC)
**File:** `cmd/vault_documents.go`

Project rule in `CLAUDE.md` / `development-rules.md`: files over 200 LoC must be modularized. Current 303. Phase spec itself budgeted ~180 LoC (line 86 of phase-03 plan).

**Fix:** extract to separate files, e.g.:
- `cmd/vault_documents.go` → commands only (`vaultDocsCmd`, list/get/create/update/delete/links-sub)
- `cmd/vault_documents_helpers.go` → `extractDocsList`, `toMapSlice`, `readFileOrStdin`

That lands both near ~170 and ~50 LoC respectively. Existing test file can stay (tests aren't in the rule's strict target, but the source file clearly is).

---

## High Priority

### H1 — `graphJSONToDOT` drops isolated nodes & node labels
**File:** `cmd/vault.go:144-179`

The loop only iterates `g.Edges`. `g.Nodes` is parsed but never written. Consequences:
1. Isolated documents (no links) disappear from the graph — visualization incomplete.
2. Server-provided node labels (`nodes[].label`) are lost; Graphviz uses node IDs as labels.
3. For a 500-doc vault with few links, DOT output may show just a handful of nodes.

**Fix:** emit nodes first:

```go
for _, n := range g.Nodes {
    id := str(n, "id")
    if id == "" { continue }
    label := str(n, "label")
    if label != "" {
        sb.WriteString(fmt.Sprintf("  %q [label=%q];\n", id, label))
    } else {
        sb.WriteString(fmt.Sprintf("  %q;\n", id))
    }
}
```

Unit test missing for this case — add `TestGraphJSONToDOT_IsolatedNode`.

---

### H2 — Upload: stdout leaks full server error body
**File:** `cmd/vault_upload.go:37-40`

```go
if resp.StatusCode >= 400 {
    body, _ := io.ReadAll(resp.Body)
    return fmt.Errorf("upload failed [%d]: %s", resp.StatusCode, string(body))
}
```

Reading entire error body into memory and inlining it into the error is fine for most cases, but:
- If server returns large HTML error page (e.g., proxy 502), the error message explodes the terminal.
- Error body may contain reflected input (uploaded filename, tags). Not a direct vuln, but not great.

**Fix:** cap body at ~2KB for the error string:

```go
body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
return fmt.Errorf("upload failed [%d]: %s", resp.StatusCode, strings.TrimSpace(string(body)))
```

Low severity; keep as H2 not C.

---

### H3 — `vault documents create` success message can show empty fields
**File:** `cmd/vault_documents.go:127-129`

```go
m := unmarshalMap(data)
printer.Success(fmt.Sprintf("Document created: %s (ID: %s)", str(m, "title"), str(m, "id")))
```

If server returns an envelope structure like `{ok:true, payload:{...}}` and `unmarshalMap` gets the envelope (not payload), `title` and `id` are empty strings → user sees `Document created:  (ID: )`.

Look at test: `TestVaultDocsCreate_CallsEndpoint` uses `vaultEnvelope(...)` which wraps in `{ok, payload}`. The test doesn't assert the success message content, so this bug passes CI. Check `do()` unwrapping behaviour — if it returns `payload`, fine; otherwise we have a silent empty print.

**Action:** quick check of `internal/client/http.go:do()` — if envelope-unwrapped, OK; otherwise fix.

---

## Medium Priority

### M1 — `--file -` with `readFileOrStdin` in create is useless
After C1 fix, re-audit whether `readFileOrStdin("-")` stdin branch is actually reachable. Coverage report explicitly calls out stdin path is untested (69.6% on `readFileOrStdin`). Either cover it or remove.

### M2 — Delete confirmation lacks cascade warning
Phase 3 spec §Risks mentions "typed delete confirmation missing for docs with many links". Current `vault documents delete` just prompts `Delete document <id>?` — no link count preview. Server-side cascade is opaque to user. Consider fetching link count and showing it in prompt. YAGNI-flag; note for future.

### M3 — `vaultDocsUpdateCmd` doesn't allow clearing fields
Update uses `cmd.Flags().Changed(...)` to detect explicit sets. Good — avoids accidental overwrite. But: no way to clear e.g. `title` to empty string except `--title=""`, which `Changed` reports as changed. Edge case; document in help or accept.

### M4 — Test: `--yes` flag leaking between tests
The `rootCmd.PersistentFlags().Set("yes", "false")` reset in `runVaultArgs` is a workaround for test isolation. Cleaner: factor vault setup into a test helper that constructs a fresh `*cobra.Command` per test, OR add a `rootCmd.ResetFlags()` hook. Not blocking.

### M5 — `vault.go:135 fmt.Println(dot)` bypasses `printer`
All other commands use `printer.Print(...)` so output honors `--output` format. DOT text directly to stdout is intentional for `--format=dot`, but the line is a direct `fmt.Println`. Fine because `--format=dot` explicitly requests raw DOT, but consider using `cmd.OutOrStdout()` for test capture.

---

## Low Priority

### L1 — Inconsistent YAGNI: enrichment stop requires `--yes`, but enrichment status doesn't (obviously correct, just noting the pattern is right).

### L2 — `vault upload` should warn on very large files (>100MB) per phase-03 §Risks. Currently streams blindly. Cheap to add: `os.Stat` → `fmt.Fprintf(os.Stderr, "warn: uploading %s, %.1f MB\n", ...)`. Could skip per YAGNI.

### L3 — `TestVaultSearch_DefaultLimit` asserts `max_results=20` but not `offset` omission. Minor coverage gap.

### L4 — `vault_documents.go:266` `strings.HasPrefix(path, "@")` — `@` prefix convention is used here but not documented in `--file` help text. Help says `use - for stdin`; could add `or @path for explicit file reference`. Minor.

---

## Positive Observations

- **Streaming multipart is textbook correct.** `ct := mw.contentType()` captured before goroutine is the exact right fix for the boundary-race bug. Defer `f.Close()` inside goroutine ties file lifetime to write completion.
- **DOT output well-formed** for the edge-only case: proper `digraph vault {...}`, quoted identifiers, edge labels bracketed. Graphviz will parse it.
- **Tree rendering** produces standard `├─`/`└─`/`│ ` prefixes matching `tree(1)`. 100% coverage on rendering functions.
- **`tui.Confirm` contract respected everywhere** — rescan / documents delete / links delete / enrichment stop all pass `cfg.Yes`. Tests verify non-interactive refuses without `--yes` AND endpoint is called with `--yes`. Strong.
- **Test envelope helper** (`vaultEnvelope`) reduces boilerplate cleanly.
- **`MarkFlagRequired`** used for `links create --from --to` and `documents create --title --path` — fail-fast before HTTP.
- **No stack-trace leakage** — all returned errors are wrapped with context (`fmt.Errorf("open %s: %w", ...)`).
- **Output formatter respected** in list/search/links (table vs json branching). Consistent.

---

## Edge Cases Found by Scout

1. **`--q` with `&` or `#`** — query injection (see C2).
2. **`vault tree --path` with spaces** — breaks without escape (see C2).
3. **Isolated vault doc** with no links → disappears from DOT (see H1).
4. **`--file=/dev/null`** or empty file → `readFileOrStdin` returns `""`, passed up then discarded anyway (C1 masks this).
5. **Upload file >2GB** — Go `io.Copy` handles fine; server may not. No CLI-side check.
6. **`tags=",,,"`** → all empty tags skipped correctly by the trim+check loop. Good.
7. **Server returns non-JSON 200 response from upload** — `json.NewDecoder(resp.Body).Decode` fails, falls through to `printer.Success("File uploaded successfully")`. Acceptable graceful degradation.
8. **Concurrent goroutine in `uploadVaultFile`** — if the consumer `c.PostRaw` fails before reading pipe, producer goroutine blocks on pipe write forever. BUT `http.NewRequest` + `Do` reads the body fully on failure, and `io.PipeWriter.CloseWithError(mw.close())` is always called → producer completes. No goroutine leak. Verified via code trace.
9. **`vault documents update` with zero changed flags** → returns `"no fields to update"` error before HTTP. Good.

---

## Security Review

- **Auth:** all commands use `newHTTP()` which sets `Authorization: Bearer`. `PostRaw` explicitly re-sets this header. No bypass.
- **Input validation:** file existence checked via `os.Stat` before opening (prevents accidental empty-path open). `--file`/`--content` mutex-checked.
- **Data leaks:** error messages include file paths (expected) but not credentials. Server error body passed through untruncated (H2).
- **PII in search query:** search query sent to server in POST body, not URL — not logged by HTTP transport. Plan spec mentions "Search query strings: not logged in CLI debug output". No debug print of query found. Good.
- **Destructive ops:** delete/rescan/enrichment-stop all gated by `tui.Confirm(cfg.Yes)`. Non-interactive without `--yes` → print hint to stderr and return false. Exit code 0 (silent refusal) — consider exit 1 for stricter automation semantics, but that's a cross-cutting change for another phase.

---

## Recommended Actions (priority order)

1. **Fix C1** (data loss in `documents create --file`) — BLOCKS SHIP.
2. **Fix C2** (URL query escaping in tree & documents list) — injection-adjacent, fix before ship.
3. **Split C3** (`vault_documents.go` 303 → 200-) — project rule violation.
4. **Fix H1** (emit nodes in DOT) — visualization correctness.
5. **Address H2** (cap upload error body length).
6. **Verify H3** (envelope unwrap for success message) — 5-minute check.
7. Defer M1–M5, L1–L4 to backlog.

---

## Metrics

- Files: 7 source + 5 test
- Total LoC (source): 857
- Vet: clean
- Tests: all green (cmd 1.10s, output 0.31s)
- Coverage: ≥69.6% for all new vault funcs per impl report
- Critical issues: 3 (C1 data loss, C2 URL injection, C3 LoC rule)
- High issues: 3
- Medium: 5
- Low: 4

---

## Unresolved Questions

1. **C1 resolution direction** — does the server `POST /v1/vault/documents` accept inline `content` field? If yes, fix is trivial; if no, we should remove `--file`/`--content` flags entirely. Need to inspect `goclaw/internal/http/vault_handler_documents.go`.
2. **H3 envelope unwrap** — confirm `client.HTTPClient.do()` returns the unwrapped `payload` or the whole `{ok, payload}` envelope. Test mocks return the full envelope, so either the helper unwraps or the `str(m, "title")` calls are always seeing "" in real usage. Quick grep of `http.go:do` will resolve.
3. **Cascade warning** (M2) — should delete show outlink/backlink count before confirming? Server-side preview endpoint exists?

---

**Status:** DONE_WITH_CONCERNS
**Score:** 7.2/10
**Critical count:** 3

Ship-blocker: **C1** (data loss). Fix before merge. C2 and C3 should land in the same follow-up PR — all three are ~30 minutes of work total.
