---
phase: 2
title: "Team Attachment Download"
status: complete
priority: P1
effort: "2h"
dependencies: [1]
---

# Phase 2: Team Attachment Download

## Overview

Add a direct CLI command for authenticated team task attachment downloads. The command writes binary content to an explicit output file only.

## Requirements

- Functional: `goclaw teams attachments download <team-id> <attachment-id> --output <file>` downloads the response body.
- Functional: `--output/-o` is mandatory.
- Functional: existing output file is refused unless `--force` is set.
- Functional: parent directories for output path are created.
- Functional: HTTP errors use existing raw response error handling.
- Non-functional: no stdout binary output; safer for automation and logs.
- Non-functional: use Bearer auth via `newHTTP().GetRaw()`.

## Architecture

Add a new teams subgroup:

```text
teams
  attachments
    download <team-id> <attachment-id> --output <file> [--force]
```

Data flow:

```text
Cobra args -> validate output path -> newHTTP()
  -> GET /v1/teams/{teamId}/attachments/{attachmentId}/download
  -> status check -> open output file -> io.Copy(response.Body, file)
  -> printer.Success("Downloaded N bytes to path")
```

Implementation details:
- Use `url.PathEscape` for both IDs.
- Use `os.OpenFile(path, O_WRONLY|O_CREATE|O_EXCL, 0644)` by default.
- With `--force`, use `O_WRONLY|O_CREATE|O_TRUNC`.
- Use `os.MkdirAll(filepath.Dir(outFile), 0755)`.
- Use `rawResponseError(resp)` for `resp.StatusCode >= 400`.
- After successful status check, `defer resp.Body.Close()` before opening/copying the output file.
- Do not parse `Content-Disposition` in this phase.

## Related Code Files

- Create: `cmd/teams_attachments.go`
- Modify: `cmd/teams.go`
- Read: `cmd/storage.go`
- Read: `cmd/media.go`
- Read: `cmd/backup.go`
- Read: `internal/client/http.go`
- Test: add coverage in a focused `cmd/p5_fillers_test.go` or existing teams test file.

## Implementation Steps

1. Create `teamsAttachmentsCmd` and `teamsAttachmentsDownloadCmd`.
2. Add flags:
   - `StringP("output", "o", "", "Output file path")`
   - `Bool("force", false, "Overwrite output file if it exists")`
3. Validate `--output` before creating HTTP client.
4. Build escaped route path.
5. Stream response to file after status check; close the response body on every success path.
6. Register `teamsAttachmentsCmd` under `teamsCmd`.
7. Add tests:
   - success writes file and hits exact route.
   - Authorization header is present through `newHTTP`.
   - missing `--output` returns validation error before network.
   - existing file without `--force` is refused.
   - existing file with `--force` overwrites.
   - response body closes on success and error paths where practical.
   - local file `--output` does not override the root output-format contract.

## Success Criteria

- [x] `teams attachments download` appears in command tree.
- [x] Binary content is written exactly to requested output path.
- [x] No accidental overwrite without `--force`.
- [x] Parent directory creation works.
- [x] Tests pass with httptest.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Path traversal in output path | This writes only local user-selected output. Do not sanitize beyond parent dir creation; user controls destination. |
| Route ID special chars | `url.PathEscape` both path segments. |
| Large file memory use | Stream with `io.Copy`; no buffering entire body. |
| Wrong auth mode | Use Bearer `GetRaw`; server accepts Bearer and signed token. |
