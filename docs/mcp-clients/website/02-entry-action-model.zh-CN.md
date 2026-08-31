# PatchXNote MCP 官网三类入口与动作方案

**日期：** 2026-08-27

**定位：** 本文用于统一 PatchXNote MCP 官网、客户端详情页、授权页体验和后续 GoServer `web/` 实现口径。本文是产品与实现边界文档，不代表所有客户端已经完成真实验收。

## 核心结论

截图里讨论的“官网、一键安装、登录授权、不同客户端入口”在用户体验上应合并成一件事：

```text
PatchXNote MCP 官网安装与授权中枢
```

用户只需要先选择正在用的 AI 工具，然后按照该工具支持的最佳路径完成接入。工程实现上仍拆成三类入口，因为它们的能力边界不同：

1. **编辑器入口**：能打开客户端、预填 MCP 配置或跳到设置页。
2. **本地 MCP / CLI 入口**：通过终端命令或让本地 AI 执行命令完成安装。
3. **云平台入口**：通过 HTTPS remote MCP URL 和 OAuth connector session 接入。

官网可以做成统一页面，但按钮文案和实际动作必须按三类区分，避免把本地 `npx`、编辑器 deeplink 和云平台 remote MCP 混成同一种能力。

实现上不需要把三类入口都做成独立一级页面。推荐结构是：

```text
/mcp 首页
 -> 选择你的 AI 工具
 -> 用 编辑器 / 云平台 / 本地 MCP tab 切换卡片列表
 -> 点击某个卡片进入详情页
 -> 详情页按该渠道展示最佳安装方式和兜底配置
```

这样用户理解成本最低，官网也更容易先闭环。`12-cloud-platform-black-chrome.png` 和 `13-local-cli-mcp-black-chrome.png` 可作为不同类型详情页的视觉参考，不一定对应独立一级页面。

如果用户没有找到自己的渠道，展示 `找不到你的渠道？` 兜底卡片，进入通用 MCP 配置页，核心内容是复制 `mcpServers` block 到任何支持 MCP 的客户端配置。

首版建议只做少数可闭环入口：

- 编辑器：`Cursor`、`VS Code`。
- 本地 MCP：`Codex`、`Claude Code`。
- 云平台：先做一个云平台模板入口，优先承接 `飞书 Aily` / `企业 WorkBuddy` 的 remote MCP 流程。
- 兜底：`通用 MCP 配置`。

## 第一类：编辑器入口

### 适用客户端

| 客户端 | 当前优先级 | 当前状态 | 官网首版动作 |
| --- | --- | --- | --- |
| Cursor | P0 | `supported / researched` | 优先做 deeplink 安装按钮，真实验收前保留为复制 setup 命令 |
| VS Code / GitHub Copilot | P0 | `supported / researched` | 复制 setup 命令，验收 `code --add-mcp` 或网页安装后升级 |
| Windsurf | P0 | `supported / researched` | 复制 setup 命令，提示刷新 MCP servers |
| Trae / Trae CN / TraeWork Code | P0 | `manual / researched` | 手动 UI 配置和复制 setup 命令，不把多个版本混作一个验收 |
| Qoder | P0 | `manual / researched` | deeplink 或手动配置候选，验收前不写一键 |
| WorkBuddy 本地模式 | P0 | `manual / researched` | 手动 MCP / CLI 设置，不等同企业平台模式 |
| Zed | P1 | `planned / researched` | 观察和手动配置候选 |
| JetBrains AI Assistant | P1 | `planned / researched` | 观察和手动配置候选 |
| Cline / Continue / Roo 等 VS Code 扩展 | P1 | `planned / researched` | 继承 VS Code 类路径，但单独记录兼容性 |

### 用户路径

```text
官网选择客户端
 -> 点击 Install in <Client> 或复制 setup 命令
 -> 浏览器或系统弹出“打开客户端”确认
 -> 客户端预填 MCP server 配置
 -> 用户确认保存或安装
 -> 用户在客户端里发送官网提供的 AI setup prompt
 -> AI 或用户运行 patchxnote-agent setup/login
 -> 授权页弹出，用户自己登录和确认
 -> 回到客户端刷新 MCP 并验证工具列表
```

### 官网动作

首版按钮优先级：

