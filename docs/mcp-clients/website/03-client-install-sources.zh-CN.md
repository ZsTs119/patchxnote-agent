# PatchXNote MCP 官网客户端安装官方依据

**日期：** 2026-08-28

**定位：** 官网实现 `web/mcp/` 时，所有客户端卡片、详情页、按钮、命令、deeplink、远程 MCP 指引都必须来自真实官方文档、当前 `clients.json` 事实源和本机验收结果。本文禁止 mock 数据、猜测路径和未验收的一键承诺。

## 总规则

- 官网可以先实现页面，但按钮状态必须保守。
- 官方文档写明支持 deeplink / CLI / 配置文件，官网才展示对应入口。
- 真实本机验收通过后，才能把按钮从 `Copy AI setup prompt` 升级为 `Install in <Client>`。
- 如果官方文档只说明手动配置，官网只给手动配置和 AI-assisted prompt。
- 如果官方文档无法抓到完整配置细节，但确认有 MCP 能力，先标 `researched`，等本机或人工复核后再写具体自动逻辑。
- 远程 MCP 平台只展示 remote MCP URL、OAuth connector 流程和验收状态，不展示本地 `npx` 作为主路径。

## 本地 callback 与官网成功页

本地 MCP 登录不能省掉 loopback callback，因为本地 CLI 需要 OAuth code 才能换 token 并保存到 OS 原生安全存储。

推荐用户可见流程：

```text
patchxnote-agent mcp login
 -> 启动 127.0.0.1:<port>/callback
 -> 打开 GoServer 官网风格授权页
 -> 用户登录和授权
 -> GoServer redirect 到 127.0.0.1:<port>/callback?code=...&state=...
 -> 本地 callback 校验 state，把 code 交给 CLI
 -> 浏览器跳转 GoServer /mcp/auth/success
 -> CLI 完成 token exchange 和安全存储
```

官网成功页文案不能写成“token 已保存成功”，除非后续增加本地状态回报机制。V1 更稳的说法是：

```text
授权信息已返回本机。请回到编辑器或终端，等待 PatchXNote MCP 完成最后的连接验证。
```

远程 MCP / 云平台不使用本地 callback。它们走平台支持的 OAuth / connector 回调流程。

## 编辑器入口

### VS Code / GitHub Copilot

官方依据：

- https://code.visualstudio.com/docs/agent-customization/mcp-servers
- https://code.visualstudio.com/docs/agents/reference/mcp-configuration

真实逻辑：

- VS Code 支持 MCP server gallery、用户 profile 配置、workspace `.vscode/mcp.json`、Dev Container 配置和命令行 `code --add-mcp`。
- VS Code 的官方配置结构是 `servers`，不是通用 MCP 示例里常见的 `mcpServers`。
- 用户 profile 配置适合“所有项目可用”，workspace 配置适合“随项目共享”。
- remote / Agent Host / Dev Container 场景有运行时差异。官网必须提醒：在哪个运行时启动 MCP，就在哪个运行时执行 PatchXNote setup/login。
- VS Code 首次启动 MCP server 时有 trust prompt。官网不能绕过用户确认。

官网动作：

- V1 主路径：`Copy AI setup prompt` + `Copy setup command`。
- 可验收路径：`code --add-mcp` 写入 user profile，或写 `.vscode/mcp.json`。
- 真实验收通过后，官网可显示 `Install in VS Code`，但仍保留手动配置。

### Cursor

官方依据：

- https://cursor.com/docs/mcp
- https://cursor.com/docs/mcp/install-links

真实逻辑：

- Cursor 支持本地 stdio、SSE、streamable HTTP 等 MCP 配置。
- Cursor 官方 deeplink 格式是：

```text
cursor://anysphere.cursor-deeplink/mcp/install?name=$NAME&config=$BASE64_ENCODED_CONFIG
```

- `config` 是 JSON stringify 后的 base64 配置，格式和 Cursor `mcp.json` 里的 server 配置一致。
- 浏览器会弹系统确认，用户确认后进入 Cursor 安装页或设置页。

官网动作：

