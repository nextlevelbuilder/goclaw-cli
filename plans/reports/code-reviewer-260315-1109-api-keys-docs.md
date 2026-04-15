## Code Review: api_keys.go & api_docs.go

### Scope
- Files: `cmd/api_keys.go` (131 LOC), `cmd/api_docs.go` (67 LOC), `README.md` (132 LOC)
- Reference: `cmd/providers.go`, `cmd/helpers.go`
- Focus: Pattern conformance, security, edge cases

### Overall Assessment
Both files are clean, well-structured, and follow established project patterns closely. No compile errors, `go vet` passes. Minor issues identified below.

---

### Critical Issues
None.

### High Priority

**1. `buildBody` silently drops empty scopes slice (api_keys.go:70)**

When `scopesRaw` is empty string and `--scopes` is required, `strings.Split("", ",")` returns `[""]`, which the trim/filter loop correctly handles to produce a nil slice. However, `buildBody` with a nil slice hits the `default` case which checks `v != nil` -- a nil `[]string` in Go is typed-nil, so `interface{}([]string(nil)) != nil` is **true** on typed interfaces. The nil slice will be sent to the server as `"scopes": null` in JSON.

- **Impact**: Server may reject or misinterpret null scopes. Since `--scopes` is marked required, this is unlikely in practice, but `--scopes ""` could trigger it.
- **Fix**: Add a guard after the split loop:
```go
if len(scopes) == 0 {
    return fmt.Errorf("at least one scope is required")
}
```

**2. No `url.PathEscape` on revoke ID (api_keys.go:112)**

`args[0]` is concatenated directly into the URL path: `/v1/api-keys/" + args[0]`. While most IDs are UUIDs, this is a path traversal/injection vector if the server uses non-UUID keys. Other commands in the codebase (e.g., `storage.go`, `memory.go`) use `url.PathEscape`.

- **Impact**: Low probability but inconsistent with safer patterns in `storage.go`.
- **Fix**: `c.Delete("/v1/api-keys/" + url.PathEscape(args[0]))` -- though note this is a codebase-wide inconsistency (most commands don't escape), not unique to this file.

### Medium Priority

**3. `api-docs open` duplicates server-empty check (api_docs.go:20-22)**

`newHTTP()` already validates `cfg.Server == ""` and returns `ErrServerRequired`. The `open` command manually checks `cfg.Server` instead of calling `newHTTP()` (since it doesn't need an HTTP client). This is fine functionally, but the error message diverges from the standard sentinel error.

- **Fix**: Use the sentinel: `return client.ErrServerRequired` instead of `fmt.Errorf("server URL required (use --server or config)")`. This keeps error messages consistent.
- **Alternative**: Acceptable as-is since `open` doesn't need auth, and the custom message is more user-friendly.

**4. `openBrowser` process leak (api_docs.go:51-62)**

`exec.Command(...).Start()` launches a child process but never calls `Wait()`. On Linux/macOS this creates zombie processes. The browser will run fine, but the Go process holds a reference to the child.

- **Impact**: Minimal for a CLI that exits immediately after, but not clean.
- **Fix**: Launch in a goroutine that waits:
```go
cmd := exec.Command(...)
if err := cmd.Start(); err != nil {
    return err
}
go cmd.Wait()
return nil
```

**5. `api-docs spec` assumes JSON object response (api_docs.go:45)**

Uses `unmarshalMap` which expects a JSON object. If the server returns the spec as a raw JSON blob or the endpoint changes shape, this silently returns nil map. Consider using `printer.PrintRaw(data)` or handling both object/array shapes.

### Low Priority

**6. Scopes display in list uses `fmt.Sprintf("%v")` (api_keys.go:35)**

This formats each scope element via `%v`, which works for strings but would produce unexpected output for nested types (e.g., `map[...]`). Unlikely given the domain but could use `%s` with a type assertion for clarity.

**7. No `--output` format flag on `api-docs spec`**

The spec command always outputs through `printer.Print(unmarshalMap(data))` which respects global `-o` format. This is correct, but for a spec fetch, users might want raw JSON passthrough without pretty-printing. Minor enhancement opportunity.

---

### README Review
- `api-keys` and `api-docs` entries correctly added to the Commands table (lines 79-80)
- Descriptions are accurate and concise
- Table ordering is logical (alphabetical grouping at bottom)
- No issues found

---

### Edge Cases Found by Scout
- `--scopes ""` (empty required flag) produces nil slice sent as JSON null
- `args[0]` with path-special characters (e.g., `../`) in revoke command -- no escaping
- `openBrowser` on unsupported platform returns error but `open` command swallows it and prints URL -- this is actually **good** graceful degradation
- `api-docs open` works without auth token (intentional and correct -- Swagger UI is public)

### Positive Observations
- Pattern adherence is excellent -- matches `providers.go` structure precisely
- Destructive `revoke` correctly uses `tui.Confirm` with `cfg.Yes` bypass
- Show-once key warning in `create` is a good UX/security practice
- `api-docs open` graceful fallback to printing URL on browser-open failure
- File sizes well within 200-line limit
- Clean separation: `api_docs.go` doesn't import auth-related packages it doesn't need

### Recommended Actions
1. Add scopes-empty validation after split loop (High)
2. Consider `url.PathEscape` for revoke ID (High, but codebase-wide decision)
3. Standardize server-empty error to use sentinel (Medium, optional)
4. Add `go cmd.Wait()` after `Start()` in `openBrowser` (Medium)

### Metrics
- Build: PASS
- go vet: PASS (0 issues)
- Type Coverage: N/A (Go, statically typed)
- Test Coverage: No tests for these files yet (new files)
- Linting Issues: 0

### Unresolved Questions
- Should `url.PathEscape` be adopted codebase-wide for all `args[0]` path segments, or only for user-supplied freeform strings (like file paths in `storage.go`)?
- Should `api-docs spec` support a `--raw` flag for unprocessed JSON output?
