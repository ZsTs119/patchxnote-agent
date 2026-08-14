# README Benchmark Notes

本文件记录 PatchXNote Agent README 的公开仓库参考和后续维护标准，避免后续文档风格漂移。

## 参考仓库

| 仓库 | 可借鉴点 |
| --- | --- |
| [larksuite/openclaw-lark](https://github.com/larksuite/openclaw-lark) | 顶部中英文切换、能力表、安全与风险提示、免责声明和安装要求。 |
| [google-gemini/gemini-cli](https://github.com/google-gemini/gemini-cli) | 徽章、首屏定位、快速安装、发布通道、认证方式和关键能力分组。 |
| [openai/codex](https://github.com/openai/codex) | 简洁定位、按系统区分安装命令、安装脚本和文档入口清晰。 |
| [github/github-mcp-server](https://github.com/github/github-mcp-server) | Use cases、remote/local MCP 分层、Host 配置示例、认证和 token 安全说明。 |
| [charmbracelet/crush](https://github.com/charmbracelet/crush) | 首屏中英双语定位、安装方式矩阵、平台覆盖说明和功能价值表达。 |
| [stripe/stripe-cli](https://github.com/stripe/stripe-cli) | 成熟 CLI 的安装/升级矩阵、npx 入口、Docker/手工二进制路径和平台注意事项。 |

## PatchXNote Agent README 标准

首屏必须回答：

- 这是什么：PatchXNote 的本地 AI 助手连接器。
- 给谁用：桌面 AI Agent / MCP Host。
- 能做什么：查看 PatchXNote 记录、AI 整理结果，生成 Markdown，并手动发送到 webhook。
- 安全边界：服务端数据只读；原始音频、硬件写操作、支付和 Admin API 不开放；webhook 只有明确调用才发送。
- 怎么开始：一条可复制的 `npx -y patchxnote-agent@<version> install --print-config`。

公开 README 必须包含：

- 中英文切换入口：`README.md` 和 `README.zh-CN.md`。
- 徽章：npm version、GitHub release、security policy。
- 首屏封面图：`docs/assets/patchxnote-agent-cover.png`。
- 关键视觉素材：quickstart、architecture、tools、safety-boundary，用于解释安装、MCP 架构、工具能力和安全边界。
- 功能表：明确支持和不支持的边界。
- 环境要求：Node 版本、系统、账号、MCP Host。
- Quickstart：安装、登录、MCP 启动、切换 base URL。
- 常用场景：用用户语言描述记录查找、AI 结果查看、Markdown 草稿、webhook 发送。
- MCP 配置：绝对路径配置示例，并说明配置里不保存 token。
- MCP 工具表：列出 19 个工具，并按用户任务分组：
  - 账号和记录查询：7 个。
  - Webhook 配置和发送：7 个。
  - AI 整理结果查看：5 个。
- CLI 命令：按安装登录、AI 整理结果、webhook 配置发送分组。
- 安全与风险提示：说明 Agent 会让 AI 访问账号、记录、原文文本和 AI 结果。
- 当前限制：beta、公测 API、Linux headless keychain 可用性、无自动发送、无原始音频下载、无硬件/支付/Admin。
- Troubleshooting：PATH、credential storage、MCP host、platform、webhook provider error、checksum、server URL。
- 验证安装：`npm view`、installer dry-run、`patchxnote version`。
- 开发和发布维护入口。
- 许可证状态。

## 公开用语标准

| 内部词 | README 用户说法 |
| --- | --- |
| `memory` | 记录 |
| `structured-result metadata` | 记录列表和基础信息 |
| `model IO trace` | AI 整理记录 / AI 处理记录 |
| `request_id` | 处理编号；命令示例里保留 `request_id` |
| `provider response` | AI 返回内容 |
| `parsed result` | AI 解析后的结果 |
| `packaged result` | 最终整理结果 |
| `source text` / safe plaintext projection | 可查看的原文文本 / 安全文本 |
| `generic webhook` | 其他 webhook 地址；命令示例里保留 `generic` |

## 图片素材事实标准

当产品事实变化时，必须同步更新 `docs/assets/patchxnote-agent-*.png`。当前公开素材应表达：

- `patchxnote-agent-cover.png`
  - 标题：`PatchXNote Agent`
  - 副标题：`让 AI 帮你查看记录、整理结果、发送 Webhook`
  - 卡片：`安装 CLI`、`连接 AI 助手`、`查看记录`、`手动发送`
- `patchxnote-agent-quickstart.png`
  - 标题：`三步接入 PatchXNote Agent`
  - 步骤：`安装 CLI`、`验证码登录`、`连接 AI 助手`
  - 底部：`开始查看记录、AI 整理结果，并手动发送 Webhook`
- `patchxnote-agent-architecture.png`
  - 主链路：`AI 助手 -> 本地 MCP -> patchxnote CLI -> Agent API`
  - 本地支路：`本地 Webhook 配置/发送`
  - 目标：`PatchXNote 服务端只读接口`、`本地安全存储`、`飞书/钉钉/Webhook`
- `patchxnote-agent-tools.png`
  - 标题：`PatchXNote Agent 工具能力`
  - 副标题：`19 个 MCP 工具，帮 AI 查看记录、整理结果和发送 Webhook`
  - 三组：`账号和记录查询 7`、`Webhook 配置和发送 7`、`AI 整理结果查看 5`
- `patchxnote-agent-safety-boundary.png`
  - 左侧：`可以明确查看`，包含账号与额度、记录列表、AI 整理结果、Webhook 别名。
  - 右侧：`不会自动操作`，包含原始音频、硬件绑定/解绑、支付/Admin、自动发送。
  - 底部：`原文文本和 AI 结果需显式调用，建议导出到本地文件`
- `patchxnote-agent-feishu-cover.png`
  - 副标题：`把 PatchXNote 记录整理成 Markdown，发送到飞书/钉钉`
  - 卡片：`查看记录`、`生成草稿`、`确认后发送`
- `patchxnote-agent-social-preview.png`
  - 英文短句：`Records · AI Results · Webhook Delivery`

图片必须保持现有浅蓝/白色背景、蓝色图标、白色卡片、柔和阴影和 MR20 产品信号。中文或英文文字如果变形、拼错、裁切、过小，不能发布。

## 禁止内容

禁止在 README、npm README、示例和截图中出现：

- access token、refresh token、OTP、原始手机号、完整 MAC、SK、真实 webhook URL、真实 request ID、原始音频、完整用户原文、用户 prompt、provider payload。
- 未验收的能力承诺，例如生产默认可用、macOS smoke 已完成、所有 MCP Host 已自动配置。
- "自动发送 webhook"、"后台推送"、"执行模型整理" 等未支持能力。
- "开源" 说法；当前仓库和 npm package 仍是 `UNLICENSED`。

## 能力变化同步清单

当能力发生变化时，优先同步：

1. `README.md`
2. `README.zh-CN.md`
3. `packages/npm/README.md`
4. `docs/assets/patchxnote-agent-*.png`
5. `docs/release-and-maintenance-runbook.zh-CN.md`
6. `docs/plans/2026-08-06-agent-v1-mvp.md`
7. 本文件

每次发布前至少确认：

```sh
grep -R --exclude='readme-benchmark.zh-CN.md' "7 个只读\\|18 个\\|seven V1\\|memory metadata" README.md README.zh-CN.md packages/npm/README.md docs
grep -R --exclude='readme-benchmark.zh-CN.md' "17795780915\\|open-apis/bot/v2/hook\\|access_token\\|refresh_token\\|sk_" README.md README.zh-CN.md packages/npm/README.md docs
```