- V1 可以生成真实 Cursor deeplink，但只有本机验收通过后才作为主按钮。
- 验收前主路径仍是 `Copy AI setup prompt`。
- 手动 fallback：`~/.cursor/mcp.json` 无密钥配置。

### Windsurf / Devin Desktop Cascade

官方依据：

- https://docs.windsurf.com/windsurf/cascade/mcp

真实逻辑：

- 该页面当前跳转到 Devin Docs 的 Cascade MCP 文档。
- Cascade 支持 marketplace、deeplink、`~/.codeium/windsurf/mcp_config.json` 手动配置。
- 文档展示 deeplink 形式：

```text
windsurf://windsurf-mcp-registry?serverName=<server-name>
```

- 文档同时说明支持 stdio、Streamable HTTP、SSE 和 OAuth，但企业/团队可通过管理员关闭 MCP 访问。
- Cascade 工具总数存在限制，官网要避免一次暴露过多 PatchXNote 工具。

官网动作：

- V1 主路径：`Copy AI setup prompt` + `Copy setup command`。
- marketplace/deeplink 只有 PatchXNote 进入对应 registry 或真实验收后才显示为主按钮。
- 手动 fallback：`~/.codeium/windsurf/mcp_config.json`。

### Trae / Trae CN / TraeWork Code

官方依据：

- https://docs.trae.ai/ide/model-context-protocol
- https://docs.trae.ai/ide/add-mcp-servers

真实逻辑：

- 官方页面确认 TraeCode MCP 支持多种 transport，并支持用户添加 MCP server。
- 当前抓取到的公开页面内容不足以确认可安全自动写入的本地配置路径、deeplink 参数和不同版本差异。
- Trae、Trae CN、TraeWork Code 不能混作同一个验收对象。

官网动作：

- V1 主路径：`Copy AI setup prompt` + 手动配置说明。
- 不写 `Install in Trae`，直到本机或官方文档确认具体 deeplink/config 路径。
- 详情页标明：`Manual setup, client acceptance pending`。

### Qoder

官方依据：

- https://docs.qoder.com/user-guide/chat/model-context-protocol
- https://docs.qoder.com/user-guide/deeplink
- https://docs.qoder.com/troubleshooting/mcp-common-issue

真实逻辑：

- 官方页面确认 Qoder 支持 MCP，并有 deeplink 能力用于快速添加 MCP 服务配置。
- Qoder 故障排查文档提醒：未打开项目目录可能处于 Ask Mode，不能调用 MCP tools；需要 Agent Mode 和已连接 server。
- 当前还需要本机验收 deeplink 参数、配置落点和重启/刷新方式。

官网动作：

- V1 主路径：`Copy AI setup prompt`。
- deeplink 可作为候选字段，不做主按钮。
- 详情页验证提示要写明：打开项目目录、切换 Agent Mode、确认 MCP server 已连接。

### WorkBuddy / Tencent CodeBuddy WorkBuddy 本地模式

官方依据：

- https://www.workbuddy.cn/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/Connector
- https://www.codebuddy.cn/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/MCP-Guide

真实逻辑：

- WorkBuddy 连接器用于外部服务接入，技术形态包含 `MCP + CLI` 和 `Skill + CLI`。
- WorkBuddy 文档说明 MCP 已集成到界面，不需要用户手动改隐藏配置即可可视化接入。
- 本地配置支持用户级 `~/.workbuddy/mcp.json` 和项目级 `<项目目录>/.workbuddy/mcp.json`。
- 配置入口在插件侧边栏的 MCP server 管理区域，保存后通过状态灯判断连接成功或异常。
- 自定义连接器和 MCP 服务存在凭据、授权和第三方共享边界，需要谨慎提示。

官网动作：

- V1 主路径：`Copy AI setup prompt` + `Manual setup guide`。
- 不把 WorkBuddy 本地模式和企业版/平台型 WorkBuddy 混成一个入口。
- 详情页要分清：本地 MCP 配置、连接器授权、企业管理员控制。

### Zed

官方依据：

- https://zed.dev/docs/ai/mcp

真实逻辑：

