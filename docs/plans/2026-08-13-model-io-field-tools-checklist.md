# Model IO Field Tools Implementation Plan

> **For Codex:** REQUIRED SKILL: Use `executing-plans` to implement this plan task-by-task. Execute sequentially in the primary agent only. Do not use sub-agents or parallel task execution.

**Goal:** Add AI-friendly CLI and MCP access to four explicit Agent model IO fields: source text, provider response, parsed result, and packaged result.

**Architecture:** Keep GoServer as the source of truth and reuse the existing Agent model IO endpoints already wired into `internal/api`. Add one shared local `internal/modelio` helper layer so CLI and MCP select, format, bound, and write model IO fields the same way. Expose semantic MCP tools for AI clients and semantic CLI subcommands for AI-driven terminal use.

**Tech Stack:** Go, Cobra, MCP stdio JSON-RPC, `internal/api`, `internal/cli`, `internal/mcp`, `internal/modelio`, `internal/keychain`, JSON `RawMessage`, local atomic file writes.

---

## Implementation Status

- [x] Shared `internal/modelio` field-selection, formatting, availability, and file-output layer implemented.
- [x] Neutral `internal/localfile` atomic writer added and reused by webhook/model IO output paths.
- [x] CLI `patchxnote model-io` command group implemented with four field subcommands plus full `export`.
- [x] MCP field tools implemented: `patchxnote_get_model_io_source_text`, `patchxnote_get_model_io_provider_response`, `patchxnote_get_model_io_parsed_result`, and `patchxnote_get_model_io_packaged_result`.
- [x] Existing webhook full model IO export command/tool kept compatible.
- [x] MVP e2e smoke updated to cover all 18 MCP tools, field-only inline output, file output, memory-id lookup, request-id lookup, and existing webhook flow.
- [x] Docs updated for CLI/MCP usage, memory/request lookup guidance, trusted-local model IO boundary, and 18-tool smoke expectations.
- [x] Real installed Windows binary smoke completed against the local installed path. The logged-in real profile was authenticated and listed 18 MCP tools, but currently returned no mobile/desktop memories, so live model IO success could not be exercised against production data. Installed-binary e2e fixture covered the successful user flow end to end.

---

## Current State

- [ ] Existing API client already has `GetMemoryModelIO(ctx, accessToken, platform, memoryID)`.
- [ ] Existing API client already has `GetModelRunIOTrace(ctx, accessToken, platform, requestID)`.
- [ ] Existing `api.AgentModelIOExport` includes:
  - `source_text`
  - `provider_response_json`
  - `parsed_result_json`
  - `packaged_result_json`
  - `field_status`
- [ ] Existing CLI can export full model IO through `patchxnote webhook export-model-io`.
- [ ] Existing MCP can export full model IO through `patchxnote_export_model_io`.
- [ ] Current MCP tool count is 14 after webhook MCP support.
- [ ] Current CLI has no top-level `model-io` command group.

## Product Decisions

- [ ] Add model IO as its own Agent/CLI/MCP capability, not as a webhook-only feature.
- [ ] Keep `patchxnote webhook export-model-io` as a compatibility command for now.
- [ ] Add `patchxnote model-io export` as the preferred complete export command.
- [ ] Add semantic CLI subcommands because CLI is also primarily operated by AI:
  - `patchxnote model-io source-text`
  - `patchxnote model-io provider-response`
  - `patchxnote model-io parsed-result`
  - `patchxnote model-io packaged-result`
  - `patchxnote model-io export`
- [ ] Add four semantic MCP tools:
  - `patchxnote_get_model_io_source_text`
  - `patchxnote_get_model_io_provider_response`
  - `patchxnote_get_model_io_parsed_result`
  - `patchxnote_get_model_io_packaged_result`
- [ ] Do not add a generic MCP `field` selector tool in V1. Tool names should be explicit so AI callers choose correctly.
- [ ] Each field command/tool accepts exactly one lookup key:
  - `memory_id`
  - `request_id`
