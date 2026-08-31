---
version: alpha
name: PatchXNote MCP Website
colors:
  background: "#07090A"
  backgroundElevated: "#0D1012"
  surface: "#121619"
  surfaceDeep: "#090C0E"
  border: "rgba(220, 230, 235, 0.13)"
  borderStrong: "rgba(225, 231, 234, 0.46)"
  text: "#F4F7F8"
  textMuted: "#A5B0B5"
  textSoft: "#6F7B82"
  metal: "#D8DEE1"
  metalDim: "#7D858A"
  accent: "#E5EAEC"
  accentSecondary: "#BFC7CB"
  accentWarm: "#C8BFAF"
  warning: "#C8BFAF"
  danger: "#B86A62"
typography:
  display:
    fontFamily: "Inter, ui-sans-serif, -apple-system, BlinkMacSystemFont, Segoe UI, PingFang SC, Hiragino Sans GB, Microsoft YaHei, sans-serif"
    fontSize: 64px
    fontWeight: 650
    lineHeight: 1.02
    letterSpacing: 0
  h1:
    fontFamily: "Inter, ui-sans-serif, -apple-system, BlinkMacSystemFont, Segoe UI, PingFang SC, Hiragino Sans GB, Microsoft YaHei, sans-serif"
    fontSize: 48px
    fontWeight: 650
    lineHeight: 1.08
    letterSpacing: 0
  h2:
    fontFamily: "Inter, ui-sans-serif, -apple-system, BlinkMacSystemFont, Segoe UI, PingFang SC, Hiragino Sans GB, Microsoft YaHei, sans-serif"
    fontSize: 32px
    fontWeight: 620
    lineHeight: 1.16
    letterSpacing: 0
  h3:
    fontFamily: "Inter, ui-sans-serif, -apple-system, BlinkMacSystemFont, Segoe UI, PingFang SC, Hiragino Sans GB, Microsoft YaHei, sans-serif"
    fontSize: 20px
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: 0
  body:
    fontFamily: "Inter, ui-sans-serif, -apple-system, BlinkMacSystemFont, Segoe UI, PingFang SC, Hiragino Sans GB, Microsoft YaHei, sans-serif"
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.65
    letterSpacing: 0
  caption:
    fontFamily: "Inter, ui-sans-serif, -apple-system, BlinkMacSystemFont, Segoe UI, PingFang SC, Hiragino Sans GB, Microsoft YaHei, sans-serif"
    fontSize: 13px
    fontWeight: 500
    lineHeight: 1.5
    letterSpacing: 0
  code:
    fontFamily: "JetBrains Mono, SFMono-Regular, Consolas, Liberation Mono, Menlo, monospace"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.7
    letterSpacing: 0
rounded:
  sm: 4px
  md: 8px
  pill: 999px
spacing:
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 32px
  sectionDesktop: 112px
  sectionMobile: 72px
---

# PatchXNote MCP Website DESIGN.md

## Overview

本文件是 PatchXNote MCP 官网给 AI coding agent 使用的视觉事实源。它把 `04-visual-system.zh-CN.md` 中已经确认的方向压缩成可执行 token、组件约束和验收边界，避免后续实现时重新发散。

官网视觉方向固定为 `Black Chrome Command Hub`：高级黑、黑铬、石墨、冷银、钛银、实体硬件可信度、AI 工具接入感和安全控制感。它不是普通深色文档站，不是硬件电商页，也不是霓虹赛博朋克页面。

阅读顺序：

1. `docs/mcp-clients/website/README.zh-CN.md`
2. `docs/mcp-clients/website/DESIGN.md`
3. `docs/mcp-clients/website/01-information-architecture.zh-CN.md`
4. `docs/mcp-clients/website/02-entry-action-model.zh-CN.md`
5. `docs/mcp-clients/website/03-client-install-sources.zh-CN.md`
6. `docs/mcp-clients/website/04-visual-system.zh-CN.md`
7. `docs/mcp-clients/website/06-implementation-readiness.zh-CN.md`
8. `docs/mcp-clients/website/07-acceptance-checklist.zh-CN.md`
9. `docs/mcp-clients/clients.json`

