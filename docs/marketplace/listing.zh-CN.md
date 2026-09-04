# PatchXNote Agent 市场文案

## 短描述

把本地 AI 客户端安全接入 PatchXNote MCP，用于查看总结、记忆、模型整理结果和用户确认后的 webhook 工作流。

## 长描述

PatchXNote Agent 让可信本地 AI 客户端通过本地 stdio MCP server 接入 PatchXNote。它会指导 AI 完成客户端识别、setup、浏览器 OAuth、验证和常见 PatchXNote 工作流，包括查看手机端或电脑端记忆、检查账号/额度投影、查看模型整理字段、生成 Markdown 草稿，以及在用户确认后手动发送 webhook。

PatchXNote 服务端数据访问是只读、按平台隔离的。本地 webhook 工具是明确的用户触发例外：可以配置本地目标别名，并手动发送用户确认过的内容。PatchXNote Agent 不开放原始音频、默认完整转写、硬件写操作、支付、Admin API、模型执行或后台自动 webhook 推送。

## 标签

`mcp`, `agent-skills`, `productivity`, `notes`, `recordings`, `memory`, `summaries`, `patchxnote`

## 安装

GitHub 发布包含 skill 后：

```sh
npx skills add ZsTs119/patchxnote-agent --skill patchxnote-mcp -g
```

MCP server setup：

```sh
npx -y patchxnote-agent@latest setup --client <client-id>
```

通用 stdio 配置：

```sh
npx -y patchxnote-agent@latest mcp config
```

## 验证

setup 后验证：

- `patchxnote_get_current_user`
- `patchxnote_list_memories`，参数 `{"platform":"mobile","limit":5}`

涉及当前工具数量或名称时，以实时 MCP tool discovery 为准。
