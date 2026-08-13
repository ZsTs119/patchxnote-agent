# MCP Webhook Tools Implementation Checklist

> **For implementation:** Execute this plan sequentially in the primary agent. Do not use sub-agents or parallel task execution.

**Goal:** Expose the existing local webhook workflow to MCP so AI clients can configure named webhook targets, render PatchXNote records, export model IO, and manually send Feishu/DingTalk/generic webhook messages by alias.

**Architecture:** Reuse the existing CLI webhook modules instead of duplicating behavior. MCP handlers should call the same `internal/webhook`, `internal/renderdoc`, `internal/api`, `internal/config`, and `internal/keychain` boundaries that the CLI already uses. Keep URL and signing secret values write-only from the MCP perspective: tools may accept them as input, but responses and config files must only expose masked metadata.

**Tech Stack:** Go, MCP stdio JSON-RPC, Cobra runtime wiring, `internal/mcp`, `internal/webhook`, `internal/renderdoc`, `internal/api`, `internal/config`, `internal/keychain`, `net/http`, `text/template`.

---

## Current State

- [x] Existing MCP tool count is 7 and all tools are read-only.
- [x] Existing CLI webhook module has 10 subcommands: `set`, `list`, `show`, `enable`, `disable`, `remove`, `test`, `draft`, `send`, `export-model-io`.
- [x] Existing webhook module supports target types `feishu`, `dingtalk`, and `generic`.
- [x] Existing webhook aliases already support Chinese and spaces.
- [x] Existing config stores non-secret webhook metadata only.
- [x] Existing keychain boundary stores webhook URLs and optional signing secrets.
- [x] Existing render module supports built-in templates and local template paths.
- [x] Existing API client can fetch delivery documents and explicit model IO traces.
- [x] Current real-install CLI smoke verified Feishu `test`, `send --file`, and `send --draft` with `status=200`.

## Product Decisions

- [x] MCP webhook support is intentionally enabled; webhook is no longer CLI-only.
- [x] This checklist is the required design record for expanding MCP from read-only tools to local config writes and external webhook sends.
- [x] Keep the existing CLI command surface unchanged.
- [x] Add MCP tools on top of the existing modules, not a second webhook implementation.
- [x] MCP may configure local webhook targets because the Agent is installed locally on the user's machine.
- [x] MCP may send external webhook messages only by named alias.
- [x] Local-only webhook tools, such as list/configure/remove/direct send/draft send, do not require PatchXNote account API access.
- [x] Memory-backed rendering, memory-backed sending, and model IO export require the existing Agent login and refresh flow.
- [x] All webhook metadata and secrets remain scoped to the active local PatchXNote profile used by the MCP server process.
- [x] MCP must never list, return, log, or persist full webhook URLs or signing secrets in plaintext.
- [x] MCP tool responses may include `masked_url`, `secret_status`, alias, type, enabled state, template, timestamps, and send results.
- [x] MCP send is manual tool execution. No background schedule, no automatic server-side push, no hidden retry loop.
- [x] Disabled targets cannot be sent unless a later design explicitly adds an override.
- [x] Duplicate target aliases in one send request fail before any HTTP request.
- [x] Provider errors, HTTP status, provider code, and bounded provider messages are surfaced to the MCP caller.
- [x] MCP must keep current App/PC/Admin flows untouched.
- [x] MCP tool count after this feature should be 14: existing 7 read tools plus 7 webhook tools.
- [x] MCP stdout remains JSON-RPC only. All diagnostics, logs, and local debug output must go to stderr or test-only evidence files.
- [x] MCP response bodies must stay within the existing MCP output cap. Large rendered Markdown/model IO should be saved to explicit files and returned as paths plus summaries.
- [x] The docs and release runbook must explicitly note that MCP is no longer read-only after this feature.

## Proposed MCP Tool Surface

### Tool 1: `patchxnote_list_webhook_targets`

- [x] Purpose: list local webhook target metadata for the active profile.
- [x] Annotation: `readOnlyHint: true`.
- [x] Input schema: no arguments, or optional `include_disabled: boolean`, default true to match CLI `webhook list`.
- [x] Output: array of targets with alias, type, enabled, masked URL, template, secret status, created/updated timestamps.
- [x] Must not output full URL or signing secret.

### Tool 2: `patchxnote_configure_webhook_target`

- [x] Purpose: create or update one webhook target by alias.
- [x] Annotation: not read-only; local config/keychain write.
- [x] Input fields:
  - `alias` string, required.
  - `type` enum: `feishu`, `dingtalk`, `generic`, required when creating.
  - `webhook_url` string, required when creating and optional for metadata-only updates.
  - `signing_secret` string, optional.
  - `clear_signing_secret` boolean, optional.
  - `enabled` boolean, optional.
  - `template` string, optional.
