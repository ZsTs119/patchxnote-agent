# PatchXNote MCP Starter Prompts

Use these prompts for README examples, marketplace submissions, local plugin cards, and review test cases.

## Primary Setup

```text
请帮我接入 PatchXNote MCP，并在客户端支持 Agent Skills 时安装 PatchXNote MCP Skill。首先，如果当前是支持本地命令执行、并且可以运行 skills CLI 的环境，请在本机执行 npx -y skills add ZsTs119/patchxnote-agent --skill patchxnote-mcp -g，让后续新会话或长上下文能记住 PatchXNote SOP；如果当前客户端不支持 Skill 安装，就直接继续 MCP setup。请识别你所在的 MCP 客户端，如果是 Codex、Cursor、VS Code、Claude Desktop、Claude Code、Windsurf 等本地 MCP 客户端，请在当前本机、并且在后续会启动 MCP 服务的同一个运行时执行 npx -y patchxnote-agent@latest setup --client <client-id>；如果不确定客户端 ID，就先执行 npx -y patchxnote-agent@latest mcp config 并使用打印出的通用 stdio 配置。登录时请打开浏览器让我完成 PatchXNote 手机号验证码授权，不要让我把验证码、OAuth code、授权码、access token、refresh token 或 webhook secret 粘贴到聊天里。完成后请调用 patchxnote_get_current_user 和 patchxnote_list_memories，参数 {"platform":"mobile","limit":5} 验证可用。参考资料：https://github.com/ZsTs119/patchxnote-agent 和 https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd。
```

```text
Help me connect PatchXNote MCP and install the PatchXNote MCP Skill when the client supports Agent Skills. First, if this is a local command-capable environment where the skills CLI can run, execute npx -y skills add ZsTs119/patchxnote-agent --skill patchxnote-mcp -g on this machine so future or long sessions remember the PatchXNote SOP; if skill installation is not supported, continue with MCP setup directly. Identify the MCP client you are running in; if this is a local MCP client such as Codex, Cursor, VS Code, Claude Desktop, Claude Code, or Windsurf, run npx -y patchxnote-agent@latest setup --client <client-id> on this machine in the same OS/runtime that will later launch the MCP server. If you are not sure which client ID to use, first run npx -y patchxnote-agent@latest mcp config and use the printed generic stdio config. For login, open the browser and let me complete PatchXNote phone-code authorization there; do not ask me to paste OTP codes, OAuth codes, authorization codes, access tokens, refresh tokens, or webhook secrets into chat. After setup, call patchxnote_get_current_user and patchxnote_list_memories with {"platform":"mobile","limit":5} to verify it works. References: https://github.com/ZsTs119/patchxnote-agent and https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd
```

## Verification

```text
请检查当前 PatchXNote MCP 是否已经登录，并调用 patchxnote_get_current_user 验证。
```

```text
List my latest 5 PatchXNote mobile summaries.
```

```text
检查当前 PatchXNote MCP 有多少工具，并说明哪些工具是读取账号/记录，哪些工具会产生本地 webhook 或发送副作用。
```

## Workflow

```text
把这条 PatchXNote 总结整理成 Markdown 草稿，我确认后再发。
```

```text
帮我按 mobile 平台统计 PatchXNote event_summary 和 daily_digest，各自有多少条。
```

```text
帮我排查为什么 Cursor 里 PatchXNote MCP 登录了但 tools/list 还是不可用。
```

## Negative Prompts

These should not activate PatchXNote-specific behavior unless PatchXNote is explicitly mentioned:

```text
Summarize this article.
```

```text
Create a generic MCP server for my SaaS dashboard.
```

```text
Publish my unrelated skill to a marketplace.
```
