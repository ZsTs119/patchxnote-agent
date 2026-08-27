# PatchXNote MCP 客户端详情页文案

以下文案用于官网客户端详情页。所有命令和配置都不包含用户密钥。

## 通用配置

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

通用中文说明：

```text
先运行 mcp login 完成浏览器 OAuth，或让 setup 自动复用同一套登录流程。PatchXNote 会把登录态保存到本机安全存储。编辑器里的 MCP 配置只负责启动本地服务，不保存手机号、验证码或 token。
```

Generic English copy:

```text
Run mcp login once, or let setup reuse the same browser OAuth flow. PatchXNote stores the connector session in the OS keychain, while the MCP config stays secret-free and only launches the local server.
```

## VS Code / GitHub Copilot

主命令：

```sh
npx -y patchxnote-agent@latest setup --client vscode
```

中文说明：

```text
适合 VS Code 和 GitHub Copilot Agent 模式。setup 会写入用户级 MCP 配置，并保留其他已有 MCP servers。写入前会确认，写入时会备份。
```

English:

```text
Use this for VS Code and GitHub Copilot Agent mode. Setup writes a user-level MCP entry, preserves existing servers, confirms before writing, and creates a backup.
```

注意事项：

```text
如果你在 VS Code Remote、Dev Container 或 WSL 中运行 MCP，请在那个远程/WSL 环境里执行 setup。
```

## Cursor

主命令：

```sh
npx -y patchxnote-agent@latest setup --client cursor
```

中文说明：

```text
适合 Cursor 本地 Agent。setup 写入用户级 Cursor MCP 配置；官网后续可以提供 Cursor deeplink，但只有真实验收后才显示为一键安装。
```

English:

```text
Use this for local Cursor agents. Setup writes the user-level Cursor MCP config. A Cursor deeplink can be shown after real client acceptance.
```

## Codex / ChatGPT Desktop / Codex IDE

主命令：

```sh
npx -y patchxnote-agent@latest setup --client codex
```

备用命令：

```sh
codex mcp add patchxnote -- npx -y patchxnote-agent@latest mcp serve
```

中文说明：

```text
Codex CLI、ChatGPT Desktop 和 Codex IDE 共享 Codex MCP 配置。setup 可以写入用户级 config.toml；如果你更想用 Codex 官方命令，可以复制备用命令。
```

English:

```text
Codex CLI, ChatGPT Desktop, and the Codex IDE extension share Codex MCP config. Setup can write the user-level config.toml, or you can use the Codex command directly.
```

## Claude Code

主命令：

```sh
npx -y patchxnote-agent@latest setup --client claude-code
```

备用命令：

```sh
claude mcp add patchxnote -- npx -y patchxnote-agent@latest mcp serve
```

中文说明：

```text
第一版不假设 Claude Code 已安装，也不直接改它的配置。setup 会给出官方 CLI 命令和 MCP 配置片段。
```

English:

```text
V1 does not assume Claude Code is installed or edit its files directly. Setup prints the official CLI command and the MCP config snippet.
```

## Claude Desktop

主命令：

```sh
npx -y patchxnote-agent@latest setup --client claude-desktop
```

中文说明：

```text
setup 写入 Claude Desktop 的本机配置文件，并创建备份。写入后请重启 Claude Desktop。
```

English:

```text
Setup writes the local Claude Desktop config and creates a backup. Restart Claude Desktop after setup.
```

## Windsurf

主命令：

```sh
npx -y patchxnote-agent@latest setup --client windsurf
```

中文说明：

```text
setup 使用 Windsurf 当前公开 MCP 配置路径写入本地 stdio server。写入后在 Windsurf 中刷新 MCP servers。
```

English:

```text
Setup writes the local stdio server using Windsurf's documented MCP config path. Refresh MCP servers in Windsurf afterward.
```

## Trae / Trae CN / TraeWork Code

主命令：

```sh
npx -y patchxnote-agent@latest setup --client trae
```

中文说明：

```text
第一版采用手动 UI 配置。setup 会输出安全 MCP JSON，你在 Trae 的 MCP 设置里粘贴即可。
```

English:

```text
V1 uses the manual UI flow. Setup prints a safe MCP JSON snippet that you can paste into Trae MCP settings.
```

## Qoder

主命令：

```sh
npx -y patchxnote-agent@latest setup --client qoder
```

中文说明：

```text
Qoder 支持 deeplink 和手动 MCP 配置。第一版先输出 deeplink 和配置片段，真实客户端验收后再把官网按钮升级为一键安装。
```

English:

```text
Qoder supports deeplinks and manual MCP configuration. V1 prints the deeplink and config snippet; the website can promote one-click after real acceptance.
```

## WorkBuddy / Tencent CodeBuddy WorkBuddy

主命令：

```sh
npx -y patchxnote-agent@latest setup --client workbuddy
```

中文说明：

```text
WorkBuddy 本地桌面形态先走 MCP + CLI 手动接入；企业/平台形态走远程 MCP 网关，不能只靠本机 npx。
```

English:

```text
Local WorkBuddy starts with the MCP + CLI manual path. Enterprise/platform mode needs the remote MCP gateway and cannot rely on local npx alone.
```

## Feishu Aily / Doubao Work Partner

远程地址：

```text
https://ws-lab.patch-x.cn/patchnote-test-api/mcp
```

中文说明：

```text
这是平台型客户端。第一版平台闭环需要远程 MCP 网关、平台授权和真实控制台验收；本地 setup 命令不能完成云平台接入。
```

English:

```text
This is a platform client. The platform loop needs a remote MCP gateway, platform authorization, and real console acceptance. Local setup is not enough.
```

## Tencent Agent Development Platform / Enterprise WorkBuddy

远程地址：

```text
https://ws-lab.patch-x.cn/patchnote-test-api/mcp
```

中文说明：

```text
平台端只接远程 MCP URL。上线前需要完成 initialize、tools/list、tools/call 和一条安全只读调用验收。
```

English:

```text
The platform path uses the remote MCP URL only. Before launch, validate initialize, tools/list, tools/call, and one safe read-only call.
```