- [ ] `platform` remains optional at schema level but recommended. If GoServer returns platform ambiguity, surface a clear message asking for `mobile` or `desktop`.
- [ ] Default inline output must stay bounded. Large fields require an explicit output file.
- [ ] `source-text` may output text; JSON fields output pretty JSON when valid and raw JSON bytes only if already valid JSON.
- [ ] Missing/unavailable fields return a clear unavailable status rather than an internal error.
- [ ] Do not copy model IO logic into MCP or CLI separately; all field selection and file writing goes through `internal/modelio`.
- [ ] `internal/modelio` must not import `internal/cli`, `internal/mcp`, or `internal/webhook`. If file writing needs a shared atomic writer, move the generic writer to a neutral helper package instead of creating a domain dependency on webhook.
- [ ] Auth, refresh, API client construction, and scope checks remain in CLI/MCP runtime layers. `internal/modelio` should operate on already-fetched `api.AgentModelIOExport` values plus local output paths.
- [ ] Source text availability is determined from `export.SourceText`: nil, non-`available` availability, or empty text must be surfaced clearly.
- [ ] JSON field availability is determined from both `export.FieldStatus` and the actual `json.RawMessage`. Status says available but payload is empty/null must not be treated as a successful content result without explanation.
- [ ] Model IO field tools may expose user source text and provider/model payloads for the logged-in user. This is allowed only because it is explicit, local, manual, and field-scoped.
- [ ] Do not return or write unrelated model IO fields when the caller requested one field. For example, provider-response must not include client request, provider request, parsed result, packaged result, or provider attempts.
- [ ] Do not add CLI memory list/search in this plan. If CLI memory discovery does not already exist, document MCP discovery as the current way to get `memory_id`, and create a separate plan if CLI discovery is needed later.
- [ ] Do not change App/PC/Admin flows.
- [ ] Do not add new GoServer APIs unless local verification proves the existing Agent endpoints do not return one of the required fields.
- [ ] MCP tool count after this feature should be 18: existing 14 tools plus 4 model IO field tools.
- [ ] Existing `patchxnote_export_model_io` MCP full export tool remains available and unchanged.

## CLI Command Shape

Examples:

```sh
patchxnote model-io source-text --memory-id <memory_id> --platform mobile
patchxnote model-io provider-response --memory-id <memory_id> --platform mobile
patchxnote model-io parsed-result --memory-id <memory_id> --platform mobile
patchxnote model-io packaged-result --memory-id <memory_id> --platform mobile
patchxnote model-io export --memory-id <memory_id> --platform mobile --out ./model-io.json
```

Request ID lookup:

```sh
patchxnote model-io provider-response --request-id <request_id> --platform mobile --out ./provider-response.json
```

Output file examples:

```sh
patchxnote model-io source-text --memory-id <memory_id> --platform mobile --out ./source-text.txt
patchxnote model-io provider-response --memory-id <memory_id> --platform mobile --out ./provider-response.json
patchxnote model-io parsed-result --memory-id <memory_id> --platform mobile --out ./parsed-result.json
patchxnote model-io packaged-result --memory-id <memory_id> --platform mobile --out ./packaged-result.json
```

Common flags:

- [ ] `--memory-id <id>`
- [ ] `--request-id <id>`
- [ ] `--platform mobile|desktop`
- [ ] `--out <path>`
- [ ] `--force`
- [ ] `--output plain|json`

Rules:

- [ ] Exactly one of `--memory-id` or `--request-id` is required.
- [ ] Field subcommands may print bounded content to stdout when `--out` is omitted.
- [ ] Field subcommands must write full field content to `--out` when provided.
- [ ] If content exceeds the CLI inline limit and `--out` is omitted, fail with guidance to pass `--out`.
- [ ] `model-io export` always requires `--out`.
- [ ] For `--output plain` without `--out`, print only the requested field content to stdout.
- [ ] For `--output plain` with `--out`, print a short success summary with the absolute output path.
- [ ] For `--output json`, return a structured object with field name, availability, status, media type, lookup metadata, inline value or output path, and `written`.
- [ ] For JSON fields in `--output json`, do not double-encode JSON as a string when it can be returned as JSON.
- [ ] Human-readable diagnostics and errors go to stderr; stdout stays parseable.

## MCP Tool Shape

Common input schema for all four tools:

```json
{
  "memory_id": "optional string",
  "request_id": "optional string",
  "platform": "mobile | desktop, optional",
  "out": "optional local file path",
  "force": false
}
```

Common behavior:

- [ ] Exactly one of `memory_id` or `request_id` is required.
- [ ] If `out` is omitted and field content is within MCP output limit, return field content inline.
- [ ] If `out` is provided, write the full field content and return only summary plus path.
- [ ] If field content exceeds MCP output limit and `out` is omitted, return `output_too_large` with guidance to pass `out`.
- [ ] For JSON fields, return/write pretty JSON where possible.
- [ ] For source text, return/write plain text plus availability metadata.
- [ ] Return `field_status`, source lookup type, platform, memory/request reference, and `written=true|false`.
- [ ] Never write unless `out` is explicit.
- [ ] All four tools have write-capable annotations because `out` can write local files. Mark them non-destructive and not open-world.
- [ ] Inline MCP responses must include only the requested field plus summary metadata.
- [ ] File-write MCP responses must not include field content by default; return summary plus absolute path.
- [ ] Provider/model payloads should be returned as JSON values where possible, not escaped JSON strings.

