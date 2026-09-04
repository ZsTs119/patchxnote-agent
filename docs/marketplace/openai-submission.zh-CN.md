# OpenAI / ChatGPT / Codex Plugin Submission Notes

更新日期：2026-09-03

首版 OpenAI 插件包位置：

```text
packages/plugins/openai/patchxnote-agent/
```

当前定位：skills-only package。它只打包 PatchXNote MCP Skill，不声明 `.app.json`、`.mcp.json` 或远程 MCP 连接。

## 当前文件

- `.codex-plugin/plugin.json`
- `skills/patchxnote-mcp/`，由 `scripts/sync-patchxnote-skill-packages.mjs` 从 canonical skill 同步
- repo marketplace：`.agents/plugins/marketplace.json`

## 本地验证

```sh
node scripts/sync-patchxnote-skill-packages.mjs --check
python3 /mnt/c/Users/11979/.codex/skills/.system/plugin-creator/scripts/validate_plugin.py packages/plugins/openai/patchxnote-agent
node scripts/validate-patchxnote-skill-packages.mjs
```

如果在 Windows 运行验证脚本，把 `/mnt/c/...` 换成对应的 `C:\Users\11979\...` 路径。

## Public Submission Gate

公开提交前必须确认：

- OpenAI Platform 组织具备 Apps Management write 权限。
- 发布者个人或企业身份已验证。
- plugin name、publisher、website、support URL、privacy policy URL、terms URL 彼此一致。
- reviewer/demo account 不要求把 PatchXNote 手机验证码、OAuth code、token 或 webhook secret 粘贴进聊天。
- 默认能力描述只写 PatchXNote MCP setup、summary/memory/model-result、用户确认 webhook workflow。
- 如果要提交 skills-plus-MCP，必须先有真实 public Streamable HTTP MCP URL、OAuth 方案和审查材料。

## 不做的事

- 不把本地 `npx -y patchxnote-agent@latest mcp serve` 塞进公开 OpenAI plugin 的 `.mcp.json`，除非该 surface 的本地执行、用户授权和更新行为都已实测。
- 不声明 raw audio、完整转写、硬件操作、支付、Admin API、模型执行或后台自动 webhook。
- 不把本地 marketplace 安装说成 public universal directory 已发布。

## Starter Prompt 候选

OpenAI plugin card 最多保留三条短 prompt：

```text
Help me connect PatchXNote MCP.
List my latest 5 PatchXNote mobile summaries.
Check my PatchXNote MCP tools.
```

更完整的 review prompt 见 `docs/marketplace/starter-prompts.md` 和 `docs/marketplace/review-test-cases.md`。
