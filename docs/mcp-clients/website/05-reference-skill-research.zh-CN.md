# PatchXNote MCP 官网参考仓库与 Skill 选型

**日期：** 2026-08-30

**定位：** 给后续实现 GoServer `web/mcp/` 官网时使用的参考清单。本文只记录可借鉴的方法、需要避开的岔路和建议调用的 skill，不把第三方页面、文案、商标或代码直接作为 PatchXNote 官网素材。

## 当前判断

PatchXNote MCP 官网应该借鉴“开发者工具官网”的信息效率，同时保留 PatchXNote 硬件和隐私记忆产品的高级感。最稳的工程路径不是把官网做成新的 Next.js/Tailwind 应用，而是在 GoServer 现有静态 web 目录下做一个轻量、数据驱动、无构建或低构建的官网：

```text
/home/zsts_119/patchxNoteGoServer/web/mcp/
  index.html
  app.css
  app.js
  data/clients.json
  assets/...
```

原因：

- GoServer 目前已有静态 web embed 和 `/download`、`/legal/privacy` 这类页面经验，首版 MCP 官网最好复用这种发布链路。
- 官网的关键难点是“安装入口、OAuth 登录、三类平台分层、视觉质感”，不是前端框架复杂度。
- 如果直接搬 Next.js/Tailwind/shadcn 模板，会引入路由、构建、依赖、部署和长期维护成本，容易偏离“放到 GoServer web 下面”的目标。
- 2026-08-30 补充判断：防止 AI 写前端跑偏，核心不是再堆一个泛用美化 skill，而是把项目事实、视觉 token、组件规则和可失败的验收放进仓库。`DESIGN.md`、`clients.json`、官网 checklist 和 Chrome 实测，比单独依赖外部 skill 更稳。

## 本地可用 skill 组合

后续进入官网实现时，建议按这个顺序使用现有本地/插件能力。默认先读项目专用文档，再考虑外部 skill：

| 阶段 | 建议 skill / 能力 | 用法 |
| --- | --- | --- |
| 项目事实源 | `AGENTS.md`、本目录 README、`DESIGN.md`、`../clients.json` | 先锁定 GoServer 静态页、三类入口、客户端状态、视觉 token 和验收边界。 |
| 产品结构 | Product Design | 用来校准 IA、用户路径、安装/登录/验证流程，不让官网只变成漂亮页面。 |
| 视觉事实源 | `DESIGN.md` + `04-visual-system.zh-CN.md` | 作为实现前必须读取的视觉合同，约束颜色、字体、按钮、状态、动效和安全文案。 |
| 视觉反跑偏 | `design-taste-frontend` 或 `impeccable` | 用于 polish / audit，不作为开写主规则；发现模板三卡片、紫蓝渐变、按钮溢出、状态过度承诺时回修。 |
| 素材生成 | Creative Production / imagegen | 基于用户提供的硬件图和信息图生成官网统一风格素材；不要直接把浅蓝信息图贴到官网。 |
| 工程实现 | `Code` | 按 GoServer 静态页面边界实现 HTML/CSS/JS 和路由嵌入。 |
| 前端验收 | Chrome plugin | 前端浏览器验收只走 Chrome，不用 Codex 内置浏览器替代。 |

不建议把 Sites 当成首版生产发布路径。Sites 适合独立托管或演示站；本项目已明确官网放在 GoServer `web/mcp/` 下，生产部署应跟 GoServer 走。

`frontend-design` 降级为视觉灵感来源。它适合打开设计思路，但不适合作为 PatchXNote 官网主执行规则，因为它容易鼓励重新发散视觉、引入过强的创意风险，或弱化客户端状态和 OAuth 安全边界。

## 候选外部 skill

以下来自 `npx skills find ...`、GitHub、Reddit / 社区经验和官方资料的筛选结果，暂不直接安装。真正需要时再逐个拉取并审查 `SKILL.md`，避免让第三方规则覆盖 PatchXNote 现有约束。

