# Agent Model IO Trace Discovery Implementation Plan

> **For Codex:** REQUIRED SKILL: Use `executing-plans` to implement this plan task-by-task. Execute sequentially in the primary agent only. Do not use sub-agents or parallel task execution.

**Goal:** Let PatchXNote Agent discover model IO trace `request_id` values and then read source text, provider response, parsed result, and packaged result without depending on structured memory records.

**Architecture:** GoServer remains the source of truth for `model_io_trace`. Add one Agent-only read list endpoint that returns lightweight trace summaries for the logged-in account and selected platform, then wire the local Agent API client, CLI, and MCP to expose that discovery path. Existing field-detail commands/tools continue to fetch actual source/model payloads by `request_id`.

**Tech Stack:** Go, Chi HTTP routing, PostgreSQL, Cobra, MCP stdio JSON-RPC, PatchXNote Agent auth refresh, `internal/agentaccess`, `internal/api`, `internal/cli`, `internal/mcp`, `internal/modelio`.

**Implementation status (2026-08-13):** Implemented and locally accepted. GoServer test environment is deployed at revision `1a88ba67fd3e2dd97a6a58773a7f4f881f705232` with OpenAPI `0.20.17`; Agent local installed binary `0.2.3-local` verified `model-io list`, request-id field exports, MCP `patchxnote_list_model_io_traces`, and MCP field exports against the real logged-in test account.

**Verification snapshot:**

- [x] GoServer `make ci`, `release-build`, `release-verify-artifact`, `deploy-test`, `deploy-test-status`, `deploy-test-verify`.
- [x] Agent `go test ./...`.
- [x] Agent `scripts/e2e/mvp-smoke.sh` with all 19 MCP tools.
- [x] Agent npm wrapper `node packages/npm/test/install.test.js`.
- [x] Real installed CLI/MCP flow wrote source/provider/parsed/packaged/model-io export files without printing payloads.
- [x] Real Feishu webhook CLI and MCP alias configure/send/remove smoke passed.

---

## Current Facts

- [ ] GoServer already writes completed model calls to `model_request`.
- [ ] GoServer already writes raw model debug/projection data to `model_io_trace`.
- [ ] GoServer already exposes Agent detail read:
  - `GET /v1/agent/model-runs/{request_id}/io-trace`
- [ ] Agent CLI already supports field reads by `request_id`:
  - `patchxnote model-io source-text --request-id <request_id>`
  - `patchxnote model-io provider-response --request-id <request_id>`
  - `patchxnote model-io parsed-result --request-id <request_id>`
  - `patchxnote model-io packaged-result --request-id <request_id>`
  - `patchxnote model-io export --request-id <request_id>`
- [ ] Agent MCP already supports field reads by `request_id`:
  - `patchxnote_get_model_io_source_text`
  - `patchxnote_get_model_io_provider_response`
  - `patchxnote_get_model_io_parsed_result`
  - `patchxnote_get_model_io_packaged_result`
- [ ] Agent currently has no way to list model IO traces, so AI cannot discover `request_id` values by itself.
- [ ] `patchxnote_list_memories` only lists `structured_result_current`; direct `model-runs:execute` traces may exist without structured memory.
- [ ] Real local investigation confirmed this can happen: completed `model_io_trace` rows exist while `structured_result_current` is empty.

## Product Decisions

- [ ] Treat `model_io_trace` as the Agent discovery source for model-chain data.
- [ ] Do not force `model_io_trace` rows to become memories in this feature.
- [ ] Keep `memory_id` flow for structured result/memory workflows.
- [ ] Use `request_id` flow for model IO workflows.
- [ ] Agent trace discovery must work even when the trace has no structured memory mapping.
- [ ] Agent trace discovery must not require `structured_result_current` to exist.
- [ ] Agent trace discovery authorization must match the existing `GET /v1/agent/model-runs/{request_id}/io-trace` detail endpoint. If the detail endpoint can read a trace by `request_id`, the list endpoint should be able to discover that trace for the same account/platform.
- [ ] Consent and structured-memory joins are only for optional `memory` metadata unless the existing detail endpoint authorization policy changes.
- [ ] Add discovery, not another full-detail dump:
  - list endpoint returns lightweight summaries;
  - actual source text/model payloads remain behind existing field-detail tools.
