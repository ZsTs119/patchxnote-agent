# README Benchmark Notes

本文件记录 PatchNote Agent README 的公开仓库参考和后续维护标准，避免后续文档风格漂移。

## 参考仓库

| 仓库 | 可借鉴点 |
| --- | --- |
| [larksuite/openclaw-lark](https://github.com/larksuite/openclaw-lark) | 顶部中英文切换、能力表、安全与风险提示、免责声明和安装要求。 |
| [google-gemini/gemini-cli](https://github.com/google-gemini/gemini-cli) | 徽章、首屏定位、快速安装、发布通道、认证方式和关键能力分组。 |
| [openai/codex](https://github.com/openai/codex) | 简洁定位、按系统区分安装命令、安装脚本和文档入口清晰。 |
| [github/github-mcp-server](https://github.com/github/github-mcp-server) | Use cases、remote/local MCP 分层、Host 配置示例、认证和 token 安全说明。 |
| [charmbracelet/crush](https://github.com/charmbracelet/crush) | 首屏中英双语定位、安装方式矩阵、平台覆盖说明和功能价值表达。 |
| [stripe/stripe-cli](https://github.com/stripe/stripe-cli) | 成熟 CLI 的安装/升级矩阵、npx 入口、Docker/手工二进制路径和平台注意事项。 |

## PatchNote Agent README 标准

首屏必须回答：

- 这是什么：PatchNote 的本地 CLI 和 MCP bridge。
- 给谁用：桌面 AI Agent / MCP Host。
- 能做什么：读取安全的 PatchNote 账号上下文。
- 不能做什么：不暴露硬件写入、原始音频、完整转写、SK、完整 MAC、支付和 Admin API。
- 怎么开始：一条可复制的 `npx -y patchnote-agent@<version> install --print-config`。

公开 README 必须包含：

- 中英文切换入口：`README.md` 和 `README.zh-CN.md`。
- 徽章：npm version、GitHub release、security policy。
- 功能表：明确支持和不支持的边界。
- 环境要求：Node 版本、系统、账号、MCP Host。
- Quickstart：安装、登录、MCP 启动、切换 base URL。
- MCP 配置：绝对路径配置示例，并说明配置里不保存 token。
- MCP 工具表：列出七个 V1 工具和用途。
- 安全与风险提示：说明 Agent 会让 AI 访问账号元数据，并列出默认安全边界。
- 当前限制：beta、测试 API、文件 keychain fallback、macOS/production pending。
- Troubleshooting：PATH、credential storage、MCP host、platform、checksum、server URL。
- 验证安装：`npm view`、installer dry-run、`patchnote version`。
- 开发和发布维护入口。
- 许可证状态。

禁止在 README、npm README、示例和截图中出现：

- access token、refresh token、OTP、原始手机号、完整 MAC、SK、原始音频、完整转写、用户 prompt、provider payload。
- 未验收的能力承诺，例如生产默认可用、OS keychain 已完成、macOS smoke 已完成、所有 MCP Host 已自动配置。

当能力发生变化时，优先同步：

1. `README.md`
2. `README.zh-CN.md`
3. `packages/npm/README.md`
4. `docs/plans/2026-08-06-agent-v1-mvp.md`
5. 本文件