- [x] Reject `signing_secret` with `clear_signing_secret`.
- [x] Reject create requests that omit `type` or `webhook_url`.
- [x] If an existing target changes `type`, require a new `webhook_url` unless the implementation deliberately records that preserving the URL is safe.
- [x] Preserve existing URL/secret when omitted.
- [x] Validate URL through `webhook.ValidateWebhookURL` before storing.
- [x] Do not validate template file paths at configure time; validate template names/paths only when rendering or sending.
- [x] Store URL/signing secret through `internal/keychain.SecretStore`.
- [x] Return only masked URL and secret status.
- [x] If keychain/secret store is unavailable, fail clearly before writing partial metadata.

### Tool 3: `patchxnote_remove_webhook_target`

- [x] Purpose: remove one local webhook target by alias.
- [x] Annotation: not read-only; local config/keychain delete.
- [x] Input fields: `alias` string, required.
- [x] Delete non-secret config metadata.
- [x] Delete stored webhook URL and signing secret.
- [x] Return removed alias, type, masked URL, and cleanup status.

### Tool 4: `patchxnote_list_webhook_templates`

- [x] Purpose: list template names available for webhook rendering.
- [x] Annotation: `readOnlyHint: true`.
- [x] Input schema: no arguments.
- [x] Output includes built-in templates:
  - `default`
  - `meeting-summary`
  - `daily-review`
  - `key-items`
  - `raw-markdown`
- [x] Include a short human-readable description for each template.
- [x] Do not scan arbitrary user directories in this tool.
- [x] Clarify that custom templates are used by passing a local file path to render/send tools, not by registering templates in this listing tool.

### Tool 5: `patchxnote_render_webhook_message`

- [x] Purpose: render a PatchXNote memory delivery document into Markdown for review or later sending.
- [x] Annotation: read-only when only returning a bounded preview; not read-only when `save_draft=true` writes local files.
- [x] Input fields:
  - `memory_id` string, required.
  - `platform` enum `mobile|desktop`, optional.
  - `template` string, optional, default `default`.
  - `title` string, optional override.
  - `save_draft` boolean, optional default false.
  - `out` string, required when `save_draft=true`.
  - `include_model_io` boolean, optional default false.
  - `force` boolean, optional default false.
- [x] Fetch content through `GetMemoryDeliveryDocument`.
- [x] Render through `renderdoc.RenderTemplate`.
- [x] Return title, Markdown, memory reference, trace reference, and optional draft paths.
- [x] If rendered Markdown would exceed MCP output limits, require `save_draft=true` and return file paths instead of inline Markdown.
- [x] If `save_draft=true`, write `source.json`, `message.md`, `metadata.json`, and optional `model-io.json` with the same overwrite rules as CLI.
- [x] Do not print raw model IO unless the user explicitly asks through `include_model_io` and the response remains bounded.
- [x] Validate local template path at render time. Reject directories, missing files, and template parse errors with clear messages.

### Tool 6: `patchxnote_export_model_io`

- [x] Purpose: explicitly export Agent model IO JSON for local debugging or user-controlled rewriting.
- [x] Annotation: not read-only because the tool writes the full export to an explicit local file.
- [x] Input fields:
  - exactly one of `memory_id` or `request_id`.
  - `platform` enum `mobile|desktop`, optional.
  - `out` string, required for full export.
  - `force` boolean, optional default false.
- [x] Fetch through `GetMemoryModelIO` or `GetModelRunIOTrace`.
- [x] Save raw export only to the explicit `out` path.
- [x] Return a small summary: source type, memory ID/request ID, platform, field statuses, output path.
- [x] Do not dump provider payload JSON into MCP response by default.
- [x] If the caller wants to inspect the full export, direct them to the output path rather than returning raw payloads through MCP.

### Tool 7: `patchxnote_send_webhook`

- [x] Purpose: send a webhook message to one or more configured targets.
- [x] Annotation: not read-only; external network side effect.
- [x] Input fields:
  - `target_aliases` array of strings, required, min 1.
  - exactly one content source:
    - direct: `title` plus `markdown`.
    - draft: `draft_dir`.
    - memory: `memory_id`, optional `platform`, optional `template`.
    - test: `test_message: true`.
  - `title` string, optional override for direct/file/memory content.
  - `timeout_seconds` integer, optional.
  - memory-source optional fields: `save_draft`, `out`, `include_model_io`, `force`.