## Task 1: Shared Model IO Package

**Files:**
- Create: `internal/modelio/modelio.go`
- Create: `internal/modelio/modelio_test.go`

**Steps:**

- [ ] Define `Field` enum values:
  - `source-text`
  - `provider-response`
  - `parsed-result`
  - `packaged-result`
- [ ] Define `Lookup` with `MemoryID`, `RequestID`, and `Platform`.
- [ ] Define `FieldResult` with:
  - `Field`
  - `Available`
  - `Status`
  - `MediaType`
  - `Text`
  - `JSON`
  - `OutputPath`
  - `LookupKind`
  - `MemoryID`
  - `RequestID`
  - `Platform`
  - `Source`
  - `Memory`
  - `Trace`
- [ ] Implement `ValidateLookup`.
- [ ] Implement `ValidateField`.
- [ ] Implement `SelectField(export, field)`.
- [ ] Implement JSON pretty formatting for `json.RawMessage`.
- [ ] Implement source text formatting as plain text.
- [ ] Implement unavailable field handling using `export.FieldStatus`.
- [ ] Implement source text availability handling using `export.SourceText.Availability`.
- [ ] Implement null/empty JSON handling for JSON fields.
- [ ] Add unit tests for each field.
- [ ] Add unit tests for unavailable fields.
- [ ] Add unit tests for invalid lookup, invalid field, and empty content.

**Verification:**

```sh
go test ./internal/modelio
```

Expected: PASS.

## Task 2: File Output Helpers

**Files:**
- Modify: `internal/modelio/modelio.go`
- Test: `internal/modelio/modelio_test.go`

**Steps:**

- [ ] Add `WriteFieldFile(path, result, force)`.
- [ ] Do not import `internal/webhook` from `internal/modelio`. Move the generic atomic writer to a neutral helper package, such as `internal/localfile`, if reuse is needed by both webhook and modelio.
- [ ] Use `0600` permissions for model IO files.
- [ ] Reject writing to directories.
- [ ] Reject symlink output paths.
- [ ] Respect `force=false` when the file already exists.
- [ ] Return absolute output path.
- [ ] For `source-text`, write plain UTF-8 text.
- [ ] For JSON fields, write pretty JSON with trailing newline.
- [ ] Add tests for write success.
- [ ] Add tests for existing file without force.
- [ ] Add tests for directory path.
- [ ] Add tests for symlink path where supported by the OS.
- [ ] Add tests for output path with Chinese characters and spaces.

**Verification:**

```sh
go test ./internal/modelio ./internal/webhook
```

Expected: PASS.

## Task 3: CLI Model IO Command Group

**Files:**
- Create: `internal/cli/model_io.go`
- Modify: `internal/cli/root.go`
- Test: `internal/cli/model_io_test.go`

**Steps:**

- [ ] Add top-level `model-io` command group.
- [ ] Add subcommands:
  - `source-text`
  - `provider-response`
  - `parsed-result`
  - `packaged-result`
  - `export`
- [ ] Reuse the existing runtime auth/refresh path used by webhook model IO export.
- [ ] For each field subcommand, fetch full `AgentModelIOExport`, select one field, and output only that field.
- [ ] For `export`, fetch full export and write the same JSON structure currently used by `webhook export-model-io`.
- [ ] Keep `webhook export-model-io` working unchanged.
- [ ] Prefer factoring shared fetch logic out of `internal/cli/webhook.go` instead of duplicating it.
- [ ] Add `--memory-id`, `--request-id`, `--platform`, `--out`, `--force`, and `--output`.
- [ ] Validate exactly-one lookup before any API call.
- [ ] If inline field output is too large, fail with guidance to use `--out`.
- [ ] Preserve global `--profile`, `--config`, and `--server-base-url` behavior.
- [ ] Keep command names stable and AI-readable; avoid hidden aliases in V1.
- [ ] Do not print provider payload content in success summaries when `--out` is used.
- [ ] Add CLI tests for `--output json` shape.
- [ ] Add CLI tests that JSON fields are not double-encoded.
- [ ] Add CLI tests for each subcommand.
- [ ] Add CLI tests for `--out` write path.
- [ ] Add CLI tests for validation errors.
- [ ] Add CLI tests that sensitive credential material is not printed.