如果视觉灵感和产品事实冲突，产品事实、安装状态、授权安全和验收口径优先。

## Colors

颜色比例：

- 黑、石墨、炭灰占 `82%` 到 `88%`。
- 钛银、冷白文字、金属边占 `8%` 到 `12%`。
- 冷银交互高光不超过 `3%`。
- 暖钛色不超过 `2%`，只用于产品高光、manual 或 pending 状态。
- 低饱和错误暖色不超过 `1%`。

禁止：

- 禁止亮蓝、亮绿、青绿色实心 CTA。
- 禁止紫蓝大渐变、彩虹渐变、装饰性光球、bokeh 和无语义霓虹边框。
- 禁止 Hero 状态条默认使用绿色圆点。
- 禁止整页读成深蓝、蓝紫、游戏 UI 或泛 AI SaaS 模板。

状态色必须配合文字标签，不能只靠颜色表达。

## Typography

中文页面优先使用系统中文字体，英文技术词和代码区保留开发者工具感。字体栈使用 frontmatter 中的 `typography` token，不为追求差异随意换字体。

规则：

- 字距固定为 `0`，不要负字距。
- Hero 大字只用于第一屏和详情页顶部。
- 卡片、按钮、状态标签内文字使用紧凑层级，不使用 hero-scale 字号。
- 按钮和标签必须有稳定宽度或响应式约束，状态切换时不能撑坏布局。
- 正文行宽控制在可读范围，长命令和 JSON 放进可横向滚动的 code block。

移动端参考：

- Display 使用 `40px / 1.08`。
- H1 使用 `34px / 1.12`。
- H2 使用 `26px / 1.2`。
- Body 使用 `15px / 1.65`。
- Code 使用 `13px / 1.65`。

## Layout

官网 V1 采用 GoServer 静态页，不引入 Next.js、React、Vue、Tailwind 构建链路。

目标目录：

```text
/home/zsts_119/patchxNoteGoServer/web/mcp/
  index.html
  app.css
  app.js
  data/clients.json
  assets/...
```

布局规则：

- 页面长滚动，不做强制 fullpage scroll。
- 最大内容宽度 `1180px`，Hero 最大宽度 `1280px`。
- 桌面安全边距 `32px` 到 `48px`，移动端 `18px` 到 `22px`。
- Section 垂直间距桌面 `96px` 到 `128px`，移动端 `64px` 到 `80px`。
- 卡片圆角固定 `8px`。
- 不做卡片套卡片。
- 不把整段 section 做成浮卡。
- 不把产品图做成暗、糊、裁切过狠、看不清主体的背景。

## Components

### Header

桌面导航：

```text
PATCHX
MCP
Clients
Security
Docs
Download App
[Get started]
```

Header 使用半透明黑色毛玻璃，但文字、焦点和按钮必须保持足够对比。`Get started` 滚动到客户端选择区，不新增独立官网登录体系。

### Hero

Hero 的任务是一眼确认这是官方 MCP 入口：

```text
PatchXNote MCP
隐私归你，AI 由你掌控
把你的真实对话，安全接入 Cursor、VS Code、Codex 和更多 AI 工具。
```

Hero 只放核心标题、副标题、两枚 CTA、产品主视觉和轻量编辑器提示。完整客户端卡片放第二屏。

产品视觉使用银黑 PatchXNote 录音卡或可解释的产品剪影，不能直接贴用户给的浅蓝白介绍图。

### Buttons

主按钮高度 `44px` 到 `48px`，圆角 `8px`。默认黑铬或石墨金属面，边框为冷银半透明，hover 时只出现窄钛银扫光和轻微位移。

常用按钮文案：

- `Copy AI setup prompt`
- `Install in Cursor`
- `Copy setup command`
- `Copy manual config`
- `Copy remote MCP URL`
- `View security model`

未验收 one-click 只能 disabled 或 coming soon，不能写成已支持。

### Segmented Controls

用于客户端分类和详情安装方式：

```text
Editors | Local MCP / CLI | Cloud Platforms
AI assisted | One-click | Manual config
```