1. `Install in <Client>`：仅当官方 deeplink 或网页安装格式已确认并真实客户端验收通过后显示为主按钮。
2. `Copy AI setup prompt`：复制一段话给当前编辑器里的 AI，让 AI 运行安装命令，并在需要用户授权时暂停。
3. `Copy setup command`：给用户复制可直接粘贴到终端的命令。
4. `Manual config`：展示无密钥 MCP JSON/TOML 作为兜底。

Cursor 这类支持 deeplink 的客户端，按钮可以打开类似系统确认弹窗，再进入客户端安装页。未完成真实验收前，官网可以展示按钮占位和 fallback，但不能写成“一键安装已支持”。

## 第二类：本地 MCP / CLI 入口

### 适用客户端

| 客户端 | 当前优先级 | 当前状态 | 官网首版动作 |
| --- | --- | --- | --- |
| Codex / ChatGPT Desktop / Codex IDE | P0 | `supported / researched` | 复制 AI setup prompt 和 `setup --client codex`，保留 `codex mcp add` 备用 |
| Claude Code | P0 | `manual / researched` | 复制 AI setup prompt 和 `claude mcp add` 备用命令 |
| Claude Desktop | P0 | `supported / researched` | 复制 setup 命令或手动 config，提示重启客户端 |
| Gemini CLI | P1 | `planned / researched` | 观察和手动配置候选 |
| Qwen Code | P1 | `planned / researched` | 观察和手动配置候选 |
| Kimi Code / Kimi CLI | P1 | `planned / researched` | 观察和手动配置候选 |
| OpenCode | P1 | `planned / researched` | 观察和手动配置候选 |

### 用户路径

```text
官网选择本地 MCP / CLI 客户端
 -> 用户复制官网提供的一句话到 AI 输入框，或复制命令到终端
 -> AI 在本机运行 setup 命令
 -> 如需打开浏览器授权、系统确认、客户端重启，AI 暂停并让用户自己操作
 -> 用户完成 PatchXNote 登录和授权
 -> AI 继续验证 MCP 工具列表
```

### AI setup prompt 模板

官网每个本地客户端详情页都提供一个可复制 prompt。以 Cursor 为例：

```text
请帮我在当前环境安装 PatchXNote MCP。请运行：
npx -y patchxnote-agent@latest setup --client cursor

如果需要打开浏览器、确认编辑器安装、登录 PatchXNote 或输入验证码，请暂停并让我自己完成。不要读取、保存或输出手机号、验证码、token、OAuth code、refresh token、webhook secret 或任何密钥。安装完成后，请刷新 MCP，并验证：List the PatchXNote MCP tools and show only their names.
```

Codex 示例：

```text
请帮我在当前 Codex 环境安装 PatchXNote MCP。请运行：
npx -y patchxnote-agent@latest setup --client codex

如果需要打开浏览器、登录 PatchXNote、输入验证码或确认授权，请暂停并让我自己完成。不要读取、保存或输出手机号、验证码、token、OAuth code、refresh token、webhook secret 或任何密钥。安装完成后，请新开或刷新 Codex 会话，并验证：List the PatchXNote MCP tools and show only their names.
```

Claude Code 示例：

```text
请帮我在当前 Claude Code 环境安装 PatchXNote MCP。请运行：
npx -y patchxnote-agent@latest setup --client claude-code

如果需要打开浏览器、登录 PatchXNote、输入验证码或确认授权，请暂停并让我自己完成。不要读取、保存或输出手机号、验证码、token、OAuth code、refresh token、webhook secret 或任何密钥。安装完成后，请刷新 MCP，并验证：List the PatchXNote MCP tools and show only their names.
```

### 命令与配置兜底

通用本地命令：

```sh
npx -y patchxnote-agent@latest setup --client <client-id>
```

通用 MCP 配置：

```json
{
  "mcpServers": {
    "patchxnote": {
      "command": "npx",
      "args": ["-y", "patchxnote-agent@latest", "mcp", "serve"]
    }
  }
}
```

关键边界：