**Verification:**

```sh
go test ./internal/cli ./internal/modelio
```

Expected: PASS.

## Task 4: MCP Model IO Field Tools

**Files:**
- Create: `internal/mcp/model_io_tools.go`
- Create: `internal/mcp/model_io_tools_test.go`
- Modify: `internal/mcp/tools.go`
- Modify: `internal/mcp/server_test.go`

**Steps:**

- [ ] Register four new MCP tools:
  - `patchxnote_get_model_io_source_text`
  - `patchxnote_get_model_io_provider_response`
  - `patchxnote_get_model_io_parsed_result`
  - `patchxnote_get_model_io_packaged_result`
- [ ] Add strict input schemas with no unknown fields.
- [ ] Add validators for exactly-one lookup, platform enum, output path, and force boolean.
- [ ] Reuse `Server.withAgentAccessToken` or equivalent refresh-aware auth path.
- [ ] Fetch `AgentModelIOExport` through existing Agent API methods.
- [ ] Select fields through `internal/modelio`.
- [ ] Return bounded inline content when safe.
- [ ] Write to `out` when requested.
- [ ] Return `output_too_large` when content is too large and no `out` is provided.
- [ ] Return unavailable fields as structured results with `available=false`, `field_status`, and no content. Do not use JSON-RPC internal errors for normal unavailability.
- [ ] Map API auth, permission, platform ambiguity, not found, and rate limit errors consistently with existing MCP behavior.
- [ ] Add tests that `tools/list` returns 18 tools.
- [ ] Add tests for each new tool success path.
- [ ] Add tests for output-file path.
- [ ] Add tests for auth required.
- [ ] Add tests for invalid lookup.
- [ ] Add tests for oversized inline output.
- [ ] Add tests for structured unavailable results.
- [ ] Add tests that provider payload content is returned only for the requested field and does not include access/refresh tokens.

**Verification:**

```sh
go test ./internal/mcp ./internal/modelio
```

Expected: PASS.

## Task 5: End-To-End MVP Smoke

**Files:**
- Modify: `test/e2e/mvp_test.go`
- Test command: `scripts/e2e/mvp-smoke.sh`

**Steps:**

- [ ] Extend the e2e fake Agent server model IO fixture if needed.
- [ ] Cover both memory-id and request-id Agent model IO endpoints in e2e if the existing fake server can support both without noisy setup.
- [ ] Update MCP expected tool count from 14 to 18.
- [ ] Call all four new MCP model IO field tools.
- [ ] Assert source text returns only source text content.
- [ ] Assert provider-response returns provider response content.
- [ ] Assert parsed-result returns parsed result content.
- [ ] Assert packaged-result returns packaged result content.
- [ ] Exercise one `out` write path.
- [ ] Exercise CLI `model-io source-text`.
- [ ] Exercise CLI `model-io provider-response --out`.
- [ ] Exercise CLI `model-io export --out`.
- [ ] Exercise one request-id lookup through CLI or MCP.
- [ ] Keep existing 14 MCP tools passing.
- [ ] Keep existing webhook e2e flow passing.
- [ ] Assert MCP stdout remains JSON-RPC only after field tool calls.

**Verification:**

```sh
scripts/e2e/mvp-smoke.sh
```

Expected: `MVP smoke PASS`.

## Task 6: Docs And User Guidance

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify: `docs/release-and-maintenance-runbook.zh-CN.md`
- Modify if needed: `docs/engineering-rules.md`

**Steps:**

- [ ] Update MCP tool count from 14 to 18.
- [ ] Add CLI examples for all `model-io` subcommands.
- [ ] Explain how to get `memory_id`:
  - MCP `patchxnote_list_memories`
  - MCP `patchxnote_search_memories`
  - CLI memory discovery only if a real CLI command exists by then; otherwise say MCP discovery is the supported path.
- [ ] Document `memory_id` vs `request_id`.
- [ ] Document that model IO field tools may expose model/provider payloads for the logged-in user and should be used only in trusted local agents.
- [ ] Document that large fields should be written to explicit local files.
- [ ] Document that `webhook export-model-io` remains compatible but `model-io export` is the preferred command.
- [ ] Document output behavior for `--output plain`, `--output json`, and `--out`.
- [ ] Document that client request, provider request, and provider attempts are intentionally not exposed as field tools in this version.
- [ ] Update release runbook MCP smoke checklist.

**Verification:**

```sh
git diff --check
```

Expected: no output.

