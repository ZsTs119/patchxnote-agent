# PatchXNote 本地 MCP 与远程 MCP 交接文档

**日期：** 2026-09-01

**定位：** 给后续官网、安装入口、平台联调会话使用的事实入口。本文只沉淀当前边界和待办，不替代具体实现计划。

## 当前结论

PatchXNote MCP 现在分两条链路：

- 本地 MCP：面向 VS Code、Cursor、Codex、Claude Code、Claude Desktop、Windsurf、Trae、Qoder、WorkBuddy 等能在用户电脑上启动本地命令的编辑器或桌面 Agent。
- 远程 MCP：面向飞书 Aily / 豆包工作伙伴、腾讯 Agent 平台、企业 WorkBuddy 等平台型智能体，它们通常不能直接运行用户本机 `npx`，需要 HTTPS MCP 网关。

两条链路共享同一套 PatchXNote Agent 工具语义和服务端授权边界，不应该在官网文案里混成一个安装方式。

## 对外术语

官网、PRD、市场页和后续竞品调研建议统一使用这些术语：

- **Hybrid MCP distribution**：混合 MCP 分发。PatchXNote 同时提供本地 stdio MCP 和远程 hosted MCP。
- **Local stdio MCP server**：本地 stdio MCP。由编辑器或 Agent 在用户电脑上启动 `npx -y patchxnote-agent@latest mcp serve`，通过 stdin/stdout 传 JSON-RPC。
- **npm stdio launcher**：npm 启动器。`patchxnote-agent` npm 包负责安装/校验平台原生 `patchxnote` 二进制，然后代理到 `patchxnote mcp serve`。
- **Local MCP bridge / stdio-to-remote proxy**：本地桥接层。用户完成 `mcp login` 后，本地 stdio MCP 可以把请求代理到 GoServer 远程 `/mcp`，同时保留旧本地 Agent fallback。
- **Browser OAuth with PKCE and loopback callback**：浏览器 OAuth + PKCE + 本机回调。`patchxnote mcp login` 先打开 GoServer 授权页，再由 `127.0.0.1:<port>/callback` 接收 authorization code。
- **OS-native secure storage**：OS 原生安全存储。Windows Credential Manager、macOS Keychain、Linux Secret Service/file fallback。MCP config 不放 token。
- **Hosted remote MCP server / remote MCP gateway**：托管远程 MCP。平台型智能体通过 HTTPS 访问 GoServer `/mcp`，不依赖用户本机 `npx`。
- **Streamable HTTP transport**：远程 MCP 的主推传输方向。若目标平台只支持 SSE，则作为兼容分支处理。
- **OAuth protected MCP resource**：需要 OAuth 保护的 MCP 资源。远程 MCP 应通过 metadata 暴露授权服务器位置，客户端按 OAuth 流程拿 token。
- **Connector session**：连接器会话。平台或本地 MCP 登录成功后，对某个账号、客户端、scope 可撤销的授权记录。
- **Client registry**：客户端登记表。官网卡片、详情页、CLI `setup --client`、验收状态共用的事实源。
- **Install entry / setup command / deeplink / marketplace listing**：安装入口。第一版优先 setup command 和无密钥 config；deeplink、插件市场、平台 marketplace 必须等真实验收后再标为一键。

一句话定义当前技术方向：

```text
PatchXNote MCP 是 Hybrid MCP distribution：本地用 npm stdio launcher + browser OAuth + OS-native secure storage 接入桌面编辑器；平台侧用 GoServer-hosted remote MCP gateway + OAuth connector session 接入云端/企业智能体。
```

## 本地 MCP 事实

本地 MCP 的用户路径是：

```text
用户运行 mcp login 或 setup
 -> 本机 patchxnote 启动临时 127.0.0.1 callback
 -> 浏览器打开 GoServer 授权页
 -> 用户在网页输入手机号和验证码
 -> GoServer OAuth authorize 返回 code 到本机 callback
 -> patchxnote 换取 token 并保存到 OS 原生安全存储
 -> 编辑器通过 npx -y patchxnote-agent@latest mcp serve 启动 stdio MCP
```

关键口径：

