# MCP Browser OAuth Login Productization Implementation Plan

**Goal:** Productize the verified browser OAuth flow into `patchnote-agent` so users can run `patchxnote mcp login` or `npx -y patchxnote-agent@latest mcp login`, finish PatchXNote login in the browser, store MCP connector credentials locally, and then use PatchXNote MCP from VS Code, Cursor, Codex, Claude Code, Claude Desktop, Windsurf, WorkBuddy, and other local stdio MCP clients.

**Architecture:** GoServer remains the web and OAuth authority: it hosts `/v1/agent/oauth/authorize`, performs phone OTP login in the browser page, issues authorization codes, exchanges them for MCP connector tokens, refreshes tokens, revokes tokens, and serves the hosted `/mcp` JSON-RPC endpoint. `patchnote-agent` becomes the local native app part of the OAuth loop: it generates PKCE/state, listens on a temporary `127.0.0.1:<port>/callback`, opens the GoServer authorization page, exchanges the callback code, stores connector credentials in OS-native secure storage, and lets local stdio MCP clients use those credentials without placing secrets in their config files.

**Tech Stack:** Go, Cobra, Viper, OS-native keychain via `internal/keychain`, stdio MCP JSON-RPC, Go `net/http` loopback callback server, OAuth authorization code + PKCE, npm launcher wrapper, PatchXNote GoServer OAuth and remote MCP endpoints.

**Execution Rule:** Work sequentially in the primary agent. Do not use sub-agents or parallel task execution. Keep the implementation small, testable, and reversible. This plan intentionally replaces the older browser setup-session direction with the now-deployed GoServer OAuth authorize/token flow.

---

## Current Baseline

- [x] GoServer test deployment has browser OAuth authorize HTML.
- [x] GoServer accepts dynamic loopback redirect URIs such as `http://127.0.0.1:<port>/callback`.
- [x] GoServer token response uses:
  - `access_token`
  - `token_type`
  - `expires_in`
  - `refresh_token`
  - `refresh_token_expires_in`
  - `scope`
  - `connector_session_id`
  - optional `patchxnote_schema_notice`
- [x] A temporary local Node callback script proved the real browser flow:
  - browser login succeeded
  - OAuth code callback succeeded
  - token exchange succeeded
  - remote MCP `patchxnote_get_current_user` call succeeded
- [x] `patchnote-agent` already has:
  - Cobra command tree under `internal/cli`
  - secure-storage boundary under `internal/keychain`
  - server client boundary under `internal/api`
  - local stdio MCP server under `internal/mcp`
  - setup client registry under `internal/setup`
  - npm wrapper under `packages/npm/bin/patchxnote-agent.js`
- [x] Current npm wrapper already delegates `login`, `setup`, `mcp config`, and `mcp serve`.
- [x] Current npm wrapper accepts `mcp login/status/logout`.
- [x] Current `patchxnote mcp` command has browser OAuth `login/status/logout` subcommands.
- [x] Current `mcp serve` uses the remote `/mcp` proxy when MCP OAuth credentials exist and keeps the local Agent fallback when they do not.
- [x] The older browser setup-session path is no longer the preferred product path after GoServer OAuth shipped.

## Implementation And Acceptance Update

**Updated:** 2026-08-27

Implemented in the local `0.2.8` candidate:

- [x] OAuth metadata discovery, token exchange, refresh, and revoke API client.
- [x] PKCE generation, loopback callback server, browser opener, credential validation, and MCP OAuth secure store.
- [x] `patchxnote mcp login/status/logout`.
- [x] Remote `/mcp` HTTP client and stdio proxy with local fallback when no MCP OAuth credential exists.
- [x] `setup --client <id>` now uses the MCP OAuth browser login path instead of terminal OTP by default.
- [x] npm wrapper delegates `mcp login/status/logout/serve` to the pinned native binary.
- [x] Public README, npm README, client registry, and release runbook now describe the browser OAuth path.

Passed local automated validation:

```text
/home/zsts_119/.local/go1.24.11/bin/go test ./... -count=1
node packages/npm/test/install.test.js
node docs/mcp-clients/validate-clients.mjs
scripts/e2e/mvp-smoke.sh
git diff --check
```

Passed real Windows local acceptance:

- [x] Windows-native `mcp login` opened the browser automatically without `--no-browser`.
- [x] User completed phone/OTP only in the GoServer browser page.
- [x] Local callback completed and saved MCP OAuth credentials for isolated profile `mcp-oauth-windows-acceptance`.
- [x] `mcp status --verify` returned `authenticated=true` and `verified=true`.
- [x] stdio MCP `initialize`, `tools/list`, and `patchxnote_get_current_user` passed through the remote `/mcp` proxy.
- [x] Authorized mobile summary records and model-IO detail tools were readable; evidence recorded only counts, fields, and bounded sizes.
- [x] npm wrapper `--from-local` candidate install and wrapper-launched `mcp serve` passed.

Still not accepted:

- [ ] Published npm registry package acceptance, because `0.2.8` has not been published yet.
- [ ] Real VS Code/Cursor/Codex/Claude/Windsurf/Trae/Qoder/WorkBuddy client UI acceptance.
- [ ] Feishu Aily / Doubao Work Partner / Tencent Agent Platform / enterprise WorkBuddy hosted platform-console acceptance.
- [ ] WSL automatic browser-open acceptance. WSL can use `--no-browser`, but it does not replace Windows desktop editor acceptance because the credential stores are separate.

Evidence:

- `docs/evidence/2026-08-27-mcp-oauth-local-acceptance.zh-CN.md`

## Confirmed Decisions

- [x] The website and authorization login page are GoServer responsibilities.
- [x] `patchnote-agent` does not host the public website.
- [x] `patchnote-agent` owns the local callback listener, browser launch, token exchange, secure storage, MCP config setup, and stdio MCP bridge.
- [x] `patchxnote login` remains the legacy terminal phone OTP Agent login path.
- [x] New browser OAuth login is exposed as `patchxnote mcp login`.
- [x] npm wrapper must support `npx -y patchxnote-agent@latest mcp login`.
- [x] MCP OAuth connector tokens must be stored separately from existing Agent OTP credentials.
- [x] MCP client config remains secret-free.
- [x] `mcp serve` must not automatically open a browser on startup in V1, because MCP hosts often have startup timeouts.
- [x] `setup` may call or reuse the same browser OAuth login flow before writing client config.
- [x] All token, code, phone, OTP, webhook URL, provider payload, full transcript, SK, and full MAC material must stay out of stdout, stderr, docs, tests, and evidence.

## Audit Additions From Plan Review