- [x] Direct source title precedence is `title` > first Markdown H1 > `PatchXNote 记录`.
- [x] Memory source template precedence is explicit `template` > target template only when exactly one target is used > `default`; for multiple targets, use one explicit/common template and avoid per-target rendering in V1.
- [x] Direct `markdown` input must be bounded below the MCP request scanner limit and below `renderdoc.MaxRenderedMarkdownBytes`.
- [x] Resolve all targets and secrets before sending.
- [x] Fail if any alias is missing, disabled, duplicated, or missing its stored URL.
- [x] For memory source, fetch delivery document and render the template before sending.
- [x] For draft source, read only `message.md` and optional `metadata.json`; never read `model-io.json` for sending.
- [x] For test source, send a small provider-appropriate test card/message.
- [x] Return per-target results with alias, type, success, HTTP status, provider code/message, masked URL, and safe error.
- [x] Provider/send failures should return an MCP tool result with `isError=true` and bounded result details, not only a JSON-RPC error that hides per-target results.

## Implementation Tasks

### Task 1: MCP Wiring Design

**Files:**
- Modify: `internal/mcp/server.go`
- Modify: `internal/cli/mcp.go`
- Modify: `internal/cli/runtime.go`

- [x] Add the API methods needed by webhook MCP tools to `mcp.AgentAPI`.
- [x] Add config and secret-store dependencies to `mcp.Options`.
- [x] Wire `runtime.Config`, `runtime.Secrets`, and existing API client into `mcp.NewServer` from `internal/cli/mcp.go`.
- [x] Preserve current MCP initialization/version behavior.
- [x] Preserve active `--profile`, `--config`, and `--server-base-url` behavior for the MCP server process.
- [x] Run focused MCP tests and confirm existing 7 tools still list when webhook tools are not added yet.

### Task 2: MCP Webhook Tool File

**Files:**
- Create: `internal/mcp/webhook_tools.go`
- Create or modify: `internal/mcp/webhook_tools_test.go`

- [x] Add `defaultWebhookTools(server)` or equivalent helper.
- [x] Register webhook tools from `defaultTools`.
- [x] Add read/write annotation helpers such as `localWriteAnnotations()`, `localDeleteAnnotations()`, and `externalSendAnnotations()`.
- [x] Use MCP annotations deliberately:
  - list/render-without-save/list-templates: `readOnlyHint=true`.
  - configure: `readOnlyHint=false`, `destructiveHint=false`, `idempotentHint` only when inputs fully replace metadata.
  - remove: `destructiveHint=true`.
  - send: `openWorldHint=true`, `destructiveHint=false`, not idempotent.
- [x] Keep tool descriptions clear that URLs/secrets are write-only and sends are external side effects.
- [x] Add tests that `tools/list` returns 14 tools after registration.

### Task 3: Input Schema And Validators

**Files:**
- Modify: `internal/mcp/webhook_tools.go`
- Modify: `internal/mcp/tools.go` if shared schema helpers are needed
- Test: `internal/mcp/webhook_tools_test.go`

- [x] Implement strict object schemas for all 7 new tools.
- [x] Reject unknown fields.
- [x] Validate aliases with existing `webhook.ValidateAlias`.
- [x] Validate target type with existing `webhook.ValidateType`.
- [x] Validate URL with existing `webhook.ValidateWebhookURL`.
- [x] Validate exactly-one source rules for `patchxnote_send_webhook`.
- [x] Validate exactly-one ID rule for `patchxnote_export_model_io`.
- [x] Validate `save_draft=true` requires `out`.
- [x] Validate `force` is only meaningful when writing files.
- [x] Validate `include_model_io=true` requires `save_draft=true` or explicit export path behavior.
- [x] Validate `timeout_seconds` has a sensible range and cannot disable timeouts.
- [x] Validate direct Markdown is non-empty and below the local cap.

### Task 4: Target Registry And Secret Handling

**Files:**
- Modify: `internal/mcp/webhook_tools.go`
- Test: `internal/mcp/webhook_tools_test.go`

- [x] Reuse `webhook.NewRegistry`.
- [x] Reuse `webhook.Secrets`.
- [x] Implement list/configure/remove handlers.
- [x] Ensure configure responses mask URLs.
- [x] Ensure list/show-style responses include `secret_status` only.
- [x] Ensure remove deletes both config metadata and keychain secrets.
- [x] Ensure configure is atomic: do not save metadata if URL/secret storage fails.
- [x] Ensure remove is best-effort for keychain cleanup but returns cleanup status if secret deletion fails.
- [x] Add memory keychain tests for create, update, preserve-secret, clear-secret, remove.

### Task 5: Template Listing And Rendering

