# PatchXNote MCP 官网页面规格

**定位：** 官网第一屏是“选择你正在用的 AI 工具”，不是文档目录。页面风格参考产品化 MCP onboarding：深色背景、紧凑客户端卡片、明确状态标签、一个主操作。

## 信息架构

### 首页 / 客户端选择页

- 顶部导航：PatchXNote MCP、Clients、Setup、Security、Resources。
- 第一屏标题：`Use PatchXNote in every MCP client.`
- 副标题：`One login. Secret-free local config. Your PatchXNote summaries inside the AI tools you already use.`
- 主区域：客户端卡片网格。
- 默认展示：P0 和 P0.5。
- 次级折叠区：P1 客户端和观察列表。

### 卡片字段

每张卡片来自 `docs/mcp-clients/clients.json`：

- Logo 或安全占位图标
- 客户端名称
- 类型：AI Editor / CLI Agent / Desktop Agent / Office Platform
- 区域：Global / Domestic / Platform
- 状态：`One-click`、`Setup command`、`Manual`、`Remote MCP`、`Coming soon`
- 主操作：复制 setup 命令、打开 deeplink、复制远程 URL、进入详情页

未确认 Logo 授权前，公开版本只使用文本、首字母图标或用户自有图形资产。

## 详情页模板

URL 建议：

```text
/mcp/clients/vscode
/mcp/clients/cursor
/mcp/clients/codex
/mcp/clients/workbuddy
/mcp/platforms/feishu-aily
```

页面块：

- Hero：客户端 Logo、客户端名、状态标签、主安装按钮。
- 60-second setup：复制命令或 deeplink。
- Config fallback：无密钥 MCP 配置片段。
- Login：说明 `setup` 会优先打开浏览器登录；服务器未启用时降级到终端 OTP。
- Verify：给用户复制一条验证提示。
- Caveats：客户端专属注意事项，例如重启、远程运行时、WSL/Windows keychain 不共享。
- Safety：说明 token 只进 OS 安全存储，MCP 配置不放密钥。

## 主操作规则

- `supported + auto_write=true`：主按钮是复制 `npx -y patchxnote-agent@latest setup --client <id>`。
- `deeplink` 已验证：可提供 deeplink 按钮，但旁边保留 setup 命令。
- `manual`：主按钮复制 MCP 配置或客户端官方 CLI 命令。
- `planned / platform`：主按钮复制远程 MCP URL 或显示等待名单；不宣称已经可用。

## P0/P0.5 首屏卡片

- VS Code / GitHub Copilot
- Cursor
- Codex / ChatGPT Desktop / Codex IDE
- Claude Code
- Claude Desktop
- Windsurf
- Trae / Trae CN / TraeWork Code
- Qoder
- WorkBuddy / Tencent CodeBuddy WorkBuddy
- Feishu Aily / Doubao Work Partner
- Tencent Agent Development Platform / Enterprise WorkBuddy

## 验证提示

每个本地客户端详情页提供同一个安全验证提示：

```text
List the PatchXNote MCP tools and show only their names.
```

已登录用户可以继续复制：

```text
Read my recent PatchXNote summaries from the mobile platform. Return titles and timestamps only.
```

这个提示不要求模型导出原始转写、provider payload 或完整模型输入输出。

## 发布保护

- 上线前加 staging/noindex，避免未验收的一键按钮被搜索引擎收录。
- 不采集包含用户路径、账号、手机号或配置内容的埋点。
- 每次改 registry 后跑 `node docs/mcp-clients/validate-clients.mjs`。
