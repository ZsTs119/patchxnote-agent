# PatchXNote MCP 官网开写前流程缺口清单

**日期：** 2026-08-28

**定位：** 本文用于回答“什么时候可以开始写 GoServer `web/mcp/` 官网页面”。它承接三类入口动作方案、视觉系统、参考仓库选型和页面规格，专门收束工程实现前还缺哪些流程。

## 已确认

- 官网第一版放在 GoServer：

```text
/home/zsts_119/patchxNoteGoServer/web/mcp/
```

- V1 不新建 Vue 项目，不引入 Next.js / React / Tailwind 构建链路。先用原生 `index.html`、`app.css`、`app.js` 和数据 JSON。
- 官网主体、手机号登录页、MCP 授权确认页、授权成功页和授权失败页都归入 GoServer `web/mcp/` 官网体验。授权页由 GoServer 动态路由返回，但源文件和视觉资产放在 `web/mcp/auth/`。
- 授权页继续复用 GoServer 现有 Agent OAuth / OTP 接口，不新增独立官网登录系统。
- 本地 MCP 登录仍需要 `127.0.0.1:<port>/callback` 接收 OAuth code。这个 callback 只做协议回传和状态校验，收到结果后跳转到官网风格的 `/mcp/auth/success` 或 `/mcp/auth/error` 展示页。
- 本地 MCP 配置不放 token。登录凭据只进入 OS 原生安全存储。
- 一键安装只在对应客户端 deeplink / 官方入口完成真实验收后才作为主按钮。
- 云平台走 remote MCP + OAuth connector session，不把本地 `npx` 当成云平台主路径。

## 可以先开写的范围

这些不需要等待所有客户端验收完成，可以先做：

- `/mcp` 首页视觉骨架。
- 三类入口 tab：`编辑器`、`云平台`、`本地 MCP`。
- 客户端卡片列表、筛选、状态标签。
- `找不到你的渠道？` 通用 MCP 配置兜底入口。
- 客户端详情视图：`AI assisted`、`One-click`、`Manual config` 三个 setup tab。
- 复制 AI setup prompt、setup command、manual config、remote MCP URL。
- toast、copy success、按钮 hover / active / disabled / coming soon 状态。
- 手机号登录页、授权确认页、授权成功页、授权失败页的官网统一视觉和错误展示修复方案。
- 使用自有抽象图标或文字占位，等正式素材生成后替换。

不能先做成“已闭环承诺”的范围：

- 未验收客户端的一键安装。
- 未验收云平台的 `platform_accepted`。
- Claude Desktop `.mcpb` 桌面扩展。
- 平台私有插件、webhook 或 HTTP tool 替代方案。

## 开写前必须补齐的流程

### 1. GoServer 静态资源嵌入与路由流程

当前 GoServer 是按页面显式 embed，不是自动嵌入整个 `web/` 目录。MCP 官网需要新增一组和下载页类似的嵌入与路由：

```text
embedded_web.go
  MCPPageFS()

internal/platform/httpapi/router.go
  Options.MCPPageFS
  registerMCPPage(router, pageFS)
```

已确认路由：

```text
/mcp                  -> 直接进入官网首页，可 301 到 /mcp/ 或直接服务 index
/mcp/                 -> index.html
/mcp/app.css          -> app.css
/mcp/app.js           -> app.js
/mcp/data/clients.json
/mcp/assets/*
/mcp/clients/{id}     -> index.html，由前端按 path 渲染详情
/mcp/platforms/{id}   -> index.html，由前端按 path 渲染详情
/mcp/clients/generic-mcp -> index.html，由前端渲染通用配置兜底
```

还要补：

- 路径清洗，禁止 `..`、反斜杠、空字节、目录结尾。
- content type：HTML、CSS、JS、JSON、SVG、PNG、JPG、WEBP。
- 缓存策略：首版 staging 可 `no-store`；稳定后 index 保持短缓存或 no-store，assets 可加版本化长缓存。
- `route_not_found` 与 API 错误 envelope 保持 GoServer 统一口径。

### 2. 官网数据快照与 schema 流程

官网不要直接在运行时读取 sibling repo。第一版应把 `patchnote-agent/docs/mcp-clients/clients.json` 转成 GoServer 内部快照：

```text
web/mcp/data/clients.json
```

这份快照不能是 mock。字段只能来自：

- `docs/mcp-clients/clients.json`
- `docs/mcp-clients/website/03-client-install-sources.zh-CN.md`
- 本机真实客户端验收记录
- GoServer 真实 remote MCP 配置

快照需要包含：

- `source_repository`
- `source_reviewed_at`
- `generated_at`
- `default_server`
- `remote_gateway`
- `clients[]`

每个 `client` 至少包含：

