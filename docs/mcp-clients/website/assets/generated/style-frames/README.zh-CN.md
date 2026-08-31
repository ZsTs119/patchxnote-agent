# PatchXNote MCP 官网 Style Frames

**日期：** 2026-08-28

本目录保存第一轮官网视觉方向稿。它们用于选择官网气质、页面密度、按钮风格、产品图处理和授权页氛围，不是最终上线页面，也不是最终文案。

## 文件清单

| 文件 | 用途 | 说明 |
| --- | --- | --- |
| `01-home-hero-style-frame.png` | 首页首屏视觉方向 | 验证高级黑官网首屏、产品主视觉、客户端卡片和主 CTA 气质 |
| `02-client-install-style-frame.png` | 客户端安装详情页方向 | 验证一键安装、AI assisted、Manual config、代码块和状态模块 |
| `03-auth-success-style-frame.png` | 授权成功页方向 | 验证授权完成、安全存储提示和回到编辑器的收束体验 |
| `04-home-hero-black-chrome-revision.png` | 首页首屏历史修正版 | 基于第一张修正：主 CTA 改为黑铬金属按钮；仍包含绿色状态点，后续不再沿用 |
| `05-home-hero-interactive-effects.png` | 首页交互效果历史修正版 | 基于 `04` 增强：状态条、黑铬 CTA、产品悬浮、细连接线、客户端卡片 hover 效果；仍包含绿色状态点，后续不再沿用 |
| `06-home-hero-no-green-status.png` | 首页当前推荐基准 | 移除绿色状态点，使用冷银刻线、细小波形和钛银边光表达系统状态 |
| `07-home-hero-slim-first-screen.png` | 首页已确认基准 | 用户确认的首页风格和产品主视觉方向：只保留 PATCHX、核心标题、两枚 CTA、产品主视觉和轻量编辑器提示 |
| `08-client-install-slim-black-chrome.png` | 详情页已确认基准 | 用户确认的产品信息展示和安装详情页方向：AI assisted、One-click、Manual config 三个 tab，右侧冷银安装进度，无绿色状态点 |
| `09-auth-success-slim-black-chrome.png` | 授权成功页历史方向 | 银色连接环、三步状态线、本地运行时回传说明；信息仍偏多，后续不再沿用 |
| `10-auth-success-minimal-black-chrome.png` | 授权成功页当前推荐基准 | 只保留授权成功结果、一句说明、返回编辑器/查看指南按钮和本机安全存储小字 |
| `11-clients-selection-black-chrome.png` | 客户端选择页当前推荐基准 | 三类入口：编辑器、云平台、本地 MCP；默认展示 Cursor、VS Code、Codex、Claude Code、WorkBuddy、飞书 Aily |
| `12-cloud-platform-black-chrome.png` | 云平台接入页当前推荐基准 | 远程 MCP 地址、发送给 AI 的一句话、官网完成授权三段式 |
| `13-local-cli-mcp-black-chrome.png` | 本地 CLI/MCP 页当前推荐基准 | 复制 setup 命令、验证一句话、本地运行时三步说明 |
| `14-login-consent-black-chrome.png` | 登录授权页当前推荐基准 | OAuth 授权中间页：当前客户端、只读权限、授权并继续、取消 |
| `15-auth-error-minimal-black-chrome.png` | 授权失败页当前推荐基准 | 与授权成功页同构，只保留失败结果、错误代号、重试和返回指南 |
| `16-security-privacy-black-chrome.png` | 安全隐私页当前推荐基准 | 本机整理、安全存储、只读工具、可撤销授权的信任说明 |
| `17-docs-footer-black-chrome.png` | Docs/底部入口页当前推荐基准 | 安装指南、客户端兼容、安全与授权、下载 App 和底部导航 |
| `18-phone-login-black-chrome.png` | 手机号登录页当前推荐基准 | 授权流程前置登录：手机号、验证码、登录并继续、返回设置指南 |

## 使用边界

- 这些图片里的文字、命令和 logo 不能直接视为最终事实。
- 后续实现必须以 `docs/mcp-clients/clients.json` 和 `03-client-install-sources.zh-CN.md` 为准。
- 这些图可以作为 `04-visual-system.zh-CN.md` 的视觉参考和 GoServer 官网实现参考。
- 当前首页实现必须优先参考 `07-home-hero-slim-first-screen.png`。
- 当前客户端安装详情页和产品信息展示必须优先参考 `08-client-install-slim-black-chrome.png`。
- 后续如需生成官网素材、产品渲染图、局部交互状态图，优先沿用 `07` 和 `08` 的黑铬、冷银、金属刻线、低饱和高对比方向。
- 当前客户端选择页实现优先参考 `11-clients-selection-black-chrome.png`。
- 当前云平台接入详情页实现优先参考 `12-cloud-platform-black-chrome.png`。
- 当前本地 CLI/MCP 详情页实现优先参考 `13-local-cli-mcp-black-chrome.png`。
- 当前登录授权页实现优先参考 `14-login-consent-black-chrome.png`。
- 当前手机号登录页实现优先参考 `18-phone-login-black-chrome.png`。
- 当前授权成功页实现优先参考 `10-auth-success-minimal-black-chrome.png`，只表达结果和下一步。
- 当前授权失败页实现优先参考 `15-auth-error-minimal-black-chrome.png`。
- 当前安全隐私页实现优先参考 `16-security-privacy-black-chrome.png`。
- 当前 Docs/底部入口页实现优先参考 `17-docs-footer-black-chrome.png`。
- `09` 不再作为实现参考，因为信息过多。
- 后续实现必须移除绿色圆点，改用冷银刻线、金属凹槽、细小波形或钛银脉冲线表达状态。
- 第一屏不要放大客户端卡片网格、四个以上功能点或 `更多客户端持续接入中` 文案。
- `编辑器`、`云平台`、`本地 MCP` 在选择区作为 tab 切换卡片列表；点击卡片进入详情页。
- `12` 和 `13` 是详情页类型参考，不要求做成独立一级页面。
- 选择区需要保留 `找不到你的渠道？` 兜底入口，展示通用 `mcpServers` block。
- 选定方向后，应再生成正式 hero 图、流程图、云平台图和 OG 图。
