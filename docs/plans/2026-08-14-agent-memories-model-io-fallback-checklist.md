# Agent Memories Model IO Fallback Implementation Plan

> **For implementation:** Execute this plan sequentially in the primary agent. Do not use sub-agents or parallel task execution.

> **2026-08-14 implementation status:** GoServer Agent-only memory fallback and Agent optional-field support have been implemented locally. GoServer unit/integration/OpenAPI/module-smoke checks pass; Agent unit/e2e checks pass; a local GoServer process plus current Windows Agent build verified `patchxnote_list_memories`, `patchxnote_get_memory`, four model IO field tools, search cache, and webhook draft rendering against the real logged-in account.

> **2026-08-14 release status:** Agent `0.2.6` has been released and published to npm. GitHub Release run `31804294579`, npm publish run `31804443521`, Linux release-asset smoke, Windows npm install smoke, default Windows install, MCP initialize, and macOS install smoke run `31804787679` all passed. Public test-server `patchxnote_list_memories` still returned an empty list while GoServer deployment was in progress; re-test that one server-dependent acceptance item after the test server is updated.

**Goal:** Make `patchxnote_list_memories` show the user's readable records, summaries, daily reviews, and key-item style outputs even when the formal structured-result vault has not been populated yet.

**Architecture:** Keep the behavior inside GoServer Agent-only read APIs. `/v1/agent/memories` continues to return formal `structured_result_current/revision` memories when they exist, and supplements them with read-only synthetic memories projected from completed `model_io_trace` rows. The Agent CLI/MCP only adds optional response fields so AI/users can see title, summary, task type, source, and request id without changing App/PC/Admin flows.

**Tech Stack:** GoServer Go HTTP/service/repository modules, PostgreSQL SQL queries, Agent Go CLI/MCP structs, Cobra CLI, existing Agent API client, existing MCP JSON output.

---

## Current Facts

- [x] `patchxnote model-io list --platform mobile` can read `model_io_trace` rows and return `request_id`, `task_type`, state, timestamps, and field availability.
- [x] `patchxnote_list_memories` currently calls GoServer `GET /v1/agent/memories`.
- [x] GoServer `GET /v1/agent/memories` currently reads only `content_storage_consent + structured_result_current + structured_result_revision`.
- [x] The real test account has many `model_request/model_io_trace` rows, but no `structured_result_current/revision` rows, so `patchxnote_list_memories` returns an empty list.
- [x] `model_io_trace.packaged_result_json` currently contains usable model outputs but no `result_id` or `revision_id` mapping.
- [x] `GET /v1/agent/model-runs/{request_id}/io-trace` can already resolve a model run by `request_id`.
- [x] `GET /v1/agent/memories/{memory_id}/delivery-document` builds a readable document from the Agent model IO export, so it can work for synthetic memories once `memory_id=request_id` resolution is supported.

## Product Decisions

- [ ] `patchxnote_list_memories` should answer: "What readable records, summaries, daily reviews, and key-item style outputs does this user currently have?"
- [ ] GoServer changes must be limited to Agent-only modules and Agent-only OpenAPI/docs/tests.
- [ ] Do not change App, PC, Admin, `/v1/content/*`, `/admin/v1/*`, model execution, billing, quota, or structured-result write flows.
- [ ] Do not add a migration for V1.
- [ ] Do not write synthetic rows into `structured_result_current` or `structured_result_revision`.
- [ ] Treat synthetic memories as read-only projections over `model_io_trace`.
- [ ] Use `request_id` as the synthetic memory id.
- [ ] Support the follow-up chain using the same id: list memory -> get memory -> get source text -> get provider response -> get parsed result -> get packaged result -> get delivery document -> webhook.
- [ ] Keep normal large/raw fields out of the list response. Full source text and provider payloads remain available only through explicit model IO/detail endpoints.
- [ ] For synthetic `model_io_trace` memories, rely on the Agent access token account/platform scope and the trace's own `account_id/platform`; do not require `content_storage_consent` to be enabled. This matches the existing Agent model-io read behavior and avoids returning an empty memory list when model IO is available but the formal result vault is not populated.
- [ ] Keep the existing consent/namespace fence for formal `structured_result_current/revision` rows because those rows are still scoped by content storage generation.
- [ ] Exclude failed/unavailable traces from the default synthetic memory list. V1 should list traces where `state='completed'` and `packaged_result_json IS NOT NULL`.
- [ ] Exclude `summary_template_draft` from the default memory list unless product later decides template drafts are user records.
- [ ] Include `transcript_correction`, `event_planning`, `event_summary`, `meeting_summary`, `daily_summary`, and `daily_digest` when they have packaged output.
- [ ] If a future trace maps to a real `structured_result_current`, show the formal memory only once and do not duplicate it as a synthetic memory.
- [ ] If a formal memory id ever equals a synthetic request id, the formal memory wins and the synthetic row is omitted.