## Task 7: Real Local Acceptance

**Commands:**

- Build local Windows binary from current workspace.
- Install with `packages/npm/bin/patchxnote-agent.js install --from-local ... --install-dir ... --print-config`.
- Run MCP stdio calls against installed binary.

**Steps:**

- [ ] Verify installed binary version and commit.
- [ ] Verify `tools/list` reports 18 tools.
- [ ] With a logged-in Agent profile, find one real `memory_id` through MCP or CLI.
- [ ] Run `patchxnote model-io source-text --memory-id <id> --platform <platform>`.
- [ ] Run `patchxnote model-io provider-response --memory-id <id> --platform <platform> --out <temp.json>`.
- [ ] Run `patchxnote model-io parsed-result --memory-id <id> --platform <platform> --out <temp.json>`.
- [ ] Run `patchxnote model-io packaged-result --memory-id <id> --platform <platform> --out <temp.json>`.
- [ ] Run one MCP field tool inline.
- [ ] Run one MCP field tool with `out`.
- [ ] If possible, run one `request_id` lookup from a known model IO trace.
- [ ] Confirm output files exist and are parseable when JSON.
- [ ] Confirm no access token, refresh token, OTP, raw phone, full webhook URL, or unrelated provider payload is printed in command logs.
- [ ] If no real memory with model IO exists, record `field_unavailable` or `not_found` as the real profile result and rely on e2e for success path.

## Verification Checklist

- [x] `go test ./internal/modelio`
- [x] `go test ./internal/cli ./internal/mcp ./internal/modelio`
- [x] `go test ./...`
- [x] `scripts/e2e/mvp-smoke.sh`
- [x] `node.exe packages/npm/test/install.test.js`
- [x] `git diff --check`
- [x] Secret scan for real phone, OTP, access token, refresh token, webhook URL fragments, and provider credentials.
- [x] Real installed CLI/MCP smoke for model IO field tools.
- [x] Confirm GoServer/app/pc projects are untouched unless a missing Agent field requires a separate accepted GoServer plan.

## Edge Cases To Cover

- [ ] `memory_id` and `request_id` both missing.
- [ ] `memory_id` and `request_id` both provided.
- [ ] `platform` omitted and server returns platform ambiguity.
- [ ] `platform` invalid.
- [ ] Memory not found.
- [ ] Request ID not found.
- [ ] Model IO export exists but source text unavailable.
- [ ] Model IO export exists but provider response unavailable.
- [ ] Model IO export exists but parsed result unavailable.
- [ ] Model IO export exists but packaged result unavailable.
- [ ] Field status says available but payload is empty.
- [ ] Source text object is nil.
- [ ] Source text availability is not `available` but text is present.
- [ ] Source text availability is `available` but text is empty.
- [ ] JSON field is `null`.
- [ ] JSON field is a scalar, array, and object.
- [ ] JSON field contains invalid JSON bytes from server.
- [ ] JSON field contains very large payload.
- [ ] Source text contains very large text.
- [ ] Inline output exceeds MCP limit.
- [ ] Inline output exceeds CLI limit.
- [ ] `out` path exists without `force`.
- [ ] `out` path parent directory does not exist.
- [ ] `out` path is a directory.
- [ ] `out` path is a symlink.
- [ ] `out` path is relative.
- [ ] `out` path has spaces or Chinese characters.
- [ ] `out` path is a Windows path during real installed smoke.
- [ ] `out` path is a WSL path during Linux/e2e smoke.
- [ ] Refresh token valid and access token expired: refresh once and retry.
- [ ] Refresh token expired: ask user to login again.
- [ ] API returns 403 permission denied.
- [ ] API returns 429 rate limited.
- [ ] API returns 5xx dependency/server error.
- [ ] Existing `patchxnote_export_model_io` MCP full export still works.
- [ ] Existing `patchxnote webhook export-model-io` CLI command still works.
- [ ] MCP stdout remains JSON-RPC only.
- [ ] CLI stdout remains machine-readable when `--output json`.

## Out Of Scope

- [ ] No model execution.
- [ ] No prompt editing.
- [ ] No provider retry or replay.
- [ ] No App/PC/Admin flow changes.
- [ ] No server-side field mutation.
- [ ] No automatic AI rewriting.
- [ ] No batch export of many memories.
- [ ] No custom provider payload filtering beyond selecting the explicit requested field.
- [ ] No field tools for `client_request_json`, `provider_request_json`, or `provider_attempts_json` in this version.
- [ ] No new CLI memory list/search command in this plan.