- `id`、`name`、`category`、`priority`、`regions`
- `support_status`、`evidence_state`
- `transport`
- `runtime_caveat`
- `primary_action`
- `setup_command`
- `ai_setup_prompt`
- `manual_config`
- `deeplink` 或 `deeplink_capability`
- `remote_mcp_url`
- `verify_prompt`
- `caveats`
- `official_sources`
- `references`

后续可以写同步脚本，但 V1 可以先手工快照。关键是页面只读快照，不在多个地方手写同一份状态。

没有真实来源的字段不要填假值。确实暂缺时，使用 `unknown`、`pending` 或直接省略，并在页面上显示为待验收/待确认。

### 3. 前端无框架渲染流程

V1 的 `app.js` 建议拆成清晰模块，即使不使用 Vue，也要保持可维护：

```text
loadData()
parseRoute()
renderHome()
renderClientDetail(client)
renderPlatformDetail(platform)
renderGenericMCP()
buildActions(client)
copyToClipboard(text)
openDeeplink(url)
showToast(message, state)
setSetupTab(tab)
```

前端状态只需要：

- 当前 category tab。
- 当前搜索关键词。
- 当前路由对象。
- 当前详情页 setup tab。
- copy/toast 的临时 UI 状态。
- reduced motion 检测。

不要在首版做复杂 SPA 状态管理，也不要引入路由库。

### 4. 安装动作生成流程

每个客户端详情页都要按同一套动作策略生成按钮：

| 条件 | 主动作 |
| --- | --- |
| deeplink 已真实验收 | `Install in <Client>` |
| 本地客户端未验收 deeplink | `Copy AI setup prompt` |
| CLI / terminal agent | `Copy AI setup prompt` 或 `Copy command` |
| 云平台 | `Copy remote MCP URL` / `Open platform guide` |
| 只在研究中 | `Coming soon` + 复制手动信息 |

每个本地客户端至少保留三层 fallback：

1. `AI assisted`：复制一句话给 AI，让 AI 运行 setup/login/config。
2. `Setup command`：复制 `npx -y patchxnote-agent@latest setup --client <id>`。
3. `Manual config`：复制无密钥 MCP 配置或官方 CLI 命令。

通用兜底入口提供：

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

Cursor / VS Code 这类官方 deeplink 要单独生成并单独验收。只要没验收，就只能作为次级或 disabled 状态，不写“一键安装已支持”。

### 5. 授权页与官网统一流程

授权页现状在 GoServer：

```text
internal/agentoauth/authorize_page.go
```

现有接口流：

```text
POST /v1/agent/auth/otp/requests
POST /v1/agent/auth/otp/verifications
GET  /v1/agent/oauth/authorize
  Authorization: Bearer <agent access token>
  X-PatchXNote-OAuth-Response: redirect_json
```

V1 建议：

- 手机号登录页源文件放到 `web/mcp/auth/phone-login.html`，授权确认页源文件放到 `web/mcp/auth/authorize.html`，脚本放到 `web/mcp/auth/auth.js`，样式复用 `web/mcp/app.css` 或拆出 `web/mcp/auth/auth.css`。
- GoServer `/v1/agent/oauth/authorize` 未登录时先返回手机号登录页，登录完成后继续进入授权确认页。
- 不引 CDN、不引 Vue、不引外部图片，保持 CSP 简单。
- 视觉 token 和官网保持一致：高级黑、黑铬按钮、钛银边框、冷银状态线、8px 左右圆角。
- 错误处理统一解析 GoServer error envelope，禁止显示 `[object Object]`。
- 授权成功页和授权失败页也放在 `web/mcp/auth/`，由官网统一渲染。
- 旧 tab、旧 callback 端口、授权过期时给可理解的重试提示。
- 验证码、手机号、token、OAuth code、refresh token 不进入页面日志、URL 展示、埋点或示例。

本地 callback 说明：

```text
patchxnote-agent mcp login
 -> 启动 127.0.0.1:<port>/callback
 -> 浏览器打开 GoServer authorize 页面
 -> 用户完成手机号/验证码登录和授权
 -> GoServer redirect 到 127.0.0.1:<port>/callback?code=...&state=...
 -> 本地 callback 校验 state 并把 code 交给 CLI
 -> 浏览器跳转到 GoServer /mcp/auth/success
 -> CLI 继续换 token 并保存到 OS 原生安全存储
```

这里不能省掉本地 callback，因为本地 CLI 必须拿到 OAuth code 才能完成 token exchange 和安全存储。可以省掉的是 callback 自己那一页简陋 HTML，把用户最终看到的成功/失败页统一跳回官网。

### 6. 素材生成与占位流程

开写页面前不必等全部视觉素材完成，但要先定素材槽位：