## Response Shape

- [ ] Keep existing `Memory` fields backward compatible:
  - `id`
  - `platform`
  - `object_type`
  - `client_object_id`
  - `revision_id`
  - `revision`
  - `schema_id`
  - `schema_version`
  - `source_availability`
  - `payload_plaintext_bytes`
  - `created_at`
  - `updated_at`
- [ ] Add optional fields to GoServer Agent `Memory`:
  - `source`: `structured_result` or `model_io_trace`
  - `request_id`: model run id when source is `model_io_trace`
  - `task_type`: model task type
  - `title`: short human-readable title
  - `summary`: short human-readable preview
- [ ] Add the same optional fields to GoServer `AgentMemoryRef`, because model-io exports and delivery documents embed this ref.
- [ ] Update GoServer Agent OpenAPI schemas with explicit optional properties. Do not leave `additionalProperties: false` schemas missing the new fields.
- [ ] Define `source` as a small enum-like string in docs/tests: `structured_result` or `model_io_trace`.
- [ ] Add the same optional fields to Agent `internal/api.AgentMemory`.
- [ ] Add the same optional fields to Agent `AgentDeliveryMemory` / memory-ref structs where present.
- [ ] Use `omitempty` JSON tags for optional Agent client fields so old server responses do not become noisy empty strings in MCP output.
- [ ] Add the same optional fields to Agent local memory cache so `patchxnote_search_memories` can search title and summary.

## Synthetic Memory Mapping

- [ ] For formal structured-result memories:
  - `source = "structured_result"`
  - `request_id` omitted unless GoServer can safely infer it
  - existing fields keep their current meaning
- [ ] For synthetic `model_io_trace` memories:
  - `id = request_id`
  - `revision_id = request_id`
  - `revision = 1`
  - `source = "model_io_trace"`
  - `request_id = request_id`
  - `platform = trace.platform`
  - `task_type = trace.task_type`
  - `schema_id = "patchnote.agent.model-io-trace"`
  - `schema_version = 1`
  - `source_availability = "text_only"` when safe source text is available, otherwise `"unavailable"`
  - `payload_plaintext_bytes = LEAST(524288, GREATEST(1, octet_length(packaged_result_json::text)))`
  - `created_at = trace.created_at`
  - `updated_at = trace.updated_at`
- [ ] Synthetic `object_type` mapping:
  - Keep values compatible with current `StructuredResultObjectType`: `event`, `key_item`, `daily_digest`, `owner_snapshot`.
  - `event_summary -> event`
  - `daily_digest -> daily_digest`
  - `daily_summary -> daily_digest`
  - `meeting_summary -> event`
  - `event_planning -> event`
  - `transcript_correction -> event`
  - Use `task_type`, `title`, and `summary` to explain the finer-grained meaning instead of adding new `object_type` enum values.
- [ ] Synthetic `client_object_id` precedence:
  - `event_id`
  - `recording_id`
  - `business_id`
  - `request_id`
