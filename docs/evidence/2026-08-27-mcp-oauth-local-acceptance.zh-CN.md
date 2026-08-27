# MCP OAuth Local Acceptance

**Date:** 2026-08-27

**Candidate:** `patchxnote-agent` / `patchxnote` `0.2.8` local candidate

**Server:** `https://ws-lab.patch-x.cn/patchnote-test-api`

**Scope:** Browser OAuth login, local secure credential storage, remote MCP proxy, stdio MCP compatibility, npm wrapper delegation, and authorized Agent content/model-IO read paths.

## Result

Local candidate acceptance passed. This evidence proves the generic Windows local stdio path works before npm publication. It does not yet prove the registry-published package, each editor UI integration, or platform-hosted MCP clients.

## Automated Gates

The following gates passed before real browser acceptance:

```text
/home/zsts_119/.local/go1.24.11/bin/go test ./... -count=1
node packages/npm/test/install.test.js
node docs/mcp-clients/validate-clients.mjs
scripts/e2e/mvp-smoke.sh
git diff --check
```

Live server contract checks also passed:

- `/mcp/health` returned the current remote MCP server shape with 19 tools and protocol `2026-07-28`.
- `/.well-known/oauth-authorization-server` returned the OAuth metadata needed by `mcp login`.

## Windows Browser OAuth Acceptance

Accepted from a Windows-native process, using isolated profile `mcp-oauth-windows-acceptance`.

Flow:

1. Built a Windows native candidate binary.
2. Cleared the isolated MCP OAuth profile with `mcp logout --local-only`.
3. Ran `mcp login` without `--no-browser`.
4. Confirmed the command attempted automatic browser launch instead of printing a manual URL.
5. User completed phone/OTP login only in the GoServer browser page.
6. Local callback returned to the CLI.
7. CLI saved the MCP OAuth credential in the Windows runtime credential store and printed only a non-secret success summary.
8. `mcp status --verify` returned `authenticated=true` and `verified=true`.

No OTP, OAuth code, PKCE verifier, access token, refresh token, or raw phone number was recorded.

## Remote MCP Stdio Acceptance

Accepted by sending JSON-RPC messages to `patchxnote mcp serve` from the same Windows runtime/profile.

Results:

- `initialize`: passed, remote protocol `2026-07-28`.
- `tools/list`: passed, 19 tools.
- `tools/call patchxnote_get_current_user`: passed and returned a current account projection with masked phone only.

## Authorized Content And Model-IO Read Acceptance

Accepted by calling remote MCP tools through local stdio proxy and recording only counts/field names.

Results:

- `patchxnote_list_memories` on `mobile`: passed, 5 items in the first page.
- `patchxnote_list_model_io_traces` on `mobile`: passed, 5 items in the first page.
- `patchxnote_list_memories` on `desktop`: passed, 0 items for the current account.
- `patchxnote_list_model_io_traces` on `desktop`: passed, 0 items for the current account.
- `patchxnote_get_memory`: passed for a mobile item and returned safe metadata fields.
- `patchxnote_get_model_io_source_text`: passed with bounded content response.
- `patchxnote_get_model_io_provider_response`: passed with bounded content response.
- `patchxnote_get_model_io_parsed_result`: passed with bounded content response.
- `patchxnote_get_model_io_packaged_result`: passed with bounded content response.

The validation intentionally did not write raw source text, complete transcript, provider payload, parsed payload, packaged payload, or model response content into this document.

## npm Wrapper Local Candidate Acceptance

Accepted by building a `0.2.8` Windows binary and installing it through the npm wrapper with `--from-local` into an isolated temporary install directory.

Results:

- `node packages/npm/bin/patchxnote-agent.js install --from-local <candidate>`: passed.
- `node packages/npm/bin/patchxnote-agent.js mcp status --install-dir <isolated-dir> -- --profile mcp-oauth-windows-acceptance --server-base-url <test-server> --output json --verify`: passed.
- `node packages/npm/bin/patchxnote-agent.js mcp serve --install-dir <isolated-dir> -- --profile mcp-oauth-windows-acceptance --server-base-url <test-server>` with `initialize` and `tools/list`: passed, 19 tools.

This proves wrapper delegation and local candidate install behavior. It is not a substitute for npm registry acceptance after publication.

## Explicit Non-Accepted States

- Published npm package acceptance: not run yet, because `0.2.8` has not been published.
- VS Code/Cursor/Codex/Claude/Windsurf UI acceptance: not run yet; only the generic stdio protocol path was accepted.
- Feishu Aily / Doubao Work Partner / Tencent Agent Platform / enterprise WorkBuddy hosted platform acceptance: not run yet; platform console integration remains pending.
- WSL browser-open acceptance: not accepted. WSL can use `--no-browser`, but Windows desktop editors should run login/setup from Windows because Windows Credential Manager and WSL/Linux keychains are separate.

## Release Implication

Before publishing `0.2.8`, keep this local evidence as the pre-release baseline. After npm publication, repeat the browser login, `mcp status --verify`, `mcp config`, and stdio `mcp serve` smoke from a normal Windows directory using the registry package and the exact published version.
