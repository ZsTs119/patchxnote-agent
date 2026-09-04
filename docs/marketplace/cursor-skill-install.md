# Cursor Skill Install Notes

Updated: 2026-09-03

Cursor documents Skills as portable Agent Skills packages that can live in a repository or be installed from GitHub links. PatchXNote MCP Skill should remain standard Agent Skills format:

```text
skills/patchxnote-mcp/SKILL.md
skills/patchxnote-mcp/references/
```

## Install Candidate

After the skill is pushed to GitHub:

```sh
npx skills add ZsTs119/patchxnote-agent --skill patchxnote-mcp --agent cursor
```

Fallback global install:

```sh
npx skills add ZsTs119/patchxnote-agent --skill patchxnote-mcp -g
```

## Cursor MCP Setup Still Required

Skill installation only teaches Cursor's agent how to perform the PatchXNote flow. MCP setup still requires:

```sh
npx -y patchxnote-agent@latest setup --client cursor
```

If Cursor rejects `npx` or the cold start is slow:

```sh
npx -y patchxnote-agent@latest install --print-config
```

Use the printed absolute-path MCP config in Cursor.

## Evidence Gate

Do not claim Cursor first-class skill support until all of these pass:

- skill installation path verified in a clean Cursor environment
- skill activates for PatchXNote setup prompt
- skill does not activate for generic summarization
- `setup --client cursor` runs in the same runtime Cursor uses for MCP
- browser OAuth completes without codes/tokens in chat
- `tools/list`, `patchxnote_get_current_user`, and `patchxnote_list_memories {"platform":"mobile","limit":5}` pass
