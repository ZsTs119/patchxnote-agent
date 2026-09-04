# PatchXNote MCP Skill Install Notes

Updated: 2026-09-03

PatchXNote MCP Skill is stored at:

```text
skills/patchxnote-mcp/
```

The skill teaches an AI assistant how to connect and verify PatchXNote MCP, then use PatchXNote summaries, memories, model results, and approved webhook workflows without losing the setup SOP in long or fresh sessions.

## Install With `npx skills`

After a commit or tag containing `skills/patchxnote-mcp/` is pushed to GitHub:

```sh
npx skills add ZsTs119/patchxnote-agent --skill patchxnote-mcp -g
```

Project-level install:

```sh
npx skills add ZsTs119/patchxnote-agent --skill patchxnote-mcp
```

Install for a specific compatible agent after confirming that agent name with current `npx skills --help`:

```sh
npx skills add ZsTs119/patchxnote-agent --skill patchxnote-mcp --agent cursor
npx skills add ZsTs119/patchxnote-agent --skill patchxnote-mcp --agent claude-code
```

For public support docs, prefer a pinned tag once the first skill release is cut:

```sh
npx skills add ZsTs119/patchxnote-agent@v0.2.10 --skill patchxnote-mcp -g
```

Keep `@latest` or `main` examples for development docs only.

## Local Development Smoke

From this repository checkout:

```sh
npx skills add . --skill patchxnote-mcp --copy -y
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
