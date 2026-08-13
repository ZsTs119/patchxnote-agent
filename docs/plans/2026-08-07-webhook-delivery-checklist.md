# Webhook Delivery Implementation Plan

> **For implementation:** Execute this plan sequentially in the primary agent. Do not use sub-agents or parallel task execution.

**Goal:** Add a user-configurable local webhook delivery workflow that can send editable Markdown to named Feishu, DingTalk, or generic webhook targets, including PatchXNote memory-backed drafts through the deployed GoServer Agent delivery projection.

**Architecture:** Keep webhook delivery inside the local `patchxnote` CLI. Store non-secret target metadata in the non-secret config file and store webhook URLs/signing secrets in OS-native keychain through a new secret store boundary. Phase A is local-only: configure targets, send test messages, and send user-edited Markdown files without any GoServer changes. Phase B is now unblocked by GoServer OpenAPI `0.20.16`: the CLI fetches one Agent delivery document per memory, normalizes it into an internal document model, optionally exports model IO JSON, renders templates, saves draft files when requested, and sends platform-specific webhook payloads.

**Tech Stack:** Go, Cobra, Viper, OS-native keychain, `net/http`, `text/template`, Markdown rendering/parsing with `goldmark` if needed, existing `internal/api`, existing `internal/config`, existing `internal/keychain`.

---

## Current Server Dependency Status

- [x] Current GoServer Agent memory APIs are metadata-only: `GET /v1/agent/memories` and `GET /v1/agent/memories/{memory_id}` do not return title, summary, transcript text, provider response, parsed result, or packaged result.
- [x] Current GoServer contract explicitly forbids Agent V1 from using `/v1/content/results:snapshot` or `:changes`; webhook must not bypass through App/PC content APIs.
- [x] GoServer now exposes the Agent-only read contract needed for memory-backed webhook drafts and sends.
- [x] Deployed GoServer contract: OpenAPI `0.20.16`, test release `0376d48ba6ae57dd691ff6a2d19acba05c7edd17`, migration `41 dirty=false`.
- [x] Integration doc: `../patchxNoteGoServer/docs/integrations/apifox/agent-model-io-read-flow.zh-CN.md`.
- [x] Test release evidence: `../patchxNoteGoServer/docs/engineering/evidence/2026-08-13-0376d48-agent-model-io-read-test-release.md`.
- [x] Normal webhook content source is `GET /v1/agent/memories/{memory_id}/delivery-document`.
- [x] Explicit local debug/export source is `GET /v1/agent/memories/{memory_id}/model-io` or `GET /v1/agent/model-runs/{request_id}/io-trace`.
- [x] Phase B is no longer blocked by GoServer; Agent CLI implementation can now include `draft --memory-id`, `send --memory-id`, and explicit model-io export after local API client/tests are added.

## Product Decisions

- [x] Split implementation into Phase A and Phase B.
- [x] Phase A is local-only and must not require any GoServer change.
- [ ] Phase A includes target configuration, keychain secret storage, provider payload rendering/signing, `webhook test`, and `webhook send --file`.
- [ ] Phase B includes `webhook draft --memory-id`, `webhook send --memory-id`, and explicit `webhook export-model-io`; it is now implementation-ready because the GoServer Agent projection is deployed to test.
- [ ] First release is manual only: no automatic server-side or background push.
- [ ] Users configure multiple webhook targets with aliases.
- [ ] Aliases support Chinese, spaces, and common punctuation; callers quote aliases with spaces in shells.
- [ ] Target aliases, metadata, webhook URLs, and signing secrets are scoped by local PatchXNote profile so different accounts/profiles cannot accidentally share webhook targets.
- [ ] Target types: `feishu`, `dingtalk`, `generic`.
- [ ] Webhook URLs may use `http` or `https`; because sends happen from the user's local machine, the CLI does not enforce host allowlists.
- [ ] Webhook URLs must parse as absolute `http` or `https` URLs with a host. Reject empty URLs, control characters, fragments, and malformed values before storing secrets.
- [ ] URL and signing secret values are secret material and must not be written to config files or command output.
- [ ] `webhook set` supports `--url`, `--url-stdin`, `--secret`, and `--secret-stdin`; stdin options are recommended when users do not want webhook secrets in shell history.
- [ ] Reject conflicting input flags such as `--url` with `--url-stdin`, or `--secret` with `--secret-stdin`.
- [ ] Updating an existing target preserves the existing signing secret unless the user passes `--secret`, `--secret-stdin`, or explicit `--clear-secret`.
- [ ] Feishu and DingTalk signing secrets are optional, but when configured the sender must apply the provider-specific signature format before POST.
- [ ] If a Feishu/DingTalk bot uses provider-side keyword or IP allowlist security instead of signing, the CLI does not emulate those policies. Provider rejection is surfaced directly to the user.
- [ ] HTTP redirects are not followed in V1. A 3xx response is surfaced as a provider response instead of silently sending record content to a different URL.
- [ ] HTTP sends use a bounded default timeout and an optional `--timeout` override.
- [ ] Non-secret metadata, such as alias, type, enabled state, template name, timestamps, and masked URL display, may live in the config file.
- [ ] Targets can be enabled or disabled. Explicit sends to disabled targets fail clearly unless a later design adds an override flag.
- [ ] Sending with no `--target` fails clearly. Duplicate `--target` aliases in one command fail before any HTTP request.
- [ ] When Phase B is enabled, each memory-backed send fetches only one `memory_id` from the server. Batch composition is handled through `--file`.
- [ ] For memory-backed commands, `--platform` is optional. If omitted, the CLI lets GoServer infer the platform; if GoServer returns `400 invalid_request` because the same `memory_id` exists on both `mobile` and `desktop`, the CLI asks the user to rerun with `--platform mobile|desktop`.
- [ ] Normal webhook sends must consume a safe delivery document projection, not the full raw `model_io_trace` debug object.
- [ ] Raw model IO trace access is now exposed through explicit Agent read endpoints. It must remain opt-in through an export/debug flag or command and must not be used by default webhook templates.
- [ ] `--template` defaults to `default` when omitted.
- [ ] Draft files are written only when the user explicitly supplies `--out`.
- [ ] `send --file` title precedence is `--title` > first Markdown H1 > file or directory name > `PatchXNote 记录`.
- [ ] The CLI does not pre-limit, truncate, or split long messages. Platform errors are safely surfaced to the user.
- [ ] The CLI may still enforce a high local safety cap for rendered Markdown/file reads to avoid accidental memory exhaustion. This cap is not a provider message-size policy.
- [ ] `generic` sends a stable PatchXNote JSON payload; it does not support custom headers, custom field mapping, or arbitrary template logic in V1.

