# PatchXNote MCP 官网规划入口

**日期：** 2026-08-28

这个目录沉淀 PatchXNote MCP 官网的产品、视觉、安装动作、素材和验收流程。后续写 GoServer `web/mcp/` 页面时，先读本文，再读编号文档和 `../clients.json`。

## 当前结论

- 官网 V1 放到 GoServer `web/mcp/`，使用原生 HTML/CSS/JS。
- 首页、客户端详情页、云平台详情页、授权登录页、授权成功页、授权失败页保持同一套高级黑视觉。
- 新增 `DESIGN.md` 作为官网给 AI coding agent 使用的视觉事实源，后续实现不得绕过它重新发散。
- 安装入口分三类：编辑器、本地 MCP / CLI、云平台。
- 客户端详情页统一提供 `AI assisted`、`One-click`、`Manual config`。
- 一键安装按钮只在真实 deeplink、CLI 或客户端流程验收后开启。
- 云平台只走 remote MCP / OAuth connector session，不把本地 `npx` 作为主路径。
- 用户提供的四张图只作为产品事实和视觉生成参考，不直接作为官网最终素材。

## 文件索引

```text
01-information-architecture.zh-CN.md
  官网页面架构、顶部/底部、section 和每屏信息。

02-entry-action-model.zh-CN.md
  编辑器、云平台、本地 MCP/CLI 三类入口动作模型。

03-client-install-sources.zh-CN.md
  各客户端官方文档来源、真实安装逻辑、不可 mock 的边界。

04-visual-system.zh-CN.md
  高级黑视觉系统、颜色、按钮、动效、素材风格。

DESIGN.md
  官网视觉 token、组件约束、内容边界和验收规则，供 AI coding agent 在实现前读取。

05-reference-skill-research.zh-CN.md
  参考站点、开源经验、可用 skill 和实现取舍。

06-implementation-readiness.zh-CN.md
  GoServer 路由、静态资源、授权页、数据快照和开写条件。

07-acceptance-checklist.zh-CN.md
  官网实现、客户端安装、授权、安全和视觉验收 checklist。

08-visual-product-optimization-checklist.zh-CN.md
  基于已认可参考图和 Chrome 实测整理的视觉、响应式、跳转和图标优化 checklist。

assets/reference/README.zh-CN.md
  原始参考素材说明。

assets/generated/style-frames/README.zh-CN.md
  第一轮官网视觉方向稿说明。
```

## 数据源

- `../clients.json` 是客户端登记表事实源。
- `../validate-clients.mjs` 用于校验客户端登记表结构。
- 官网上线快照应复制到 GoServer `web/mcp/data/clients.json`，不要运行时读取 sibling repo。

## 维护规则

- 新增或调整客户端时，先更新 `../clients.json`，再更新本目录文档。
- 所有客户端安装动作都要保留官方来源或真实验收记录。
- 未验收的一键安装只显示为待验收或次级入口。
- 原始素材只能作为参考来源，正式官网素材放在 GoServer 官网目录或后续生成资产目录，并记录生成依据。