- [ ] Title extraction:
  - `event_summary`: `packaged_result_json.title`, fallback `PatchXNote 总结`
  - `daily_digest`: `packaged_result_json.title`, fallback `PatchXNote 日回顾`
  - `daily_summary`: `packaged_result_json.title`, fallback `PatchXNote 日回顾`
  - `meeting_summary`: `packaged_result_json.title`, fallback `PatchXNote 总结`
  - `event_planning`: `packaged_result_json.title`, fallback `PatchXNote 记录规划`
  - `transcript_correction`: `packaged_result_json.title`, fallback `PatchXNote 转写整理`
- [ ] Summary extraction:
  - Prefer `summary`, `abstract`, or a bounded excerpt from `summary_markdown`.
  - For `event_planning`, summarize event count and first few title hints without dumping the whole payload.
  - For `transcript_correction`, use a bounded safe excerpt only if already present in packaged/parsed result; otherwise leave empty.
  - Bound summary length with UTF-8 safe truncation. Current local implementation uses 120 runes for title and 500 runes for summary.
- [ ] Source-text availability:
  - Reuse or mirror existing `extractAgentSourceText` behavior from the Agent model-io projection.
  - Ensure `patchxnote_list_model_io_traces` and `patchxnote_list_memories` do not disagree about whether source text is available for the same request.

## GoServer Checklist

### Task G1: Add Contract Tests For Synthetic Memories

**Files:**
- Modify: `../patchxNoteGoServer/internal/agentaccess/repository_integration_test.go`
- Modify: `../patchxNoteGoServer/internal/agentaccess/service_test.go`
- Modify: `../patchxNoteGoServer/internal/agentaccess/http_test.go` if the HTTP response shape has explicit tests

- [ ] Add a failing repository test where consent is enabled, structured-result tables are empty, and one completed `model_io_trace` with packaged JSON exists.
- [ ] Assert `ListMemories(platform=mobile)` returns one synthetic memory with `id=request_id`, `source=model_io_trace`, `request_id`, `task_type`, `title`, and `summary`.
- [ ] Add a test proving failed or unavailable traces do not appear in default memories.
- [ ] Add a test proving `summary_template_draft` does not appear in default memories.
- [ ] Add a test proving formal `structured_result_current` memories still appear.
- [ ] Add a dedupe test where a trace maps to a formal structured memory and the list returns only one item.
- [ ] Add a test proving synthetic memories are still listed when `content_storage_consent` is disabled or absent, as long as the Agent token has the correct account/platform content scope.
- [ ] Add a test proving formal structured-result memories still require the existing consent namespace fence.
- [ ] Add pagination test covering mixed formal and synthetic rows ordered by newest `updated_at`.
- [ ] Add a serialization safety test proving list/get memory responses do not include raw transcript text, provider request, provider response, parsed JSON body, packaged JSON body, token, phone, MAC, or secret markers.
- [ ] Run the targeted tests and confirm they fail for the expected reason.

### Task G2: Extend Agent Memory Domain Shape

**Files:**
- Modify: `../patchxNoteGoServer/internal/agentaccess/domain.go`

- [ ] Add optional `Source string json:"source,omitempty"` to `Memory`.
- [ ] Add optional `RequestID string json:"request_id,omitempty"` to `Memory`.
- [ ] Add optional `TaskType string json:"task_type,omitempty"` to `Memory`.
- [ ] Add optional `Title string json:"title,omitempty"` to `Memory`.
- [ ] Add optional `Summary string json:"summary,omitempty"` to `Memory`.
- [ ] Add the same optional fields to `AgentMemoryRef`.
- [ ] Keep existing fields and JSON names unchanged.
- [ ] Set `Source="structured_result"` when scanning formal memories.
- [ ] Ensure `memoryRefFromMemory` preserves the new optional fields.

### Task G3: Implement Synthetic Memory Projection

**Files:**
- Modify: `../patchxNoteGoServer/internal/agentaccess/repository.go`
- Add or modify helper tests as needed under `../patchxNoteGoServer/internal/agentaccess/`

