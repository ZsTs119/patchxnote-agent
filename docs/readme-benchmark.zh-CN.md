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

- 所有 README / npm / 对外公开图片都必须同时出现：
  - PatchXNote / PATCHX 品牌标识，不能只用抽象图标或无品牌卡片。
  - AI 录音卡真实产品主体，不能只用概念设备、几何装饰或纯文字说明。
- 图片风格参考现有产品宣发图：浅蓝底、白色纸面模块、黑色主文字、简洁留白、真实产品图；避免 AI 生成感、渐变炫光、抽象科技背景和过度装饰。
- 产品图和 logo 不能只是“裁剪后贴进去”的小插图。每张公开图必须围绕产品做版面设计，让 AI 录音卡成为主视觉或清晰的视觉锚点，再用少量信息解释 Agent 能力。
- 设计目标是“产品宣发 + 功能说明平衡”：比技术流程图更像产品宣传页，但仍能让 README 读者快速知道能做什么、怎么开始、边界在哪里。
- 每张图只突出一个重点。不要把全部功能堆到每张图里；用户扫图时应能直接知道这张图负责解释什么。
- `patchxnote-agent-cover.png`
  - 重点：Agent 的首屏价值。
  - 主标题：`让 AI 读懂你的记录`
  - 表达：记录查询、AI 整理和 Webhook 发送。
  - 右侧使用产品宣发纸面，必须出现 PatchXNote / PATCHX 品牌信号和 AI 录音卡真实产品图。
- `patchxnote-agent-quickstart.png`
  - 重点：用户三步接入。
  - 标题：`三步接入 AI 助手`
  - 步骤：`安装`、`登录`、`接入 MCP`
  - 命令示例：`npx -y patchxnote-agent install --print-config`、`patchxnote login`、`patchxnote mcp serve`
- `patchxnote-agent-records.png`
  - 重点：按平台找到记录。
  - 标题：`找到你的每条记录`
  - 表达：按手机端或电脑端查看记录，支持搜索和读取记录详情。
- `patchxnote-agent-tools.png`
  - 重点：19 个 MCP 工具怎么分组。
  - 标题：`19 个 MCP 工具`
  - 三组：`记录查询 7`、`Webhook 7`、`AI 结果 5`
- `patchxnote-agent-model-io.png`
  - 重点：AI 整理背后的内容可以显式查看或导出。
  - 标题：`看见 AI 整理背后的内容`
  - 四组：`原始文本`、`AI 返回`、`解析结果`、`最终结果`
- `patchxnote-agent-webhook-delivery.png`
  - 重点：Webhook 协作发送。
  - 标题：`整理好，再发送`
  - 主链路：`Markdown 草稿 -> 确认发送 -> 飞书/钉钉/Webhook`
  - 表达：先生成草稿，用户确认后再手动发送。
- `patchxnote-agent-safety-boundary.png`
  - 重点：能看什么、不会做什么。
  - 标题：`本地运行，明确边界`
  - 左侧：读取记录、查看 AI 结果、生成本地草稿、手动发送 Webhook。
  - 右侧：原始音频、硬件操作、支付/Admin、后台自动发送都不做。

图片必须保持现有浅蓝/白色背景、白色卡片、黑色主文字、真实产品图和 PATCHX/PatchXNote 品牌信号。中文或英文文字如果变形、拼错、裁切、过小，不能发布。

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
