# PatchXNote MCP Onboarding

Use this reference when installing, repairing, or verifying PatchXNote MCP for a local AI client.

## One-Line User Prompt

Chinese:

```text
请帮我接入 PatchXNote MCP，并安装或使用 PatchXNote MCP Skill。请识别你所在的客户端，如果是 Codex、Cursor、VS Code、Claude Desktop、Claude Code、Windsurf 等本地 MCP 客户端，请在当前本机、并且在后续会启动 MCP 服务的同一个运行时执行 npx -y patchxnote-agent@latest setup --client <client-id>；如果不确定客户端 ID，就先执行 npx -y patchxnote-agent@latest mcp config 并使用打印出的通用 stdio 配置。登录时请打开浏览器让我完成 PatchXNote 手机号验证码授权，不要让我把验证码、OAuth code、授权码、access token、refresh token 或 webhook secret 粘贴到聊天里。完成后请调用 patchxnote_get_current_user 和 patchxnote_list_memories，参数 {"platform":"mobile","limit":5} 验证可用。参考资料：https://github.com/ZsTs119/patchxnote-agent 和 https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd。
```

English:

```text
Help me connect PatchXNote MCP, and install or use the PatchXNote MCP Skill. Identify the MCP client you are running in; if this is a local MCP client such as Codex, Cursor, VS Code, Claude Desktop, Claude Code, or Windsurf, run npx -y patchxnote-agent@latest setup --client <client-id> on this machine in the same OS/runtime that will later launch the MCP server. If you are not sure which client ID to use, first run npx -y patchxnote-agent@latest mcp config and use the printed generic stdio config. For login, open the browser and let me complete PatchXNote phone-code authorization there; do not ask me to paste OTP codes, OAuth codes, authorization codes, access tokens, refresh tokens, or webhook secrets into chat. After setup, call patchxnote_get_current_user and patchxnote_list_memories with {"platform":"mobile","limit":5} to verify it works. References: https://github.com/ZsTs119/patchxnote-agent and https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd
```

## Client Detection

Prefer the actual host the current session is running in:

| Client family | `setup --client` value | Notes |
| --- | --- | --- |
| VS Code / GitHub Copilot | `vscode` | Run setup in the same local, WSL, SSH, or Dev Container runtime that launches MCP. |
| Cursor | `cursor` | Local config merge is supported after user confirmation. |
| Codex / ChatGPT Desktop / Codex IDE | `codex` | New session or reload may be needed after config changes. |
| Claude Code | `claude-code` | V1 may return manual commands; install plugin separately if using Claude plugin marketplace. |
| Claude Desktop | `claude-desktop` | Desktop app restart is usually needed. |
| Windsurf | `windsurf` | Run where Cascade launches MCP servers. |
| Trae / Trae CN / TraeWork Code | `trae` | Manual UI/config path in V1. |
| Qoder | `qoder` | Manual UI/deeplink path in V1 until platform acceptance is recorded. |
| WorkBuddy | `workbuddy` | Local desktop and enterprise platform modes are different. |

If the client is unknown, do not invent an ID. Print generic config:

```sh
npx -y patchxnote-agent@latest mcp config
```

## Local Setup

For local stdio MCP clients:

```sh
npx -y patchxnote-agent@latest setup --client <client-id>
```

Setup may plan, confirm, back up, and write supported client config. It should not delete unrelated MCP servers. If an existing `patchxnote` entry exists, ask before replacing it. Use `--force` only when the user explicitly agrees to replace that entry.

Manual fallback:

```sh
npx -y patchxnote-agent@latest mcp config
```

Absolute-path fallback for slow first download or clients that reject `npx`:

```sh
npx -y patchxnote-agent@latest install --print-config
```

## Login

Use browser OAuth for MCP:

```sh
npx -y patchxnote-agent@latest mcp login
```

`setup --client <id>` can reuse the same browser OAuth flow. `mcp serve` does not open a browser when the editor starts, so do not expect a fresh editor launch to log the user in.

The user completes phone verification in the browser page. Do not ask for codes, token-shaped strings, or webhook secrets in chat.

## Verification

First verify the local CLI auth state:

```sh
npx -y patchxnote-agent@latest mcp status --verify
```

Then verify real MCP capability:

1. `initialize`
2. `tools/list`
3. `tools/call patchxnote_get_current_user`
4. `tools/call patchxnote_list_memories` with `{"platform":"mobile","limit":5}`

If the client asks for a current tool count, answer from `tools/list`, not from this skill file or memory.

## Cloud Or Hosted Clients

Cloud-only clients cannot run the user's local `npx` process. For Feishu Aily, Doubao Work Partner, Tencent Agent Development Platform, and enterprise WorkBuddy platform mode, use a hosted remote MCP path only when that channel has evidence. Do not claim local stdio setup proves remote platform acceptance.
