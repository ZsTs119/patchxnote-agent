# Remote MCP GoServer Parity Implementation Plan

> **For Codex:** Execute this plan sequentially, task by task. Follow repository `AGENTS.md`: sub-agents and parallel task execution are disabled.

**Goal:** Add a GoServer-hosted remote MCP gateway that gives platform agents the same PatchXNote MCP tool capabilities as the local npm/stdio Agent.

**Architecture:** Keep `patchnote-agent` as the local CLI/npm/stdio runtime. Add a `patchxNoteGoServer` remote MCP HTTP gateway that exposes the same 19 MCP tools over HTTPS, uses official MCP/OAuth authorization discovery, and reuses existing `internal/agentaccess` service methods wherever possible. Local-only file/config state becomes server-side state or bounded response content in remote mode.

**Tech Stack:** Go, chi HTTP router, PostgreSQL migrations, existing GoServer `agentaccess` services, OpenAPI/Apifox artifacts, MCP JSON-RPC over Streamable HTTP, OAuth 2.1 authorization code with PKCE, protected resource metadata, existing smoke-module gate.

---

## Confirmed Decisions

- [x] Remote MCP belongs in `patchxNoteGoServer`, not in npm.
- [x] Test endpoint can first live under `https://ws-lab.patch-x.cn/patchnote-test-api/mcp`.
- [x] Production should later use a cleaner MCP subdomain, such as `https://mcp.patch-x.cn/mcp` or `https://mcp.patchxnote.com/mcp`.
- [x] Remote MCP must be functionally equivalent to local MCP, not a smaller tool subset.
- [x] The remote platform path should use the official MCP/OAuth style instead of asking users to paste PatchXNote access tokens into third-party platforms.
- [x] Platforms that can run a local command or cloud-computer terminal may still use the existing npm command.
- [x] Pure platform connectors should use the hosted remote MCP URL.

## Functional Parity Definition

Remote MCP is accepted only when these are true:

- [ ] `tools/list` exposes the same 19 `patchxnote_*` tool names as local MCP.
- [ ] Tool descriptions, input schemas, required fields, bounds, and error semantics match local MCP unless a local filesystem field is explicitly documented as remote-inapplicable.
- [ ] Tool schemas have a parity test against the current local MCP schema fixture, so future local MCP changes cannot silently drift from remote MCP.
- [ ] Read tools return the same logical data shape as local MCP.
- [ ] Webhook tools exist in remote MCP and use GoServer-backed storage instead of local files/keychain.
- [ ] Export/render tools can return bounded content directly when no local filesystem exists.
- [ ] No remote tool consumes or replaces App/PC `client_installation`.
- [ ] Mobile and desktop content remain platform-scoped; remote MCP must not merge them implicitly.
- [ ] `tools/call` side-effect tools are idempotent under platform retries.
- [ ] Remote MCP tool output never treats user content as tool instructions; user-generated transcript/summary/model content is returned only as data.

## Audit Additions 2026-08-27

These gaps were found during plan review and are now part of the execution scope:

- [ ] Add an explicit scope matrix per MCP tool and show it on the OAuth consent page.
- [ ] Add connector session listing/revocation APIs for users/operators, not only token revocation.
- [ ] Harden OAuth with exact redirect URI allowlists, `state`, PKCE verifier expiry, one-time codes, refresh-token rotation, and reuse detection.
- [ ] Support both root and path-prefixed `.well-known` metadata URLs for the current `/patchnote-test-api` test deployment.
- [ ] Add notification handling for `notifications/initialized` and other JSON-RPC notifications common clients may send.
- [ ] Add idempotency/replay protection for remote webhook configure/remove/send tools.
- [ ] Add response contract for remote file-output fields: no silent ignore, no local path writes, bounded returned content or short-lived handle only.
- [ ] Add cleanup/retention jobs for OAuth codes, connector tokens, connector sessions, webhook send audit, and expired download handles.
- [ ] Add secret/key rotation notes for connector token signing, webhook target encryption, and OAuth client secrets.
- [ ] Add Caddy/reverse-proxy path-prefix acceptance checks before platform validation.
- [ ] Add platform tool-cache/schema-version handling so stale `tools/list` does not break callers during rollout.
- [ ] Add SSRF/DNS-rebinding protections for server-side webhook sends.

## Required Tool Surface

Remote MCP must expose all current local MCP tools:

- [ ] `patchxnote_get_current_user`
- [ ] `patchxnote_list_recorder_cards`
- [ ] `patchxnote_get_quota_summary`
- [ ] `patchxnote_get_model_usage_summary`
- [ ] `patchxnote_list_memories`
- [ ] `patchxnote_search_memories`
- [ ] `patchxnote_get_memory`
- [ ] `patchxnote_list_webhook_targets`
- [ ] `patchxnote_configure_webhook_target`
- [ ] `patchxnote_remove_webhook_target`
- [ ] `patchxnote_list_webhook_templates`
- [ ] `patchxnote_render_webhook_message`
- [ ] `patchxnote_export_model_io`
- [ ] `patchxnote_send_webhook`
- [ ] `patchxnote_list_model_io_traces`
- [ ] `patchxnote_get_model_io_source_text`
- [ ] `patchxnote_get_model_io_provider_response`
- [ ] `patchxnote_get_model_io_parsed_result`
- [ ] `patchxnote_get_model_io_packaged_result`

## OAuth Scope Matrix

OAuth consent and connector tokens must carry explicit scopes. Do not infer permission only from the fact that a platform can reach `/mcp`.

| Scope | Tools | Notes |
| --- | --- | --- |
| `agent:account.read` | `patchxnote_get_current_user` | Existing Agent scope. |
| `agent:hardware.read` | `patchxnote_list_recorder_cards` | Existing Agent scope; output remains masked only. |
| `agent:quota.read` | `patchxnote_get_quota_summary` | Existing Agent scope. |
| `agent:model_usage.read` | `patchxnote_get_model_usage_summary` | Existing Agent scope. |
| `agent:content.read:mobile` | `patchxnote_list_memories`, `patchxnote_search_memories`, `patchxnote_get_memory`, model IO tools for `mobile` | Existing content scope; platform parameter still required where local MCP requires it. |
| `agent:content.read:desktop` | `patchxnote_list_memories`, `patchxnote_search_memories`, `patchxnote_get_memory`, model IO tools for `desktop` | Existing content scope; never merge with mobile. |
| `agent:model_io.read` | `patchxnote_list_model_io_traces`, `patchxnote_get_model_io_source_text`, `patchxnote_get_model_io_provider_response`, `patchxnote_get_model_io_parsed_result`, `patchxnote_get_model_io_packaged_result`, `patchxnote_export_model_io` | New explicit remote consent scope for model IO fields; still also requires matching platform content scope. |
| `agent:webhook.read` | `patchxnote_list_webhook_targets`, `patchxnote_list_webhook_templates` | New remote scope; list output is masked. |
| `agent:webhook.write` | `patchxnote_configure_webhook_target`, `patchxnote_remove_webhook_target`, `patchxnote_render_webhook_message` with server-side saved state | New remote scope; configure inputs are write-only. |
| `agent:webhook.send` | `patchxnote_send_webhook` | New explicit external-send scope; also requires `agent:webhook.read` to resolve aliases. |