- Zed 支持 MCP Tools 和 Prompts。
- 可以通过 Extensions 安装，也可以在 Settings -> AI -> MCP Servers 添加 Local Server 或 Remote Server。
- Zed 配置键是 `context_servers`，不是 `mcpServers`。
- Remote MCP 如果没有配置 Authorization header，Zed 会按标准 MCP OAuth 流程提示认证。
- 连接后可通过设置页状态点确认 server active。

官网动作：

- V1 可以给 Zed 专用 `context_servers` 配置，不要复用通用 `mcpServers` JSON。
- 本地 PatchXNote 用 stdio；远程平台候选可以走 URL + OAuth。
- 当前标 `planned / researched`，等用户安装 Zed 或官方安装入口确认后再升级。

### JetBrains AI Assistant

官方依据：

- https://www.jetbrains.com/help/ai-assistant/settings-reference-mcp.html
- https://www.jetbrains.com/help/ai-assistant/mcp.html

真实逻辑：

- JetBrains AI Assistant 有 MCP 配置页，可添加 MCP server。
- JetBrains 2025.2 起还内置 IDE MCP Server，用来把 IDE 能力暴露给 Codex、Claude Code、VS Code 等外部客户端。这和 PatchXNote 作为外部 MCP server 接入 JetBrains AI Assistant 是两个方向。
- 部分 JetBrains agent 需要开启“Pass custom MCP servers”才能把已配置 MCP server 暴露给 agent。

官网动作：

- V1 只写 `Manual setup guide`，不写一键。
- 详情页要避免把“JetBrains IDE 作为 MCP server”误写成“PatchXNote 已自动接入 JetBrains”。

### Cline / Continue / Roo-derived VS Code agents

官方依据：

- https://docs.cline.bot/mcp/mcp-overview
- https://docs.continue.dev/customize/mcp-tools
- https://docs.continue.dev/customize/deep-dives/mcp
- https://roocodeinc.github.io/Roo-Code/features/mcp/using-mcp-in-roo/

真实逻辑：

- Cline、Continue、Roo 都是 VS Code 生态内的独立 MCP Host / extension，需要单独配置和验收。
- Continue 使用 `mcpServers` 相关配置块，并且 MCP 只能在 agent mode 使用。
- Cline 有 MCP Servers 管理入口、manual config 和 CLI wizard。
- Roo 有自己的 MCP 配置/管理入口，不等同 VS Code 官方 MCP 配置。

官网动作：

- V1 放在 P1 或“VS Code-derived agents”折叠区。
- 不把 VS Code 官方验收自动继承成 Cline / Continue / Roo 已验收。

## 本地 MCP / CLI 入口

### Codex / ChatGPT Desktop / Codex IDE

官方依据：

- https://developers.openai.com/codex/mcp
- https://developers.openai.com/codex/config-reference
- https://developers.openai.com/codex/config-basic

真实逻辑：

- Codex MCP 配置在 `config.toml`，默认 `~/.codex/config.toml`，也可在受信任项目里用 `.codex/config.toml`。
- ChatGPT desktop、Codex CLI、IDE extension 共享 Codex MCP 配置。
- Codex CLI 支持 `codex mcp add <name> --url <remote-url>` 这类远程配置，也支持本地 stdio 配置。
- 本地 PatchXNote 仍建议走 `patchxnote-agent setup --client codex`，因为我们还要处理本地 login 和 OS 安全存储。

官网动作：

- V1 主路径：`Copy AI setup prompt`。
- fallback：`codex mcp add patchxnote -- npx -y patchxnote-agent@latest mcp serve` 或 TOML 配置。
- 不把 Codex plugin / ChatGPT app 目录当成已发布入口。

### Claude Code

官方依据：

- https://code.claude.com/docs/en/mcp

真实逻辑：

- Claude Code 支持 `claude mcp add` 添加 remote HTTP 或 local stdio server。
- HTTP 是连接远程 MCP server 的推荐 transport。
- Claude Code 会在 `/mcp` 和 `claude mcp list` 里显示配置问题，例如空 URL、前后隐藏空格、同名多 scope 冲突。
- Claude Code 有 tools 输出限制、approval、组织控制等边界。

官网动作：

- V1 主路径：`Copy AI setup prompt`。
- fallback：`claude mcp add patchxnote -- npx -y patchxnote-agent@latest mcp serve`。
- 远程 MCP 可作为平台/remote 分支，不要和本地 stdio 混写。

