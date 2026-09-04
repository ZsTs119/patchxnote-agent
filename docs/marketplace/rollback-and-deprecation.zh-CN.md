# PatchXNote Skill And Marketplace Rollback

更新日期：2026-09-03

本文用于 skill、plugin、MCP registry 和第三方 listing 出问题时快速收口。

## 触发条件

- skill 指令误导 AI 请求验证码、OAuth code、token 或 webhook secret。
- skill 错误扩大能力边界，例如暗示 raw audio、硬件写、支付、Admin 或模型执行。
- plugin manifest 导致无法安装、无法更新或加载旧缓存。
- marketplace/listing 文案误称已平台验收、已生产可用或已官方认证。
- npm package、GitHub release、MCP registry 或第三方目录引用了错误版本。
- 安全报告指出返回内容存在 prompt injection、secret 泄漏或过宽数据暴露。

## 快速处理

1. 暂停宣传错误渠道。
2. 在 `docs/marketplace/evidence-log.md` 将渠道状态改为 `blocked`、`deprecated` 或 `removed`。
3. 修正 canonical source：`skills/patchxnote-mcp/` 或 registry/listing 源文件。
4. 运行同步和校验：

```sh
node scripts/sync-patchxnote-skill-packages.mjs
node scripts/sync-patchxnote-skill-packages.mjs --check
node scripts/validate-patchxnote-skill-packages.mjs
git diff --check
```

5. 需要公开发版时，发布新的 patch 版本，不覆盖已发布版本。
6. 如果 npm 版本有严重问题，按主 runbook 使用 `npm deprecate`。
7. 如果 marketplace 支持下架或隐藏，先下架错误版本，再提交修正版。

## 用户沟通口径

应说明：

- 影响渠道
- 受影响版本
- 建议升级或回滚命令
- 哪些能力可继续使用
- 哪些能力暂停使用

不应包含：

- 用户手机号、验证码、token、webhook secret、原始内容、完整转写、provider payload
- 未确认的根因
- 未通过验收的恢复承诺