- `127.0.0.1:<port>/callback` 是桌面 OAuth loopback 回调，不是官网页面。正常成功时它只负责接收 authorization code 并校验 state，然后跳转到 GoServer 官网风格的 `/mcp/auth/success` 成功页。
- `mcp serve` 启动时不自动弹浏览器。缺少凭据时应提示用户先运行 `mcp login`，避免编辑器 MCP Host 启动超时。
- MCP 配置文件必须无密钥，不出现手机号、验证码、access token、refresh token、webhook secret。
- 凭据跟运行时绑定。Windows 桌面、WSL、VS Code Remote、Dev Container、远端 Linux 不能默认共享同一份安全存储。

当前状态：

- `patchxnote-agent` 已产品化正式 `mcp login/status/logout`，替代临时 Node 验收脚本。
- `patchxnote-agent@0.2.9` 已发布到 npm 和 GitHub Releases。
- 发布包 install、`mcp config`、clean-profile `mcp status`、`setup --client cursor --dry-run --print-config`、Windows 安装、Linux checksum、macOS 安装/MCP smoke 均已通过。
- `0.2.9` 本地源码候选已完成真实浏览器 OAuth 登录、latest OTP request 生效、loopback callback、`mcp status --verify` 和远程 MCP 读工具验证。
- Fresh registry 包浏览器 OAuth 登录仍需单独重测：发布后的 clean-profile OTP 登录没有在本次 npm closeout 中重复完成。

相关文件：

- `README.md`
- `README.zh-CN.md`
- `docs/evidence/2026-09-01-mcp-oauth-local-acceptance.zh-CN.md`
- `docs/evidence/2026-09-01-release-0.2.9.zh-CN.md`
- `docs/plans/2026-08-27-mcp-browser-oauth-login-productization-checklist.md`

## 远程 MCP 事实

远程 MCP 的用户路径是：

```text
用户在官网或平台控制台选择平台客户端
 -> 用户完成 PatchXNote 授权
 -> 平台拿到可调用远程 MCP 的连接配置或连接会话
 -> 平台通过 HTTPS 调用 GoServer /mcp
 -> GoServer 按 connector session / OAuth / Agent access 边界映射到当前账号
 -> 平台调用同一套 PatchXNote MCP 工具
```

当前远程入口：

```text
https://ws-lab.patch-x.cn/patchnote-test-api/mcp
```

长期可包装为独立域名：

```text
https://mcp.patchxnote.com/mcp
```

关键口径：

- 平台型客户端不应要求用户把 PatchXNote access token 粘贴进第三方平台。
- 能走平台官方 OAuth / remote MCP 的，优先走远程 MCP。
- 不能走 remote MCP 的，记录为平台限制；如业务必须接入，再评估 webhook / HTTP tool / 平台私有插件方案。
- 远程 MCP 不能访问用户本机文件系统，render/export/webhook 等本地能力需要服务端持久化或返回 bounded content/download handle。

相关文件：

- `docs/plans/2026-08-27-remote-mcp-platform-gateway-design.md`
- `docs/plans/2026-08-27-remote-mcp-goserver-parity-checklist.md`
- `docs/evidence/2026-08-27-platform-client-poc-status.zh-CN.md`

## 同类产品官网调研入口

以下页面已在 2026-08-27 核验，可作为官网信息架构和安装入口参考。不要直接照抄文案、视觉或商标资产。