## Provider Contract Facts To Verify Before Task 8

Provider docs change over time. Confirm these facts against official Feishu/DingTalk docs immediately before implementing payload/signing code and update golden tests if any item has changed.

- [ ] Feishu custom bot signing: when signing is enabled, webhook POST body includes provider-required `timestamp` and `sign` fields.
- [ ] Feishu card send: custom bot card payload uses `msg_type=interactive`; Markdown text inside cards should use the provider-supported card Markdown field shape such as `lark_md`.
- [ ] DingTalk markdown send: custom robot Markdown payload uses `msgtype=markdown`, `markdown.title`, and `markdown.text`.
- [ ] DingTalk signing: when signing is enabled, append provider-required `timestamp` and `sign` query parameters to the webhook URL.
- [ ] DingTalk provider responses include `errcode`/`errmsg`; non-zero `errcode` is a failed send even when HTTP status is 200.
- [ ] DingTalk custom robots have provider-side rate limits. V1 does not queue or retry; 429 or rate-limit style provider errors are surfaced to the user.
- [ ] Provider keyword/IP allowlist failures are not prevalidated locally; show the provider error message with a bounded safe excerpt.

## GoServer Agent Interfaces For Phase B

These interfaces are implemented in GoServer OpenAPI `0.20.16` and deployed to the test server. Agent CLI work must implement against this deployed contract, not the older metadata-only memory APIs.

### Delivery Document Projection

Implemented endpoint:

```http
GET /v1/agent/memories/{memory_id}/delivery-document
GET /v1/agent/memories/{memory_id}/delivery-document?platform=<mobile|desktop>
Authorization: Bearer <agent_access_token>
```

Operation ID: `getAgentMemoryDeliveryDocument`

Required scope: `agent:content.read:<resolved platform>`

Purpose: return a bounded, webhook-safe document projection for normal user sending. This endpoint is the default source for `draft --memory-id` and `send --memory-id`.

Implemented response shape, shortened:

```json
{
  "source": "patchxnote",
  "version": "1",
  "title": "本周复盘",
  "summary": "本次会议确认了...",
  "markdown": "# 本周复盘\n\n## 摘要\n\n...",
  "sections": [
    {
      "title": "关键结论",
      "markdown": "- ..."
    }
  ],
  "key_items": [
    {
      "title": "跟进客户报价",
      "status": "open",
      "owner": "",
      "due_at": "",
      "markdown": "- 跟进客户报价"
    }
  ],
  "memory": {
    "id": "mem_example",
    "platform": "desktop",
    "object_type": "daily_digest",
    "client_object_id": "day_2026_08_13",
    "revision_id": "rev_xxx",
    "revision": 1,
    "schema_id": "patchnote.daily-digest",
    "schema_version": 1,
    "source_availability": "text_only"
  },
  "trace": {
    "trace_id": "mitr_xxx",
    "request_id": "mrun_xxx",
    "platform": "desktop",
    "task_type": "daily_digest",
    "state": "completed",
    "safe_error_code": null,
    "created_at": "2026-08-13T00:00:00Z",
    "updated_at": "2026-08-13T00:00:00Z",
    "completed_at": "2026-08-13T00:01:00Z"
  },
  "generated_at": "2026-08-13T00:00:00Z"
}
```

Projection rules:

- [x] Response includes title, summary, safe Markdown, structured key items, optional memory reference, and trace reference.
- [x] It excludes Provider API keys, Authorization headers, access tokens, refresh tokens, TOS AK/SK, database connection strings, object-store signed URLs, raw audio, full MAC, and SK.
- [x] It prefers `packaged_result_json` / `parsed_result_json` / approved structured result payload over raw provider response text.
- [x] If GoServer cannot map `memory_id` to `model_io_trace`, it returns `404 resource_not_found`; the Agent CLI must not guess from metadata.
- [x] It is additive and does not change App/PC `/v1/content/**`, `/v1/model-runs:execute`, quota settlement, or structured-result encryption behavior.
- [x] `platform` is optional. If omitted and the same `memory_id` exists on both platforms, GoServer returns `400 invalid_request`.

### Explicit Agent Model IO Export Projection

Implemented endpoints:

```http
GET /v1/agent/memories/{memory_id}/model-io
GET /v1/agent/memories/{memory_id}/model-io?platform=<mobile|desktop>
GET /v1/agent/model-runs/{request_id}/io-trace
GET /v1/agent/model-runs/{request_id}/io-trace?platform=<mobile|desktop>
Authorization: Bearer <agent_access_token>
```

Required scope: `agent:content.read:<resolved platform>`

Purpose: allow local debugging/export of model input/output traces, not normal webhook sending.

Rules:

- [x] Endpoints are separate from `delivery-document`.
- [ ] Default webhook templates must not call or require model IO export.
- [ ] Agent CLI may expose model IO only through explicit `export-model-io` or `--include-model-io` behavior.
- [x] Responses may include `source_text`, `client_request_json`, `provider_request_json`, `provider_response_json`, `parsed_result_json`, `packaged_result_json`, and `provider_attempts_json` with `field_status`.
- [x] Query access is owner-scoped and platform-scoped; Agent must not read another account or platform trace.
- [x] If a `request_id` trace is not mapped to a memory, the `memory` field is omitted rather than returned as `null`.
- [x] If only `model_request` exists but no `model_io_trace` exists, GoServer returns `404 resource_not_found`.
- [x] `platform` is optional. Explicit wrong platform returns `404`; ambiguous memory ID without platform returns `400 invalid_request`.
- [ ] CLI must save model IO JSON only to explicit user-chosen paths and must not print raw provider payloads to stdout/stderr.

## Command Surface

```sh
patchxnote webhook set "产品群 飞书" --type feishu --url <webhook_url>
patchxnote webhook set "运营群 钉钉" --type dingtalk --url <webhook_url> --secret-stdin
patchxnote webhook set "内部网关" --type generic --url <webhook_url>
patchxnote webhook set "运营群 钉钉" --clear-secret

patchxnote webhook list
patchxnote webhook show "产品群 飞书"
patchxnote webhook disable "产品群 飞书"
patchxnote webhook enable "产品群 飞书"
patchxnote webhook test "产品群 飞书"
patchxnote webhook remove "产品群 飞书"

patchxnote webhook draft \
  --memory-id <memory_id> \
  --platform desktop \
  --template meeting-summary \
  --out "./patchxnote-drafts/本周复盘" \
  --include-model-io

patchxnote webhook send \
  --target "产品群 飞书" \
  --target "运营群 钉钉" \
  --memory-id <memory_id> \
  --platform desktop \
  --template meeting-summary \
  --save-draft \
  --out "./patchxnote-drafts/本周复盘"

patchxnote webhook export-model-io \
  --memory-id <memory_id> \
  --platform desktop \
  --out "./patchxnote-drafts/本周复盘/model-io.json"

patchxnote webhook export-model-io \
  --request-id <request_id> \
  --out "./patchxnote-drafts/model-run.json"

patchxnote webhook send \
  --target "产品群 飞书" \
  --file "./patchxnote-drafts/本周复盘/message.md"

patchxnote webhook send \
  --target "产品群 飞书" \
  --draft "./patchxnote-drafts/本周复盘"
```