- [ ] Require `platform=mobile|desktop` in V1 list calls to keep content platform-scoped and avoid accidental cross-platform merging.
- [ ] Allow optional filters that are cheap and useful for AI:
  - `request_id`
  - `task_type`
  - `state`
  - `recording_id`
  - `event_id`
  - `business_id`
  - `date_from`
  - `date_to`
  - `limit`
  - `cursor`
- [ ] Default list limit is 20, max is 50.
- [ ] Sort newest first by `created_at DESC, trace_id DESC`.
- [ ] Return `request_id` as the primary follow-up key.
- [ ] Return `trace_id` only if it is already part of Agent-safe contract or needed for pagination/debug; do not require it for follow-up reads.
- [ ] Do not return `client_request_json`, `provider_request_json`, `provider_response_json`, `parsed_result_json`, `packaged_result_json`, `provider_attempts_json`, source text, prompt, transcript, or model body in list responses.
- [ ] Include field availability and byte-size metadata so AI can choose the right follow-up field command.
- [ ] Do not modify App/PC/Admin request or response flows.
- [ ] Do not add write APIs.

## Target User Flow

AI/MCP:

```text
1. patchxnote_list_model_io_traces(platform="mobile", limit=10)
2. Pick a request_id from the returned summaries.
3. patchxnote_get_model_io_source_text(request_id=..., platform="mobile")
4. patchxnote_get_model_io_provider_response(request_id=..., platform="mobile")
5. patchxnote_get_model_io_parsed_result(request_id=..., platform="mobile")
6. patchxnote_get_model_io_packaged_result(request_id=..., platform="mobile")
```

CLI:

```sh
patchxnote model-io list --platform mobile
patchxnote model-io source-text --request-id <request_id> --platform mobile
patchxnote model-io provider-response --request-id <request_id> --platform mobile --out ./provider.json
patchxnote model-io parsed-result --request-id <request_id> --platform mobile --out ./parsed.json
patchxnote model-io packaged-result --request-id <request_id> --platform mobile --out ./packaged.json
```

## GoServer API Contract

Add:

```http
GET /v1/agent/model-io-traces?platform=mobile&limit=20
```

Query parameters:

- [ ] `platform` required: `mobile|desktop`
- [ ] `request_id` optional, exact request/run id
- [ ] `task_type` optional, exact model task enum
- [ ] `state` optional, exact trace state enum; default no state filter
- [ ] `recording_id` optional, exact trace recording id
- [ ] `event_id` optional, exact trace event id
- [ ] `business_id` optional, exact trace business id
- [ ] `date_from` optional RFC3339 timestamp, filters `created_at >= date_from`
- [ ] `date_to` optional RFC3339 timestamp, filters `created_at < date_to`
- [ ] `limit` optional int, default 20, min 1, max 50
- [ ] `cursor` optional opaque cursor

Accepted `task_type` values:

- [ ] `transcript_correction`
- [ ] `meeting_summary`
- [ ] `daily_summary`
- [ ] `daily_digest`
- [ ] `event_planning`
- [ ] `event_summary`
- [ ] `summary_template_draft`
- [ ] `legacy_classification`

Accepted `state` values:

- [ ] `created`
- [ ] `reserved`
- [ ] `executing`
- [ ] `validating`
- [ ] `completed`
- [ ] `provider_failed`
- [ ] `reconciliation_required`
- [ ] `response_cache_expired`
- [ ] `trace_failed`

Response:

```json
{
  "items": [
    {
      "request_id": "mrun_example",
      "platform": "mobile",
      "api_contract_version": "1.0.0",
      "task_type": "event_planning",
      "state": "completed",
      "safe_error_code": null,
      "recording_id": "rec_example",
      "event_id": "event_example",
      "business_id": "local-example",
      "created_at": "2026-08-13T12:00:00Z",
      "updated_at": "2026-08-13T12:00:05Z",
      "completed_at": "2026-08-13T12:00:05Z",
      "source_text_availability": "available",
      "field_status": {
        "client_request_json": "available",
        "provider_request_json": "available",
        "provider_response_json": "available",
        "parsed_result_json": "available",
        "packaged_result_json": "available",
        "provider_attempts_json": "available"
      },
      "field_bytes": {
        "client_request_json": 3423,
        "provider_request_json": 2840,
        "provider_response_json": 1423,
        "parsed_result_json": 1011,
        "packaged_result_json": 1011,
        "provider_attempts_json": 720
      },
      "memory": null
    }
  ],
  "next_cursor": "opaque"
}
```

