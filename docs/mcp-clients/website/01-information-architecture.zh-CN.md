# PatchXNote MCP 官网信息架构

**定位：** 官网第一屏先建立产品识别和核心转化，不把所有客户端、卖点和安装细节塞进首屏。客户端选择、安装方式、安全说明和平台接入放到后续屏幕。

## 信息架构

### 首页第一屏 / Product Hero

第一屏只保留：

- 顶部导航：`PATCHX`、`MCP`、`Clients`、`Security`、`Docs`、`Download App`、`Get started`。
- 产品名：`PatchXNote MCP`。
- 主标题：`隐私归你，AI 由你掌控`。
- 副标题：`把你的真实对话，安全接入 Cursor、VS Code、Codex 和更多 AI 工具。`
- 主 CTA：`连接我的 AI 工具`。
- 次 CTA：`查看支持的编辑器`。
- 右侧或中心：高级黑风格 PatchXNote 录音卡主视觉。
- 底部下一屏提示：`Cursor / VS Code / Codex / Claude Code / 更多编辑器` 的轻量文字或单色图标。

第一屏不放：

- 大客户端卡片网格。
- 四个以上功能卖点。
- 价格、参数、配件、长说明。
- 具体安装命令和配置代码。
- `更多客户端持续接入中` 这类容易让用户误解为“还没接好”的文案。

下一屏入口文案建议：

```text
查看全部支持的编辑器
选择你的 AI 工具
更多编辑器
```

不要使用：

```text
更多客户端持续接入中
```

### 第二屏 / 客户端选择页

- 标题：`选择你的 AI 工具`。
- 副标题：`同一个 PatchXNote MCP，接入你已经在用的编辑器、CLI 和云平台。`
- 顶部切换：`编辑器`、`云平台`、`本地 MCP`。
- 主区域：按当前切换项展示对应渠道卡片。
- 点击卡片：进入对应详情页，详情页复用同一套模板，只替换客户端名称、主动作、命令、配置和注意事项。
- 兜底卡片：`找不到你的渠道？`，进入通用 MCP 配置说明，提供 `mcpServers` block 让用户粘贴到支持 MCP 的客户端配置里。
- 首版默认只做能闭环的重点渠道，不把观察列表做成强曝光入口。

首版建议闭环渠道：

- 编辑器：`Cursor`、`VS Code`。
- 本地 MCP：`Codex`、`Claude Code`。
- 云平台：`飞书 Aily`、`企业 WorkBuddy` 或先保留一个云平台模板入口。
- 兜底：`通用 MCP 配置`。

### 卡片字段

每张卡片来自 `docs/mcp-clients/clients.json`：

- Logo 或安全占位图标。未确认商标使用规范前，公开版本优先用单色银灰图标或文字首字母。
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
/mcp/clients/generic-mcp
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

这些客户端进入第二屏默认卡片区，不进入第一屏完整网格：

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

## 授权结果页

授权页组不是教学页，也不是产品展示页。它只负责把 OAuth 流程清楚收口。

### 手机号登录页

当用户打开授权链接但官网未登录时，先展示手机号登录弹窗或页面。登录完成后继续进入授权确认页。

保留：

- 顶部简化品牌：`PATCHX` + `MCP`。
- 标题：`登录 PatchXNote`。
- 说明：`登录后继续授权当前 AI 工具。`
- 手机号输入框。
- 验证码输入框。
- `获取验证码` 按钮。
- 主 CTA：`登录并继续`。
- 次级入口：`返回设置指南`。
- 底部小字：`验证码仅用于本次登录，不会写入 MCP 配置。`

不保留：

- 产品大图。
- 价格、参数、卖点。
- 第三方客户端 Logo 和卡片。
- token、OAuth code、验证码明文、请求日志或调试参数。

### 授权确认页

用户已登录后，再展示授权确认页。页面只说明当前客户端、只读权限和继续/取消动作。

保留：

- 当前客户端，例如 `Cursor`。
- 只读权限范围：账号状态、额度概览、最近摘要。
- 主 CTA：`授权并继续`。
- 次 CTA：`取消`。
- 安全说明：`授权完成后，凭据只写入本机安全存储。`

### 授权成功页

保留：

- 顶部简化品牌：`PATCHX` + `MCP`。
- 一个银色完成符号：连接环、短刻线或金属确认符号。
- 标题：`授权已完成`。
- 说明：`授权信息已返回本地运行时。请回到编辑器或终端，等待 PatchXNote MCP 完成连接验证。`
- 主 CTA：`返回 AI 编辑器`。
- 次 CTA：`查看设置指南`。
- 底部小字：`凭据保存在本机安全存储，MCP 配置不包含密钥。`

不保留：

- 产品大图。
- 三步流程大图。
- 网站到编辑器/终端的大型示意图。
- 客户端 Logo、客户端卡片、平台入口。
- 价格、参数、卖点、安装命令、配置代码。
- token、OAuth code、手机号、验证码、日志或 QR code。

### 授权失败页

保留：

- 标题：`授权未完成`。
- 一句可读原因：过期、取消、state mismatch、网络失败或未知错误。
- 主 CTA：`重新授权`。
- 次 CTA：`返回设置指南`。
- 小字提示：不要展示敏感参数，只展示安全的错误代号或 request id。

## 发布保护

- 上线前加 staging/noindex，避免未验收的一键按钮被搜索引擎收录。
- 不采集包含用户路径、账号、手机号或配置内容的埋点。
- 每次改 registry 后跑 `node docs/mcp-clients/validate-clients.mjs`。
