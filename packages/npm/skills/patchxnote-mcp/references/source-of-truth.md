# PatchXNote MCP Source Of Truth

Use this reference when links, versions, marketplace claims, or platform-specific publishing steps matter.

## PatchXNote Links

- GitHub repository: `https://github.com/ZsTs119/patchxnote-agent`
- npm package: `https://www.npmjs.com/package/patchxnote-agent`
- Feishu public guide: `https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd`

## Repository Docs

Check these before changing onboarding or publishing claims:

- `README.md`
- `README.zh-CN.md`
- `packages/npm/README.md`
- `docs/engineering-rules.md`
- `docs/release-and-maintenance-runbook.zh-CN.md`
- `docs/plans/2026-08-06-agent-v1-mvp.md`
- `docs/mcp-clients/clients.json`
- `docs/plans/2026-09-03-patchxnote-mcp-skill-marketplace-checklist.md`

Server contracts take precedence over this repository if they conflict:

- `../patchxNoteGoServer/docs/engineering/agent-access-v1.md`
- `../patchxNoteGoServer/docs/integrations/apifox/shared/integration-guide.zh-CN.md`

## Platform Docs To Recheck Before Release

- OpenAI skills: `https://developers.openai.com/plugins/build/skills`
- OpenAI plugin packaging: `https://developers.openai.com/plugins/build/plugins`
- OpenAI plugin submission: `https://developers.openai.com/plugins/deploy/submission`
- Claude Code plugin marketplace: `https://code.claude.com/docs/en/plugin-marketplaces`
- Agent Skills specification: `https://agentskills.io/specification`
- MCP Registry quickstart: `https://github.com/modelcontextprotocol/registry/blob/main/docs/modelcontextprotocol-io/quickstart.mdx`
- Smithery publishing: `https://smithery.ai/docs/build/publish`
- Cursor Skills docs: `https://cursor.com/docs/skills`

Platform schemas and submission flows can change. Re-check these docs during each release. Do not rely only on old plan notes.