Scope rules:

- [ ] Consent page displays platform name, requested scopes, and high-risk scopes such as `agent:model_io.read` and `agent:webhook.send`.
- [ ] Users can revoke the whole connector session.
- [ ] Scope downgrades create a new consent version or revoke/reissue the connector token.
- [ ] `tools/list` may show all 19 tools for parity, but `tools/call` must return `forbidden` for missing scopes.
- [ ] Error text for missing scope must name the missing scope without leaking user data.

## Remote Contract Adaptations

Remote MCP keeps the same tool names and logical capabilities, but must make these local-environment differences explicit:

| Local MCP Behavior | Remote MCP Behavior |
| --- | --- |
| Credentials live in local secure storage. | Connector credentials live in GoServer token/session storage. |
| Webhook targets live in local config/keychain. | Webhook targets live in GoServer account-scoped encrypted storage. |
| `patchxnote_search_memories` searches current local MCP metadata cache. | Remote search queries server-authorized safe metadata because HTTP calls may be stateless. |
| `patchxnote_render_webhook_message` can write a draft directory. | Remote render returns Markdown/content fields or a short-lived download handle. |
| `patchxnote_export_model_io` requires `out` for local file export. | Remote export returns bounded JSON/content fields or a short-lived download handle; local path fields are rejected with a clear MCP error. |
| `patchxnote_send_webhook` sends from the user's machine. | Remote send originates from GoServer and must enforce SSRF, timeout, payload, and idempotency controls. |

Remote schema rule:

- [ ] Prefer adding backward-compatible optional fields such as `return_content` or `delivery_mode` only if one shared schema cannot describe both local and remote modes cleanly.
- [ ] If a remote tool cannot honor `out`, `save_draft`, `draft_dir`, or other local path fields, return `invalid_request`; do not ignore the field.
- [ ] If content is too large for MCP output, return `content_truncated`, byte counts, and a short-lived handle instead of dumping the full body.

## Client Support Matrix

| Client / Platform | First Path | Remote Path | V1 Acceptance |
| --- | --- | --- | --- |
| VS Code / GitHub Copilot | npm stdio | optional later | local setup already owns V1 |
| Cursor | npm stdio / install link | optional later | local setup already owns V1 |
| Codex / ChatGPT Desktop / Codex IDE | npm stdio / `codex mcp add` | optional later | local setup already owns V1 |
| Claude Code | npm stdio / `claude mcp add` | optional later | local setup already owns V1 |
| Claude Desktop | local config file | optional later | local setup already owns V1 |
| Windsurf | local config file | optional later | local setup already owns V1 |
| Trae | local command/config | optional later | manual command V1 |
| Qoder | local command/config | optional later | manual command V1 |
| CodeBuddy CLI | npm stdio | HTTP/SSE optional | validate after release |
| WorkBuddy local/CLI mode | npm stdio if command runner exists | HTTP if platform connector | verify actual client mode |
| Feishu Aily | maybe local/cloud computer command | hosted MCP URL | remote gateway required for pure platform connector |
| Doubao Work Partner | local/cloud computer command if available | hosted MCP URL | support both paths where product allows |
| Tencent Agent Development Platform | not local by default | hosted MCP URL | remote gateway required |
| Enterprise WorkBuddy | tenant-dependent | hosted MCP URL | remote gateway required unless local runner is confirmed |

## Endpoint Shape

External test URLs:

```text
GET  https://ws-lab.patch-x.cn/patchnote-test-api/mcp/health
POST https://ws-lab.patch-x.cn/patchnote-test-api/mcp
```

Future production URLs:

```text
GET  https://mcp.patch-x.cn/health
POST https://mcp.patch-x.cn/mcp
```

Discovery and auth URLs:

```text
GET  /.well-known/oauth-protected-resource
GET  /.well-known/oauth-protected-resource/mcp
GET  /.well-known/oauth-protected-resource/patchnote-test-api/mcp
GET  /.well-known/oauth-authorization-server
GET  /v1/agent/oauth/authorize
POST /v1/agent/oauth/token
POST /v1/agent/oauth/revoke
```

Implementation note:

- [ ] In the existing test-domain path-prefix deployment, make the public URLs resolve under `/patchnote-test-api/...`.
- [ ] Return `WWW-Authenticate` with `resource_metadata` from unauthenticated `/mcp` responses.
- [ ] For test endpoint `https://ws-lab.patch-x.cn/patchnote-test-api/mcp`, point `resource_metadata` to an exact reachable URL, preferably `https://ws-lab.patch-x.cn/.well-known/oauth-protected-resource/patchnote-test-api/mcp`.
- [ ] Also support `/patchnote-test-api/.well-known/oauth-protected-resource/mcp` as a compatibility fallback if a platform preserves the deployment prefix.
- [ ] Return `authorization_servers` from protected resource metadata.
- [ ] Keep root/subdomain well-known behavior simple in test, then tighten it when the production MCP subdomain is created.
- [ ] Add a route test that fails if Caddy/path-prefix rewriting makes metadata URLs point at non-existent paths.

## Scope Split

### Local MCP Stays As Is

- [ ] Keep `npx -y patchxnote-agent@latest mcp serve` as the universal local stdio entrypoint.
- [ ] Keep stdout as JSON-RPC only.
- [ ] Keep local credentials in OS-native secure storage.
- [ ] Keep local setup/client config work in `patchnote-agent`.

### Remote MCP Added In GoServer

- [ ] Add hosted Streamable HTTP MCP route.
- [ ] Add OAuth connector authorization.
- [ ] Add server-side webhook target storage.
- [ ] Add bounded content-return mode for render/export tools.
- [ ] Add platform-client audit and revocation.
- [ ] Add deployment and platform acceptance evidence.

---

## Task 0: Baseline And Required Reads