- [ ] Add OAuth metadata discovery before login instead of blindly hard-coding every endpoint. The default can still derive conventional paths from `server.base_url`, but implementation must verify GoServer metadata when reachable.
- [ ] Treat `server_base_url` as part of the MCP OAuth credential identity. A token issued for test must not be silently reused against production, another tenant, or a local dev server.
- [ ] Request and store the exact OAuth `scope` returned by GoServer. Do not reuse old setup-session placeholder scopes such as `agent:memories.read`.
- [ ] Keep `state`, PKCE verifier, authorization code, and callback URL in process memory only. They are transient non-bearer materials and must not be written to keychain, config, logs, test evidence, or docs.
- [ ] Support `application/json` and be explicit about `text/event-stream` handling for remote MCP HTTP responses. If GoServer returns SSE/Streamable HTTP framing later, the proxy must either parse it correctly or fail with a clear compatibility error.
- [ ] Add a dedicated MCP OAuth refresh lock, separate from the legacy Agent OTP refresh lock, so multiple editor MCP processes do not rotate the same refresh token concurrently.
- [ ] Add an escape hatch for rollout and support, for example an internal `PATCHXNOTE_MCP_MODE=local|remote|auto`, but keep public docs on the simple default path.
- [ ] Decide and document command exit codes: `mcp status` with no credential should be script-friendly, `mcp logout --local-only` with no credential should be idempotent, and real login failures should be non-zero.
- [ ] Treat Windows, WSL, VS Code Remote, Dev Containers, macOS, Linux desktop, and Linux headless as separate acceptance states because browser callback reachability and keychain storage differ.
- [ ] Add explicit cleanup behavior for abandoned browser login attempts: close callback listener, clear in-memory verifier/state, and avoid storing partial credentials.
- [ ] Keep the older setup-session code path either removed or clearly marked historical. Do not leave two browser-login product paths active without a feature flag and tests.
- [ ] Add release-version and artifact acceptance tasks; npm `latest` cannot prove the feature until the package and native binary are both published at the same version.
- [ ] Separate local-stdio client acceptance from hosted platform acceptance. VS Code, Cursor, Codex, Claude Code, Claude Desktop, and Windsurf can be validated through local config or command-based setup. Feishu Aily, Doubao WorkBuddy, Tencent Agent platforms, and enterprise WorkBuddy-style platforms require hosted remote MCP gateway acceptance in the platform console.
- [ ] Define stale credential behavior for older agent versions and changed metadata schema: new code must fail closed with a clean relogin path instead of trying to reinterpret unknown token records.
- [ ] Treat GoServer metadata, token schema, and remote MCP tool list as drift-prone deployment dependencies. The agent implementation must verify the live contract before release acceptance.

## V1 User Flows

### Direct MCP Login

```text
User runs:
  npx -y patchxnote-agent@latest mcp login

npm wrapper installs/verifies patchxnote binary
 -> delegates to patchxnote mcp login
 -> agent starts temporary 127.0.0.1 callback listener
 -> agent generates state + PKCE verifier/challenge
 -> agent opens GoServer authorization URL in browser
 -> user enters phone and OTP on GoServer page
 -> GoServer redirects to 127.0.0.1:<port>/callback?code=...&state=...
 -> agent validates state
 -> agent exchanges code with /v1/agent/oauth/token
 -> agent stores MCP OAuth connector credential in OS-native keychain
 -> agent runs a safe remote MCP smoke, defaulting to patchxnote_get_current_user
 -> CLI prints non-secret success summary
```

### Setup Reuse

```text
User runs:
  npx -y patchxnote-agent@latest setup --client cursor

setup builds the Cursor install plan
 -> setup checks MCP OAuth credential
 -> if missing or expired, setup invokes the same OAuth browser login helper
 -> setup writes or prints the client MCP config
 -> setup verifies stdio MCP initialize/tools/list
 -> setup optionally verifies safe read tool if credentials are available
```

### MCP Serve Runtime

```text
MCP host starts:
  npx -y patchxnote-agent@latest mcp serve

npm wrapper installs/verifies patchxnote binary
 -> delegates to patchxnote mcp serve
 -> mcp serve loads MCP OAuth connector credential if present
 -> if present, local stdio requests are forwarded to GoServer /mcp with Bearer connector token
 -> if token is near expiry, agent refreshes it before forwarding
 -> if remote MCP returns 401, agent refreshes once and retries the request
 -> if no MCP OAuth credential exists, V1 keeps the existing local MCP path for compatibility
 -> if neither credential model can authenticate, tools return auth_required with guidance to run patchxnote mcp login
```

### Logout

```text
User runs:
  npx -y patchxnote-agent@latest mcp logout

agent revokes refresh token and access token best-effort through /v1/agent/oauth/revoke
 -> agent deletes MCP OAuth connector credential from OS-native keychain
 -> agent prints a non-secret success summary
```

## Non-Goals

- [ ] Do not move the GoServer authorization page into this repository.
- [ ] Do not build the public MCP website in this iteration.
- [ ] Do not publish a VS Code extension or Codex plugin in this iteration.
- [ ] Do not add new MCP tools in this iteration.
- [ ] Do not add new GoServer OAuth endpoints in this iteration.
- [ ] Do not place bearer tokens or refresh tokens in MCP config files.
- [ ] Do not require users to paste authorization codes or tokens into chat.
- [ ] Do not auto-open browser from `mcp serve` during normal editor startup.
- [ ] Do not remove legacy terminal `patchxnote login` or the current local MCP path until a separate migration plan accepts that risk.

## Expected Files

- Create: `internal/oauthflow/pkce.go`
- Create: `internal/oauthflow/discovery.go`
- Create: `internal/oauthflow/callback.go`
- Create: `internal/oauthflow/browser.go`
- Create: `internal/oauthflow/store.go`
- Create: `internal/oauthflow/refresh.go`
- Create: `internal/oauthflow/flow.go`
- Create: `internal/oauthflow/*_test.go`
- Modify: `internal/api/types.go`
- Modify: `internal/api/client.go`
- Modify: `internal/api/client_test.go`
- Modify: `internal/cli/mcp.go`
- Create or modify: `internal/cli/mcp_login.go`
- Create or modify: `internal/cli/mcp_login_test.go`
- Modify: `internal/cli/setup.go`
- Modify: `internal/cli/setup_test.go`
- Create: `internal/remotemcp/client.go`
- Create: `internal/remotemcp/proxy.go`
- Create: `internal/remotemcp/*_test.go`
- Modify as needed: `internal/mcp/server.go`
- Modify: `packages/npm/bin/patchxnote-agent.js`
- Modify: `packages/npm/test/install.test.js`
- Modify: `scripts/e2e/mvp-smoke.sh`
- Modify as needed: `.github/workflows/release.yml`
- Modify as needed: `.github/workflows/publish-npm.yml`
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify: `docs/release-and-maintenance-runbook.zh-CN.md`
- Modify as needed: `docs/mcp-clients/README.zh-CN.md`
- Modify as needed: `docs/mcp-clients/client-detail-copy.zh-CN.md`