- [ ] Add helper functions to map task type to object type/title fallback.
- [ ] Add helper functions to extract title and summary from `packaged_result_json` safely.
- [ ] Keep helper output bounded and UTF-8 safe.
- [ ] Update `ListMemories` repository logic to combine formal structured memories with synthetic `model_io_trace` memories.
- [ ] Prefer a simple SQL union when it keeps pagination clear; otherwise run two bounded queries and merge/sort/page in Go.
- [ ] Keep the existing consent fence only for formal structured memories.
- [ ] Do not require consent fence for synthetic model-io memories.
- [ ] Filter synthetic traces to completed rows with packaged result JSON.
- [ ] Exclude `summary_template_draft` by default.
- [ ] Dedupe synthetic traces when they already map to a formal structured memory.
- [ ] Preserve cursor pagination with a stable `(sort_at, id)` ordering.
- [ ] Keep all SQL inside `internal/agentaccess`; do not touch App/PC/Admin repositories.

### Task G4: Resolve Synthetic Memory IDs In Detail APIs

**Files:**
- Modify: `../patchxNoteGoServer/internal/agentaccess/repository.go`
- Modify: `../patchxNoteGoServer/internal/agentaccess/service.go` if service-level validation needs adjustment

- [ ] Update `GetMemory` to return a synthetic memory when `memory_id` is a `request_id`.
- [ ] Update `resolveAgentMemory` to fall back to synthetic memory resolution by `request_id`.
- [ ] Update `GetMemoryModelIO` so `memory_id=request_id` resolves directly to the matching `model_io_trace`.
- [ ] Confirm `GetMemoryDeliveryDocument` works without extra special casing because it already builds from `GetMemoryModelIO`.
- [ ] Keep `GET /v1/agent/memories/{memory_id}` requiring explicit `platform`; callers should use the `platform` returned by the list item.
- [ ] Preserve ambiguous-platform behavior: if platform is omitted and the id could resolve on multiple platforms, return the existing ambiguous-platform error.
- [ ] Keep auth checks scoped to `agent:content.read:<platform>`.
- [ ] Confirm current `validOpaque` validation accepts `mrun_...` request ids used as synthetic memory ids.

### Task G5: Update GoServer Agent Contract Docs

**Files:**
- Modify: `../patchxNoteGoServer/openapi/` Agent OpenAPI artifact if generated or maintained in repo
- Modify: `../patchxNoteGoServer/docs/integrations/apifox/agent-model-io-read-flow.zh-CN.md`
- Modify or add evidence notes under `../patchxNoteGoServer/docs/engineering/evidence/` only after implementation/deploy

- [ ] Document that `/v1/agent/memories` returns formal memories plus synthetic model-run memories.
- [ ] Document optional fields: `source`, `request_id`, `task_type`, `title`, `summary`.
- [ ] Document that synthetic memory `object_type` intentionally stays within the existing enum and the precise model task is in `task_type`.
- [ ] Document that `memory_id=request_id` is valid for Agent memory detail/model-io/delivery-document endpoints.
- [ ] Document that `GET /v1/agent/memories/{memory_id}` still requires `platform`, while model-io and delivery-document endpoints may infer platform when unambiguous.
- [ ] State clearly that this is Agent-only and does not change App/PC/Admin content APIs.
- [ ] Keep examples short and free of raw transcript/provider payloads.

### Task G6: GoServer Verification

**Commands:**
- `cd ../patchxNoteGoServer && go test ./internal/agentaccess -run 'Test.*Memory|Test.*ModelIO' -count=1`
- `cd ../patchxNoteGoServer && go test ./internal/agentaccess ./internal/modeliotrace -count=1`
- `cd ../patchxNoteGoServer && go test ./... -count=1`