Rules:

- [ ] Response must be account-scoped through Agent bearer token.
- [ ] Response must be platform-scoped through `agent:content.read:<platform>`.
- [ ] List visibility is based on `model_io_trace.account_id` and requested platform, not on structured memory existence.
- [ ] Response must set `Cache-Control: no-store`.
- [ ] Missing `platform` returns `400 invalid_request`.
- [ ] Invalid `request_id`, `task_type`, `state`, `recording_id`, `event_id`, `business_id`, timestamp, `limit`, or `cursor` returns `400 invalid_request`.
- [ ] `date_from >= date_to` returns `400 invalid_request`.
- [ ] Empty results return `{"items":[]}`.
- [ ] Cursor is opaque base64url JSON, same style as existing memory cursor.
- [ ] Cursor payload should use `created_at` and `trace_id` or `request_id`.
- [ ] Cursor payload should bind to the active filter set, for example through a short filter fingerprint, so a cursor from another platform/filter/date range is rejected.
- [ ] Byte sizes should use database-safe size calculations, not serialize payloads into logs.
- [ ] `source_text_availability` should reuse existing Agent source-text extraction logic if feasible.
- [ ] If source extraction requires reading `client_request_json`, it may be scanned inside repository code only to compute availability; it must never be returned, logged, cached, or included in tests/evidence.
- [ ] If source extraction is too heavy for list, return only availability derived from existing field status and document the limitation.
- [ ] Do not expose `source_text.text`, prompt, provider request body, provider response body, parsed body, packaged body, attempts body, raw transcript, full phone, token, MAC, SK, or webhook URLs in list output.

## GoServer Task 1: Domain Types

**Files:**
- Modify: `../patchxNoteGoServer/internal/agentaccess/domain.go`

**Steps:**

- [ ] Add `ModelIOTraceListInput`.
- [ ] Add `ModelIOTracePage`.
- [ ] Add `ModelIOTraceSummary`.
- [ ] Add `AgentModelIOFieldBytes` or equivalent struct.
- [ ] Reuse `AgentModelIOFieldStatus`.
- [ ] Include optional `Memory *AgentMemoryRef`.
- [ ] Include optional trace identifiers:
  - `api_contract_version`
  - `recording_id`
  - `event_id`
  - `business_id`
- [ ] Include `SourceTextAvailability`.
- [ ] Keep JSON tags stable and snake_case.

**Verification:**

```sh
cd ../patchxNoteGoServer
GOTOOLCHAIN=go1.26.5 go test ./internal/agentaccess
```

Expected: compile fails first if repository/service are not implemented yet, then PASS after later tasks.

## GoServer Task 2: Repository List Query

**Files:**
- Modify: `../patchxNoteGoServer/internal/agentaccess/repository.go`
- Modify or extend tests: `../patchxNoteGoServer/internal/agentaccess/repository_integration_test.go`

**Steps:**

- [ ] Add repository interface method:
  - `ListModelIOTraces(ctx, accountID, input, cursor, limitPlusOne)`
- [ ] Validate:
  - account id is opaque;
  - platform is valid;
  - limit is 1..51;
  - cursor fields are valid when present.
- [ ] Query `model_io_trace` by `account_id` and `platform`.
- [ ] Apply optional filters:
  - `request_id`
  - `task_type`
  - `state`
  - `recording_id`
  - `event_id`
  - `business_id`
  - `date_from`
  - `date_to`
- [ ] Left join enabled `content_storage_consent` and structured result tables exactly like `GetModelRunIOTrace`.
- [ ] Return optional memory reference when `trace.result_id` maps to a current structured result.
- [ ] Compute field status with the same helper logic used by detail projection.
- [ ] Compute byte sizes with `pg_column_size(...)` or an equivalent DB-safe expression.
- [ ] Do not select or scan raw JSON payload values into the summary unless needed only for field status/source availability.
- [ ] Add integration fixture with:
  - trace with memory;
  - trace without memory;
  - trace for another account;
  - trace on the other platform;
  - trace with same `created_at` but different tie-breaker;
  - old trace for cursor/date filtering.