## Storage Contract

MCP OAuth credentials must use `internal/keychain.SecretStore` instead of the existing `keychain.Credential` record used by terminal Agent OTP login.

Suggested logical secret names:

```text
mcp_oauth:access_token
mcp_oauth:refresh_token
mcp_oauth:metadata
```

Suggested non-secret refresh lock path:

```text
<config_dir>/mcp-oauth-refresh.lock
```

Suggested metadata JSON:

```json
{
  "schema_version": "1",
  "server_base_url": "https://ws-lab.patch-x.cn/patchnote-test-api",
  "client_id": "patchxnote-local-dev",
  "token_type": "Bearer",
  "access_token_expires_at": "2026-08-27T12:00:00Z",
  "refresh_token_expires_at": "2026-09-26T12:00:00Z",
  "connector_session_id": "redacted-in-docs",
  "scope": "agent:account.read agent:content.read:mobile",
  "patchxnote_schema_notice": ""
}
```

Checklist:

- [ ] Metadata must not contain access token or refresh token values.
- [ ] Metadata may contain connector session ID because it is not a bearer secret, but command output should avoid printing it by default.
- [ ] Metadata must include `server_base_url`; `mcp serve` must ignore or reject the credential if the current runtime points at a different base URL.
- [ ] Metadata must include `client_id`; changing the OAuth client ID requires a new login.
- [ ] Metadata must include `token_type` and it must be `Bearer`.
- [ ] Metadata must include parsed access/refresh expiry timestamps; zero or past refresh expiry means `mcp login` is required.
- [ ] Scope strings must be parsed from the OAuth token response, normalized, and stored as returned by GoServer.
- [ ] Native keychain stores token fields separately from metadata.
- [ ] Secret names should be scoped by profile and server base URL, or the stored metadata must be checked before any token is used, so a test login cannot bleed into production.
- [ ] File keychain remains development-only and is still gated by `PATCHXNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true`.
- [ ] Linux Secret Service/keyring unavailable failures must fail closed for public login and include a short recovery hint.
- [ ] `patchxnote logout` must not accidentally delete MCP OAuth credentials unless a future UX explicitly merges logout behavior.
- [ ] `patchxnote mcp logout` must not delete legacy Agent OTP credentials.
- [ ] The refresh lock must live in the config directory and must not contain token material.
- [ ] Metadata decoder should ignore unknown future fields but reject missing required fields, unsupported schema versions, invalid expiry timestamps, and non-Bearer token types.

## API Contract

Add OAuth-specific client methods under `internal/api`.

Suggested request types:

```go
type OAuthTokenRequest struct {
    GrantType    string
    Code         string
    RedirectURI  string
    ClientID     string
    CodeVerifier string
    RefreshToken string
}

type OAuthTokenResponse struct {
    AccessToken            string `json:"access_token"`
    TokenType              string `json:"token_type"`
    ExpiresIn              int    `json:"expires_in"`
    RefreshToken           string `json:"refresh_token,omitempty"`
    RefreshTokenExpiresIn  int    `json:"refresh_token_expires_in,omitempty"`
    Scope                  string `json:"scope"`
    ConnectorSessionID     string `json:"connector_session_id,omitempty"`
    PatchXNoteSchemaNotice string `json:"patchxnote_schema_notice,omitempty"`
}
```

Suggested methods:

```go
func (c *Client) GetOAuthAuthorizationServer(ctx context.Context) (OAuthAuthorizationServerMetadata, error)
func (c *Client) ExchangeOAuthCode(ctx context.Context, request OAuthTokenRequest) (OAuthTokenResponse, error)
func (c *Client) RefreshOAuthToken(ctx context.Context, request OAuthTokenRequest) (OAuthTokenResponse, error)
func (c *Client) RevokeOAuthToken(ctx context.Context, token string) error
```

Checklist:

- [ ] Metadata endpoint uses `GET /.well-known/oauth-authorization-server` under the configured `server.base_url`.
- [ ] Metadata response must include `authorization_endpoint`, `token_endpoint`, `revocation_endpoint`, `response_types_supported`, `grant_types_supported`, and `code_challenge_methods_supported`.
- [ ] Login must verify metadata supports `response_type=code`, `grant_type=authorization_code`, `grant_type=refresh_token`, and `S256`.
- [ ] Login must verify discovered endpoints stay on the expected public HTTPS origin and path prefix, except explicit localhost development URLs.
- [ ] If metadata fetch fails in a local development test, implementation may fall back to conventional paths only when the caller explicitly allows it in tests.
- [ ] Token endpoint uses `application/x-www-form-urlencoded`, not JSON.
- [ ] Token exchange sends `grant_type=authorization_code`.
- [ ] Refresh sends `grant_type=refresh_token`.
- [ ] Public client sends no client secret.
- [ ] Revoke sends the token in a form body and never logs it.
- [ ] Token response validation requires non-empty access token, `token_type=Bearer`, positive `expires_in`, non-empty scope, and refresh-token fields for login responses.
- [ ] Refresh response validation must handle refresh-token rotation. If GoServer returns a new refresh token, replace the old one atomically; if it omits a refresh token, keep the previous one only if the server contract explicitly allows it.
- [ ] Clock skew must be handled conservatively by refreshing before actual expiry, for example with a short safety window.
- [ ] API errors map through the existing `api.Error` path.
- [ ] Tests assert Authorization header is not sent to token/revoke endpoints.
- [ ] Tests assert request bodies are form-encoded.
- [ ] Tests assert token values are never included in error strings.
- [ ] Tests assert non-2xx OAuth responses do not echo raw response bodies if the body could contain user or token material.

## Implementation Checklist

### Task 0: Confirm Workspace And Contracts

**Files:**

- Read: `AGENTS.md`
- Read: `README.md`
- Read: `README.zh-CN.md`
- Read: `docs/engineering-rules.md`
- Read: `docs/release-and-maintenance-runbook.zh-CN.md`
- Read: `docs/plans/2026-08-27-context7-style-setup-and-client-platform-mcp-checklist.md`
- Read: `../patchxNoteGoServer/internal/agentoauth/domain.go`
- Read: `../patchxNoteGoServer/internal/agentoauth/http.go`
- Read: `../patchxNoteGoServer/internal/agentoauth/service.go`
- Read: `../patchxNoteGoServer/internal/remotemcp/*`

