# PatchNote Agent

[![npm version](https://img.shields.io/npm/v/patchnote-agent.svg)](https://www.npmjs.com/package/patchnote-agent)
[![GitHub release](https://img.shields.io/github/v/release/ZsTs119/patchnote-agent)](https://github.com/ZsTs119/patchnote-agent/releases)

PatchNote Agent is the local CLI and MCP bridge that lets desktop agents read safe PatchNote account context: account status, bound recorder cards, quota, model usage, and structured-result metadata.

It installs a versioned `patchnote` binary, runs a local stdio MCP server, and talks only to dedicated read-only `/v1/agent/**` PatchNote server APIs. It does not expose App/PC hardware write flows, raw audio, full transcripts, SK, full MAC, or provider payloads.

```sh
npx -y patchnote-agent@0.1.1 install --print-config
```

## What You Get

| Capability | Available in `0.1.1` | Notes |
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

## Quickstart

Install the npm wrapper. It downloads the matching `patchnote` binary from GitHub Releases, verifies `checksums.txt`, and installs it into a user-writable directory.

```sh
npx -y patchnote-agent@0.1.1 install --print-config
```

The installer prints:

- the installed binary path
- a PATH hint if `patchnote` is not already on your terminal PATH
- an MCP config snippet using the absolute binary path

The first beta build defaults to the PatchNote test API:

```text
https://ws-lab.patch-x.cn/patchnote-test-api
```

Log in and check your session:

```sh
PATCHNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true patchnote login
patchnote auth status
```

Start the MCP server:

```sh
patchnote mcp serve
```

To target a different PatchNote environment:

```sh
PATCHNOTE_SERVER_BASE_URL=<PatchNote API base URL> \
PATCHNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true \
patchnote login
```

## MCP Configuration

Use the `--print-config` output from the installer. A typical config looks like this:

```json
{
  "mcpServers": {
    "patchnote": {
      "command": "/absolute/path/to/patchnote",
      "args": ["mcp", "serve"],
      "env": {
        "PATCHNOTE_AUTH_INSECURE_FILE_KEYCHAIN": "true"
      }
    }
  }
}
```

MCP config never contains access tokens or refresh tokens. The beta `PATCHNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true` setting is explicit because OS keychain adapters are not shipped yet.

## MCP Tools

| Tool | Purpose |
| --- | --- |
| `patchnote_get_current_user` | Read the current PatchNote account projection. |
| `patchnote_list_recorder_cards` | List bound recorder cards with masked identifiers only. |
| `patchnote_get_quota_summary` | Read the current account quota summary. |
| `patchnote_get_model_usage_summary` | Read current-month model usage summary. |
| `patchnote_list_memories` | List safe structured-result metadata for one platform. |
| `patchnote_search_memories` | Search local authorized memory metadata cache. |
| `patchnote_get_memory` | Read safe metadata for one structured result. |

Memory tools require an explicit `platform` argument: `mobile` or `desktop`. V1 memory responses are safe metadata only; direct model-run response bodies and old summary text are not reconstructed by the Agent.

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
--server-base-url <url>   PatchNote API base URL
--profile <name>          local profile name
--output json             machine-readable output where supported
--config <path>           non-secret config file path
```

The npm package itself is only an installer/update wrapper:

```sh
npx -y patchnote-agent@0.1.1 install
npx -y patchnote-agent@0.1.1 update
npx -y patchnote-agent@0.1.1 uninstall
```

## Security Model

- Agent auth is separate from App/PC `mobile` and `desktop` installations.
- Agent calls only dedicated read-only `/v1/agent/**` server routes.
- MCP runs locally over stdio; stdout is reserved for JSON-RPC.
- MCP config does not store bearer tokens, refresh tokens, OTPs, SK, or full MAC values.
- Recorder-card identifiers are masked; live BLE state, battery, storage, and recording status are not exposed.
- Structured content is platform-scoped. The Agent does not merge mobile and desktop content.
- Tool outputs are bounded and validated before being returned to the MCP client.

## Current Limitations

`0.1.1` is a first beta release.

- The default server points to the PatchNote test API.
- Credential storage uses an explicit beta file-store opt-in: `PATCHNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true`.
- OS-native keychain adapters are pending.
- macOS execution smoke is pending.
- Production Agent route rollout is pending.
- `patchnote_search_memories` searches only metadata cached during the current MCP session.
- Raw audio, full transcripts, complete model responses, hardware write actions, quota purchase/reward actions, payment, and Admin APIs are out of scope.

## Verify The Install

```sh
npm view patchnote-agent@0.1.1 version --registry https://registry.npmjs.org
npx -y --registry https://registry.npmjs.org patchnote-agent@0.1.1 install --dry-run --print-config
patchnote version
```

The release binary should report version `0.1.1` and commit `8c82973d690b7ca58b79ddbab7d57e5a2a82f470`.

## Development

Local checks:

```sh
go test ./...
scripts/e2e/mvp-smoke.sh
node packages/npm/bin/patchnote-agent.js install --dry-run --print-config
```

The MVP smoke builds the CLI, runs installer dry-run, logs in against an in-process Agent V1 test server, checks `auth status`, starts `patchnote mcp serve`, calls all seven V1 MCP tools, logs out, and scans evidence for secret-like values.

Before changing CLI behavior, installer logic, MCP tools, authentication, local cache, or release configuration, read:

- [AGENTS.md](AGENTS.md)
- [docs/engineering-rules.md](docs/engineering-rules.md)
- [docs/plans/2026-08-06-agent-v1-mvp.md](docs/plans/2026-08-06-agent-v1-mvp.md)

## Release Notes For Operators

1. Confirm the target PatchNote GoServer exposes the required `/v1/agent/**` routes.
2. Confirm `packages/npm/package.json` version matches the release tag without the leading `v`.
3. Push a clean tag, for example `v0.1.1`.
4. Wait for GitHub Release assets: `checksums.txt` plus Linux/macOS/Windows amd64 and arm64 binaries.
5. Configure npm Trusted Publishing for this GitHub Actions workflow before npm publish:
   - owner/user: `ZsTs119`
   - repository: `patchnote-agent`
   - workflow filename: `publish-npm.yml`
   - allowed action: `npm publish`
6. Publish npm only after release assets exist and the trusted publisher is configured.
7. After a successful trusted publish, revoke the old npm automation token and disallow token-based publishing for this package.

Security reports should use the private process in [SECURITY.md](SECURITY.md).

## License

This repository is currently published without an open-source license. Contact PatchNote before redistributing or embedding it in another product.