- [ ] Assert account/platform isolation.
- [ ] Assert newest-first ordering.
- [ ] Assert cursor returns the next page.
- [ ] Assert cursor from a different filter set is rejected by service validation.
- [ ] Assert traces without consent/memory mapping still appear for the authenticated account/platform.
- [ ] Assert response does not contain raw source/model payload fields.

**Verification:**

```sh
cd ../patchxNoteGoServer
GOTOOLCHAIN=go1.26.5 go test ./internal/agentaccess -run 'Test.*ModelIOTrace'
```

Expected: PASS.

## GoServer Task 3: Service Validation And Cursor

**Files:**
- Modify: `../patchxNoteGoServer/internal/agentaccess/service.go`
- Test: `../patchxNoteGoServer/internal/agentaccess/service_test.go`

**Steps:**

- [ ] Add `ListModelIOTraces(ctx, actor, input)`.
- [ ] Require `contentScope(input.Platform)`.
- [ ] Use `validateActorForScope(ctx, actor, agent:content.read:<platform>)`.
- [ ] Default limit to 20 and cap at 50.
- [ ] Reject cursor longer than 512 bytes.
- [ ] Add `modelIOTraceCursor` and encode/decode helpers.
- [ ] Cursor payload should include:
  - `created_at`
  - stable tie-breaker `trace_id` or `request_id`
  - filter fingerprint for platform and optional filters
- [ ] Return `next_cursor` when repository returns `limit+1` rows.
- [ ] Add service tests:
  - missing/invalid platform;
  - invalid limit;
  - invalid cursor;
  - cursor filter mismatch;
  - invalid date range;
  - invalid task/state enum;
  - missing content scope;
  - successful pagination.

**Verification:**

```sh
cd ../patchxNoteGoServer
GOTOOLCHAIN=go1.26.5 go test ./internal/agentaccess
```

Expected: PASS.

## GoServer Task 4: HTTP Route

**Files:**
- Modify: `../patchxNoteGoServer/internal/agentaccess/http.go`
- Test: existing `agentaccess` HTTP tests if present, otherwise service/repository coverage plus smoke.

**Steps:**

- [ ] Register route:
  - `GET /v1/agent/model-io-traces`
- [ ] Decode query with strict unknown-parameter rejection.
- [ ] Require exactly one value for each query key.
- [ ] Trim-check all string query values.
- [ ] Parse RFC3339 timestamps.
- [ ] Reject `date_from >= date_to`.
- [ ] Parse and validate optional `request_id`, `task_type`, `state`, `recording_id`, `event_id`, and `business_id`.
- [ ] Call `service.ListModelIOTraces`.
- [ ] Set `Cache-Control: no-store`.
- [ ] Return JSON page.
- [ ] Map validation to existing `invalid_request`.
- [ ] Preserve existing routes:
  - `/v1/agent/memories`
  - `/v1/agent/memories/{memory_id}/model-io`
  - `/v1/agent/model-runs/{request_id}/io-trace`

**Verification:**

```sh
cd ../patchxNoteGoServer
GOTOOLCHAIN=go1.26.5 go test ./internal/agentaccess
```

Expected: PASS.

## GoServer Task 5: OpenAPI, Docs, Smoke

**Files:**
- Modify: `../patchxNoteGoServer/openapi/openapi.yaml` or current OpenAPI source
- Modify: `../patchxNoteGoServer/docs/integrations/apifox/agent-model-io-read-flow.zh-CN.md`
- Modify if required: `../patchxNoteGoServer/docs/integrations/apifox/patchnote-openapi.zh-CN.json`
- Modify: `../patchxNoteGoServer/tests/smoke/registry.yaml`
- Create or modify smoke module under: `../patchxNoteGoServer/tests/smoke/agent-model-io-read/`
- Add evidence file only during implementation if repository practice requires it.

**Steps:**

- [ ] Add `listAgentModelIOTraces` operation.
- [ ] Document list first, detail second:
  - list gets `request_id`;
  - detail gets fields.
- [ ] Make docs clear that list does not return raw source text or provider/model bodies.
- [ ] Add smoke case:
  - login Agent;
  - list traces by platform;
  - take a `request_id`;
  - call `/v1/agent/model-runs/{request_id}/io-trace`;
  - verify field statuses.
- [ ] Add smoke case where a trace exists without a structured memory and is still listed.
- [ ] Keep Admin `/admin/v1/model-io-traces*` unchanged.
- [ ] Keep App/PC `POST /v1/model-runs:execute` unchanged.