| 槽位 | V1 可用占位 | 正式素材 |
| --- | --- | --- |
| Hero product visual | CSS/HTML 构成的 command hub + 产品剪影占位 | 基于用户硬件图生成的高级黑产品渲染 |
| Setup preview | 静态代码窗口和状态灯 | 可加入 MCP 工具连接动画 |
| Client icons | 文字首字母 / 自有抽象图标 | 上线前确认商标授权或使用规范 |
| Cloud platform visual | remote MCP URL / OAuth flow 图形 | 平台验收后补真实状态图 |
| OG image | 暂不阻塞 | 生成 1200x630 官方分享图 |

素材规则：

- 用户给的浅蓝产品信息图只作为事实参考，不直接贴到官网。
- 生成素材要保留硬件形态和 PatchX 标识，不夸大硬件外观。
- 图片尺寸按页面实际显示尺寸压缩，不提交超大原图。
- 每个关键图都要有 alt 文案。

### 7. 验收与安全扫描流程

开写前要定好验收口径，避免页面写完后不知道怎么算完成：

前端验收：

- Chrome 桌面宽屏。
- 1366 宽度桌面。
- iPhone 宽度。
- 按钮文字不换行、不溢出。
- 卡片、代码块、toast、tabs 不重叠。
- reduced motion 可关闭主要动画。

GoServer 验收：

- `/mcp`、`/mcp/`、`/mcp/app.css`、`/mcp/app.js`、`/mcp/data/clients.json`、`/mcp/assets/*` 可访问。
- `/mcp/clients/cursor`、`/mcp/clients/generic-mcp`、`/mcp/platforms/feishu-aily`、`/mcp/auth/login`、`/mcp/auth/success`、`/mcp/auth/error` 能返回对应官网页面。
- 非法路径返回统一 `route_not_found`。
- 授权页 OTP、verify、redirect_json 流程不回归。

安全扫描：

- 页面源码和数据 JSON 不包含手机号、验证码、access token、refresh token、OAuth code、webhook secret。
- 示例 MCP 配置不包含 bearer token。
- 不采集用户配置内容或账号敏感信息。
- 外链加 `rel="noopener noreferrer"`。

基础命令：

```sh
go test ./...
node docs/mcp-clients/validate-clients.mjs
git diff --check
```

如果只改 GoServer 官网静态页，GoServer 侧还要补静态资源路由测试；如果只改 `patchnote-agent` 文档，则不需要 GoServer runtime 测试。

## 已确认的产品决策与实现默认值

这些作为 V1 官网实现输入：

1. 首版语言：中文为主，保留 PatchXNote、MCP、AI Agent 等英文技术词；后续再加英文站。
2. URL 形式：保留 `/mcp/clients/{id}` 和 `/mcp/platforms/{id}` 可分享路径。
3. 上线方式：先本地和测试环境验收；生产上线前再单独确认域名、缓存和 noindex。
4. 正式 remote MCP 域名：页面先从 GoServer 配置/快照读取；没有正式域名前，不把测试 URL 包装成生产入口。
5. 客户端 logo：首版用自有抽象图标或文字符号，避免未确认商标使用规范；真实 logo 后置确认。
6. Hero 正式素材：基于用户提供的硬件图和产品信息生成高级黑风格素材，不直接使用浅蓝信息图。
7. 授权页归属：手机号登录页、授权确认页、授权成功页、授权失败页都放入 GoServer `web/mcp/auth/`，视觉与官网一致。
8. 客户端数据：只展示真实登记表、官方文档、现有代码能力和本机验收结果，不写 mock 数据。

## 开写判断

满足下面条件，就可以进入页面实现：

- [x] 确认 V1 使用原生 HTML/CSS/JS。
- [x] 确认 GoServer `/mcp`、`/mcp/clients/{id}`、`/mcp/platforms/{id}` 路由方案。
- [x] 确认 `web/mcp/data/clients.json` 快照字段，并要求所有字段来自官方文档、现有登记表或真实验收记录。
- [x] 确认授权页源文件放入 GoServer `web/mcp/auth/`，由 GoServer OAuth 路由读取并返回。
- [x] 确认一键安装按钮只按真实验收状态开启。
- [x] 确认首版可以用自有占位/生成素材，不直接使用用户浅蓝信息图。
- [x] 确认选择页使用三类 tab 筛选卡片，点击卡片进入详情页。
- [x] 确认通用 MCP 配置作为 `找不到你的渠道？` 兜底入口。
- [x] 确认手机号登录页是授权流程前置页。

我的建议是：上述条件已经满足。下一步可以先做 GoServer 静态页面骨架、授权页 web 资产、数据快照和路由 embed，再继续推进视觉素材生成与客户端一键安装验收。