Checklist:

- [ ] Run `git status --short --branch`.
- [ ] Record unrelated modified or untracked files and leave them untouched.
- [ ] Confirm GoServer test domain is still `https://ws-lab.patch-x.cn/patchnote-test-api`.
- [ ] Confirm GoServer authorize path is `/v1/agent/oauth/authorize`.
- [ ] Confirm GoServer token path is `/v1/agent/oauth/token`.
- [ ] Confirm GoServer revoke path is `/v1/agent/oauth/revoke`.
- [ ] Confirm GoServer remote MCP path is `/mcp`.
- [ ] Confirm GoServer still allows dynamic loopback callback ports.
- [ ] Confirm current local MCP tool count and remote MCP tool count before changing serve behavior.

Validation:

```sh
git status --short --branch
curl -fsS https://ws-lab.patch-x.cn/patchnote-test-api/mcp/health
curl -fsS https://ws-lab.patch-x.cn/patchnote-test-api/.well-known/oauth-authorization-server
```

Expected:

- Worktree state is understood.
- Remote MCP health returns current tool count.
- OAuth metadata points to the same public base URL used by the CLI.

### Task 1: Add OAuth Token API Client

**Files:**

- Modify: `internal/api/types.go`
- Modify: `internal/api/client.go`
- Modify: `internal/api/client_test.go`

Checklist:

- [ ] Write a failing test for authorization-code token exchange.
- [ ] Test path is `/v1/agent/oauth/token`.
- [ ] Test method is `POST`.
- [ ] Test content type is form-encoded.
- [ ] Test body includes `grant_type=authorization_code`, `code`, `redirect_uri`, `client_id`, and `code_verifier`.
- [ ] Test body does not include phone, OTP, access token, refresh token, or client secret.
- [ ] Implement `ExchangeOAuthCode`.
- [ ] Write a failing test for refresh token rotation.
- [ ] Implement `RefreshOAuthToken`.
- [ ] Write a failing test for revoke.
- [ ] Implement `RevokeOAuthToken`.
- [ ] Validate `idempotency-key` is not required for OAuth token/revoke unless GoServer later adds it.
- [ ] Ensure `api.Error` mapping works for 400, 401, 403, 404, and 429.

Validation:

```sh
go test ./internal/api -run 'TestOAuth|TestAgentOTP|TestRefreshAgentSession' -count=1
```

Expected:

- OAuth form requests pass.
- Existing Agent OTP and refresh tests still pass.
- No test prints token material.

### Task 2: Implement PKCE And Callback Flow Helpers

**Files:**

- Create: `internal/oauthflow/pkce.go`
- Create: `internal/oauthflow/discovery.go`
- Create: `internal/oauthflow/callback.go`
- Create: `internal/oauthflow/browser.go`
- Create: `internal/oauthflow/flow.go`
- Create: `internal/oauthflow/*_test.go`

Checklist:

- [ ] Write `GeneratePKCE` test.
- [ ] Generate verifier using crypto randomness.
- [ ] Verifier must match OAuth PKCE allowed character/length expectations.
- [ ] Challenge must be SHA-256 base64url without padding.
- [ ] Write authorize URL builder test.
- [ ] Authorize URL must include `response_type=code`, `client_id`, `redirect_uri`, `state`, `code_challenge`, and `code_challenge_method=S256`.
- [ ] Authorize URL must preserve `/patchnote-test-api` base path.
- [ ] Authorize URL must not include verifier, token, phone, or OTP.
- [ ] Write loopback callback server test.
- [ ] Callback server binds with an OS-assigned port on `127.0.0.1:0`; do not use a fixed port in normal login.
- [ ] Callback server listens only on `127.0.0.1`, not `0.0.0.0`, and should avoid `localhost` ambiguity where IPv6 or hosts-file behavior can differ.
- [ ] Callback path is `/callback`.
- [ ] Callback validates `state`.
- [ ] Callback rejects missing code.
- [ ] Callback rejects state mismatch.
- [ ] Callback handles OAuth denial responses such as `error=access_denied` and shows a short failure page without exposing raw query parameters.
- [ ] Callback returns a minimal local success/failure HTML page.
- [ ] Callback HTML responses set defensive headers: `Cache-Control: no-store`, `Content-Security-Policy`, `Referrer-Policy: no-referrer`, and `X-Content-Type-Options: nosniff`.
- [ ] Callback HTML must not include code, state, verifier, token, phone, OTP, or raw query string values.
- [ ] Callback shuts down after first terminal success/failure.
- [ ] Callback listener shuts down on context cancellation, timeout, or browser-launch failure.
- [ ] Browser opener uses argument arrays:
  - Windows: `rundll32 url.dll,FileProtocolHandler <url>`
  - macOS: `open <url>`
  - Linux: `xdg-open <url>`
- [ ] Browser opener has `--no-browser` equivalent path that prints the authorize URL to stderr only when explicitly needed; normal browser-open success should avoid printing real state/challenge URLs.

Validation:

```sh
go test ./internal/oauthflow -count=1
```

Expected:

- PKCE generation is deterministic only in tests via injected random source.
- Callback server is single-use and state-safe.
- No shell-concatenated browser command is used.

### Task 3: Add MCP OAuth Secure Store

**Files:**

- Create: `internal/oauthflow/store.go`
- Modify as needed: `internal/keychain/native_test.go`
- Modify as needed: `internal/keychain/file_test.go`
- Modify as needed: `internal/keychain/memory_test.go`

Checklist:

- [ ] Write a failing store test using `keychain.MemoryStore`.
- [ ] Store access token, refresh token, and metadata under MCP-specific secret names.
- [ ] Read returns a typed credential only when metadata and required token fields are complete.
- [ ] Missing store returns `ok=false`, not a fatal error.
- [ ] Unavailable store returns a user-facing error for login, but `mcp serve` maps it to auth_required where appropriate.
- [ ] Delete removes all MCP OAuth keys.
- [ ] Saving new credentials replaces old MCP OAuth credentials atomically as far as the current keychain abstraction allows.
- [ ] Existing Agent OTP credentials are not touched.
- [ ] Metadata JSON decode errors fail closed and tell user to run `patchxnote mcp logout` then login again.

Validation:

```sh
go test ./internal/oauthflow ./internal/keychain -count=1
```

Expected:

- MCP OAuth storage works on memory/file/native test doubles.
- Existing keychain tests still pass.

### Task 4: Implement `patchxnote mcp login`

**Files:**