## GoServer Task 6: Test Server Deployment

**Files:**
- Read: `../patchxNoteGoServer/docs/engineering/runbooks/release-update-sop.md`
- Read: any current test-server deployment guide referenced by `../patchxNoteGoServer/docs/plans/CURRENT.md`
- Modify only if required by repository practice: deployment evidence under `../patchxNoteGoServer/docs/engineering/evidence/`

**Steps:**

- [ ] Build the GoServer test binary/artifacts using the repository runbook.
- [ ] Deploy to the test server only after local module and smoke tests pass.
- [ ] Verify online health/ready endpoints.
- [ ] Verify online OpenAPI exposes `GET /v1/agent/model-io-traces`.
- [ ] Verify the real test account with known `model_io_trace` rows can list traces through the Agent endpoint.
- [ ] Verify existing Agent detail endpoint still returns by `request_id`.
- [ ] Record deployment version/commit and smoke result without printing raw phone, source text, model payloads, tokens, or provider keys.

**Verification:**

```sh
cd ../patchxNoteGoServer
curl -fsS https://ws-lab.patch-x.cn/patchnote-test-api/healthz
curl -fsS https://ws-lab.patch-x.cn/patchnote-test-api/readyz
```

Expected: both checks PASS, then real Agent list/detail acceptance PASS.

**Verification:**

```sh
cd ../patchxNoteGoServer
MODULE=agent-model-io-read make test-module
MODULE=agent-model-io-read make smoke-module
```

Expected: PASS.

## Agent Task 6: API Client Types And Method

**Files:**
- Modify: `internal/api/types.go`
- Modify: `internal/api/client.go`
- Modify: `internal/api/client_test.go`
- Modify: `internal/cli/runtime.go`
- Modify: `internal/mcp/server.go`

**Steps:**

- [ ] Add `AgentModelIOTracePage`.
- [ ] Add `AgentModelIOTraceSummary`.
- [ ] Add `AgentModelIOFieldBytes`.
- [ ] Add `ListModelIOTracesParams`.
- [ ] Add client method:
  - `ListModelIOTraces(ctx, accessToken, params)`
- [ ] Request path:
  - `/v1/agent/model-io-traces`
- [ ] Validate platform is `mobile|desktop`.
- [ ] Validate limit range only when non-zero.
- [ ] Pass optional `request_id`, `task_type`, `state`, `recording_id`, `event_id`, `business_id`, `date_from`, `date_to`, `cursor`.
- [ ] Update `agentAPI` interface.
- [ ] Update MCP `AgentAPI` interface.
- [ ] Update fake APIs in tests.
- [ ] Add client tests:
  - URL/query encoding;
  - invalid platform;
  - response decoding;
  - auth header present;
  - no raw body fields required.

**Verification:**

```sh
go test ./internal/api ./internal/cli ./internal/mcp
```

Expected: PASS.

## Agent Task 7: CLI `model-io list`

**Files:**
- Modify: `internal/cli/model_io.go`
- Modify: `internal/cli/model_io_test.go`
- Modify if needed: `README.md`
- Modify if needed: `README.zh-CN.md`
- Modify if needed: `packages/npm/README.md`

**Command Shape:**

```sh
patchxnote model-io list --platform mobile
patchxnote model-io list --platform mobile --task-type event_planning --limit 10
patchxnote model-io list --platform mobile --state completed --date-from 2026-08-13T00:00:00Z
patchxnote model-io list --platform mobile --business-id local-example
patchxnote model-io list --platform mobile --output json
```

**Steps:**

- [ ] Add `list` subcommand under `model-io`.
- [ ] Flags:
  - `--platform`
  - `--request-id`
  - `--task-type`
  - `--state`
  - `--recording-id`
  - `--event-id`
  - `--business-id`
  - `--date-from`
  - `--date-to`
  - `--limit`
  - `--cursor`
- [ ] Reuse existing auth refresh path.
- [ ] For `--output json`, print the raw decoded page as JSON.
- [ ] For plain output, print a compact table:
  - request id
  - platform
  - task type
  - state
  - source text availability
  - field availability summary for provider/parsed/packaged
  - created/completed time
  - memory yes/no
- [ ] Include `next_cursor` in plain output when present.
- [ ] Do not print raw source text or model JSON in list output.
- [ ] Add tests for:
  - required platform;
  - JSON output;
  - plain output has request id but not provider body;
  - task/business/date filters are passed to API;
  - cursor passthrough;
  - auth refresh behavior.