## Output And File Contracts

Draft output directory:

```text
<out>/
  source.json
  message.md
  metadata.json
  model-io.json        # only when explicitly requested with --include-model-io or export-model-io
```

`source.json` contains the safe Agent delivery document projection returned by the Agent API. It must not contain raw audio, complete transcript text outside the approved projection, access tokens, refresh tokens, full MAC, SK, or provider credentials.

For Phase B, `source.json` must be based on the GoServer Agent delivery document projection, not the current metadata-only `AgentMemory` response. Model IO exports must use `model-io.json` or an explicit `--out` path and must not silently replace `source.json` for webhook drafts.

`message.md` is the editable user-facing Markdown rendered from the selected template.

`metadata.json` contains non-secret facts:

```json
{
  "source": "patchxnote",
  "version": "1",
  "platform": "desktop",
  "memory_id": "mem_example",
  "template": "meeting-summary",
  "delivery_request_id": "mrun_xxx",
  "model_io_included": false,
  "generated_at": "2026-08-07T00:00:00Z"
}
```

For `send --file`, `memory` is omitted unless metadata is loaded from a draft directory. File-only generic payloads still include `source`, `version`, `title`, `markdown`, and `metadata.source="file"`.

For `send --draft <dir>`, the CLI loads `<dir>/message.md` and optional `<dir>/metadata.json`. It never reads `<dir>/model-io.json` for normal webhook payloads.

`model-io.json` contains the explicit Agent model IO export response. It may contain provider request/response JSON allowed by the GoServer Agent contract. The CLI must only write it when the user explicitly requests it, never include it in normal webhook payloads, and never print its raw contents to stdout/stderr.

Generic webhook payload:

```json
{
  "source": "patchxnote",
  "version": "1",
  "title": "本周复盘",
  "markdown": "## 摘要\n\n...",
  "memory": {
    "id": "mem_example",
    "platform": "desktop"
  },
  "metadata": {
    "template": "meeting-summary"
  }
}
```

## Proposed Package Layout

- [ ] Create `internal/webhook/target.go` for target metadata domain types and alias validation.
- [ ] Create `internal/webhook/registry.go` for config-backed target registry operations.
- [ ] Create `internal/webhook/secrets.go` for keychain-backed URL and signing secret storage.
- [ ] Create `internal/webhook/sender.go` for target resolution, HTTP send, result aggregation, safe error formatting.
- [ ] Create `internal/webhook/payload.go` for Feishu, DingTalk, and generic payload structs.
- [ ] Create `internal/webhook/signing.go` for Feishu and DingTalk signing helpers.
- [ ] Create `internal/webhook/url.go` for webhook URL validation and masked display helpers if it keeps target validation small.
- [ ] Create `internal/webhook/files.go` or reuse a local helper for atomic draft/export writes.
- [ ] Create `internal/renderdoc/document.go` for the internal normalized document model.
- [ ] Create `internal/renderdoc/templates.go` for built-in template loading and local template path handling.
- [ ] Create `internal/renderdoc/markdown.go` for Markdown title extraction and optional Markdown normalization.
- [ ] Modify `internal/api/types.go` to include the Agent delivery document projection response type.
- [ ] Modify `internal/api/types.go` to include the Agent model IO export response type.
- [ ] Modify `internal/api/client.go` to add Agent delivery document and explicit model IO export client methods.
- [ ] Modify `internal/cli/runtime.go` to expose webhook dependencies through testable factories if needed.
- [ ] Create `internal/cli/webhook.go` for the Cobra command group.
- [ ] Modify `internal/cli/root.go` to register `newWebhookCommand(state)`.
- [ ] Add sanitized fixtures under `testdata/api/`.

## Task 0: Confirm Server Contract

**Files:**
- Read: `../patchxNoteGoServer/docs/integrations/apifox/agent-model-io-read-flow.zh-CN.md`
- Read: `../patchxNoteGoServer/docs/engineering/evidence/2026-08-13-0376d48-agent-model-io-read-test-release.md`
- Read: `../patchxNoteGoServer/docs/integrations/apifox/patchnote-openapi.zh-CN.json`
- Read official provider docs for current Feishu custom bot and DingTalk custom robot webhook payload/signing/response behavior:
  - `https://open.feishu.cn/document/client-docs/bot-v3/add-custom-bot`
  - `https://open.feishu.cn/document/uAjLw4CM/ukzMukzMukzM/feishu-cards/quick-start/send-message-cards-with-custom-bot`
  - `https://open.dingtalk.com/document/orgapp/custom-robots-send-group-messages`
- Modify if needed: `docs/plans/2026-08-07-webhook-delivery-checklist.md`

**Checklist:**
- [x] Confirm current Agent memories are metadata-only and cannot power `draft --memory-id`.
- [x] Identify the exact Agent endpoint and operation ID for the delivery document projection: `GET /v1/agent/memories/{memory_id}/delivery-document`, `getAgentMemoryDeliveryDocument`.
- [x] Confirm required scope is `agent:content.read:<resolved platform>`.
- [x] Confirm response fields available for title, summary, sections, key items, memory, trace, timestamps, and generated time.
- [x] Confirm the projection excludes raw audio, full MAC, SK, provider credentials, object-store signed URLs, access tokens, and refresh tokens.
- [x] Confirm GoServer exposes explicit Agent model IO export endpoints: memory `model-io` and request `io-trace`.
- [x] Confirm `memory_id` / `result_id` / `revision_id` mapping to `model_io_trace` is handled by GoServer; Agent must not guess from metadata.
- [ ] Confirm Feishu request shape, signing fields, and success/failure response shape.
- [ ] Confirm DingTalk markdown request shape, signing query parameters, and `errcode`/`errmsg` response shape.
- [ ] Record any contract gap before implementation.

