# PatchXNote Agent npm wrapper

[English README](https://github.com/ZsTs119/patchxnote-agent#readme) | [简体中文说明](https://github.com/ZsTs119/patchxnote-agent/blob/main/README.zh-CN.md) | [飞书公开使用指南](https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd)

This npm package is the installer and thin launcher wrapper for PatchXNote Agent. It downloads or verifies the matching native `patchxnote` CLI binary from GitHub Releases, verifies `checksums.txt`, and installs it into a user-writable directory before delegating runtime commands to that binary.

PatchXNote Agent is the local AI assistant connector for PatchXNote. It lets an MCP-capable AI assistant find PatchXNote records, inspect AI-generated results, create Markdown drafts, and manually send user-approved messages to Feishu, DingTalk, or another webhook.

Give this one-line prompt to a local-command-capable AI assistant:

```text
Install PatchXNote Agent for this local MCP client: run npx -y patchxnote-agent@latest setup --client <my-client> in the same OS/runtime that will launch the MCP server, guide me through browser login there, and keep MCP config secret-free. If setup cannot write config, run npx -y patchxnote-agent@latest mcp config and use the printed stdio config. Do not ask me to paste OTP codes, OAuth codes, access tokens, refresh tokens, or webhook secrets into chat. References: https://github.com/ZsTs119/patchxnote-agent and https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd
```

Run setup for a supported local client:

```sh
npx -y patchxnote-agent@latest setup --client cursor
npx -y patchxnote-agent@latest setup --client vscode
npx -y patchxnote-agent@latest setup --client codex
```

PatchXNote Agent currently exposes two login surfaces and two MCP service shapes:

| Mode | Entry | Use when |
| --- | --- | --- |
| Browser MCP login | `npx -y patchxnote-agent@latest mcp login` | Recommended for desktop editors and local MCP hosts. |
| Terminal CLI login | `npx -y patchxnote-agent@latest login` | Terminal-only users, headless runtimes, or the legacy local Agent path. |
| Local MCP service | `npx -y patchxnote-agent@latest mcp serve` | VS Code, Cursor, Codex, Claude Desktop, Windsurf, Trae, Qoder, WorkBuddy, and other stdio MCP clients. |
| Hosted remote MCP gateway | `https://ws-lab.patch-x.cn/patchnote-test-api/mcp` | Feishu Aily, Doubao Work Partner, Tencent Agent Development Platform, enterprise WorkBuddy, and other platform clients that cannot run local commands. |

Log in with browser OAuth, then check or clear the MCP login state:

```sh
npx -y patchxnote-agent@latest mcp login
npx -y patchxnote-agent@latest mcp status
npx -y patchxnote-agent@latest mcp logout --local-only
```

Print the generic MCP config manually:

```sh
npx -y patchxnote-agent@latest mcp config
```

The generated config uses this universal local stdio MCP command:

```sh
npx -y patchxnote-agent@latest mcp serve
```

The older terminal Agent login remains available for legacy local CLI/MCP fallback:

```sh
npx -y patchxnote-agent@latest login
```

`setup` reuses the same `mcp login` browser OAuth flow. The GoServer website owns phone OTP input; the local Agent owns the loopback callback, token exchange, secure storage, and stdio bridge. `mcp serve` never opens a browser during editor startup. It stores credentials in the OS-native keychain and keeps MCP config free of phone numbers, OTP codes, access tokens, refresh tokens, webhook secrets, and base URL by default.

P0 local client IDs: `vscode`, `cursor`, `codex`, `claude-code`, `claude-desktop`, `windsurf`, `trae`, `qoder`, `workbuddy`.

Platform clients such as Feishu Aily, Doubao Work Partner, Tencent Agent Development Platform, and enterprise WorkBuddy need the hosted remote MCP gateway plus platform-console acceptance; local `npx` setup is for desktop and terminal clients.

If a client rejects `npx` or times out during first binary download, run the stable absolute-path fallback:

```sh
npx -y patchxnote-agent@latest install --print-config
```

It currently exposes 19 MCP tools grouped around account/record lookup, webhook configuration and sending, and explicit AI result inspection.

Common CLI examples:

```sh
patchxnote model-io list --platform mobile
patchxnote model-io packaged-result --memory-id <memory_or_request_id> --platform mobile --out ./packaged-result.json
patchxnote model-io provider-response --request-id <request_id> --platform mobile --out ./provider-response.json
patchxnote webhook set "Product Feishu" --type feishu --url-stdin
patchxnote webhook send --target "Product Feishu" --file ./message.md
```

Webhook aliases can contain Chinese text, spaces, and dots.

Server-backed PatchXNote data access remains read-only through dedicated `/v1/agent/**` APIs. Record lookup can include formal saved results and readable model-generated outputs returned by the server. Local webhook tools can configure named targets and perform explicit manual sends. AI result tools can inspect source text, AI response, parsed result, and final result when explicitly called. MCP config contains no phone number, OTP, access token, refresh token, webhook secret, or base URL by default. The CLI stores credentials and webhook secrets in the OS-native keychain when available, never lists webhook URLs or signing secrets back, and does not expose raw audio, audio downloads, hardware write actions, payment flows, or Admin APIs.

For full installation, MCP setup, security notes, and troubleshooting, read the GitHub documentation:

- [English README](https://github.com/ZsTs119/patchxnote-agent#readme)
- [简体中文 README](https://github.com/ZsTs119/patchxnote-agent/blob/main/README.zh-CN.md)
- [PatchXNote Agent 公测使用指南（飞书公开版）](https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd)
