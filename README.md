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

## Engineering Rules

Before changing CLI behavior, installer logic, MCP tools, authentication, local cache, or release configuration, read:

- [AGENTS.md](AGENTS.md)
- [docs/engineering-rules.md](docs/engineering-rules.md)
- [docs/plans/2026-08-06-agent-v1-mvp.md](docs/plans/2026-08-06-agent-v1-mvp.md)