**Validation:**
- [x] Server contract is no longer missing. Implementation can proceed through Phase A and Phase B after API client tests are added.
- [ ] If Feishu or DingTalk official docs have changed, update payload/signing tests before implementing sender.

## Task 0A: Completed GoServer Agent Delivery Contract Gate

**Files:**
- Read: `../patchxNoteGoServer/docs/integrations/apifox/agent-model-io-read-flow.zh-CN.md`
- Read: `../patchxNoteGoServer/docs/engineering/evidence/2026-08-13-0376d48-agent-model-io-read-test-release.md`

**Checklist:**
- [x] Exact delivery document route and operation ID are decided and deployed.
- [x] Agent model IO export routes are in scope and deployed.
- [x] Current deployed contract uses Agent content read scopes; no extra Agent debug scope was introduced.
- [x] OpenAPI schemas exist for `AgentDeliveryDocument`, `AgentDeliverySection`, `AgentDeliveryKeyItem`, `AgentModelIOExport`, `AgentSourceTextProjection`, and `AgentModelIOFieldStatus`.
- [x] Contract tests keep App/PC model execution request and response schema unchanged.
- [x] Smoke tests prove Agent delivery document/model IO is account-scoped and platform-scoped.
- [x] Smoke tests prove missing mapping and ambiguous platform behavior are stable.
- [x] GoServer support is deployed to test before Agent CLI memory-backed commands are released.

**Validation:**
- [x] GoServer gate PASS. Agent work may proceed with `draft --memory-id`, `send --memory-id`, and explicit `export-model-io`.

## Task 1: Webhook Target Domain And Alias Validation

**Files:**
- Create: `internal/webhook/target.go`
- Test: `internal/webhook/target_test.go`

**Checklist:**
- [ ] Add `TargetType` enum values: `feishu`, `dingtalk`, `generic`.
- [ ] Add `Target` metadata type with alias, type, enabled, masked URL, created/updated timestamps.
- [ ] Add `ValidateAlias(alias string) (string, error)` that trims surrounding whitespace.
- [ ] Permit Chinese, spaces, letters, numbers, and normal punctuation.
- [ ] Reject empty aliases, aliases longer than 64 runes, control characters, tabs, newlines, and path-like separator abuse if it would break storage keys.
- [ ] Reject aliases that become identical only after trimming when setting a new target.
- [ ] Treat aliases as exact-match identifiers after trimming; do not add case folding or Unicode normalization in V1.
- [ ] Add `ValidateType`.
- [ ] Add `MaskWebhookURL`.
- [ ] Add `ValidateWebhookURL` that accepts only absolute `http`/`https` URLs with host and rejects fragments, control characters, and malformed values.

**Tests:**
- [ ] Chinese alias passes.
- [ ] Alias with spaces passes.
- [ ] Empty alias fails.
- [ ] Alias with newline or tab fails.
- [ ] Unknown target type fails.
- [ ] URL masking never reveals token-like query values in full.
- [ ] URL validation accepts local `http://127.0.0.1:...` and normal HTTPS webhook URLs.
- [ ] URL validation rejects missing scheme, unsupported scheme, missing host, fragment, and newline/control characters.

**Commands:**
```sh
go test ./internal/webhook -run 'TestValidateAlias|TestValidateType|TestMaskWebhookURL'
```

## Task 2: Config-Backed Target Registry

**Files:**
- Create: `internal/webhook/registry.go`
- Test: `internal/webhook/registry_test.go`
- Modify: `internal/config/config.go`
- Test: `internal/config/config_test.go`

**Checklist:**
- [ ] Add non-secret `webhooks.targets` config shape.
- [ ] Include the active PatchXNote profile in the target registry scope.
- [ ] Implement registry load from `config.Config`.
- [ ] Implement atomic write to `cfg.Paths.ConfigFile`, creating the config directory if missing.
- [ ] Write config files with owner-readable permissions where the platform allows it.
- [ ] Preserve unrelated config fields.
- [ ] `set` creates or updates by exact normalized alias.
- [ ] `set` must not merge targets across profiles even when aliases are identical.
- [ ] `enable` and `disable` update the enabled state without touching URL or signing secret.
- [ ] `list` returns targets sorted by alias for stable output.
- [ ] `remove` deletes metadata but leaves secret cleanup to the secret store call.

**Tests:**
- [ ] Registry loads empty config.
- [ ] Registry writes first target.
- [ ] Registry updates existing target without duplicate alias.
- [ ] Registry enables and disables an existing target.
- [ ] Registry preserves server/profile/auth config.
- [ ] Registry isolates targets by profile.
- [ ] Registry removes target.

**Commands:**
```sh
go test ./internal/webhook ./internal/config
```

## Task 3: Keychain Secret Store For Webhooks

**Files:**
- Create: `internal/webhook/secrets.go`
- Test: `internal/webhook/secrets_test.go`
- Modify if needed: `internal/keychain/store.go`
- Modify if needed: `internal/keychain/native.go`
- Test if modified: `internal/keychain/*_test.go`

**Checklist:**
- [ ] Store webhook URL by profile and alias.
- [ ] Store optional signing secret by profile and alias.
- [ ] Do not mix webhook secrets with Agent auth credential fields.
- [ ] Fail closed when secure keychain storage is unavailable in a public build; do not silently write webhook secrets into plaintext config.
- [ ] Prefer opaque or escaped keychain account names so aliases with spaces, Chinese, punctuation, or path separators cannot break key construction.
- [ ] Support memory/fake secret store in tests.
- [ ] Delete secrets on `webhook remove`.
- [ ] Support explicit signing secret removal through `--clear-secret` without deleting the webhook URL.
- [ ] When target metadata exists but the keychain URL is missing, report "webhook secret missing; run webhook set again" without printing raw stored values.
- [ ] Return not-found errors that can be mapped to clear CLI messages.
- [ ] Support secret input from flags, stdin, and interactive prompts without echoing secret values.

**Tests:**
- [ ] Put/get/delete URL.
- [ ] Put/get/delete signing secret.
- [ ] Clear signing secret preserves URL.
- [ ] Metadata-present/keychain-missing path maps to a clear secret-missing error.
- [ ] Keychain unavailable path fails closed.
- [ ] Aliases with Chinese and spaces survive key construction.
- [ ] Aliases with punctuation or slashes cannot collide in key construction.
- [ ] Secret values are not included in error strings.

**Commands:**
```sh
go test ./internal/webhook ./internal/keychain
```

## Task 4: Agent Delivery And Model IO API Client

