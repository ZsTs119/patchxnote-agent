# PatchXNote Agent

[English](./README.md) | [简体中文](./README.zh-CN.md)

[![npm version](https://img.shields.io/npm/v/patchnote-agent.svg)](https://www.npmjs.com/package/patchnote-agent)
[![GitHub release](https://img.shields.io/github/v/release/ZsTs119/patchnote-agent)](https://github.com/ZsTs119/patchnote-agent/releases)
[![Security policy](https://img.shields.io/badge/security-policy-blue.svg)](./SECURITY.md)

![PatchXNote Agent hero](./docs/assets/patchxnote-agent-cover.png)

PatchXNote Agent is the local CLI and MCP bridge for PatchXNote. It lets desktop AI agents read safe PatchXNote account context, including account status, bound recorder cards, quota, model usage, and structured-result metadata.

Agent V1 is deliberately read-only. It uses dedicated `/v1/agent/**` PatchXNote server APIs and does not expose App/PC hardware write flows, raw audio, full transcripts, SK, full MAC values, provider payloads, quota purchase flows, or Admin APIs.

```sh
npx -y patchnote-agent@0.1.3 install --print-config
```

## At A Glance

| Area | Agent V1 behavior |
| --- | --- |
| Runtime | Installs a versioned native `patchnote` binary through an npm wrapper. |
| Agent protocol | Runs a local stdio MCP server with `patchnote mcp serve`. |
| Login | Phone OTP login creates an independent Agent session, not a mobile/desktop installation. |
| Data access | Reads bounded account, recorder-card, quota, usage, and structured-result metadata projections. |
| Safety boundary | Read-only, masked, platform-scoped, and routed through dedicated Agent server endpoints. |
| Package status | Public beta `0.1.3`, defaulting to the PatchXNote test API. |

## Features

| Capability | Available in `0.1.3` | Notes |
| --- | --- | --- |
| Phone OTP Agent login | Yes | Uses Agent-specific server auth, not mobile/desktop installation slots. |
| Local MCP server | Yes | `patchnote mcp serve` over stdio. |
| Current account projection | Yes | Status, masked phone, registration platform, state version. |
| Recorder-card list | Yes | Read-only projection with masked identifiers. |
| Quota summary | Yes | Current account token balance summary. |
| Model usage summary | Yes | Current-month usage and charged quota summary. |
| Structured-result metadata | Yes | Platform-scoped `mobile` or `desktop` safe metadata. |
| Local memory search | Yes | Searches authorized metadata cached during the MCP session. |
| Hardware bind/release/recovery | No | Owned by App/PC and MR20 flows, not Agent V1. |
| Raw audio/transcripts/downloads | No | Intentionally not exposed. |
| Model execution | No | Agent V1 is read-only. |

## Requirements

- Node.js `18` or newer for the npm installer wrapper.
- Windows, macOS, or Linux on `amd64` or `arm64`.
- A PatchXNote account that can receive the phone OTP login code.
- An MCP host that supports stdio MCP servers, such as Codex, Claude Desktop, Cursor, VS Code, or another compatible desktop agent.

> `0.1.3` is a beta build. The default server is the PatchXNote test API. Credentials are stored in the OS-native keychain by default.

## Quickstart

![PatchXNote Agent quickstart](./docs/assets/patchxnote-agent-quickstart.png)

Install the npm wrapper. It downloads the matching `patchnote` binary from GitHub Releases, verifies `checksums.txt`, and installs it into a user-writable directory.

```sh
npx -y patchnote-agent@0.1.3 install --print-config
```

The installer prints:

- the installed binary path
- a PATH hint if `patchnote` is not already on your terminal PATH
- an MCP config snippet using the absolute binary path

The first beta build defaults to the PatchXNote test API:

```text
https://ws-lab.patch-x.cn/patchnote-test-api
```

Log in and check your session.

macOS/Linux:

```sh
patchnote login
patchnote auth status
```

Windows PowerShell:

```powershell
patchnote login
patchnote auth status
```

Start the MCP server:

```sh
patchnote mcp serve
```

To target a different PatchXNote environment:

```sh
PATCHNOTE_SERVER_BASE_URL=<PatchXNote API base URL> \
patchnote login
```

## MCP Configuration

![PatchXNote Agent architecture](./docs/assets/patchxnote-agent-architecture.png)

Use the `--print-config` output from the installer. A typical config looks like this:

```json
{
  "mcpServers": {
    "patchnote": {
      "command": "/absolute/path/to/patchnote",
      "args": ["mcp", "serve"]
    }
  }
}
```

MCP config never contains access tokens or refresh tokens. PatchXNote Agent stores credential material in macOS Keychain, Windows Credential Manager, or Linux Secret Service when available. The explicit `PATCHNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true` file store remains for local development and CI smoke only.

## MCP Tools

![PatchXNote Agent tools](./docs/assets/patchxnote-agent-tools.png)

| Tool | Purpose |
| --- | --- |
| `patchnote_get_current_user` | Read the current PatchXNote account projection. |
| `patchnote_list_recorder_cards` | List bound recorder cards with masked identifiers only. |
| `patchnote_get_quota_summary` | Read the current account quota summary. |
| `patchnote_get_model_usage_summary` | Read current-month model usage summary. |
| `patchnote_list_memories` | List safe structured-result metadata for one platform. |
| `patchnote_search_memories` | Search local authorized memory metadata cache. |
| `patchnote_get_memory` | Read safe metadata for one structured result. |

Memory tools require an explicit `platform` argument: `mobile` or `desktop`. V1 memory responses are safe metadata only; direct model-run response bodies and old summary text are not reconstructed by the Agent.

Example prompts for your desktop agent:

```text
Show my PatchXNote account and quota status.
List my PatchXNote recorder cards.
Search my desktop PatchXNote memories for roadmap.
```

## CLI Commands

```sh
patchnote version
patchnote login
patchnote auth status
patchnote logout
patchnote mcp serve
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
npx -y patchnote-agent@0.1.3 install
npx -y patchnote-agent@0.1.3 update
npx -y patchnote-agent@0.1.3 uninstall
```

## Security And Risk Notice

![PatchXNote Agent safety boundary](./docs/assets/patchxnote-agent-safety-boundary.png)

PatchXNote Agent gives an AI agent access to account metadata that belongs to the logged-in PatchXNote user. Treat the MCP host as trusted software and review any prompts, tool calls, or logs that may reveal private account context.

Default safety boundaries:

- Agent auth is separate from App/PC `mobile` and `desktop` installations.
- Agent calls only dedicated read-only `/v1/agent/**` server routes.
- MCP runs locally over stdio; stdout is reserved for JSON-RPC.
- MCP config does not store bearer tokens, refresh tokens, OTPs, SK, or full MAC values.
- Recorder-card identifiers are masked; live BLE state, battery, storage, and recording status are not exposed.
- Structured content is platform-scoped. The Agent does not merge mobile and desktop content.
- Tool outputs are bounded and validated before being returned to the MCP client.

Do not paste access tokens, refresh tokens, OTP codes, raw phone numbers, full MAC values, SK values, raw audio, transcripts, prompts, or provider payloads into public issues. Use the private process in [SECURITY.md](./SECURITY.md) for vulnerability reports.

## Current Limitations

`0.1.3` is a beta release.

- The default server points to the PatchXNote test API.
- Linux headless environments may not have Secret Service available; use the explicit development file-store fallback only for local smoke.
- Production Agent route rollout is pending.
- `patchnote_search_memories` searches only metadata cached during the current MCP session.
- Raw audio, full transcripts, complete model responses, hardware write actions, quota purchase/reward actions, payment, and Admin APIs are out of scope.

## Troubleshooting

| Problem | What to check |
| --- | --- |
| `patchnote` is not found after install | Add the printed install directory to PATH, then open a new terminal. |
| Login says credential storage is unavailable | Check that macOS Keychain, Windows Credential Manager, or Linux Secret Service is available and unlocked. For local development only, set `PATCHNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true`. |
| MCP host cannot start the server | Use the absolute `command` path printed by `--print-config`. |
| Memory list is empty | Check that you selected the correct `platform`: `mobile` or `desktop`. |
| Checksum verification fails | Retry later or pin a known version; the installer refuses unchecked binaries. |
| Wrong server environment | Set `PATCHNOTE_SERVER_BASE_URL=<PatchXNote API base URL>`. |

## Verify The Install

```sh
npm view patchnote-agent@0.1.3 version --registry https://registry.npmjs.org
npx -y --registry https://registry.npmjs.org patchnote-agent@0.1.3 install --dry-run --print-config
patchnote version
```

The release binary should report version `0.1.3` and the commit attached to the `v0.1.3` GitHub Release.

## Development

Local checks:

```sh
go test ./...
scripts/e2e/mvp-smoke.sh
node packages/npm/bin/patchnote-agent.js install --dry-run --print-config
```

The MVP smoke builds the CLI, runs installer dry-run, logs in against an in-process Agent V1 test server, checks `auth status`, starts `patchnote mcp serve`, calls all seven V1 MCP tools, logs out, and scans evidence for secret-like values.

Before changing CLI behavior, installer logic, MCP tools, authentication, local cache, or release configuration, read:

- [AGENTS.md](./AGENTS.md)
- [docs/engineering-rules.md](./docs/engineering-rules.md)
- [docs/plans/2026-08-06-agent-v1-mvp.md](./docs/plans/2026-08-06-agent-v1-mvp.md)

## Release Notes For Operators

1. Confirm the target PatchXNote GoServer exposes the required `/v1/agent/**` routes.
2. Confirm `packages/npm/package.json` version matches the release tag without the leading `v`.
3. Push a clean tag, for example `v0.1.3`.
4. Wait for GitHub Release assets: `checksums.txt` plus Linux/macOS/Windows amd64 and arm64 binaries.
5. Configure npm Trusted Publishing for this GitHub Actions workflow before npm publish:
   - owner/user: `ZsTs119`
   - repository: `patchnote-agent`
   - workflow filename: `publish-npm.yml`
   - allowed action: `npm publish`
6. Publish npm only after release assets exist and the trusted publisher is configured.
7. After a successful trusted publish, revoke the old npm automation token and disallow token-based publishing for this package.

## License

This repository is currently published without an open-source license. Contact PatchXNote before redistributing or embedding it in another product.
