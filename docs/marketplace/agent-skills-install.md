# PatchXNote MCP Skill Install Notes

Updated: 2026-09-04

PatchXNote MCP Skill is stored at:

```text
skills/patchxnote-mcp/
```

The skill teaches an AI assistant how to connect and verify PatchXNote MCP, then use PatchXNote summaries, memories, model results, and approved webhook workflows without losing the setup SOP in long or fresh sessions.

## Install With npm

The npm package bundles a copy of the canonical skill. This is the primary public install path because it uses the same package users already run for PatchXNote MCP setup:

```sh
npx -y patchxnote-agent@latest skill install
```

Pin a published release for troubleshooting or rollback:

```sh
npx -y patchxnote-agent@0.2.11 skill install
```

Install into an agent-specific local skill directory only after that client path has been verified:

```sh
npx -y patchxnote-agent@latest skill install --agent codex
npx -y patchxnote-agent@latest skill install --agent cursor
npx -y patchxnote-agent@latest skill install --agent claude-code
```

Dry-run the target paths before writing:

```sh
npx -y patchxnote-agent@latest skill install --dry-run --json
```

The installer refuses to replace an existing different skill folder unless `--force` is passed. Use `--force` only after confirming the target folder should be replaced.

## Skill Ecosystem Search

Keep `skills/patchxnote-mcp/` in the repository root as the canonical Agent Skills source so discovery tools can still index it from GitHub. Direct `npx skills add ...` remains a candidate ecosystem path, not the primary install command, because remote GitHub clone behavior can be slow or unavailable on some hosts.

When checking discoverability, record the command, date, and result in `docs/marketplace/evidence-log.md`:

```sh
npx -y skills find patchxnote
npx -y skills find patchxnote --owner ZsTs119
```

## Local Development Smoke

From this repository checkout:

```sh
node packages/npm/bin/patchxnote-agent.js skill install --dry-run --json
```

Then start a fresh supported AI session and ask:

```text
Help me connect PatchXNote MCP.
```

Expected behavior:

- The PatchXNote MCP skill activates.
- The AI identifies the local client/runtime.
- The AI runs or recommends `npx -y patchxnote-agent@latest setup --client <client-id>`.
- If the client ID is unclear, the AI runs `npx -y patchxnote-agent@latest mcp config`.
- The AI opens browser OAuth for the user and never asks for codes or tokens in chat.

## What This Does Not Install

This skill does not by itself authenticate PatchXNote or start the MCP server. Users still need:

```sh
npx -y patchxnote-agent@latest setup --client <client-id>
```

or the manual MCP config from:

```sh
npx -y patchxnote-agent@latest mcp config
```

Skill installation teaches the AI the process. MCP setup installs/configures the local server.