**Files:**
- Modify: `internal/api/types.go`
- Modify: `internal/api/client.go`
- Test: `internal/api/client_test.go`
- Create: `testdata/api/agent_memory_delivery_document_success.json`
- Create: `testdata/api/agent_memory_model_io_success.json`
- Create: `testdata/api/agent_model_run_io_trace_success.json`
- Create: `testdata/api/agent_model_run_io_trace_without_memory_success.json`

**Checklist:**
- [ ] Add typed response for the server-approved Agent delivery document projection.
- [ ] Add typed response for the explicit Agent model IO export projection.
- [ ] Add client method `GetMemoryDeliveryDocument(ctx, token, platform, memoryID)`.
- [ ] Add client method `GetMemoryModelIO(ctx, token, platform, memoryID)`.
- [ ] Add client method `GetModelRunIOTrace(ctx, token, platform, requestID)`.
- [ ] Treat `platform` as optional for these three methods; when non-empty, validate `mobile|desktop`.
- [ ] Validate non-empty `memory_id` / `request_id`.
- [ ] Use Agent bearer token.
- [ ] On memory-backed delivery/model-IO reads, reuse the existing Agent auth refresh path for one 401 retry. If refresh fails, ask the user to log in again.
- [ ] Reuse retry policy only for safe idempotent reads.
- [ ] Sanitize API error mapping.
- [ ] Map `400 invalid_request` from ambiguous platform into a clear CLI-facing error that tells users to pass `--platform`.
- [ ] Map `404 resource_not_found` into clear "not found or no exportable model IO" wording without guessing from metadata.
- [ ] Keep model IO methods out of default webhook send path; they are for explicit local export only.

**Tests:**
- [ ] Delivery document success fixture decodes title, markdown, sections, key items, memory, trace, and generated time.
- [ ] Memory model IO success fixture decodes source text, field status, JSON fields, memory, and trace.
- [ ] Request ID model IO success fixture decodes with memory present.
- [ ] Request ID model IO success fixture decodes when `memory` is omitted.
- [ ] Unauthorized error maps safely.
- [ ] Expired access token refresh success retries the original read once.
- [ ] Expired access token refresh failure maps to login-required guidance.
- [ ] Forbidden scope error maps safely.
- [ ] Not found error maps safely.
- [ ] Ambiguous platform error maps safely.
- [ ] Delivery document fixture scan contains no token, OTP, raw phone, full MAC, SK, raw audio, prompt, or provider payload.
- [ ] Model IO fixtures are synthetic only and contain no token, OTP, raw phone, full MAC, SK, raw audio, provider key, Authorization header, or real user text.

**Commands:**
```sh
go test ./internal/api
```

## Task 5: Internal Document Model

**Files:**
- Create: `internal/renderdoc/document.go`
- Create: `internal/renderdoc/from_api.go`
- Test: `internal/renderdoc/document_test.go`

**Checklist:**
- [ ] Define `Document` with title, summary, sections, key items, code blocks, links, metadata, and source.
- [ ] Map the Agent delivery document API response into `Document`.
- [ ] Keep unavailable fields empty; do not guess title or summary from unrelated metadata.
- [ ] Keep source metadata bounded and non-secret.

**Tests:**
- [ ] Event-like projection maps into document.
- [ ] Key-item-like projection maps into document.
- [ ] Day-like projection maps into document.
- [ ] Missing optional fields do not panic.
- [ ] Disallowed fields are not part of `Document`.

**Commands:**
```sh
go test ./internal/renderdoc
```

## Task 6: Built-In And Local Templates

**Files:**
- Create: `internal/renderdoc/templates.go`
- Create: `internal/renderdoc/templates/default.tmpl`
- Create: `internal/renderdoc/templates/meeting-summary.tmpl`
- Create: `internal/renderdoc/templates/daily-review.tmpl`
- Create: `internal/renderdoc/templates/key-items.tmpl`
- Create: `internal/renderdoc/templates/raw-markdown.tmpl`
- Test: `internal/renderdoc/templates_test.go`

**Checklist:**
- [ ] Use Go `text/template`.
- [ ] Resolve built-in templates by name.
- [ ] Resolve local template paths when `--template` points to an existing file.
- [ ] Add helper funcs only when needed, such as simple date formatting or Markdown escaping.
- [ ] Do not expose environment, filesystem, network, shell, or process helpers to templates.
- [ ] Reject missing templates with clear errors.
- [ ] Built-in templates should stay within a conservative webhook Markdown subset: headings, bullets, links, and fenced code blocks.
- [ ] Bound rendered Markdown size only for local memory safety; do not apply provider message-size policy in the CLI.
- [ ] Keep template execution deterministic for golden tests.

**Tests:**
- [ ] Each built-in template renders a representative document.
- [ ] Local template path renders.
- [ ] Missing template returns a clear error.
- [ ] Rendered Markdown contains expected headings and bullet structure.
- [ ] Oversized rendered output fails with local safety-cap wording, not provider-limit wording.

**Commands:**
```sh
go test ./internal/renderdoc
```

## Task 7: Markdown Utilities

**Files:**
- Create: `internal/renderdoc/markdown.go`
- Test: `internal/renderdoc/markdown_test.go`

**Checklist:**
- [ ] Extract first H1 for `send --file` title inference.
- [ ] Derive fallback title from file or directory name.
- [ ] Optionally use `goldmark` if plain parsing becomes brittle.
- [ ] Keep Markdown body unchanged unless normalization is explicitly needed.

**Tests:**
- [ ] H1 title wins.
- [ ] Missing H1 falls back to filename.
- [ ] Empty file falls back to `PatchXNote 记录`.
- [ ] Unicode H1 works.

**Commands:**
```sh
go test ./internal/renderdoc
```

## Task 8: Platform Payload Renderers

**Files:**
- Create: `internal/webhook/payload.go`
- Create: `internal/webhook/feishu.go`
- Create: `internal/webhook/dingtalk.go`
- Create: `internal/webhook/generic.go`
- Create: `internal/webhook/signing.go`
- Test: `internal/webhook/payload_test.go`

**Checklist:**
- [ ] Feishu renderer produces interactive card payload with `lark_md` content.
- [ ] DingTalk renderer produces `msgtype=markdown` payload with title/text.
- [ ] Generic renderer produces stable JSON with source, version, title, markdown, memory, metadata.
- [ ] Feishu signing helper adds the provider-required timestamp/sign fields when a signing secret is configured.
- [ ] DingTalk signing helper appends provider-required timestamp/sign query parameters when a signing secret is configured.
- [ ] Payload rendering uses `Content-Type: application/json` for all provider types.
- [ ] Provider keyword/IP allowlist failures are handled by sender error mapping, not by renderer validation.
- [ ] Generic file-only payload omits `memory` and sets metadata source to `file`.
- [ ] Do not include webhook URL or signing secret in payload output.
- [ ] Keep output stable for tests.