**Verification:**

```sh
go test ./internal/cli ./internal/api
```

Expected: PASS.

## Agent Task 8: MCP Tool `patchxnote_list_model_io_traces`

**Files:**
- Modify or create: `internal/mcp/model_io_trace_tools.go`
- Modify: `internal/mcp/tools.go`
- Modify: `internal/mcp/server_test.go`
- Modify or create: `internal/mcp/model_io_trace_tools_test.go`
- Modify: `internal/mcp/tools_test.go`
- Modify: `test/e2e/mvp_test.go`

**Tool Schema:**

```json
{
  "platform": "mobile | desktop",
  "request_id": "optional string",
  "task_type": "optional string",
  "state": "optional string",
  "recording_id": "optional string",
  "event_id": "optional string",
  "business_id": "optional string",
  "date_from": "optional RFC3339 string",
  "date_to": "optional RFC3339 string",
  "limit": "optional integer 1..50",
  "cursor": "optional string"
}
```

**Steps:**

- [ ] Register `patchxnote_list_model_io_traces`.
- [ ] Mark annotations as read-only.
- [ ] Require `platform`.
- [ ] Validate unknown fields are rejected.
- [ ] Validate `request_id`, `task_type`, `state`, `recording_id`, `event_id`, `business_id`, `limit`, `date_from`, `date_to`, `cursor`.
- [ ] Use refresh-aware access token path.
- [ ] Call API `ListModelIOTraces`.
- [ ] Return concise JSON text page.
- [ ] Keep output bounded; if future fields grow, summarize rather than dumping raw payloads.
- [ ] Update expected MCP tool count:
  - current 18 + 1 = 19.
- [ ] Add MCP tests:
  - tools/list includes new tool;
  - success path;
  - invalid platform;
  - filter passthrough;
  - auth required;
  - no provider/source payload leakage.
- [ ] Update e2e smoke:
  - call `patchxnote_list_model_io_traces`;
  - take a fixture request id or assert known fixture id is present;
  - call one existing field tool by request id.

**Verification:**

```sh
go test ./internal/mcp ./internal/api
scripts/e2e/mvp-smoke.sh
```

Expected: PASS.

## Agent Task 9: Docs And User Guidance

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify: `docs/release-and-maintenance-runbook.zh-CN.md`
- Modify if needed: `docs/engineering-rules.md`

**Steps:**

- [ ] Update MCP tool count from 18 to 19.
- [ ] Add `patchxnote model-io list` examples.
- [ ] Add `patchxnote_list_model_io_traces` MCP example.
- [ ] Explain:
  - `memory_id` is for structured memories;
  - `request_id` is for model IO trace;
  - list traces first when memory list is empty.
- [ ] Explain common filters:
  - task type;
  - business id;
  - recording id;
  - date range.
- [ ] Document that list output is metadata only.
- [ ] Document that field commands/tools may return source text and provider/model payloads for the logged-in user.
- [ ] Document recommended AI flow:
  - list traces;
  - choose latest completed trace;
  - fetch specific fields.

**Verification:**

```sh
git diff --check
```

Expected: no output.

## Agent Task 10: Real Local Acceptance

**Prerequisites:**

- [ ] GoServer test server has the new `GET /v1/agent/model-io-traces` endpoint deployed.
- [ ] A real logged-in Agent account has at least one `model_io_trace`.
- [ ] Do not print source text, provider request, provider response, parsed result, packaged result, OTP, token, phone, or webhook URL in terminal logs.

**Steps:**

- [ ] Build current Windows CLI binary.
- [ ] Install through npm wrapper into the normal local install path.
- [ ] Confirm `patchxnote version`.
- [ ] Confirm `patchxnote auth status --output json` is authenticated.
- [ ] Run:

```sh
patchxnote model-io list --platform mobile --output json
```

- [ ] Confirm at least one `request_id` is returned for a real trace account.
- [ ] Run source text to a temp file:

```sh
patchxnote model-io source-text --request-id <request_id> --platform mobile --out <temp-source.txt> --force
```

- [ ] Run provider response to a temp JSON file:

```sh
patchxnote model-io provider-response --request-id <request_id> --platform mobile --out <temp-provider.json> --force
```

