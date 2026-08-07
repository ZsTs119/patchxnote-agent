# Agent Auth Refresh Implementation Plan

> **For implementation:** Execute this plan sequentially in the primary agent. Do not use sub-agents or parallel task execution.

**Goal:** Add a 30-day Agent/CLI/MCP refresh-token flow so local PatchXNote Agent users do not need to re-login when the short-lived access token expires.

**Architecture:** Keep App/PC authentication unchanged. Add an Agent-only refresh contract under `/v1/agent/auth/refresh`, store Agent refresh state in `agent_session` plus Agent-specific replay storage, and teach the local CLI/MCP credential provider to refresh automatically before making protected Agent API calls.

**Tech Stack:** Go, Chi HTTP routes, PostgreSQL migrations, existing PatchXNote auth key material, OpenAPI, Cobra CLI, OS-native keychain, stdio MCP.

**Execution Update (2026-08-07 Asia/Shanghai):** Server support is implemented and deployed to the test environment at revision `788973838facbf31588b30dde38fd279bc35cad0`; Agent CLI/MCP auto-refresh is implemented and validated against the test environment with a real logged-in profile. Remaining follow-up is production release/npm publication and a `patchxnote_get_memory` success-path smoke with an account that has at least one readable structured result.

---

## Product Decisions

- [ ] This change targets only Agent/CLI/MCP authentication.
- [ ] Do not modify App/PC `/v1/auth/refresh` request/response semantics.
- [ ] Do not modify App/PC `client_installation` slot behavior.
- [ ] Do not add `agent` to `identity.Platform`.
- [ ] Do not let Agent refresh tokens validate through the App/PC client refresh endpoint.
- [ ] Keep access token TTL short, using the existing server default such as `15m`.
- [ ] Add a 30-day refresh token for Agent sessions, using the existing `auth.refresh_token_ttl` default of `720h` unless server config says otherwise.
- [ ] Store refresh tokens only in OS-native secure storage on the CLI side.
- [ ] Store only refresh-token hashes on the server side.
- [ ] Rotate the Agent refresh token on every successful refresh.
- [ ] Old Agent sessions created before this feature may not have refresh metadata; they should require one explicit re-login.
- [ ] Refresh failure should surface the server error clearly and ask the user to run `patchxnote login` again.
- [ ] `patchxnote logout` should revoke the Agent session when possible and always clear local credentials afterward.

## Non-Goals

- [ ] No App/PC protocol migration.
- [ ] No App/PC client release requirement for this change.
- [ ] No changes to hardware binding, installation replacement, quota claim, model execution, content write, payment, or Admin flows.
- [ ] No server-side background worker for Agent refresh.
- [ ] No long-lived 30-day access token.
- [ ] No token values in logs, command output, examples, fixtures, screenshots, or docs.

## Expected User Flow

```text
patchxnote login
  -> server returns access_token + refresh_token
  -> CLI stores both in secure storage

patchxnote auth status
  -> CLI checks local access expiry
  -> if access is fresh, call /v1/agent/me
  -> if access is stale, call /v1/agent/auth/refresh first
  -> save rotated credentials
  -> call /v1/agent/me

patchxnote mcp serve
  -> every protected MCP tool asks the credential provider for a valid access token
  -> provider refreshes automatically when needed
```

## Key Edge Cases

- [ ] Existing installed users with no stored refresh token: command fails with `auth_required` and a re-login hint.
- [ ] Access token expired but refresh token valid: CLI refreshes and retries the intended command.
- [ ] Refresh token expired: server returns `401`, CLI clears or marks stale credentials, user re-runs `patchxnote login`.
- [ ] Refresh token reused after rotation: server rejects the request and revokes the Agent session.
- [ ] Multiple local CLI/MCP processes refresh at the same time: CLI uses a local refresh lock and re-reads credentials after acquiring it.
- [ ] Network timeout after refresh succeeds but before CLI saves the response: same in-process retry should reuse the same `Idempotency-Key`; if recovery is impossible, user can re-login.
- [ ] Server returns a 5xx or retryable error: CLI surfaces the original error and does not discard refresh credentials immediately.
- [ ] Server clock/client clock skew: CLI refreshes early enough before expiry, using a conservative refresh window.
- [ ] Logout with expired access but valid refresh: CLI refreshes once, then calls Agent logout.
- [ ] Logout when remote revoke cannot complete: CLI clears local credentials and warns that remote revocation may not have completed.
- [ ] MCP stdio logs: no refresh diagnostics go to stdout; stdout remains JSON-RPC only.
- [ ] `patchxnote_search_memories` must require a valid Agent credential before searching local cached projections.