**Tests:**
- [ ] Feishu payload JSON matches golden structure.
- [ ] DingTalk payload JSON matches golden structure.
- [ ] Generic payload JSON matches golden structure.
- [ ] Feishu signed payload/query matches an official-doc-compatible test vector.
- [ ] DingTalk signed URL query matches an official-doc-compatible test vector.
- [ ] Unsigned Feishu/DingTalk payloads omit signature fields/query parameters.
- [ ] Empty markdown fails before send.

**Commands:**
```sh
go test ./internal/webhook
```

## Task 9: HTTP Sender And Safe Error Mapping

**Files:**
- Create: `internal/webhook/sender.go`
- Test: `internal/webhook/sender_test.go`

**Checklist:**
- [ ] Send targets sequentially in the order provided by `--target`.
- [ ] Fail before sending when target list is empty.
- [ ] Fail before sending when the same target alias appears more than once in one command.
- [ ] Use bounded HTTP timeout.
- [ ] Support a CLI-level timeout override that is parsed once and passed into sender options.
- [ ] Do not follow redirects; return 3xx as a provider response.
- [ ] Treat generic 2xx responses as success.
- [ ] Treat HTTP 204 with empty body as success for `generic`.
- [ ] For Feishu and DingTalk, parse provider response bodies and treat provider-level error codes as failed sends even when HTTP status is 2xx.
- [ ] For failed sends, return target alias, HTTP status code, provider error code/message when available, safe response body excerpt, and masked URL.
- [ ] Handle invalid provider JSON as a failed provider response with a bounded safe excerpt.
- [ ] Classify timeout/context-canceled failures clearly.
- [ ] Surface HTTP 429 and 5xx provider errors without automatic retry.
- [ ] If one target fails, continue sending remaining targets and aggregate results.
- [ ] Overall command returns non-zero if any target fails.
- [ ] Do not retry automatically in V1.

**Tests:**
- [ ] Success response records sent status.
- [ ] Non-2xx response surfaces safe error.
- [ ] Feishu HTTP 200 with provider error code fails.
- [ ] DingTalk HTTP 200 with non-zero `errcode` fails.
- [ ] HTTP redirect fails without following the redirected URL.
- [ ] HTTP 204 generic success is accepted.
- [ ] HTTP 429 is surfaced and not retried.
- [ ] Timeout failure is clear and secret-free.
- [ ] Invalid provider JSON is bounded and secret-free.
- [ ] Duplicate target alias fails before the first request.
- [ ] Mixed target success/failure continues after first failure.
- [ ] Response body excerpt is bounded.
- [ ] Error text does not contain raw webhook URL.

**Commands:**
```sh
go test ./internal/webhook
```

## Task 10: CLI Webhook Set/List/Show/Remove

**Files:**
- Create: `internal/cli/webhook.go`
- Modify: `internal/cli/root.go`
- Test: `internal/cli/webhook_test.go`

**Checklist:**
- [ ] Register `webhook` command group.
- [ ] Implement `webhook set <alias> --type <type> (--url <url>|--url-stdin) [--secret <secret>|--secret-stdin]`.
- [ ] Implement `webhook set <alias> --clear-secret` for removing an existing signing secret while preserving the URL.
- [ ] Reject `--clear-secret` combined with `--secret` or `--secret-stdin`.
- [ ] Reject `--url` combined with `--url-stdin`, and `--secret` combined with `--secret-stdin`.
- [ ] Implement `webhook list`.
- [ ] Implement `webhook show <alias>`.
- [ ] Implement `webhook enable <alias>`.
- [ ] Implement `webhook disable <alias>`.
- [ ] Implement `webhook remove <alias>`.
- [ ] Use Cobra `Args` validators, not manual arg-count checks in `RunE`.
- [ ] Use `StringArray` for repeated `--target` later; do not use comma-splitting flags for aliases because aliases may contain commas.
- [ ] Support `--output plain` and `--output json`.
- [ ] Write human diagnostics to stderr and command results to stdout.
- [ ] Never print raw webhook URL.

**Tests:**
- [ ] Set creates target and stores secret.
- [ ] Set reads URL and signing secret from stdin.
- [ ] Set updates existing target.
- [ ] Set update preserves existing signing secret when no secret flag is provided.
- [ ] Set with `--clear-secret` removes only signing secret.
- [ ] Conflicting input flags fail before reading stdin or writing config/keychain.
- [ ] List shows alias/type/enabled/masked URL.
- [ ] Show shows one target with masked URL.
- [ ] List/show surface secret-missing state when metadata exists but keychain URL is missing.
- [ ] Enable/disable changes enabled state.
- [ ] Remove deletes metadata and secret.
- [ ] JSON output is valid and secret-free.
- [ ] Alias with Chinese and spaces works.

**Commands:**
```sh
go test ./internal/cli -run TestWebhook
```

## Task 11: CLI Webhook Test

**Files:**
- Modify: `internal/cli/webhook.go`
- Test: `internal/cli/webhook_test.go`

**Checklist:**
- [ ] Implement `webhook test <alias>`.
- [ ] Send a small fixed sample document through the selected target renderer.
- [ ] Make it clear in output that this is a test message.
- [ ] Do not require PatchXNote auth for target-only test sends.
- [ ] Fail clearly when the target is disabled.

**Tests:**
- [ ] Test sends sample payload to `httptest.Server`.
- [ ] Missing alias fails clearly.
- [ ] Disabled alias fails before sending.
- [ ] Raw webhook URL does not appear in stdout/stderr.

**Commands:**
```sh
go test ./internal/cli -run TestWebhookTest
```

## Task 12: CLI Webhook Draft

**Files:**
- Modify: `internal/cli/webhook.go`
- Test: `internal/cli/webhook_test.go`
- Test: `test/e2e/mvp_test.go` if the smoke helper needs coverage

**Checklist:**
- [ ] Implement `webhook draft --memory-id <id> [--platform <mobile|desktop>] --template <name-or-path> --out <dir> [--include-model-io]`.
- [ ] Require `--out`.
- [ ] Require authenticated Agent credential.
- [ ] Fetch the Agent delivery document projection; do not use metadata-only `getAgentMemory` as a substitute.
- [ ] If `--platform` is omitted and GoServer returns ambiguous platform, fail with guidance to pass `--platform mobile|desktop`.
- [ ] Render `source.json`, `message.md`, and `metadata.json`.
- [ ] If `--include-model-io` is provided, also fetch explicit model IO and write `model-io.json`.
- [ ] Do not include `model-io.json` contents in stdout/stderr or normal webhook payloads.
- [ ] Write draft files atomically by writing temporary files in the target directory and renaming them into place.
- [ ] Refuse to overwrite existing directory unless `--force` is provided.
- [ ] If `--force` is provided, overwrite only the known draft files (`source.json`, `message.md`, `metadata.json`, optional `model-io.json`) and never recursively delete arbitrary directory contents.
- [ ] Refuse output paths where known draft files would be symlinks or directories instead of regular files.
- [ ] Ensure created files use user-readable permissions but avoid world-writable output.