**Files:**
- Read: `/home/zsts_119/patchxNoteGoServer/AGENTS.md`
- Read: `/home/zsts_119/patchxNoteGoServer/README.md`
- Read: `/home/zsts_119/patchxNoteGoServer/docs/README.md`
- Read: `/home/zsts_119/patchxNoteGoServer/docs/product-engineering-profile.md`
- Read: `/home/zsts_119/patchxNoteGoServer/docs/engineering/api-module-tdd-smoke-gate.md`
- Read: `/home/zsts_119/patchxNoteGoServer/docs/plans/CURRENT.md`
- Read: `/home/zsts_119/patchxNoteGoServer/docs/integrations/apifox/shared/integration-guide.zh-CN.md`
- Read: `/home/zsts_119/patchnote-agent/internal/mcp/tools.go`
- Read: `/home/zsts_119/patchnote-agent/internal/mcp/webhook_tools.go`
- Read: `/home/zsts_119/patchnote-agent/internal/mcp/model_io_tools.go`

**Checklist:**
- [ ] Verify current GoServer branch, HEAD, upstream, and dirty state.
- [ ] Verify current `patchnote-agent` branch, HEAD, upstream, and dirty state.
- [ ] Record current npm latest and local MCP tool count.
- [ ] Record current GoServer OpenAPI version and latest migration number.
- [ ] Confirm test base URL is reachable: `https://ws-lab.patch-x.cn/patchnote-test-api`.

**Commands:**

```sh
cd /home/zsts_119/patchxNoteGoServer
git status --short
git rev-parse --abbrev-ref HEAD
git rev-parse HEAD
grep -n "version:" openapi/openapi.yaml | head
ls migrations | tail
curl -fsS https://ws-lab.patch-x.cn/patchnote-test-api/healthz
```

**Acceptance:**
- [ ] Baseline is written into the implementation notes before code changes.
- [ ] Any unrelated dirty files are identified and left untouched.

---

## Task 1: Capture The Local MCP Contract

**Files:**
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/tool_contract.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/tool_contract_test.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/testdata/local_tools_19.json`
- Create: `/home/zsts_119/patchxNoteGoServer/scripts/dev/export-local-mcp-tools.mjs` or a Go equivalent
- Reference: `/home/zsts_119/patchnote-agent/internal/mcp/tools.go`
- Reference: `/home/zsts_119/patchnote-agent/internal/mcp/webhook_tools.go`
- Reference: `/home/zsts_119/patchnote-agent/internal/mcp/model_io_tools.go`

**Checklist:**
- [ ] Define the 19 remote MCP tool names in one ordered slice.
- [ ] Define required input schemas with the same required fields and bounds as local MCP.
- [ ] Mark read-only tools as read-only.
- [ ] Mark webhook configure/remove/render/export as write or local/remote state changing.
- [ ] Mark `patchxnote_send_webhook` as external network send.
- [ ] Add a test that fails if the tool count is not 19.
- [ ] Add a test that fails if any expected local tool name is missing.
- [ ] Add a test that every remote tool name starts with `patchxnote_`.
- [ ] Generate a local MCP `tools/list` fixture from a real local Agent binary or source test server.
- [ ] Compare remote tool names, descriptions, required properties, bounds, and annotations with the fixture.
- [ ] Store a contract hash in test output/evidence so platform cache issues can be diagnosed.

**Acceptance:**
- [ ] `go test ./internal/remotemcp -run TestToolContract -count=1` passes.
- [ ] Fixture regeneration produces no diff unless the local MCP contract intentionally changed.

---

## Task 2: Add Remote MCP Protocol Handler

**Files:**
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/jsonrpc.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/handler.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/errors.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/handler_test.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/cmd/api/main.go`

**Checklist:**
- [ ] Support `POST /mcp`.
- [ ] Support `GET /mcp/health`.
- [ ] Decode one JSON-RPC 2.0 request object per HTTP request.
- [ ] Reject batch requests in V1 with `invalid_request`.
- [ ] Accept JSON-RPC notifications without `id`; return the transport response expected by the selected MCP protocol revision without inventing a JSON-RPC response id.
- [ ] Treat `notifications/initialized` as a no-op success path for clients that still send it after `initialize`.
- [ ] Support `initialize`.
- [ ] Support `tools/list`.
- [ ] Support `tools/call`.
- [ ] Support `ping` if a target platform probes it.
- [ ] Return JSON-RPC errors without leaking upstream raw bodies.
- [ ] Distinguish transport/auth failures from MCP method/tool failures: unauthenticated HTTP requests can be `401`, while authenticated `tools/call` business failures should normally return a JSON-RPC result with `isError` or a JSON-RPC error according to local MCP semantics.
- [ ] Keep HTTP response headers `Cache-Control: no-store` and `X-Content-Type-Options: nosniff`.
- [ ] Enforce `Content-Type: application/json` for non-SSE POST requests.
- [ ] Return `Accept`/content-type behavior compatible with clients that ask for `application/json` or Streamable HTTP responses.
- [ ] Cap request body size with existing router max body rules or a route-specific override.
- [ ] Add request timeout handling so abandoned platform requests do not keep DB or webhook work running unbounded.

**Protocol compatibility notes:**
- [ ] Accept current common MCP clients that still send `initialize`.
- [ ] Accept `MCP-Protocol-Version` if present.
- [ ] Validate `Mcp-Method` and `Mcp-Name` where the 2026-07-28 client sends them.
- [ ] If `Mcp-Method`/`Mcp-Name` disagree with the JSON-RPC body, reject the request before calling any tool.
- [ ] Do not require SSE unless Feishu/Tencent/WorkBuddy acceptance proves it is needed.
- [ ] If SSE is added later, keep it behind the same auth/scope/rate-limit code path as Streamable HTTP.

**Acceptance:**
- [ ] Unauthenticated `initialize` returns protocol metadata if no user data is accessed.
- [ ] Unauthenticated `tools/list` either returns public tool metadata or returns a proper auth challenge; choose one and document it.
- [ ] Authenticated `tools/list` returns exactly 19 tools.
- [ ] Malformed JSON returns JSON-RPC parse/invalid-request error.
- [ ] Unknown method returns method-not-found style MCP error.
- [ ] `notifications/initialized` does not break clients.
- [ ] Header/body method mismatch never reaches a tool handler.

---

## Task 3: Add Official MCP OAuth Discovery

