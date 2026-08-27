# Agent Browser Setup Session Contract

**Status:** historical design note. The active product path is now `patchxnote mcp login` using GoServer OAuth authorize/token/revoke with PKCE and a loopback callback.

**Goal:** let `patchxnote setup --client <id>` open a PatchXNote web login page, receive an Agent session after user approval, store credentials in OS-native secure storage, then install a secret-free MCP config into the selected local client.

## Historical Client Behavior

- Older setup-session code was designed to check existing local Agent credentials, try a browser setup-session API, and fall back to terminal phone OTP.
- Public V1 onboarding should not present this as the active browser-login path.
- `patchxnote setup --client <id>` now reuses the MCP OAuth login helper. If a local stdio client needs credentials and the MCP OAuth credential is missing or expired, setup starts the same browser OAuth flow as `patchxnote mcp login`.
- `--no-browser` prints the OAuth URL and waits for the loopback callback; it does not fall back to terminal phone OTP.
- Setup never writes phone numbers, OTPs, access tokens, refresh tokens, source text, provider payloads, or webhook secrets to client config.

## Active OAuth Product Path

```text
patchxnote mcp login
 -> discover /.well-known/oauth-authorization-server
 -> open /v1/agent/oauth/authorize with PKCE S256
 -> receive http://127.0.0.1:<port>/callback
 -> exchange code at /v1/agent/oauth/token
 -> store MCP OAuth credential in MCP-specific keychain secret names
 -> patchxnote mcp serve proxies local stdio requests to <server_base_url>/mcp when that credential is present
```

## Proposed Server Endpoints

### Create setup session

```http
POST /v1/agent/setup-sessions
Content-Type: application/json
Idempotency-Key: <opaque id>
```

Request:

```json
{
  "client_id": "cursor",
  "client_name": "Cursor",
  "profile": "default",
  "scopes": ["agent:account.read", "agent:memories.read"]
}
```

Response:

```json
{
  "session_id": "setup_fixture",
  "status": "pending",
  "user_code": "PXNOTE-1234",
  "verification_uri": "https://patchxnote.com/mcp/setup",
  "verification_uri_complete": "https://patchxnote.com/mcp/setup?code=PXNOTE-1234",
  "expires_in_seconds": 600,
  "poll_interval_seconds": 2
}
```

### Poll setup session

```http
GET /v1/agent/setup-sessions/{session_id}
```

Pending response:

```json
{
  "session_id": "setup_fixture",
  "status": "pending"
}
```

Approved response:

```json
{
  "session_id": "setup_fixture",
  "status": "approved",
  "session": {
    "access_token": "<server-only response>",
    "access_expires_in_seconds": 3600,
    "refresh_token": "<server-only response>",
    "refresh_expires_in_seconds": 2592000,
    "account": {
      "id": "acct_fixture",
      "status": "active"
    },
    "scopes": ["agent:account.read", "agent:memories.read"]
  }
}
```

The CLI stores the token fields immediately in secure storage and never prints them.

## Required Server Guarantees

- Setup session is short-lived, single-use, and bound to `client_id`, requested scopes, account, and server environment.
- Browser approval page displays client name, setup code, scopes, environment, expiry, and security warning.
- User must confirm the same setup code shown by the CLI before approval.
- Reuse, denial, expiry, mismatched client, and mismatched environment return stable non-secret errors.
- Setup session does not create, consume, replace, or bind `mobile` or `desktop` installations.
- Logs record only safe diagnostics: request id, stable status, client id, account id, and error code.

## CLI Fallback

`patchxnote login` remains as the legacy terminal OTP Agent login path. It is still useful for the old direct `/v1/agent/**` local MCP fallback, but it is no longer the setup product path for MCP browser login.