**Tests:**
- [ ] Draft creates all three files.
- [ ] Draft with `--include-model-io` creates `model-io.json`.
- [ ] Draft refuses missing `--out`.
- [ ] Draft refuses existing output without `--force`.
- [ ] Draft with `--force` overwrites only known draft files.
- [ ] Draft write failure leaves no partial known output file.
- [ ] Draft refuses symlink or directory collision for known draft files.
- [ ] Draft without `--platform` succeeds when server response is unambiguous.
- [ ] Draft without `--platform` maps ambiguous platform error to clear guidance.
- [ ] Draft with local template path works.
- [ ] Draft output scans clean for token-like values.

**Commands:**
```sh
go test ./internal/cli -run TestWebhookDraft
```

## Task 13: CLI Webhook Send

**Files:**
- Modify: `internal/cli/webhook.go`
- Test: `internal/cli/webhook_test.go`

**Checklist:**
- [ ] Implement `webhook send --target <alias> [--target <alias>...] --file <message.md> [--title <title>]`.
- [ ] Implement `webhook send --target <alias> [--target <alias>...] --draft <dir>`.
- [ ] Implement `webhook send --target <alias> [--target <alias>...] --memory-id <id> [--platform <platform>] --template <name-or-path>`.
- [ ] Make `--file`, `--draft`, and `--memory-id` mutually exclusive content sources.
- [ ] Make `--platform` valid only with `--memory-id`.
- [ ] Use repeated `--target` with `StringArray`; do not split aliases on commas.
- [ ] Require at least one `--target`.
- [ ] Reject duplicate target aliases before resolving secrets or sending.
- [ ] For `--file`, title follows `--title` > first H1 > file/directory name > `PatchXNote 记录`.
- [ ] For `--draft`, load `message.md` and optional `metadata.json` from the draft directory; never load `model-io.json` into payload.
- [ ] For `--memory-id`, fetch the Agent delivery document projection and render through template without saving files unless `--save-draft --out <dir>` is provided.
- [ ] If `--platform` is omitted and GoServer returns ambiguous platform, fail with guidance to pass `--platform mobile|desktop`.
- [ ] If `--save-draft` is set, require `--out`, write draft files before sending, and keep those files even when sends fail.
- [ ] If `--save-draft` file writing fails, abort before any webhook HTTP request.
- [ ] If `--save-draft --include-model-io` is set, write `model-io.json` before sending but do not include it in the webhook payload.
- [ ] Fail clearly when any selected target is disabled; do not silently skip disabled explicit targets.
- [ ] Send targets sequentially and aggregate per-target results.
- [ ] Return non-zero if any target fails.

**Tests:**
- [ ] Send file to one target.
- [ ] Send file to multiple targets.
- [ ] Send draft directory to one target.
- [ ] Send memory through template to one target.
- [ ] Send memory with `--save-draft --out` writes draft before send.
- [ ] Send memory with `--save-draft` write failure sends no HTTP request.
- [ ] Send memory without `--platform` succeeds when server response is unambiguous.
- [ ] Send memory ambiguous platform error is clear.
- [ ] Disabled target fails before sending.
- [ ] Missing target fails before sending.
- [ ] Duplicate target fails before sending.
- [ ] Mixed target result returns non-zero with per-target output.
- [ ] `--file`, `--draft`, and `--memory-id` conflicts fail before sending.
- [ ] `--platform` with `--file` or `--draft` fails before sending.
- [ ] Title inference follows precedence.
- [ ] Raw webhook URLs do not appear in output.

**Commands:**
```sh
go test ./internal/cli -run TestWebhookSend
```

## Task 13A: CLI Webhook Export Model IO

**Files:**
- Modify: `internal/cli/webhook.go`
- Test: `internal/cli/webhook_test.go`

**Checklist:**
- [ ] Implement `webhook export-model-io --memory-id <id> [--platform <mobile|desktop>] --out <file>`.
- [ ] Implement `webhook export-model-io --request-id <id> [--platform <mobile|desktop>] --out <file>`.
- [ ] Require authenticated Agent credential.
- [ ] Require exactly one of `--memory-id` or `--request-id`.
- [ ] Require `--out`.
- [ ] Refuse to overwrite an existing file unless `--force` is provided.
- [ ] If `--platform` is omitted and GoServer returns ambiguous platform, fail with guidance to pass `--platform mobile|desktop`.
- [ ] Write the server JSON response as pretty JSON to the exact output file.
- [ ] Write export files atomically and refuse symlink/directory collisions for the output file.
- [ ] Write command success metadata to stdout, not the raw model IO JSON.
- [ ] Do not send webhook messages from this command.

**Tests:**
- [ ] Export by memory ID writes model IO JSON.
- [ ] Export by request ID writes model IO JSON.
- [ ] Export by request ID works when `memory` is omitted in the response.
- [ ] Missing `--out` fails.
- [ ] Both `--memory-id` and `--request-id` fail.
- [ ] Existing output without `--force` fails.
- [ ] Export write failure leaves no partial output file.
- [ ] Export refuses symlink or directory output collision.
- [ ] Ambiguous platform error is clear.
- [ ] stdout/stderr do not contain raw provider JSON, webhook URL, access token, refresh token, OTP, raw phone, full MAC, or SK.

**Commands:**
```sh
go test ./internal/cli -run TestWebhookExportModelIO
```

## Task 14: MCP Surface Decision

**Files:**
- Modify only if accepted: `internal/mcp/tools.go`
- Test only if accepted: `internal/mcp/tools_test.go`

**Checklist:**
- [ ] Decide whether V1 exposes webhook sending through MCP.
- [ ] Recommended V1 default: no MCP webhook tool until CLI flow is proven.
- [ ] If added later, require explicit target aliases and never expose raw webhook URLs through MCP.
- [ ] If added later, mark the tool as write/destructive-adjacent, not read-only.

**Validation:**
- [ ] If MCP surface is not added, confirm existing tool list remains exactly seven.
- [ ] If MCP surface is added later, run MCP protocol smoke.