- [ ] Run parsed result to a temp JSON file.
- [ ] Run packaged result to a temp JSON file.
- [ ] Confirm files exist and JSON files parse.
- [ ] Run MCP `tools/list`; confirm tool count is 19.
- [ ] Run MCP `patchxnote_list_model_io_traces`.
- [ ] Run MCP one field tool by returned `request_id`.
- [ ] Confirm list output did not include raw source/model payloads.
- [ ] Confirm field output/write paths are scoped to the requested field only.

## Verification Checklist

- [ ] GoServer: `GOTOOLCHAIN=go1.26.5 go test ./internal/agentaccess`
- [ ] GoServer: `MODULE=agent-model-io-read make test-module`
- [ ] GoServer: `MODULE=agent-model-io-read make smoke-module`
- [ ] GoServer: deploy to test server and verify online `GET /v1/agent/model-io-traces`
- [ ] Agent: `go test ./internal/api ./internal/cli ./internal/mcp`
- [ ] Agent: `go test ./...`
- [ ] Agent: `scripts/e2e/mvp-smoke.sh`
- [ ] Agent: `node.exe packages/npm/test/install.test.js`
- [ ] Agent: `git diff --check`
- [ ] Secret scan for raw phone, OTP, access token, refresh token, webhook URL fragments, provider keys, source text fixtures from real users, and provider payloads.
- [ ] Real installed CLI/MCP acceptance against test server.
- [ ] Confirm App/PC/Admin routes are unchanged except OpenAPI docs adding the Agent-only route.

## Edge Cases

- [ ] Missing `platform`.
- [ ] Invalid `platform`.
- [ ] Unknown query parameter.
- [ ] Duplicate query parameter values.
- [ ] `request_id` invalid.
- [ ] `task_type` invalid.
- [ ] `state` invalid.
- [ ] `recording_id` invalid.
- [ ] `event_id` invalid.
- [ ] `business_id` invalid.
- [ ] `date_from` invalid timestamp.
- [ ] `date_to` invalid timestamp.
- [ ] `date_from >= date_to`.
- [ ] `limit=0`.
- [ ] `limit>50`.
- [ ] Invalid cursor base64.
- [ ] Cursor from another platform.
- [ ] Cursor from another task/state/business/date filter.
- [ ] Cursor after deleted/missing row.
- [ ] Account has no traces.
- [ ] Account has traces but no structured memories.
- [ ] Account has traces while mobile consent is enabled but no memory exists.
- [ ] Account has traces while memory mapping is absent because result promotion is not applicable.
- [ ] Account has traces on both platforms.
- [ ] Account lacks `agent:content.read:<platform>`.
- [ ] Trace has provider response but no source text.
- [ ] Trace source text availability is `not_applicable`.
- [ ] Trace source text availability is `not_recorded`.
- [ ] Trace source text availability is `truncated`.
- [ ] Trace has parsed result but no packaged result.
- [ ] Trace is `provider_failed`.
- [ ] Trace is `trace_failed`.
- [ ] Trace is still `executing`.
- [ ] Trace is `completed` but a JSON field is missing.
- [ ] Trace has memory mapping.
- [ ] Trace has no memory mapping.
- [ ] Large provider response exists; list still returns only metadata.
- [ ] MCP stdout remains JSON-RPC only.
- [ ] CLI `--output json` remains machine-readable.

## Out Of Scope

- [ ] No model execution.
- [ ] No model replay or retry.
- [ ] No mutation of `model_io_trace`.
- [ ] No backfill from `model_io_trace` into structured memories.
- [ ] No change to App/PC `model-runs:execute`.
- [ ] No change to Admin model IO trace APIs.
- [ ] No batch export of raw provider/model payloads from the list endpoint.
- [ ] No automatic webhook sending.
- [ ] No custom AI rewriting.

## Implementation Order

1. [ ] Implement GoServer `GET /v1/agent/model-io-traces`.
2. [ ] Verify the endpoint locally against a fixture that has traces but no memories.
3. [ ] Deploy GoServer to the test server.
4. [ ] Verify the endpoint online against a real account that has traces but no memories.
5. [ ] Add Agent API client method.
6. [ ] Add CLI `patchxnote model-io list`.
7. [ ] Add MCP `patchxnote_list_model_io_traces`.
8. [ ] Update docs and e2e.
9. [ ] Run local installed acceptance end to end.