- `mcp serve` 不在编辑器启动时自动弹浏览器，避免 MCP Host 启动超时。
- 登录应由 `mcp login` 或 `setup` 主动触发。
- 凭据只进 OS 原生安全存储，MCP 配置不放 token。
- 哪个运行时启动 MCP，就在哪个运行时登录和 setup。Windows、WSL、VS Code Remote、Dev Container、远端 Linux 的安全存储不能默认共享。

## 第三类：云平台入口

### 适用平台

| 平台 | 当前优先级 | 当前状态 | 官网首版动作 |
| --- | --- | --- | --- |
| Feishu Aily / Doubao Work Partner | P0.5 | `planned / researched` | `Connect remote MCP`，展示 remote MCP URL 和平台验收状态 |
| Tencent Agent Development Platform | P0.5 | `planned / researched` | `Connect remote MCP`，展示 remote MCP URL 和平台验收状态 |
| Enterprise WorkBuddy | P0.5 | `planned / researched` | 走企业/平台 remote MCP，不等同本地 WorkBuddy |
| 其他只能接 HTTP tool 或 webhook 的平台 | 后续观察 | `research` | 标记平台限制，另行评估 webhook / HTTP 方案 |

### 用户路径

```text
官网选择云平台
 -> 复制 remote MCP URL 或打开平台配置指引
 -> 在平台控制台添加 PatchXNote remote MCP
 -> 平台按 OAuth protected resource metadata 发现授权入口
 -> 用户在 PatchXNote 授权页登录和确认
 -> GoServer 创建可撤销 connector session
 -> 平台通过 HTTPS 调用 /mcp
 -> 验证 initialize、tools/list 和一个安全只读 tools/call
```

测试服 remote MCP URL：

```text
https://ws-lab.patch-x.cn/patchnote-test-api/mcp
```

正式域名候选：

```text
https://mcp.patchxnote.com/mcp
```

云平台关键边界：

- 不要求用户把 PatchXNote access token 粘贴到第三方平台。
- 能走平台官方 OAuth / remote MCP 的，优先走 remote MCP。
- 如果平台只支持 header/token，首版标记为平台限制，不把它包装成已闭环。
- 如果平台不支持 MCP，再单独评估 webhook、HTTP tool 或平台私有插件。
- 远程 MCP 不能访问用户本机文件系统，本地导出和 webhook 草稿类能力需要返回 bounded content 或短期 download handle。

## 官网页面信息架构

首页第一屏先建立产品识别和转化，不在首屏塞卡片墙。第二屏或 `Clients` 锚点再让用户选择工具：

```text
PatchXNote MCP
隐私归你，AI 由你掌控
连接我的 AI 工具
```

客户端选择区使用三段筛选：

| 筛选 | 展示对象 | 主动作 |
| --- | --- | --- |
| Editors | Cursor、VS Code、Windsurf、Trae、Qoder、WorkBuddy 本地、Zed、JetBrains | `Install in <Client>` 或 `Copy AI setup prompt` |
| Local MCP | Codex、Claude Code、Claude Desktop、Gemini CLI、Qwen Code、Kimi Code、OpenCode | `Copy AI setup prompt` 或 `Copy command` |
| Cloud Platforms | 飞书 Aily、豆包工作伙伴、腾讯 Agent、企业 WorkBuddy | `Connect remote MCP` |

点击任意卡片进入对应详情页。详情页使用同一个模板，只按渠道类型切换内容：

- 编辑器详情页：优先展示一键/AI assisted/setup command/manual config。
- 本地 MCP 详情页：优先展示复制 prompt、复制命令、验证提示。
- 云平台详情页：优先展示 remote MCP URL、给 AI 的一句话、官网授权。
- 通用配置详情页：展示 `mcpServers` block 和安全边界。

每个详情页固定模块：

1. Hero：客户端名、类型、状态标签、主动作。
2. Quick setup：一键安装、AI setup prompt 或 remote MCP URL。
3. Login：说明浏览器授权和用户确认边界。
4. Manual config：无密钥配置兜底。
5. Verify：安全验证提示。
6. Caveats：客户端专属限制。
7. Safety：token、验证码、原始内容和 provider payload 不进官网、配置和示例。

## 按钮状态规则

