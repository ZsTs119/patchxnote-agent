# Platform Client PoC Status

**Date:** 2026-08-27

**Scope:** Feishu Aily / Doubao Work Partner, Tencent Agent Development Platform, and WorkBuddy enterprise/platform mode.

## Current Result

Platform PoC acceptance is not yet passed because the remote MCP gateway is not deployed and no platform connector credential/tenant acceptance has been completed.

This is an explicit blocked state, not a local setup failure. Local clients can use the npm stdio server; platform clients need the planned HTTPS endpoint:

```text
https://mcp.patchxnote.com/mcp
```

## Client States

| Client | Connection Method | Transport | Tool Count | Safe Read Call | Status |
| --- | --- | --- | --- | --- | --- |
| Feishu Aily / Doubao Work Partner | remote MCP URL | Streamable HTTP or SSE | not verified | not run | blocked: gateway not deployed |
| Tencent Agent Development Platform | remote MCP URL | Streamable HTTP or SSE | not verified | not run | blocked: gateway not deployed |
| Enterprise WorkBuddy | remote MCP URL | Streamable HTTP or SSE | not verified | not run | blocked: gateway not deployed |
| WorkBuddy desktop local path | local MCP + CLI | stdio | not verified in app | not run in app | manual V1 path only |

## Acceptance Needed Next

- Deploy or stage the remote MCP gateway.
- Create a revocable platform connector session.
- Add the remote MCP URL in at least one Feishu/Doubao or Tencent platform test workspace.
- Run `initialize`, `tools/list`, and one safe read-only `tools/call`.
- Record only sanitized evidence: client name, connection method, transport, tool count, and success/failure code.

## Sensitive Data Review

No raw phone number, OTP, access token, refresh token, webhook secret, raw audio, full transcript, source text dump, provider payload, full MAC, SK, or real platform credential is recorded in this evidence note.
