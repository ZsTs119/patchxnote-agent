# PatchXNote Agent

[English](./README.md) | [简体中文](./README.zh-CN.md)

[![npm version](https://img.shields.io/npm/v/patchxnote-agent.svg)](https://www.npmjs.com/package/patchxnote-agent)
[![GitHub release](https://img.shields.io/github/v/release/ZsTs119/patchxnote-agent)](https://github.com/ZsTs119/patchxnote-agent/releases)
[![Security policy](https://img.shields.io/badge/security-policy-blue.svg)](./SECURITY.md)

Official documentation: [Public Chinese user guide on Feishu](https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd)

Repository: [https://github.com/ZsTs119/patchxnote-agent](https://github.com/ZsTs119/patchxnote-agent)

![PatchXNote Agent hero](./docs/assets/patchxnote-agent-cover.png)

PatchXNote Agent is the local AI assistant connector for PatchXNote. After installing it, you can let an AI assistant find PatchXNote records, inspect AI-generated results, create Markdown, and send user-approved messages to Feishu, DingTalk, or another webhook.

It helps your AI assistant search records synced from mobile or desktop, inspect the source text and AI output behind a processing run, save editable local drafts, and manually send the final text to a named webhook target.

PatchXNote server data access remains read-only. Agent does not operate hardware binding, read raw audio, handle payments, or expose Admin APIs. Webhook configuration and sending happen locally, and messages are sent only when you or your AI assistant explicitly call a send command.

Give this one-line prompt to a local-command-capable AI assistant:

```text
Help me connect PatchXNote MCP and install the PatchXNote MCP Skill when the client supports Agent Skills. First, if this is a local command-capable environment, execute npx -y patchxnote-agent@latest skill install on this machine so future or long sessions remember the PatchXNote SOP; if skill installation is not supported, continue with MCP setup directly. Identify the MCP client you are running in; if this is a local MCP client such as Codex, Cursor, VS Code, Claude Desktop, Claude Code, or Windsurf, run npx -y patchxnote-agent@latest setup --client <client-id> on this machine in the same OS/runtime that will later launch the MCP server. If you are not sure which client ID to use, first run npx -y patchxnote-agent@latest mcp config and use the printed generic stdio config. For login, open the browser and let me complete PatchXNote phone-code authorization there; do not ask me to paste OTP codes, OAuth codes, authorization codes, access tokens, refresh tokens, or webhook secrets into chat. After setup, call patchxnote_get_current_user and patchxnote_list_memories with {"platform":"mobile","limit":5} to verify it works. References: https://github.com/ZsTs119/patchxnote-agent and https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd
```

For a supported client, start with setup:

```sh
npx -y patchxnote-agent@latest setup --client cursor
```

Or print the generic MCP config manually:

```sh
npx -y patchxnote-agent@latest mcp config
```

## PatchXNote MCP Skill

This repository also includes an Agent Skills package at `skills/patchxnote-mcp/`. It teaches compatible AI clients the PatchXNote setup, browser OAuth, runtime, verification, summary, memory, model-result, webhook, and evidence rules so the workflow survives fresh or long sessions.

`patchxnote-agent` on npm is the MCP installer/runtime package and also bundles the PatchXNote MCP Skill. Install or refresh the skill with:

```sh
npx -y patchxnote-agent@latest skill install
```

Use `--agent <id>` only after that client's local skills directory is verified; the default writes to the portable Agent Skills directory.

The skill does not authenticate PatchXNote or start MCP by itself. It complements the MCP setup command:

```sh
npx -y patchxnote-agent@latest setup --client <client-id>
```

For troubleshooting or rollback after `0.2.11` is published, pin the skill installer:

```sh
npx -y patchxnote-agent@0.2.11 skill install
```

## At A Glance

| Area | Agent V1 behavior |
| --- | --- |
| Runtime | Installs or verifies a versioned native `patchxnote` binary through an npm wrapper. |
| Browser MCP login | `patchxnote mcp login` opens the PatchXNote OAuth page, receives the loopback callback, and stores MCP connector credentials in the OS-native keychain. |
| Terminal CLI login | `patchxnote login` remains available for terminal-only use and the legacy local Agent path. |
| Local MCP service | `patchxnote mcp serve` runs the local stdio MCP server for editors and desktop agents. |
| Hosted MCP gateway | GoServer-hosted `/mcp` is the remote/platform path; platform-console acceptance is tracked separately. |
| Data access | Shows account, recorder cards, quota, records, and AI-generated results. |
| Webhook | Locally configures named targets and manually sends to Feishu, DingTalk, or another webhook. |
| Safety boundary | Server data is read-only; raw audio, hardware, payment, and Admin APIs are not exposed. |
| Package status | Current public beta release `0.2.11`, defaulting to the PatchXNote public beta API. |

## Features

| Capability | Available in `0.2.11` | Notes |
| --- | --- | --- |
| Browser OAuth MCP login | Yes | `patchxnote mcp login` opens the PatchXNote authorization page, completes phone OTP on the GoServer page, then stores MCP credentials locally. |
| Terminal phone OTP Agent login | Yes | `patchxnote login` keeps the terminal login path for CLI-first users and fallback environments. |
| Agent session refresh | Yes | Automatically rotates Agent access and refresh tokens from the local keychain. |
| Local MCP server | Yes | Lets MCP-capable AI assistants call PatchXNote Agent. |
| Hosted remote MCP gateway | Beta | Used for platform agents that cannot execute local `npx`; each platform still needs real console acceptance. |
| Local client setup wizard | Yes | `patchxnote setup --client <id>` plans, confirms, backs up, writes supported client config, and falls back to manual instructions where needed. |
| Account, recorder cards, quota | Yes | Shows account status, recorder-card list, quota, and current-month model usage. |
| Record list and search | Yes | Lists readable record entries by `mobile` or `desktop`, including saved results and model-generated outputs, then searches basics cached in the current MCP session. |
| Single record details | Yes | Shows safe basic information for one record. |
| AI processing lookup | Yes | Finds an AI processing run and the follow-up `request_id`. |
| Source text and AI result export | Yes | Explicitly inspect or export source text, AI response, parsed result, and final result. |
| Multiple webhook aliases | Yes | Configure multiple Feishu, DingTalk, or generic webhooks with custom names, including Chinese names and spaces. |
| Markdown drafts | Yes | Render a record into a local Markdown draft so the user can edit before sending. |
| Manual webhook send | Yes | Sends only when the user or AI explicitly invokes a send command; no background push. |
| Raw audio/audio download | No | Agent does not read raw audio and does not provide audio downloads. |
| Hardware bind/release/recovery | No | Owned by App/PC and MR20 flows. |
| Model execution/replay | No | Agent does not trigger new model runs or replay model calls. |
| Payment/purchase/Admin APIs | No | Quota purchase, payments, and Admin APIs are out of scope. |

## Requirements

- Node.js `18` or newer for the npm installer/launcher wrapper.
- Windows, macOS, or Linux on `amd64` or `arm64`.
- A PatchXNote account that can receive the phone OTP login code.
- An MCP host that supports stdio MCP servers, such as Codex, Claude Desktop, Cursor, VS Code, or another compatible desktop agent.

> `0.2.11` is the current public beta release. The default server is the PatchXNote public beta API. Credentials are stored in the OS-native keychain by default.

## Login And MCP Modes

PatchXNote Agent now has two productized login surfaces and two MCP service shapes:

| Mode | Entry | Use when | Notes |
| --- | --- | --- | --- |
| Browser MCP login | `npx -y patchxnote-agent@latest mcp login` | The user is installing PatchXNote into a desktop editor or local MCP host. | Opens a browser to the PatchXNote authorization page, completes phone OTP there, receives the loopback callback, exchanges the code, and stores credentials in the OS keychain. |
| Terminal CLI login | `npx -y patchxnote-agent@latest login` | The user prefers a terminal-only flow, is on a headless runtime, or needs the legacy local Agent fallback. | Keeps the phone OTP conversation inside the terminal. Do not paste OTP codes or tokens into an AI chat. |
| Local MCP service | `npx -y patchxnote-agent@latest mcp serve` | VS Code, Cursor, Codex, Claude Desktop, Windsurf, Trae, Qoder, WorkBuddy, and other local stdio MCP clients. | MCP config stays secret-free. `mcp serve` does not open a browser during editor startup. |
| Hosted remote MCP gateway | `https://ws-lab.patch-x.cn/patchnote-test-api/mcp` | Feishu Aily, Doubao Work Partner, Tencent Agent Development Platform, enterprise WorkBuddy, and other platform clients that cannot run local commands. | Server route is the platform-facing MCP shape; real platform-console acceptance is still tracked separately. |

## Quickstart

![PatchXNote Agent quickstart](./docs/assets/patchxnote-agent-quickstart.png)

Run setup for the client you use:

```sh
npx -y patchxnote-agent@latest setup --client vscode
npx -y patchxnote-agent@latest setup --client cursor
npx -y patchxnote-agent@latest setup --client codex
npx -y patchxnote-agent@latest setup --client workbuddy
```

Setup checks the MCP browser OAuth login, keeps credentials in the OS keychain, writes or prints the client MCP configuration, and creates a backup before modifying supported config files. This is the recommended path for local editor integrations.

You can also log in explicitly before adding the client:

```sh
npx -y patchxnote-agent@latest mcp login
npx -y patchxnote-agent@latest mcp status
```

You can still print the generic MCP config and paste it into any local stdio MCP-capable assistant or editor:

```sh
npx -y patchxnote-agent@latest mcp config
```

The generated config starts PatchXNote Agent through:

```sh
npx -y patchxnote-agent@latest mcp serve
```

On first start, the npm wrapper downloads the matching `patchxnote` binary from GitHub Releases, verifies `checksums.txt`, installs it into a user-writable directory, and then delegates to `patchxnote mcp serve`. MCP stdout stays reserved for JSON-RPC.

The terminal Agent login remains available for terminal-only users and legacy local CLI/MCP fallback:

```sh
npx -y patchxnote-agent@latest login
```

If an MCP host times out during the first binary download, run the stable fallback once and paste the printed absolute-path config:

```sh
npx -y patchxnote-agent@latest install --print-config
```

To pin the current published public beta version for troubleshooting or rollback:

```sh
npx -y patchxnote-agent@0.2.11 install --print-config
```

The public beta build defaults to the PatchXNote public beta API:

```text
https://ws-lab.patch-x.cn/patchnote-test-api
```

To target a different PatchXNote environment:

```sh
PATCHXNOTE_SERVER_BASE_URL=<PatchXNote API base URL> \
patchxnote mcp login
```

## Common Workflows

Ask your AI assistant:

```text
Find today's mobile records.
Show the AI result behind this record.
Send this Markdown to my Product Feishu webhook.
Export the AI response to a local JSON file.
Create a Markdown draft from this record so I can edit it before sending.
```

## Client Setup

Local setup supports these P0 client IDs:

```text
vscode, cursor, codex, claude-code, claude-desktop, windsurf, trae, qoder, workbuddy
```

`vscode`, `cursor`, `codex`, `claude-desktop`, and `windsurf` can write a local config file after confirmation. `claude-code`, `trae`, `qoder`, and `workbuddy` return manual commands or copyable config in V1. Platform clients such as Feishu Aily, Doubao Work Partner, Tencent Agent Development Platform, and enterprise WorkBuddy require the hosted remote MCP gateway and platform-console acceptance instead of local `npx`.

Useful setup flags:

```sh
patchxnote setup --client cursor --dry-run --print-config
patchxnote setup --client cursor --yes
patchxnote setup --client cursor --no-browser
patchxnote setup --all-local-supported --dry-run
patchxnote setup --client cursor --output json
```

Run setup in the same OS/runtime that will later launch MCP. For example, a Windows desktop editor needs Windows Credential Manager credentials, while a WSL or remote VS Code session needs credentials in that Linux runtime.

## MCP Configuration

For generic local stdio MCP hosts, use the pure JSON printed by:

```sh
npx -y patchxnote-agent@latest mcp config
```

The default config looks like this:

```json
{
  "mcpServers": {
    "patchxnote": {
      "command": "npx",
      "args": ["-y", "patchxnote-agent@latest", "mcp", "serve"]
    }
  }
}
```

Some clients may require a wrapper-specific field such as `type: "stdio"` or a different top-level key, but the `command` and `args` stay the same. If a client rejects `npx`, kills slow cold starts, or requires allowlisted absolute paths, use the fallback printed by:

```sh
npx -y patchxnote-agent@latest install --print-config
```

The fallback config uses the installed binary path:

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

MCP config never contains access tokens, refresh tokens, OTP codes, phone numbers, webhook secrets, or a base URL by default. PatchXNote Agent stores credential material in macOS Keychain, Windows Credential Manager, or Linux Secret Service when available. The explicit `PATCHXNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true` file store remains for local development and CI smoke only.

## MCP Tools

![PatchXNote Agent tools](./docs/assets/patchxnote-agent-tools.png)

PatchXNote Agent `0.2.11` exposes the same **19 local MCP tools** as the current public local server. End users can think of them as three groups; exact tool names are for MCP hosts and AI assistants.

### Account And Record Lookup

![PatchXNote Agent record lookup](./docs/assets/patchxnote-agent-records.png)

| Tool | Purpose |
| --- | --- |
| `patchxnote_get_current_user` | Show the current PatchXNote account status. |
| `patchxnote_list_recorder_cards` | List bound recorder cards with masked identifiers only. |
| `patchxnote_get_quota_summary` | Show current quota. |
| `patchxnote_get_model_usage_summary` | Show current-month AI usage and charged quota. |
| `patchxnote_list_memories` | List records for `mobile` or `desktop`. |
| `patchxnote_search_memories` | Search record basics cached in the current MCP session. |
| `patchxnote_get_memory` | Show safe basic information for one record. |

### Webhook Configuration And Sending

![PatchXNote Agent webhook delivery](./docs/assets/patchxnote-agent-webhook-delivery.png)

| Tool | Purpose |
| --- | --- |
| `patchxnote_list_webhook_targets` | List local webhook aliases and masked metadata. |
| `patchxnote_configure_webhook_target` | Create or update a webhook alias; URL and secret inputs are write-only. |
| `patchxnote_remove_webhook_target` | Remove a webhook alias and best-effort clean up stored secrets. |
| `patchxnote_list_webhook_templates` | List built-in Markdown templates. |
| `patchxnote_render_webhook_message` | Render a record into Markdown and optionally save a local draft. |
| `patchxnote_export_model_io` | Export a complete AI processing record to a user-chosen local file. |
| `patchxnote_send_webhook` | Manually send Markdown, a draft, a rendered record, or a test message to target aliases. |

### AI Result Inspection

![PatchXNote Agent AI result inspection](./docs/assets/patchxnote-agent-model-io.png)

| Tool | Purpose |
| --- | --- |
| `patchxnote_list_model_io_traces` | Find AI processing runs and the follow-up `request_id`. |
| `patchxnote_get_model_io_source_text` | Inspect or export the source text used for that run. |
| `patchxnote_get_model_io_provider_response` | Inspect or export the AI response. |
| `patchxnote_get_model_io_parsed_result` | Inspect or export the parsed AI result. |
| `patchxnote_get_model_io_packaged_result` | Inspect or export the final result. |

Record tools require an explicit `platform` argument: `mobile` or `desktop`. The record list now includes formal saved results plus readable model-generated outputs when the server has model IO data. `patchxnote model-io list` remains the lower-level AI processing list for finding request IDs and filtering by task or state.

Webhook MCP tools share the same local config, keychain, templates, and sender modules as the CLI. They do not return full webhook URLs or signing secrets, and send calls perform external network requests only when the MCP client explicitly invokes the send tool.

AI result tools are explicit inspection tools. They may expose source text or AI payloads for the logged-in user, so use them only from trusted local MCP hosts. Large fields should be written to an explicit local `out` file.

## CLI Commands

Browser MCP login and local MCP service:

```sh
patchxnote version
patchxnote mcp login
patchxnote mcp status
patchxnote mcp config
patchxnote setup --client cursor
patchxnote mcp serve
patchxnote mcp logout
```

Terminal CLI login:

```sh
patchxnote login
patchxnote auth status
patchxnote logout
```

List AI processing runs and export results:

```sh
patchxnote model-io list --platform mobile
patchxnote model-io source-text --request-id <request_id> --platform mobile --out ./source.txt
patchxnote model-io provider-response --request-id <request_id> --platform mobile --out ./provider-response.json
patchxnote model-io parsed-result --request-id <request_id> --platform mobile --out ./parsed-result.json
patchxnote model-io packaged-result --request-id <request_id> --platform mobile --out ./packaged-result.json
patchxnote model-io export --request-id <request_id> --platform mobile --out ./model-io.json
```

Get `request_id` from `patchxnote model-io list --platform mobile|desktop` when you need a lower-level AI processing run. MCP `patchxnote_list_memories` returns `id` and `platform` for record rendering, drafts, webhook workflows, and model IO field tools; for model-generated entries, that `id` can be the same value as `request_id`.

Configure and send webhooks:

```sh
patchxnote webhook set "Product Feishu" --type feishu --url-stdin
patchxnote webhook list
patchxnote webhook test "Product Feishu"
patchxnote webhook draft --memory-id <memory_id> --platform mobile --out ./patchxnote-drafts/example
patchxnote webhook send --target "Product Feishu" --file ./message.md
patchxnote webhook send --target "Product Feishu" --draft ./patchxnote-drafts/example
patchxnote webhook remove "Product Feishu"
```

Useful global flags:

```sh
--server-base-url <url>   PatchXNote API base URL
--profile <name>          local profile name
--output json             machine-readable output where supported
--config <path>           non-secret config file path
```

The npm package is a small installer/launcher wrapper:

```sh
npx -y patchxnote-agent@latest mcp login
npx -y patchxnote-agent@latest mcp status
npx -y patchxnote-agent@latest mcp config
npx -y patchxnote-agent@latest mcp serve
npx -y patchxnote-agent@latest mcp logout --local-only
npx -y patchxnote-agent@latest login
npx -y patchxnote-agent@latest setup --client cursor
npx -y patchxnote-agent@latest install
npx -y patchxnote-agent@latest update
npx -y patchxnote-agent@latest uninstall
```

Webhook URLs and optional Feishu/DingTalk signing secrets are stored in the local secure credential store, not in the non-secret config file. `--url-stdin` and `--secret-stdin` avoid shell history. CLI and MCP webhook sending is manual only, does not follow redirects, and surfaces provider errors directly.

`patchxnote model-io export` is the preferred complete AI processing export command. `patchxnote webhook export-model-io` remains available for compatibility.

## Security And Risk Notice

![PatchXNote Agent safety boundary](./docs/assets/patchxnote-agent-safety-boundary.png)

PatchXNote Agent gives an AI assistant access to account and record information for the logged-in PatchXNote user. Treat the MCP host as trusted software and review prompts, tool calls, local files, and logs that may contain account context, source text, or AI results.

Default safety boundaries:

- Agent auth is separate from App/PC `mobile` and `desktop` installations.
- Agent calls only dedicated read-only `/v1/agent/**` server routes for PatchXNote server data.
- MCP webhook tools can write local non-secret target metadata, write URL/secret material to local secure storage, and manually send external webhook HTTP requests.
- Webhook sending never happens in the background and is not scheduled automatically.
- MCP runs locally over stdio; stdout is reserved for JSON-RPC.
- MCP config does not store bearer tokens, refresh tokens, OTPs, SK, or full MAC values.
- Recorder-card identifiers are masked; live BLE state, battery, storage, and recording status are not exposed.
- Content is platform-scoped. The Agent does not merge mobile and desktop content.
- Tool outputs are bounded and validated before being returned to the MCP client.
- Webhook target URLs and signing secrets stay in local secure storage; normal webhook payloads never include access tokens, refresh tokens, or exported AI processing JSON.
- AI result tools return only the requested field. They do not replay model calls and do not include unrelated fields in single-field responses.
- Agent does not read raw audio and does not provide audio downloads. Source text and AI results are available only through explicit tool calls, and large/sensitive content should be exported to local files.

Do not paste access tokens, refresh tokens, OTP codes, raw phone numbers, full MAC values, SK values, raw audio, source text, prompts, or provider payloads into public issues. Use the private process in [SECURITY.md](./SECURITY.md) for vulnerability reports.

## Current Limitations

`0.2.11` is the current public beta release.

- The default server points to the PatchXNote public beta API and does not imply a production SLA.
- `mcp serve` never opens a browser during editor startup. Run `mcp login` first, or let `setup --client <id>` reuse that same OAuth flow.
- The GoServer website and authorization page own phone OTP input. The local Agent owns loopback callback, token exchange, secure storage, and stdio bridging.
- Remote platform MCP for Feishu Aily, Doubao Work Partner, Tencent Agent Development Platform, and enterprise WorkBuddy requires hosted platform-console acceptance; local `npx` setup only closes desktop/terminal clients.
- Linux headless environments may not have Secret Service available; use the explicit development file-store fallback only for local smoke.
- Public beta users should expect iterative improvements to setup guidance, MCP client examples, and webhook formatting.
- `patchxnote_search_memories` searches only record basics cached during the current MCP session.
- If the record list is empty, first check whether you selected `mobile` or `desktop`. Lower-level AI processing runs can also be listed separately with `patchxnote model-io list`.
- Raw audio, audio downloads, hardware write actions, model execution/replay, automatic webhook pushes, quota purchase/reward actions, payment, and Admin APIs are out of scope.

## Troubleshooting

| Problem | What to check |
| --- | --- |
| `patchxnote` is not found after install | Add the printed install directory to PATH, then open a new terminal. |
| Login says credential storage is unavailable | Check that macOS Keychain, Windows Credential Manager, or Linux Secret Service is available and unlocked. For local development only, set `PATCHXNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true`. |
| MCP login expired or points at the wrong server | Run `npx -y patchxnote-agent@latest mcp logout --local-only`, then run `npx -y patchxnote-agent@latest mcp login` again in the same runtime. |
| MCP host cannot start the server | If first start is slow or the client rejects `npx`, run `npx -y patchxnote-agent@latest install --print-config` once and use the printed absolute `command` path. |
| Setup writes credentials in the wrong place | Run setup from the same OS/runtime that will launch MCP. Windows desktop apps, WSL terminals, and VS Code Remote do not automatically share keychain credentials. |
| Need to undo setup | Restore the timestamped `.bak-YYYYMMDDTHHMMSSZ` file printed by setup, or remove only the `patchxnote` MCP server entry from the client config. |
| Record list is empty | Check that you selected the correct `platform`: `mobile` or `desktop`; use `model-io list` for lower-level AI processing runs. |
| Webhook did not send | Confirm the alias exists, the target is enabled, and check the provider error returned by the command. |
| Checksum verification fails | Retry later or pin a known version; the installer refuses unchecked binaries. |
| Wrong server environment | Set `PATCHXNOTE_SERVER_BASE_URL=<PatchXNote API base URL>`. |

## Verify The Install

```sh
npm view patchxnote-agent@latest version --registry https://registry.npmjs.org
npx -y --registry https://registry.npmjs.org patchxnote-agent@latest mcp config
npx -y --registry https://registry.npmjs.org patchxnote-agent@latest mcp status --output json
npx -y --registry https://registry.npmjs.org patchxnote-agent@latest mcp logout --local-only --output json
npx -y --registry https://registry.npmjs.org patchxnote-agent@latest setup --client cursor --dry-run --print-config
npx -y --registry https://registry.npmjs.org patchxnote-agent@latest install --dry-run --print-config
patchxnote version
```

The release binary should report the npm package version and the commit attached to the matching GitHub Release tag.

## 0.2.11 Highlights

- Bundles the canonical PatchXNote MCP Skill inside the npm package.
- Adds `npx -y patchxnote-agent@latest skill install` for npm-based skill installation without relying on a separate skills CLI or GitHub clone.
- Extends skill package sync and validation so OpenAI, Claude, and npm copies stay byte-identical to `skills/patchxnote-mcp/`.
- Updates the one-line setup prompt and discovery metadata to prefer the npm-bundled skill installer while preserving MCP setup, browser OAuth, and tool verification.

## 0.2.10 Highlights

- Adds the reusable PatchXNote MCP Skill at `skills/patchxnote-mcp/` so compatible AI clients can keep the setup and usage SOP across fresh or long sessions.
- Adds OpenAI/Codex, Claude Code, Agent Skills, MCP Registry, Smithery, and third-party directory draft packaging and listing materials.
- Adds MCP Registry metadata through `server.json` and `package.json#mcpName`, plus local validation and stdio smoke scripts for release evidence.
- Updates the one-line setup prompt so agents install the skill when supported, then run MCP setup, browser OAuth, and tool verification without asking users to paste codes or tokens into chat.

## 0.2.9 Highlights

- Polishes the browser OAuth loopback success and failure pages shown after `patchxnote mcp login`.
- Keeps the post-login result page focused on plain user guidance, without exposing OAuth codes, state values, or token-shaped details.
- Adds regression coverage for the browser callback pages so future changes keep those sensitive details out of the UI.

## 0.2.8 Highlights

- Adds `patchxnote setup --client <id>` and npm wrapper delegation with dry-run, JSON output, confirmation, config printing, force repair, and local MCP smoke hooks.
- Adds a client registry for VS Code, Cursor, Codex, Claude Code, Claude Desktop, Windsurf, Trae, Qoder, WorkBuddy, Feishu/Doubao, Tencent platform, and P1 follow-up clients.
- Adds JSON and TOML config merge adapters with backup, conflict detection, rollback, and manual JSONC mode.
- Adds `patchxnote mcp login/status/logout`, browser OAuth with PKCE, MCP OAuth secure storage, and remote `/mcp` stdio proxy mode with local fallback.
- Adds website page specs, detail-page copy, and remote platform gateway design for product-style onboarding.

## 0.2.6 Highlights

- MCP expands to 19 tools across account/record lookup, webhook delivery, and AI result inspection.
- Webhook workflows are available to MCP: configure named aliases and manually send to Feishu, DingTalk, or generic webhooks.
- AI processing runs can be listed first, then inspected by `request_id` for source text, AI response, parsed result, and final result.
- Record lists can now include readable model-generated outputs returned by the server, so users can find a record first and then inspect its source text, AI response, parsed result, or final result.
- Webhook aliases containing dots, Chinese text, or spaces now persist and reload correctly from the local config file.
- README, npm README, and public visual assets have been refreshed for the new user-facing positioning.

## Development

Local checks:

```sh
go test ./...
scripts/e2e/mvp-smoke.sh
node packages/npm/test/install.test.js
node docs/mcp-clients/validate-clients.mjs
```

The MVP smoke builds the CLI, runs installer dry-run, checks the npm universal MCP launcher plus `mcp login/status/logout` non-interactive paths, logs in against an in-process Agent V1 test server, checks `auth status`, starts `patchxnote mcp serve`, calls all 19 V1 MCP tools, exercises AI processing discovery, field export, and local webhook delivery, logs out, and scans evidence for secret-like values.

Before changing CLI behavior, installer logic, MCP tools, authentication, local cache, or release configuration, read:

- [AGENTS.md](./AGENTS.md)
- [docs/engineering-rules.md](./docs/engineering-rules.md)
- [docs/release-and-maintenance-runbook.zh-CN.md](./docs/release-and-maintenance-runbook.zh-CN.md)
- [docs/plans/2026-08-06-agent-v1-mvp.md](./docs/plans/2026-08-06-agent-v1-mvp.md)

## Release Notes For Operators

The detailed release and documentation maintenance checklist lives in [docs/release-and-maintenance-runbook.zh-CN.md](./docs/release-and-maintenance-runbook.zh-CN.md).

1. Confirm the target PatchXNote GoServer exposes the required `/v1/agent/**` routes.
2. Confirm `packages/npm/package.json` version matches the release tag without the leading `v`.
3. Push a clean tag, for example `v0.2.11`.
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