| 候选 skill / 资料 | 适合程度 | 可借鉴点 | PatchXNote 使用方式 |
| --- | --- | --- | --- |
| `google-labs-code/design.md` | 高 | 把视觉身份、token 和设计理由写成仓库内 AI 可读合同，并支持 lint / diff。 | 已采用为本目录 `DESIGN.md`；这是防跑偏主方案，不只是参考资料。 |
| `vercel-labs/agent-skills@web-design-guidelines` | 高 | 文件级 UI / a11y / UX 审查，输出适合修复的 `file:line` 问题。 | 适合官网实现后的 review，不作为开写规则；安装前需审查其远程拉取规则。 |
| `anthropics/skills@frontend-design` | 中高 | 先做设计计划、再自我批判、再实现，能减少模板化页面。 | 可在视觉方案空白时参考；不能覆盖 `DESIGN.md`、`clients.json` 和验收状态。 |
| `tristanmanchester/agent-skills@designing-beautiful-websites` | 中高 | 从用户目标、IA、wireframe、视觉系统到验证的完整网站设计流程。 | 旧候选保留，但优先级低于本目录文档；适合做设计 PRD 前补流程。 |
| `paidax01/web-to-design-md@website-to-design-md` | 中 | 将参考官网整理成 `design.md` / `DESIGN.md` 和预览页面。 | 只用于把 1Server、Context7 等参考站沉淀为设计观察，不用于克隆页面。 |
| `connorads/dotfiles@web-animation-design` | 中 | 动效时长、缓动、reduced motion、性能边界。 | 仅在需要精修按钮、copy success、OAuth 状态灯和连接线动效时再考虑。 |
| `openai/plugins@frontend-app-builder` | 中 | 先生成完整视觉概念，再按概念高保真实现，适合大型视觉前端。 | 当前不直接使用：其默认 Image Gen 和浏览器验收路径与本项目 Chrome-only、GoServer 静态页边界冲突，需要适配后才可借鉴。 |
| `nexu-io/open-design@frontend-skill` | 中低 | 作为 OpenDesign 目录里的前端 playbook 入口，强调克制构图和官网 UI。 | 暂不作为主 skill；其上游 OpenAI skill 路径已变化，使用前必须审查真实 `SKILL.md` 和依赖。 |
| `skills-101/superpowers@landing-page-design` | 低 | 转化型 landing page 和外部图片生成流程。 | 不建议用于本项目：偏增长页和 `belt`/inference.sh 链路，容易弱化安装事实和验收边界。 |
| `rmyndharis/antigravity-skills@frontend-developer` | 低 | React / Next.js / RSC / 状态管理 / 性能优化。 | 不建议用于 V1：GoServer 原生静态页不需要 React / Next.js 架构。 |
| `nexu-io/open-design@frontend-dev` | 低 | cinematic hero、生成式网页、展示站效果。 | 不建议作为官网主路径：容易把页面带向炫技视觉，而不是安装和授权闭环。 |
| `flitzrrr/frontend-design-skills@landing-pages` | 低 | landing page 专项规则。 | 与本地 `design-taste-frontend`、`impeccable` 和 `DESIGN.md` 重叠，暂不优先。 |

社区经验补充：

- 只靠 `CLAUDE.md`、泛 prompt 或“更会设计的模型”不稳定，长会话后容易掉出上下文。
- 更可靠的方案是设计 token、组件约束、lint / audit、真实浏览器验收和人审结合。
- 对 PatchXNote 来说，`DESIGN.md` + `clients.json` + GoServer 静态页边界 + Chrome 验收，是比堆外部 skill 更适合的防跑偏组合。

## 参考仓库与页面

### 1Server Clients

链接：

- https://1server.ai/clients/
- https://1server.ai/clients/cursor/

可借鉴：

- 首页先让用户选 AI 工具，再进入客户端详情页。
- 每个详情页固定展示客户端简介、`60-second install`、一键按钮、手动配置和 fallback。
- 底部用“不同工具继续选择”的方式把用户留在安装任务里。

PatchXNote 调整：

- 我们不能照搬“same vault / marketplace”的叙事，因为 PatchXNote 不是 MCP server marketplace，而是用户自己的记录、总结和记忆接入 AI。
- 我们的主张应是：`One login. Secret-free local config. Your PatchXNote memory in every AI tool.`
- 详情页主路径不只是一键安装，还要展示“复制一句话给 AI，让 AI 帮你完成 setup”的 AI-assisted 路径。

### GitHub MCP Server

链接：

- https://github.com/github/github-mcp-server

可借鉴：

- 一个官方 MCP Server 同时维护 VS Code、Cursor、Codex、Claude、OpenCode、Windsurf、Zed 等安装入口。
- VS Code 远程 MCP 示例使用 `type: http` + URL，也展示 PAT 输入变量方案。
- 文档明确提醒：不同 MCP Host 的 OAuth / remote 支持细节不同，要以宿主文档为准。

