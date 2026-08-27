# PatchXNote MCP 客户端登记表

这组文件是 PatchXNote MCP 官网、CLI `setup --client`、后续市场提交和平台联调共用的事实源。第一版先把本地可闭环和平台待闭环分开，避免把“能复制配置”误写成“已经一键安装验证通过”。

## 状态口径

- `supported`：CLI 已能给出可执行安装路径，或能安全写入已验证格式的本地配置。
- `manual`：CLI 给出复制命令、配置片段和验证提示，但 V1 不自动写配置。
- `planned`：已确认方向，等待平台端、远程 MCP 网关或真实客户端验收。
- `research`：仅作为观察列表，还不能放进首版主路径。

`evidence_state` 比 `support_status` 更细，允许从 `researched` 逐步推进到 `implemented`、`locally_smoked`、`published_smoked`、`platform_accepted`。

## 本地闭环

截至 `2026-08-27`，`0.2.8` 本地候选已通过 Windows-native 通用链路验收：浏览器 OAuth 自动打开、GoServer 页面登录、Windows 凭据保存、`mcp status --verify`、stdio `initialize/tools/list/current-user`、mobile 总结记录和 model-IO 字段读取、以及 npm wrapper `--from-local` 候选安装代理。证据见 `docs/evidence/2026-08-27-mcp-oauth-local-acceptance.zh-CN.md`。

这个结论只代表“通用本地 stdio MCP 链路可用”。VS Code、Cursor、Codex、Claude Desktop、Claude Code、Windsurf、Trae、Qoder、WorkBuddy 等每个客户端自己的 UI/配置写入/刷新流程，还需要在发布候选包或 registry 包上分别验收后，才能把对应客户端标记为 `locally_smoked` 或 `published_smoked`。

本地客户端第一步统一执行浏览器 OAuth 登录：

```sh
npx -y patchxnote-agent@latest mcp login
```

然后使用无密钥 stdio 配置：

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

`setup --client <id>` 会复用同一套 `mcp login`。编辑器启动 `mcp serve` 时不会自动弹浏览器，避免 MCP Host 启动超时。令牌只存 OS 原生安全存储。官网、编辑器配置、MCP JSON/TOML 都不能出现手机号、验证码、access token、refresh token、webhook secret、录音原文、完整转写或模型 provider 原始响应。

关键要求：在哪个运行时启动 MCP，就在哪个运行时执行登录和 setup。Windows 桌面编辑器、WSL 终端、VS Code Remote、Dev Container、远端 Linux 主机不能默认共享同一个本机安全存储。

## 平台闭环

飞书 Aily / 豆包工作伙伴、腾讯 Agent 平台、企业版 WorkBuddy 这类平台通常不能运行用户本机的 `npx`，所以它们走远程 MCP 网关：

```text
https://ws-lab.patch-x.cn/patchnote-test-api/mcp
```

远程 MCP 的目标是和本地 Agent MCP 使用同一套 curated 工具语义，当前公测服务端 `/mcp` 已按 19 个 V1 工具暴露；平台控制台最终可用范围还要以平台权限、OAuth 授权和真实验收为准。官网文案不要把平台型客户端写成“本地一键安装”，只写“远程 MCP 待验收”。

## 维护规则

- 新客户端先进入 registry，再进入官网和 CLI。
- 未经真实客户端验证，不把 `website_buttons` 写成一键安装。
- 自动写配置前必须有确认、备份、回滚路径和 fixture 测试。
- 官方文档链接需要按发布日期重新审查；本文件最后审查日期是 `2026-08-27`。
- 客户端 Logo、名称和商标使用在公开官网发布前单独确认授权或使用规范。