| 产品/页面 | 可学习点 | 对 PatchXNote 的启发 |
| --- | --- | --- |
| [1Server Clients](https://1server.ai/clients/) | 客户端卡片网格、`Use ... in every MCP client`、每个客户端详情页、通用 `mcpServers` fallback。 | 首页可以先让用户选择正在用的 AI 工具；详情页再给 setup command、config fallback、限制说明。 |
| [Context7 Install](https://context7.com/install) | `One command` + `Pick your agent`，把 Claude Code、Cursor、Codex 等入口压得很短。 | PatchXNote 可以把第一屏主路径做成 `npx -y patchxnote-agent@latest setup --client <id>`，降低文档感。 |
| [Context7 MCP Clients](https://context7.com/docs/resources/all-clients) | 每个客户端有本地连接配置，也有 VS Code 扩展自动注册路径。 | 我们也要把“自动安装”和“手动配置”分开，不能未验收就写一键。 |
| [Composio MCP for VS Code](https://composio.dev/toolkits/composio/framework/vscode) | 单客户端详情页包含 one-click、manual config、authorize、authenticate 四段。 | PatchXNote 每个详情页也按安装、登录授权、验证、故障恢复来写。 |
| [Zapier MCP](https://zapier.com/mcp) | 面向普通用户强调“AI action bridge”、平台无关、企业治理、无终端/无配置的引导。 | 我们官网可以强调“让 AI 读 PatchXNote 记录和总结”，但第一版不能承诺无终端，除非 deeplink/插件真实闭环。 |
| [Zapier Remote MCP Client](https://help.zapier.com/hc/en-us/articles/38777069364109-Connect-remote-MCP-servers-to-Zapier-using-MCP-Client) | 远程 MCP 表单明确 Server URL、Transport、OAuth、Bearer Token。 | 平台型客户端页应明确 remote URL、transport、OAuth；不鼓励用户粘贴 PatchXNote token。 |
| [Pipedream MCP Developers](https://pipedream.com/docs/connect/mcp/developers) | Remote MCP + 内置用户授权 + behalf-of 用户工具调用。 | 远程 MCP 需要把“代表用户调用”讲清楚：不是平台公共 token，而是用户授权的 connector session。 |
| [Smithery](https://smithery.ai/) | MCP marketplace、server 详情、Setup snippet、Tools/Resources/Prompts 展示。 | PatchXNote 详情页可以展示工具能力分组，但保持工具数量克制，不做 full OpenAPI 暴露。 |
| [21st MCP](https://github.com/21st-dev/magic-mcp) | CLI `init --client` 覆盖 Cursor/Claude/VS Code/Windsurf/Codex，并提供 HTTP MCP fallback。 | 我们的 `setup --client` 路线是对的；远程 HTTP fallback 要和本地 stdio 写清楚。 |
| [DevUtils MCP Server](https://mcpservers.org/servers/paladini/devutils-mcp-server) | one-click/plugin、npx、npm、不同客户端分段安装。 | 官网可以按“推荐安装 / 手动安装 / 插件市场”三层展示。 |
| [VS Code MCP docs](https://code.visualstudio.com/docs/agent-customization/mcp-servers) | 官方支持 MCP gallery、用户/工作区配置、`code --add-mcp`、trust prompt、remote/dev container 差异。 | VS Code 页必须区分 user profile、workspace、remote/dev container；自动写配置要提示 trust。 |
| [Cursor MCP docs](https://cursor.com/docs/mcp) 和 [Cursor install links](https://cursor.com/docs/mcp/install-links) | Cursor 支持 `mcp.json`、Customize 页面、deeplink install。 | Cursor 可以优先做 deeplink，但必须保留 setup command fallback。 |
| [OpenAI Codex MCP docs](https://developers.openai.com/codex/mcp) | Codex 本地客户端和 ChatGPT desktop/IDE 共享 MCP 配置。 | Codex 页要给 `codex mcp add` / `config.toml` 路径，也要规划后续 Codex plugin/ChatGPT plugin 发现入口。 |

## 官网需要表达的产品分层

首页不要只写“支持 MCP”。更清楚的分层是：

- **Use PatchXNote in local AI editors**：给 VS Code、Cursor、Codex、Claude Code、Claude Desktop、Windsurf、Trae、Qoder、WorkBuddy。
- **Connect PatchXNote to platform agents**：给飞书 Aily / 豆包工作伙伴、腾讯 Agent 平台、企业 WorkBuddy。
- **One login, secret-free config**：本地 token 存 OS 安全存储，MCP 配置不含密钥。
- **Same PatchXNote tools everywhere**：本地和远程尽量工具语义一致，但安装/授权/文件能力不同。
- **Read-first Agent boundary**：强调读 PatchXNote 记录、总结、AI 结果；不要把硬件绑定、支付、Admin、模型执行写进去。

## 还没做的三件事

### 1. GoServer `web/` 官网与授权页统一

- [ ] 在 GoServer `web/` 下做产品化官网，不放在 `patchnote-agent` 仓库。
- [ ] 官网和 MCP OAuth 授权页使用同一套前端体验与品牌样式。
- [ ] 首页采用客户端卡片网格，而不是文档目录。
- [ ] 卡片点击进入客户端详情页，详情页展示安装入口、配置片段、验证提示和客户端注意事项。
- [ ] 登录授权页继续复用 GoServer OAuth authorize/token/revoke，不新增一套独立登录体系。
- [ ] 修复授权页错误展示：统一 API 错误 envelope 不能显示成 `[object Object]`。
- [ ] 授权成功页要明确告诉用户“可以回到编辑器继续使用”。
- [ ] 本地 loopback callback 成功接收 code 后跳转到 GoServer `/mcp/auth/success`；callback 本身不再作为最终展示页。
- [ ] loopback callback 失败或旧 tab/旧端口失效时，要给用户可理解的重试提示。
- [ ] 官网页面应复用 `docs/mcp-clients/clients.json` 或等价数据源，避免官网、CLI、文档三套状态漂移。
- [ ] 外链调研页面只作为参考，不复制对方商标、文案、视觉和 screenshots。

### 2. 逐个补客户端安装入口

- [ ] VS Code / GitHub Copilot。
- [ ] Cursor。
- [ ] Codex / ChatGPT Desktop / Codex IDE。
- [ ] Claude Code。
- [ ] Claude Desktop。
- [ ] Windsurf。
- [ ] Trae / Trae CN / TraeWork Code。
- [ ] Qoder。
- [ ] WorkBuddy / Tencent CodeBuddy WorkBuddy 本地模式。

安装入口分层：

- `setup command`：第一版最稳，复制 `npx -y patchxnote-agent@latest setup --client <id>`。
- `config fallback`：复制无密钥 MCP 配置。
- `deeplink`：只有官方格式和真实客户端验收通过后，才能作为“一键安装”按钮。
- `marketplace/plugin`：单独作为发现渠道，不要和 MCP 可用性混淆。

首版安装入口建议：

| 客户端 | 官网主按钮 | 备用入口 | 验收前不能说 |
| --- | --- | --- | --- |
| VS Code | Copy setup command | `code --add-mcp` / `.vscode/mcp.json` / user profile config | 未验收前不说 marketplace 已上架。 |
| Cursor | Copy setup command | Cursor deeplink / `~/.cursor/mcp.json` | deeplink 未真机验证前不说一键安装。 |
| Codex | Copy setup command | `codex mcp add patchxnote -- npx -y patchxnote-agent@latest mcp serve` / `~/.codex/config.toml` | 未做插件前不说 Codex 插件市场可搜。 |
| Claude Code | Copy setup command | `claude mcp add` / `.mcp.json` | 不把 Claude Desktop 配置当作 Claude Code 已验收。 |
| Claude Desktop | Copy setup command | desktop config JSON / 后续 `.mcpb` | `.mcpb` 未做前不说桌面扩展一键装。 |
| Windsurf | Copy setup command | Windsurf config path | 未核对当前 schema 前不自动写。 |
| Trae / TraeWork | Manual setup command | 客户端配置页截图/说明 | 不把 Trae、Trae CN、TraeWork 混为一个验收。 |
| Qoder | Manual setup command | Qoder deeplink 或配置文件 | deeplink 未确认前只写 planned。 |
| WorkBuddy 本地 | Manual setup command | WorkBuddy MCP/CLI 设置页 | 不把企业平台 WorkBuddy 写成本地闭环。 |

每个客户端都要单独记录：

- 当前支持状态。
- 配置路径或官方命令。
- 是否能自动写配置。
- 是否需要用户确认和备份。
- 是否完成真实客户端 UI 验收。
- 是否完成发布包验收。

### 3. 平台型客户端真实验收

- [ ] 飞书 Aily / 豆包工作伙伴。
- [ ] 腾讯 Agent 平台。
- [ ] 企业 WorkBuddy。
- [ ] 其他国内外平台型 Agent，如果只能配置 HTTPS tool 或 webhook，也要单独记录。

验收路径：

- [ ] 平台能配置 remote MCP：接 `https://ws-lab.patch-x.cn/patchnote-test-api/mcp` 或正式域名。
- [ ] 平台能走 OAuth：按平台官方强调方式接授权。
- [ ] 平台只能配置 header/token：第一版不要让用户粘贴 PatchXNote access token，先标记为限制。
- [ ] 平台不能接 MCP：记录为平台限制，评估 webhook/HTTP 方案。
- [ ] 每个平台至少验收 `initialize`、`tools/list`、一个安全读工具。
- [ ] 只记录工具名、数量、状态码、request id、脱敏账号状态；不记录手机号、验证码、token、原始转写、完整模型输入输出或 provider payload。

平台页首版需要准备这些字段：

- 平台名称和平台类型。
- 支持 transport：Streamable HTTP / SSE / unknown。
- 支持 auth：OAuth / Bearer token / custom header / unknown。
- 远程 MCP URL。
- 是否能自动发现 OAuth protected resource metadata。
- 是否需要平台后台白名单、应用审核、企业管理员授权。
- 当前验收状态和下一步动作。

## 后续需要做到什么

- [ ] 官网上线前：GoServer `web/` 提供产品化客户端选择页和详情页。
- [ ] Fresh published package OAuth：用 `patchxnote-agent@0.2.9` clean profile 完成浏览器登录、callback 成功页、`mcp status --verify`、`mcp serve` 工具调用。
- [ ] VS Code：验证 `setup --client vscode`、manual `.vscode/mcp.json`、用户 profile config、`code --add-mcp` 至少一种真实路径。
- [ ] Cursor：验证 setup command 和 deeplink 两条路径。
- [ ] Codex：验证 `codex mcp add` / `~/.codex/config.toml`，后续评估 plugin 发现入口。
- [ ] Claude Code / Claude Desktop：分别验证 CLI add 和 Desktop config；后续评估 `.mcpb`。
- [ ] Windsurf / Trae / Qoder / WorkBuddy：先确认当前官方配置 schema，再决定自动写或手动。
- [ ] 平台远程 MCP：分别在飞书 Aily / 豆包工作伙伴、腾讯 Agent 平台、企业 WorkBuddy 上验收 remote MCP 或记录平台限制。
- [ ] 官网状态系统：每个按钮绑定 `researched / implemented / locally_smoked / published_smoked / platform_accepted`，避免过度承诺。
- [ ] 商标和素材：上线前确认客户端 Logo 使用规范；不确定时用文字或自有抽象图标。

## 官网新会话建议入口

新会话写官网时，先读这些文件：

```text
docs/mcp-local-remote-handoff.zh-CN.md
docs/mcp-clients/README.zh-CN.md
docs/mcp-clients/clients.json
docs/mcp-clients/website/README.zh-CN.md
docs/mcp-clients/website/01-information-architecture.zh-CN.md
docs/mcp-clients/website/02-entry-action-model.zh-CN.md
docs/mcp-clients/website/03-client-install-sources.zh-CN.md
docs/mcp-clients/website/04-visual-system.zh-CN.md
docs/mcp-clients/website/05-reference-skill-research.zh-CN.md
docs/mcp-clients/website/06-implementation-readiness.zh-CN.md
docs/mcp-clients/website/07-acceptance-checklist.zh-CN.md
docs/mcp-clients/client-detail-copy.zh-CN.md
docs/evidence/2026-09-01-release-0.2.9.zh-CN.md
docs/evidence/2026-09-01-mcp-oauth-local-acceptance.zh-CN.md
../patchxNoteGoServer/docs/requirements/cloud-config-center/new_current
```

官网第一版不要承诺所有客户端都“一键安装”。更稳的表达是：

- 本地客户端：`Install locally` / `Setup command`。
- 已验证 deeplink：`One-click install`。
- 平台客户端：`Connect remote MCP`。
- 未验收平台：`Platform acceptance pending`。

## 不要改错的边界

- 不把 GoServer OAuth 页面搬到 `patchnote-agent`。
- 不把 token 写进 MCP config、官网 URL、安装命令、文档示例或截图。
- 不把平台远程 MCP 写成本地 `npx` 安装。
- 不把本地 stdio MCP 写成云端平台自动可用。
- 不把“发布包可启动”写成“某个编辑器已验收”。
- 不把“远程 `/mcp` 可初始化”写成“飞书/豆包/腾讯/WorkBuddy 已平台验收”。