PatchXNote 调整：

- 客户端详情页需要有统一结构：推荐路径、手动配置、验证、故障恢复、支持状态。
- 远程 MCP 平台不能让用户粘贴 PatchXNote access token；若平台需要 secret，只能是平台 connector session 或独立授权。

### Playwright MCP

链接：

- https://github.com/microsoft/playwright-mcp

可借鉴：

- 一个 README 用“标准配置 + 各客户端差异命令”覆盖大量 MCP 客户端。
- 对 Codex、Claude Code、Cursor、VS Code、Windsurf、OpenCode 等都给了对应入口。
- 先给通用配置，再给 host-specific 命令，信息效率很高。

PatchXNote 调整：

- `clients.json` 应继续作为官网卡片和详情页的事实源，避免每个页面手写一遍。
- 每个客户端只展示已验收的按钮。没有真实验收时只放复制命令和手动配置，不写“已一键安装”。

### Context7

链接：

- https://github.com/upstash/context7

可借鉴：

- `npx ctx7 setup` 一条命令把 OAuth、API key、skill/MCP 模式选择和目标 agent 安装串起来。
- 同时支持 `CLI + Skills` 和 `MCP` 两种模式，这和我们讨论的“给 AI 一句话”和“本地 MCP”很接近。
- 文案重点是减少用户切换窗口和减少过时配置，而不是解释太多协议细节。

PatchXNote 调整：

- 官网按钮可以同时提供：
  - `Install in Cursor/VS Code`：真正 deeplink 验收后启用。
  - `Ask your AI to set it up`：复制一句话给当前 AI。
  - `Manual config`：永远保留的 fallback。
- 本地登录仍由 `npx -y patchxnote-agent@latest mcp login` 完成，不在 MCP config 里放 token。

### Unkey Marketing

链接：

- https://github.com/unkeyed/marketing

可借鉴：

- 把 marketing site、playground、glossary generator 放在一个 marketing repo/monorepo 下，并区分 `www`、`play`、`generator`。
- `apps/www` 里有 `app`、`components`、`content`、`images`、`public`、`lib` 等清晰分层。

PatchXNote 调整：

- 首版 GoServer 静态页不需要照搬 monorepo，但可以借鉴目录分层思想：页面、数据、图片、组件样式分开。
- 未来如果 MCP 官网扩展成多页面文档，再考虑把 `web/mcp/` 抽成独立前端工程。

### Supabase Website

链接：

- https://github.com/supabase/supabase/tree/master/apps/www

可借鉴：

- 图片治理规则非常细：不同用途的缩略图、社交图、客户 logo、暗色/亮色版本分开。
- `/go/*` campaign pages 是用结构化 page object + section renderer 管理的，不是散落 HTML。
- PR 检查里有可访问性与页面针对性验收思路。

PatchXNote 调整：

- 我们可以提前规定官网素材命名：hero、OG、client icons、generated panels 分开。
- 首版不做复杂动态 OG 服务，但要预留 `og-image` 静态素材。
- 客户端 logo 和商标使用在公开发布前必须单独确认授权或使用规范。

### Launch UI / shadcn/ui / TailGrids

链接：

- https://github.com/launch-ui/launch-ui
- https://github.com/shadcn-ui/ui
- https://github.com/Tailgrids/tailgrids

可借鉴：

- 组件化思路：导航、hero、feature、tabs、toast、dialog、cards、button states。
- 可访问性、深浅色、响应式、SEO 等检查项比较完整。
- shadcn/ui 的价值是“open code, make it your own”，不是照搬默认视觉。

PatchXNote 调整：

- 因为 GoServer 首版更适合静态页，不建议引入 React/shadcn runtime。
- 可以把组件设计经验翻译成原生 HTML/CSS/JS：segmented controls、copy buttons、hover/focus/active、toast、dialog。

### Plain HTML SaaS Template

链接：

- https://github.com/hannah-wright/saas-landing-page-template

可借鉴：

- 同时提供 React、Vue、HTML 版本；HTML 版可以直接打开，适合无构建场景。
- 设计 token 放在一个 `globals.css`，方便 AI 或工程师统一调整。
- 多页面结构和 dark/light token 对小团队落地很友好。

PatchXNote 调整：

- 这是最接近 GoServer 静态落地方式的参考，但视觉不能直接使用模板样式。
- 我们可以采用“一个 CSS token 文件 + 一个轻 JS 行为层”的工程形态。

### Trigger.dev

链接：