## Task 15: End-To-End Smoke

**Files:**
- Modify: `scripts/e2e/mvp-smoke.sh`
- Test: `test/e2e/mvp_test.go`

**Checklist:**
- [ ] Extend smoke to configure a local `httptest` or in-process webhook endpoint.
- [ ] Run `webhook set`.
- [ ] Run `webhook set --url-stdin`.
- [ ] Run `webhook list`.
- [ ] Run `webhook test`.
- [ ] Run `webhook send --file`.
- [ ] Run `webhook send --draft`.
- [ ] Cover duplicate `--target` alias failing before any HTTP request.
- [ ] Cover missing keychain URL for existing target metadata.
- [ ] Cover URL validation rejection for fragments and control characters.
- [ ] Cover `--clear-secret`.
- [ ] Phase A smoke may stop here and must not require GoServer delivery-document support.
- [ ] Add mocked Agent API coverage for `webhook draft --memory-id`.
- [ ] Add mocked Agent API coverage for `webhook send --memory-id`.
- [ ] Add mocked Agent API coverage for `webhook export-model-io`.
- [ ] Add a real test-server acceptance lane for memory-backed draft/send only when a disposable test account and synthetic memory fixture are available.
- [ ] Include provider-level failure cases where HTTP is 200 but provider body reports failure.
- [ ] Scan default draft/send evidence for tokens, raw phone, full MAC, SK, raw audio, prompt, provider payload, model IO JSON, and raw webhook URL.
- [ ] Scan explicit `export-model-io` evidence for tokens, raw phone, full MAC, SK, raw audio, provider key, Authorization header, and real user text.

**Commands:**
```sh
go test ./...
scripts/e2e/mvp-smoke.sh
```

## Task 16: Documentation

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify: `docs/release-and-maintenance-runbook.zh-CN.md`
- Modify if public guide changes: `../patchxNoteGoServer/docs/integrations/patchnote-agent-feishu-guide.zh-CN.md`

**Checklist:**
- [ ] Document webhook target setup.
- [ ] Document `--url-stdin` and `--secret-stdin` as safer alternatives to shell-history-visible flags.
- [ ] Document Phase A file-based send workflow and document memory-backed draft/send/export as Phase B backed by GoServer OpenAPI `0.20.16`.
- [ ] Document target types.
- [ ] Document optional Feishu/DingTalk signing secret support.
- [ ] Document `--clear-secret`, `--timeout`, and `send --draft <dir>`.
- [ ] Document secret storage boundary.
- [ ] Document `http`/`https` local-user responsibility.
- [ ] Document that redirects are not followed in V1.
- [ ] Document that automatic sending is not included.
- [ ] Document provider-side keyword/IP allowlist restrictions as user/provider configuration; the CLI surfaces those failures but does not prevalidate them.
- [ ] Document DingTalk/Feishu provider rate limits as provider behavior; V1 does not queue or retry.
- [ ] Document platform ambiguity handling for memory-backed commands: rerun with `--platform mobile|desktop` when GoServer reports ambiguous memory ID.
- [ ] Document that `export-model-io` is explicit local export/debug only and never part of normal webhook payloads.
- [ ] Document that long-message platform errors are surfaced as returned by the webhook provider.
- [ ] Include examples with synthetic aliases and fake URLs only.

**Validation:**
```sh
git diff --check
go test ./...
scripts/e2e/mvp-smoke.sh
node packages/npm/test/install.test.js
```

## Security Checklist

- [ ] Phase A does not call GoServer for webhook message content.
- [ ] Phase B uses only Agent delivery document projection for default webhook sends.
- [ ] Raw `model_io_trace` data is available only through explicit local export behavior and is never used by default webhook templates.
- [ ] Profile-scoped webhook targets prevent one local profile from accidentally using another profile's target alias or secret.
- [ ] No webhook URL is printed in full.
- [ ] No webhook URL is stored in config files.
- [ ] No signing secret is stored in config files.
- [ ] Secure keychain unavailable fails closed in public builds instead of storing webhook secrets in plaintext.
- [ ] `--url-stdin` and `--secret-stdin` are available for users who want to avoid shell history.
- [ ] No access token or refresh token enters webhook payloads.
- [ ] Normal webhook payloads never include `model-io.json` or raw model IO fields.
- [ ] Default webhook fixtures/docs contain no OTP, raw phone, full MAC, SK, raw audio, prompt, provider payload, or model IO JSON.
- [ ] Explicit model IO fixtures are synthetic-only and contain no OTP, raw phone, full MAC, SK, raw audio, provider key, Authorization header, or real user text.
- [ ] Draft output is explicit and user-chosen.
- [ ] Draft/export writes are atomic and do not recursively delete arbitrary user files.
- [ ] Draft/export refuses symlink or directory collisions for known output files.
- [ ] Redirect responses are not followed automatically.
- [ ] Provider-level error bodies are parsed and treated as failures.
- [ ] Platform errors are bounded before display.
- [ ] Generic payload schema is stable and documented.
- [ ] Draft/export commands should make it easy for users to save both the rendered Markdown and the original model IO JSON locally, then edit or AI-polish before sending.
- [ ] MCP stdio stdout remains JSON-RPC only.
- [ ] Existing MCP tool list remains exactly seven unless a separate accepted MCP webhook plan changes it.
- [ ] Duplicate `--target` aliases fail before any webhook request.

## Final Validation Checklist

- [x] Server Agent delivery document projection contract confirmed.
- [ ] Current metadata-only Agent memories are not used as a fake delivery document.
- [x] GoServer Agent delivery document route is deployed before `draft --memory-id` / `send --memory-id` are enabled.
- [ ] `go test ./...` PASS.
- [ ] `scripts/e2e/mvp-smoke.sh` PASS.
- [ ] `node packages/npm/test/install.test.js` PASS if installer docs/package behavior changed.
- [ ] Webhook CLI tests cover aliases with Chinese and spaces.
- [ ] Webhook CLI tests cover `send --draft`, `--clear-secret`, duplicate targets, missing keychain URL, and URL fragment/control-character rejection.
- [ ] Webhook sender tests cover partial failure.
- [ ] Webhook sender tests cover timeout, 204 generic success, 429 no-retry behavior, invalid provider JSON, and bounded provider excerpts.
- [ ] Draft file tests cover `source.json`, `message.md`, `metadata.json`.
- [ ] Draft/export file tests cover atomic write and symlink/directory collision behavior.
- [ ] Sensitive-value scan PASS.
- [ ] README and Chinese README updated.
- [ ] Release runbook updated if command surface changes.