## Phase 0: Baseline And Branch Safety

Files:

- Inspect Agent repo: `README.md`
- Inspect Agent repo: `docs/engineering-rules.md`
- Inspect Agent repo: `docs/plans/2026-08-06-agent-v1-mvp.md`
- Inspect GoServer repo: `../patchxNoteGoServer/docs/engineering/agent-access-v1.md`
- Inspect GoServer repo: `../patchxNoteGoServer/docs/integrations/apifox/integration-guide.zh-CN.md`

Checklist:

- [ ] Confirm Agent repo dirty state and keep unrelated webhook plan files out of auth commits.
- [ ] Confirm GoServer branch is the intended implementation branch.
- [ ] Record current server OpenAPI version and commit in the implementation notes.
- [ ] Confirm deployed environment has `auth.refresh_token_ttl=720h` or equivalent 30-day value.
- [ ] Confirm Agent access TTL remains within the existing short-lived access-token range.

Validation:

- [ ] `git status --short --branch` is recorded for both repositories.
- [ ] No unrelated docs, release artifacts, or local temp files are staged.

## Phase 1: Server Contract And Design Docs

Files:

- Modify GoServer: `../patchxNoteGoServer/docs/engineering/agent-access-v1.md`
- Modify GoServer: `../patchxNoteGoServer/openapi/openapi.yaml`
- Modify GoServer: `../patchxNoteGoServer/tests/smoke/registry.yaml`
- Modify GoServer: `../patchxNoteGoServer/tests/smoke/agent-access/cases.yaml`
- Regenerate GoServer: `../patchxNoteGoServer/docs/integrations/apifox/patchnote-openapi.zh-CN.json`

Checklist:

- [ ] Mark Agent refresh as an accepted V1 design decision.
- [ ] Add `POST /v1/agent/auth/refresh` with `operationId: refreshAgentSession`.
- [ ] Define `AgentRefreshRequest` with only `refresh_token`.
- [ ] Extend `AgentSessionResponse` with `refresh_token` and `refresh_expires_in_seconds`.
- [ ] Keep `AgentSessionResponse.access_expires_in_seconds` max at the short access-token bound.
- [ ] Document that Agent refresh does not require `installation_proof`.
- [ ] Document that Agent refresh does not touch `client_installation`.
- [ ] Add smoke cases for refresh success, rotation, replay, and reuse rejection.
- [ ] Add registry operation contract metadata for the refresh write operation.

Validation:

- [ ] `make openapi-check`
- [ ] `make smoke-coverage`
- [ ] `make apifox-bundle`

## Phase 2: Server Migration

Files:

- Create GoServer: `../patchxNoteGoServer/migrations/000037_agent_session_refresh.up.sql`
- Create GoServer: `../patchxNoteGoServer/migrations/000037_agent_session_refresh.down.sql`
- Modify GoServer: `../patchxNoteGoServer/migrations/checksums.sha256`

Checklist:

- [ ] Add nullable refresh columns to `agent_session`: `refresh_hash`, `refresh_key_id`, `refresh_expires_at`, and `refresh_rotated_at`.
- [ ] Add constraints so refresh fields are either all present or all absent.
- [ ] Add a partial unique index on active refresh hash/key fields.
- [ ] Preserve existing Agent sessions without refresh fields for backward-compatible re-login handling.
- [ ] Add Agent-specific idempotency/replay storage if OTP verify and refresh responses now contain refresh tokens.
- [ ] Do not alter `client_refresh_session`.
- [ ] Do not alter `client_installation`.

Validation:

- [ ] `make migration-check`
- [ ] Migration up/down works in an isolated smoke schema.

## Phase 3: Server Repository And Service

Files:

- Modify GoServer: `../patchxNoteGoServer/internal/agentaccess/domain.go`
- Modify GoServer: `../patchxNoteGoServer/internal/agentaccess/secrets.go`
- Modify GoServer: `../patchxNoteGoServer/internal/agentaccess/runtime.go`
- Modify GoServer: `../patchxNoteGoServer/internal/agentaccess/repository.go`
- Modify GoServer: `../patchxNoteGoServer/internal/agentaccess/service.go`
- Modify GoServer tests: `../patchxNoteGoServer/internal/agentaccess/service_test.go`
- Modify GoServer tests: `../patchxNoteGoServer/internal/agentaccess/repository_integration_test.go`

Checklist:

- [ ] Add Agent refresh token fields to `SessionResponse`.
- [ ] Add `RefreshInput` and refresh repository parameter types.
- [ ] Load refresh-token hash key material into `agentaccess.Secrets`.
- [ ] Use an Agent-only hash purpose such as `agent-refresh-token`.
- [ ] Add `RefreshTTL` to Agent service configuration.
- [ ] Generate a refresh token during Agent OTP verification.
- [ ] Store only the refresh hash/key/expires metadata in `agent_session`.
- [ ] Return the plaintext refresh token only in the successful JSON response.
- [ ] Add service method `Refresh(ctx, RefreshInput)`.
- [ ] Verify refresh token candidates against current and previous refresh hash keys.
- [ ] Rotate refresh metadata on successful refresh.
- [ ] Return a fresh access token and fresh refresh token after rotation.
- [ ] Reject expired, revoked, malformed, or missing refresh tokens with stable service errors.
- [ ] Detect old-token reuse after rotation and revoke the Agent session.
- [ ] Preserve current read-scope checks for all Agent read projections.
- [ ] Keep App/PC `clientaccess` code untouched unless a shared helper is extracted without behavior changes.

Validation:

- [ ] Unit test: Agent OTP verify returns refresh fields.
- [ ] Unit test: Agent refresh returns a new access token and new refresh token.
- [ ] Unit test: malformed refresh input maps to `invalid_request`.
- [ ] Integration test: Agent session stores refresh hash, not plaintext.
- [ ] Integration test: refresh rotates hash and extends refresh expiry.
- [ ] Integration test: reused old refresh token is rejected and session is revoked.
- [ ] Integration test: App/PC refresh endpoint does not accept Agent refresh tokens.
- [ ] Integration test: Agent refresh endpoint does not accept App/PC refresh tokens.

## Phase 4: Server HTTP Route

Files:

- Modify GoServer: `../patchxNoteGoServer/internal/agentaccess/http.go`
- Modify GoServer smoke: `../patchxNoteGoServer/tests/smoke/agent-access/agent_access_test.go`

Checklist:

- [ ] Add `Refresh(context.Context, RefreshInput)` to `AgentAPI`.
- [ ] Register `POST /v1/agent/auth/refresh`.
- [ ] Require `Content-Type: application/json`.
- [ ] Require `Idempotency-Key`.
- [ ] Reject unknown JSON fields.
- [ ] Do not require `Authorization` for refresh.
- [ ] Never write refresh token values to logs or smoke artifacts.
- [ ] Add smoke refresh call after OTP verify and before read tools.
- [ ] Assert old access token and new access token both preserve Agent audience isolation.
- [ ] Assert refresh success does not create App/PC installation rows.
- [ ] Assert refresh success does not change mobile/desktop installation state.

Validation:

- [ ] `go test ./internal/agentaccess`
- [ ] `go test -tags integration ./internal/agentaccess` when smoke config is available.
- [ ] `make smoke-module MODULE=agent-access` or repository-equivalent module smoke command.

## Phase 5: CLI API Client

Files:

- Modify Agent repo: `internal/api/types.go`
- Modify Agent repo: `internal/api/client.go`
- Modify Agent tests near existing API client tests.

Checklist:

- [ ] Add `RefreshToken` and `RefreshExpiresInSeconds` to `AgentSessionResponse`.
- [ ] Add `AgentRefreshRequest`.
- [ ] Add `RefreshAgentSession(ctx, refreshToken, idempotencyKey)`.
- [ ] Keep access token and refresh token out of formatted error output.
- [ ] Preserve current login behavior for servers that return only access token by showing a clear "server does not support refresh; please re-login after expiry" state.

Validation:

- [ ] Unit test: refresh request sends only the refresh token body and idempotency header.
- [ ] Unit test: refresh response parses rotated credentials.
- [ ] Unit test: API error response is surfaced without token leakage.

## Phase 6: CLI Credential Refresh Manager

Files:

- Modify Agent repo: `internal/auth/manager.go`
- Add or modify Agent repo: `internal/auth/session.go`
- Modify Agent repo: `internal/cli/runtime.go`
- Modify Agent repo: `internal/cli/auth.go`
- Modify Agent tests near existing auth/CLI tests.

Checklist:

