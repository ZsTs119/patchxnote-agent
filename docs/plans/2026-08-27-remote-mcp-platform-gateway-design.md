# Remote MCP Platform Gateway Design

**Status:** background design. The implementation checklist is now superseded by `docs/plans/2026-08-27-remote-mcp-goserver-parity-checklist.md`, which changes the V1 target from a reduced read-only subset to 19-tool functional parity with the local MCP server.

**V1 decision:** build the first remote MCP gateway as a GoServer-integrated route, fronted by `https://mcp.patchxnote.com/mcp`. This keeps account auth, quotas, rate limits, audit IDs, and existing Agent API authorization in one service boundary. A separate gateway service can come later if traffic or deployment isolation requires it.

## Why Remote MCP Exists

Local editors can launch:

```sh
npx -y patchxnote-agent@latest mcp serve
```

Cloud/platform agents such as Feishu Aily, Doubao Work Partner, Tencent Agent Development Platform, and enterprise WorkBuddy normally need an HTTPS MCP URL. They cannot depend on a user's local terminal, npm cache, keychain, or machine state.

## Endpoint Shape

```text
GET  https://mcp.patchxnote.com/health
POST https://mcp.patchxnote.com/mcp
```

Transport:

- Streamable HTTP first.
- SSE only if a target P0.5 platform requires it in acceptance.
- No stdio in platform mode.

MCP methods:

- `initialize`
- `tools/list`
- `tools/call`

## V1 Tool Surface

The earlier reduced read-only V1 surface is superseded. Remote MCP V1 now targets all 19 local `patchxnote_*` MCP tools for functional parity. Local filesystem-backed behavior must be adapted to server-side state or bounded returned content:

- webhook configure/list/remove/send uses GoServer account-scoped storage and outbound send;
- render/export returns bounded content or a short-lived download handle instead of writing to a user's local filesystem;
- search runs against server-authorized safe metadata instead of a local MCP session cache.

Still do not expose raw audio download, unbounded transcripts, arbitrary OpenAPI access, App/PC installation mutation, hardware connect/release, or model-run execution.

## Auth Model

Preferred V1 auth:

1. User starts platform connector authorization from PatchXNote web.
2. Server creates a revocable connector session for the selected platform.
3. Platform stores only the connector credential required to call remote MCP, never PatchXNote Agent access/refresh tokens.
4. Gateway exchanges connector context server-side into authorized Agent reads.
5. User/operator can list and revoke connector sessions.

Do not ask users to paste PatchXNote access tokens into Feishu, Tencent, WorkBuddy, or any third-party platform config.

## Request Binding And Audit

Every remote MCP call must bind:

- PatchXNote account
- platform client id
- connector session id
- scopes
- MCP request id
- server audit request id

Logs may include only safe diagnostics: version, platform client, status, stable error code, request id, tool name, latency bucket, and bounded item counts.

## Error Mapping

Use stable MCP errors:

- `unauthenticated`
- `forbidden`
- `not_found`
- `invalid_request`
- `rate_limited`
- `upstream_unavailable`

Never return raw upstream error bodies, token values, phone numbers, prompts, transcripts, provider payloads, or webhook URLs.

## Limits

- Page size default 10, max 50.
- Tool result text bounded by server-side byte caps.
- No caching of full record content in V1.
- Rate limit by account, connector, platform client, IP/platform source, and tool.
- Keep tool descriptions concise so platform agents select the right tool reliably.

## Platform Acceptance Matrix

| Platform | Required Transport | Auth | V1 Status |
| --- | --- | --- | --- |
| Feishu Aily / Doubao Work Partner | Streamable HTTP or SSE | connector session | blocked until remote endpoint and platform test tenant are ready |
| Tencent Agent Development Platform | Streamable HTTP or SSE | connector session | blocked until remote endpoint and platform test tenant are ready |
| Enterprise WorkBuddy | Streamable HTTP or SSE | connector session | blocked until enterprise connector docs and account are ready |

## Deployment Gates

- Health endpoint returns non-sensitive status.
- MCP inspector initializes and lists the same 19 local MCP tools.
- One safe read call returns account-owned metadata only.
- Revocation removes platform access.
- Sensitive-value scan passes for logs and checked-in evidence.
- Platform evidence notes contain no raw phone, token, OTP, raw source text, full transcript, provider payload, full MAC, SK, or webhook secret.