- Modify: `internal/cli/mcp.go`
- Create: `internal/cli/mcp_login.go`
- Create: `internal/cli/mcp_login_test.go`
- Modify as needed: `internal/cli/root_test.go`

Suggested command:

```sh
patchxnote mcp login
patchxnote mcp login --no-browser
patchxnote mcp login --force
patchxnote mcp login --callback-timeout 5m
patchxnote mcp login --skip-smoke
patchxnote mcp login --output json
```

Checklist:

- [ ] Add `newMCPLoginCommand(state)`.
- [ ] Wire it into `newMCPCommand`.
- [ ] Reuse `loadRuntime(state)` and current `--server-base-url` default.
- [ ] Default OAuth client ID is `patchxnote-local-dev` for the current test/beta line.
- [ ] Reject non-HTTPS `--server-base-url` values unless they are explicit localhost development URLs.
- [ ] If MCP OAuth credential exists and is refresh-valid, print already logged-in unless `--force`.
- [ ] If `--force`, replace the existing MCP OAuth credential after a successful new login.
- [ ] Start loopback callback listener before opening browser.
- [ ] Build authorize URL from runtime server base URL.
- [ ] Open browser best-effort.
- [ ] Print manual URL to stderr if browser opening fails or `--no-browser` is set.
- [ ] Wait for callback with bounded timeout.
- [ ] Exchange code with `ExchangeOAuthCode`.
- [ ] Validate token response before storing: access token present, token type Bearer, positive expiry, refresh token present for initial login, scope present, and server metadata still matches current runtime.
- [ ] Store connector tokens through MCP OAuth store.
- [ ] Run safe remote MCP smoke unless `--skip-smoke`.
- [ ] Default smoke calls `tools/call patchxnote_get_current_user` through `/mcp`.
- [ ] All progress, browser-open guidance, warnings, and human diagnostics go to stderr when stdout is used for JSON output.
- [ ] `--output json` prints exactly one non-secret JSON object and exits with a deterministic status code.
- [ ] Plain output prints only:
  - login success
  - profile
  - server base URL
  - scopes count or scope names
  - expiry summary
- [ ] JSON output omits access token, refresh token, authorization code, verifier, phone, and OTP.
- [ ] On timeout, callback listener shuts down and no partial credential is stored.
- [ ] On token exchange failure, no partial credential is stored.
- [ ] On smoke failure after token exchange, keep credentials but return a clear warning or non-zero error according to the final UX decision.
- [ ] Exit code policy is covered by tests: success/already logged in returns 0, timeout/state mismatch/token failure returns non-zero, and skipped smoke is still success if credentials are stored.

Validation:

```sh
go test ./internal/cli -run 'TestMCPLogin|TestRootCommandIncludesMCP' -count=1
```

Expected:

- Command is visible in root help.
- Tests cover success, state mismatch, timeout, existing credential, force replacement, no-browser, JSON output, and no secret leakage.

### Task 5: Add `patchxnote mcp status` And `patchxnote mcp logout`

**Files:**

- Modify: `internal/cli/mcp.go`
- Create or modify: `internal/cli/mcp_login.go`
- Create or modify: `internal/cli/mcp_login_test.go`

Suggested commands:

```sh
patchxnote mcp status
patchxnote mcp logout
patchxnote mcp logout --local-only
patchxnote mcp logout --output json
```

Checklist:

- [ ] `mcp status` reports whether MCP OAuth credential exists.
- [ ] `mcp status` with no credential exits 0 and reports `authenticated=false` in JSON mode so scripts can inspect it.
- [ ] `mcp status` reports profile, server base URL, token expiry, refresh expiry, and scope names.
- [ ] `mcp status` can optionally call remote current-user smoke if `--verify` is added.
- [ ] `mcp status` never prints token values.
- [ ] `mcp logout` calls revoke on refresh token when present.
- [ ] `mcp logout` calls revoke on access token when present.
- [ ] `mcp logout --local-only` with no credential exits 0 and is idempotent.
- [ ] If revoke fails due to network, `--local-only` can clear local credentials.
- [ ] Default logout should delete local credentials even if remote revoke returns already-invalid or unauthorized.
- [ ] `patchxnote logout` remains legacy Agent OTP logout and does not delete MCP OAuth credentials in this iteration.

Validation:

```sh
go test ./internal/cli -run 'TestMCPStatus|TestMCPLogout' -count=1
```

Expected:

- Status and logout pass.
- Revoke requests never leak token strings in output or errors.

### Task 6: Implement Remote MCP HTTP Client And Stdio Proxy

**Files:**

- Create: `internal/remotemcp/client.go`
- Create: `internal/remotemcp/proxy.go`
- Create: `internal/remotemcp/client_test.go`
- Create: `internal/remotemcp/proxy_test.go`
- Modify: `internal/cli/mcp.go`
- Modify as needed: `internal/mcp/server.go`

Checklist:

- [ ] Write a remote MCP client test for `initialize`.
- [ ] Write a remote MCP client test for `tools/list`.
- [ ] Write a remote MCP client test for `tools/call`.
- [ ] HTTP method is `POST`.
- [ ] HTTP path is `<server_base_url>/mcp`.
- [ ] `Authorization: Bearer <access_token>` is sent only when a token is available.
- [ ] Accept header includes `application/json` and explicitly allows `text/event-stream` only if the proxy parser supports it.
- [ ] Content-Type for outbound JSON-RPC requests is `application/json`.
- [ ] HTTP client sets a bounded timeout suitable for editor startup and tool calls.
- [ ] HTTP client sends a non-secret User-Agent with agent version and OS/arch when available.
- [ ] Request body forwards the JSON-RPC object without adding secrets.
- [ ] Request body size is bounded before forwarding.
- [ ] Response body size is bounded.
- [ ] JSON-RPC response parser rejects malformed JSON, mismatched IDs, unsupported batch shape, and unexpected content types with stable MCP errors.
- [ ] If GoServer returns SSE/Streamable HTTP framing, the V1 client must parse the supported event format or fail with a clear compatibility error rather than treating event text as raw JSON.
- [ ] Non-200 HTTP responses map to JSON-RPC errors.
- [ ] HTTP 401 maps to `auth_required`.
- [ ] HTTP 403 maps to `permission_denied`.
- [ ] HTTP 429 maps to `rate_limited`.
- [ ] Stdio proxy preserves request IDs exactly.
- [ ] Notifications are either forwarded safely or ignored according to current local MCP semantics.
- [ ] `mcp serve` stdout remains JSON-RPC only.
- [ ] Refresh diagnostics, if any, go to stderr and avoid token material.
- [ ] If MCP OAuth credential exists, `mcp serve` uses remote proxy.
- [ ] If MCP OAuth credential is missing, `mcp serve` keeps the existing local MCP server behavior for compatibility in this release.
- [ ] If remote proxy gets 401 and refresh token exists, refresh once and retry the original JSON-RPC request.
- [ ] If refresh fails unauthorized, delete MCP OAuth credential and return `auth_required`.
- [ ] If remote proxy cannot reach GoServer, return `api_unavailable` or `transport_error` without falling into infinite retries.
- [ ] If current `server.base_url` differs from stored credential metadata, do not send the stored token; return `auth_required` with relogin guidance.