- [ ] Add a credential provider method that guarantees a usable access token or returns an auth error.
- [ ] Refresh when access token is expired.
- [ ] Refresh when access token will expire inside the configured early-refresh window.
- [ ] Use a local refresh mutex for in-process MCP calls.
- [ ] Use a local file lock or OS lock for separate CLI/MCP processes.
- [ ] Re-read credentials after acquiring the refresh lock.
- [ ] Save rotated access/refresh credentials atomically through the existing secure store boundary.
- [ ] Keep non-secret expiry metadata in the current metadata store.
- [ ] If refresh token is absent, return a re-login-required error.
- [ ] If refresh returns `401`, clear stale credentials or mark them unusable and ask the user to login.
- [ ] If refresh returns retryable 5xx/network error, keep existing refresh token and surface the failure.
- [ ] Update `patchxnote auth status` to refresh before calling `/v1/agent/me`.
- [ ] Update `patchxnote logout` to refresh once if needed, then revoke remote session.
- [ ] Keep human diagnostics on stderr and machine-readable output on stdout.

Validation:

- [ ] Unit test: expired access plus valid refresh performs refresh before status.
- [ ] Unit test: fresh access skips refresh.
- [ ] Unit test: missing refresh token requires re-login.
- [ ] Unit test: refresh 401 clears or invalidates credentials.
- [ ] Unit test: concurrent refresh attempts use one rotated credential result instead of double-refreshing.

## Phase 7: MCP Credential Integration

Files:

- Modify Agent repo: `internal/mcp/server.go`
- Modify Agent repo: `internal/cli/mcp.go`
- Modify Agent tests: `internal/mcp/server_test.go`
- Modify Agent tests: `internal/cli/mcp_test.go`

Checklist:

- [ ] Inject the refresh-capable credential provider into MCP server startup.
- [ ] Ensure every server-backed MCP tool requests a fresh access token through that provider.
- [ ] Require authentication for local-cache search before returning local memory results.
- [ ] Do not print refresh logs to stdout in stdio mode.
- [ ] Preserve MCP `initialize.serverInfo.version` behavior fixed in `0.2.2`.
- [ ] Keep the public MCP tool count at seven unless a separate feature plan changes it.

Validation:

- [ ] MCP initialize still returns the release version.
- [ ] MCP tools/list still returns exactly seven tools.
- [ ] Expired access credential refreshes before `patchxnote_get_current_user`.
- [ ] Unauthenticated `patchxnote_search_memories` returns `auth_required` instead of unauthenticated local results.

## Phase 8: Release And Install Flow

Files:

- Modify Agent repo: `README.md`
- Modify Agent repo: `README.zh-CN.md`
- Modify Agent repo: `packages/npm/package.json`
- Modify Agent repo: `.github/workflows/macos-install-smoke.yml`
- Modify Agent repo as needed: release runbook docs.

Checklist:

- [ ] Bump Agent CLI version after implementation.
- [ ] Document that existing users must login once after server support is deployed if their stored credentials lack refresh token.
- [ ] Document that normal users do not need to reinstall only because refresh tokens rotate.
- [ ] Reinstall is needed only to pick up a newly released CLI binary.
- [ ] Keep MCP JSON config free of access and refresh tokens.

Validation:

- [ ] `go test ./...`
- [ ] `scripts/e2e/mvp-smoke.sh`
- [ ] Windows npm wrapper test from a local temp path, not UNC.
- [ ] `npm pack --dry-run`
- [ ] Fresh install prints MCP config without secrets.
- [ ] Real login stores refresh metadata without printing token values.
- [ ] After forced/stubbed access expiry, an MCP tool succeeds through auto-refresh.

## Deployment Order

- [ ] Merge and deploy GoServer support first.
- [ ] Confirm live `verifyAgentOTP` returns refresh fields.
- [ ] Release Agent CLI with auto-refresh support.
- [ ] Ask existing Agent users to run `patchxnote login` once after both sides are live.
- [ ] Verify `patchxnote auth status` remains valid after access-token expiry without another OTP.

## Rollback Plan

- [ ] If server refresh has issues, disable or remove `/v1/agent/auth/refresh` while leaving OTP login and read-only Agent access intact.
- [ ] If CLI auto-refresh has issues, release a CLI patch that falls back to current re-login behavior.
- [ ] Do not roll back App/PC auth modules for Agent refresh issues.
- [ ] If a refresh-token reuse bug revokes sessions too aggressively, affected Agent users can recover by re-running `patchxnote login`.

## Definition Of Done

- [ ] Agent login returns short-lived access token plus 30-day refresh token.
- [ ] Agent refresh rotates refresh token and preserves read-only Agent scopes.
- [ ] CLI/MCP refresh automatically before protected Agent calls.
- [ ] Old App/PC login, installation, replacement, refresh, and logout tests remain green.
- [ ] No access token, refresh token, OTP, raw phone, full MAC, SK, raw audio, transcript, or provider payload appears in logs, docs, smoke artifacts, or command output.
- [ ] Real installed CLI can login once and keep working after access-token expiry without another SMS OTP.