| 状态 | 官网标签 | 可展示动作 |
| --- | --- | --- |
| `researched` | Setup guide | 复制 AI prompt、复制命令、手动配置 |
| `implemented` | Setup command | 复制已实现命令 |
| `locally_smoked` | Local smoke passed | 可以提示本地链路通过，但不代表某个客户端 UI 通过 |
| `published_smoked` | Published package smoked | 可以说明发布包可启动和代理，但不代表客户端 UI 通过 |
| 客户端 deeplink/UI 验收通过 | One-click install | 可以显示 `Install in <Client>` 主按钮 |
| `platform_accepted` | Platform accepted | 云平台可以显示已验收连接 |

禁止口径：

- 不把“复制命令”写成“一键安装”。
- 不把“deeplink 格式可生成”写成“真实客户端已验收”。
- 不把“发布包可启动”写成“Cursor / VS Code / Codex 已验收”。
- 不把“远程 `/mcp` 可初始化”写成“飞书/豆包/腾讯/WorkBuddy 已平台验收”。
- 不展示或采集手机号、验证码、token、OAuth code、refresh token、webhook secret、原始转写、完整模型输入输出或 provider payload。

## GoServer 实现建议

官网实现放在 GoServer：

```text
/home/zsts_119/patchxNoteGoServer/web/mcp/
```

建议静态结构：

```text
web/mcp/
  index.html
  app.css
  app.js
  data/clients.json
  auth/
    authorize.html
    phone-login.html
    success.html
    error.html
    auth.js
```

路由建议：

```text
/mcp
/mcp/
/mcp/clients/{client-id}
/mcp/platforms/{platform-id}
/mcp/auth/login
/mcp/auth/success
/mcp/auth/error
/mcp/assets/*
/mcp/data/clients.json
```

数据源建议：

- 第一版把 `patchnote-agent/docs/mcp-clients/clients.json` 转成 GoServer `web/mcp/data/clients.json` 快照，并保留 `source_reviewed_at` 和 `source_repository` 字段。
- `web/mcp/data/clients.json` 不能写 mock 数据，字段只能来自客户端登记表、官方文档、本机验收记录或 GoServer 真实 remote MCP 配置。
- 后续再做脚本化同步，避免官网、CLI 和文档三套状态漂移。
- 官网构建时不依赖 sibling repo 的运行时路径。

授权页建议：

- `/v1/agent/oauth/authorize` 继续使用 GoServer OAuth authorize/token/revoke。
- 不新增一套官网登录系统。
- 手机号登录页、授权确认页、授权成功页、授权失败页都放进 GoServer `web/mcp/auth/`，视觉和 `web/mcp/` 共用设计 token 或同一份轻量 CSS。
- 用户未登录时先展示手机号登录；登录完成后继续展示授权确认页。
- 用户已登录时直接展示授权确认页。
- 修复错误展示，统一 API error envelope，不能显示 `[object Object]`。
- 本地 loopback callback 不能省掉，因为它要接收 OAuth code 并交给 CLI 换 token；但它收到 code 后可以跳转到官网 `/mcp/auth/success` 展示统一成功页。
- 授权成功页明确提示“授权信息已返回本机，请回到编辑器或终端等待最后连接验证”。
- loopback callback 失败、旧 tab 或旧端口失效时，跳转或提示用户回到终端或 AI 工具重新发起登录。

## 首版验收清单

- [ ] GoServer `/mcp` 首页可打开并展示三类入口。
- [ ] P0/P0.5 客户端卡片来自同一份 `clients.json` 快照。
- [ ] 每个本地客户端详情页都有 AI setup prompt、setup command、manual config 和 verify prompt。
- [ ] Cursor deeplink 只有真实安装页验收通过后才设为主按钮。
- [ ] VS Code、Codex、Claude Code、Claude Desktop 等分别保留官方命令或配置兜底。
- [ ] 云平台详情页展示 remote MCP URL、OAuth 口径、当前验收状态和下一步。
- [ ] 手机号登录页、授权确认页、授权成功页、授权失败页与官网视觉统一，不新增登录体系。
- [ ] 本地 callback 成功后跳转官网成功页，但 CLI 仍能拿到 code 并完成 token exchange。
- [ ] 授权页错误不会显示 `[object Object]`。
- [ ] 页面、示例、日志和验收记录不包含密钥或敏感内容。
- [ ] 前端移动端和桌面端不出现按钮文字溢出、遮挡或布局错位。
