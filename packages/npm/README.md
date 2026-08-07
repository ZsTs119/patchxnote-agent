# PatchXNote Agent npm wrapper

[English README](https://github.com/ZsTs119/patchxnote-agent#readme) | [简体中文说明](https://github.com/ZsTs119/patchxnote-agent/blob/main/README.zh-CN.md)

This npm package is the installer wrapper for PatchXNote Agent. It downloads the matching native `patchxnote` CLI binary from GitHub Releases, verifies `checksums.txt`, and installs it into a user-writable directory.

```sh
npx -y patchxnote-agent@0.2.1 install --print-config
```

PatchXNote Agent runs a local stdio MCP server:

```sh
patchxnote mcp serve
```

Agent V1 is read-only. It stores credentials in the OS-native keychain when available, exposes safe PatchXNote account projections through dedicated `/v1/agent/**` APIs, and does not expose raw audio, full transcripts, SK, full MAC values, hardware write actions, payment flows, or Admin APIs.

For full installation, MCP setup, security notes, and troubleshooting, read the GitHub documentation:

- [English README](https://github.com/ZsTs119/patchxnote-agent#readme)
- [简体中文 README](https://github.com/ZsTs119/patchxnote-agent/blob/main/README.zh-CN.md)
