# PatchXNote Skill And Marketplace Evidence Log

Updated: 2026-09-04

This log tracks repository-side skill and marketplace distribution work. It does not claim public marketplace acceptance unless a channel has explicit evidence.

| Channel | Artifact/version | Status | Evidence | Notes |
| --- | --- | --- | --- | --- |
| GitHub Release | `v0.2.10` | `published_smoked` | Release workflow `33846214903` passed and assets were verified on 2026-09-04 | Evidence: `docs/evidence/2026-09-04-release-0.2.10.zh-CN.md`. |
| npm package | `patchxnote-agent@0.2.10` | `published_smoked` | npm workflow `33846367389` passed; npm registry reports latest `0.2.10` on 2026-09-04 | Windows install smoke and MCP smoke passed. |
| Canonical Agent Skill | `skills/patchxnote-mcp` `0.1.0` | `locally_smoked` | `quick_validate.py skills/patchxnote-mcp` passed; GitHub API confirmed the skill file at `v0.2.10` on 2026-09-04 | Full remote `npx skills add ... --list` was interrupted after prolonged clone slowness on this host. |
| OpenAI/Codex plugin | `packages/plugins/openai/patchxnote-agent` `0.1.0` | `locally_smoked` | `validate_plugin.py packages/plugins/openai/patchxnote-agent` passed on 2026-09-03 | Skills-only; no public submission yet. |
| Claude Code plugin | `packages/plugins/claude/patchxnote-agent` `0.1.0` | `drafted` | Local manifest and marketplace | `claude --version` was unavailable on this host on 2026-09-03, so install smoke is pending in a Claude Code environment. |
| `npx skills` | GitHub repo source | `locally_smoked` | Local-path `npx -y skills add . --skill patchxnote-mcp --list` found `patchxnote-mcp` on 2026-09-03; GitHub file visibility confirmed on 2026-09-04 | Full remote clone smoke remains pending because this host's GitHub clone was too slow. |
| Local stdio MCP smoke | `patchxnote-agent@latest` / `0.2.10` | `published_smoked` | `node scripts/smoke-mcp-stdio.mjs` passed on 2026-09-04: `tools/list` returned 19 tools and `patchxnote_get_current_user` / `patchxnote_list_memories` returned `isError:false` | Script output intentionally omits account details and memory content. |
| Official MCP Registry | `server.json`, `mcpName` | `drafted` | Metadata files exist and npm `patchxnote-agent@0.2.10` exposes `mcpName=io.github.zsts119/patchxnote-agent` | Official `mcp-publisher` binary was unavailable on this host. The npm package named `mcp-publisher@0.4.2` starts a stdio service and was not counted as registry validation. |
| Smithery | `smithery.yaml` | `drafted` | Draft stdio config only | Hosted URL/MCPB decision pending. |
| Glama / PulseMCP / MCP.so | listing docs | `drafted` | Listing copy exists | No external submission yet. |
| officialskills.sh / awesome-agent-skills | listing docs | `drafted` | Listing copy exists | No external submission yet. |
| Cursor Skills | install notes | `drafted` | Docs exist | Clean Cursor acceptance pending. |
| Current Codex-mounted PatchXNote MCP tools | Connected runtime | `blocked` | Tool discovery returned 7 tools and direct tool calls returned `invalid_params` on 2026-09-03 | This is separate from the npm latest stdio smoke, which passed. Reconnect/reload the MCP entry before using this mounted tool set as acceptance evidence. |

## Evidence Rules

- Record tool names, counts, versions, pass/fail, and masked projections only.
- Do not record OTP, OAuth codes, tokens, webhook secrets, raw phone numbers, raw content, full MAC, SK, prompts, or provider payloads.
- Do not upgrade a status to `accepted` or `indexed` until the platform or directory proves it.
