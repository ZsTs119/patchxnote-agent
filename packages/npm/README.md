# PatchXNote Agent npm wrapper

[English README](https://github.com/ZsTs119/patchxnote-agent#readme) | [简体中文说明](https://github.com/ZsTs119/patchxnote-agent/blob/main/README.zh-CN.md) | [飞书公开使用指南](https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd)

This npm package is the installer wrapper for PatchXNote Agent. It downloads the matching native `patchxnote` CLI binary from GitHub Releases, verifies `checksums.txt`, and installs it into a user-writable directory.

Give this one-line prompt to a local-command-capable AI assistant:

```text
Install and connect PatchXNote Agent by following the public guide at https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd: run npx -y patchxnote-agent install --print-config, guide me through patchxnote login, and add the printed MCP JSON config to this AI assistant. Do not ask me to paste OTP codes, access tokens, or refresh tokens into chat. GitHub repository: https://github.com/ZsTs119/patchxnote-agent
```

Or run the install command manually:

```sh
npx -y patchxnote-agent install --print-config
```

PatchXNote Agent runs a local stdio MCP server:

```sh
patchxnote mcp serve
```

Agent V1 is read-only. It stores credentials in the OS-native keychain when available, exposes safe PatchXNote account projections through dedicated `/v1/agent/**` APIs, and does not expose raw audio, full transcripts, SK, full MAC values, hardware write actions, payment flows, or Admin APIs.

For full installation, MCP setup, security notes, and troubleshooting, read the GitHub documentation:

- [English README](https://github.com/ZsTs119/patchxnote-agent#readme)
- [简体中文 README](https://github.com/ZsTs119/patchxnote-agent/blob/main/README.zh-CN.md)
- [PatchXNote Agent 公测使用指南（飞书公开版）](https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd)
