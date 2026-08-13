# PatchXNote Agent

[English](./README.md) | [简体中文](./README.zh-CN.md)

[![npm version](https://img.shields.io/npm/v/patchxnote-agent.svg)](https://www.npmjs.com/package/patchxnote-agent)
[![GitHub release](https://img.shields.io/github/v/release/ZsTs119/patchxnote-agent)](https://github.com/ZsTs119/patchxnote-agent/releases)
[![Security policy](https://img.shields.io/badge/security-policy-blue.svg)](./SECURITY.md)

官方文档：[PatchXNote Agent 公测使用指南（飞书公开版）](https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd)

GitHub 仓库：[https://github.com/ZsTs119/patchxnote-agent](https://github.com/ZsTs119/patchxnote-agent)

![PatchXNote Agent 封面](./docs/assets/patchxnote-agent-cover.png)

PatchXNote Agent 是 PatchXNote 的本地 CLI 和 MCP 桥接工具。它让桌面 AI Agent 可以读取安全的 PatchXNote 账号上下文，包括账号状态、已绑定录音卡、额度、模型使用情况和结构化结果元数据。它也可以把用户确认后的 Markdown 手动发送到本地配置的飞书、钉钉或 generic webhook 目标。

Agent 服务端数据访问明确保持只读。它只调用专用的 `/v1/agent/**` PatchXNote 服务端 API，不暴露 App/PC 的硬件写入流程、原始音频、完整转写、SK、完整 MAC、额度购买流程或 Admin API。webhook 工具属于本地配置写入和手动外部发送；webhook URL 和签名密钥只作为写入输入，不会被列表或查询返回。

将以下一句话发送给支持本地命令执行的 AI 助手即可：

```text
请按照 PatchXNote Agent 使用指南（https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd）帮我安装并接入 MCP：先执行 npx -y patchxnote-agent install --print-config，然后引导我执行 patchxnote login 登录 PatchXNote 账号，并把安装器打印的 MCP JSON 配置接入当前 AI 助手；过程中不要让我把验证码、access token 或 refresh token 粘贴到对话里；GitHub 仓库是 https://github.com/ZsTs119/patchxnote-agent。
```

也可以手动执行安装命令：

```sh
npx -y patchxnote-agent install --print-config
```

## 快速了解

| 维度 | Agent V1 行为 |
| --- | --- |
| 运行方式 | 通过 npm 安装壳下载并安装版本化的原生 `patchxnote` 二进制。 |
| Agent 协议 | 通过 `patchxnote mcp serve` 启动本地 stdio MCP server。 |
| 登录 | 手机验证码登录，创建独立 Agent 会话，不占用 mobile/desktop 安装位。 |
| 数据访问 | 读取有边界的账号、录音卡、额度、模型使用和结构化结果元数据投影。 |
| 安全边界 | 服务端数据只读、脱敏、按平台隔离，并且只走专用 Agent 接口；webhook 是本地手动发送副作用。 |
| 包状态 | 公开 beta 版 `0.2.3`，默认连接 PatchXNote 公测 API。 |

## 功能

| 能力 | `0.2.3` 是否支持 | 说明 |
| --- | --- | --- |
| 手机验证码 Agent 登录 | 支持 | 使用 Agent 专用登录态，不影响 mobile/desktop 安装位。 |
| Agent 会话自动刷新 | 支持 | 从本机安全钥匙串自动轮换 Agent access token 和 refresh token。 |
| 本地 MCP server | 支持 | 通过 `patchxnote mcp serve` 使用 stdio 通信。 |
| 当前账号投影 | 支持 | 返回状态、脱敏手机号、注册平台、状态版本。 |
| 录音卡列表 | 支持 | 只读投影，只返回脱敏标识。 |
| 额度汇总 | 支持 | 返回当前账号 token 额度概览。 |
| 模型使用汇总 | 支持 | 返回当月模型使用和扣费额度概览。 |
| 结构化结果元数据 | 支持 | 按 `mobile` 或 `desktop` 平台隔离。 |
| 本地记忆搜索 | 支持 | 搜索当前 MCP 会话中已授权缓存的元数据。 |
| 本地 webhook 发送 | 支持 | 配置带别名的飞书、钉钉或 generic webhook，手动发送可编辑 Markdown。 |
| 记忆 webhook 草稿 | 支持 | 读取 Agent delivery-document 投影，保存可编辑草稿，并可显式导出 model IO JSON。 |
| Model IO 字段查看 | 支持 | 可按 memory 或 request 显式查看 source text、provider response、parsed result、packaged result。 |
| 硬件绑定/解绑/恢复 | 不支持 | 仍由 App/PC 和 MR20 流程负责。 |
| 原始音频/完整转写/下载 | 不支持 | V1 明确不暴露。 |
| 模型执行 | 不支持 | Agent 服务端数据访问保持只读。 |

## 环境要求

- Node.js `18` 或更高版本，用于 npm 安装壳。
- Windows、macOS 或 Linux，支持 `amd64` 和 `arm64`。
- 可以接收手机验证码的 PatchXNote 账号。
- 支持 stdio MCP server 的 MCP Host，例如 Codex、Claude Desktop、Cursor、VS Code 或其他兼容桌面 Agent。

> `0.2.3` 是公测 beta 构建。默认服务端是 PatchXNote 公测 API，凭据默认写入系统原生安全钥匙串。

## 快速开始

![三步接入 PatchXNote Agent](./docs/assets/patchxnote-agent-quickstart.png)

安装 npm 包。它会从 GitHub Releases 下载匹配平台的 `patchxnote` 二进制，校验 `checksums.txt`，并安装到用户可写目录。

```sh
npx -y patchxnote-agent install --print-config
```

安装器会打印：

- 已安装的二进制路径
- 如果 `patchxnote` 还不在 PATH 中，会打印 PATH 配置提示
- 使用绝对二进制路径的 MCP 配置片段

如果需要固定当前公测版本用于排障或回滚：

```sh
npx -y patchxnote-agent@0.2.3 install --print-config
```

公测版本默认连接 PatchXNote 公测 API：

```text
https://ws-lab.patch-x.cn/patchnote-test-api
```

登录并检查会话状态。

macOS/Linux：

```sh
patchxnote login
patchxnote auth status
```

Windows PowerShell：

```powershell
patchxnote login
patchxnote auth status
```

启动 MCP server：

```sh
patchxnote mcp serve
```

如果要切换到其他 PatchXNote 环境：

```sh
PATCHXNOTE_SERVER_BASE_URL=<PatchXNote API base URL> \
patchxnote login
```

## MCP 配置

![PatchXNote Agent 架构](./docs/assets/patchxnote-agent-architecture.png)

使用安装器 `--print-config` 打印的配置。典型配置如下：

```json
{
  "mcpServers": {
    "patchxnote": {
      "command": "/absolute/path/to/patchxnote",
      "args": ["mcp", "serve"]
    }
  }
}
```

MCP 配置中不会保存 access token 或 refresh token。PatchXNote Agent 默认把凭据写入 macOS Keychain、Windows Credential Manager 或 Linux Secret Service。显式的 `PATCHXNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true` 文件存储仅保留给本地开发和 CI 冒烟。

## MCP 工具

![PatchXNote Agent 工具能力](./docs/assets/patchxnote-agent-tools.png)

| 工具 | 用途 |
| --- | --- |
| `patchxnote_get_current_user` | 读取当前 PatchXNote 账号投影。 |
| `patchxnote_list_recorder_cards` | 读取已绑定录音卡列表，只返回脱敏标识。 |
| `patchxnote_get_quota_summary` | 读取当前账号额度汇总。 |
| `patchxnote_get_model_usage_summary` | 读取当月模型使用汇总。 |
| `patchxnote_list_memories` | 按平台列出安全的结构化结果元数据。 |
| `patchxnote_search_memories` | 搜索当前会话已授权缓存的记忆元数据。 |
| `patchxnote_get_memory` | 读取单条结构化结果的安全元数据。 |
| `patchxnote_list_webhook_targets` | 列出本地 webhook 别名和脱敏元数据。 |
| `patchxnote_configure_webhook_target` | 创建或更新本地 webhook 目标；URL/密钥是只写输入。 |
| `patchxnote_remove_webhook_target` | 删除本地 webhook 目标并尽力清理已保存密钥。 |
| `patchxnote_list_webhook_templates` | 列出内置 webhook Markdown 模板。 |
| `patchxnote_render_webhook_message` | 把 delivery-document 投影渲染成 Markdown，可选保存草稿。 |
| `patchxnote_export_model_io` | 把显式 model IO JSON 导出到用户指定的本地文件。 |
| `patchxnote_send_webhook` | 手动把 Markdown、草稿、记忆渲染结果或测试消息发送到目标别名。 |
| `patchxnote_list_model_io_traces` | 按平台列出 model IO trace 元数据和 request ID。 |
| `patchxnote_get_model_io_source_text` | 读取显式 source text/安全转写投影字段。 |
| `patchxnote_get_model_io_provider_response` | 只读取模型供应商响应 JSON 字段。 |
| `patchxnote_get_model_io_parsed_result` | 只读取模型解析结果 JSON 字段。 |
| `patchxnote_get_model_io_packaged_result` | 只读取最终封装结构 JSON 字段。 |

记忆类工具必须显式传入 `platform`：`mobile` 或 `desktop`。V1 的记忆响应只包含安全元数据，不重建模型运行响应正文，也不返回旧的完整摘要文本。
webhook MCP 工具复用 CLI 的本地配置、钥匙串、模板和发送模块。工具不会返回完整 webhook URL 或签名密钥；只有 MCP client 明确调用发送工具时才会发起外部网络请求。
Model IO 字段工具是显式、按字段隔离的能力。它可能返回当前登录用户的 source text 或模型/供应商 payload，因此只建议在可信本地 MCP Host 中使用。大字段建议写入显式 `out` 本地文件。

可以在桌面 Agent 中这样询问：

```text
查看我的 PatchXNote 账号和额度状态。
列出我的 PatchXNote 录音卡。
搜索 desktop 平台里和 roadmap 相关的 PatchXNote 记忆。
配置一个名为“产品群 飞书”的飞书 webhook，然后把这段 Markdown 总结发送过去。
列出今天 mobile 平台的 PatchXNote model IO traces，再用 request ID 查看 provider response。
查看这条 PatchXNote 记忆的 provider response，并把 parsed result 保存到本地 JSON 文件。
列出我配置过的 PatchXNote webhook 目标。
删除名为“产品群 飞书”的 PatchXNote webhook 目标。
```

## CLI 命令

```sh
patchxnote version
patchxnote login
patchxnote auth status
patchxnote logout
patchxnote mcp serve
patchxnote webhook set "产品群 飞书" --type feishu --url-stdin
patchxnote webhook test "产品群 飞书"
patchxnote webhook send --target "产品群 飞书" --file ./message.md
patchxnote webhook draft --memory-id <memory_id> --out ./patchxnote-drafts/example
patchxnote webhook send --target "产品群 飞书" --draft ./patchxnote-drafts/example
patchxnote webhook export-model-io --memory-id <memory_id> --out ./patchxnote-drafts/example/model-io.json
patchxnote model-io list --platform mobile
patchxnote model-io source-text --memory-id <memory_id> --platform mobile
patchxnote model-io provider-response --memory-id <memory_id> --platform mobile --out ./provider-response.json
patchxnote model-io parsed-result --memory-id <memory_id> --platform mobile --out ./parsed-result.json
patchxnote model-io packaged-result --request-id <request_id> --platform mobile
patchxnote model-io export --memory-id <memory_id> --platform mobile --out ./model-io.json
```

常用全局参数：

```sh
--server-base-url <url>   PatchXNote API base URL
--profile <name>          本地 profile 名称
--output json             支持时输出机器可读 JSON
--config <path>           非 secret 配置文件路径
```

npm 包本身只是安装/更新/卸载壳：

```sh
npx -y patchxnote-agent@0.2.3 install
npx -y patchxnote-agent@0.2.3 update
npx -y patchxnote-agent@0.2.3 uninstall
```

webhook URL 和飞书/钉钉可选签名密钥只写入本机安全钥匙串，不写普通配置文件。建议用 `--url-stdin` 和 `--secret-stdin` 避免 shell history。CLI 和 MCP 的 webhook 发送都只支持用户手动执行，不跟随重定向，下游平台错误会直接透传给用户。
`patchxnote model-io export` 是完整 model IO 导出的推荐命令。`patchxnote webhook export-model-io` 会继续兼容保留。

## 安全与风险提示

![PatchXNote Agent 安全边界](./docs/assets/patchxnote-agent-safety-boundary.png)

PatchXNote Agent 会让 AI Agent 访问当前登录 PatchXNote 用户的账号元数据。请把 MCP Host 视为受信软件，并注意 prompt、工具调用和日志中可能出现的账号上下文。

默认安全边界：

- Agent 登录态独立于 App/PC 的 `mobile` 和 `desktop` 安装位。
- Agent 读取 PatchXNote 服务端数据时只调用专用只读 `/v1/agent/**` 服务端路由。
- MCP webhook 工具可以写入本地非 secret 目标元数据，把 URL/密钥写入本机安全存储，并手动发送外部 webhook HTTP 请求。
- MCP 在本地通过 stdio 运行；stdout 只用于 JSON-RPC。
- MCP 配置不保存 bearer token、refresh token、验证码、SK 或完整 MAC。
- 录音卡标识会被脱敏；不暴露实时 BLE 状态、电量、存储和录音状态。
- 结构化内容按平台隔离。Agent 不会合并 mobile 和 desktop 内容。
- 工具输出在返回给 MCP client 前会做边界控制和校验。
- webhook URL 和签名密钥只保存在本机安全存储；普通 webhook payload 不包含 access token、refresh token 或导出的 model IO JSON。
- Model IO 字段工具只返回用户请求的字段，不重放模型调用，也不会在单字段响应里夹带无关 model IO 字段。

不要把 access token、refresh token、验证码、原始手机号、完整 MAC、SK、原始音频、完整转写、prompt 或供应商 payload 发到公开 Issue。安全问题请使用 [SECURITY.md](./SECURITY.md) 中的私密流程。

## 当前限制

`0.2.3` 是 beta 版本。

- 默认服务端指向 PatchXNote 公测 API。
- Linux 无桌面/headless 环境可能没有 Secret Service；此时仅本地冒烟可显式开启开发文件存储 fallback。
- 公测期间会持续优化安装流程、MCP 客户端示例和 webhook 格式效果。
- `patchxnote_search_memories` 只搜索当前 MCP 会话中已缓存的元数据。
- 原始音频、完整转写、硬件写操作、模型执行/重放、额度购买/领取、支付和 Admin API 都不在 V1 范围内。

## 常见问题排查

| 问题 | 检查项 |
| --- | --- |
| 安装后找不到 `patchxnote` | 把安装器打印的目录加入 PATH，然后打开新终端。 |
| 登录提示凭据存储不可用 | 检查 macOS Keychain、Windows Credential Manager 或 Linux Secret Service 是否可用且已解锁。本地开发才使用 `PATCHXNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true`。 |
| MCP Host 启动失败 | 使用 `--print-config` 打印出的绝对 `command` 路径。 |
| 记忆列表为空 | 检查是否选择了正确的 `platform`：`mobile` 或 `desktop`。 |
| checksum 校验失败 | 稍后重试或固定已知版本；安装器会拒绝未校验二进制。 |
| 连到了错误服务端 | 设置 `PATCHXNOTE_SERVER_BASE_URL=<PatchXNote API base URL>`。 |

## 验证安装

```sh
npm view patchxnote-agent@0.2.3 version --registry https://registry.npmjs.org
npx -y --registry https://registry.npmjs.org patchxnote-agent@0.2.3 install --dry-run --print-config
patchxnote version
```

发布二进制应报告版本 `0.2.3`，commit 应为 GitHub Release `v0.2.3` 对应的提交。

## 开发

本地检查：

```sh
go test ./...
scripts/e2e/mvp-smoke.sh
node packages/npm/bin/patchxnote-agent.js install --dry-run --print-config
```

MVP smoke 会构建 CLI，执行安装器 dry-run，登录进程内 Agent V1 测试服务，检查 `auth status`，启动 `patchxnote mcp serve`，调用全部 19 个 V1 MCP 工具，覆盖 model IO 发现、字段工具和本地 webhook 发送，登出，并扫描 evidence 中是否出现 secret-like 内容。

修改 CLI 行为、安装器逻辑、MCP 工具、认证、本地缓存或发布配置前，请先阅读：

- [AGENTS.md](./AGENTS.md)
- [docs/engineering-rules.md](./docs/engineering-rules.md)
- [docs/release-and-maintenance-runbook.zh-CN.md](./docs/release-and-maintenance-runbook.zh-CN.md)
- [docs/plans/2026-08-06-agent-v1-mvp.md](./docs/plans/2026-08-06-agent-v1-mvp.md)

## 发布维护说明

详细发包和文档同步 checklist 见 [docs/release-and-maintenance-runbook.zh-CN.md](./docs/release-and-maintenance-runbook.zh-CN.md)。

1. 确认目标 PatchXNote GoServer 已暴露所需 `/v1/agent/**` 路由。
2. 确认 `packages/npm/package.json` 版本与 release tag 一致，tag 不带前缀 `v` 时要匹配包版本。
3. 推送干净 tag，例如 `v0.2.3`。
4. 等待 GitHub Release 产物：`checksums.txt`，以及 Linux/macOS/Windows 的 amd64 和 arm64 二进制。
5. npm publish 前确认 npm Trusted Publishing 已配置：
   - owner/user：`ZsTs119`
   - repository：`patchxnote-agent`
   - workflow filename：`publish-npm.yml`
   - allowed action：`npm publish`
6. 只有 release 产物存在且 trusted publisher 配好后，才发布 npm。
7. Trusted publish 成功后，撤销旧 npm automation token，并禁止该包继续使用 token-based publishing。

## 许可证

当前仓库尚未发布开源许可证。重新分发或嵌入其他产品前，请先联系 PatchXNote。