**Files:**
- Modify: `internal/renderdoc/templates.go` if needed
- Modify: `internal/mcp/webhook_tools.go`
- Test: `internal/renderdoc/templates_test.go`
- Test: `internal/mcp/webhook_tools_test.go`

- [x] Expose a small function that returns built-in template metadata.
- [x] Implement `patchxnote_list_webhook_templates`.
- [x] Implement `patchxnote_render_webhook_message`.
- [x] Use delivery document projection for normal rendering.
- [x] Preserve platform ambiguity behavior from CLI: surface `400 invalid_request` as guidance to provide `platform`.
- [x] Keep rendered Markdown bounded by `renderdoc.MaxRenderedMarkdownBytes`.
- [x] Keep Feishu heading normalization covered through sender/payload golden tests so Markdown `#` does not appear bare in cards.

### Task 6: File Output Helpers

**Files:**
- Refactor or reuse: `internal/cli/webhook.go`
- Create if needed: `internal/webhook/files.go` or `internal/renderdoc/draft_files.go`
- Test: `internal/cli/webhook_test.go`
- Test: `internal/mcp/webhook_tools_test.go`

- [x] Extract draft writing helpers from CLI if MCP needs the same behavior.
- [x] Keep atomic write behavior.
- [x] Preserve `--force` style overwrite rules.
- [x] Refuse symlink/directory collisions.
- [x] Ensure `send` never reads `model-io.json`.
- [x] Ensure `export_model_io` writes full JSON only to explicit `out`.
- [x] Normalize and validate output paths before recursive writes so a computed path cannot escape the intended output directory.
- [x] Return absolute output paths in MCP responses.

### Task 7: Send Handler

**Files:**
- Modify: `internal/mcp/webhook_tools.go`
- Test: `internal/mcp/webhook_tools_test.go`

- [x] Implement direct `title + markdown` sending.
- [x] Implement test-message sending.
- [x] Implement draft-dir sending.
- [x] Implement memory-backed render-and-send.
- [x] Reuse `webhook.Sender`.
- [x] Reuse provider payload rendering and signing.
- [x] Return aggregate send results even when one target fails.
- [x] Fail before sending when targets are invalid, duplicated, disabled, or missing URL secret.
- [x] Ensure memory-backed send uses the same access-token refresh path as CLI.
- [x] Ensure local direct/draft/test sends do not fail merely because the PatchXNote API session is logged out.

### Task 8: Error Mapping

**Files:**
- Modify: `internal/mcp/errors.go`
- Modify: `internal/mcp/webhook_tools.go`
- Test: `internal/mcp/webhook_tools_test.go`

- [x] Map invalid input to MCP `invalid_params`.
- [x] Map missing target to stable tool error.
- [x] Map missing URL secret to stable tool error with reconfiguration guidance.
- [x] Map provider failures to tool error while including bounded per-target result details.
- [x] Use `CallToolResult.IsError=true` for provider/send failures where a structured result is useful.
- [x] Map auth failures consistently with existing MCP auth behavior.
- [x] Do not leak token, URL, signing secret, provider auth header, or full provider payload.
- [x] Add tests that errors generated after `configure` with a raw URL never echo that raw URL.

### Task 9: End-To-End Smoke

**Files:**
- Modify: `test/e2e/mvp_test.go`
- Test command: `scripts/e2e/mvp-smoke.sh`

- [x] Extend e2e server with generic webhook receiver if not already enough.
- [x] Exercise `tools/list` and assert tool count is 14.
- [x] Exercise `patchxnote_configure_webhook_target`.
- [x] Exercise configure with URL input and assert the raw URL is absent from MCP response.
- [x] Exercise `patchxnote_list_webhook_targets`.
- [x] Exercise `patchxnote_list_webhook_templates`.
- [x] Exercise `patchxnote_send_webhook` with direct Markdown.
- [x] Exercise local direct send while logged out, if the e2e harness supports isolated auth state.
- [x] Exercise duplicate target failure before request.
- [x] Exercise remove target and verify list is empty.
- [x] Keep existing 7 MCP read tools passing.

### Task 10: Real Local Acceptance

**Commands:**
- Build local Windows binary from current commit.
- Install with `packages/npm/bin/patchxnote-agent.js install --from-local ... --install-dir ... --print-config`.
- Run MCP stdio calls against the installed binary.