- https://github.com/triggerdotdev/trigger.dev

可借鉴：

- 面向 AI agents / workflows 的开发者产品叙事，强调 durable、observable、人参与审批等能力。
- 大型开源项目把 webapp、docs、examples、self-hosting、roadmap 分得很清楚。

PatchXNote 调整：

- 我们不是 workflow 平台，不能把官网写成“运行 AI agent”的平台文案。
- 可借鉴的是“把复杂技术能力翻译成用户可理解路径”：安装、授权、使用、验证。

## PatchXNote 官网定制规则

### 信息架构

- 首页继续使用长滚动，不做强制全屏翻页。
- 首页主选择区使用三类 tab：`Editors`、`Local MCP / CLI`、`Cloud Platforms`。
- 客户端详情页统一使用：`AI assisted`、`One-click`、`Manual config` 三个 setup tab。
- 页面顶部保留 `PATCHX`、`MCP`、`Clients`、`Security`、`Docs`、`Download App`、`Get started`。
- 页面底部要提供：产品、安装、开发者、法律、状态口径和测试/生产 MCP URL 说明。

### 工程边界

- 首版采用原生 `index.html` + `app.css` + `app.js`，不引入 Next.js、React、Tailwind build。
- `clients.json` 作为官网数据源，后续可从 `patchnote-agent/docs/mcp-clients/clients.json` 同步到 GoServer。
- 客户端卡片、详情内容、按钮状态由数据驱动渲染，避免复制粘贴造成状态漂移。
- 复制命令、deeplink 生成、toast、tabs、filter/search、展开详情可以用少量原生 JS 实现。

### 视觉边界

- 延续 `04-visual-system.zh-CN.md` 的 `Black Chrome Command Hub`。
- 高级黑不是整页纯黑，要有冷银金属、黑铬按钮、钛银边框和低饱和暖金属点缀；不使用青绿色状态光、电蓝辅助光或蓝绿实心按钮。
- 按钮要有真实状态：hover edge light、active press、focus ring、copy success morph、disabled/coming soon。
- 避免紫蓝大渐变、模板化三卡片、纯文字 hero、把用户浅蓝信息图直接贴进暗色官网。

### 安装/登录边界

- `One-click` 只在对应客户端 deeplink 或官方安装入口真实验收后启用。
- `AI assisted` 是首版默认推荐路径：复制一句话到当前 AI，让 AI 帮用户运行 setup/login/config。
- `Manual config` 永远保留，且明确运行时边界：Windows 桌面、WSL、Remote、Dev Container 不共享安全存储。
- 云平台只给 remote MCP / OAuth connector 流程，不给本地 `npx` 指令作为主路径。

### 验收边界

- 浏览器验收走 Chrome plugin。
- 每个客户端至少验收：安装入口、登录/授权、MCP tools/list、一个只读工具调用、重启后仍可用、退出/撤销路径。
- 官网自测至少覆盖：桌面宽屏、1366 桌面、iPhone 宽度、按钮文字不换行、文本不重叠、copy toast、reduced motion。
- 文案不要把 `implemented`、`locally_smoked`、`published_smoked`、`platform_accepted` 混成同一个“已支持”。

## 推荐执行顺序

1. 先确认 `01-information-architecture.zh-CN.md`、`02-entry-action-model.zh-CN.md`、`03-client-install-sources.zh-CN.md`、`04-visual-system.zh-CN.md`、`06-implementation-readiness.zh-CN.md` 和本文无冲突。
2. 把 `clients.json` 转成官网需要的 data schema，保留 evidence/status 字段。
3. 在 GoServer 增加 `web/mcp/` 静态资源骨架、embed 和 `/mcp/*` 路由。
4. 生成或处理 hero、OG、client category、setup preview 所需素材。
5. 在 GoServer `web/mcp/` 写静态官网。
6. 用 Chrome 做视觉和交互验收，再补 GoServer 静态资源路由测试。

## 暂缓事项

- 暂不安装外部 skill，除非进入具体设计/动效/克隆分析阶段。
- 暂不引入 Tailwind/React/Next.js，除非静态页已经无法支撑交互复杂度。
- 暂不承诺 Claude Desktop `.mcpb` 一键包。MCPB 是有潜力的后续方向，但需要单独打包、签名、测试和发布流程。
- 暂不做平台私有插件或 webhook/HTTP tool 替代方案。飞书、豆包、腾讯、WorkBuddy 先按远程 MCP + OAuth connector 走。