- [ ] Targeted Agent Access tests pass.
- [ ] Model IO trace tests pass.
- [ ] Full GoServer tests pass or any unrelated pre-existing failure is documented with exact package and reason.
- [ ] Local real-process smoke starts GoServer against an isolated or test config.
- [ ] Smoke call `GET /v1/agent/memories?platform=mobile&limit=3` returns synthetic memory rows for the real test account.
- [ ] Verify those rows include `source=model_io_trace`, `request_id`, `task_type`, and a readable `title`.
- [ ] Smoke call `GET /v1/agent/memories/{request_id}` returns the synthetic memory.
- [ ] Smoke call `GET /v1/agent/memories/{request_id}/model-io` returns the model IO export.
- [ ] Smoke call `GET /v1/agent/memories/{request_id}/delivery-document` returns readable Markdown.
- [ ] Do not print raw source text, provider payloads, tokens, raw phone, or full user content in smoke evidence.
- [ ] Confirm App/PC/Admin smoke registry is not changed by this task.

## Agent Checklist

### Task A1: Add Optional Memory Fields To API Types

**Files:**
- Modify: `internal/api/types.go`
- Modify: `internal/api/client_test.go`

- [ ] Add optional fields to `AgentMemory`: `Source`, `RequestID`, `TaskType`, `Title`, `Summary`.
- [ ] Add optional fields to `AgentDeliveryMemory`: `Source`, `RequestID`, `TaskType`, `Title`, `Summary`.
- [ ] Use `json:"...,omitempty"` tags for all new optional fields.
- [ ] Keep all existing fields unchanged.
- [ ] Add or update API client test fixtures to include the new fields.
- [ ] Confirm old responses without the new fields still decode correctly.

### Task A2: Preserve Fields In MCP Cache And Search

**Files:**
- Modify: `internal/cache/memory.go`
- Modify: `internal/mcp/tools.go`
- Modify: `internal/mcp/tools_test.go`
- Modify: `internal/cache/memory_test.go` if present or add focused tests nearby

- [ ] Add `Source`, `RequestID`, `TaskType`, `Title`, `Summary` to cached memory records.
- [ ] Update `apiMemoriesToCache` to preserve the new fields.
- [ ] Update search matching so query can match title, summary, task type, request id, object type, and client object id.
- [ ] Keep search bounded and local-only.
- [ ] Add tests showing a listed synthetic memory can later be found by title/summary.
- [ ] Document that `patchxnote_search_memories` searches the local MCP cache after `patchxnote_list_memories` or `patchxnote_get_memory` has populated it; it is not a server-side global search.

### Task A3: Keep MCP Tool Output Backward Compatible

**Files:**
- Modify: `internal/mcp/tools_test.go`
- Modify: `internal/mcp/model_io_tools_test.go` if model-io lookup tests need synthetic ids

- [ ] Confirm `patchxnote_list_memories` returns the new optional fields because JSON output comes from `api.AgentMemory`.
- [ ] Confirm `patchxnote_get_memory` returns the new optional fields.
- [ ] Confirm `patchxnote_get_memory` uses the platform returned by `patchxnote_list_memories`.
- [ ] Confirm `patchxnote_get_model_io_source_text` accepts the synthetic memory id if GoServer supports it.
- [ ] Confirm `patchxnote_get_model_io_provider_response` accepts the synthetic memory id if GoServer supports it.
- [ ] Confirm `patchxnote_get_model_io_parsed_result` accepts the synthetic memory id if GoServer supports it.
- [ ] Confirm `patchxnote_get_model_io_packaged_result` accepts the synthetic memory id if GoServer supports it.
- [ ] Confirm `patchxnote_export_model_io` accepts the synthetic memory id if GoServer supports it.

### Task A4: Update CLI/User Docs

**Files:**
- Modify: `README.md`
- Modify: `docs/plans/2026-08-13-agent-model-io-trace-discovery-checklist.md` only if status tracking needs updating

- [ ] Explain in plain language that "records list" includes formal saved results and recent model-generated readable outputs.
- [ ] Add an example flow:
  - `patchxnote model-io list --platform mobile`
  - MCP `patchxnote_list_memories`
  - use returned `id` and `platform` with detail/model IO field tools
  - send via webhook
- [ ] Make clear that `patchxnote_list_memories` is currently an MCP tool name, while `patchxnote model-io list` is a CLI command.
- [ ] Avoid describing database table names in user-facing README unless needed for troubleshooting.

### Task A5: Agent Verification