### Claude Desktop

官方依据：

- https://modelcontextprotocol.io/docs/2026-07-28/develop/connect-local-servers
- https://www.anthropic.com/engineering/desktop-extensions

真实逻辑：

- Claude Desktop 本地 MCP 传统路径是编辑 `claude_desktop_config.json`，macOS 和 Windows 路径不同。
- 配置后需要完全退出并重启 Claude Desktop。
- Anthropic Desktop Extensions 使用 `.mcpb` 包实现一键安装，但这需要独立打包、manifest、依赖和发布流程。

官网动作：

- V1 主路径：`Copy AI setup prompt` + manual `claude_desktop_config.json`。
- 不写 `.mcpb` 一键，直到我们真的完成打包和验收。

### Gemini CLI

官方依据：

- https://github.com/google-gemini/gemini-cli/blob/main/docs/tools/mcp-server.md

真实逻辑：

- Gemini CLI 从 `settings.json` 的 `mcpServers` 读取 MCP 配置。
- 支持 stdio、SSE、Streamable HTTP 三类 transport。
- 文档说明远程 OAuth 可自动发现，OAuth 登录需要本机能打开浏览器并接收 localhost redirect。
- Gemini CLI 有环境变量扩展和敏感环境变量清洗逻辑。

官网动作：

- V1 先给 manual config 和 AI-assisted prompt。
- 本地 PatchXNote 用 stdio；远程 PatchXNote 可在后续验证 OAuth discovery 后开放。

### Qwen Code

官方依据：

- https://github.com/QwenLM/qwen-code/blob/main/docs/users/features/mcp.md
- https://qwenlm.github.io/qwen-code-docs/en/users/overview/

真实逻辑：

- Qwen Code 从 `settings.json` 的 `mcpServers` 读取配置，也支持 `qwen mcp` 命令。
- 默认 user scope 是 `~/.qwen/settings.json`，project scope 是 `.qwen/settings.json`。
- Transport 选择中，HTTP 用于远程服务，SSE 属于 legacy，stdio 用于本地进程。
- 添加 server 后如果 Qwen Code 已经运行，通常需要同项目重启后再使用。

官网动作：

- V1 先给 `Copy AI setup prompt` 和 manual config。
- 远程 HTTP 和本地 stdio 分开展示。

### Kimi Code / Kimi CLI

官方依据：

- https://www.kimi.com/code/docs/en/kimi-code-cli/customization/mcp.html
- https://moonshotai.github.io/kimi-cli/en/customization/mcp.html

真实逻辑：

- Kimi Code CLI 支持 stdio、HTTP、SSE 三种 MCP 连接方式。
- 配置文件是 `mcp.json`，user level 在 `~/.kimi-code/mcp.json` 或 `$KIMI_CODE_HOME/mcp.json`，project level 在 `.kimi-code/mcp.json`。
- TUI 里可用 `/mcp-config` 管理配置，用 `/mcp` 查看连接状态。
- project-level server 在未信任目录会触发 workspace trust prompt。

官网动作：

- V1 给 AI-assisted prompt、manual config 和 `/mcp` 验证提示。
- 不写一键。

### OpenCode

官方依据：

- https://opencode.ai/docs/tools/
- https://opencode.ai/docs/config/
- https://opencode.ai/docs/cli/

真实逻辑：

- OpenCode 支持通过 custom tools / MCP servers 扩展工具。
- CLI 提供 `opencode mcp add` 引导添加 local 或 remote MCP server。
- CLI 提供 `opencode mcp list` / `opencode mcp ls` 查看配置状态，也有 OAuth auth/logout/debug 子命令。
- 配置文件支持 JSON / JSONC。

官网动作：

- V1 给 `Copy AI setup prompt` 和 `opencode mcp add` 手动指引。
- 不写一键。

## 云平台入口

### Feishu Aily / Doubao Work Partner

官方依据：

- https://www.feishu.cn/content/article/7576921890476788922
- https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/mcp_integration/mcp_introduction?lang=zh-CN
- https://open.feishu.cn/document/mcp_open_tools/developers-call-remote-mcp-server?lang=zh-CN