Validation:

```sh
go test ./internal/remotemcp ./internal/mcp ./internal/cli -run 'TestRemote|TestMCPServe|TestServer' -count=1
```

Expected:

- Remote proxy tests pass.
- Existing local MCP tests still pass.
- No MCP stdout pollution.

### Task 7: Replace Setup Browser Setup-Session Path With MCP OAuth Login

**Files:**

- Modify: `internal/cli/setup.go`
- Modify: `internal/cli/setup_test.go`
- Modify as needed: `internal/api/types.go`
- Modify as needed: `internal/api/client.go`
- Modify as needed: `internal/api/client_test.go`
- Update docs: `docs/plans/2026-08-27-agent-browser-setup-session-contract.md`

Checklist:

- [ ] Keep setup planning, config write, backup, rollback, and manual client behavior unchanged.
- [ ] Replace `tryBrowserSetupSession` primary path with a call to the same OAuth browser login helper used by `mcp login`.
- [ ] Keep old setup-session API types only if tests or docs still need historical compatibility.
- [ ] If old setup-session code remains, hide it from public docs and mark it historical in code comments or tests.
- [ ] Do not call terminal phone OTP fallback for MCP setup by default once remote MCP OAuth is required.
- [ ] If browser OAuth is unavailable, return clear guidance:
  - run `patchxnote mcp login --no-browser`
  - or run `patchxnote login` only for legacy local MCP mode
- [ ] `--no-browser` should print the OAuth URL and wait for callback, not fall back to asking for phone OTP in the terminal.
- [ ] Existing `--dry-run` does not start login.
- [ ] Existing `--print-config` still prints secret-free config.
- [ ] Existing `--yes` only bypasses config-write confirmation, not browser account approval.
- [ ] Setup JSON output reports `auth_method=mcp_oauth` without secrets.
- [ ] Setup preserves existing client config backups and rollback behavior if OAuth login succeeds but config writing fails.
- [ ] Setup must not start browser login for platform-only clients that cannot use local stdio config; those should show hosted remote MCP instructions instead.

Validation:

```sh
go test ./internal/cli ./internal/setup -run 'TestSetup|TestMCPLogin' -count=1
```

Expected:

- Setup tests pass.
- Existing dry-run and manual client behavior remains stable.
- Browser setup-session is no longer the claimed product path.

### Task 8: Update npm Wrapper For `mcp login/status/logout`

**Files:**

- Modify: `packages/npm/bin/patchxnote-agent.js`
- Modify: `packages/npm/test/install.test.js`
- Modify as needed: `packages/npm/package.json`

Checklist:

- [ ] Parser accepts `mcp login`.
- [ ] Parser accepts `mcp status`.
- [ ] Parser accepts `mcp logout`.
- [ ] `mcp login/status/logout` install or verify the pinned binary, then delegate to:
  - `patchxnote mcp login`
  - `patchxnote mcp status`
  - `patchxnote mcp logout`
- [ ] Launcher options like `--install-dir`, `--platform`, `--arch`, and `--from-local` stay wrapper-level.
- [ ] Runtime args after launcher parsing pass through to Go binary.
- [ ] `mcp config` remains pure stdout JSON and should not install if current behavior intentionally avoids install.
- [ ] `mcp serve` keeps stdout reserved for JSON-RPC.
- [ ] Install diagnostics for `mcp serve` still go to stderr.
- [ ] `mcp login/status/logout` human diagnostics go to stderr when `--output json` is used.
- [ ] Wrapper tests include Windows path-with-spaces command construction.
- [ ] Tests use fake binary log capture to assert delegated argv.

Validation:

```sh
node packages/npm/test/install.test.js
```

Expected:

- Existing npm install/update/uninstall/setup/login tests pass.
- New `mcp login/status/logout` parser and delegation tests pass.

### Task 9: Update MVP Smoke And Local Acceptance Scripts

**Files:**

- Modify: `scripts/e2e/mvp-smoke.sh`
- Modify or create as needed: `test/e2e/mvp_test.go`
- Modify as needed: `.github/workflows/macos-install-smoke.yml`

Checklist:

- [ ] Add smoke case for `patchxnote mcp login --help`.
- [ ] Add smoke case for `patchxnote mcp status --output json` with no credentials.
- [ ] Add smoke case for `patchxnote mcp logout --local-only` with no credentials.
- [ ] Add npm wrapper dry-run or fake-binary case for `mcp login`.
- [ ] Keep smoke scripts from requiring real browser login in CI.
- [ ] Add a manual acceptance script or documented command for real browser login against test GoServer.
- [ ] The manual acceptance script must never print token, code, verifier, phone, or OTP.
- [ ] CI smoke covers no-credential `mcp serve` startup without requiring browser login.
- [ ] CI smoke covers fake credential metadata mismatch so tokens are not sent to the wrong base URL.

Validation:

```sh
scripts/e2e/mvp-smoke.sh
go test ./test/e2e -count=1
```

Expected:

- CI-safe smoke passes without real browser input.
- Manual browser login acceptance is documented but not run in non-interactive CI.

### Task 10: Documentation Updates

**Files:**

- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify: `docs/release-and-maintenance-runbook.zh-CN.md`
- Modify as needed: `docs/mcp-clients/README.zh-CN.md`
- Modify as needed: `docs/mcp-clients/client-detail-copy.zh-CN.md`
- Modify: `docs/plans/2026-08-27-context7-style-setup-and-client-platform-mcp-checklist.md`
- Modify: `docs/plans/2026-08-27-agent-browser-setup-session-contract.md`

Checklist:

- [ ] Public docs recommend `npx -y patchxnote-agent@latest mcp login` for browser OAuth login.
- [ ] Public docs keep `npx -y patchxnote-agent@latest mcp config` as secret-free config generator.
- [ ] Public docs explain that `setup --client <id>` reuses the same browser OAuth login.
- [ ] Public docs explain Windows vs WSL keychain separation.
- [ ] Public docs explain `mcp serve` does not auto-open browser on editor startup.
- [ ] Docs state that the GoServer website/login page owns phone OTP input.
- [ ] Docs state that the local agent owns callback, token exchange, secure storage, and stdio bridging.
- [ ] Remove or downgrade older setup-session copy so it is not presented as the active product path.
- [ ] Document `mcp logout` and local-only recovery.
- [ ] Do not publish raw OAuth URLs containing real state/code/challenge examples copied from acceptance runs.
- [ ] Docs list supported client tiers:
  - P0 local stdio: VS Code, Cursor, Codex, Claude Code, Claude Desktop, Windsurf.
  - P0/P1 platform-hosted remote MCP: Feishu Aily, Doubao WorkBuddy, Tencent Agent platforms, enterprise WorkBuddy-style platforms.
  - Manual generic MCP: any client that accepts a command/args/env MCP server definition.
- [ ] Docs do not claim one-click install for Codex, WorkBuddy, Feishu Aily, Doubao WorkBuddy, or Tencent Agent platforms until each exact flow has been manually accepted.

Validation:

```sh
git diff --check
node docs/mcp-clients/validate-clients.mjs
```

Expected:

- Docs are consistent with the deployed GoServer OAuth model.
- Client registry remains valid.

### Task 11: Local Test Matrix

**Files:**

- No direct code changes unless tests fail.

Checklist:

- [ ] Run API client tests.
- [ ] Run OAuth flow tests.
- [ ] Run remote MCP proxy tests.
- [ ] Run CLI tests.
- [ ] Run setup tests.
- [ ] Run existing MCP tests.
- [ ] Run keychain tests.
- [ ] Run npm wrapper tests.
- [ ] Run MVP smoke.
- [ ] Run docs/client registry validation.
- [ ] Run secret-output grep over generated stdout/stderr fixtures.
- [ ] Run `go test ./...` once targeted suites are green, and record any known unrelated baseline failures separately.
- [ ] Run tests sequentially; if any test harness defaults to workers, force single-worker mode where supported.

Validation:

```sh
go test ./internal/api ./internal/oauthflow ./internal/remotemcp ./internal/cli ./internal/setup ./internal/mcp ./internal/keychain -count=1
go test ./... -count=1
node packages/npm/test/install.test.js
node docs/mcp-clients/validate-clients.mjs
scripts/e2e/mvp-smoke.sh
git diff --check
```

Expected:

- All targeted tests pass.
- No generated output contains:
  - `access_token`
  - `refresh_token`
  - `authorization_code`
  - phone numbers
  - OTP values
  - webhook URLs
  - provider payloads

### Task 12: Real Browser Acceptance Against Test GoServer

**Files:**

- No tracked file writes required, except optional evidence note with secrets redacted.

Checklist:

- [ ] Build local binary.
- [ ] Run `patchxnote mcp logout --local-only` for the test profile to start clean.
- [ ] Run `patchxnote mcp login --server-base-url https://ws-lab.patch-x.cn/patchnote-test-api`.
- [ ] Confirm browser opens the GoServer authorization page.
- [ ] User enters phone and OTP only in the browser.
- [ ] Confirm local callback success page appears.
- [ ] Confirm CLI prints non-secret success.
- [ ] Run `patchxnote mcp status --output json`.
- [ ] Run a stdio MCP smoke through `patchxnote mcp serve`:
  - `initialize`
  - `tools/list`
  - `tools/call patchxnote_get_current_user`
- [ ] Confirm `patchxnote_get_current_user` returns the expected authorized account projection.
- [ ] Confirm no token/code/verifier/phone/OTP appears in terminal logs.
- [ ] Repeat or document Windows-native acceptance separately from WSL acceptance, because Windows editors normally use Windows Credential Manager while WSL uses Linux keyring behavior.
- [ ] Record only redacted evidence: command names, version, pass/fail status, response shape, and account projection without raw phone, tokens, codes, or complete content records.

Suggested commands:

```sh
go build -trimpath -o .tmp/acceptance/patchxnote ./cmd/patchxnote
.tmp/acceptance/patchxnote mcp logout --local-only --profile mcp-oauth-acceptance
.tmp/acceptance/patchxnote mcp login --profile mcp-oauth-acceptance --server-base-url https://ws-lab.patch-x.cn/patchnote-test-api
.tmp/acceptance/patchxnote mcp status --profile mcp-oauth-acceptance --output json
```

Expected:

- Browser OAuth closes through the real GoServer.
- MCP connector token is saved locally.
- `mcp serve` can call the remote safe read tool.

### Task 13: Published Package Acceptance

**Files:**

- No tracked file writes required, except optional evidence note with secrets redacted.

Checklist:

- [ ] Publish or install a candidate package version only after code review and local tests pass.
- [ ] Confirm package version, native binary version, Git tag, release commit, and checksums all refer to the same source revision.
- [ ] From a normal Windows working directory, run `npx -y patchxnote-agent@<candidate> mcp login`.
- [ ] From a Windows path containing spaces, run `npx -y patchxnote-agent@<candidate> mcp status --output json`.
- [ ] Complete browser login against the test GoServer.
- [ ] Run `npx -y patchxnote-agent@<candidate> mcp status --output json`.
- [ ] Run `npx -y patchxnote-agent@<candidate> mcp config` and confirm stdout is pure JSON.
- [ ] Run a stdio MCP smoke through `npx -y patchxnote-agent@<candidate> mcp serve`.
- [ ] Verify Cursor or VS Code setup through `npx -y patchxnote-agent@<candidate> setup --client cursor --yes`.
- [ ] Verify Codex manual command output remains correct.
- [ ] Record exact package version, binary version, OS, and test server base URL.
- [ ] Do not store tokens or OTP in evidence.

Expected:

- Published artifact works outside the source checkout.
- Windows path with spaces and npm cold install both work.

### Task 14: Client Acceptance Matrix

**Files:**

- Modify as needed: `docs/mcp-clients/README.zh-CN.md`
- Modify as needed: `docs/mcp-clients/client-detail-copy.zh-CN.md`
- Optional evidence only: `docs/evidence/<date>-mcp-client-acceptance.zh-CN.md`

Checklist:

- [ ] VS Code: generated user-level MCP config is secret-free, starts `npx -y patchxnote-agent@<candidate> mcp serve`, lists tools, and calls `patchxnote_get_current_user`.
- [ ] Cursor: setup writes or prints the expected MCP config, editor reload sees the server, lists tools, and calls `patchxnote_get_current_user`.
- [ ] Codex: manual command/config instructions are correct; only claim marketplace/plugin distribution after separate Codex plugin submission exists.
- [ ] Claude Code: command-based MCP add path is documented and accepted with the candidate package.
- [ ] Claude Desktop: JSON config path, backup behavior, and restart requirement are documented and accepted.
- [ ] Windsurf: setup/manual config path is documented and accepted with the candidate package.
- [ ] Generic local MCP: `mcp config` output can be pasted into clients that accept command/args MCP server definitions.
- [ ] Feishu Aily / Doubao WorkBuddy / Tencent Agent platforms: mark as hosted remote MCP acceptance, requiring GoServer `/mcp` URL, OAuth authorization, platform permissions, and platform console testing.
- [ ] Enterprise WorkBuddy-style clients: document as platform-dependent until the target tenant confirms whether it supports local stdio, remote MCP, or only vendor-approved integrations.
- [ ] Every client entry records status as one of `accepted`, `manual-only`, `remote-platform-pending`, or `unsupported-in-v1`; avoid ambiguous marketing labels.

Validation:

```sh
node docs/mcp-clients/validate-clients.mjs
```

Expected:

- Client support claims match actual tested installation paths.
- Platform clients are not marked accepted until the real platform-side OAuth/MCP flow has been verified.

## Boundary And Risk Checklist

- [ ] Browser URL launch must not use shell string concatenation.
- [ ] Loopback callback must bind only `127.0.0.1`, not `0.0.0.0`.
- [ ] Callback port conflicts are avoided by using an OS-assigned port; retries must create a new state and verifier.
- [ ] Multiple concurrent `mcp login` attempts should not overwrite each other's callback state or stored credentials before success.
- [ ] State mismatch must fail closed.
- [ ] Token exchange must fail closed on missing code or verifier.
- [ ] Token refresh must be single-flight or lock-protected enough for multiple editor MCP processes.
- [ ] Refresh-token rotation race must be handled so an older process does not write back a stale refresh token after a newer refresh succeeds.
- [ ] `mcp serve` must not block `initialize` waiting for browser login.
- [ ] `mcp serve` must keep stdout JSON-RPC clean even during install, refresh, and error paths.
- [ ] MCP OAuth credentials and legacy Agent OTP credentials must not overwrite each other.
- [ ] Changing `--profile`, `server.base_url`, or OAuth client ID must not silently reuse another credential.
- [ ] Windows Credential Manager and WSL/Linux Secret Service are separate stores; docs and setup warnings must say so.
- [ ] Linux headless environments may not have `xdg-open`; `--no-browser` must work.
- [ ] Corporate machines may block loopback browser callbacks; timeout and retry guidance must be clear.
- [ ] Corporate proxies may block GoServer; error should mention server connectivity without leaking request bodies.
- [ ] Default-browser unavailable, cancelled, or non-zero opener exit should fall back to manual URL guidance without leaving a listener behind after timeout.
- [ ] OAuth server metadata origin/path mismatch must fail closed to avoid sending credentials to an unexpected endpoint.
- [ ] Remote MCP tool output size must remain bounded by GoServer and local proxy read limits.
- [ ] Refresh failure should delete invalid MCP OAuth credentials only after a clear unauthorized response, not after transient network errors.
- [ ] Revoke failure should not leave the user unable to clear local credentials.
- [ ] Setup should not modify any client config unless user confirmed or passed `--yes`.
- [ ] Setup must only add or replace the `patchxnote` MCP server entry, not remove unrelated servers.
- [ ] Do not claim one-click install for a client until that exact client path has been tested.

## Rollout Slices

### Slice A: CLI Browser OAuth Login

- [ ] Implement API form methods.
- [ ] Implement OAuth metadata discovery and validation.
- [ ] Implement OAuth flow helpers.
- [ ] Implement MCP OAuth secure store.
- [ ] Implement `patchxnote mcp login/status/logout`.
- [ ] Validate with local fake server.

### Slice B: Remote MCP Proxy

- [ ] Implement remote MCP HTTP client.
- [ ] Add stdio proxy mode when MCP OAuth credentials exist.
- [ ] Preserve legacy local MCP fallback.
- [ ] Validate with local httptest remote MCP server.

### Slice C: Setup Integration

- [ ] Replace setup-session primary path with MCP OAuth login helper.
- [ ] Keep config install behavior unchanged.
- [ ] Validate setup dry-run, manual clients, and auto-write clients.

### Slice D: npm And Docs

- [ ] Add npm wrapper delegation for `mcp login/status/logout`.
- [ ] Update public docs and runbook.
- [ ] Run npm tests and docs checks.

### Slice E: Real Acceptance And Release Candidate

- [ ] Run real browser OAuth against `ws-lab.patch-x.cn`.
- [ ] Run stdio MCP safe read through remote proxy.
- [ ] Test candidate npm package from a clean Windows directory.
- [ ] Record redacted evidence.

### Slice F: Client Matrix Acceptance

- [ ] Accept at least VS Code and Cursor or Codex on the candidate package before claiming local-client V1 ready.
- [ ] Mark Claude Code, Claude Desktop, Windsurf, and generic MCP as accepted or manual-only based on the real flow tested.
- [ ] Mark Feishu Aily, Doubao WorkBuddy, Tencent Agent platforms, and enterprise WorkBuddy-style clients as remote-platform-pending until platform-side hosted MCP acceptance is complete.
- [ ] Feed the accepted matrix back into website copy and install cards later; do not let website copy lead the tested reality.

## Acceptance Criteria

- [x] `patchxnote mcp login` completes browser OAuth and stores MCP connector credentials.
- [ ] `npx -y patchxnote-agent@latest mcp login` delegates correctly.
- [x] OAuth metadata discovery succeeds against the configured GoServer base URL or fails closed with a relogin/config hint.
- [x] No MCP config file contains tokens, phone numbers, OTPs, webhook secrets, or server secrets.
- [x] `patchxnote mcp status` reports login state without secrets.
- [x] `patchxnote mcp logout` revokes best-effort and clears local MCP OAuth credentials.
- [x] `patchxnote mcp serve` can initialize, list tools, and call `patchxnote_get_current_user` using MCP OAuth credentials.
- [x] `patchxnote mcp serve` still supports existing local Agent credentials as a compatibility fallback for this release.
- [ ] `setup --client cursor` and at least one other P0 local client reuse MCP OAuth login and remain secret-free.
- [ ] Windows real browser login works from a normal non-repo working directory.
- [x] WSL behavior is documented as a separate keychain/runtime boundary.
- [x] `go test` targeted suites pass.
- [x] `go test ./...` passes or any unrelated baseline failure is explicitly recorded and accepted before release.
- [x] npm wrapper tests pass.
- [x] MVP smoke passes.
- [x] Client acceptance matrix records each supported editor/platform with a concrete state and no untested one-click claims.
- [ ] Published artifact smoke passes before marking the feature release accepted.
