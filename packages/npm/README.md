# PatchXNote Agent npm wrapper

[English README](https://github.com/ZsTs119/patchxnote-agent#readme) | [简体中文说明](https://github.com/ZsTs119/patchxnote-agent/blob/main/README.zh-CN.md) | [飞书公开使用指南](https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd)

This npm package is the installer wrapper for PatchXNote Agent. It downloads the matching native `patchxnote` CLI binary from GitHub Releases, verifies `checksums.txt`, and installs it into a user-writable directory.

PatchXNote Agent is the local AI assistant connector for PatchXNote. It lets an MCP-capable AI assistant find PatchXNote records, inspect AI-generated results, create Markdown drafts, and manually send user-approved messages to Feishu, DingTalk, or another webhook.

Give this one-line prompt to a local-command-capable AI assistant:

```text
Install and connect PatchXNote Agent by following the public guide at https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd: run npx -y patchxnote-agent install --print-config, guide me through patchxnote login, and add the printed MCP JSON config to this AI assistant. Do not ask me to paste OTP codes, access tokens, or refresh tokens into chat. GitHub repository: https://github.com/ZsTs119/patchxnote-agent
```

Or run the install command manually:

```sh
npx -y patchxnote-agent@0.2.5 install --print-config
```

PatchXNote Agent runs a local stdio MCP server:

```sh
patchxnote mcp serve
```

It currently exposes 19 MCP tools grouped around account/record lookup, webhook configuration and sending, and explicit AI result inspection.

Common CLI examples:

```sh
patchxnote model-io list --platform mobile
patchxnote model-io provider-response --request-id <request_id> --platform mobile --out ./provider-response.json
patchxnote webhook set "Product Feishu" --type feishu --url-stdin
patchxnote webhook send --target "Product Feishu" --file ./message.md
```

Webhook aliases can contain Chinese text, spaces, and dots.

Server-backed PatchXNote data access remains read-only through dedicated `/v1/agent/**` APIs. Local webhook tools can configure named targets and perform explicit manual sends. AI result tools can inspect source text, AI response, parsed result, and final result when explicitly called. The CLI stores credentials and webhook secrets in the OS-native keychain when available, never lists webhook URLs or signing secrets back, and does not expose raw audio, audio downloads, hardware write actions, payment flows, or Admin APIs.

For full installation, MCP setup, security notes, and troubleshooting, read the GitHub documentation:

- [English README](https://github.com/ZsTs119/patchxnote-agent#readme)
- [简体中文 README](https://github.com/ZsTs119/patchxnote-agent/blob/main/README.zh-CN.md)
- [PatchXNote Agent 公测使用指南（飞书公开版）](https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd)