真实逻辑：

- 飞书 Aily 官方材料说明 Aily 支持工具 / MCP 服务和服务市场，企业也可开发自己的 MCP 服务。
- 飞书开放平台文档提供 MCP 能力说明和远程 MCP 调用方向。
- 这类平台不能当成本地编辑器处理，不能让用户复制本地 `npx` 命令作为主路径。

官网动作：

- V1 展示 `Connect remote MCP`，复制 remote MCP URL 和平台接入说明。
- 当前标 `planned / researched`，等真实平台控制台接入后才能升级。

### Tencent Agent Development Platform

官方依据：

- https://cloud.tencent.com/document/product/1759/117855
- https://intl.cloud.tencent.com/document/product/1254/69956
- https://github.com/TencentCloudADP

真实逻辑：

- 腾讯云 ADP 是企业级 Agent 开发平台，支持构建、发布、管理企业智能体。
- 当前公开资料不足以确认它对第三方 remote MCP 的具体配置表单、OAuth discovery 和 connector session 细节。

官网动作：

- V1 展示为云平台研究/验收中，不写已接入。
- 如果平台实际只支持 HTTP tool / webhook，需要单独评估，不伪装成 MCP 已闭环。

### Enterprise WorkBuddy

官方依据：

- https://www.workbuddy.cn/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/Connector
- https://www.codebuddy.cn/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/MCP-Guide

真实逻辑：

- WorkBuddy 有连接器和 MCP 配置体系，也有企业/管理员边界。
- 企业版 WorkBuddy 与本地 WorkBuddy/CodeBuddy 不应混为一个安装入口。

官网动作：

- V1 先写 `Connect remote MCP` 或 `Enterprise setup pending`。
- 真实接入时必须记录管理员开关、连接器授权、状态灯、解绑路径和工具调用验收。

## 页面数据要求

`web/mcp/data/clients.json` 不能是 mock。每条记录都应该从以下来源之一生成：

- `patchnote-agent/docs/mcp-clients/clients.json`
- 本文官方依据
- 本机真实验收记录
- 服务端真实 remote MCP 配置

建议每条 client 加这些字段：

```json
{
  "id": "cursor",
  "name": "Cursor",
  "category": "ai-editor",
  "support_status": "supported",
  "evidence_state": "researched",
  "official_sources": [
    "https://cursor.com/docs/mcp",
    "https://cursor.com/docs/mcp/install-links"
  ],
  "actions": {
    "recommended": "copy_ai_prompt",
    "one_click": {
      "enabled": false,
      "reason": "deeplink format confirmed; local client acceptance pending"
    },
    "manual": {
      "enabled": true
    }
  }
}
```

上线前如果某个字段没有真实来源，就不要输出这个字段，或显式标 `unknown / pending`。

## 第一批真实验收建议

用户本机已具备 VS Code 和 Cursor，第一批建议只做这两个。

### Cursor 验收

- 生成 PatchXNote MCP deeplink。
- 浏览器点击后确认系统弹窗。
- Cursor 打开安装页并预填 `patchxnote` server。
- 用户确认安装。
- 运行或触发 `mcp login`，完成官网授权页。
- callback 跳官网成功页。
- Cursor 刷新 MCP server。
- 验证 `tools/list` 只显示 PatchXNote 工具名。
- 调用一个只读工具，只返回标题/时间等脱敏字段。

### VS Code 验收

- 备份或使用测试 profile。
- 测 `code --add-mcp` 或 `.vscode/mcp.json`。
- 打开 VS Code MCP 列表，确认 trust prompt。
- 启动 PatchXNote MCP server。
- 运行或触发 `mcp login`，完成官网授权页。
- callback 跳官网成功页。
- 重载窗口或刷新 MCP server。
- 验证 `tools/list` 和一个只读工具。

验收记录只保存：

- 客户端名称、版本、系统。
- 安装方式。
- 工具数量和工具名。
- 脱敏 request id / 状态码。
- 成功/失败截图或文字结论。

不要保存：

- 手机号、验证码、token、OAuth code、refresh token。
- 原始录音、完整转写、完整模型输入输出。
- 用户主配置里的私有路径，除非脱敏。
