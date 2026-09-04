# PatchXNote MCP Skill Platform Matrix

更新日期：2026-09-03

本文记录 PatchXNote MCP Skill 和 PatchXNote Agent MCP server 的首版分发渠道。状态只描述仓库侧准备度，不代表平台已经审核通过。

## 状态定义

| 状态 | 含义 |
| --- | --- |
| `drafted` | 仓库中已有草稿文件或文案。 |
| `locally_smoked` | 已在本机完成该渠道当前可执行的结构、发现、安装或加载验证；具体证据以 `docs/marketplace/evidence-log.md` 为准。 |
| `submitted` | 已提交到平台或目录。 |
| `accepted` | 平台审核通过。 |
| `indexed` | 目录页面可搜索或可访问。 |
| `blocked` | 等待账号、审核、生产 URL、认证方案或平台能力。 |

## 渠道矩阵

| 渠道 | 分发对象 | 仓库文件 | 当前状态 | 不能提前声明 |
| --- | --- | --- | --- | --- |
| Agent Skills standard | 支持 Agent Skills 的本地 AI 客户端 | `skills/patchxnote-mcp/` | `locally_smoked` | 不能说所有客户端都已自动触发。 |
| `npx skills` / skills.sh | Codex、Cursor、Claude Code、OpenCode、Gemini CLI 等兼容生态 | `skills/patchxnote-mcp/` | `locally_smoked` | 不能说 public GitHub install 可用，直到包含 skill 的 commit/tag 已推送。 |
| OpenAI / ChatGPT / Codex plugin | ChatGPT 和 Codex 插件市场 | `packages/plugins/openai/patchxnote-agent/`、`.agents/plugins/marketplace.json` | `locally_smoked` | 不能说已进入 universal public directory。 |
| Claude Code plugin marketplace | Claude Code 插件市场 | `packages/plugins/claude/patchxnote-agent/`、`.claude-plugin/marketplace.json` | `drafted` | 不能说 URL marketplace 可用，直到路径和缓存更新实测。 |
| Claude.ai / Claude API custom skills | 上传 zip 的 custom skill | `skills/patchxnote-mcp/` 生成 zip | `drafted` | 不能说已上传或组织可用。 |
| Cursor Skills | Cursor 的 skills 功能 | `skills/patchxnote-mcp/`、`docs/marketplace/cursor-skill-install.md` | `drafted` | 不能说 Cursor 自动安装已验收。 |
| MCPB / Desktop Extension | 可安装的本地 MCP bundle | 暂不生成正式 `.mcpb` | `blocked` | 不能说首版有 MCPB 一键包。 |
| Official MCP Registry | MCP server 元数据发现 | `server.json`、`packages/npm/package.json#mcpName` | `drafted` | 不能说已发布到 registry。 |
| Smithery | MCP server 搜索、分发、扫描 | `smithery.yaml`、listing docs | `drafted` | 不能说 hosted/managed 分发已通过。 |
| Glama / PulseMCP / MCP.so | 第三方 MCP 目录 | listing docs | `drafted` | 不能说官方认证或平台 acceptance。 |
| officialskills.sh / awesome-agent-skills | Skill 发现目录 | listing docs、starter prompts | `drafted` | 不能说收录前已经 indexed。 |

## 首版推荐顺序

1. 先发布 canonical skill 文件和 README 安装说明。
2. 再做 OpenAI/Codex 和 Claude Code 本地 marketplace 冒烟。
3. 再做 MCP Registry 和 Smithery 元数据验证。
4. 最后提交第三方目录和官方平台审核。