**Commands:**
- `cd /home/zsts_119/patchnote-agent && go test ./... -count=1`
- `cd /home/zsts_119/patchnote-agent && node packages/test/install.test.js`
- `patchxnote -o json model-io list --platform mobile --limit 3`
- MCP smoke against installed binary after release/update

- [ ] Agent unit tests pass.
- [ ] Installer test passes if package metadata or installed binary behavior changes.
- [ ] Real installed CLI decodes `patchxnote_list_memories` rows with the new fields after GoServer test deploy.
- [ ] Real MCP tool `patchxnote_list_memories` returns readable synthetic memory rows.
- [ ] Real MCP field tools can use a returned synthetic `id` to fetch source text/provider response/parsed result/packaged result.
- [ ] Real webhook draft/send can use a returned synthetic `id` through delivery document.

## Cross-Repo Acceptance Checklist

- [ ] GoServer test deployment is updated only after local GoServer tests and smoke pass.
- [x] Agent release/update is done after Agent optional-field support is verified locally. Public test-server memory fallback is tracked below because GoServer deployment was still in progress during the Agent release.
- [ ] On a clean user-style install, login succeeds without token copy/paste.
- [ ] `patchxnote_list_memories` for the real test account returns non-empty mobile records.
- [ ] Each returned synthetic memory includes a human-readable `title` or fallback title.
- [ ] Returned synthetic memories use `object_type` values accepted by the published OpenAPI enum.
- [ ] At least one `event_summary` synthetic memory can fetch:
  - source text
  - provider response
  - parsed result
  - packaged result
  - delivery document
- [ ] At least one `daily_digest` synthetic memory can fetch packaged result and delivery document if available.
- [ ] `patchxnote_search_memories` can find a synthetic memory by title or summary after listing has populated the local cache.
- [ ] Desktop platform behavior remains correct: if there is no desktop model IO/formal data, it returns empty or the expected authorization result without affecting mobile.
- [ ] App/PC/Admin routes are not modified and their relevant tests/smoke remain unchanged.

## Risks And Edge Cases

- [ ] A model run can be completed but have no packaged result. It should not appear as a readable memory in V1.
- [ ] A model run can be failed or reconciliation-required. It should stay visible through `model-io list`, not through the default memories list.
- [ ] A user can have both formal structured memories and synthetic model IO memories. Ordering and pagination must stay stable.
- [ ] A future formal memory can map to an older model trace. Dedupe must prevent duplicate list items.
- [ ] Synthetic IDs are request IDs. Detail endpoints must clearly support this so users do not need to understand database internals.
- [ ] `request_id` length/pattern must satisfy both path validation and OpenAPI `OpaqueID`/`ModelLocalID` constraints.
- [ ] Synthetic `payload_plaintext_bytes` must remain within OpenAPI bounds even if packaged JSON is larger than formal structured-result limits.
- [ ] Synthetic rows must not introduce new `object_type` enum values.
- [ ] Summary extraction must not dump full transcript or provider JSON in list results.
- [ ] Formal structured memories may not have title/summary because vault payload is not decrypted by list APIs. That is acceptable; delivery-document/detail endpoints remain the readable-content path.
- [ ] Agent 0.2.5 and older clients will ignore new optional fields, so GoServer must remain backward compatible.
- [ ] If Agent is not updated, users can still list synthetic rows but may not see title/summary in old clients.
- [ ] If GoServer is not updated, Agent-only changes cannot make `patchxnote_list_memories` non-empty.
- [ ] If many model traces exist for one recording, V1 may show multiple processing outputs. This is acceptable for Agent power-user workflows; use `title/task_type/created_at` to make them distinguishable.

## Rollout Notes

- [ ] Implement and deploy GoServer first.
- [ ] Verify deployed GoServer with direct Agent API smoke.
- [ ] Implement Agent optional field support.
- [ ] Release Agent after GoServer test deployment is verified.
- [x] Reinstall/update local CLI to the released Agent version.
- [x] Run real user-style CLI and MCP smoke from the installed binary.
- [x] Commit and push each repository separately with clear messages.
