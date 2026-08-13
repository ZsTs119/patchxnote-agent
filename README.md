# PatchXNote Agent

[English](./README.md) | [简体中文](./README.zh-CN.md)

[![npm version](https://img.shields.io/npm/v/patchxnote-agent.svg)](https://www.npmjs.com/package/patchxnote-agent)
[![GitHub release](https://img.shields.io/github/v/release/ZsTs119/patchxnote-agent)](https://github.com/ZsTs119/patchxnote-agent/releases)
[![Security policy](https://img.shields.io/badge/security-policy-blue.svg)](./SECURITY.md)

Official documentation: [Public Chinese user guide on Feishu](https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd)

Repository: [https://github.com/ZsTs119/patchxnote-agent](https://github.com/ZsTs119/patchxnote-agent)

![PatchXNote Agent hero](./docs/assets/patchxnote-agent-cover.png)

PatchXNote Agent is the local CLI and MCP bridge for PatchXNote. It lets desktop AI agents read safe PatchXNote account context, including account status, bound recorder cards, quota, model usage, and structured-result metadata. It can also manually deliver user-approved Markdown to locally configured Feishu, DingTalk, or generic webhook targets.

Server-backed Agent data access is deliberately read-only. PatchXNote Agent uses dedicated `/v1/agent/**` server APIs and does not expose App/PC hardware write flows, raw audio, full transcripts, SK, full MAC values, quota purchase flows, or Admin APIs. Webhook tools are local configuration writes plus manual external sends; webhook URLs and signing secrets are write-only inputs and are never listed back.

Give this one-line prompt to a local-command-capable AI assistant:

```text
Install and connect PatchXNote Agent by following the public guide at https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd: run npx -y patchxnote-agent install --print-config, guide me through patchxnote login, and add the printed MCP JSON config to this AI assistant. Do not ask me to paste OTP codes, access tokens, or refresh tokens into chat. GitHub repository: https://github.com/ZsTs119/patchxnote-agent
```

Or run the install command manually:

```sh
npx -y patchxnote-agent install --print-config
```

## At A Glance

| Area | Agent V1 behavior |
| --- | --- |
| Runtime | Installs a versioned native `patchxnote` binary through an npm wrapper. |
| Agent protocol | Runs a local stdio MCP server with `patchxnote mcp serve`. |
| Login | Phone OTP login creates an independent Agent session, not a mobile/desktop installation. |
| Data access | Reads bounded account, recorder-card, quota, usage, and structured-result metadata projections. |
| Safety boundary | Server data is read-only, masked, platform-scoped, and routed through dedicated Agent endpoints; webhook sends are local, manual side effects. |
| Package status | Public beta `0.2.3`, defaulting to the PatchXNote public beta API. |

## Features

| Capability | Available in `0.2.3` | Notes |
| --- | --- | --- |
| Phone OTP Agent login | Yes | Uses Agent-specific server auth, not mobile/desktop installation slots. |
| Agent session refresh | Yes | Automatically rotates Agent access and refresh tokens from the local keychain. |
| Local MCP server | Yes | `patchxnote mcp serve` over stdio. |
| Current account projection | Yes | Status, masked phone, registration platform, state version. |
| Recorder-card list | Yes | Read-only projection with masked identifiers. |
| Quota summary | Yes | Current account token balance summary. |
| Model usage summary | Yes | Current-month usage and charged quota summary. |
| Structured-result metadata | Yes | Platform-scoped `mobile` or `desktop` safe metadata. |
| Local memory search | Yes | Searches authorized metadata cached during the MCP session. |
| Local webhook delivery | Yes | Configure named Feishu, DingTalk, or generic webhook targets and manually send editable Markdown. |
| Memory-backed webhook drafts | Yes | Fetches the Agent delivery-document projection, saves editable drafts, and can explicitly export model IO JSON. |
| Model IO field inspection | Yes | Explicitly inspect source text, provider response, parsed result, or packaged result by memory or request ID. |
| Hardware bind/release/recovery | No | Owned by App/PC and MR20 flows, not Agent V1. |
| Raw audio/transcripts/downloads | No | Intentionally not exposed. |
| Model execution | No | Server-backed Agent data access remains read-only. |

## Requirements

- Node.js `18` or newer for the npm installer wrapper.
- Windows, macOS, or Linux on `amd64` or `arm64`.
- A PatchXNote account that can receive the phone OTP login code.
- An MCP host that supports stdio MCP servers, such as Codex, Claude Desktop, Cursor, VS Code, or another compatible desktop agent.

> `0.2.3` is a public beta build. The default server is the PatchXNote public beta API. Credentials are stored in the OS-native keychain by default.

## Quickstart

![PatchXNote Agent quickstart](./docs/assets/patchxnote-agent-quickstart.png)

Install the npm wrapper. It downloads the matching `patchxnote` binary from GitHub Releases, verifies `checksums.txt`, and installs it into a user-writable directory.

```sh
npx -y patchxnote-agent install --print-config
```

The installer prints:

- the installed binary path
- a PATH hint if `patchxnote` is not already on your terminal PATH
- an MCP config snippet using the absolute binary path

To pin the current public beta version for troubleshooting or rollback:

```sh
npx -y patchxnote-agent@0.2.3 install --print-config
```

The public beta build defaults to the PatchXNote public beta API:

```text
https://ws-lab.patch-x.cn/patchnote-test-api
```

Log in and check your session.

macOS/Linux:

```sh
patchxnote login
patchxnote auth status
```

Windows PowerShell:

```powershell
patchxnote login
patchxnote auth status
```

Start the MCP server:

```sh
patchxnote mcp serve
```

To target a different PatchXNote environment:

```sh
PATCHXNOTE_SERVER_BASE_URL=<PatchXNote API base URL> \
patchxnote login
```

## MCP Configuration

![PatchXNote Agent architecture](./docs/assets/patchxnote-agent-architecture.png)

Use the `--print-config` output from the installer. A typical config looks like this:

```json
{
  "mcpServers": {
    "patchxnote": {
      "command": "/absolute/path/to/patchxnote",
      "args": ["mcp", "serve"]
    }
  }
}
```

MCP config never contains access tokens or refresh tokens. PatchXNote Agent stores credential material in macOS Keychain, Windows Credential Manager, or Linux Secret Service when available. The explicit `PATCHXNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true` file store remains for local development and CI smoke only.

## MCP Tools

![PatchXNote Agent tools](./docs/assets/patchxnote-agent-tools.png)

| Tool | Purpose |
| --- | --- |
| `patchxnote_get_current_user` | Read the current PatchXNote account projection. |
| `patchxnote_list_recorder_cards` | List bound recorder cards with masked identifiers only. |
| `patchxnote_get_quota_summary` | Read the current account quota summary. |
| `patchxnote_get_model_usage_summary` | Read current-month model usage summary. |
| `patchxnote_list_memories` | List safe structured-result metadata for one platform. |
| `patchxnote_search_memories` | Search local authorized memory metadata cache. |
| `patchxnote_get_memory` | Read safe metadata for one structured result. |
| `patchxnote_list_webhook_targets` | List local webhook target aliases with masked metadata only. |
| `patchxnote_configure_webhook_target` | Create or update a local webhook target; URL/secret inputs are write-only. |
| `patchxnote_remove_webhook_target` | Remove a local webhook target and best-effort clean up stored secrets. |
| `patchxnote_list_webhook_templates` | List built-in webhook Markdown templates. |
| `patchxnote_render_webhook_message` | Render a delivery-document projection into Markdown and optionally save a draft. |
| `patchxnote_export_model_io` | Export explicit model IO JSON to a user-chosen local file. |
| `patchxnote_send_webhook` | Manually send Markdown, a draft, a memory render, or a test message to target aliases. |
| `patchxnote_list_model_io_traces` | List model IO trace metadata and request IDs for one platform. |
| `patchxnote_get_model_io_source_text` | Read the explicit source text/safe transcript projection field. |
| `patchxnote_get_model_io_provider_response` | Read only the model provider response JSON field. |
| `patchxnote_get_model_io_parsed_result` | Read only the parsed model result JSON field. |
| `patchxnote_get_model_io_packaged_result` | Read only the packaged structured result JSON field. |

Memory tools require an explicit `platform` argument: `mobile` or `desktop`. V1 memory responses are safe metadata only; direct model-run response bodies and old summary text are not reconstructed by the Agent.
Webhook MCP tools share the same local config, keychain, templates, and sender modules as the CLI. They do not return full webhook URLs or signing secrets, and send calls perform external network requests only when the MCP client explicitly invokes the send tool.
Model IO field tools are explicit and field-scoped. They may expose source text or provider/model payloads for the logged-in user, so use them only from trusted local MCP hosts. Large fields should be written to an explicit local `out` file.

Example prompts for your desktop agent:

```text
Show my PatchXNote account and quota status.
List my PatchXNote recorder cards.
Search my desktop PatchXNote memories for roadmap.
Configure a Feishu webhook target named Product Feishu, then send this Markdown summary to it.
List my mobile PatchXNote model IO traces from today and use the request ID to inspect the provider response.
Show the provider response for this PatchXNote memory and save the parsed result to a local JSON file.
List my configured PatchXNote webhook targets.
Remove the PatchXNote webhook target named Product Feishu.
```

## CLI Commands

```sh
patchxnote version
patchxnote login
patchxnote auth status
patchxnote logout
patchxnote mcp serve
patchxnote webhook set "Product Feishu" --type feishu --url-stdin
patchxnote webhook test "Product Feishu"
patchxnote webhook send --target "Product Feishu" --file ./message.md
patchxnote webhook draft --memory-id <memory_id> --out ./patchxnote-drafts/example
patchxnote webhook send --target "Product Feishu" --draft ./patchxnote-drafts/example
patchxnote webhook export-model-io --memory-id <memory_id> --out ./patchxnote-drafts/example/model-io.json
patchxnote model-io list --platform mobile
patchxnote model-io source-text --memory-id <memory_id> --platform mobile
patchxnote model-io provider-response --memory-id <memory_id> --platform mobile --out ./provider-response.json
patchxnote model-io parsed-result --memory-id <memory_id> --platform mobile --out ./parsed-result.json
patchxnote model-io packaged-result --request-id <request_id> --platform mobile
patchxnote model-io export --memory-id <memory_id> --platform mobile --out ./model-io.json
```

Useful global flags:

```sh
--server-base-url <url>   PatchXNote API base URL
--profile <name>          local profile name
--output json             machine-readable output where supported
--config <path>           non-secret config file path
```

The npm package itself is only an installer/update wrapper:

```sh
npx -y patchxnote-agent@0.2.3 install
npx -y patchxnote-agent@0.2.3 update
npx -y patchxnote-agent@0.2.3 uninstall
```

Webhook URLs and optional Feishu/DingTalk signing secrets are stored in the local secure credential store, not in the non-secret config file. `--url-stdin` and `--secret-stdin` avoid shell history. CLI and MCP webhook sending is manual only, does not follow redirects, and surfaces provider errors directly.
`patchxnote model-io export` is the preferred complete model IO export command. `patchxnote webhook export-model-io` remains available for compatibility.

## Security And Risk Notice

![PatchXNote Agent safety boundary](./docs/assets/patchxnote-agent-safety-boundary.png)

PatchXNote Agent gives an AI agent access to account metadata that belongs to the logged-in PatchXNote user. Treat the MCP host as trusted software and review any prompts, tool calls, or logs that may reveal private account context.

Default safety boundaries:

- Agent auth is separate from App/PC `mobile` and `desktop` installations.
- Agent calls only dedicated read-only `/v1/agent/**` server routes for PatchXNote server data.
- MCP webhook tools can write local non-secret target metadata, write URL/secret material to local secure storage, and manually send external webhook HTTP requests.
- MCP runs locally over stdio; stdout is reserved for JSON-RPC.
- MCP config does not store bearer tokens, refresh tokens, OTPs, SK, or full MAC values.
- Recorder-card identifiers are masked; live BLE state, battery, storage, and recording status are not exposed.
- Structured content is platform-scoped. The Agent does not merge mobile and desktop content.
- Tool outputs are bounded and validated before being returned to the MCP client.
- Webhook target URLs and signing secrets stay in local secure storage; normal webhook payloads never include access tokens, refresh tokens, or exported model IO JSON.
- Model IO field tools return only the requested field. They do not replay model calls and do not include unrelated model IO fields in single-field responses.

Do not paste access tokens, refresh tokens, OTP codes, raw phone numbers, full MAC values, SK values, raw audio, transcripts, prompts, or provider payloads into public issues. Use the private process in [SECURITY.md](./SECURITY.md) for vulnerability reports.

## Current Limitations

`0.2.3` is a beta release.

- The default server points to the PatchXNote public beta API.
- Linux headless environments may not have Secret Service available; use the explicit development file-store fallback only for local smoke.
- Public beta users should expect iterative improvements to setup guidance, MCP client examples, and webhook formatting.
- `patchxnote_search_memories` searches only metadata cached during the current MCP session.
- Raw audio, full transcripts, hardware write actions, model execution/replay, quota purchase/reward actions, payment, and Admin APIs are out of scope.

## Troubleshooting

| Problem | What to check |
| --- | --- |
| `patchxnote` is not found after install | Add the printed install directory to PATH, then open a new terminal. |
| Login says credential storage is unavailable | Check that macOS Keychain, Windows Credential Manager, or Linux Secret Service is available and unlocked. For local development only, set `PATCHXNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true`. |
| MCP host cannot start the server | Use the absolute `command` path printed by `--print-config`. |
| Memory list is empty | Check that you selected the correct `platform`: `mobile` or `desktop`. |
| Checksum verification fails | Retry later or pin a known version; the installer refuses unchecked binaries. |
| Wrong server environment | Set `PATCHXNOTE_SERVER_BASE_URL=<PatchXNote API base URL>`. |

## Verify The Install

```sh
npm view patchxnote-agent@0.2.3 version --registry https://registry.npmjs.org
npx -y --registry https://registry.npmjs.org patchxnote-agent@0.2.3 install --dry-run --print-config
patchxnote version
```

The release binary should report version `0.2.3` and the commit attached to the `v0.2.3` GitHub Release.

## Development

Local checks:

```sh
go test ./...
scripts/e2e/mvp-smoke.sh
node packages/npm/bin/patchxnote-agent.js install --dry-run --print-config
```

The MVP smoke builds the CLI, runs installer dry-run, logs in against an in-process Agent V1 test server, checks `auth status`, starts `patchxnote mcp serve`, calls all 19 V1 MCP tools, exercises model IO discovery and field tools plus local webhook delivery, logs out, and scans evidence for secret-like values.

Before changing CLI behavior, installer logic, MCP tools, authentication, local cache, or release configuration, read:

- [AGENTS.md](./AGENTS.md)
- [docs/engineering-rules.md](./docs/engineering-rules.md)
- [docs/release-and-maintenance-runbook.zh-CN.md](./docs/release-and-maintenance-runbook.zh-CN.md)
- [docs/plans/2026-08-06-agent-v1-mvp.md](./docs/plans/2026-08-06-agent-v1-mvp.md)

## Release Notes For Operators

The detailed release and documentation maintenance checklist lives in [docs/release-and-maintenance-runbook.zh-CN.md](./docs/release-and-maintenance-runbook.zh-CN.md).

1. Confirm the target PatchXNote GoServer exposes the required `/v1/agent/**` routes.
2. Confirm `packages/npm/package.json` version matches the release tag without the leading `v`.
3. Push a clean tag, for example `v0.2.3`.
4. Wait for GitHub Release assets: `checksums.txt` plus Linux/macOS/Windows amd64 and arm64 binaries.
5. Configure npm Trusted Publishing for this GitHub Actions workflow before npm publish:
   - owner/user: `ZsTs119`
   - repository: `patchxnote-agent`
   - workflow filename: `publish-npm.yml`
   - allowed action: `npm publish`
6. Publish npm only after release assets exist and the trusted publisher is configured.
7. After a successful trusted publish, revoke the old npm automation token and disallow token-based publishing for this package.

## License

This repository is currently published without an open-source license. Contact PatchXNote before redistributing or embedding it in another product.
