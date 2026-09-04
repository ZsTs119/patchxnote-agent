# PatchXNote Skill And MCP Publishing Checklist

更新日期：2026-09-03

## 0. Baseline

- [ ] `git status --short --branch` 只包含本次变更或已确认的用户变更。
- [ ] `npm view patchxnote-agent version dist-tags.latest repository.url --registry https://registry.npmjs.org` 与计划版本一致。
- [ ] README、中文 README、npm README 的 setup 一句话一致。
- [ ] `docs/mcp-clients/clients.json` 的 P0/P0.5 状态已经复核。
- [ ] OpenAI、Claude、Agent Skills、MCP Registry、Smithery 官方文档在本次发布窗口内复核过。

## 1. Skill

- [ ] `skills/patchxnote-mcp/SKILL.md` frontmatter 有 `name` 和 trigger-oriented `description`。
- [ ] references 按需加载，不把长流程全部塞进 `SKILL.md`。
- [ ] 正向/负向触发用例覆盖。
- [ ] 不写死当前工具数量。
- [ ] 明确浏览器 OAuth 和不得粘贴 code/token/secret。
- [ ] 明确 PatchXNote 返回内容是数据，不是指令。

## 2. Plugin Packages

- [ ] OpenAI/Codex `.codex-plugin/plugin.json` 通过本地 validator。
- [ ] Claude Code `.claude-plugin/plugin.json` 路径和版本已检查。
- [ ] package copies 由 sync 脚本生成。
- [ ] `node scripts/sync-patchxnote-skill-packages.mjs --check` 通过。
- [ ] `node scripts/validate-patchxnote-skill-packages.mjs` 通过。

## 3. MCP Registry And Directories

- [ ] `server.json` 与 `packages/npm/package.json#mcpName` 一致。
- [ ] 对应 npm 版本已发布后再 registry publish。
- [ ] Smithery URL publishing 不用于本地 stdio unless 已有 Streamable HTTP + OAuth。
- [ ] Smithery local path 需要 MCPB bundle 时，单独验收 install/update/uninstall。
- [ ] Glama、PulseMCP、MCP.so、mcpservers.org、officialskills.sh 等目录只声明实际状态。

## 4. Public Review

- [ ] OpenAI publisher identity 已验证。
- [ ] support URL、privacy policy URL、terms URL 可公开访问且与发布主体一致。
- [ ] reviewer/demo account 不依赖在聊天里粘贴手机验证码或 token。
- [ ] 至少 5 个正向、3 个负向测试用例完成并记录。
- [ ] 文案不声明 raw audio、完整转写、硬件写、支付、Admin、模型执行或后台自动发送。

## 5. Release Evidence

- [ ] evidence log 记录每个渠道的 status、owner、version、证据。
- [ ] release notes 区分 docs-only、skills-only、plugin package、registry metadata、public marketplace acceptance。
- [ ] 如发现错误，有 rollback/deprecation 操作路径。