**Files:**
- Create: `/home/zsts_119/patchxNoteGoServer/internal/agentoauth/domain.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/agentoauth/service.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/agentoauth/http.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/agentoauth/http_test.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/cmd/api/main.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/config.yaml.example`
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/platform/config/*.go`

**Checklist:**
- [ ] Add protected resource metadata endpoint.
- [ ] Add authorization server metadata endpoint.
- [ ] Add `GET /v1/agent/oauth/authorize`.
- [ ] Add `POST /v1/agent/oauth/token`.
- [ ] Add `POST /v1/agent/oauth/revoke`.
- [ ] Add `GET /v1/agent/oauth/connectors` for current-user connector listing.
- [ ] Add `DELETE /v1/agent/oauth/connectors/{connector_session_id}` or equivalent revocation API for user/operator cleanup.
- [ ] Implement authorization code flow with PKCE.
- [ ] Support pre-registered platform clients with `client_id`.
- [ ] Support confidential client validation when a platform requires `client_secret`.
- [ ] Do not require Dynamic Client Registration in V1 unless a target platform forces it.
- [ ] Prefer pre-registered clients and client metadata documents over open dynamic client registration.
- [ ] Store only hashed client secrets.
- [ ] Validate exact redirect URI against the registered allowlist; no wildcard path, open redirect, or scheme downgrade.
- [ ] Require and verify OAuth `state`.
- [ ] Bind authorization code to `client_id`, redirect URI, code challenge, scope set, account id, consent version, and expiry.
- [ ] Use one-time authorization codes and delete or mark them redeemed atomically.
- [ ] Rotate connector refresh tokens on every refresh and detect refresh-token reuse.
- [ ] Bind access/refresh tokens to issuer/audience/client id so they cannot be replayed against normal Agent APIs.
- [ ] Add a browser login/consent page that can ask for phone OTP if the user has no web session.
- [ ] Consent copy must clearly state that the platform can read PatchXNote Agent data and call configured webhook tools.
- [ ] Tokens issued for remote MCP must be connector tokens, not raw Agent access/refresh tokens.
- [ ] Token TTL and refresh behavior must be configurable.
- [ ] Token responses must include only OAuth connector token fields; never include PatchXNote Agent refresh/access tokens.
- [ ] If a platform only supports static headers in V1, support a temporary connector-token mode only as an internal test fallback and mark it non-production.

**Acceptance:**
- [ ] `/mcp` unauthenticated returns `401` with `WWW-Authenticate` metadata pointer.
- [ ] Protected resource metadata includes the authorization server URL.
- [ ] Authorization server metadata includes authorization, token, revocation, and supported scopes.
- [ ] Authorization code cannot be redeemed twice.
- [ ] PKCE mismatch fails.
- [ ] Redirect URI mismatch fails.
- [ ] State mismatch fails or is rejected by the browser flow before token issuance.
- [ ] Refresh token reuse revokes the connector session.
- [ ] Connector listing returns only platform/client/session metadata and scopes, never tokens.
- [ ] Revoked connector token cannot call `/mcp`.

---

## Task 4: Add Connector Session Storage

**Files:**
- Create: `/home/zsts_119/patchxNoteGoServer/migrations/000056_agent_mcp_connector_sessions.up.sql`
- Create: `/home/zsts_119/patchxNoteGoServer/migrations/000056_agent_mcp_connector_sessions.down.sql`
- Modify: `/home/zsts_119/patchxNoteGoServer/migrations/checksums.sha256`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/agentoauth/repository.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/agentoauth/repository_integration_test.go`

**Migration note:**
- [ ] Recheck the latest migration number immediately before implementation. If `000056` or `000057` already exists, use the next available number and update this plan in-place before coding.

**Schema checklist:**
- [ ] `agent_mcp_oauth_client`
- [ ] `agent_mcp_authorization_code`
- [ ] `agent_mcp_connector_session`
- [ ] `agent_mcp_connector_token`
- [ ] Store token hashes only.
- [ ] Store account id, connector session id, platform client id, scopes, consent version, status, created/updated/expires timestamps.
- [ ] Store user-facing connector name and source client, for example `feishu_aily`, `doubao_work_partner`, `tencent_adp`, `workbuddy_enterprise`.
- [ ] Store platform tenant/workspace identifiers only when supplied by the platform and safe to persist.
- [ ] Store exact redirect URI on authorization code redemption records.
- [ ] Store refresh-token predecessor hash to detect reuse.
- [ ] Add revocation fields.
- [ ] Add indexes by account, session, token hash, refresh token hash, client id, status, expiry, and cleanup due time.
- [ ] Add uniqueness constraints that prevent two active sessions from sharing the same connector token hash.

**Acceptance:**
- [ ] Migration applies and rolls back cleanly on an isolated test database.
- [ ] Repository tests prove create/redeem/revoke/expire behavior.
- [ ] No plaintext token is persisted.
- [ ] Cleanup query can find expired codes/tokens/sessions without table scans.

---

## Task 5: Bind Remote MCP Requests To Connector Actors

**Files:**
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/handler.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/agentaccess/service.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/agentaccess/domain.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/auth_test.go`

**Checklist:**
- [ ] Parse `Authorization: Bearer <connector_access_token>` in `/mcp`.
- [ ] Resolve connector token to PatchXNote account id and connector session.
- [ ] Do not pass connector tokens into `agentaccess.ParseAccessToken`; connector tokens are a separate OAuth audience.
- [ ] Choose and document one principal strategy before implementation:
  - Option A: create a first-class connector principal adapter that lets `agentaccess` validate account/scopes without requiring a local `agent_session` token;
  - Option B: create an internal `agent_session`-compatible record for remote MCP connector sessions, without ever exposing Agent access/refresh tokens to the platform.
- [ ] Map connector session scopes to the chosen `agentaccess` principal shape.
- [ ] Reject Client/Admin/Provider tokens.
- [ ] Attach safe debug metadata: account id, connector session id, platform client id, tool name, request id.
- [ ] Do not log raw token, raw phone, OTP, prompt, transcript, provider payload, webhook URL, or secret.
- [ ] Validate platform client status before every `tools/call`, so disabled clients stop immediately.

**Acceptance:**
- [ ] Wrong audience token fails.
- [ ] Expired connector token fails.
- [ ] Revoked connector token fails.
- [ ] Token with missing scope cannot call the corresponding tool.
- [ ] Same account isolation is covered by tests.
- [ ] Connector token cannot call existing `/v1/agent/...` REST endpoints unless explicitly designed and documented.

---

## Task 6: Implement Read Tool Handlers

**Files:**
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/read_tools.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/read_tools_test.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/agentaccess/service.go` if `SearchMemories` is added.
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/agentaccess/repository.go` if `SearchMemories` is added.

**Tool mapping:**

| MCP Tool | GoServer Source |
| --- | --- |
| `patchxnote_get_current_user` | `agentaccess.Service.Me` |
| `patchxnote_list_recorder_cards` | `agentaccess.Service.ListRecorderCards` |
| `patchxnote_get_quota_summary` | `agentaccess.Service.QuotaSummary` |
| `patchxnote_get_model_usage_summary` | `agentaccess.Service.ModelUsageSummary` |
| `patchxnote_list_memories` | `agentaccess.Service.ListMemories` |
| `patchxnote_get_memory` | `agentaccess.Service.GetMemory` |
| `patchxnote_list_model_io_traces` | `agentaccess.Service.ListModelIOTraces` |

**Search checklist:**
- [ ] Do not use local MCP session cache remotely; remote HTTP may be stateless.
- [ ] Add server-side `SearchMemories` over safe metadata fields only: title, summary, object type, source, request id.
- [ ] Keep platform required.
- [ ] Keep limit max 50.
- [ ] Do not search raw transcript, prompt, provider request, provider response, parsed payload, packaged payload, or raw audio.
- [ ] Use deterministic ordering and stable cursor encoding for remote search/list responses.
- [ ] Bind every cursor to account id, platform, filter fingerprint, and expiry or reject stale/foreign cursors.
- [ ] Keep empty result and not-found behavior aligned with local MCP text/structured output.

**Acceptance:**
- [ ] All read tools return MCP `CallToolResult` with structured JSON and concise text.
- [ ] Invalid platform fails.
- [ ] Limit over 50 fails.
- [ ] Cursor over max length fails.
- [ ] Cross-platform content is rejected or empty as appropriate.
- [ ] Cursor from another account/platform/filter fails safely.

---

## Task 7: Implement Model IO Field Tools

**Files:**
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/model_io_tools.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/model_io_tools_test.go`

**Checklist:**
- [ ] Implement `patchxnote_get_model_io_source_text`.
- [ ] Implement `patchxnote_get_model_io_provider_response`.
- [ ] Implement `patchxnote_get_model_io_parsed_result`.
- [ ] Implement `patchxnote_get_model_io_packaged_result`.
- [ ] Resolve by `memory_id` or `request_id` consistently with local MCP.
- [ ] Support optional `platform` where existing Agent API supports it.
- [ ] Return bounded content directly in remote mode.
- [ ] If content exceeds response cap, return safe truncation metadata or a short-lived download handle.
- [ ] Never write raw model IO fields to logs or evidence.
- [ ] Include `field_status`, original byte size, returned byte size, and truncation/download-handle status in every field response.
- [ ] Make `agent:model_io.read` consent copy explicit that model provider response and parsed/packaged result may contain user-generated content.
- [ ] Redact only transport/secrets; do not mutate legitimate model output content unless a documented output cap requires truncation.
- [ ] Do not add provider request export unless it already exists in local MCP; 19-tool parity is the boundary.

**Acceptance:**
- [ ] Each field tool passes unit tests for memory id path.
- [ ] Each field tool passes unit tests for request id path.
- [ ] Missing field returns stable not-found/unavailable output.
- [ ] Oversized field does not exceed the MCP response cap.
- [ ] Evidence fixtures use synthetic content and contain no real model/user payload.

---

## Task 8: Add Server-Side Webhook Target Storage

**Files:**
- Create: `/home/zsts_119/patchxNoteGoServer/migrations/000057_agent_mcp_webhook_targets.up.sql`
- Create: `/home/zsts_119/patchxNoteGoServer/migrations/000057_agent_mcp_webhook_targets.down.sql`
- Modify: `/home/zsts_119/patchxNoteGoServer/migrations/checksums.sha256`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/agentwebhook/domain.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/agentwebhook/repository.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/agentwebhook/service.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/agentwebhook/repository_integration_test.go`

**Checklist:**
- [ ] Store webhook targets per PatchXNote account.
- [ ] Store alias, type, enabled, template, source client, timestamps, and state version.
- [ ] Encrypt webhook URL and signing secret with existing server envelope/secret mechanism.
- [ ] Return only masked webhook metadata in list responses.
- [ ] Add optimistic versioning if updates can race.
- [ ] Reject alias collisions within one account.
- [ ] Reject non-HTTPS webhook URLs.
- [ ] Reject localhost, loopback, link-local, and private-network webhook targets.
- [ ] Re-resolve target DNS immediately before send and reject private, link-local, loopback, multicast, and unspecified IPs after resolution.
- [ ] Disable redirects or re-validate every redirect target before following.
- [ ] Store only normalized target host metadata for display/audit; never store raw URL in logs.

**Acceptance:**
- [ ] Configure target creates and updates one alias.
- [ ] Remove target revokes/deletes one alias.
- [ ] Listing never returns raw URL or signing secret.
- [ ] Secret rotation does not expose old secret.
- [ ] DNS rebinding and private-IP target tests fail safely.

---

## Task 9: Implement Remote Webhook MCP Tools

**Files:**
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/webhook_tools.go`
- Create: `/home/zsts_119/patchxNoteGoServer/internal/remotemcp/webhook_tools_test.go`
- Reuse: `/home/zsts_119/patchxNoteGoServer/internal/agentwebhook/service.go`

**Checklist:**
- [ ] Implement `patchxnote_list_webhook_targets`.
- [ ] Implement `patchxnote_configure_webhook_target`.
- [ ] Implement `patchxnote_remove_webhook_target`.
- [ ] Implement `patchxnote_list_webhook_templates`.
- [ ] Implement `patchxnote_render_webhook_message`.
- [ ] Implement `patchxnote_export_model_io`.
- [ ] Implement `patchxnote_send_webhook`.
- [ ] `render_webhook_message` returns Markdown content directly when remote.
- [ ] `export_model_io` returns bounded JSON content directly when remote.
- [ ] `send_webhook` sends from GoServer to server-side stored targets.
- [ ] Add timeout, max payload size, and stable error mapping for external sends.
- [ ] External send audit records target alias, type, status, latency bucket, and request id only.
- [ ] Add idempotency key support for `configure`, `remove`, and `send` paths.
- [ ] For `send`, replay the same result for the same connector session, tool, idempotency key, and request fingerprint.
- [ ] Reject idempotency-key reuse with a different request fingerprint.
- [ ] If a platform does not pass an idempotency key, use a short replay window keyed by connector session, JSON-RPC id, and request fingerprint; document the reduced guarantee.

**Remote file-field behavior:**
- [ ] `out`, `save_draft`, and `force` are accepted only where they make sense.
- [ ] If a platform sends local file fields to remote MCP, return a clear `invalid_request` explaining that remote mode returns content instead of writing local files.
- [ ] Do not silently ignore file-write parameters.

**Acceptance:**
- [ ] Webhook target CRUD works with masked outputs.
- [ ] Render returns Markdown content and does not write files.
- [ ] Export returns bounded JSON/content and does not write files.
- [ ] Send works with a local test webhook receiver.
- [ ] Repeated `send_webhook` retry does not duplicate delivery when idempotency data is present.
- [ ] Failed webhook send returns a structured tool error without leaking URL or secret.

---

## Task 10: Update Local Agent Schema Only Where Needed

**Files:**
- Modify only if required: `/home/zsts_119/patchnote-agent/internal/mcp/webhook_tools.go`
- Modify only if required: `/home/zsts_119/patchnote-agent/internal/webhook/*.go`
- Modify only if required: `/home/zsts_119/patchnote-agent/internal/modelio/*.go`
- Test only if modified: `/home/zsts_119/patchnote-agent/internal/mcp/*_test.go`

**Checklist:**
- [ ] Avoid changing local MCP if remote can adapt cleanly.
- [ ] If adding `return_content`, make it backward compatible and optional.
- [ ] Keep existing `out` file export behavior working.
- [ ] Keep existing `save_draft` behavior working.
- [ ] Keep local MCP stdout JSON-RPC only.

**Acceptance:**
- [ ] `go test ./...` passes in `patchnote-agent`.
- [ ] `node packages/npm/test/install.test.js` passes if npm wrapper changed.
- [ ] Existing local MCP protocol smoke still reports 19 tools.

---

## Task 11: Wire Routes, Config, Limits, And Observability

**Files:**
- Modify: `/home/zsts_119/patchxNoteGoServer/cmd/api/main.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/config.yaml.example`
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/platform/config/*.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/platform/httpapi/router.go` only if route body limit overrides are needed.
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/apirequestdebug/*` only if route classification needs connector actor support.

**Checklist:**
- [ ] Add feature flag: `agent_mcp.enabled`.
- [ ] Add test/prod public base URL config.
- [ ] Add max request bytes.
- [ ] Add max response bytes per tool.
- [ ] Add default page size and max page size.
- [ ] Add per-tool timeout.
- [ ] Add webhook send timeout.
- [ ] Add rate-limit config placeholders if the existing limiter can be reused.
- [ ] Add safe metrics labels only: route, status, tool, platform client, coarse outcome.
- [ ] Add schema/contract version and contract hash to `initialize` or `tools/list` metadata.
- [ ] Add reverse-proxy trusted host/path-prefix config so generated OAuth URLs match the public URL, not the internal upstream URL.
- [ ] Add per-tool rate limits, with stricter defaults for model IO field tools and `patchxnote_send_webhook`.
- [ ] Add concurrency caps for remote MCP requests per connector session and per account.
- [ ] Add graceful shutdown handling so in-flight webhook sends and DB operations are bounded on deploy.

**Acceptance:**
- [ ] Disabled feature flag leaves `/mcp` unavailable.
- [ ] Enabled feature flag registers `/mcp` and discovery/auth routes.
- [ ] Logs and request debug metadata contain no secrets or raw user content.
- [ ] Generated metadata URLs are correct behind the test Caddy `/patchnote-test-api` prefix.
- [ ] Rate-limit responses include stable error code and optional retry-after, without raw internals.

---

## Task 11A: Retention, Cleanup, And Key Rotation

**Files:**
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/agentoauth/repository.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/agentoauth/service.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/agentwebhook/repository.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/jobs/*` only if cleanup is wired into the existing job runtime.
- Modify: `/home/zsts_119/patchxNoteGoServer/config.yaml.example`
- Test: `/home/zsts_119/patchxNoteGoServer/internal/agentoauth/repository_integration_test.go`
- Test: `/home/zsts_119/patchxNoteGoServer/internal/agentwebhook/repository_integration_test.go`

**Checklist:**
- [ ] Delete or expire unused authorization codes shortly after expiry.
- [ ] Expire connector access tokens without deleting connector session history needed for revocation/audit.
- [ ] Rotate connector refresh tokens and retain only safe predecessor hashes needed for reuse detection.
- [ ] Prune expired short-lived download/content handles.
- [ ] Prune webhook send replay/idempotency records after the configured replay window.
- [ ] Keep connector session audit metadata long enough for user-visible revocation history, but never keep raw tokens or raw model/content bodies.
- [ ] Document key rotation for connector token signing, OAuth client secrets, and webhook target encryption.
- [ ] Ensure `config.yaml.example` includes only secret file paths or non-secret schema, never real key material.

**Acceptance:**
- [ ] Cleanup tests prove expired codes/tokens/handles are removed or marked expired.
- [ ] Active connector sessions survive cleanup.
- [ ] Rotated keys can still read existing encrypted webhook targets during the overlap window.
- [ ] No cleanup log prints token hashes, raw URLs, secrets, user content, or model payload.

---

## Task 12: OpenAPI, Docs, And Client Website Copy

**Files:**
- Modify: `/home/zsts_119/patchxNoteGoServer/openapi/openapi.yaml`
- Modify: `/home/zsts_119/patchxNoteGoServer/openapi/CHANGELOG.md`
- Modify: `/home/zsts_119/patchxNoteGoServer/docs/integrations/apifox/agent/agent-model-io-read-flow.zh-CN.md`
- Modify: `/home/zsts_119/patchxNoteGoServer/docs/integrations/apifox/shared/integration-guide.zh-CN.md`
- Modify: `/home/zsts_119/patchnote-agent/docs/mcp-clients/clients.json`
- Modify: `/home/zsts_119/patchnote-agent/docs/mcp-clients/README.zh-CN.md`
- Modify: `/home/zsts_119/patchnote-agent/docs/mcp-clients/client-detail-copy.zh-CN.md`
- Modify: `/home/zsts_119/patchnote-agent/docs/plans/2026-08-27-remote-mcp-platform-gateway-design.md`

**Checklist:**
- [ ] Document test remote MCP URL.
- [ ] Document future production remote MCP URL.
- [ ] Document OAuth browser authorization flow.
- [ ] Document which clients use npm stdio and which use hosted MCP URL.
- [ ] Replace the old “remote V1 exposes fewer tools” design note with “remote V1 targets 19-tool parity”.
- [ ] Document remote file-output behavior.
- [ ] Document webhook target storage as account-scoped server state.
- [ ] Document revocation and session management.
- [ ] Document scope matrix and which tools require high-risk scopes.
- [ ] Document connector session list/revoke flow.
- [ ] Document idempotency expectations for remote write/send tools.
- [ ] Document `.well-known` metadata URLs for both test path-prefix and future production subdomain.
- [ ] Document that test-domain success is test acceptance, not production readiness.

**Acceptance:**
- [ ] No docs ask users to paste PatchXNote raw access/refresh tokens into third-party MCP configs.
- [ ] Docs clearly show Feishu Aily / Doubao Work Partner / Tencent ADP / enterprise WorkBuddy platform route.
- [ ] Public-facing docs do not include real tenant IDs, callback URLs with secrets, tokens, OTPs, raw phone numbers, webhook URLs, raw source text, or model payload.

---

## Task 13: Add Smoke Module And API Gate Coverage

**Files:**
- Modify: `/home/zsts_119/patchxNoteGoServer/tests/smoke/registry.yaml`
- Create: `/home/zsts_119/patchxNoteGoServer/tests/smoke/agent-remote-mcp/cases.yaml`
- Create: `/home/zsts_119/patchxNoteGoServer/tests/smoke/agent-remote-mcp/agent_remote_mcp_test.go`
- Modify: `/home/zsts_119/patchxNoteGoServer/internal/platform/openapigate/openapi_test.go`

**Checklist:**
- [ ] Add new smoke module `agent-remote-mcp`.
- [ ] Add cases for discovery metadata.
- [ ] Add cases for OAuth authorize/token/revoke.
- [ ] Add cases for `/mcp initialize`.
- [ ] Add cases for `/mcp tools/list`.
- [ ] Add cases for one read tool.
- [ ] Add cases for one model IO field tool.
- [ ] Add cases for webhook target CRUD.
- [ ] Add cases for webhook send using a test receiver.
- [ ] Add negative cases: no auth, bad token, revoked token, wrong scope, wrong platform, over-limit, unknown tool.
- [ ] Add negative cases: redirect URI mismatch, PKCE mismatch, state mismatch, reused authorization code, reused refresh token.
- [ ] Add negative cases: cursor from another account/platform/filter, JSON-RPC notification, batch request rejection, header/body method mismatch.
- [ ] Add negative cases: duplicate webhook send retry, idempotency-key mismatch, private-IP webhook URL, DNS rebinding webhook URL.
- [ ] Add path-prefix cases for `/patchnote-test-api/mcp` and well-known metadata.
- [ ] Update operation count expectations if OpenAPI gate enforces exact counts.

**Acceptance:**
- [ ] `go test ./internal/remotemcp ./internal/agentoauth ./internal/agentwebhook -count=1` passes.
- [ ] `go test ./tests/smoke/agent-remote-mcp -count=1` passes with isolated test config.
- [ ] At least one test compares remote `tools/list` with the local MCP fixture contract.

---

## Task 14: Protocol Smoke Against A Real Local Process

**Files:**
- Create: `/home/zsts_119/patchxNoteGoServer/scripts/smoke/remote-mcp-smoke.mjs` or a Go equivalent.
- Create: `/home/zsts_119/patchxNoteGoServer/docs/engineering/evidence/YYYY-MM-DD-agent-remote-mcp-local-validation.md`

**Checklist:**
- [ ] Start one GoServer process with remote MCP enabled.
- [ ] Fetch protected resource metadata.
- [ ] Complete OAuth flow with non-production review OTP where allowed.
- [ ] Call `/mcp initialize`.
- [ ] Call `/mcp tools/list` and assert 19 tools.
- [ ] Record remote tool contract hash and compare it with the local MCP fixture hash.
- [ ] Call `patchxnote_get_current_user`.
- [ ] Call `patchxnote_list_memories` for `mobile`.
- [ ] Call `patchxnote_list_model_io_traces` for `mobile`.
- [ ] Call one model IO field tool only if safe sample exists.
- [ ] Configure a test webhook alias.
- [ ] Render a webhook message.
- [ ] Send to a local test webhook receiver.
- [ ] Repeat the same send request with the same idempotency key and verify no duplicate delivery.
- [ ] Revoke connector session.
- [ ] Confirm post-revocation `/mcp` call fails.

**Acceptance:**
- [ ] Evidence records status, request IDs, counts, versions, and timestamps only.
- [ ] Evidence does not include raw phone, OTP, token, raw transcript, provider payload, full model output, webhook URL, or secret.
- [ ] Smoke script exits non-zero on tool count drift, auth failure, duplicate delivery, or response-size cap violation.

---

## Task 15: Deploy To Test Domain

**Files:**
- Modify only if required: `/home/zsts_119/patchxNoteGoServer/deploy/*`
- Modify only if required: `/home/zsts_119/patchxNoteGoServer/docs/engineering/deployment-ops.md`
- Create: `/home/zsts_119/patchxNoteGoServer/docs/engineering/evidence/YYYY-MM-DD-agent-remote-mcp-test-release.md`

**Checklist:**
- [ ] Build GoServer binary with version metadata.
- [ ] Apply migrations to the test environment.
- [ ] Deploy with `agent_mcp.enabled=true`.
- [ ] Verify `https://ws-lab.patch-x.cn/patchnote-test-api/mcp/health`.
- [ ] Verify well-known metadata under the test path.
- [ ] Verify the `WWW-Authenticate` `resource_metadata` URL is publicly reachable and not an internal upstream URL.
- [ ] Verify OAuth authorize/token/revoke URLs generated behind Caddy use public HTTPS and the `/patchnote-test-api` prefix where required.
- [ ] Verify `/mcp initialize`.
- [ ] Verify `/mcp tools/list` returns 19.
- [ ] Verify one safe real-account read call.
- [ ] Verify one bounded model IO field call with synthetic or user-approved safe sample only.
- [ ] Verify one webhook target CRUD/send path against a controlled test receiver.
- [ ] Verify revocation.
- [ ] Verify logs contain no secret/content leaks.

**Acceptance:**
- [ ] Test deployment evidence has commit SHA, config fingerprint, migration version, route checks, and smoke result.
- [ ] Failed platform calls produce stable error envelopes and usable request IDs.
- [ ] Feature flag can be disabled and re-enabled without leaving broken routes or stale metadata.

---

## Task 16: Platform Acceptance

**Files:**
- Create: `/home/zsts_119/patchnote-agent/docs/evidence/YYYY-MM-DD-platform-remote-mcp-acceptance.zh-CN.md`
- Modify: `/home/zsts_119/patchnote-agent/docs/mcp-clients/clients.json`
- Modify: `/home/zsts_119/patchnote-agent/docs/mcp-clients/README.zh-CN.md`

**Checklist:**
- [ ] For every platform, record whether it supports local command mode, hosted HTTP MCP mode, OAuth callback mode, static header mode, or only manual/internal connectors.
- [ ] For every OAuth platform, register exact redirect URI and client id before testing.
- [ ] For every platform, record whether it caches `tools/list`, how to refresh the cache, and whether stale schemas affect rollout.
- [ ] Feishu Aily: add custom MCP using hosted URL, complete OAuth, run initialize/tools/list/tools/call.
- [ ] Doubao Work Partner: test local/cloud-computer command mode if available.
- [ ] Doubao Work Partner: test hosted MCP URL mode if available.
- [ ] Tencent ADP: add custom connector URL, configure OAuth, complete authorization, pull tools, call one safe tool.
- [ ] Enterprise WorkBuddy: test tenant connector path if account/docs are available.
- [ ] CodeBuddy CLI: verify local npm path; optionally verify HTTP path if supported.
- [ ] If a platform only supports static header credentials, mark it as test-only unless it can use revocable connector tokens without exposing PatchXNote Agent tokens.

**Acceptance:**
- [ ] Each accepted platform has screenshots or text evidence with no secrets.
- [ ] Each unsupported mode is documented as blocked with the exact reason.
- [ ] Website/client registry status is updated from `planned` to `validated` only after real platform proof.
- [ ] Evidence separates local command acceptance from hosted remote MCP acceptance.

---

## Task 17: Final Regression Gate

**GoServer commands:**

```sh
cd /home/zsts_119/patchxNoteGoServer
go test ./internal/remotemcp ./internal/agentoauth ./internal/agentwebhook -count=1
go test ./internal/agentaccess ./internal/platform/httpapi ./internal/platform/openapigate -count=1
go test ./tests/smoke/agent-access -count=1
go test ./tests/smoke/agent-model-io-read -count=1
go test ./tests/smoke/agent-remote-mcp -count=1
```

**Agent commands if `patchnote-agent` changed:**

```sh
cd /home/zsts_119/patchnote-agent
go test ./...
node packages/npm/test/install.test.js
node docs/mcp-clients/validate-clients.mjs
```

**External smoke:**

```sh
curl -fsS https://ws-lab.patch-x.cn/patchnote-test-api/mcp/health
```

Then use the MCP smoke script to call:

```text
initialize
tools/list
tools/call patchxnote_get_current_user
tools/call patchxnote_list_memories
```

**Acceptance:**
- [ ] Local npm MCP still works.
- [ ] GoServer remote MCP works on test domain.
- [ ] `tools/list` parity is 19 tools.
- [ ] Remote/local tool contract hash comparison passes.
- [ ] OAuth connector authorization and revocation work.
- [ ] Scope matrix enforcement works.
- [ ] Remote render/export content path works.
- [ ] Remote webhook storage/send works.
- [ ] Remote webhook send idempotency works.
- [ ] OAuth/session/token cleanup works.
- [ ] Sensitive scan passes.
- [ ] OpenAPI and docs are updated.

---

## Non-Goals For This Slice

- [ ] Do not add raw audio download.
- [ ] Do not expose full App/PC OpenAPI as MCP.
- [ ] Do not merge mobile and desktop content.
- [ ] Do not make remote platform clients depend on the user's local npm cache.
- [ ] Do not require users to paste PatchXNote access/refresh tokens into platform configs.
- [ ] Do not add a separate microservice unless GoServer integration is proven insufficient.
- [ ] Do not claim production readiness from local or test-domain smoke alone.

## Main Risks And Handling

- [ ] **Platform local execution ambiguity:** Some products can run local/cloud-computer commands, some cannot. Support both npm command and hosted URL in docs.
- [ ] **OAuth variance:** Platforms differ in OAuth client-secret, callback, PKCE, and discovery support. Start with standards plus pre-registered platform clients; document per-platform quirks during acceptance.
- [ ] **Path-prefix discovery:** The test domain lives under `/patchnote-test-api`; wrong `.well-known` or callback URL generation can make OAuth discovery fail even when `/mcp` itself is reachable.
- [ ] **Connector principal mapping:** Existing Agent REST auth validates Agent sessions; connector tokens need a clear adapter or internal session strategy before tool handlers call Agent services.
- [ ] **Tool count vs model selection quality:** 19 tools is accepted for parity, but descriptions must stay concise and names explicit.
- [ ] **Schema drift:** Local MCP tools may change before remote MCP ships; contract fixture/hash comparison must fail loudly.
- [ ] **Remote file writes:** No user local filesystem exists. Return bounded content or download handles.
- [ ] **Webhook SSRF/security:** Remote server sends outbound webhooks. Enforce HTTPS, deny private networks, cap body/timeout, and log only aliases/status.
- [ ] **Webhook duplicate delivery:** Platform retries can duplicate external sends; idempotency/replay is required for send paths.
- [ ] **Large model IO content:** Cap remote MCP result size and never store raw fields in evidence/logs.
- [ ] **Token leakage:** Platform stores connector token only; GoServer never exposes Agent refresh/access tokens to platform config.
- [ ] **Retention drift:** Expired codes/tokens/download handles and webhook replay records need cleanup so test data does not become long-lived production data.
- [ ] **Spec drift:** Validate current target-platform MCP transport requirements on implementation day.

## Done Definition

- [ ] GoServer exposes hosted remote MCP on the test domain.
- [ ] Official MCP/OAuth discovery works.
- [ ] Remote `tools/list` returns all 19 local MCP tools.
- [ ] Remote/local MCP contract parity test passes.
- [ ] Every tool has at least one unit or smoke coverage path.
- [ ] At least one platform connector completes initialize/tools/list/tools/call.
- [ ] OAuth scopes, connector listing, revocation, token rotation, and cleanup are verified.
- [ ] Remote webhook SSRF protection and idempotency are verified.
- [ ] Local npm MCP behavior is unchanged.
- [ ] Docs and client registry describe both local command and hosted remote MCP paths.
- [ ] Evidence contains no secrets or raw user/model content.

## References To Recheck On Implementation Day

- Model Context Protocol 2026-07-28 specification: `https://modelcontextprotocol.io/specification/2026-07-28`
- MCP authorization server discovery: `https://modelcontextprotocol.io/specification/2026-07-28/basic/authorization/authorization-server-discovery`
- Tencent ADP custom MCP connector: `https://www.tencentcloud.com/zh/document/product/1254/81634`
- CodeBuddy MCP transport and permissions: `https://www.codebuddy.cn/docs/cli/mcp`
- Existing local Agent MCP design: `docs/plans/2026-08-06-agent-v1-mvp.md`
- Existing local/client setup plan: `docs/plans/2026-08-27-context7-style-setup-and-client-platform-mcp-checklist.md`
