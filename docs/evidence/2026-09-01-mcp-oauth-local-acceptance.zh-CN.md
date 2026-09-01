# MCP OAuth Local Acceptance

**Date:** 2026-09-01

**Candidate:** `patchxnote-agent` / `patchxnote` `0.2.9` source candidate

**Server:** local GoServer real-OTP test runtime on loopback, plus PatchXNote public beta API for wrapper smoke

**Scope:** Browser MCP login callback pages, OTP resend/change-phone interaction, token exchange, local credential storage, MCP status verification, and release preflight checks.

## Result

Local candidate acceptance passed before publishing `0.2.9`. This proves the source candidate can complete the real browser login loop and that the Agent callback UI no longer exposes technical OAuth details. It does not claim per-editor UI acceptance or hosted platform-client acceptance.

## Automated Gates

Passed:

```text
go test -count=1 -p=1 -parallel=1 ./...
node packages/npm/test/install.test.js
node docs/mcp-clients/validate-clients.mjs
scripts/e2e/mvp-smoke.sh
npm pack --dry-run
git diff --check
```

The npm pack dry-run was executed from a normal Windows temporary directory to avoid WSL UNC path behavior.

## Browser Login Acceptance

Accepted with a real local GoServer API and worker runtime:

- GoServer OAuth metadata returned local authorize/token/revoke endpoints.
- `patchxnote mcp login` opened the browser flow through the local Agent callback.
- The user completed phone OTP only in the browser page.
- The flow handled `send code -> change phone -> send again -> login` using the latest OTP request state.
- GoServer observed the expected sequence: OTP request accepted, OTP verification accepted, authorize completed, token exchange completed, and remote MCP read check completed.
- `patchxnote mcp status --verify` passed for the same profile.

No phone number, OTP, authorization code, PKCE verifier, access token, refresh token, source text, model payload, full MAC, SK, or webhook secret is recorded in this note.

## UI Boundary

The local Agent loopback result page now shows ordinary user-facing success/failure states:

- Success: `登录已完成`
- Failure: `登录未完成`

Callback page tests assert that the page does not expose OAuth code, state, token-shaped values, or related implementation details.

## Pending Acceptance

- Fresh `patchxnote-agent@0.2.9` registry-package browser OAuth login with user OTP: pending, because this release closeout did not ask the user to complete a new published-package OTP login.
- VS Code / Cursor / Codex / Claude / Windsurf / Trae / Qoder / WorkBuddy UI acceptance: pending per client.
- Feishu Aily / Doubao Work Partner / Tencent Agent Platform / enterprise WorkBuddy hosted platform-console acceptance: pending.