- [x] Verify installed binary version and commit.
- [x] Verify `tools/list` reports 14 tools.
- [x] Configure a Feishu target by MCP call using the test webhook URL.
- [x] Confirm response contains masked URL only.
- [x] Send test message by MCP call and confirm Feishu receives it.
- [x] Send longer Markdown by MCP call and confirm Feishu card formatting has no bare Markdown `#` issue.
- [x] Send direct Markdown through MCP while no memory record is required.
- [x] If a real memory record exists, run `render_webhook_message`, memory-backed `send_webhook`, and `export_model_io` success paths. Real installed isolated profile returned stable `auth_required`; success path is covered by e2e fake Agent server.
- [x] Remove target by MCP call.
- [x] Scan repo and temp evidence for real webhook URL fragments before commit.

### Task 11: Docs And Release Notes

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify: `docs/release-and-maintenance-runbook.zh-CN.md`

- [x] Update MCP tool count from 7 to 14.
- [x] Document that webhook tools have local config writes and external send side effects.
- [x] Document that URLs and signing secrets are accepted only as write-only inputs and are never listed back.
- [x] Add minimal MCP examples for configure/list/send/remove.
- [x] Update release checklist to include real webhook MCP smoke.
- [x] Update any read-only MCP wording in `docs/engineering-rules.md` or related release docs so the documented product boundary matches the new decision.

## Verification Checklist

- [x] `go test ./internal/mcp ./internal/webhook ./internal/renderdoc ./internal/cli`
- [x] `go test ./...`
- [x] `scripts/e2e/mvp-smoke.sh`
- [x] `node.exe packages/npm/test/install.test.js`
- [x] `git diff --check`
- [x] Secret scan for real webhook URL fragments, phone numbers, access tokens, refresh tokens, OTP, and provider payloads.
- [x] Real installed CLI MCP smoke with Feishu test robot.
- [x] MCP protocol smoke confirms stdout contains only JSON-RPC messages.
- [x] Confirm App/PC/Admin GoServer projects are untouched.

## Edge Cases To Cover

> Covered by the existing CLI/webhook/renderdoc test suite plus the new MCP unit tests, MVP smoke, and real installed Feishu smoke unless a line explicitly notes e2e-only coverage.

- [x] Alias contains Chinese and spaces.
- [x] Alias trims leading/trailing spaces but preserves internal spaces.
- [x] Alias with slash, backslash, tab, newline, or control characters is rejected.
- [x] Empty target list is rejected.
- [x] Duplicate targets are rejected before network sends.
- [x] Disabled target fails clearly.
- [x] Configure disabled target on create, then enable later.
- [x] Configure existing target to disabled without changing URL.
- [x] Configure existing target's template only.
- [x] Configure existing target's type change without URL is rejected or explicitly handled.
- [x] Target metadata exists but URL secret is missing.
- [x] Secret store/keychain unavailable.
- [x] Config file path missing or config directory cannot be created.
- [x] Updating target without URL preserves existing URL secret.
- [x] Updating target without signing secret preserves existing signing secret.
- [x] `clear_signing_secret=true` removes signing secret only.
- [x] Invalid URL scheme, missing host, malformed URL, URL fragment, and control characters are rejected.
- [x] Feishu provider returns HTTP 200 with provider error code.
- [x] DingTalk provider returns HTTP 200 with non-zero `errcode`.
- [x] Provider returns 3xx redirect, 4xx, 5xx, invalid JSON, empty response, timeout.
- [x] Generic webhook returns 204 with empty body.
- [x] Direct Markdown input is empty.
- [x] Direct Markdown input exceeds local cap.
- [x] Rendered Markdown fits `MaxRenderedMarkdownBytes` but exceeds MCP output cap, requiring draft output.
- [x] Local custom template path is missing, points to a directory, fails parsing, or renders too large.
- [x] Rendered Markdown exceeds local cap.
- [x] Draft output path already exists without force.
- [x] Draft output path is a file when directory is expected.
- [x] Symlink collision in output path.
- [x] Draft directory has `metadata.json` with invalid JSON; send still uses `message.md`.
- [x] Draft directory has `model-io.json`; send ignores it.
- [x] `memory_id` exists on both platforms and caller omitted `platform`.
- [x] Wrong platform returns not found.
- [x] No real memory records exist; missing memory errors remain clear.
- [x] `export_model_io` by `request_id` works when memory reference is absent.
- [x] Access token expired but refresh token valid; memory-backed MCP tools refresh and retry once.
- [x] Refresh token expired; memory-backed MCP tools ask user to run login again.

## Out Of Scope For This Plan

- [x] No automatic scheduled webhook sends.
- [x] No server-side webhook storage or delivery.
- [x] No App/PC/Admin flow changes.
- [x] No arbitrary custom HTTP headers.
- [x] No provider retry queue.
- [x] No batch rendering of many memories in one MCP call.
- [x] No MCP exposure of full webhook URL or signing secret after configuration.
