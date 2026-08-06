# PatchNote Agent

PatchNote Agent is the local CLI and MCP bridge for exposing safe PatchNote account tools to desktop agents.

Planned user entry:

```sh
npx -y @patchnote/agent install
```

The npm package is only an installer wrapper. The long-lived local runtime is a versioned `patchnote` binary that provides:

- `patchnote setup` for first-run login and agent configuration.
- `patchnote login` for phone OTP login.
- `patchnote mcp` for the local stdio MCP server.

Initial MCP scope is read-only:

- current account and profile projection
- bound recorder card list
- quota summary
- structured result listing and detail lookup

The server-side PatchNote API remains the source of truth. This repository owns local distribution, credential storage, MCP tool schema, and desktop agent integration.

## Current MVP Smoke

Local MVP smoke command:

```sh
scripts/e2e/mvp-smoke.sh
```

The smoke builds the `patchnote` binary, runs the npm installer wrapper in dry-run mode, logs in against an in-process Agent V1 test server, checks `auth status`, starts `patchnote mcp serve`, calls the seven V1 MCP tools, logs out, and scans smoke evidence for secret-like values.

Useful local commands:

```sh
go test ./...
go run ./cmd/patchnote version
go run ./cmd/patchnote auth status
go run ./cmd/patchnote mcp serve
node packages/npm/bin/patchnote-agent.js install --dry-run --print-config
```

V1 limitations:

- Agent access is read-only and uses dedicated `/v1/agent/...` server routes.
- Recorder-card battery, live BLE state, storage, recording status, SK, full MAC, raw audio, and full transcripts are not exposed.
- `patchnote_search_memories` searches only local authorized metadata cache populated during the MCP session.
- Public npm release, signed binaries, real OS keychain adapters, and cross-machine install validation remain Phase 12 release gates.

## Engineering Rules

Before changing CLI behavior, installer logic, MCP tools, authentication, local cache, or release configuration, read:

- [AGENTS.md](AGENTS.md)
- [docs/engineering-rules.md](docs/engineering-rules.md)
- [docs/plans/2026-08-06-agent-v1-mvp.md](docs/plans/2026-08-06-agent-v1-mvp.md)
