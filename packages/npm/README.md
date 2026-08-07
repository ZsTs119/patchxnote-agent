# PatchNote Agent npm wrapper

[English README](https://github.com/ZsTs119/patchnote-agent#readme) | [简体中文说明](https://github.com/ZsTs119/patchnote-agent/blob/main/README.zh-CN.md)

This npm package is the installer wrapper for PatchNote Agent. It downloads the matching native `patchnote` CLI binary from GitHub Releases, verifies `checksums.txt`, and installs it into a user-writable directory.

```sh
npx -y patchnote-agent@0.1.1 install --print-config
```

PatchNote Agent runs a local stdio MCP server:

```sh
patchnote mcp serve
```

Agent V1 is read-only. It exposes safe PatchNote account projections through dedicated `/v1/agent/**` APIs and does not expose raw audio, full transcripts, SK, full MAC values, hardware write actions, payment flows, or Admin APIs.

For full installation, MCP setup, security notes, and troubleshooting, read the GitHub documentation:

- [English README](https://github.com/ZsTs119/patchnote-agent#readme)
- [简体中文 README](https://github.com/ZsTs119/patchnote-agent/blob/main/README.zh-CN.md)