active 状态使用细下划线、边框高亮或小型冷银状态刻线，不使用大面积彩色胶囊。

### Client Cards

字段顺序：

```text
图标 / 首字母
客户端名
类型 + 状态标签
一句话说明
主动作
```

卡片尺寸稳定，桌面 3 列，平板 2 列，移动 1 列。hover 时只允许上浮 `2px`、边框变亮、角箭头移动 `2px` 到 `4px`，不做整卡发光。

### Code Blocks

代码块像终端面板，顶栏显示 `command`、`mcp.json` 或 `config.toml`，右侧复制按钮使用图标或短文案。复制成功后显示 `Copied` 或勾选图标。

代码示例禁止包含 token、手机号、验证码、OAuth code、refresh token、webhook secret、真实用户内容和 provider payload。

### Status Pills

状态文案使用事实分层：

- `One-click candidate`
- `Setup command`
- `Manual`
- `Remote MCP`
- `Pending acceptance`
- `Accepted`

不要把 `researched`、`implemented`、`locally_smoked`、`published_smoked`、`platform_accepted` 混成同一个“已支持”。

### Auth Result Pages

授权成功页和失败页要克制：

- 只展示一个结果。
- 最多一个银色完成或失败符号和两枚按钮。
- 不放产品图、三步流程大图、客户端卡片、安装命令或配置代码。
- 成功态不用绿色圆点或绿色大对勾。
- 错误页要说明过期、取消、state mismatch、网络失败等可恢复状态。

## Motion

动效表达材质和连接，不表达霓虹和炫彩。

预算：

- 每个视口最多 1 个主材质动效。
- 每个视口最多 1 个流程动效。
- 每个视口最多 1 个用户触发动效。

建议：

- Hero 连接线周期 `6s` 到 `10s`。
- 主按钮扫光 `260ms` 到 `360ms`。
- 复制成功边框亮起 `700ms`。
- Tab 切换使用淡入和轻微位移。
- 移动端关闭产品 3D 视差。
- `prefers-reduced-motion` 下关闭连接流动、扫光和视差。

不要同时叠加背景光、产品发光、按钮发光、卡片发光和连接线流动。

## Content

官网文案要具体、可验证、不过度承诺。

必须保留的事实边界：

- 本地链路：npm stdio launcher、browser OAuth、OS-native secure storage。
- 云平台链路：GoServer-hosted remote MCP gateway、OAuth connector session。
- 本地 MCP 配置不放 token。
- 云平台不把本地 `npx` 当成主路径。
- 一键安装只在真实 deeplink、CLI 或客户端流程验收后开启。
- 测试 URL 不能包装成生产入口。

不要在页面、示例、截图、埋点或日志里展示手机号、验证码、OAuth code、PKCE verifier、access token、refresh token、webhook secret、原始转写、完整模型输入输出、provider payload、完整 MAC 或 SK。

## Data Source

官网页面和详情内容必须由数据驱动。

事实源：

- `docs/mcp-clients/clients.json`
- `docs/mcp-clients/website/03-client-install-sources.zh-CN.md`
- 真实客户端验收记录
- GoServer 真实 remote MCP 配置

GoServer 上线快照放在：

```text
web/mcp/data/clients.json
```

快照需要记录 `source_repository`、`source_reviewed_at`、`generated_at`、`default_server`、`remote_gateway` 和 `clients[]`。没有真实来源的字段使用 `unknown`、`pending` 或省略。

## Acceptance

实现完成后至少验证：

- Chrome 桌面 `1440px`、`1280px`、`1366px`。
- Chrome 移动 `390px`、`430px`。
- Header、Hero、客户端卡片、安装 Tab、代码块、toast、Footer 不遮挡、不溢出。
- 主按钮 hover、active、copy、disabled、coming soon 状态清晰。
- `prefers-reduced-motion` 生效。
- 页面源码和数据 JSON 不包含敏感信息。
- `node docs/mcp-clients/validate-clients.mjs` 通过。
- `git diff --check` 通过。
- 如果改 GoServer 路由或静态资源，再运行相关 Go 测试，并明确区分局部测试、测试环境和生产验收。
