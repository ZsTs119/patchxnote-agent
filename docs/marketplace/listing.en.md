# PatchXNote Agent Listing Copy

## Short Description

Connect local AI clients to PatchXNote MCP for safe access to summaries, memories, model results, and approved webhook workflows.

## Long Description

PatchXNote Agent helps trusted local AI clients connect to PatchXNote through a local stdio MCP server. It guides setup, browser OAuth, verification, and common PatchXNote workflows such as listing recent mobile or desktop memories, checking account and quota projections, inspecting model-result fields, creating Markdown drafts, and manually sending user-approved webhook messages.

PatchXNote server-backed data access is read-only and platform-scoped. The local webhook tools are explicit user-triggered exceptions for configuring local target aliases and sending approved content. PatchXNote Agent does not expose raw audio, full transcripts by default, hardware write operations, payments, Admin APIs, model execution, or background automatic webhook pushes.

## Tags

`mcp`, `agent-skills`, `productivity`, `notes`, `recordings`, `memory`, `summaries`, `patchxnote`

## Install

Agent Skill:

```sh
npx -y patchxnote-agent@latest skill install
```

MCP server setup:

```sh
npx -y patchxnote-agent@latest setup --client <client-id>
```

Generic stdio config:

```sh
npx -y patchxnote-agent@latest mcp config
```

## Verification

After setup, verify:

- `patchxnote_get_current_user`
- `patchxnote_list_memories` with `{"platform":"mobile","limit":5}`

When tool names or count matter, use live MCP tool discovery.
